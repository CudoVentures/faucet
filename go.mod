module github.com/tendermint/faucet

go 1.16

replace github.com/cosmos/cosmos-sdk => github.com/CudoVentures/cosmos-sdk v0.0.0-20220111092913-4117cd46b688

require (
	github.com/ghodss/yaml v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/joho/godotenv v1.5.1
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/sirupsen/logrus v1.8.1
	github.com/tendermint/spm v0.1.8
	github.com/tendermint/starport v0.18.6
	lukechampine.com/uint128 v1.2.0
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
