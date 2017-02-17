package main

import (
	//"github.com/bitfinexcom/bitfinex-api-go"

	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/Bitcoin_Trader/marketfeed"
	//"os"
)

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

	runtime.GOMAXPROCS(int(float64(runtime.NumCPU()) * 1.25))

	log.Println("Starting application...")

	log.Println("Starting Bitcoin Trader...")

	log.Println("Starting HTTP Server")

	/*
	 * When SIGINT or SIGTERM is caught write to the quitChannel
	 */
	quitChannel := make(chan os.Signal)
	errChannel := make(chan error, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

	shutdownChannel := make(chan bool)
	waitGroup := &sync.WaitGroup{}

	waitGroup.Add(maxGoRoutines)

	/*
	 * Create a goroutine that does imaginary work
	 */
	for i := 0; i < maxGoRoutines; i++ {
		go func(shutdownChannel chan bool, errChannel chan error, waitGroup *sync.WaitGroup, id int) {
			log.Println("Starting work goroutine...")
			//defer waitGroup.Done()

			for {
				/*
				 * Listen on channels for message.
				 */
				select {
				case _ = <-shutdownChannel:
					log.Printf("Received shutdown on goroutine %d\n", id)
					return
				case _ = <-errChannel:
					log.Println(errChannel)
					return

				default:
				}

				// Do some hard work here!
				_, err := marketfeed.GetHistoricTrades(waitGroup, "bitfinexUSD")

				if err != nil {
					errChannel <- err
				}

			}
		}(shutdownChannel, errChannel, waitGroup, i)
	}

	/*
	 * Wait until we get the quit message
	 */
	<-errChannel

	log.Println("Received quit. Sending shutdown and waiting on goroutines...")

	for i := 0; i < maxGoRoutines; i++ {
		shutdownChannel <- true
	}

	/*
	 * Block until wait group counter gets to zero
	 */
	waitGroup.Wait()
	log.Println("Done.")
}

// package main
//
// import (
//         "log"
//         "os"
//         "os/signal"
//         "runtime"
//         "sync"
//         "syscall"
// )
//

// }
