module github.com/smartcontractkit/substrate-adapter

go 1.13

require (
	github.com/centrifuge/go-substrate-rpc-client v1.1.0
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/ethereum/go-ethereum v1.9.21 // indirect
	github.com/gin-gonic/gin v1.5.0
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/pierrec/xxHash v0.1.5 // indirect
	github.com/pkg/errors v0.9.1
	github.com/robertkrimen/otto v0.0.0-20170205013659-6a77b7cbc37d // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/stretchr/testify v1.5.1
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
)

replace github.com/centrifuge/go-substrate-rpc-client => github.com/LaurentTrk/go-substrate-rpc-client v2.0.0-alpha.6.4+incompatible
