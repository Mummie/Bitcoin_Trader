package main

import
//"github.com/bitfinexcom/bitfinex-api-go"

(
	"fmt"
	"time"

	"github.com/Bitcoin_Trader/marketfeed"
	"github.com/Bitcoin_Trader/server"
	"github.com/Bitcoin_Trader/trade"
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
	maxGoRoutines = 8
)

func main() {

	resc, errc := make(chan []*marketfeed.HistoricTradeData), make(chan error)

	tickChan, errT := make(chan []*marketfeed.TickerData), make(chan error)
	chinaTick := make(chan *marketfeed.BTCChinaTick)
	blockchain, errBC := make(chan *marketfeed.BlockChain), make(chan error)
	var highBids []marketfeed.BiddingDataList
	go server.RunHTTPServer()
	for tasks := 0; tasks < 4; tasks++ {
		// go marketfeed.RunTickerSocket()

		go func() {
			trade.ConnectToMarginAccount()

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

		go func() {

			s, err := marketfeed.GetBlockchainStats()
			if err != nil {
				errBC <- err
				return
			}

			blockchain <- s
		}()

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

	// runtime.GOMAXPROCS(int(float64(runtime.NumCPU()) * 1.25))
	//
	// log.Println("Starting application...")
	//
	// log.Println("Starting Bitcoin Trader...")
	//
	// /*
	//  * When SIGINT or SIGTERM is caught write to the quitChannel
	//  */
	// quitChannel := make(chan os.Signal)
	// errChannel := make(chan error, 1)
	// signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	//
	// shutdownChannel := make(chan bool)
	// waitGroup := &sync.WaitGroup{}
	//
	// waitGroup.Add(maxGoRoutines)
	//
	// /*
	//  * Create a goroutine that does imaginary work
	//  */
	// for i := 0; i < maxGoRoutines; i++ {
	// 	go func(shutdownChannel chan bool, errChannel chan error, waitGroup *sync.WaitGroup, id int) {
	// 		log.Println("Starting work goroutine...")
	// 		//defer waitGroup.Done()
	//
	// 		for {
	// 			/*
	// 			 * Listen on channels for message.
	// 			 */
	// 			select {
	// 			case _ = <-shutdownChannel:
	// 				log.Printf("Received shutdown on goroutine %d\n", id)
	// 				return
	// 			case _ = <-errChannel:
	// 				log.Println(errChannel)
	// 				return
	//
	// 			default:
	// 			}
	//
	// 			// Do some hard work here!
	// 			_, err := marketfeed.GetHistoricTrades(waitGroup, "bitfinexUSD", "okcoinCNY", "coinbaseUSD")
	//
	// 			if err != nil {
	// 				errChannel <- err
	// 			}
	//
	// 		}
	// 	}(shutdownChannel, errChannel, waitGroup, i)
	// }
	//
	// /*
	//  * Wait until we get the quit message
	//  */
	// <-errChannel
	//
	// log.Println("Received quit. Sending shutdown and waiting on goroutines...")
	//
	// for i := 0; i < maxGoRoutines; i++ {
	// 	shutdownChannel <- true
	// }
	//
	// /*
	//  * Block until wait group counter gets to zero
	//  */
	// waitGroup.Wait()
	// log.Println("Done.")
}
