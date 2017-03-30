package main

import
//"github.com/bitfinexcom/bitfinex-api-go"

(
	"fmt"
	"os"
	"time"

	"github.com/Bitcoin_Trader/marketfeed"
	"github.com/Bitcoin_Trader/server"
	"github.com/DannyBen/quandl"
)

//"os"

type Config struct {
	API_KEY    string
	API_SECRET string
}

// consistently check for data races from websocket response and ticker data $ go run -race mysrc.go  // to run the source file

// go server.RunServer()

// tick, err := marketfeed.RunTicker("BTCUSD")

// if err != nil {
// 	log.Fatal(err)
// }
// log.Println(tick)

// pair, err := marketfeed.PairStats("BTCUSD")

// if err != nil {
// 	log.Fatal(err)
// }

// log.Println(pair)

const (
	maxGoRoutines = 1
)

func main() {

	quandl.APIKey = os.Getenv("QUANDL_API_KEY")

	resc, errc := make(chan []*marketfeed.HistoricTradeData), make(chan error)

	tickChan, errT := make(chan []*marketfeed.TickerData), make(chan error)
	chinaTick := make(chan *marketfeed.BTCChinaTick)
	bitstampExchangeRateTimeSeries, bitstampErr := make(chan *quandl.SymbolResponse), make(chan error)
	blockchain, errBC := make(chan *marketfeed.BlockChain), make(chan error)
	// cbTickChan, errC := make(chan []*marketfeed.CoinbaseTicker), make(chan error)

	var highBids []marketfeed.BiddingDataList
	go server.RunHTTPServer()
	for tasks := 0; tasks < 3; tasks++ {
		// go marketfeed.RunTickerSocket()

		go func() {

			tickData, err := marketfeed.RunTicker("btcUSD")
			if err != nil {
				errT <- err
				return
			}

			chinatick, err := marketfeed.GetBTCChinaTickData()
			if err != nil {
				errT <- err
				return
			}

			tickChan <- tickData
			chinaTick <- chinatick
		}()
		// clear messages by linebreak every minute
		ticker := time.NewTicker(60 * time.Second)
		go func() {
			for {
				select {
				case <-ticker.C:
					fmt.Println()
				}
			}
		}()

		go func() {

			s, err := marketfeed.GetBlockchainStats()
			if err != nil {
				errBC <- err
				return
			}

			blockchain <- s
		}()

		go func() {
			t, err := marketfeed.GetHistoricBitstampExchangeRate()
			if err != nil {
				bitstampErr <- err
				return
			}

			bitstampExchangeRateTimeSeries <- t
		}()

		// go func() {
		// 	tickData, err := marketfeed.RunCoinBaseTicker(os.Getenv("CB_SECRET"), os.GetEnv("CB_KEY"))
		// 	if err != nil {
		// 		errC <- err
		// 		return
		// 	}

		// 	cbTickChan <- tickChan
		// }()

		// go func() {
		// 	trades, err := marketfeed.GetHistoricTrades("bitfinexUSD")
		// 	if err != nil {
		// 		errc <- err
		// 		return
		// 	}
		//
		// 	resc <- trades
		// }()

		for i := 0; i <= 1; i++ {
			select {
			case res := <-resc:
				fmt.Println(&res)
			case err := <-errc:
				fmt.Println("HISTORICAL DATA ERROR: ", err)
			case tick := <-chinaTick:
				fmt.Printf("BTC CHINA: %+v", tick)
			case stats := <-blockchain:
				fmt.Printf("Blockchain Stats: %+v", stats)
			// case t := <-cbTickChan:
			// 	fmt.Printf("Finished Coinbase Socket Process with %+v", t)
			// case err := <-errC:
			// 	fmt.Println("COINBASE TICKER ERR: ", err)
			case bitstampTimeSeries := <-bitstampExchangeRateTimeSeries:
				fmt.Printf("Bitstamp Exchange Rate Time Series: %+v ", bitstampTimeSeries)
			case err := <-bitstampErr:
				fmt.Printf("Bitstamp Time Series Error: %+v", err)
			case ticker := <-tickChan:
				var highestBid float64
				var h float64
				var hDiff float64
				for i, t := range ticker {
					if i == 1 {
						highestBid = t.Bid
						h = t.High
					}
					if t.Bid >= highestBid {
						highestBid = t.Bid
					}

					if t.High >= h {
						h = t.High
					}

					hDiff = highestBid - h
				}
				highBid := marketfeed.BiddingDataList{
					highestBid,
					hDiff,
					time.Now().Add(-1 * time.Minute).Format("15:04:05"),
				}
				highBids = append(highBids, highBid)

			case err := <-errT:
				fmt.Println("TICKER ERROR: ", err)
				continue
			}
		}
		time.Sleep(5 * time.Second)
		tasks++
	}
	for _, v := range highBids {
		fmt.Printf("key[%v]", v)
	}

}
