package marketfeed

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

//BlockChain holds all blockchain data
type BlockChain struct {
	MarketPriceUsd                float64 `json:"market_price_usd"`
	HashRate                      float64 `json:"hash_rate"`
	TotalFeesBtc                  int64   `json:"total_fees_btc"`
	NBtcMined                     int64   `json:"n_btc_mined"`
	NTx                           int     `json:"n_tx"`
	NBlocksMined                  int     `json:"n_blocks_mined"`
	MinutesBetweenBlocks          float64 `json:"minutes_between_blocks"`
	Totalbc                       int64   `json:"totalbc"`
	NBlocksTotal                  int     `json:"n_blocks_total"`
	EstimatedTransactionVolumeUsd float64 `json:"estimated_transaction_volume_usd"`
	BlocksSize                    int     `json:"blocks_size"`
	MinersRevenueUsd              float64 `json:"miners_revenue_usd"`
	Nextretarget                  int     `json:"nextretarget"`
	Difficulty                    int64   `json:"difficulty"`
	EstimatedBtcSent              int64   `json:"estimated_btc_sent"`
	MinersRevenueBtc              int     `json:"miners_revenue_btc"`
	TotalBtcSent                  int64   `json:"total_btc_sent"`
	TradeVolumeBtc                float64 `json:"trade_volume_btc"`
	TradeVolumeUsd                float64 `json:"trade_volume_usd"`
	Timestamp                     int64   `json:"timestamp"`
}

//BlockChainPool holds the mining pool info and shows the amount of blocks mined by miners
type BlockChainPool struct {
	GHashIO         int `json:"GHash.IO"`
	Nine512848209   int `json:"95.128.48.209"`
	NiceHashSolo    int `json:"NiceHash Solo"`
	SoloCKPool      int `json:"Solo CKPool"`
	One76931178     int `json:"176.9.31.178"`
	OneHash         int `json:"1Hash"`
	Two1711225189   int `json:"217.11.225.189"`
	Unknown         int `json:"Unknown"`
	BitClubNetwork  int `json:"BitClub Network"`
	Telco214        int `json:"Telco 214"`
	HaoBTC          int `json:"HaoBTC"`
	GBMiners        int `json:"GBMiners"`
	SlushPool       int `json:"SlushPool"`
	Nine122013139   int `json:"91.220.131.39"`
	KanoCKPool      int `json:"Kano CKPool"`
	BTCCPool        int `json:"BTCC Pool"`
	Six020510755    int `json:"60.205.107.55"`
	BitMinter       int `json:"BitMinter"`
	BitFury         int `json:"BitFury"`
	AntPool         int `json:"AntPool"`
	F2Pool          int `json:"F2Pool"`
	ViaBTC          int `json:"ViaBTC"`
	BWCOM           int `json:"BW.COM"`
	BTCCom          int `json:"BTC.com"`
	Four7895125     int `json:"47.89.51.25"`
	Seven4118157122 int `json:"74.118.157.122"`
}

// SendBlockChainData will create a websocket to bitfenix blockchain and return the output message from the websocket to a channel
func SendBlockChainData() {
	var addr = flag.String("addr", "https://blockchain.info/api/api_websocket", "https://blockchain.info/api/api_websocket")
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			// To cleanly close a connection, a client should send a close
			// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			c.Close()
			return
		}
	}
}

//GetBlockchainStats will return the stats behind Blockchain
func GetBlockchainStats() (b *BlockChain, err error) {
	res, err := getJSONData("https://api.blockchain.info/stats")

	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(res, &b)

	if err != nil {
		return nil, err
	}

	return b, nil
}

//GetBlockchainPools will return blockchain pools info by timespan
func GetBlockchainPools(timespan int) (p *BlockChainPool, err error) {
	if timespan > 10 {
		return nil, fmt.Errorf("Error maximum timespan is 10 days")
	}

	url := fmt.Sprintf("https://api.blockchain.info/pools?timespan=%vdays", timespan)
	res, err := getJSONData(url)

	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(res, &p)

	if err != nil {
		return nil, err
	}

	return p, nil
}
