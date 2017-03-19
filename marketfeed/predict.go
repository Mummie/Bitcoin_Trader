package marketfeed

/*
Draw a square on the ground, then inscribe a circle within it.
Uniformly scatter some objects of uniform size (grains of rice or sand) over the square.
Count the number of objects inside the circle and the total number of objects.
The ratio of the two counts is an estimate of the ratio of the two areas, which is π/4. Multiply the result by 4 to estimate π.*/

/*eliminate survivorship bias
test for statistical significance
make sure your standard deviations aren't too large
back-test in different exchanges
back-test for at least ten years
account for trading commissions/fees
account for liquidity issues
paper trade*/
import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/Knetic/govaluate"
)

type AverageMarketPrice struct {
	Average float64
	Rate    float64
}

//ARIMA stands for Autoregressive Integrated Moving Average. It's a type of time series model which outputs a prediction and prediction interval given a time series data input. Mathematically, an ARIMA(p,d,q) looks like:
//yt=(∑i=1pαiyt−i)+(∑i=1qβiϵt−i)+ϵ
//Intuitively, this model predicts a y value at time t, given its near correlations to its past terms (until lag p) and its near correlations to its past residuals (until lag q).
// PredictBitcoinVolatility will predict the volatility by the called tick data at given time
func (tick *TickerData) PredictBitcoinVolatility() (float64, error) {
	var difficulty float64
	res, err := getJSONData("https://blockchain.info/q/getdifficulty")
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(res, &difficulty)

	if err != nil {
		return 0, err
	}

	expression, err := govaluate.NewEvaluableExpression("(Price) / (diff * 2^32) / 3600")
	if err != nil {
		return 0, err
	}

	parameters := make(map[string]interface{}, 8)
	parameters["Price"] = tick.LastPrice
	parameters["diff"] = difficulty

	prediction, err := expression.Evaluate(parameters)
	if err != nil {
		return 0, err
	}

	return prediction.(float64), nil

}

// CalculateOrderBookRegression will calculate order book regression where vbid is total volume people are willing to buy in the top x orders and vask is the total volume people are willing to sell in the top x orders based on current order book data
func (tick *TickerData) CalculateOrderBookRegression() float64 {
	expression, _ := govaluate.NewEvaluableExpression("(vbid - vask) / (vbid + vask)")
	parameters := make(map[string]interface{}, 8)
	parameters["vbid"] = tick.Bid
	parameters["vask"] = tick.Ask
	r, _ := expression.Evaluate(parameters)
	return r.(float64)
}

// CalculateMarketAveragePrice will take an amount of BTC to find and return the average market price for by the specified rate argument
func CalculateMarketAveragePrice(amount float64, rate float64) (a AverageMarketPrice, err error) {
	baseURL := "https://api.bitfinex.com/v2/calc/trade/avg"
	params := url.Values{}
	params.Add("symbol", "tBTCUSD")
	params.Add("amount", FloatToString(amount))
	params.Add("rate_limit", FloatToString(rate))

	finalURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	resp, err := postJSONData(finalURL, nil)
	if err != nil {
		return AverageMarketPrice{}, err
	}

	data := make([]float64, 2)
	err = json.Unmarshal(resp, &data)
	if err != nil {
		return AverageMarketPrice{}, err
	}
	a.Average = data[0]
	a.Rate = data[1]
	return a, nil
}

// FloatToString will convert a type float64 to a string with precision point 6
func FloatToString(input_num float64) string {
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}

// The trading strategy is very
// simple: at each time, we either maintain position
// of +1 Bitcoin, 0 Bitcoin or −1 Bitcoin. At each
// time instance, we predict the average price movement
// over the 10 seconds interval, say ∆p, using Bayesian
// regression (precise details explained below) - if ∆p > t,
// a threshold, then we buy a bitcoin if current bitcoin
// position is ≤ 0; if ∆p < −t, then we sell a bitcoin if
// current position is ≥ 0; else do nothing. The choice
// of time steps when we make trading decisions as
// mentioned above are chosen carefully by looking at
// the recent trends

// tell program to assume Gaussian distribution. dont program parameters but leave them unknown and computer afterwards
// get sample vector and process. probability density function of Gaussian distribution
// do the math, where u will be maximum likelihood for the mean and o will be maximum likelihood for standard deviation
