package marketfeed

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	//"os"
)

// BTC/day = [(BitcoinPrice) / (difficulty * 2^32) / 3600]
// ex. let o = 24 * 2^32 / 3600

type TickerData struct {
	Mid       json.Number `json:"mid"`
	Bid       json.Number `json:"bid"`
	Ask       json.Number `json:"ask"`
	LastPrice json.Number `json:"last_price"`
	Low       json.Number `json:"low"`
	High      json.Number `json:"high"`
	Volume    json.Number `json:"volume"`
	Timestamp json.Number `json:"timestamp"`
}

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

func SendBlockChainData() {
	var addr = flag.String("addr", "localhost:8080", "http service address")
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

func getJSONData(url string) (body []byte, err error) {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}

func GetTickDataBySymbol(symbol string) (tick TickerData, err error) {
	res, err := getJSONData("https://api.bitfinex.com/v1/pubticker/" + symbol)
	if err != nil {
		return TickerData{}, err
	}

	err = json.Unmarshal(res, &tick)

	if err != nil {
		return TickerData{}, err
	}

	return tick, nil

}

func RunTicker(symbol string) (tick TickerData, err error) {
	timeout := time.After(30 * time.Second)
	processing := time.Tick(500 * time.Millisecond)
	// Keep trying until we're timed out or got a result or got an error
	for {
		select {
		// Got a timeout! fail with a timeout error
		case <-timeout:
			return TickerData{}, errors.New("timed out")
		// Got a tick, we should check on doSomething()
		case <-processing:
			tick, err = GetTickDataBySymbol(symbol)
			if err != nil {
				return TickerData{}, err
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
