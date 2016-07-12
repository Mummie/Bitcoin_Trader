package trade

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"
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
func GetTickData(symbol string) (tick TickerData, err error) {

	req, err := http.NewRequest("GET", "https://api.bitfinex.com/v1/pubticker/"+symbol, nil)

	if err != nil {

		log.Println(err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TickerData{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return TickerData{}, err
	}

	err = json.Unmarshal(body, &tick)

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
			tick, err = GetTickData(symbol)
			if err != nil {
				return TickerData{}, err
			}
			log.Print("Tick Returns:  ")
			return tick, nil

		}
	}
}
