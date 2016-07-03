package main

import (
	//"github.com/Mummie/Bitcoin_Trader/market_feed/trade"
	//"github.com/labstack/echo"
	"github.com/bitfinexcom/bitfinex-api-go"
	//"github.com/labstack/echo/engine/standard"
	"log"
)

func main() {
	/*e := echo.New()
	e.GET("/ticker", trade.Ticker)
	e.Run(standard.New(":8008"))
	log.Println("Server running on 8008")
	*/
	api := bitfinex.NewClient()
	// Create new connection
	e := api.WebSocket.Connect()

	if e != nil {
		log.Fatal("Error connecting to web socket")

	}
	defer api.WebSocket.Close()

	book_btcusd_chan := make(chan []float64)
	book_ltcusd_chan := make(chan []float64)
	trades_chan := make(chan []float64)
	ticker_chan := make(chan []float64)

	api.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.BTCUSD, book_btcusd_chan)
	api.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.LTCUSD, book_ltcusd_chan)
	api.WebSocket.AddSubscribe(bitfinex.CHAN_TRADE, bitfinex.BTCUSD, trades_chan)
	api.WebSocket.AddSubscribe(bitfinex.CHAN_TICKER, bitfinex.BTCUSD, ticker_chan)
	go api.WebSocket.Subscribe()

	// after api client successfully connect to remote web socket
	// channel will reveive current payload as separate messages.
	// each channel will receive order book updates: [price, count, ±amount]
	for {
		select {
		case btcusd_msg := <-book_btcusd_chan:
			log.Println("BOOK BTCUSD:", btcusd_msg)
		case ltcusd_msg := <-book_ltcusd_chan:
			log.Println("BOOK LTCUSD:", ltcusd_msg)
		case trade_msg := <-trades_chan:
			log.Println("TRADES:", trade_msg)
		case ticker_msg := <-ticker_chan:
			log.Println("TICKER:", ticker_msg)
		}
	}
}
