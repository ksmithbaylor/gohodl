package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var UniswapUniversalAbi abi.ABI

//go:embed uniswap_universal.json
var uniswapUniversalAbiJson string
var UNISWAP_UNIVERSAL_EXECUTE string
var UNISWAP_UNIVERSAL_EXECUTE_0 string

func init() {
	abi, err := abi.JSON(strings.NewReader(uniswapUniversalAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Uniswap universal router ABI: %s\n", err.Error())
	}
	UniswapUniversalAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "execute":
			UNISWAP_UNIVERSAL_EXECUTE = selector
		case "execute0":
			UNISWAP_UNIVERSAL_EXECUTE_0 = selector
		}
	}
}
