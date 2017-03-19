package trade

import (
	"fmt"
	"os"
	"time"

	bitfinex "github.com/bitfinexcom/bitfinex-api-go"
)

type CoinbaseCredentials struct {
	APIKey string
	Secret string
}

type Wallet struct {
	Type              string
	Currency          string
	Balance           float64
	UnsettledInterest float64
	BalanceAvailable  float64
}

type CoinbasePurchase struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PaymentMethod struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"payment_method"`
	Transaction struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"transaction"`
	Amount struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Total struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"total"`
	Subtotal struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"subtotal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	Resource     string    `json:"resource"`
	ResourcePath string    `json:"resource_path"`
	Committed    bool      `json:"committed"`
	Instant      bool      `json:"instant"`
	Fee          struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"fee"`
	PayoutAt string `json:"payout_at"`
}

type CoinbaseSell struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	PaymentMethod struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"payment_method"`
	Transaction struct {
		ID           string `json:"id"`
		Resource     string `json:"resource"`
		ResourcePath string `json:"resource_path"`
	} `json:"transaction"`
	Amount struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`
	Total struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"total"`
	Subtotal struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"subtotal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    string    `json:"updated_at"`
	Resource     string    `json:"resource"`
	ResourcePath string    `json:"resource_path"`
	Committed    bool      `json:"committed"`
	Instant      bool      `json:"instant"`
	Fee          struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"fee"`
	PayoutAt string `json:"payout_at"`
}

type PaymentMethodResource struct {
	ID            string    `json:"id"`
	Type          string    `json:"type"`
	Name          string    `json:"name"`
	Currency      string    `json:"currency"`
	PrimaryBuy    bool      `json:"primary_buy"`
	PrimarySell   bool      `json:"primary_sell"`
	AllowBuy      bool      `json:"allow_buy"`
	AllowSell     bool      `json:"allow_sell"`
	AllowDeposit  bool      `json:"allow_deposit"`
	AllowWithdraw bool      `json:"allow_withdraw"`
	InstantBuy    bool      `json:"instant_buy"`
	InstantSell   bool      `json:"instant_sell"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
	Resource      string    `json:"resource"`
	ResourcePath  string    `json:"resource_path"`
}

//Trade needs to be able to call environment variables for bitfenix api use and be able to pass into payload for POST request to make a trade on platform
// with x amount of bitcoin for today
// will need to be inserted to db (dear god it better be)
// will need to perform trade function after scripts pass okay

func ConnectToMarginAccount() {
	key := os.Getenv("BFX_API_KEY")
	secret := os.Getenv("BFX_API_SECRET")
	client := bitfinex.NewClient().Auth(key, secret)

	info, err := client.Account.Info()
	if err != nil {
		fmt.Println("Error ", err)
	}

	fmt.Println("BFX INFO", info)
}
