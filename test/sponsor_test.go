package test

import (
	"context"
	"os"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/node-real/megafuel-go-sdk/pkg/sponsorclient"
)

// Constants for API
const (
	PAYMASTER_URL     = "https://bsc-megafuel-testnet.nodereal.io/97"
	POLICY_UUID       = "72191372-5550-4cf6-956e-b70d1e4786cf"
	RECIPIENT_ADDRESS = "0xDE08B1Fd79b7016F8DD3Df11f7fa0FbfdF07c941"
)

// sponsorSetup initializes a sponsor client using the environment variable.
func sponsorSetup(t *testing.T) (sponsorclient.Client, error) {
	t.Helper()
	key := os.Getenv("OPEN_PLATFORM_PRIVATE_KEY")
	if key == "" {
		log.Fatal("Environment variable OPEN_PLATFORM_PRIVATE_KEY is not set")
	}
	return sponsorclient.New(context.Background(), "https://open-platform.nodereal.io/"+key+"/megafuel-testnet")
}

// TestSponsorAPI conducts several whitelist operations.
func TestSponsorAPI(t *testing.T) {
	sponsorClient, err := sponsorSetup(t)
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
