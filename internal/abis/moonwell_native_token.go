package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var MoonwellNativeTokenAbi abi.ABI

//go:embed moonwell_native_token.json
var moonwellNativeTokenAbiJson string
var MOONWELL_NATIVE_MINT string

func init() {
	abi, err := abi.JSON(strings.NewReader(moonwellNativeTokenAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Moonwell native token ABI: %s\n", err.Error())
	}
	MoonwellNativeTokenAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "mint":
			MOONWELL_NATIVE_MINT = selector
		}
	}
}
