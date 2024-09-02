package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

type Big big.Int

var (
	zeroBigInt = big.NewInt(0)
	ZeroBig    = (*Big)(zeroBigInt)
)

func (i *Big) Scan(value interface{}) error {
	if i == nil {
		return nil
	}
	var signByte uint8

	bts, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal Big value:", value))
	}

	if len(bts) >= 1 {
		signByte = bts[0]
		bts = bts[1:]
	}

	i.Raw().SetBytes(bts)
	// if the sign byte indicate negative sign, set value as negative
	if signByte == 1 {
		i.Raw().Neg(i.Raw())
	}

	return nil
}

// Value of big would add one byte in the front of original value to indicate sign (+,-)
func (i *Big) Value() (driver.Value, error) {
	if i == nil {
		return []byte{}, nil
	}
	var signByte uint8

	// Use byte 1 to indicate original negative sign, byte 0 to indicate original positive sign
	if i.Raw().Sign() == -1 {
		signByte = 1
	} else {
		signByte = 0
	}

	return append([]byte{signByte}, i.Raw().Bytes()...), nil
}

func (i *Big) Raw() *big.Int {
	return (*big.Int)(i)
}

func (i *Big) MarshalText() ([]byte, error) {
	return []byte(hexutil.EncodeBig((*big.Int)(i))), nil
}
func (i *Big) UnmarshalText(text []byte) error {
	if i == nil {
		return nil
	}
	// Decode the hex string
	bigInt, err := hexutil.DecodeBig(string(text))
	if err != nil {
		return err
	}

	// Set the Big value
	i.Raw().Set(bigInt)

	return nil
}
