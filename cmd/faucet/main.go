package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/starport/starport/pkg/chaincmd"
	chaincmdrunner "github.com/tendermint/starport/starport/pkg/chaincmd/runner"
	"github.com/tendermint/starport/starport/pkg/cosmosfaucet"
	"github.com/tendermint/starport/starport/pkg/cosmosver"
	"github.com/tendermint/starport/starport/pkg/xhttp"
)

type SiteVerifyResponse struct {
	Success     bool      `json:"success"`
	Score       float64   `json:"score"`
	Action      string    `json:"action"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
}

type TransferRequest struct {
	AccountAddress  string   `json:"address"`
	Coins           []string `json:"coins"`
	CaptchaResponse string   `json:"captchaResponse"`
}

func main() {
	flag.Parse()

	configKeyringBackend, err := chaincmd.KeyringBackendFromString(keyringBackend)
	if err != nil {
		log.Fatal(err)
	}

	ccoptions := []chaincmd.Option{
		chaincmd.WithKeyringPassword(keyringPassword),
		chaincmd.WithKeyringBackend(configKeyringBackend),
		chaincmd.WithAutoChainIDDetection(),
		chaincmd.WithNodeAddress(nodeAddress),
	}

	if legacySendCmd {
		ccoptions = append(ccoptions, chaincmd.WithLegacySendCommand())
	}

	if sdkVersion == string(cosmosver.Stargate) {
		ccoptions = append(ccoptions,
			chaincmd.WithVersion(cosmosver.StargateFortyFourVersion),
		)
	} else {
		log.Fatal("The chain is not using cosmossdk > 0.44")
		// ccoptions = append(ccoptions,
		// 	chaincmd.WithVersion(cosmosver.LaunchpadAny),
		// 	chaincmd.WithLaunchpadCLI(appCli),
		// )
	}

	cr, err := chaincmdrunner.New(context.Background(), chaincmd.New(appCli, ccoptions...))
	if err != nil {
		log.Fatal(err)
	}

	coins := strings.Split(defaultDenoms, denomSeparator)

	faucetOptions := make([]cosmosfaucet.Option, len(coins))
	for i, coin := range coins {
		faucetOptions[i] = cosmosfaucet.Coin(creditAmount, maxCredit, coin)
	}

	faucetOptions = append(faucetOptions, cosmosfaucet.Account(keyName, keyMnemonic))

	faucet, err := cosmosfaucet.New(context.Background(), cr, faucetOptions...)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		originHeader := r.Header.Get("Origin")
		if originHeader == "http://localhost:3000" || originHeader == corsDomainPrivateTestnet || originHeader == corsDomainPublicTestnet {
			w.Header().Set("Access-Control-Allow-Origin", originHeader)
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			if (*r).Method == "OPTIONS" {
				return
			}
		}

		buf, _ := ioutil.ReadAll(r.Body)
		rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))

		var req TransferRequest
		err := json.NewDecoder(rdr1).Decode(&req)

		isValidCudosAddress, _ := regexp.MatchString(
			"^cudos[0-9a-z]{39}$",
			req.AccountAddress)

		if !isValidCudosAddress {
			http.Error(w, "Wrong address format", http.StatusUnauthorized)
		}

		if err == nil {
			captchaErr := checkCaptchaWithKey(req.CaptchaResponse)

			if captchaErr != nil {
				http.Error(w, "Wrong captcha", http.StatusUnauthorized)
				return
			}

			coin := req.Coins[0]
			cosmosCoin, err := types.ParseCoinNormalized(coin)
			if err == nil {
				if cosmosCoin.Amount.GT(types.NewIntFromUint64(maxCredit)) {
					var transfers []cosmosfaucet.Transfer
					t := cosmosfaucet.Transfer{
						Coin:   cosmosCoin.Denom,
						Status: "error",
						Error:  fmt.Sprintf("maximum credit (%d)", maxCredit),
					}

					transfers = append(transfers, t)

					xhttp.ResponseJSON(w, http.StatusOK, cosmosfaucet.TransferResponse{
						Transfers: transfers,
					})
					return
				}

				rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
				r.Body = rdr2
				faucet.ServeHTTP(w, r)
			}
		}
	})
	log.Infof("listening on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func checkCaptchaWithKey(captcha string) error {
	// secret api key for google recaptcha
	secret := googleApiKey

	siteVerifyURL := "https://www.google.com/recaptcha/api/siteverify"
	req, err := http.NewRequest(http.MethodPost, siteVerifyURL, nil)
	if err != nil {
		return err
	}

	// Add necessary request parameters.
	q := req.URL.Query()
	q.Add("secret", secret)
	q.Add("response", captcha)
	req.URL.RawQuery = q.Encode()

	// Make request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Decode response
	var body SiteVerifyResponse
	if err = json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return err
	}

	if body.Score < 0.3 {
		fmt.Printf("Captcha score %f\n", body.Score)
		// Score is always returned 0 even for testing keys temporary workaround
		//return errors.New("invalid captcha")
	}

	return nil
}
