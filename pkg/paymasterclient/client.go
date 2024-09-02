package paymasterclient

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gofrs/uuid"
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

	// GetSponsorTxByTxHash returns a sponsor transaction by hash
	GetSponsorTxByTxHash(ctx context.Context, txHash common.Hash) (sponsorTx *SponsorTx, err error)
	// GetSponsorTxByBundleUUID returns a sponsor transaction by bundle UUID
	GetSponsorTxByBundleUUID(ctx context.Context, bundleUUID uuid.UUID) (sponsorTx *SponsorTx, err error)
	// GetBundleByUUID returns a bundle by UUID
	GetBundleByUUID(ctx context.Context, bundleUUID uuid.UUID) (bundle *Bundle, err error)
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

func (c *client) GetSponsorTxByTxHash(ctx context.Context, txHash common.Hash) (*SponsorTx, error) {
	var result SponsorTx
	err := c.c.CallContext(ctx, &result, "pm_getSponsorTxByTxHash", txHash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetSponsorTxByBundleUUID(ctx context.Context, bundleUUID uuid.UUID) (*SponsorTx, error) {
	var result SponsorTx
	err := c.c.CallContext(ctx, &result, "pm_getSponsorTxByBundleUUID", bundleUUID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetBundleByUUID(ctx context.Context, bundleUUID uuid.UUID) (*Bundle, error) {
	var result Bundle
	err := c.c.CallContext(ctx, &result, "pm_getBundleByUUID", bundleUUID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
