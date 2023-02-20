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

	"github.com/tendermint/faucet/customChaincmd"
	customChaincmdrunner "github.com/tendermint/faucet/customChaincmdrunner"
	"github.com/tendermint/faucet/customFaucet"
	"github.com/tendermint/starport/starport/pkg/cosmosver"
	"github.com/tendermint/starport/starport/pkg/xhttp"
	bigNumber "lukechampine.com/uint128"
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

type TransferRequestWithFees struct {
	AccountAddress  string   `json:"address"`
	Fees            string   `json:"fees"`
	Coins           []string `json:"coins"`
	CaptchaResponse string   `json:"captchaResponse"`
}

func main() {
	flag.Parse()

	configKeyringBackend, err := customChaincmd.KeyringBackendFromString(keyringBackend)
	if err != nil {
		log.Fatal(err)
	}

	ccoptions := []customChaincmd.Option{
		customChaincmd.WithKeyringPassword(keyringPassword),
		customChaincmd.WithKeyringBackend(configKeyringBackend),
		customChaincmd.WithAutoChainIDDetection(),
		customChaincmd.WithNodeAddress(nodeAddress),
	}

	if legacySendCmd {
		ccoptions = append(ccoptions, customChaincmd.WithLegacySendCommand())
	}

	if sdkVersion == string(cosmosver.Stargate) {
		ccoptions = append(ccoptions,
			customChaincmd.WithVersion(cosmosver.StargateFortyFourVersion),
		)
	} else {
		log.Fatal("The chain is not using cosmossdk > 0.44")
		// ccoptions = append(ccoptions,
		// 	chaincmd.WithVersion(cosmosver.LaunchpadAny),
		// 	chaincmd.WithLaunchpadCLI(appCli),
		// )
	}

	cr, err := customChaincmdrunner.New(context.Background(), customChaincmd.New(appCli, ccoptions...))
	if err != nil {
		log.Fatal(err)
	}

	coins := strings.Split(defaultDenoms, denomSeparator)

	faucetOptions := make([]customFaucet.Option, len(coins))
	for i, coin := range coins {
		amount, _ := bigNumber.FromString(creditAmount)
		credit, _ := bigNumber.FromString(maxCredit)
		faucetOptions[i] = customFaucet.Coin(amount, credit, coin)
	}

	faucetOptions = append(faucetOptions, customFaucet.Account(keyName, keyMnemonic))

	faucet, err := customFaucet.New(context.Background(), cr, faucetOptions...)
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

		var req TransferRequestWithFees
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
			cosmosCoin, _ := bigNumber.FromString(strings.Split(coin, defaultDenoms)[0])
			credit, _ := bigNumber.FromString(maxCredit)

			if err == nil {
				if cosmosCoin.Cmp(credit) > 0 {
					var transfers []customFaucet.Transfer
					t := customFaucet.Transfer{
						Coin:   defaultDenoms,
						Status: "error",
						Error:  fmt.Sprintf("maximum credit (%s)", maxCredit),
					}

					transfers = append(transfers, t)

					xhttp.ResponseJSON(w, http.StatusOK, customFaucet.TransferResponse{
						Transfers: transfers,
					})
					return
				}

				// Update the request body with fees
				req.Fees = fees + defaultDenoms
				buf, _ = json.Marshal(req)
				r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))

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
