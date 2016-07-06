package main

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"github.com/Mummie/Bitcoin_Trader/market_feed/trade"
	"log"
	//"os"
)

type Config struct {
	API_KEY    string
	API_SECRET string
}

func main() {

	log.Println("Starting Bitcoin Trader...")

	tick := trade.Ticker("BTCUSD")

	log.Println(tick)

	pair, err := trade.PairStats("BTCUSD")

	if err != nil {
		log.Fatal(err)
	}

	log.Println(pair)

}
