package paymasterclient

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

type Client interface {
	// ChainID returns the chain ID of the connected domain
	ChainID(ctx context.Context) (string, error)
	// IsSponsorable checks if a transaction is sponsorable
	IsSponsorable(ctx context.Context, tx TransactionArgs) (*IsSponsorableResponse, error)
	// SendRawTransaction sends a raw transaction to the connected domain
	SendRawTransaction(ctx context.Context, input hexutil.Bytes) (common.Hash, error)
	// GetGaslessTransactionByHash returns a gasless transaction by hash
	GetGaslessTransactionByHash(ctx context.Context, txHash common.Hash) (userTx *TransactionResponse, err error)
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

func (c *client) ChainID(ctx context.Context) (string, error) {
	var result string
	err := c.c.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return "", err
	}
	return result, nil
}

func (c *client) IsSponsorable(ctx context.Context, tx TransactionArgs) (*IsSponsorableResponse, error) {
	var result IsSponsorableResponse
	err := c.c.CallContext(ctx, &result, "pm_isSponsorable", tx)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) SendRawTransaction(ctx context.Context, input hexutil.Bytes) (common.Hash, error) {
	var result common.Hash
	err := c.c.CallContext(ctx, &result, "eth_sendRawTransaction", input)
	if err != nil {
		return common.Hash{}, err
	}
	return result, nil
}

func (c *client) GetGaslessTransactionByHash(ctx context.Context, txHash common.Hash) (*TransactionResponse, error) {
	var result TransactionResponse
	err := c.c.CallContext(ctx, &result, "eth_getGaslessTransactionByHash", txHash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
