package marketfeed

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
	//"os"
)

// BTC/day = [(BitcoinPrice) / (difficulty * 2^32) / 3600]
// ex. let o = 24 * 2^32 / 3600

// @todo: make historic function to download csv data, parse data and return as slice
// Trade data is available as CSV, delayed by approx. 15 minutes. It will return the 2000 most recent trades.
//
// http://api.bitcoincharts.com/v1/trades.csv?symbol=SYMBOL[&start=UNIXTIME]
//
// returns CSV:
//
// unixtime,price,amount
// You can use the start parameter to specify an earlier unix timestamp in order to retrieve older data.

// @todo create function that will return real time data of blockchain. find be used to find patterns, transactions, amount of bitcoin mined, average mined and hash rate
// create channel to receive info from connection
// run goroutine for below url to dial
//dial with gorilla and loop reading data, send each response to channel
// https://blockchain.info/api/api_websocket

// https://blockchain.info/api/blockch

//@todo create ohShit func to stop all goroutines by closing created quit channel

// GET gets tick data or symbol
// change func to be dynamic for making api request to url and return json response []byte

type TickerData struct {
	Mid       float64 `json:"mid,string"`
	Bid       float64 `json:"bid,string"`
	Ask       float64 `json:"ask,string"`
	LastPrice float64 `json:"last_price,string"`
	Low       float64 `json:"low,string"`
	High      float64 `json:"high,string"`
	Volume    float64 `json:"volume,string"`
	Timestamp float64 `json:"timestamp,string"`
}

type Trade struct {
	Timestamp int64   `json:"timestamp"`
	Tid       int64   `json:"tid"`
	Price     float64 `json:"price,string"`
	Amount    float64 `json:"amount,string"`
	Exchange  string  `json:"exchange"`
	Type      string  `json:"type"`
}

func getJSONData(url string) (body []byte, err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := netClient.Get(url)

	if err != nil {
		return nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return body, nil
}

// GetTickData gets tick data or symbol
func GetTickData(symbol string) (tick *TickerData, err error) {

	res, err := getJSONData("https://api.bitfinex.com/v1/pubticker/" + symbol)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &tick)

	if err != nil {
		return nil, err
	}

	return tick, nil

}

// GetTradesData will return a slice of Trade data by the given day
func GetTradesData(symbol string) (trade []Trade, err error) {

	res, err := getJSONData("https://api.bitfinex.com/v1/trades/btcusd" + symbol)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &trade)

	if err != nil {
		return nil, err
	}

	return trade, nil
}

func RunTicker(symbol string) (tick *TickerData, err error) {
	timeout := time.After(30 * time.Second)
	processing := time.Tick(500 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return nil, errors.New("timed out")
		// Got a tick, we should check on doSomething()
		case <-processing:
			tick, err = GetTickData(symbol)
			if err != nil {
				return nil, err
			}
			log.Print("Tick Returns:  ")
			return tick, nil

		}
	}
}

const WorkerCount = 10

func compute() int {
	// Some input data to operate on.
	// Each worker gets an equal share to work on.
	data := make([]int, WorkerCount*10)

	for i := range data {
		data[i] = i
	}

	// Sum all the entries.
	result := sum(data)

	fmt.Printf("Sum: %d\n", result)
	return result
}

// sum adds up the numbers in the given list, by having the operation delegated
// to workers operating in parallel on sub-slices of the input data.
func sum(data []int) int {
	result := make(chan int)

	// The WaitGroup will track completion of all our workers.
	wg := &sync.WaitGroup{}
	wg.Add(WorkerCount)

	// Divide the work up over the number of workers.
	chunkSize := len(data) / WorkerCount

	// Spawn workers.
	for i := 0; i < WorkerCount; i++ {
		offset := i * chunkSize
		go worker(result, data[offset:offset+chunkSize], wg)
	}

	// Wait for all workers to finish, before returning the result.
	go func() {
		wg.Wait()
		close(result)
	}()

	// Accumulate results from workers.
	sum := 0
	for value := range result {
		sum += value
	}

	return sum
}

// worker sums up the numbers in the given list.
func worker(result chan int, data []int, wg *sync.WaitGroup) {
	defer wg.Done()

	sum := 0
	for _, v := range data {
		sum += v
	}

	result <- sum
}
