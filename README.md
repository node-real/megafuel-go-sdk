# megafuel-go-sdk

This Golang SDK is thin wrapper of MegaFuel clients, offering a streamlined interface to interact with the [MegaFuel](https://docs.nodereal.io/docs/megafuel-overview).

## Network Endpoint

|    Network    |        [Paymaster]( https://docs.nodereal.io/reference/pm-issponsorable)        |                [Sponsor](https://docs.nodereal.io/reference/pm-addtowhitelist)                 |
|:-------------:|:-------------------------------------------------------------------------------:|:----------------------------------------------------------------------------------------------:|
|  BSC mainnet  |                        https://bsc-megafuel.nodereal.io                         |                   https://open-platform-ap.nodereal.io/{YOUR_API_KEY}/megafuel                    |
|  BSC testnet  |                    https://bsc-megafuel-testnet.nodereal.io                     |               https://open-platform-ap.nodereal.io/{YOUR_API_KEY}/megafuel-testnet                |
| opBNB mainnet |                       https://opbnb-megafuel.nodereal.io                        |                   https://open-platform-ap.nodereal.io/{YOUR_API_KEY}/megafuel                    |
| opBNB testnet |                   https://opbnb-megafuel-testnet.nodereal.io                    |               https://open-platform-ap.nodereal.io/{YOUR_API_KEY}/megafuel-testnet                |


## Quick Start

1. Install dependency

```shell
$ go get -u github.com/nodereal/megafuel-go-sdk
```

2. Example

```go
package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gofrs/uuid"
	"github.com/node-real/megafuel-go-sdk/pkg/paymasterclient"
	"github.com/node-real/megafuel-go-sdk/pkg/sponsorclient"
)

const PAYMASTER_URL = "https://bsc-megafuel-testnet.nodereal.io"
const CHAIN_URL = "https://data-seed-prebsc-2-s1.binance.org:8545/"
const SPONSOR_URL = "https://open-platform-ap.nodereal.io/<api-key>/megafuel-testnet"

const POLICY_UUID = "a2381160-xxxx-xxxx-xxxxceca86556834"
const RECIPIENT_ADDRESS = "0x8e9......3EA2"
const YOUR_PRIVATE_KEY = "69......929"

func main() {
	sponsorClient, err := sponsorclient.New(context.Background(), SPONSOR_URL)
	if err != nil {
		panic(err)
	}

	policyUUID, _ := uuid.FromString(POLICY_UUID)

	success, err := sponsorClient.AddToWhitelist(context.Background(), sponsorclient.WhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Values:        []string{RECIPIENT_ADDRESS},
	})
	if err != nil || !success {
		panic("failed to add token contract whitelist")
	}

	// Connect to an Ethereum node (for transaction assembly)
	client, err := ethclient.Dial(CHAIN_URL)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum network: %v", err)
	}
	// Create a PaymasterClient (for transaction sending)
	paymasterClient, err := paymasterclient.New(context.Background(), PAYMASTER_URL)
	if err != nil {
		log.Fatalf("Failed to create PaymasterClient: %v", err)
	}

	// Load your private key
	privateKey, err := crypto.HexToECDSA(YOUR_PRIVATE_KEY)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get the latest nonce for the from address
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	toAddress := common.HexToAddress(RECIPIENT_ADDRESS)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(0),
		Gas:      21000,
		To:       &toAddress,
		Value:    big.NewInt(1e18),
	})

	// Convert to Transaction struct for IsSponsorable check
	gasLimit := tx.Gas()
	sponsorableTx := paymasterclient.TransactionArgs{
		To:    &toAddress,
		From:  fromAddress,
		Value: (*hexutil.Big)(big.NewInt(1e18)),
		Gas:   (*hexutil.Uint64)(&gasLimit),
	}

	// Check if the transaction is sponsorable
	sponsorableInfo, err := paymasterClient.IsSponsorable(context.Background(), sponsorableTx)
	if err != nil {
		log.Fatalf("Error checking sponsorable status: %v", err)
	}

	fmt.Printf("Sponsorable Information:\n%+v\n", sponsorableInfo)

	if sponsorableInfo.Sponsorable {
		// Get the chain ID
		chainID, err := client.ChainID(context.Background())
		if err != nil {
			log.Fatalf("Failed to get chain ID: %v", err)
		}

		// Sign the transaction
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			log.Fatalf("Failed to sign transaction: %v", err)
		}

		txInput, err := signedTx.MarshalBinary()
		if err != nil {
			log.Fatalf("Failed to marshal transaction: %v", err)
		}

		// Send the transaction using PaymasterClient
		_, err = paymasterClient.SendRawTransaction(context.Background(), txInput)
		if err != nil {
			log.Fatalf("Failed to send sponsorable transaction: %v", err)
		}
		fmt.Printf("Sponsorable transaction sent: %s\n", signedTx.Hash())
	} else {
		fmt.Println("Transaction is not sponsorable. You may need to send it as a regular transaction.")
	}
}
```

More examples can be found in the [examples](https://github.com/node-real/megafuel-client-example).

