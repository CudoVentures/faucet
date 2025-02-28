package main

import (
	"flag"

	"github.com/tendermint/starport/starport/pkg/cosmosfaucet"

	"github.com/tendermint/faucet/internal/environ"
)

const (
	denomSeparator = ","
)

var (
	captchSiteKey            string
	googleApiKey             string
	googleProjectId          string
	port                     int
	keyringBackend           string
	sdkVersion               string
	keyName                  string
	keyMnemonic              string
	keyringPassword          string
	appCli                   string
	defaultDenoms            string
	creditAmount             uint64
	maxCredit                uint64
	nodeAddress              string
	legacySendCmd            bool
	corsDomainPublicTestnet  string
	corsDomainPrivateTestnet string
)

func init() {
	flag.IntVar(&port, "port",
		environ.GetInt("PORT", 8000),
		"tcp port where faucet will be listening for requests",
	)
	flag.StringVar(&captchSiteKey, "captcha-site-key",
		environ.GetString("CAPTCHA_SITE_KEY", ""),
		"site key of recaptcha",
	)
	flag.StringVar(&googleApiKey, "google-api-key",
		environ.GetString("GOOGLE_API_KEY", ""),
		"google api key",
	)
	flag.StringVar(&googleProjectId, "google-project-id",
		environ.GetString("GOOGLE_PROJECT_ID", ""),
		"google project id",
	)
	flag.StringVar(&keyringBackend, "keyring-backend",
		environ.GetString("KEYRING_BACKEND", ""),
		"keyring backend to be used",
	)
	flag.StringVar(&sdkVersion, "sdk-version",
		environ.GetString("SDK_VERSION", "stargate"),
		"version of sdk (launchpad or stargate)",
	)
	flag.StringVar(&keyName, "account-name",
		environ.GetString("ACCOUNT_NAME", cosmosfaucet.DefaultAccountName),
		"name of the account to be used by the faucet",
	)
	flag.StringVar(&keyMnemonic, "mnemonic",
		environ.GetString("MNEMONIC", ""),
		"mnemonic for restoring an account",
	)
	flag.StringVar(&keyringPassword, "keyring-password",
		environ.GetString("KEYRING_PASSWORD", ""),
		"password for accessing keyring",
	)
	flag.StringVar(&appCli, "cli-name",
		environ.GetString("CLI_NAME", "gaiad"),
		"name of the cli executable",
	)
	flag.StringVar(&defaultDenoms, "denoms",
		environ.GetString("DENOMS", cosmosfaucet.DefaultDenom),
		"denomination of the coins sent by default (comma separated)",
	)
	flag.Uint64Var(&creditAmount,
		"credit-amount",
		environ.GetUint64("CREDIT_AMOUNT", cosmosfaucet.DefaultAmount),
		"amount to credit in each request",
	)
	flag.Uint64Var(&maxCredit,
		"max-credit", environ.GetUint64("MAX_CREDIT", cosmosfaucet.DefaultMaxAmount),
		"maximum credit per account",
	)
	flag.StringVar(&nodeAddress, "node",
		environ.GetString("NODE", ""),
		"address of tendermint RPC endpoint for this chain",
	)
	flag.BoolVar(&legacySendCmd, "legacy-send",
		environ.GetBool("LEGACY_SEND", false),
		"whether to use legacy send command",
	)
	flag.StringVar(&corsDomainPublicTestnet, "cors-domain-public-testnet",
		environ.GetString("CORS_DOMAIN_PUBLIC_TESTNET", ""),
		"cors domain for public testnet",
	)
	flag.StringVar(&corsDomainPrivateTestnet, "cors-domain-private-testnet",
		environ.GetString("CORS_DOMAIN_PRIVATE_TESTNET", ""),
		"cors domain for private testnet",
	)
}
