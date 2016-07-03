package trade

import (
	"github.com/bitfinexcom/bitfinex-api-go"
	"github.com/labstack/echo"
	//"github.com/labstack/echo/engine/standard"
	//"io/ioutil"
	"log"
	//"net/http"
	//"os"
	"time"
)

type Config struct {
	API_KEY    string
	API_SECRET string
}

type TickerData struct {
	Mid       string    `json:"mid"`
	Bid       string    `json:"bid"`
	Ask       string    `json:"ask"`
	LastPrice string    `json:"last_price"`
	Low       string    `json:"low"`
	High      string    `json:"high"`
	Volume    string    `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type Subscribe struct {
	Book   chan []float64
	Trade  chan []float64
	Ticker chan []float64
}

// GET gets tick data or symbol
// needs to be marshaled into struct and returned, with error return ticker TickerData{}, err
func Ticker(c echo.Context) (btcusd_msg string, ltcusd_msg, err error) {

	api := bitfinex.NewClient()
	// Create new connection
	err = api.WebSocket.Connect()

	if err != nil {
		log.Fatal("Error connecting to web socket")
		return nil, err
	}
	defer api.WebSocket.Close()

	book_btcusd_chan := make(chan []float64)
	book_ltcusd_chan := make(chan []float64)
	trades_chan := make(chan []float64)
	ticker_chan := make(chan []float64)

	api.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.BTCUSD, book_btcusd_chan)
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
			return btcusd_msg, nil
		case trade_msg := <-trades_chan:
			log.Println("TRADES:", trade_msg)
		case ticker_msg := <-ticker_chan:
			log.Println("TICKER:", ticker_msg)
		}

		return nil

	}
}
