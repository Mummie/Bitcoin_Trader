package marketfeed

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/bitfinexcom/bitfinex-api-go"

	"time"
	//"os"
)

// @todo create function that will return real time data of blockchain. find be used to find patterns, transactions, amount of bitcoin mined, average mined and hash rate
// create channel to receive info from connection
// run goroutine for below url to dial
//dial with gorilla and loop reading data, send each response to channel
// https://blockchain.info/api/api_websocket

// https://blockchain.info/api/blockch

//@todo create ohShit func to stop all goroutines by closing created quit channel

//https://api.blockchain.info/stats

//https://blockchain.info/blocks/$time_in_milliseconds?format=json

//Blocks for one day: https://blockchain.info/blocks/$time_in_milliseconds?format=json
//Blocks for specific pool: https://blockchain.info/blocks/$pool_name?format=json

type Timestamp time.Time

type TickerData struct {
	Mid       float64 `json:"mid,string"`
	Bid       float64 `json:"bid,string"`
	Ask       float64 `json:"ask,string"`
	LastPrice float64 `json:"last_price,string"`
	Low       float64 `json:"low,string"`
	High      float64 `json:"high,string"`
	Volume    float64 `json:"volume,string"`
	// Timestamp Timestamp `json:"timestamp,string"`
	Predict float64 `json:"-"`
}

type Trade struct {
	Timestamp int64   `json:"timestamp"`
	Tid       int64   `json:"tid"`
	Price     float64 `json:"price,string"`
	Amount    float64 `json:"amount,string"`
	Exchange  string  `json:"exchange"`
	Type      string  `json:"type"`
}

type BTCChinaTick struct {
	Ticker struct {
		High   float64 `json:"high,string"`
		Low    float64 `json:"low,string"`
		Buy    float64 `json:"buy,string"`
		Sell   float64 `json:"sell,string"`
		Last   float64 `json:"last,string"`
		Volume float64 `json:"vol,string"`
		//	Time      time.Time `json:"date"`
		Vwap      float64 `json:"vwap,string"`
		PrevClose float64 `json:"prev_close,string"`
		Open      float64 `json:"open,string"`
	} `json:"ticker"`
}

type BiddingDataList struct {
	HighestBid float64
	BidDiff    float64
	Time       string
}

type TickerSocket struct {
	Bid             float64 `json:"bid,string"`
	BidSize         float64 `json:"bidsize,string"`
	Ask             float64 `json:"ask,string"`
	AskSize         float64 `json:"asksize,string"`
	DailyChange     float64 `json:"dailychange,string"`
	DailyChangePerc float64 `json:"dailychangeperc,string"`
	LastPrice       float64 `json:"last_price,string"`
	Low             float64 `json:"low,string"`
	High            float64 `json:"high,string"`
	Volume          float64 `json:"volume,string"`
}

type WebsocketChannel struct {
	Event     string `json:"subscribe"`
	Channel   string `json:"channel"`
	Error     string `json:"msg"`
	ErrorCode int64  `json:"code"`
}

func getJSONData(url string) (body []byte, err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 15,
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

func postJSONData(url string, r io.Reader) (body []byte, err error) {
	var netClient = &http.Client{
		Timeout: time.Second * 15,
	}

	resp, err := netClient.Post(url, "application/json", r)

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

	res, err := getJSONData("https://api.bitfinex.com/v1/trades/" + symbol)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &trade)

	if err != nil {
		return nil, err
	}

	return trade, nil
}

func GetBTCChinaTickData() (t *BTCChinaTick, err error) {
	res, err := getJSONData("https://data.btcchina.com/data/ticker")
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &t)
	if err != nil {
		return nil, err
	}

	return t, nil
}

//RunTicker will gather tick data by the time tick defined in processing and store the data inside a slice when timedout is reached
func RunTicker(symbol string) (tickData []*TickerData, err error) {
	timeout := time.After(60 * time.Second)
	processing := time.Tick(10 * time.Microsecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			if len(tickData) <= 0 {
				return nil, errors.New("Timed out with empty tick response")
			}
			return tickData, nil
		// Got a tick, we should check on doSomething()
		case <-processing:
			tick, err := GetTickData(symbol)
			if err != nil {
				return nil, err
			}

			predict, err := tick.PredictBitcoinVolatility()
			if err != nil {
				return nil, err
			}
			tick.Predict = predict

			r := tick.CalculateOrderBookRegression()

			avg, err := CalculateMarketAveragePrice(tick.LastPrice, tick.Volume)
			if err != nil {
				return nil, err
			}

			fmt.Printf("\rBid: %+v Prediction: %+v Regression: %+v Average Market Price: %f", tick.Bid, tick.Predict, r, avg.Average)

			tickData = append(tickData, tick)
		}
	}
}

func RunTickerSocket() chan []float64 {
	c := bitfinex.NewClient()

	// in case your proxy is using a non valid certificate set to TRUE
	c.WebSocketTLSSkipVerify = false

	err := c.WebSocket.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer c.WebSocket.Close()
	//
	// book_btcusd_chan := make(chan []float64)
	// book_ltcusd_chan := make(chan []float64)
	// trades_chan := make(chan []float64)
	ticker_chan := make(chan []float64)

	// c.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.BTCUSD, book_btcusd_chan)
	// c.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.LTCUSD, book_ltcusd_chan)
	// c.WebSocket.AddSubscribe(bitfinex.CHAN_TRADE, bitfinex.BTCUSD, trades_chan)
	c.WebSocket.AddSubscribe(bitfinex.CHAN_TICKER, bitfinex.BTCUSD, ticker_chan)

	// go listen(book_btcusd_chan, "BOOK BTCUSD:")
	// go listen(book_ltcusd_chan, "BOOK LTCUSD:")
	// go listen(trades_chan, "TRADES BTCUSD:")

	err = c.WebSocket.Subscribe()
	if err != nil {
		log.Fatal(err)
	}
	return ticker_chan
}

func listen(in chan []float64, message string) {
	for {
		msg := <-in
		fmt.Println(message, msg)
	}
}

// func (t *Timestamp) MarshalJSON() ([]byte, error) {
// 	ts := time.Time(*t).Unix()
// 	stamp := fmt.Sprint(ts)
//
// 	return []byte(stamp), nil
// }
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	unix := strings.Split(string(b), ".")
	ts, err := strconv.ParseInt(unix[0], 10, 64)
	if err != nil {
		return err
	}
	*t = Timestamp(time.Unix(ts, 0))
	return nil
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
