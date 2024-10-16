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
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/node-real/megafuel-go-sdk/pkg/paymasterclient"
	"github.com/node-real/megafuel-go-sdk/pkg/sponsorclient"
)

// Constants for API
const (
	POLICY_UUID       = "72191372-5550-4cf6-956e-b70d1e4786cf"
	RECIPIENT_ADDRESS = "0xDE08B1Fd79b7016F8DD3Df11f7fa0FbfdF07c941"
)

// sponsorSetup initializes a sponsor client using the environment variable.
func sponsorSetup(t *testing.T) (*ethclient.Client, sponsorclient.Client, string, error) {
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

	// Create a client (for transaction sending)
	sponsorURL := fmt.Sprintf("https://open-platform-ap.nodereal.io/%s/megafuel-testnet/97", key)
	sponsorClient, err := sponsorclient.New(context.Background(), sponsorURL)
	if err != nil {
		log.Fatalf("Failed to create PaymasterClient: %v", err)
	}

	return client, sponsorClient, yourPrivateKey, nil
}

// TestSponsorAPI conducts several whitelist operations.
func TestSponsorAPI(t *testing.T) {
	_, sponsorClient, _, err := sponsorSetup(t)
	require.NoError(t, err, "Setup should not fail")

	policyUUID, err := uuid.FromString(POLICY_UUID)
	require.NoError(t, err, "Failed to parse UUID")

	testAddToWhitelist(t, sponsorClient, policyUUID, RECIPIENT_ADDRESS)
	testAddToWhitelist(t, sponsorClient, policyUUID, RECIPIENT_ADDRESS)
	testGetWhitelist(t, sponsorClient, policyUUID)
	testRemoveFromWhitelist(t, sponsorClient, policyUUID, RECIPIENT_ADDRESS)
	testEmptyWhitelist(t, sponsorClient, policyUUID)
}

// testAddToWhitelist tests the addition of an address to the ToAccountWhitelist.
func testAddToWhitelist(t *testing.T, client sponsorclient.Client, policyUUID uuid.UUID, address string) {
	success, err := client.AddToWhitelist(context.Background(), sponsorclient.WhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Values:        []string{address},
	})
	require.NoError(t, err, "Should be able to add to whitelist")
	assert.True(t, success, "Whitelist addition should be successful")
	log.Infof("Added %s to whitelist successfully", address)
}

// testGetWhitelist tests retrieving the whitelist.
func testGetWhitelist(t *testing.T, client sponsorclient.Client, policyUUID uuid.UUID) {
	whitelist, err := client.GetWhitelist(context.Background(), sponsorclient.GetWhitelistArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Offset:        0,
		Limit:         2,
	})
	require.NoError(t, err, "Should be able to retrieve whitelist")
	assert.NotEmpty(t, whitelist, "Whitelist should not be empty")
	log.Info("Retrieved whitelist successfully")
}

// testRemoveFromWhitelist tests the removal of an address from the whitelist.
func testRemoveFromWhitelist(t *testing.T, client sponsorclient.Client, policyUUID uuid.UUID, address string) {
	success, err := client.RmFromWhitelist(context.Background(), sponsorclient.WhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
		Values:        []string{address},
	})
	require.NoError(t, err, "Should be able to remove from whitelist")
	assert.True(t, success, "Whitelist removal should be successful")
	log.Infof("Removed %s from whitelist successfully", address)
}

// testEmptyWhitelist tests clearing all entries from a specific whitelist type.
func testEmptyWhitelist(t *testing.T, client sponsorclient.Client, policyUUID uuid.UUID) {
	success, err := client.EmptyWhitelist(context.Background(), sponsorclient.EmptyWhiteListArgs{
		PolicyUUID:    policyUUID,
		WhitelistType: sponsorclient.ToAccountWhitelist,
	})
	require.NoError(t, err, "Should be able to empty whitelist")
	assert.True(t, success, "Whitelist emptying should be successful")
	log.Info("Emptied whitelist successfully")
}

// TestPaymasterAPI tests the critical functionalities related to the Paymaster API.
func TestPrivatePolicyGaslessTransaction(t *testing.T) {
	// Setup Ethereum client and Paymaster client. Ensure no errors during the setup.
	client, sponsorClient, yourPrivateKey, err := sponsorSetup(t)
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

	privatePolicySponsorableInfo, err := sponsorClient.IsSponsorable(context.Background(), privatePolicySponsorableTx, &paymasterclient.IsSponsorableOptions{PrivatePolicyUUID: PRIVATE_POLICY})
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

	transaction, err := sponsorClient.SendRawTransaction(context.Background(), txInput, &paymasterclient.SendRawTransactionOptions{PrivatePolicyUUID: PRIVATE_POLICY, UserAgent: "Test User Agent"})
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
