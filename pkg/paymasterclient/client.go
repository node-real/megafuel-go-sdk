package paymasterclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/gofrs/uuid"
)

type Client interface {
	// ChainID returns the chain ID of the connected domain
	ChainID(ctx context.Context) (*big.Int, error)
	// IsSponsorable checks if a transaction is sponsorable
	IsSponsorable(ctx context.Context, tx TransactionArgs, opts IsSponsorableOptions) (*IsSponsorableResponse, error)
	// SendRawTransaction sends a raw transaction to the connected domain
	SendRawTransaction(ctx context.Context, input hexutil.Bytes, opts SendRawTransactionOptions) (common.Hash, error)
	// GetGaslessTransactionByHash returns a gasless transaction by hash
	GetGaslessTransactionByHash(ctx context.Context, txHash common.Hash) (userTx *TransactionResponse, err error)

	// GetSponsorTxByTxHash returns a sponsor transaction by hash
	GetSponsorTxByTxHash(ctx context.Context, txHash common.Hash) (sponsorTx *SponsorTx, err error)
	// GetSponsorTxByBundleUUID returns a sponsor transaction by bundle UUID
	GetSponsorTxByBundleUUID(ctx context.Context, bundleUUID uuid.UUID) (sponsorTx *SponsorTx, err error)
	// GetBundleByUUID returns a bundle by UUID
	GetBundleByUUID(ctx context.Context, bundleUUID uuid.UUID) (bundle *Bundle, err error)
	// GetTransactionCount returns the number of transactions sent from an address
	GetTransactionCount(ctx context.Context, address common.Address, blockNrOrHash rpc.BlockNumberOrHash) (uint64, error)
}

type client struct {
	userClient    *rpc.Client
	sponsorClient *rpc.Client
}

func New(ctx context.Context, userURL, sponsorURL string, options ...rpc.ClientOption) (Client, error) {
	userClient, err := rpc.DialOptions(ctx, userURL, options...)
	if err != nil {
		return nil, err
	}

	sponsorClient, err := rpc.DialOptions(ctx, sponsorURL, options...)
	if err != nil {
		userClient.Close() // Close the user client if user client creation fails
		return nil, err
	}

	return &client{
		userClient:    userClient,
		sponsorClient: sponsorClient,
	}, nil
}

func (c *client) ChainID(ctx context.Context) (*big.Int, error) {
	var result hexutil.Big
	err := c.userClient.CallContext(ctx, &result, "eth_chainId")
	if err != nil {
		return nil, err
	}
	return (*big.Int)(&result), err
}

func (c *client) IsSponsorable(ctx context.Context, tx TransactionArgs, opts IsSponsorableOptions) (*IsSponsorableResponse, error) {
	var result IsSponsorableResponse
	if opts.PrivatePolicyUUID != "" {
		c.sponsorClient.SetHeader("X-MegaFuel-Policy-Uuid", opts.PrivatePolicyUUID)
		err := c.sponsorClient.CallContext(ctx, &result, "pm_isSponsorable", tx)
		if err != nil {
			return nil, err
		}
		return &result, nil
	}
	err := c.userClient.CallContext(ctx, &result, "pm_isSponsorable", tx)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) SendRawTransaction(ctx context.Context, input hexutil.Bytes, opts SendRawTransactionOptions) (common.Hash, error) {
	var result common.Hash
	if opts.PrivatePolicyUUID != "" {
		c.sponsorClient.SetHeader("X-MegaFuel-Policy-Uuid", opts.PrivatePolicyUUID)
		err := c.sponsorClient.CallContext(ctx, &result, "eth_sendRawTransaction", input)
		if err != nil {
			return common.Hash{}, err
		}
		return result, nil
	}
	err := c.userClient.CallContext(ctx, &result, "eth_sendRawTransaction", input)
	if err != nil {
		return common.Hash{}, err
	}
	return result, nil
}

func (c *client) GetGaslessTransactionByHash(ctx context.Context, txHash common.Hash) (*TransactionResponse, error) {
	var result TransactionResponse
	err := c.userClient.CallContext(ctx, &result, "eth_getGaslessTransactionByHash", txHash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetSponsorTxByTxHash(ctx context.Context, txHash common.Hash) (*SponsorTx, error) {
	var result SponsorTx
	err := c.userClient.CallContext(ctx, &result, "pm_getSponsorTxByTxHash", txHash)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetSponsorTxByBundleUUID(ctx context.Context, bundleUUID uuid.UUID) (*SponsorTx, error) {
	var result SponsorTx
	err := c.userClient.CallContext(ctx, &result, "pm_getSponsorTxByBundleUuid", bundleUUID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetBundleByUUID(ctx context.Context, bundleUUID uuid.UUID) (*Bundle, error) {
	var result Bundle
	err := c.userClient.CallContext(ctx, &result, "pm_getBundleByUuid", bundleUUID)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) GetTransactionCount(ctx context.Context, address common.Address, blockNrOrHash rpc.BlockNumberOrHash) (uint64, error) {
	var result hexutil.Uint64
	err := c.userClient.CallContext(ctx, &result, "eth_getTransactionCount", address, blockNrOrHash)
	if err != nil {
		return 0, err
	}
	return uint64(result), nil
}
