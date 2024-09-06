package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gofrs/uuid"

	"github.com/node-real/megafuel-go-sdk/pkg/paymasterclient"
	"github.com/node-real/megafuel-go-sdk/pkg/sponsorclient"
)

const PAYMASTER_URL = "https://opbnb-megafuel.nodereal.io/204"
const CHAIN_URL = "https://opbnb-mainnet-rpc.bnbchain.org"
const POLICY_UUID = "72191372-5550-4cf6-956e-b70d1e4786cf"
const RECIPIENT_ADDRESS = "0xDfbA0Ce6349C7205C8951304a67f36F65EBc1B2e"

func main() {
	yourPrivateKey := os.Getenv("YOUR_PRIVATE_KEY")
	if yourPrivateKey == "" {
		log.Fatal("Environment variable YOUR_PRIVATE_KEY is not set")
	}

	sponsorURL := os.Getenv("SPONSOR_URL")
	if sponsorURL == "" {
		log.Fatal("Environment variable SPONSOR_URL is not set")
	}
	sponsorClient, err := sponsorclient.New(context.Background(), sponsorURL)
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
	privateKey, err := crypto.HexToECDSA(yourPrivateKey)
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
		Value:    big.NewInt(0),
	})

	// Convert to Transaction struct for IsSponsorable check
	gasLimit := tx.Gas()
	sponsorableTx := paymasterclient.TransactionArgs{
		To:    &toAddress,
		From:  fromAddress,
		Value: (*hexutil.Big)(big.NewInt(0)),
		Gas:   (*hexutil.Uint64)(&gasLimit),
		Data:  &hexutil.Bytes{},
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

	chainID, err := paymasterClient.ChainID(context.Background())
	if err != nil {
		log.Fatalf("failed to get chain id err:%v", err)
	}
	fmt.Println(chainID)
	txResp, err := paymasterClient.GetGaslessTransactionByHash(context.Background(), common.HexToHash("0x3f4dc87385533e4134ed96c986fc841b7e291ae92898e03e79fc6d229d68afa9"))
	if err != nil {
		log.Fatalf("failed to GetGaslessTransactionByHash err:%v", err)
	}
	fmt.Println(txResp.TxHash, txResp.BundleUUID)

	sponsorTx, err := paymasterClient.GetSponsorTxByTxHash(context.Background(), common.HexToHash("0x970ce1f01ef50fcc5bcbbaadf37c21b2f49551df641940b99b1be066577d179f"))
	if err != nil {
		log.Fatalf("failed to GetSponsorTxByTxHash err:%v", err)
	}
	fmt.Println(sponsorTx.TxHash)

	bundleUuid := txResp.BundleUUID
	sponsorTx, err = paymasterClient.GetSponsorTxByBundleUUID(context.Background(), bundleUuid)
	if err != nil {
		log.Fatalf("failed to GetSponsorTxByBundleUUID err:%v", err)
	}
	fmt.Println(sponsorTx.TxHash)

	bundle, err := paymasterClient.GetBundleByUUID(context.Background(), bundleUuid)
	if err != nil {
		log.Fatalf("failed to GetBundleByUUID err:%v", err)
	}
	fmt.Println(bundle.BundleUUID)
	blockNumber := rpc.PendingBlockNumber
	count, err := paymasterClient.GetTransactionCount(context.Background(), common.HexToAddress("0x04d63aBCd2b9b1baa327f2Dda0f873F197ccd186"), rpc.BlockNumberOrHash{BlockNumber: &blockNumber})
	if err != nil {
		log.Fatalf("failed to GetTransactionCount err:%v", err)
	}
	fmt.Println(count.String())

	res, err := sponsorClient.AddToWhitelist(context.Background(), sponsorclient.WhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Values:        []string{"0x04d63aBCd2b9b1baa327f2Dda0f873F197ccd186"},
	})
	if err != nil {
		log.Fatalf("failed to AddToWhitelist err:%v", err)
	}
	fmt.Println(res)

	wl, err := sponsorClient.GetWhitelist(context.Background(), sponsorclient.GetWhitelistArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Offset:        0,
		Limit:         2,
	})
	if err != nil {
		log.Fatalf("failed to GetWhitelist err:%v", err)
	}
	fmt.Println(wl)

	res, err = sponsorClient.RmFromWhitelist(context.Background(), sponsorclient.WhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Values:        []string{"0x04d63aBCd2b9b1baa327f2Dda0f873F197ccd186"},
	})
	if err != nil {
		log.Fatalf("failed to RmFromWhitelist err:%v", err)
	}
	fmt.Println(res)

	res, err = sponsorClient.EmptyWhitelist(context.Background(), sponsorclient.EmptyWhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
	})
	if err != nil {
		log.Fatalf("failed to EmptyWhitelist err:%v", err)
	}
	fmt.Println(res)

	pUUID, _ := uuid.FromString("7cb16eee-3a95-4d41-b280-41955e617a36")
	UserSpendData, err := sponsorClient.GetUserSpendData(context.Background(), common.HexToAddress("0x04d63aBCd2b9b1baa327f2Dda0f873F197ccd186"), pUUID)
	if err != nil {
		log.Fatalf("failed to GetUserSpendData err:%v", err)
	}
	fmt.Println(UserSpendData.GasCost.Raw().String())

	PolicySpendData, err := sponsorClient.GetPolicySpendData(context.Background(), pUUID)
	if err != nil {
		log.Fatalf("failed to GetPolicySpendData err:%v", err)
	}
	fmt.Println(PolicySpendData.Cost.Raw().String())
}
