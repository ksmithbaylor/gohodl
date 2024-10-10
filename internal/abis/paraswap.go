package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var ParaswapAbi abi.ABI

//go:embed paraswap.json
var paraswapAbiJson string
var PARASWAP_SIMPLE_BUY string
var PARASWAP_SIMPLE_SWAP string
var PARASWAP_SWAP_ON_UNISWAP string
var PARASWAP_MEGA_SWAP string
var PARASWAP_SWAP_ON_UNISWAP_V2_FORK string

func init() {
	abi, err := abi.JSON(strings.NewReader(paraswapAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Paraswap ABI: %s\n", err.Error())
	}
	ParaswapAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "simpleBuy":
			PARASWAP_SIMPLE_BUY = selector
		case "simpleSwap":
			PARASWAP_SIMPLE_SWAP = selector
		case "swapOnUniswap":
			PARASWAP_SWAP_ON_UNISWAP = selector
		case "megaSwap":
			PARASWAP_MEGA_SWAP = selector
		case "swapOnUniswapV2Fork":
			PARASWAP_SWAP_ON_UNISWAP_V2_FORK = selector
		}
	}
}
