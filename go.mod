module github.com/tendermint/faucet

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/tendermint/starport v0.16.1
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
