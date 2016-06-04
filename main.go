package main

import (
	"fmt"
	//"github.com/bitfinexcom/bitfinex-api-go"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
	"time"
)

type Config struct {
	API_KEY    string
	API_SECRET string
}

type TickerData struct {
	Mid       string `json:"mid"`
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
	LastPrice string `json:"last_price"`
	Low       string `json:"low"`
	High      string `json:"high"`
	Volume    string `json:"volume"`
	Timestamp string `json:"timestamp"`
}

// GET gets tick data or symbol
func Ticker(symbol string) string {

	req, err := http.NewRequest("GET", "https://api.bitfinex.com/v1/pubticker/LTCUSD", nil)

	if err != nil {

		log.Fatal(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(body)
}

func main() {

	for i := 1; i < 1000; i++ {
		time.Sleep(3000 * time.Millisecond)
		fmt.Println(Ticker("BTCUSD"))
	}

}
