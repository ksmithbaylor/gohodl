package evm

import (
	"github.com/ethereum/go-ethereum/common"
)

type EvmAddress string

const EvmNullAddress EvmAddress = "0x0000000000000000000000000000000000000000"

func (a EvmAddress) ToGeth() common.Address {
	return common.HexToAddress(string(a))
}

func (a EvmAddress) String() string {
	// Round-trip to get casing/checksum
	return a.ToGeth().String()
}
