package marketfeed

//Various stats about the requested pair
import (
	//"github.com/bitfinexcom/bitfinex-api-go"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	//"os"
)

type Pair []struct {
	Period json.Number `json:"period"`
	Volume json.Number `json:"volume"`
}

func PairStats(symbol string) (p Pair, err error) {
	req, err := http.NewRequest("GET", "https://api.bitfinex.com/v1/stats/"+symbol, nil)

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
	log.Println(string(body))

	err = json.Unmarshal(body, &p)

	if err != nil {
		log.Println("Exception occured at marshaling pair struct data", err)
		return Pair{}, err
	}

	return p, nil
}
