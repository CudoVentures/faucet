module github.com/tendermint/faucet

go 1.16

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/sirupsen/logrus v1.8.0
	github.com/tendermint/starport v0.15.0
	golang.org/x/mod v0.4.2 // indirect
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5 // indirect
	golang.org/x/sys v0.0.0-20210525143221-35b2ab0089ea // indirect
	google.golang.org/genproto v0.0.0-20210524171403-669157292da3
	google.golang.org/grpc v1.38.0 // indirect
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4
