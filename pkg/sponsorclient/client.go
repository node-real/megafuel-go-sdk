package sponsorclient

import (
	"context"

	"github.com/ethereum/go-ethereum/rpc"
)

type Client interface {
	// AddToWhitelist adds a list of values to the whitelist of a policy
	AddToWhitelist(ctx context.Context, args WhiteListArgs) (bool, error)
	// RmFromWhitelist removes a list of values from the whitelist of a policy
	RmFromWhitelist(ctx context.Context, args WhiteListArgs) (bool, error)
	// EmptyWhitelist clear the whitelist of a policy
	EmptyWhitelist(ctx context.Context, args EmptyListArgs) (bool, error)
	// GetWhitelist returns the whitelist of a policy
	GetWhitelist(ctx context.Context, args GetWhitelistArgs) (interface{}, error)
}

type client struct {
	c *rpc.Client
}

func New(ctx context.Context, url string, options ...rpc.ClientOption) (Client, error) {
	c, err := rpc.DialOptions(ctx, url, options...)
	if err != nil {
		return nil, err
	}

	return &client{c}, nil
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

func (c *client) EmptyWhitelist(ctx context.Context, args EmptyListArgs) (bool, error) {
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
