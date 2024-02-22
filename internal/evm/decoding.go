package evm

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func decodeUint8(data []byte) uint8 {
	if len(data) == 0 {
		return 0
	}
	return uint8(data[len(data)-1])
}

func decodeString(data []byte) (string, error) {
	stringType, err := abi.NewType("string", "", nil)
	if err != nil {
		return "", err
	}

	decoded, err := abi.Arguments{{Type: stringType}}.Unpack(data)
	if err != nil {
		return "", err
	}

	sym, ok := decoded[0].(string)
	if !ok {
		return "", errors.New("Result was not a string")
	}

	return sym, nil
}

func decodeBigInt(data []byte) *big.Int {
	var value big.Int
	value.SetBytes(data)
	return &value
}
