package marketfeed

//Various stats about the requested pair
import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"log"
	"time"
	//"os"
)

type Pair []struct {
	Period json.Number `json:"period"`
	Volume json.Number `json:"volume"`
}

type VolatilityEstimate struct {
	Date             time.Time `json:"Date"`
	Volatility       float64   `json:"Volatility"`
	Volatility60Days float64   `json:"Volatility60"`
}

type FundingBook struct {
	Bids []struct {
		Rate      float64   `json:"rate,string"`
		Amount    float64   `json:"amount,string"`
		Period    int64     `json:"period,string"`
		Timestamp time.Time `json:"timestamp,string"`
		Frr       string    `json:"frr,string"`
	} `json:"bids"`
	Asks []struct {
		Rate      float64   `json:"rate,string"`
		Amount    float64   `json:"amount,string"`
		Period    int64     `json:"period,string"`
		Timestamp time.Time `json:"timestamp,string"`
		Frr       string    `json:"frr,string"`
	} `json:"asks"`
}

type OrderBook struct {
	Bids []struct {
		Price     float64   `json:"price"`
		Amount    float64   `json:"amount"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"bids"`
	Asks []struct {
		Price     float64   `json:"price"`
		Amount    float64   `json:"amount"`
		Timestamp time.Time `json:"timestamp"`
	} `json:"asks"`
}

func PairStats(symbol string) (p Pair, err error) {
	res, err := getJSONData("https://api.bitfinex.com/v1/stats/" + symbol)

	if err != nil {
		return Pair{}, err
	}

	err = json.Unmarshal(res, &p)

	if err != nil {
		log.Println("Exception occured at marshaling pair struct data", err)
		return Pair{}, err
	}

	return p, nil
}

// GetFundingBook returns the full margin funding book
// for Frr returns Yes if the offer is at Flash Return Rate, no if fixed rate
func GetFundingBook(currency string) (f FundingBook, err error) {
	res, err := getJSONData("https://api.bitfinex.com/v1/lendbook/" + currency)

	if err != nil {
		return FundingBook{}, err
	}

	err = json.Unmarshal(res, &f)

	if err != nil {
		return FundingBook{}, err
	}

	return f, nil
}

// GetOrderBook will return the full order book for the given symbol off the bitfenix exchange
func GetOrderBook(symbol string) (o OrderBook, err error) {
	res, err := getJSONData("https://api.bitfinex.com/v1/book/" + symbol)

	if err != nil {
		return OrderBook{}, err
	}

	err = json.Unmarshal(res, &o)

	if err != nil {
		return OrderBook{}, err
	}

	return o, nil
}

func GetLatestBTCVolatilityEstimate() (v VolatilityEstimate, err error) {
	res, err := getJSONData("https://btcvol.info/latest")

	if err != nil {
		return VolatilityEstimate{}, err
	}

	err = json.Unmarshal(res, &v)

	if err != nil {
		return VolatilityEstimate{}, err
	}

	return v, nil
}
