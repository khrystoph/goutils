package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"time"
)

const (
	RSI_PERIODS = 14
)

type StockPrice struct {
	open, close, high, low float64
}

type StockInstance struct {
	periodTime                                            time.Time
	periodType                                            string //This can be one of 1min, 5min, 15min, 30min, 1hr, 2hr, 4hr, 12hr, 1day, 1wk, 1mo, 3mo, 6mo, 1yr
	open, close, high, low, percentChange, absoluteChange float64
}

func CalculatePercentLossGain(candlestick *StockPrice) (periodPercentGainedLost float64) {
	periodPercentGainedLost = (candlestick.close - candlestick.open) / candlestick.open * 100.0
	return periodPercentGainedLost
}

func CalculateTotalLossGain(candlestick *StockPrice) (periodGainedLost float64) {
	periodGainedLost = candlestick.close - candlestick.open
	return periodGainedLost
}

func CalculateAverageGain(vals []float64) (averageGainLoss float64) {
	averageGainLoss = (vals[len(vals)-1] - vals[0]) / float64(len(vals))
	return averageGainLoss
}

func RelativeStrengthIndex(prices []float64, period int64, prevAvgGain, prevAvgLoss float64) (relativeStrengthIndex, avgGain, avgLoss float64, err error) {
	err = nil
	totalGain, totalLoss := 0.0, 0.0
	fmt.Printf("length of prices slice: %v\nPrices: %v\n", len(prices), prices)
	if int64(len(prices)) < period {
		err = errors.New("not enough data points to calculate RSI, previous Avg. Gain, and previous Avg. Loss for given lookback period\n")
		return 0.0, 0.0, 0.0, err
	} else if int64(len(prices)) == period {
		for _, dailyGain := range prices {
			if dailyGain <= 0.0 {
				totalLoss += math.Abs(dailyGain)
			} else {
				totalGain += math.Abs(dailyGain)
			}
		}
		avgGain = totalGain / float64(period)
		avgLoss = totalLoss / float64(period)
	} else {
		fmt.Printf("prevAvgGain: %.2f\n", prevAvgGain)
		fmt.Printf("prevAvgLoss: %.2f\n", prevAvgLoss)
		fmt.Printf("last price in slice: %.2f\n", prices[len(prices)-1])
		fmt.Printf("length: %d\n", len(prices))
		if prices[len(prices)-1] > 0.0 {
			avgGain = (prevAvgGain*(float64(period)-1.0) + math.Abs(prices[len(prices)-1])) / float64(period)
			avgLoss = prevAvgLoss * (float64(period) - 1.0) / float64(period)
		} else {
			avgGain = prevAvgGain * (float64(period) - 1.0) / float64(period)
			avgLoss = (prevAvgLoss*(float64(period)-1.0) + math.Abs(prices[len(prices)-1])) / float64(period)
		}
	}
	relativeStrengthIndex = 100.0 - (100 / (1 + avgGain/avgLoss))
	return relativeStrengthIndex, avgGain, avgLoss, err
}

func SimpleMovingAverage(vals []float64) (simpleMovingAverage float64, err error) {
	var sum = float64(0.0)
	err = nil
	if len(vals) == 0 {
		err = errors.New("Invalid array length of zero")
		return 0.00, err
	}
	for _, price := range vals {
		sum += price
	}
	simpleMovingAverage = sum / float64(len(vals))
	if simpleMovingAverage < 0.0 {
		err = errors.New("Simple Moving average is below zero. Expected value greater than or equal to zero.")
	}
	return simpleMovingAverage, err
}

func main() {
	defer os.Exit(0)
	testPriceArray := []float64{3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0, 11.0, 12.0, 13.0, 14.0, 15.0}
	testRSIArray := []float64{-2.77, 4.79, 8.6, -0.18, 6.02, 1.23, -16.7, 9.64, 8.68, -0.8, 4.88, -1.83, -3.96, -9.09, -7.79, 2.3, -1.94, -7.72, 7.67, -9.48, 4.06, -1.22, 1, 4.91, 7.96, -6.99, 7.09, 4.92, -5.61}
	sma, err := SimpleMovingAverage(testPriceArray)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Simple Moving Average: %.2f\n", sma)

	//calculate first RSI value using the
	rsi, prevAvgGain, prevAvgLoss, err := RelativeStrengthIndex(testRSIArray[0:RSI_PERIODS], RSI_PERIODS, 0.0, 0.0)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Relative Strength Index: %.2f\n", rsi)
	fmt.Printf("Previous Average Gain: %.2f\n", prevAvgGain)
	fmt.Printf("Previous Average Loss: %.2f\n", prevAvgLoss)

	rsi2, prevAvgGain2, prevAvgLoss2, err := RelativeStrengthIndex(testRSIArray[0:RSI_PERIODS+1], RSI_PERIODS, prevAvgGain, prevAvgLoss)
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Relative Strength Index 2: %.2f\nPrevious Average Gain 2: %.2f\nPrevious Average Loss 2: %.2f\n", rsi2, prevAvgGain2, prevAvgLoss2)

	fmt.Printf("==== Blank Lines Denoting Start of Run through RSI Test Array ====\n\n\n\n")

	for index, _ := range testRSIArray {
		rsi, prevAvgGain, prevAvgLoss, err = RelativeStrengthIndex(testRSIArray[0:index+1], RSI_PERIODS, prevAvgGain, prevAvgLoss)
		fmt.Printf("RSI for index %v is %.2f\n", index, rsi)
	}
}
