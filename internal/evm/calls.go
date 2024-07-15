package evm

import (
	"fmt"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

func ethCall(to common.Address, sig string, args ...any) (ethereum.CallMsg, error) {
	selector, err := encodeSelector(sig)
	if err != nil {
		return ethereum.CallMsg{}, fmt.Errorf("Unable to encode selector for signature '%s': %w", sig, err)
	}

	encodedArgs, err := encodeArgs(sig, args)
	if err != nil {
		return ethereum.CallMsg{}, fmt.Errorf("Unable to encode args '%v': %w", args, err)
	}

	data := make([]byte, 0, len(selector)+len(encodedArgs))
	data = append(data, selector...)
	data = append(data, encodedArgs...)

	return ethereum.CallMsg{
		To:   &to,
		Data: data,
	}, nil
}

func encodeSelector(sig string) ([]byte, error) {
	// TODO use something from geth to encode the signature to the 4-byte selector
	return make([]byte, 0), nil
}

func encodeArgs(sig string, args []any) ([]byte, error) {
	// TODO use something from geth to encode the arguments
	return make([]byte, 0), nil
}

// TODO: replace with something like:
// msg, err := ethCall(contract, "decimals()")
func decimalsCall(to common.Address) ethereum.CallMsg {
	return ethereum.CallMsg{
		To:   &to,
		Data: []byte{0x31, 0x3c, 0xe5, 0x67},
	}
}

// TODO: replace with something like:
// msg, err := ethCall(contract, "symbol()")
func symbolCall(to common.Address) ethereum.CallMsg {
	return ethereum.CallMsg{
		To:   &to,
		Data: []byte{0x95, 0xd8, 0x9b, 0x41},
	}
}

// TODO: replace with something like:
// msg, err := ethCall(contract, "balanceOf(address)", addr)
func balanceCall(to common.Address, a common.Address) ethereum.CallMsg {
	data := make([]byte, 0, 36)

	data = append(data, 0x70, 0xa0, 0x82, 0x31)
	data = append(data, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0)
	data = append(data, a.Bytes()...)

	return ethereum.CallMsg{
		To:   &to,
		Data: data,
	}
}
