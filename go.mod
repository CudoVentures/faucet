module github.com/tendermint/faucet

go 1.16

require (
	github.com/cosmos/cosmos-sdk v0.44.5
	github.com/sirupsen/logrus v1.8.1
	github.com/tendermint/starport v0.18.6
	google.golang.org/genproto v0.0.0-20210903162649-d08c68adba83
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
