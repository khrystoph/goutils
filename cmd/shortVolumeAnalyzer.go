package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	inputURL string = "https://cdn.finra.org/equity/regsho/daily/CNMSshvol20220118.txt"
)

type dailyShortData struct {
	symbol           string    `json:symbol`
	tradingDate      time.Time `json:timestamp`
	tradingDayString string    `json:dateString`
	shortVol         float64   `json:shortVolume`
	shortExemptVol   float64   `json:shortExemptVolume`
	totalVolume      float64   `json:totalVolume`
	market           []string  `json:market`
}

type shortData struct {
	stockSymbol     string         `json:symbol`
	shortVolumeData dailyShortData `json:finraShortVolumeData`
}

func fetchStockVolumeData(url string) (finraShortData shortData, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return shortData{}, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return shortData{}, err
	}

	responseString := strings.Split(string(responseBody), "\n")
	for _, lines := range responseString {
		fmt.Println(lines)
	}
	return shortData{}, err
}

func main() {
	finraData, err := fetchStockVolumeData(inputURL)
	if err != nil {
		fmt.Printf("Error retrieving stock volume data from finra: %v\n", err)
	}
	fmt.Printf("%v", finraData)
}
