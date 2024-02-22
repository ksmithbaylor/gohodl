package evm

import (
	"github.com/ethereum/go-ethereum/common"
)

type Address string

const NullAddress Address = "0x0000000000000000000000000000000000000000"

func (a Address) ToGeth() common.Address {
	return common.HexToAddress(string(a))
}

func (a Address) String() string {
	// Round-trip to get casing/checksum
	return a.ToGeth().String()
}
