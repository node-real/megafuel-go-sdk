package sponsorclient

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gofrs/uuid"

	"github.com/node-real/megafuel-go-sdk/pkg/paymasterclient"
)

// Client interface defines the methods available for the sponsor client.
// This client combines sponsor-specific functionality with all paymaster client (USER API) methods.
type Client interface {
	// AddToWhitelist adds a list of values to the whitelist of a policy
	AddToWhitelist(ctx context.Context, args WhiteListArgs) (bool, error)
	// RmFromWhitelist removes a list of values from the whitelist of a policy
	RmFromWhitelist(ctx context.Context, args WhiteListArgs) (bool, error)
	// EmptyWhitelist clear the whitelist of a policy
	EmptyWhitelist(ctx context.Context, args EmptyWhiteListArgs) (bool, error)
	// GetWhitelist returns the whitelist of a policy
	GetWhitelist(ctx context.Context, args GetWhitelistArgs) (interface{}, error)

	// GetUserSpendData returns the user spend data on a policy
	GetUserSpendData(ctx context.Context, fromAddress common.Address, policyUUID uuid.UUID) (*UserSpendData, error)
	// GetPolicySpendData returns the spend data of a policy
	GetPolicySpendData(ctx context.Context, policyUUID uuid.UUID) (*PolicySpendData, error)
	paymasterclient.Client
}

type client struct {
	c *rpc.Client
	paymasterclient.Client
}

func New(ctx context.Context, url string, options ...rpc.ClientOption) (Client, error) {
	c, err := rpc.DialOptions(ctx, url, options...)
	if err != nil {
		return nil, err
	}

	c2, err := paymasterclient.New(ctx, url, options...)
	if err != nil {
		return nil, err
	}
	return &client{c, c2}, nil
}

func (c *client) AddToWhitelist(ctx context.Context, args WhiteListArgs) (bool, error) {
	var result bool
	err := c.c.CallContext(ctx, &result, "pm_addToWhitelist", args)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (c *client) RmFromWhitelist(ctx context.Context, args WhiteListArgs) (bool, error) {
	var result bool
	err := c.c.CallContext(ctx, &result, "pm_rmFromWhitelist", args)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (c *client) EmptyWhitelist(ctx context.Context, args EmptyWhiteListArgs) (bool, error) {
	var result bool
	err := c.c.CallContext(ctx, &result, "pm_emptyWhitelist", args)
	if err != nil {
		return false, err
	}
	return result, nil
}

func (c *client) GetWhitelist(ctx context.Context, args GetWhitelistArgs) (interface{}, error) {
	var result interface{}
	err := c.c.CallContext(ctx, &result, "pm_getWhitelist", args)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *client) GetUserSpendData(ctx context.Context, fromAddress common.Address, policyUUID uuid.UUID) (*UserSpendData, error) {
	var result UserSpendData
	err := c.c.CallContext(ctx, &result, "pm_getUserSpendData", fromAddress, policyUUID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetPolicySpendData(ctx context.Context, policyUUID uuid.UUID) (*PolicySpendData, error) {
	var result PolicySpendData
	err := c.c.CallContext(ctx, &result, "pm_getPolicySpendData", policyUUID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
