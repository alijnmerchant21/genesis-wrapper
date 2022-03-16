module github.com/crescent-network/genesis-wrapper

go 1.17

require (
	github.com/cosmos/cosmos-sdk v0.44.5
	github.com/crescent-network/crescent v0.0.0-20220316143238-14f27e2b6364
	github.com/spf13/cobra v1.2.1
	github.com/tendermint/budget v1.1.0
	github.com/tendermint/tendermint v0.34.15
)

replace (
	github.com/cosmos/cosmos-sdk => github.com/crescent-network/cosmos-sdk v1.0.2-sdk-0.44.5
	github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
	github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4
	google.golang.org/grpc => google.golang.org/grpc v1.33.2
)
