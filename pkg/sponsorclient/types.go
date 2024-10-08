package sponsorclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/gofrs/uuid"

	"github.com/node-real/megafuel-go-sdk/pkg/types"
)

// WhitelistType represents the type of whitelist.
type WhitelistType string

const (
	// FromAccountWhitelist represents a whitelist of from accounts.
	FromAccountWhitelist WhitelistType = "FromAccountWhitelist"

	// ToAccountWhitelist represents a whitelist of to accounts.
	ToAccountWhitelist WhitelistType = "ToAccountWhitelist"

	// ContractMethodSigWhitelist represents a whitelist of contract method signatures.
	ContractMethodSigWhitelist WhitelistType = "ContractMethodSigWhitelist"

	// BEP20ReceiverWhiteList represents a whitelist of BEP20 receiver addresses.
	BEP20ReceiverWhiteList WhitelistType = "BEP20ReceiverWhiteList"
)

type WhiteListArgs struct {
	PolicyUUID    uuid.UUID     `json:"policyUuid"`    // The uuid of policy for which this request is attempt to add the white list values  . Required.
	WhitelistType WhitelistType `json:"whitelistType"` // enum, supported values are "FromAccountWhitelist", "ToAccountWhitelist", "ContractMethodSigWhitelist", "BEP20ReceiverWhiteList"
	Values        []string      `json:"values"`        // a list of values for given WhitelistType.  The max length of this list is WhiteListDataMaxBatchSize. To insert more than WhiteListDataMaxBatchSize records, please invoke this API multiple times.
}

type EmptyWhiteListArgs struct {
	PolicyUUID    uuid.UUID     `json:"policyUuid"`    // The uuid of policy for which this request is attempt to empty the white list values  . Required.
	WhitelistType WhitelistType `json:"whitelistType"` // enum, supported values are "FromAccountWhitelist", "ToAccountWhitelist", "ContractMethodSigWhitelist", "BEP20ReceiverWhiteList"
}

type GetWhitelistArgs struct {
	PolicyUUID    uuid.UUID     `json:"policyUuid"`    // The uuid of policy for which this request is attempt to fetch the white list. Required.
	WhitelistType WhitelistType `json:"whitelistType"` // enum, supported values are "FromAccountWhitelist", "ToAccountWhitelist", "ContractMethodSigWhitelist", "BEP20ReceiverWhiteList"
	Offset        int           `json:"offset"`        // Offset must be less than MaxOffset. Default value is 0
	Limit         int           `json:"limit"`         // Limit must be less than MaxOffset. Default value is 0
}

type UserSpendData struct {
	UserAddress   common.Address `json:"userAddress"`
	GasCost       *types.Big     `json:"gasCost"`
	GasCostCurDay *types.Big     `json:"gasCostCurDay"`
	TxCountCurDay uint64         `json:"txCountCurDay"`
	UpdateAt      uint64         `json:"updateAt"`
	ChainID       int            `json:"chain_id"`
}

type PolicySpendData struct {
	Cost     *types.Big `json:"cost"`
	UpdateAt uint64     `json:"updateAt"`
	ChainID  int        `json:"chain_id"`
}
