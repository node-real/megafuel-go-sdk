package test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/node-real/megafuel-go-sdk/pkg/paymasterclient"
)

const PAYMASTER_URL = "https://bsc-megafuel-testnet.nodereal.io/97"

var log = logrus.New()

func paymasterSetup(t *testing.T) (*ethclient.Client, paymasterclient.Client, string, error) {
	t.Helper()

	key := os.Getenv("OPEN_PLATFORM_PRIVATE_KEY")
	if key == "" {
		log.Fatal("Environment variable OPEN_PLATFORM_PRIVATE_KEY is not set")
	}

	yourPrivateKey := os.Getenv("YOUR_PRIVATE_KEY")
	if yourPrivateKey == "" {
		log.Fatal("Environment variable YOUR_PRIVATE_KEY is not set")
	}

	// Connect to an Ethereum node (for transaction assembly)
	client, err := ethclient.Dial(fmt.Sprintf("https://bsc-testnet.nodereal.io/v1/%s", key))
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum network: %v", err)
	}

	// Create a PaymasterClient (for transaction sending)
	paymasterClient, err := paymasterclient.New(context.Background(), PAYMASTER_URL)
	if err != nil {
		log.Fatalf("Failed to create PaymasterClient: %v", err)
	}

	return client, paymasterClient, yourPrivateKey, nil
}

func TestPaymasterAPI(t *testing.T) {
	client, paymasterClient, yourPrivateKey, err := paymasterSetup(t)
	require.NoError(t, err, "failed to set up paymaster")

	privateKey, err := crypto.HexToECDSA(yourPrivateKey)
	require.NoError(t, err, "Failed to load private key")

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err, "Failed to get nonce")

	toAddress := common.HexToAddress(RECIPIENT_ADDRESS)

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(0),
		Gas:      21000,
		To:       &toAddress,
		Value:    big.NewInt(0),
	})

	gasLimit := tx.Gas()
	sponsorableTx := paymasterclient.TransactionArgs{
		To:    &toAddress,
		From:  fromAddress,
		Value: (*hexutil.Big)(big.NewInt(0)),
		Gas:   (*hexutil.Uint64)(&gasLimit),
		Data:  &hexutil.Bytes{},
	}

	sponsorableInfo, err := paymasterClient.IsSponsorable(context.Background(), sponsorableTx)
	require.NoError(t, err, "Error checking sponsorable status")
	require.True(t, sponsorableInfo.Sponsorable)

	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err, "Failed to get chain ID")

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	require.NoError(t, err, "Failed to sign transaction")

	txInput, err := signedTx.MarshalBinary()
	require.NoError(t, err, "Failed to marshal transaction")

	paymasterTx, err := paymasterClient.SendRawTransaction(context.Background(), txInput)
	require.NoError(t, err, "Failed to send sponsorable transaction")
	log.Infof("Sponsorable transaction sent: %s", signedTx.Hash())
	log.Info("Waiting for transaction confirmation")
	time.Sleep(5 * time.Second)

	payMasterChainID, err := paymasterClient.ChainID(context.Background())
	require.NoError(t, err, "failed to get paymaster chain id")
	assert.Equal(t, payMasterChainID, "0x61")

	txResp, err := paymasterClient.GetGaslessTransactionByHash(context.Background(), paymasterTx)
	require.NoError(t, err, "failed to GetGaslessTransactionByHash")
	assert.Equal(t, txResp.TxHash.String(), paymasterTx.String())

	bundleUuid := txResp.BundleUUID
	sponsorTx, err := paymasterClient.GetSponsorTxByBundleUUID(context.Background(), bundleUuid)
	require.NoError(t, err)

	bundle, err := paymasterClient.GetBundleByUUID(context.Background(), bundleUuid)
	require.NoError(t, err)

	sponsorTx, err = paymasterClient.GetSponsorTxByTxHash(context.Background(), sponsorTx.TxHash)
	require.NoError(t, err)

	log.Infof("Bundle UUID: %s", bundle.BundleUUID)
	blockNumber := rpc.PendingBlockNumber
	count, err := paymasterClient.GetTransactionCount(context.Background(), common.HexToAddress(RECIPIENT_ADDRESS), rpc.BlockNumberOrHash{BlockNumber: &blockNumber})
	require.NoError(t, err, "failed to GetTransactionCount")
	assert.Greater(t, *count, hexutil.Uint64(0))
}
