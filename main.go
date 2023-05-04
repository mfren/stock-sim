package main

import (
	"fmt"
	"math"
	"math/rand"
)

func main() {

	const simLength int = 100

	const avgReturn float64 = 0.01
	const standDev float64 = 0.1
	const variance float64 = 0.01

	const drift float64 = avgReturn + (0.5 * variance)

	var values [simLength]float64

	values[0] = 100.0

	for i := 1; i < simLength; i++ {

		var lastPrice = values[i-1]

		var random = rand.Float64()
		var volatility = standDev * random
		var brownianMotion = drift + volatility

		var nowPrice = lastPrice * math.Pow(math.E, brownianMotion)
		values[i] = nowPrice

		fmt.Println(nowPrice)
	}

}
