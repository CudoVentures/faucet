package customFaucet

import (
	"context"
	"fmt"
	"strings"
	"time"

	customChaincmdrunner "github.com/tendermint/faucet/customChaincmdrunner"
	bigNumber "lukechampine.com/uint128"
)

// TotalTransferredAmount returns the total transferred amount from faucet account to toAccountAddress.
func (f Faucet) TotalTransferredAmount(ctx context.Context, toAccountAddress, denom string) (amount bigNumber.Uint128, error error) {
	fromAccount, err := f.runner.ShowAccount(ctx, f.accountName)
	if err != nil {
		num, _ := bigNumber.FromString("0")
		return num, err
	}

	events, err := f.runner.QueryTxEvents(ctx,
		customChaincmdrunner.NewEventSelector("message", "sender", fromAccount.Address),
		customChaincmdrunner.NewEventSelector("transfer", "recipient", toAccountAddress))
	if err != nil {
		num, _ := bigNumber.FromString("0")
		return num, err
	}

	for _, event := range events {
		if event.Type == "transfer" {
			for _, attr := range event.Attributes {
				if attr.Key == "amount" {
					if !strings.HasSuffix(attr.Value, denom) {
						continue
					}

					if time.Since(event.Time) < f.limitRefreshWindow {
						amountStr := strings.TrimRight(attr.Value, denom)
						num, error := bigNumber.FromString(amountStr)
						if error == nil {
							amount = amount.Add(num)
						}
					}
				}
			}
		}
	}

	return amount, nil
}

// Transfer transfer amount of tokens from the faucet account to toAccountAddress.
func (f Faucet) Transfer(ctx context.Context, toAccountAddress string, amount bigNumber.Uint128, denom string, fees string) error {
	amountToTransfer := amount.String() + denom

	totalSent, err := f.TotalTransferredAmount(ctx, toAccountAddress, denom)
	if err != nil {
		return err
	}

	if !f.coinsMax[denom].IsZero() {
		if totalSent.Cmp(f.coinsMax[denom]) > 0 {
			return fmt.Errorf("account has reached maximum credit allowed per account (%d)", f.coinsMax[denom])
		}

		if totalSent.Add(amount).Cmp(f.coinsMax[denom]) > 0 {
			return fmt.Errorf("account is about to reach maximum credit allowed per account. it can only receive up to (%d) in total", f.coinsMax[denom])
		}
	}

	fromAccount, err := f.runner.ShowAccount(ctx, f.accountName)
	if err != nil {
		return err
	}

	return f.runner.BankSend(ctx, fromAccount.Address, toAccountAddress, amountToTransfer, fees)
}
