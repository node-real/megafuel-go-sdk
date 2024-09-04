# megafuel-go-sdk

This is Golang SDK for the MegaFuel clients, it provides a simple way to interact with the MegaFuel.

For more information, please refer to the [API documentation](https://docs.nodereal.io/docs/megafuel-api).

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
	"encoding/json"
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
const SPONSOR_URL = "https://open-platform.nodereal.io/<api-key>/megafuel-testnet"

const POLICY_UUID = "a2381160-xxxx-xxxx-xxxxceca86556834"
const TOKEN_CONTRACT_ADDRESS = "0xeD2.....12Ee"
const RECIPIENT_ADDRESS = "0x8e9......3EA2"
const YOUR_PRIVATE_KEY = "69......929"

func createERC20TransferData(to common.Address, amount *big.Int) ([]byte, error) {
	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]
	paddedAddress := common.LeftPadBytes(to.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	return data, nil
}

func main() {
	sponsorClient, err := sponsorclient.New(context.Background(), SPONSOR_URL)
	if err != nil {
		panic(err)
	}

	policyUUID, _ := uuid.FromString(POLICY_UUID)

	success, err := sponsorClient.AddToWhitelist(context.Background(), sponsorclient.WhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Values:        []string{TOKEN_CONTRACT_ADDRESS},
	})
	if err != nil || !success {
		panic("failed to add token contract whitelist")
	}

	println(success)

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
	// Amount of tokens to transfer (adjust based on token decimals)
	amount := big.NewInt(1000000000000000000) // 1 token for a token with 18 decimals

	// Create ERC20 transfer data
	data, err := createERC20TransferData(common.HexToAddress(RECIPIENT_ADDRESS), amount)
	if err != nil {
		log.Fatalf("Failed to create ERC20 transfer data: %v", err)
	}

	// Get the latest nonce for the from address
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("Failed to get nonce: %v", err)
	}

	tokenContractAddress := common.HexToAddress(TOKEN_CONTRACT_ADDRESS)
	// Create the transaction
	gasPrice := big.NewInt(0)
	tx := types.NewTransaction(nonce, tokenContractAddress, big.NewInt(0), 300000, gasPrice, data)

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

	// Convert to Transaction struct for IsSponsorable check
	gasLimit := tx.Gas()
	sponsorableTx := paymasterclient.TransactionArgs{
		To:    &tokenContractAddress,
		From:  fromAddress,
		Value: (*hexutil.Big)(big.NewInt(0)),
		Gas:   (*hexutil.Uint64)(&gasLimit),
		Data:  (*hexutil.Bytes)(&data),
	}

	// Check if the transaction is sponsorable
	sponsorableInfo, err := paymasterClient.IsSponsorable(context.Background(), sponsorableTx)
	if err != nil {
		log.Fatalf("Error checking sponsorable status: %v", err)
	}

	jsonInfo, _ := json.MarshalIndent(sponsorableInfo, "", "  ")
	fmt.Printf("Sponsorable Information:\n%s\n", string(jsonInfo))

	if sponsorableInfo.Sponsorable {
		// Send the transaction using PaymasterClient
		_, err := paymasterClient.SendRawTransaction(context.Background(), txInput)
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

