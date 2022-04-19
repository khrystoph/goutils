package main

import (
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
	DEFAULT_TIME_FORMAT     string = "2006-01-02"
	DEFAULT_TIME_URL_STRING string = "20060102"
	HOURS_PER_DAY           int64  = 24
)

var (
	defaultDays     int64
	trailingDays    int64
	startDateString string
	endDateString   string
	inputURLPrefix  string = "https://cdn.finra.org/equity/regsho/daily/CNMSshvol"
	inputURLPost    string = ".txt"
)

func init() {
	flag.Int64Var(&defaultDays, "days", 30, "enter a number of days to go back and retrieve records from today's date (calendar days). Default: 30.")
	flag.Int64Var(&defaultDays, "d", 30, "enter a number of days to go back and retrieve records from today's date (calendar days). Default: 30.")
	flag.StringVar(&startDateString, "start", "NA", "enter a date in format: -start 'YYYY-mm-dd'. Default NA.")
	flag.StringVar(&endDateString, "end", "NA", "enter a date in format: -start 'YYYY-mm-dd'. Default NA.")
}

//DailyShortData contains the data for a single symbol for a single day of data.
type DailyShortData struct {
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

type ShortData struct {
	StockSymbol                string                    `json:symbol`
	ShortVolumeData            map[string]DailyShortData `json:finraShortVolumeData` //string representation of trade date in form yyyymmdd
	TotalShortVol              float64                   `json:totalShortVol`
	TotalExemptShortVol        float64                   `json:totalExemptShortVol`
	TotalBuyVol                float64                   `json:totalBuyVol`
	TotalSharesShort           float64                   `json:totalSharesShort`
	TotalVol                   float64                   `json:totalVol`
	TotalBuyVolPercent         float64                   `json:totalBuyVolPercent`
	TotalExemptShortVolPercent float64                   `json:totalExemptShortVolPercent`
	TotalShortInterestPercent  float64                   `json:totalShortInterestPercent`
	TotalSharesOutstanding     float64                   `json:totalSharesOutstanding`
	TotalFloat                 float64                   `json:totalFloat`
}

//instantiate a new array of size N to contain all the DailyShortData structs of specific size so that you can end up with
func NewDailyShortDataMap(numElems int) (dailyShortDataMapPtr *map[string]DailyShortData) {
	dailyShortDataMap := make(map[string]DailyShortData, numElems)
	return &dailyShortDataMap
}

func MakeShortDataMap() (shortDataMap *map[string]ShortData) {
	shortDataList := make(map[string]ShortData)
	return &shortDataList
}

func MakeNewShortData(symbol string, numElems int) (newShortDataElemPtr *ShortData) {
	newShortDataElemPtr = &ShortData{
		StockSymbol:     symbol,
		ShortVolumeData: *NewDailyShortDataMap(numElems),
	}
	return
}

//simple function to check if a flag was passed in at the command line
func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

//fetchStockVolumeData grabs the data from Finra and returns a struct of type shortData containing all
//the data in the finra daily volume txt web pages
func fetchStockVolumeData(url string, finraShortDataMap *map[string]ShortData) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("error retrieving responseBody")
		return err
	}
	respBodyBytes := strings.ReplaceAll(string(responseBody), ",", ";")
	respBodyBytes = strings.ReplaceAll(respBodyBytes, "|", ",")
	respBodyBytesSlice := strings.Split(respBodyBytes, "\n")
	respBodyBytes = strings.Join(respBodyBytesSlice[0:len(respBodyBytesSlice)-2], "\n")

	var columnMapping = make(map[string]int)
	for _, line := range strings.Split(respBodyBytes, "\n") {
		line = strings.Trim(line, "\r") //remove carriage returns, which were polluting the "Market" field in the line data

		//If there is not already a key mapping for the Columns, Create it if it's the first line of the csv-style formatting
		if strings.Contains(strings.ToLower(line), "symbol") {
			for index, key := range strings.Split(line, ",") {
				columnMapping[key] = index
			}
			fmt.Printf("column Mapping:\n%v\n", columnMapping)

			//Bulk of logic to create the struct is here.
		} else {
			tradeDate, err := time.Parse(DEFAULT_TIME_URL_STRING, strings.Split(line, ",")[columnMapping["Date"]])
			if err != nil {
				fmt.Printf("unable to parse trade date: %v\n", err)
				return err
			}
			shortVolume, err := strconv.ParseFloat(strings.Split(line, ",")[columnMapping["ShortVolume"]], 64)
			if err != nil {
				fmt.Printf("unable to parse Short Volume: %v\n", err)
				return err
			}
			shortExemptVolume, err := strconv.ParseFloat(strings.Split(line, ",")[columnMapping["ShortExemptVolume"]], 64)
			if err != nil {
				fmt.Printf("unable to parse Short Exempt Volume: %v\n", err)
				return err
			}
			totalTradeVolume, err := strconv.ParseFloat(strings.Split(line, ",")[columnMapping["TotalVolume"]], 64)
			if err != nil {
				fmt.Printf("unable to parse Total Trade Volume: %v\n", err)
				return err
			}
			marketArray := strings.Split((strings.Split(line, ",")[columnMapping["Market"]]), ";")
			symbolShortData := DailyShortData{
				Symbol:           strings.Split(line, ",")[columnMapping["Symbol"]],
				TradingDate:      tradeDate,
				TradingDayString: tradeDate.Format(DEFAULT_TIME_FORMAT),
				ShortVol:         shortVolume,
				ShortExemptVol:   shortExemptVolume,
				TotalVolume:      totalTradeVolume,
				Market:           marketArray,
			}
			symbolShortData.ShortVolPercent = (symbolShortData.ShortVol) / symbolShortData.TotalVolume * 100
			symbolShortData.ShortExemptVolPercent = (symbolShortData.ShortExemptVol) / symbolShortData.TotalVolume * 100
			symbolShortData.BuyVol = symbolShortData.TotalVolume - symbolShortData.ShortVol
			symbolShortData.BuyVolPercent = (symbolShortData.BuyVol) / symbolShortData.TotalVolume * 100

		}
	}
	return err
}

//worker function to coordinate fetching data and adding it to the map of tickers.
func worker(tickers *map[string]ShortData, start, end time.Time, deltaDays int64) (err error) {
	/*
	* Retrieve a day of data and store it in the map.
	* Handle page retrieve errors on days that don't trade
	* Tabulate total short shares outstanding. If a day exists where buy vol exceeds 50% of trading (ie. delta positive shares),
	* subtract those from the existing shares short.
	*
	* tickers: the pointer to the map of ticker symbols to track short interest data on.
	* deltaDays: number of days that are to be retrieved back on the calendar.
	* start: a Time var indicating the first trading day to retrieve.
	* end: a Time var indicating the last trading day to retrieve.
	*
	 */
	fmt.Printf("start: %v\n", start)
	url := inputURLPrefix + start.Format("20060102") + inputURLPost
	fmt.Printf("%v\n", url)
	fetchStockVolumeData(url, tickers)
	os.Exit(0)

	for i := 0; i < int(deltaDays); i++ {
		retrieveDate := start.Add(time.Duration(int64(i * int(time.Hour) * 24)))
		fmt.Printf("date of retrieval is: %v\n", retrieveDate.Format(DEFAULT_TIME_URL_STRING))
	}

	return nil
}

//Retrieve short vol data from FINRA, organize the data, and then present data based on highest short interest.
func main() {

	flag.Parse()
	var (
		startDate, endDate time.Time
		numDays            int64
		err                error
	)

	//Need to check for existence of both walkbaack days and start/end dates. If both exist, either error out or choose one to have preceedence over the other.

	//check if the walk-back of days is set correctly.
	if isFlagPassed("days") || isFlagPassed("d") {
		startDate = time.Now().Add(time.Duration(-1*defaultDays*24) * time.Hour)
		endDate = time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.UTC)
		numDays = defaultDays
	} else {
		fmt.Printf("startDateString: %s\n", startDateString)
		startDate, err = time.Parse("2006-01-02", startDateString)
		if err != nil {
			fmt.Printf("Error converting start Date into a date format. Required format: YYYY-mm-dd, where Y is year, m is month, and d is day. Error msg:\n %v", err)
			os.Exit(1)
		}
		fmt.Printf("startDate: %v\n", startDate)

		fmt.Printf("endDateString: %s\n", endDateString)
		endDate, err = time.Parse(DEFAULT_TIME_FORMAT, endDateString)
		if err != nil {
			fmt.Printf("Error converting end Date into a date format. Required format: YYYY-mm-dd, where Y is year, m is month, and d is day. Error msg:\n %v", err)
			os.Exit(1)
		}
		fmt.Printf("endDate: %v\n", endDate)
		numDays = int64(endDate.Sub(startDate).Hours()) / HOURS_PER_DAY
	}

	//create map of tickers
	tickerMap := make(map[string]ShortData, numDays)

	//kick off the function that will do most of the work.

	fmt.Printf("start date: %v\n", startDate)
	err = worker(&tickerMap, startDate, endDate, numDays)
	if err != nil {
		fmt.Printf("Error running main worker. Error is: %v", err)
		os.Exit(1)
	}

	fmt.Printf("Completed tasks. Exiting Cleanly.")
	return
}
