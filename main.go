package main

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

const apiAddr string = "https://www.alphavantage.co"

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

var myClient = &http.Client{Timeout: 10 * time.Second}

func getData(stockName string, target interface{}) error {

	var url = fmt.Sprintf(
		"%s/query?function=TIME_SERIES_DAILY_ADJUSTED&symbol=%s&outputsize=full&apikey=%s",
		apiAddr, stockName, os.Getenv("API_KEY"),
	)

	fmt.Println(url)

	r, err := myClient.Get(url)

	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
	}
	defer r.Body.Close()

	// fmt.Print(res.Body.Read())
	fmt.Printf("client: got response!\n")
	fmt.Printf("client: status code: %d\n", r.StatusCode)

	return json.NewDecoder(r.Body).Decode(target)
}

func main() {

	data := new(ResponseBody)

	getData("BA.LON", data)

	// Calculate the average daily increase
	// Calculate the variance of daily increases across the data
	//		We need the mean of the daily increases
	//		total number of entries

	var avgReturn float64 = 0
	var totalEntries int = 0
	var diffs []float64

	for _, value := range data.TimeSeriesDaily {

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
		// diff := (close / open) - 1
		fmt.Println(diff)
		diffs = append(diffs, diff)
		avgReturn += diff
		totalEntries++
	}

	avgReturn /= float64(totalEntries)
	fmt.Println("Average increase:", avgReturn)

	// Calculate variance
	var variance float64 = 0
	for _, diff := range diffs {
		variance += math.Pow(diff-avgReturn, 2)
	}
	variance /= float64(totalEntries)
	fmt.Println("Variance:", variance)

	standDev := math.Sqrt(variance)
	fmt.Println("Standard Deviation:", standDev)

	// ^^^ Stats ^^^
	// vvv  Sim  vvv

	const numSims = 10000
	const simLength = 365

	var simValues [numSims][simLength]float64

	var drift float64 = avgReturn - (0.5 * variance)
	fmt.Println("Drift:", drift)

	for simNum := 0; simNum < numSims; simNum++ {
		var values = &simValues[simNum]

		values[0] = 1

		for i := 1; i < simLength; i++ {

			var lastPrice = values[i-1]

			// random := rand.NormFloat64()
			// change := lastPrice * ((avgReturn * 1) + (standDev * random * math.Sqrt(1)))
			// values[i] = lastPrice + change

			var random = rand.NormFloat64()
			var volatility = standDev * random
			var brownianMotion = drift + volatility

			var newPrice = lastPrice * math.Pow(math.E, brownianMotion)
			values[i] = newPrice

		}

	}

	// ^^^     Sims      ^^^
	// vvv Meta Analysis vvv

	// Sort the sims based on the end value
	sort.Slice(simValues[:], func(i, j int) bool {
		return simValues[i][simLength-1] < simValues[j][simLength-1]
	})

	// We want to select the 99th, 95th, 85th and 75th percentile sims
	selectedSims := []([simLength]float64){simValues[4000], simValues[4500], simValues[5000], simValues[5500], simValues[6000]}

	// vvv CSV  vvv

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

	for i := 0; i < simLength; i++ {
		fmt.Fprintf(fo, "%d,", i)
		for simNum := 0; simNum < 5; simNum++ {
			fmt.Fprintf(fo, "%.5f,", selectedSims[simNum][i]-1)
		}
		fmt.Fprint(fo, "\n")
	}
}
