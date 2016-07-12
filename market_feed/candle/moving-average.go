package candle

import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
)

// get moving price
// last_price * number of coins purchased / coin trades
func GetMovingAverage(price string, trades string, totaltrades string) int {
	price = int(price)
	trades = int(trades)
	totaltrades = int(totaltrades)

	return price * trades / totaltrades
}
