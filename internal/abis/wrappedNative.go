package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var WrappedNativeAbi abi.ABI
var WRAPPED_NATIVE_WITHDRAW string
var WRAPPED_NATIVE_DEPOSIT string

//go:embed wrappedNative.json
var wrappedNativeAbiJson string

func init() {
	abi, err := abi.JSON(strings.NewReader(wrappedNativeAbiJson))
	if err != nil {
		log.Fatalf("Could not parse wrapped native ABI: %s\n", err.Error())
	}
	WrappedNativeAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "withdraw":
			WRAPPED_NATIVE_WITHDRAW = selector
		case "deposit":
			WRAPPED_NATIVE_DEPOSIT = selector
		}
	}
}
