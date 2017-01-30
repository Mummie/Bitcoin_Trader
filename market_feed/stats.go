package market_feed

//Various stats about the requested pair
import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"log"
	//"os"
)

type Pair struct {
	Period json.Number `json:"period"`
	Volume json.Number `json:"volume"`
}

func PairStats(symbol string) (p *Pair, err error) {
	res, err := getJSONData("https://api.bitfinex.com/v1/stats/" + symbol)

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(res, &p)

	if err != nil {
		log.Println("Exception occured at unmarshaling pair struct data", err)
		return nil, err
	}

	return p, nil
}
