package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	inputURL string = "https://cdn.finra.org/equity/regsho/daily/CNMSshvol20220118.txt"
)

type dailyShortData struct {
	Symbol           string    `json:symbol`
	TradingDate      time.Time `json:timestamp`
	TradingDayString string    `json:dateString`
	ShortVol         float64   `json:shortVolume`
	ShortExemptVol   float64   `json:shortExemptVolume`
	TotalVolume      float64   `json:totalVolume`
	Market           []string  `json:market`
}

type shortData struct {
	StockSymbol     string         `json:symbol`
	ShortVolumeData dailyShortData `json:finraShortVolumeData`
}

//fetchStockVolumeData grabs the data from Finra and returns a struct of type shortData containing all
//the data in the finra daily volume txt web pages
func fetchStockVolumeData(url string) (finraShortData []shortData, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return []shortData{}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error retrieving responseBody")
		return []shortData{}, err
	}
	respBodyBytes := strings.ReplaceAll(string(responseBody), "|", ",")
	respBodyBytesSlice := strings.Split(respBodyBytes, "\n")
	respBodyBytes = strings.Join(respBodyBytesSlice[0:len(respBodyBytesSlice)-2], "\n")

	var columnMapping = make(map[string]int)
	for _, line := range strings.Split(respBodyBytes, "\n") {
		if strings.Contains(strings.ToLower(line), "symbol") {
			for index, key := range strings.Split(line, ",") {
				columnMapping[key] = index
			}
			fmt.Printf("column Mapping:\n%v", columnMapping)
		} else {
			tradeDate, err := time.Parse("20060102", strings.Split(line, ",")[columnMapping["Date"]])
			if err != nil {
				fmt.Printf("unable to parse trade date: %v", err)
				return []shortData{}, err
			}
			shortVolume, err := strconv.ParseFloat(strings.Split(line, ",")[columnMapping["ShortVolume"]], 64)
			if err != nil {
				fmt.Printf("unable to parse Short Volume: %v", err)
				return []shortData{}, err
			}
			shortExemptVolume, err := strconv.ParseFloat(strings.Split(line, ",")[columnMapping["ShortExemptVolume"]], 64)
			if err != nil {
				fmt.Printf("unable to parse Short Exempt Volume: %v", err)
				return []shortData{}, err
			}
			totalTradeVolume, err := strconv.ParseFloat(strings.Split(line, ",")[columnMapping["TotalVolume"]], 64)
			if err != nil {
				fmt.Printf("unable to parse Total Trade Volume: %v", err)
				return []shortData{}, err
			}
			symbolShortData := dailyShortData{
				Symbol:           strings.Split(line, ",")[columnMapping["Symbol"]],
				TradingDate:      tradeDate,
				TradingDayString: tradeDate.Format("2006-01-02"),
				ShortVol:         shortVolume,
				ShortExemptVol:   shortExemptVolume,
				TotalVolume:      totalTradeVolume,
				Market:           strings.Split(strings.Split(line, ",")[columnMapping["Market"]], ","),
			}
			finraShortData = append(finraShortData,
				shortData{
					StockSymbol:     symbolShortData.Symbol,
					ShortVolumeData: symbolShortData,
				})
		}
	}
	return finraShortData, err
}

func main() {
	finraData, err := fetchStockVolumeData(inputURL)
	if err != nil {
		fmt.Printf("Error retrieving stock volume data from finra: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Finra data returned:\n%v", finraData[0])
}
