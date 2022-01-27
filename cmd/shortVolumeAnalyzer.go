package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_TIME_FORMAT     = "2006-01-02"
	DEFAULT_TIME_URL_STRING = "20060102"
)

var (
	startDate      string
	endDate        string
	singleDate     string
	trailingDays   int64
	inputURLPrefix string = "https://cdn.finra.org/equity/regsho/daily/CNMSshvol"
	inputURLPost   string = ".txt"
)

func init() {
	flag.StringVar(&startDate, "start", "2021-01-01",
		"enter a date string in yyyy-mm-dd format for the day to start retrieving records from. Default: 2021-01-01.")
	flag.StringVar(&startDate, "s", "2021-01-01",
		"enter a date string in yyyy-mm-dd format for the day to start retrieving records from. Default: 2021-01-01.")
	flag.StringVar(&endDate, "end", time.Now().Format("2006-01-02"),
		"enter a date string in yyyy-mm-dd format to stop retrieving records from. Default: Today's date.")
	flag.StringVar(&endDate, "e", time.Now().Format("2006-01-02"),
		"enter a date string in yyyy-mm-dd format to stop retrieving records from. Default: Today's date.")
	flag.StringVar(&singleDate, "one-day", time.Now().Format("2006-01-02"),
		"enter a single date string in yyyy-mm-dd format to retrieve a single day of Finra Data. Default: Today's date.")
	flag.StringVar(&singleDate, "o", time.Now().Format("2006-01-02"),
		"enter a single date string in yyyy-mm-dd format to retrieve a single day of Finra Data. Default: Today's date.")
}

type dailyShortData struct {
	Symbol                string    `json:symbol`                //provided
	TradingDate           time.Time `json:timestamp`             //provided
	TradingDayString      string    `json:dateString`            //provided
	ShortVol              float64   `json:shortVolume`           //provided
	ShortExemptVol        float64   `json:shortExemptVol`        //provided
	TotalVolume           float64   `json:totalVol`              //provided
	Market                []string  `json:market`                //provided
	ShortVolPercent       float64   `json:shortVolPercent`       //calculated
	ShortExemptVolPercent float64   `json:shortExemptVolPercent` //calculated
	BuyVolPercent         float64   `json:buyVolPercent`         //calculated
	BuyVol                float64   `json:buyVol`                //calculated
}

type shortData struct {
	StockSymbol                string         `json:symbol`
	TradingDate                string         `json:dateString`
	ShortVolumeData            dailyShortData `json:finraShortVolumeData`
	TotalShortVol              float64        `json:totalShortVol`
	TotalExemptShortVol        float64        `json:totalExemptShortVol`
	TotalBuyVol                float64        `json:totalBuyVol`
	TotalSharesShort           float64        `json:totalSharesShort`
	TotalVol                   float64        `json:totalVol`
	TotalBuyVolPercent         float64        `json:totalBuyVolPercent`
	TotalExemptShortVolPercent float64        `json:totalExemptShortVolPercent`
	TotalShortInterestPercent  float64        `json:totalShortInterestPercent`
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
				TradingDayString: tradeDate.Format(DEFAULT_TIME_FORMAT),
				ShortVol:         shortVolume,
				ShortExemptVol:   shortExemptVolume,
				TotalVolume:      totalTradeVolume,
				Market:           strings.Split(strings.Split(line, ",")[columnMapping["Market"]], ","),
			}
			symbolShortData.ShortVolPercent = (symbolShortData.TotalVolume - symbolShortData.ShortVol) / symbolShortData.TotalVolume * 100
			symbolShortData.ShortExemptVolPercent = (symbolShortData.TotalVolume - symbolShortData.ShortExemptVol) / symbolShortData.TotalVolume * 100
			symbolShortData.BuyVol = symbolShortData.TotalVolume - symbolShortData.ShortVol
			symbolShortData.BuyVolPercent = (symbolShortData.TotalVolume - symbolShortData.BuyVol) / symbolShortData.TotalVolume * 100
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

	flag.Parse()

	//determine whether or not to select single date of data, start and end date, or range of dates based on input flags.

	inputDate, err := time.Parse(DEFAULT_TIME_FORMAT, singleDate)
	if err != nil {
		fmt.Printf("Unable to parse single date input. Error message:\n%s", err)
		os.Exit(1)
	}
	inputDateString := inputDate.Format(DEFAULT_TIME_URL_STRING)

	inputURL := inputURLPrefix + inputDateString + inputURLPost
	finraData, err := fetchStockVolumeData(inputURL)
	if err != nil {
		fmt.Printf("Error retrieving stock volume data from finra: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Finra data returned:")
	for _, stockSymbol := range finraData {
		fmt.Printf("%s:\n", stockSymbol.StockSymbol)
		jsonData, err := json.MarshalIndent(stockSymbol, "", "    ")
		if err != nil {
			fmt.Printf("Unable to Marshal JSON data")
		}
		fmt.Printf(string(jsonData))
	}
}
