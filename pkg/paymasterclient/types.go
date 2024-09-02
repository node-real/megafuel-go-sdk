package paymasterclient

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/gofrs/uuid"

	"github.com/node-real/megafuel-go-sdk/pkg/types"
)

type TransactionArgs struct {
	To    *common.Address `json:"to"`
	From  common.Address  `json:"from"`
	Value *hexutil.Big    `json:"value"`
	Gas   *hexutil.Uint64 `json:"gas"`
	Data  *hexutil.Bytes  `json:"data"`
}

type IsSponsorableResponse struct {
	Sponsorable    bool   `json:"sponsorable"`              // Sponsorable is a mandatory field, bool value, indicating if a given tx is able to sponsor.
	SponsorName    string `json:"sponsorName,omitempty"`    // SponsorName is an optional field, string value, shows the name of the policy sponsor.
	SponsorIcon    string `json:"sponsorIcon,omitempty"`    // SponsorIcon is an optional field, string value, shows the icon of the policy sponsor.
	SponsorWebsite string `json:"sponsorWebsite,omitempty"` // SponsorWebsite is an optional field, string value, shows the website of the policy sponsor.
}

type Status int8 // enum: new/pending/failed/confirmed/invalid

const (
	StatusNew Status = iota
	StatusPending
	StatusConfirmed
	StatusFailed
	StatusInvalid
)

type TransactionResponse struct {
	TxHash          common.Hash     `json:"txHash"`
	BundleID        uint64          `json:"bundleID"`
	FromAddress     common.Address  `json:"fromAddress"`
	ToAddress       *common.Address `json:"ToAddress"`
	RawData         []byte          `json:"rawData"`
	Status          Status          `json:"status"`
	GasUsed         uint64          `json:"gasUsed"`
	PolicyID        uint64          `json:"policyID"`
	Source          string          `json:"source"`          // user-agent
	BornBlockNumber int64           `json:"bornBlockNumber"` // the height when the tx is sent to builders.
	ChainID         int             `json:"chain_id"`
}

type SponsorTx struct {
	TxHash          common.Hash    `json:"txHash"`
	Address         common.Address `json:"address"`
	BundleUUID      uuid.UUID      `json:"bundleUUID"`
	Status          Status         `json:"status"`
	GasPrice        *types.Big     `json:"gasPrice"`
	GasFee          *types.Big     `json:"gasFee"`
	BornBlockNumber int64          `json:"bornBlockNumber"` // the height when the tx is sent to builders.
	ChainID         int            `json:"chain_id"`
}

type Bundle struct {
	BundleUUID           uuid.UUID  `json:"bundleUUID"`
	Status               Status     `json:"status"`
	AvgGasPrice          *types.Big `json:"avgGasPrice"`
	BornBlockNumber      int64      `json:"bornBlockNumber"` // the height when the tx is sent to builders.
	ConfirmedBlockNumber int64      `json:"confirmedBlockNumber"`
	ConfirmedDate        uint64     `json:"confirmedDate"`
	ChainID              int        `json:"chain_id"`
}
