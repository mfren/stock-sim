package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

var NUM_SIMS = flag.Int("sims", 10_000, "Number of simulations to run")
var SIM_LENGTH = flag.Int("len", 365, "Length of each simulation")
var DISPLAY_STATS = flag.Bool("stats", false, "Display dataset statistics")
var API_KEY = flag.String("apikey", "", "API Key")
var TICKER = flag.String("stock", "GOOG", "Stock ticker to simulate")

const API_ADDR = "https://www.alphavantage.co"

var HTTP_CLIENT = &http.Client{Timeout: 10 * time.Second}

type ResponseBody struct {
	MetaData        MetaData             `json:"Meta Data"`
	TimeSeriesDaily map[string]DataPoint `json:"Time Series (Daily)"`
}

type MetaData struct {
	Information   string `json:"1. Information"`
	Symbol        string `json:"2. Symbol"`
	LastRefreshed string `json:"3. Last Refreshed"`
	OutputSize    string `json:"4. Output Size"`
	TimeZone      string `json:"5. Time Zone"`
}

type DataPoint struct {
	Open             string `json:"1. open"`
	High             string `json:"2. high"`
	Low              string `json:"3. low"`
	Close            string `json:"4. close"`
	AdjustedClose    string `json:"5. adjusted close"`
	Volume           string `json:"6. volume"`
	DividendAmount   string `json:"7. dividend amount"`
	SplitCoefficient string `json:"8. split coefficient"`
}

/* Gets the stock data from the web api */
func getData(stockName string, target interface{}) error {

	var url = fmt.Sprintf(
		"%s/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&outputsize=full&apikey=%s",
		API_ADDR, stockName, *API_KEY,
	)

	r, err := HTTP_CLIENT.Get(url)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

/* Calculates the differences between open/closes for all datapoints */
func calcDiffs(data map[string]DataPoint, target []float64) []float64 {

	for _, value := range data {

		var open, close float64
		var err error

		if open, err = strconv.ParseFloat(value.Close, 64); err != nil {
			fmt.Println(err.Error())
			err = nil
		}

		if close, err = strconv.ParseFloat(value.Open, 64); err != nil {
			fmt.Println(err.Error())
			err = nil
		}

		diff := math.Log(close / open)
		target = append(target, diff)
	}

	return target
}

/* Calculates the mean average across a dataset of float64s */
func calcMeanAvg(data []float64) float64 {

	total := 0.0
	num := 0

	for _, val := range data {
		total += val
		num++
	}

	return total / float64(num)
}

/* Calculates the variance across a dataset of float64s */
func calcVariance(avg float64, data []float64) float64 {

	// Variance uses the formula:
	//		(Sum((x - avg) ^ 2)) / n
	// Where n is the number of values, x is the value, and avg is the average across the set
	// https://en.wikipedia.org/wiki/Variance

	variance := 0.0
	num := 0

	for _, diff := range data {
		variance += math.Pow(diff-avg, 2)
		num++
	}

	return variance / float64(num)
}

func main() {

	// Pull in necessary CLI args, setting defaults if not there
	flag.Parse()

	// Get data from the web datasource
	data := new(ResponseBody)
	getData(*TICKER, data)

	// We're using brownian motion to determine stock price rises, so we need the following info from the dataset
	//	1. Average daily increase
	//	2. Variance of daily increases
	//	3. Standard Deviation of daily increases (simply the square root of the variance)

	// Allocate space for the daily increases, and use the helper function to calculate them
	var diffs []float64
	diffs = calcDiffs(data.TimeSeriesDaily, diffs)

	avg := calcMeanAvg(diffs)
	variance := calcVariance(avg, diffs)
	stdDev := math.Sqrt(variance)
	drift := avg - (0.5 * variance)

	if *DISPLAY_STATS {
		fmt.Printf("Average: %.5f\nVariance: %.5f\nStandard Deviation: %.5f\n\nDrift: %.5f\n", avg, variance, stdDev, drift)
	}

	// Define storage for simulation numbers
	var simValues [][]float64 = make([][]float64, *NUM_SIMS)

	for simNum := 0; simNum < *NUM_SIMS; simNum++ {
		var values = make([]float64, *SIM_LENGTH)
		simValues[simNum] = values

		// We start at 1, as this allows the simualtion to then be applied to a number of different starting points
		values[0] = 1

		for i := 1; i < *SIM_LENGTH; i++ {

			random := rand.NormFloat64()
			volatility := stdDev * random
			brownianMotion := drift + volatility

			newValue := values[i-1] * math.Pow(math.E, brownianMotion)
			values[i] = newValue
		}
	}

	// Sort the sims based on the end value
	sort.Slice(simValues[:], func(i, j int) bool {
		return simValues[i][*SIM_LENGTH-1] < simValues[j][*SIM_LENGTH-1]
	})

	// We want to select the 40th, 45th, 50th, 55th and 60th percentile simulations
	selectedSims := []([]float64){
		simValues[int(math.Floor(float64(*NUM_SIMS)*0.4))],
		simValues[int(math.Floor(float64(*NUM_SIMS)*0.45))],
		simValues[int(math.Floor(float64(*NUM_SIMS)*0.5))],
		simValues[int(math.Floor(float64(*NUM_SIMS)*0.55))],
		simValues[int(math.Floor(float64(*NUM_SIMS)*0.60))],
	}

	// We now are going to write the data to a CSV, however, we need to write each time interval as a row

	// Open a new CSV file
	fo, err := os.Create("output.csv")
	if err != nil {
		panic(err)
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// Loop through each time period and write all the selected simulation's data for that time period on the row
	for i := 0; i < *SIM_LENGTH; i++ {
		fmt.Fprintf(fo, "%d,", i)
		for simNum := 0; simNum < len(selectedSims); simNum++ {
			fmt.Fprintf(fo, "%.5f,", selectedSims[simNum][i]-1)
		}
		fmt.Fprint(fo, "\n")
	}
}
