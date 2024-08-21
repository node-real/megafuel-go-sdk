module github.com/node-real/megafuel-go-sdk

go 1.21

require (
	github.com/ethereum/go-ethereum v1.14.8
	github.com/gofrs/uuid v4.3.0+incompatible
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.3.4 // indirect
	github.com/crate-crypto/go-kzg-4844 v1.0.0 // indirect
	github.com/deckarep/golang-set/v2 v2.6.0 // indirect
	github.com/ethereum/c-kzg-4844 v1.0.0 // indirect
	github.com/go-ole/go-ole v1.3.0 // indirect
	github.com/gorilla/websocket v1.5.1 // indirect
	github.com/holiman/uint256 v1.3.1 // indirect
	github.com/panjf2000/ants/v2 v2.4.5 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.13 // indirect
	github.com/tklauser/numcpus v0.7.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.3 // indirect
	golang.org/x/crypto v0.22.0 // indirect
	golang.org/x/exp v0.0.0-20240213143201-ec583247a57a // indirect
	golang.org/x/net v0.23.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
)

replace (
	github.com/cometbft/cometbft => github.com/bnb-chain/greenfield-tendermint v0.0.0-20230417032003-4cda1f296fb2
	github.com/ethereum/go-ethereum => github.com/bnb-chain/bsc v1.4.13
	github.com/syndtr/goleveldb v1.0.1 => github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/tendermint/tendermint => github.com/bnb-chain/tendermint v0.31.16
)
