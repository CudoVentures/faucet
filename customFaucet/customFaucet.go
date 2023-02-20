package customFaucet

import (
	"context"
	"fmt"
	"time"

	customChaincmdrunner "github.com/tendermint/faucet/customChaincmdrunner"
	bigNumber "lukechampine.com/uint128"
)

const (
	// DefaultAccountName is the default account to transfer tokens from.
	DefaultAccountName = "faucet"

	// DefaultDenom is the default denomination to distribute.
	DefaultDenom = "acudos"

	// DefaultAmount specifies the default amount to transfer to an account
	// on each request.
	DefaultAmount = "100000000000000000000"

	// DefaultMaxAmount specifies the maximum amount that can be tranffered to an
	// account in all times.
	DefaultMaxAmount = "100000000000000000000"

	// DefaultMaxAmount specifies the --fees flag amount for the BankSend Tx
	DefaultFees = "1000000000000000000"

	// DefaultLimitRefreshWindow specifies the time after which the max amount limit
	// is refreshed for an account [1 year]
	DefaultRefreshWindow = time.Hour * 24 * 365
)

// Faucet represents a faucet.
type Faucet struct {
	// runner used to intereact with blockchain's binary to transfer tokens.
	runner customChaincmdrunner.Runner

	// chainID is the chain id of the chain that faucet is operating for.
	chainID string

	// accountName to transfer tokens from.
	accountName string

	// accountMnemonic is the mnemonic of the account.
	accountMnemonic string

	// coins keeps a list of coins that can be distributed by the faucet.
	coins []coin

	transferCoins []transferCoin

	// coinsMax is a denom-max pair.
	// it holds the maximum amounts of coins that can be sent to a single account.
	coinsMax map[string]bigNumber.Uint128

	limitRefreshWindow time.Duration

	// openAPIData holds template data customizations for serving OpenAPI page & spec.
	openAPIData openAPIData
}

type coin struct {
	// amount is the amount of the coin can be distributed per request.
	amount bigNumber.Uint128

	// denom is denomination of the coin to be distributed by the faucet.
	denom string
}

type transferCoin struct {
	// amount is the amount of the coin can be distributed per request.
	amount string

	// denom is denomination of the coin to be distributed by the faucet.
	denom string
}

func (c coin) String() string {
	return fmt.Sprintf("%d%s", c.amount, c.denom)
}

// Option configures the faucetOptions.
type Option func(*Faucet)

// Account provides the account information to transfer tokens from.
// when mnemonic isn't provided, account assumed to be exists in the keyring.
func Account(name, mnemonic string) Option {
	return func(f *Faucet) {
		f.accountName = name
		f.accountMnemonic = mnemonic
	}
}

// Coin adds a new coin to coins list to distribute by the faucet.
// the first coin added to the list considered as the default coin during transfer requests.
//
// amount is the amount of the coin can be distributed per request.
// maxAmount is the maximum amount of the coin that can be sent to a single account.
// denom is denomination of the coin to be distributed by the faucet.
func Coin(amount bigNumber.Uint128, maxAmount bigNumber.Uint128, denom string) Option {
	return func(f *Faucet) {
		f.coins = append(f.coins, coin{amount, denom})
		f.coinsMax[denom] = maxAmount
	}
}

// RefreshWindow adds the duration to refresh the transfer limit to the faucet
func RefreshWindow(refreshWindow time.Duration) Option {
	return func(f *Faucet) {
		f.limitRefreshWindow = refreshWindow
	}
}

// ChainID adds chain id to faucet. faucet will automatically fetch when it isn't provided.
func ChainID(id string) Option {
	return func(f *Faucet) {
		f.chainID = id
	}
}

// OpenAPI configures how to serve Open API page and and spec.
func OpenAPI(apiAddress string) Option {
	return func(f *Faucet) {
		f.openAPIData.APIAddress = apiAddress
	}
}

// New creates a new faucet with ccr (to access and use blockchain's CLI) and given options.
func New(ctx context.Context, ccr customChaincmdrunner.Runner, options ...Option) (Faucet, error) {
	f := Faucet{
		runner:      ccr,
		accountName: DefaultAccountName,
		coinsMax:    make(map[string]bigNumber.Uint128),
		openAPIData: openAPIData{"Blockchain", "http://localhost:1317"},
	}

	for _, apply := range options {
		apply(&f)
	}

	if len(f.coins) == 0 {
		amount, _ := bigNumber.FromString(DefaultAmount)
		max, _ := bigNumber.FromString(DefaultMaxAmount)
		Coin(amount, max, DefaultDenom)(&f)
	}

	if f.limitRefreshWindow == 0 {
		RefreshWindow(DefaultRefreshWindow)(&f)
	}

	// import the account if mnemonic is provided.
	if f.accountMnemonic != "" {
		_, err := f.runner.AddAccount(ctx, f.accountName, f.accountMnemonic)
		if err != nil && err != customChaincmdrunner.ErrAccountAlreadyExists {
			return Faucet{}, err
		}
	}

	if f.chainID == "" {
		status, err := f.runner.Status(ctx)
		if err != nil {
			return Faucet{}, err
		}

		f.chainID = status.ChainID
		f.openAPIData.ChainID = status.ChainID
	}

	return f, nil
}
