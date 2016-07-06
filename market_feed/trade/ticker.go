package trade

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
)

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

// GET gets tick data or symbol
func Ticker(symbol string) string {

	req, err := http.NewRequest("GET", "https://api.bitfinex.com/v1/pubticker/"+symbol, nil)

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
