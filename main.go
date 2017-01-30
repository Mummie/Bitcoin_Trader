package main

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"log"

	"github.com/Bitcoin_Trader/market_feed"
	//"os"
)

type Config struct {
	API_KEY    string
	API_SECRET string
}

// consistently check for data races from websocket response and ticker data $ go run -race mysrc.go  // to run the source file

func main() {

	log.Println("Starting Bitcoin Trader...")

	for {

		tick, err := market_feed.RunTicker("BTCUSD")

		if err != nil {
			log.Fatal(err)
		}
		log.Println(tick)
	}

	pair, err := market_feed.PairStats("BTCUSD")

	if err != nil {
		log.Fatal(err)
	}

	log.Println(pair)

}
