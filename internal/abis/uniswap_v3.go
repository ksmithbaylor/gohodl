package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var UniswapV3Abi abi.ABI

//go:embed uniswap_v3.json
var uniswapV3AbiJson string
var UNISWAP_V3_MULTICALL_0 string
var UNISWAP_V3_MULTICALL_1 string

func init() {
	abi, err := abi.JSON(strings.NewReader(uniswapV3AbiJson))
	if err != nil {
		log.Fatalf("Could not parse Uniswap V3 ABI: %s\n", err.Error())
	}
	UniswapV3Abi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "multicall0":
			UNISWAP_V3_MULTICALL_0 = selector
		case "multicall1":
			UNISWAP_V3_MULTICALL_1 = selector
		}
	}
}
