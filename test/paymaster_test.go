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

const (
	PAYMASTER_URL  = "https://bsc-megafuel-testnet.nodereal.io/"
	PRIVATE_POLICY = "90f1ba4c-1f93-4759-b8a9-da4d59c668b4"
)

var log = logrus.New()

// setupCommon contains the common setup logic for both paymaster types
func setupCommon(t *testing.T) (*ethclient.Client, string, string, error) {
	t.Helper()

	key := os.Getenv("OPEN_PLATFORM_PRIVATE_KEY")
	if key == "" {
		return nil, "", "", fmt.Errorf("environment variable OPEN_PLATFORM_PRIVATE_KEY is not set")
	}

	yourPrivateKey := os.Getenv("YOUR_PRIVATE_KEY")
	if yourPrivateKey == "" {
		return nil, "", "", fmt.Errorf("environment variable YOUR_PRIVATE_KEY is not set")
	}

	// Connect to an Ethereum node (for transaction assembly)
	client, err := ethclient.Dial(fmt.Sprintf("https://bsc-testnet.nodereal.io/v1/%s", key))
	if err != nil {
		return nil, "", "", fmt.Errorf("failed to connect to the Ethereum network: %v", err)
	}

	return client, key, yourPrivateKey, nil
}

// paymasterSetup initializes a standard paymaster client using the environment variable.
func paymasterSetup(t *testing.T) (*ethclient.Client, paymasterclient.Client, string, error) {
	client, _, yourPrivateKey, err := setupCommon(t)
	if err != nil {
		return nil, nil, "", err
	}

	// Create a PaymasterClient (for transaction sending)
	paymasterClient, err := paymasterclient.New(context.Background(), PAYMASTER_URL)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create PaymasterClient: %v", err)
	}

	return client, paymasterClient, yourPrivateKey, nil
}

// privatePaymasterSetup initializes a private paymaster client using the environment variable.
func privatePaymasterSetup(t *testing.T) (*ethclient.Client, paymasterclient.Client, string, error) {
	client, key, yourPrivateKey, err := setupCommon(t)
	if err != nil {
		return nil, nil, "", err
	}

	sponsorURL := fmt.Sprintf("https://open-platform-ap.nodereal.io/%s/megafuel-testnet/97", key)
	// Create a Private PaymasterClient (for transaction sending)
	paymasterClient, err := paymasterclient.NewPrivatePaymaster(context.Background(), sponsorURL, PRIVATE_POLICY)
	if err != nil {
		return nil, nil, "", fmt.Errorf("failed to create Private PaymasterClient: %v", err)
	}

	return client, paymasterClient, yourPrivateKey, nil
}

// TestPaymasterAPI tests the critical functionalities related to the Paymaster API.
func TestPaymasterAPI(t *testing.T) {
	// Setup Ethereum client and Paymaster client. Ensure no errors during the setup.
	client, paymasterClient, yourPrivateKey, err := paymasterSetup(t)
	require.NoError(t, err, "failed to set up paymaster")

	// Convert the private key from hex string to ECDSA format and check for errors.
	privateKey, err := crypto.HexToECDSA(yourPrivateKey)
	require.NoError(t, err, "Failed to load private key")

	// Extract the public key from the private key and assert type casting to ECDSA.
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Fetch the current nonce for the account to ensure the transaction can be processed sequentially.
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err, "Failed to get nonce")

	// Define the recipient Ethereum address.
	toAddress := common.HexToAddress(RECIPIENT_ADDRESS)

	// Construct a new Ethereum transaction.
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(0),
		Gas:      21000,
		To:       &toAddress,
		Value:    big.NewInt(0),
	})

	// Prepare a transaction argument for checking if it's sponsorable.
	gasLimit := tx.Gas()
	sponsorableTx := paymasterclient.TransactionArgs{
		To:    &toAddress,
		From:  fromAddress,
		Value: (*hexutil.Big)(big.NewInt(0)),
		Gas:   (*hexutil.Uint64)(&gasLimit),
		Data:  &hexutil.Bytes{},
	}

	// Verify if the transaction can be sponsored under the current policy.
	sponsorableInfo, err := paymasterClient.IsSponsorable(context.Background(), sponsorableTx)
	require.NoError(t, err, "Error checking sponsorable status")
	require.True(t, sponsorableInfo.Sponsorable)

	// Retrieve the blockchain ID to ensure that the transaction is signed correctly.
	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err, "Failed to get chain ID")

	// Sign the transaction using the provided private key and the current chain ID.
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	require.NoError(t, err, "Failed to sign transaction")

	// Marshal the signed transaction into a binary format for transmission.
	txInput, err := signedTx.MarshalBinary()
	require.NoError(t, err, "Failed to marshal transaction")

	// Send the signed transaction and check for successful submission.
	paymasterTx, err := paymasterClient.SendRawTransaction(context.Background(), txInput, nil)
	require.NoError(t, err, "Failed to send sponsorable transaction")
	log.Infof("Sponsorable transaction sent: %s", signedTx.Hash())
	log.Info("Waiting for transaction confirmation")
	time.Sleep(5 * time.Second) // Consider replacing with a non-blocking wait or event-driven notification.

	// Check the Paymaster client's chain ID for consistency.
	payMasterChainID, err := paymasterClient.ChainID(context.Background())
	require.NoError(t, err, "failed to get paymaster chain id")
	assert.Equal(t, payMasterChainID.String(), "97")

	// Retrieve and verify the transaction details by its hash.
	txResp, err := paymasterClient.GetGaslessTransactionByHash(context.Background(), paymasterTx)
	require.NoError(t, err, "failed to GetGaslessTransactionByHash")
	assert.Equal(t, txResp.TxHash.String(), paymasterTx.String())

	// Check for the related transaction bundle based on the UUID.
	bundleUuid := txResp.BundleUUID
	sponsorTx, err := paymasterClient.GetSponsorTxByBundleUUID(context.Background(), bundleUuid)
	require.NoError(t, err)

	// Retrieve the full bundle using the UUID and verify its existence.
	bundle, err := paymasterClient.GetBundleByUUID(context.Background(), bundleUuid)
	require.NoError(t, err)

	// Further validate the bundle by fetching the transaction via its hash.
	sponsorTx, err = paymasterClient.GetSponsorTxByTxHash(context.Background(), sponsorTx.TxHash)
	require.NoError(t, err)

	// Log the UUID of the bundle for reference.
	log.Infof("Bundle UUID: %s", bundle.BundleUUID)

	// Obtain and verify the transaction count for the recipient address.
	blockNumber := rpc.PendingBlockNumber
	count, err := paymasterClient.GetTransactionCount(context.Background(), common.HexToAddress(RECIPIENT_ADDRESS), rpc.BlockNumberOrHash{BlockNumber: &blockNumber})
	require.NoError(t, err, "failed to GetTransactionCount")
	assert.Greater(t, count, hexutil.Uint64(0))
}

// TestPaymasterAPI tests the critical functionalities related to the Paymaster API.
func TestPrivatePolicyGaslessTransaction(t *testing.T) {
	// Setup Ethereum client and Paymaster client. Ensure no errors during the setup.
	client, sponsorClient, yourPrivateKey, err := privatePaymasterSetup(t)
	require.NoError(t, err, "failed to set up paymaster")

	// Convert the private key from hex string to ECDSA format and check for errors.
	privateKey, err := crypto.HexToECDSA(yourPrivateKey)
	require.NoError(t, err, "Failed to load private key")

	// Extract the public key from the private key and assert type casting to ECDSA.
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Fetch the current nonce for the account to ensure the transaction can be processed sequentially.
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	require.NoError(t, err, "Failed to get nonce")

	// Define the recipient Ethereum address.
	toAddress := common.HexToAddress(RECIPIENT_ADDRESS)

	// Construct a new Ethereum transaction.
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(0),
		Gas:      21000,
		To:       &toAddress,
		Value:    big.NewInt(0),
	})

	// Prepare a transaction argument for checking if it's sponsorable.
	gasLimit := tx.Gas()

	privatePolicySponsorableTx := paymasterclient.TransactionArgs{
		To:    &toAddress,
		From:  fromAddress,
		Value: (*hexutil.Big)(big.NewInt(0)),
		Gas:   (*hexutil.Uint64)(&gasLimit),
		Data:  &hexutil.Bytes{},
	}

	privatePolicySponsorableInfo, err := sponsorClient.IsSponsorable(context.Background(), privatePolicySponsorableTx)
	require.NoError(t, err, "Error checking sponsorable private policy status")
	require.True(t, privatePolicySponsorableInfo.Sponsorable)

	// Retrieve the blockchain ID to ensure that the transaction is signed correctly.
	chainID, err := client.ChainID(context.Background())
	require.NoError(t, err, "Failed to get chain ID")

	// Sign the transaction using the provided private key and the current chain ID.
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	require.NoError(t, err, "Failed to sign transaction")

	// Marshal the signed transaction into a binary format for transmission.
	txInput, err := signedTx.MarshalBinary()
	require.NoError(t, err, "Failed to marshal transaction")

	transaction, err := sponsorClient.SendRawTransaction(context.Background(), txInput, &paymasterclient.SendRawTransactionOptions{UserAgent: "Test User Agent"})
	require.NoError(t, err, "Failed to send sponsorable private policy transaction")
	log.Infof("Sponsorable private policy transaction sent: %s", signedTx.Hash())
	time.Sleep(10 * time.Second) // Consider replacing with a non-blocking wait or event-driven notification.

	// Check the Paymaster client's chain ID for consistency.
	payMasterChainID, err := sponsorClient.ChainID(context.Background())
	require.NoError(t, err, "failed to get paymaster chain id")
	assert.Equal(t, payMasterChainID.String(), "97")

	// Retrieve and verify the transaction details by its hash.
	txResp, err := sponsorClient.GetGaslessTransactionByHash(context.Background(), transaction)
	require.NoError(t, err, "failed to GetGaslessTransactionByHash")
	assert.Equal(t, txResp.TxHash.String(), transaction.String())

	// Check for the related transaction bundle based on the UUID.
	bundleUuid := txResp.BundleUUID
	sponsorTx, err := sponsorClient.GetSponsorTxByBundleUUID(context.Background(), bundleUuid)
	require.NoError(t, err)

	// Retrieve the full bundle using the UUID and verify its existence.
	bundle, err := sponsorClient.GetBundleByUUID(context.Background(), bundleUuid)
	require.NoError(t, err)

	// Further validate the bundle by fetching the transaction via its hash.
	sponsorTx, err = sponsorClient.GetSponsorTxByTxHash(context.Background(), sponsorTx.TxHash)
	require.NoError(t, err)

	// Log the UUID of the bundle for reference.
	log.Infof("Bundle UUID: %s", bundle.BundleUUID)

	// Obtain and verify the transaction count for the recipient address.
	blockNumber := rpc.PendingBlockNumber
	count, err := sponsorClient.GetTransactionCount(context.Background(), common.HexToAddress(RECIPIENT_ADDRESS), rpc.BlockNumberOrHash{BlockNumber: &blockNumber})
	require.NoError(t, err, "failed to GetTransactionCount")
	assert.Greater(t, count, hexutil.Uint64(0))
}
