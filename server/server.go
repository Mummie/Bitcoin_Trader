package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go"
	"golang.org/x/net/websocket"
)

type Connections struct {
	clients      map[chan string]bool
	addClient    chan chan string
	removeClient chan chan string
	messages     chan string
}

var hub = &Connections{
	clients:      make(map[chan string]bool),
	addClient:    make(chan (chan string)),
	removeClient: make(chan (chan string)),
	messages:     make(chan string),
}

// Init will initialize client connections and broadcast to http server
func (hub *Connections) Init() {
	go func() {
		for {
			select {
			case s := <-hub.addClient:
				hub.clients[s] = true
				log.Println("Added new client")
			case s := <-hub.removeClient:
				delete(hub.clients, s)
				log.Println("Removed client")
			case msg := <-hub.messages:
				for s := range hub.clients {
					s <- msg
				}
				log.Printf("Broadcast \"%v\" to %d clients", msg, len(hub.clients))
			}
		}
	}()
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
		return
	}

	if r.URL.Path == "/send" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		str := string(body)
		hub.messages <- str
		f.Flush()
		return
	}

	messageChannel := make(chan string)
	hub.addClient <- messageChannel
	notify := w.(http.CloseNotifier).CloseNotify()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for i := 0; i < 1440; {
		select {
		case msg := <-messageChannel:
			jsonData, _ := json.Marshal(msg)
			str := string(jsonData)
			if r.URL.Path == "/events/sse" {
				fmt.Fprintf(w, "data: {\"str\": %s, \"time\": \"%v\"}\n\n", str, time.Now())
			} else if r.URL.Path == "/events/lp" {
				fmt.Fprintf(w, "{\"str\": %s, \"time\": \"%v\"}", str, time.Now())
			}
			f.Flush()
		case <-time.After(time.Second * 60):
			if r.URL.Path == "/events/sse" {
				fmt.Fprintf(w, "data: {\"str\": \"No Data\"}\n\n")
			} else if r.URL.Path == "/events/lp" {
				fmt.Fprintf(w, "{\"str\": \"No Data\"}")
			}
			f.Flush()
			i++
		case <-notify:
			f.Flush()
			i = 1440
			hub.removeClient <- messageChannel
		}
	}
}

//websocketHandler will handle any http incoming messages/data and output it to it's respective page
// todo: attach tickChan for listening to incoming messages and send to socket
func websocketHandler(ws *websocket.Conn) {
	var in string
	messageChannel := make(chan string)
	hub.addClient <- messageChannel

	c := bitfinex.NewClient()

	// in case your proxy is using a non valid certificate set to TRUE
	c.WebSocketTLSSkipVerify = false

	err := c.WebSocket.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer c.WebSocket.Close()
	//
	// book_btcusd_chan := make(chan []float64)
	// book_ltcusd_chan := make(chan []float64)
	// trades_chan := make(chan []float64)
	ticker_chan := make(chan []float64)

	// c.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.BTCUSD, book_btcusd_chan)
	// c.WebSocket.AddSubscribe(bitfinex.CHAN_BOOK, bitfinex.LTCUSD, book_ltcusd_chan)
	// c.WebSocket.AddSubscribe(bitfinex.CHAN_TRADE, bitfinex.BTCUSD, trades_chan)
	c.WebSocket.AddSubscribe(bitfinex.CHAN_TICKER, bitfinex.BTCUSD, ticker_chan)

	// go listen(book_btcusd_chan, "BOOK BTCUSD:")
	// go listen(book_ltcusd_chan, "BOOK LTCUSD:")
	// go listen(trades_chan, "TRADES BTCUSD:")

	err = c.WebSocket.Subscribe()
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < 1440; {
		select {
		case tick := <-ticker_chan:
			jsonData, _ := json.Marshal(tick)
			str := string(jsonData)
			in = fmt.Sprintf("{\"tick\": %s, \"time\": \"%v\"}\n\n", str, time.Now())
			if err := websocket.Message.Send(ws, in); err != nil {
				i = 1440
			}
		case msg := <-messageChannel:
			jsonData, _ := json.Marshal(msg)
			str := string(jsonData)
			in = fmt.Sprintf("{\"str\": %s, \"time\": \"%v\"}\n\n", str, time.Now())

			if err := websocket.Message.Send(ws, in); err != nil {
				hub.removeClient <- messageChannel
				i = 1440
			}
		case <-time.After(time.Second * 60):
			in = fmt.Sprintf("{\"str\": \"No Data\"}\n\n")
			if err := websocket.Message.Send(ws, in); err != nil {
				hub.removeClient <- messageChannel
				i = 1440
			}
		}
		i++
	}
}

func telnetHandler(c *net.TCPConn) {
	defer c.Close()
	fmt.Printf("Connection from %s to %s established.\n", c.RemoteAddr(), c.LocalAddr())
	io.WriteString(c, fmt.Sprintf("Welcome on %s\n", c.LocalAddr()))
	buf := make([]byte, 4096)
	for {
		n, err := c.Read(buf)
		if (err != nil) || (n == 0) {
			c.Close()
			break
		}
		str := strings.TrimSpace(string(buf[0:n]))
		hub.messages <- str
		io.WriteString(c, "sent to "+strconv.Itoa(len(hub.clients))+" clients\n")
	}
	time.Sleep(150 * time.Millisecond)
	fmt.Printf("Connection from %v closed.\n", c.RemoteAddr())
	c.Close()
	return
}

func listenForTelnet(ln *net.TCPListener) {
	for {
		conn, err := ln.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go telnetHandler(conn)
	}
}

// RunHTTPServer will start the http application server and host client connections, connectivity to websocket and stream data from bitcoin app
func RunHTTPServer() {
	fmt.Printf("application started at: %s\n", time.Now().Format(time.RFC822))
	var starttime = time.Now().Unix()
	runtime.GOMAXPROCS(8)

	hub.Init()

	http.HandleFunc("/send", httpHandler)
	http.HandleFunc("/events/sse", httpHandler)
	http.HandleFunc("/events/lp", httpHandler)
	http.Handle("/events/ws", websocket.Handler(websocketHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})
	// http.Handle("/", http.FileServer(http.Dir("./public/")))

	go http.ListenAndServe("localhost:5001", nil)

	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		Port: 8001,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go listenForTelnet(ln)

	var input string
	for input != "exit" {
		_, _ = fmt.Scanf("%v", &input)
		if input != "exit" {
			switch input {
			case "", "0", "5", "help", "info":
				fmt.Print("you can type \n1: \"exit\" to kill this application")
				fmt.Print("\n2: \"clients\" to show the amount of connected clients")
				fmt.Print("\n3: \"system\" to show info about the server")
				fmt.Print("\n4: \"time\" to show since when this application is running")
				fmt.Print("\n5: \"help\" to show this information")
				fmt.Println()
			case "1", "exit", "kill":
				fmt.Println("application get killed in 5 seconds")
				input = "exit"
				time.Sleep(5 * time.Second)
			case "2", "clients":
				fmt.Printf("connected to %d clients\n", len(hub.clients))
			case "3", "system":
				fmt.Printf("CPU cores: %d\nGo calls: %d\nGo routines: %d\nGo version: %v\nProcess ID: %v\n", runtime.NumCPU(), runtime.NumCgoCall(), runtime.NumGoroutine(), runtime.Version(), syscall.Getpid())
			case "4", "time":
				fmt.Printf("application running since %d minutes\n", (time.Now().Unix()-starttime)/60)
			}
		}
	}
	os.Exit(0)
}

// type Config struct {
// 	Port             string `json:"port"`
// 	GrantTTL         int    `json:"grant_ttl"`
// 	btcsChannelGroup string `json:"btcs_channel_goroup"`
// 	HistoryChannel   string `json:"history_channel"`
// 	ChatChannel      string `json:"chat_channel"`
// 	Keys             struct {
// 		Pub    string `json:"publish_key"`
// 		Sub    string `json:"subscribe_key"`
// 		Secret string `json:"secret_key"`
// 		Auth   string `json:"auth_key"`
// 	}
// }

//http://jumflot.jumware.com/ for candlestick
// http://www.flotcharts.org/ for graphing

// func LoadConfig() {
// 	configPath := "../"
//
// 	// Fallback to local config files
// 	if configPath == "" {
// 		configPath = "./config/"
// 	}
//
// 	// Auths
// 	file, err := os.Open(configPath + "config.json")
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	config = Config{}
//
// 	decoder := json.NewDecoder(file)
//
// 	err = decoder.Decode(&config)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	// btcs
// 	file, err = os.Open(configPath + btcS_FILE)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	decoder = json.NewDecoder(file)
//
// 	err = decoder.Decode(&btcs)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
//
// 	// btc names
// 	btcNamesArray := make([]string, 0, len(btcs))
// 	for _, v := range btcs {
// 		btcNamesArray = append(btcNamesArray, v.Name)
// 	}
//
// 	btcNames = strings.Join(btcNamesArray, ",")
//
// 	// Authentication key for management instance inside GrantPermissions()
// 	// function
// 	bootstrapAuth = config.Keys.Auth + BOOTSTRAP_INSTANCE_SUFFIX
// }

// func SetUpChannelGroup() {
// 	errorChannel := make(chan []byte)
// 	successChannel := make(chan []byte)
// 	done := make(chan bool)
//
// 	pubnub := messaging.NewPubnub(config.Keys.Pub, config.Keys.Sub, "", "",
// 		false, "")
//
// 	pubnub.SetAuthenticationKey(bootstrapAuth)
//
// 	// Remove Group
// 	go pubnub.ChannelGroupRemoveGroup(config.btcsChannelGroup,
// 		successChannel, errorChannel)
// 	go handleResponse(successChannel, errorChannel,
// 		messaging.GetNonSubscribeTimeout(), done)
//
// 	<-done
//
// 	// Create it from the scratch
// 	go pubnub.ChannelGroupAddChannel(config.btcsChannelGroup, btcNames,
// 		successChannel, errorChannel)
// 	go handleResponse(successChannel, errorChannel,
// 		messaging.GetNonSubscribeTimeout(), done)
//
// 	<-done
// }

type StreamMessage struct {
	Time       string `json:"time"`
	Price      string `json:"price"`
	Delta      string `json:"delta"`
	Percentage string `json:"perc"`
	Vol        int    `json:"vol"`
}

// Exposing keys for clients throught HTTP
// func GetConfigsHandler(w http.ResponseWriter, req *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Write([]byte(
// 		fmt.Sprintf("{\"publish_key\": \"%s\", \"subscribe_key\": \"%s\"}",
// 			config.Keys.Pub, config.Keys.Sub)))
// }
//
// func ServeHttp() {
// 	publicPath := os.Getenv("PUBNUB_btcS_PUBLIC")
//
// 	http.Handle("/", http.FileServer(http.Dir(publicPath)))
// 	http.HandleFunc("/get_configs", GetConfigsHandler)
//
// 	err := http.ListenAndServe(fmt.Sprintf(":%s", config.Port), nil)
// 	if err != nil {
// 		log.Panic(err)
// 	}
// 	log.Println("Starting HTTP Server")
//
// }
//
// // Handlers
// func handleResponse(successChannel, errorChannel chan []byte, timeout uint16,
// 	finishedChannel chan bool) {
//
// 	select {
// 	case success := <-successChannel:
// 		fmt.Printf("%s\n", success)
// 	case failure := <-errorChannel:
// 		fmt.Printf("ERROR: %s\n", failure)
// 	case <-time.After(time.Second * time.Duration(timeout)):
// 		fmt.Println("Request timeout")
// 	}
//
// 	finishedChannel <- true
// }
//
// func (btc *marketfeed.TickerData) RunCycle() {
// 	cycle := make(chan bool)
// 	i := 0
// 	pubnub := messaging.NewPubnub(config.Keys.Pub, config.Keys.Sub, "", "",
// 		false, "")
// 	pubnub.SetAuthenticationKey(config.Keys.Auth)
//
// 	for {
// 		go btc.UpdateValuesAndPublish(pubnub, cycle)
// 		<-cycle
// 		i++
// 	}
// }
//
// //todo: change the struct value changes to match marketfeed.TickerData
// func (btc *marketfeed.TickerData) UpdateValuesAndPublish(pubnub *messaging.Pubnub,
// 	cycle chan bool) {
//
// 	rand.Seed(int64(time.Now().Nanosecond()))
//
// 	change := float64(rand.Intn(btc.Volume)-btc.Volume/2) / 100
// 	btc.Ask = btc.Ask + float64(change)
// 	delta := btc.Ask - btc.LastPrice
// 	percentage := Roundn((1-btc.LastPrice/btc.Ask)*100, 2)
// 	vol := Randn(btc.Volume, 1000) * 10
//
// 	streamMessage := StreamMessage{
// 		Time:       time.Now().Format(TIME_FORMAT),
// 		Price:      fmt.Sprintf("%.2f", btc.Ask),
// 		Delta:      fmt.Sprintf("%.2f", delta),
// 		Percentage: fmt.Sprintf("%.2f", percentage),
// 		Vol:        vol}
//
// 	if math.Abs(percentage) > float64(btc.MaxDelta)/100 {
// 		btc.Ask = btc.LastPrice
// 	}
//
// 	errorChannel := make(chan []byte)
// 	successChannel := make(chan []byte)
// 	done := make(chan bool)
//
// 	go pubnub.Publish(btc.Name, streamMessage, successChannel, errorChannel)
// 	go handleResponse(successChannel, errorChannel,
// 		messaging.GetNonSubscribeTimeout(), done)
//
// 	time.Sleep(time.Duration(600) * time.Second)
//
// 	cycle <- <-done
// }
