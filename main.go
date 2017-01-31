package main

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"github.com/Mummie/Bitcoin_Trader/marketfeed"
	"log"
	//"os"
)

type Config struct {
	API_KEY    string
	API_SECRET string
}

func main() {

	log.Println("Starting Bitcoin Trader...")

		tick, err := marketfeed.RunTicker("BTCUSD")

		if err != nil {
			log.Println(err)
		}
		log.Println(tick)
	

	pair, err := marketfeed.PairStats("BTCUSD")

	if err != nil {
		log.Fatal(err)
	}

	log.Println(pair)

}
