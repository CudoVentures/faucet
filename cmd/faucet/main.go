package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/tendermint/starport/starport/pkg/chaincmd"
	chaincmdrunner "github.com/tendermint/starport/starport/pkg/chaincmd/runner"
	"github.com/tendermint/starport/starport/pkg/cosmoscoin"
	"github.com/tendermint/starport/starport/pkg/cosmosfaucet"
	"github.com/tendermint/starport/starport/pkg/cosmosver"
	"github.com/tendermint/starport/starport/pkg/xhttp"

	recaptchapb "google.golang.org/genproto/googleapis/cloud/recaptchaenterprise/v1"
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
	CaptchaResponse string   `json: "captchaResponse"`
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
			chaincmd.WithVersion(cosmosver.StargateZeroFourtyAndAbove),
		)
	} else {
		ccoptions = append(ccoptions,
			chaincmd.WithVersion(cosmosver.LaunchpadAny),
			chaincmd.WithLaunchpadCLI(appCli),
		)
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
		if originHeader == "http://localhost:3000" || originHeader == "http://104.154.17.186:3000" || originHeader == "http://explorer.cudos.org" || originHeader == "https://explorer.cudos.org" {
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

		if err == nil {
			captchaErr := checkCaptchaWithKey(req.CaptchaResponse)

			if captchaErr != nil {
				http.Error(w, "Wrong captcha", http.StatusUnauthorized)
				return
			}

			coin := req.Coins[0]
			amount, denom, err := cosmoscoin.Parse(coin)
			if err == nil {
				if amount > maxCredit {
					var transfers []cosmosfaucet.Transfer
					t := cosmosfaucet.Transfer{
						Coin:   denom,
						Status: "error",
						Error:  fmt.Sprintf("max credit (%d)", maxCredit),
					}

					transfers = append(transfers, t)

					xhttp.ResponseJSON(w, http.StatusOK, cosmosfaucet.TransferResponse{
						Transfers: transfers,
					})
					return
				}

				bigCoinAmount := new(big.Int).SetUint64(amount)
				bigCoinMultiplier := new(big.Int).SetUint64(1000000000000)
				bigCoinAmount.Mul(bigCoinAmount, bigCoinMultiplier)

				req.Coins[0] = fmt.Sprintf("%sacudos", bigCoinAmount.String())
				reqBytes, _ := json.Marshal(req)
				rdr2 := ioutil.NopCloser(bytes.NewBuffer(reqBytes))
				r.Body = rdr2
				faucet.ServeHTTP(w, r)
			}
		}
	})
	log.Infof("listening on :%d", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func checkCaptchaWithKey(captcha string) error {
	event := &recaptchapb.Event{
		Token:          captcha,
		SiteKey:        captchSiteKey,
		ExpectedAction: "login",
	}

	assessment := &recaptchapb.Assessment{
		Event: event,
	}

	reqJson, err := json.Marshal(assessment)
	if err != nil {
		return err
	}

	siteVerifyURL := "https://recaptchaenterprise.googleapis.com/v1beta1/projects/" + googleProjectId + "/assessments?key=" + googleApiKey
	req, err := http.NewRequest(http.MethodPost, siteVerifyURL, bytes.NewBuffer(reqJson))
	if err != nil {
		return err
	}

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

	fmt.Printf("Captcha score %f\n", body.Score)
	if body.Score < 0.3 {
		return errors.New("invalid captcha")
	}

	return nil
}
