package marketfeed

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/DannyBen/quandl"
	"github.com/cavaliercoder/grab"
	"github.com/gocarina/gocsv"
)

// http://api.bitcoincharts.com/v1/trades.csv?symbol=mtgoxUSD

//HistoricTradeHistory stores a map with key name being the exchange and a struct of that exchanges trading data
type HistoricTradeHistory struct {
	Trades map[string][]HistoricTradeData
}

//HistoricTradeData stores 2000 most recent trades by exchange. Data is unmarshaled from a csv to struct
type HistoricTradeData struct {
	Time   time.Time `csv:"unixtime"`
	Price  float64   `csv:"price"`
	Amount float64   `csv:"amount"`
}

type TimeSeriesExchangeRate struct {
	Dataset struct {
		ID                  int         `json:"id"`
		DatasetCode         string      `json:"dataset_code"`
		DatabaseCode        string      `json:"database_code"`
		Name                string      `json:"name"`
		Description         string      `json:"description"`
		RefreshedAt         time.Time   `json:"refreshed_at"`
		NewestAvailableDate string      `json:"newest_available_date"`
		OldestAvailableDate string      `json:"oldest_available_date"`
		ColumnNames         []string    `json:"column_names"`
		Frequency           string      `json:"frequency"`
		Type                string      `json:"type"`
		Premium             bool        `json:"premium"`
		Limit               interface{} `json:"limit"`
		Transform           interface{} `json:"transform"`
		ColumnIndex         interface{} `json:"column_index"`
		StartDate           string      `json:"start_date"`
		EndDate             string      `json:"end_date"`
		Data                []struct {
			Date           string  `json:"0"`
			Open           float64 `json:"1"`
			High           float64 `json:"2"`
			Low            float64 `json:"3"`
			Close          float64 `json:"4"`
			VolumeBTC      float64 `json:"5"`
			VolumeUSD      float64 `json:"6"`
			Weighted_Price float64 `json:"7"`
		} `json:"data"`
		Collapse   interface{} `json:"collapse"`
		Order      interface{} `json:"order"`
		DatabaseID int         `json:"database_id"`
	} `json:"dataset"`
}

//GetHistoricTrades will concurrently download the latest 2000 trades from array of exchanges, extract and unmarshal file content into struct
// e.g Trades -> m := map["bitfenixUSD"][]HistoricTradeData
func GetHistoricTrades(exchanges ...string) (trades []*HistoricTradeData, err error) {

	var b bytes.Buffer
	var urls []string
	for _, exchange := range exchanges {
		b.Reset()
		b.Write([]byte("http://api.bitcoincharts.com/v1/csv/"))
		b.WriteString(exchange)
		b.WriteString(".csv.gz")
		urls = append(urls, b.String())
	}

	err = DownloadToFile(urls)
	if err != nil {
		return nil, err
	}

	for _, exchange := range exchanges {
		b.Reset()
		b.Write([]byte(exchange))
		b.WriteString(".csv.gz")
		err = Unzip(b.String(), "historic_trades")
		if err != nil {
			return nil, err
		}

		if _, err := os.Stat(b.String()); err == nil {
			err = os.Remove(b.String())
			if err != nil {
				return nil, err
			}
		}
	}

	for _, exchange := range exchanges {
		b.Reset()
		b.WriteString("historic_trades/")
		b.WriteString(".")
		b.Write([]byte(exchange))
		b.WriteString(".csv")
		tradesFile, err := os.OpenFile(b.String(), os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, err
		}
		defer tradesFile.Close()

		if err := gocsv.UnmarshalFile(tradesFile, &trades); err != nil { // Load trades from file
			return nil, err
		}
	}

	return trades, nil
}

//DownloadToFile will download historical csv data for one or more exchanges passed and download in parallel
func DownloadToFile(urls []string) error {

	// create a custom client
	client := grab.NewClient()

	// create request for each URL given on the command line
	reqs := make([]*grab.Request, 0)
	for _, url := range urls {
		log.Println(url)
		req, err := grab.NewRequest(url)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}

		req.HTTPRequest.Method = "GET"
		req.HTTPRequest.Header.Add("Accept-Encoding", "gzip")

		reqs = append(reqs, req)
	}

	// start file downloads
	fmt.Printf("Downloading %d files...\n", len(reqs))
	respch := client.DoBatch(len(reqs), reqs...)

	// start a ticker to update progress every 200ms
	t := time.NewTicker(200 * time.Millisecond)

	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// clear lines
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}

			// update completed downloads
			for i, resp := range responses {
				if resp != nil && resp.IsComplete() {
					// print final result
					if resp.Error != nil {
						fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Error)
					} else {

						if resp.Size < 1024 {
							_, err := fmt.Fprintf(os.Stderr, "Expected filesize to be greater than 1kb, got %v", resp.Size)
							return err
						}

						fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))

					}
					responses[i] = nil
					completed++
				}
			}

			// update downloads in progress
			inProgress = 0
			for _, resp := range responses {
				if resp != nil {
					inProgress++
					fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
				}
			}
		}
	}

	defer t.Stop()

	return nil
}

//Unzip will take a src filename, create a gzip reader and extract to dest argument
func Unzip(src, dest string) (err error) {
	file, err := os.Open(src)
	if err != nil {
		return err
	}

	defer file.Close()
	log.Println("Extracting Gzip...")
	r, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := os.Stat(dest); os.IsNotExist(err) {
		os.MkdirAll(dest, 0755)
	}

	path := filepath.Join(dest, r.Name)
	writer, err := os.Create(path)
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = io.Copy(writer, r)
	log.Println("Copied zip contents to csv file")

	return err
}

func GetHistoricBitcoinVolatility() (v []VolatilityEstimate, err error) {
	res, err := getJSONData("https://btcvol.info/all")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &v)

	if err != nil {
		return nil, err
	}

	return v, nil
}

func GetHistoricBitfenixUSDExchangeRates() (b *quandl.SymbolResponse, err error) {
	b, err = quandl.GetSymbol("BITFENIX/BTCUSD", nil)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GetHistoricBitstampUSDExchangeRates() (b *quandl.SymbolResponse, err error) {
	b, err = quandl.GetSymbol("BITSTAMP/USD", nil)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GetHistoricBitstampExchangeRate() (b *quandl.SymbolResponse, err error) {
	b, err = quandl.GetSymbol("BCHARTS/BITSTAMPUSD", nil)
	if err != nil {
		return nil, err
	}

	log.Println(b)

	return b, nil
}

// >- below is a func for single downloading a file concurrently by downloading pieces
// func Test() {
// 	res, _ := http.Head("http://api.bitcoincharts.com/v1/csv/bitfenixUSD.csv.gz") // 187 MB file of random numbers per line
// 	maps := res.Header
// 	length, _ := strconv.Atoi(maps["Content-Length"][0]) // Get the content length from the header request
// 	limit := 5                                           // 5 Go-routines for the process so each downloads 18.7MB
// 	len_sub := length / limit                            // Bytes for each Go-routine
// 	diff := length % limit                               // Get the remaining for the last request
// 	body := make([]string, 11)                           // Make up a temporary array to hold the data to be written to the file
// 	for i := 0; i < limit; i++ {
// 		wg.Add(1)

// 		min := len_sub * i       // Min range
// 		max := len_sub * (i + 1) // Max range

// 		if i == limit-1 {
// 			max += diff // Add the remaining bytes in the last request
// 		}

// 		go func(min int, max int, i int) {
// 			client := &http.Client{}
// 			req, _ := http.NewRequest("GET", "http://api.bitcoincharts.com/v1/csv/bitfenixUSD.csv.gz", nil)
// 			range_header := "bytes=" + strconv.Itoa(min) + "-" + strconv.Itoa(max-1) // Add the data for the Range header of the form "bytes=0-100"
// 			req.Header.Add("Range", range_header)
// 			resp, _ := client.Do(req)
// 			defer resp.Body.Close()
// 			reader, _ := ioutil.ReadAll(resp.Body)
// 			body[i] = string(reader)
// 			ioutil.WriteFile(strconv.Itoa(i), []byte(string(body[i])), 0x777) // Write to the file i as a byte array
// 			wg.Done()
// 		}(min, max, i)
// 	}
// 	wg.Wait()
// }
