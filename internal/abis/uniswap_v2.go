package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var UniswapV2Abi abi.ABI

//go:embed uniswap_v2.json
var uniswapV2AbiJson string
var UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_TOKENS string
var UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_TOKENS string
var UNISWAP_V2_SWAP_EXACT_ETH_FOR_TOKENS string
var UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_ETH string
var UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_ETH string
var UNISWAP_V2_SWAP_ETH_FOR_EXACT_TOKENS string
var UNISWAP_V2_ADD_LIQUIDITY string
var UNISWAP_V2_ADD_LIQUIDITY_ETH string
var UNISWAP_V2_REMOVE_LIQUIDITY_ETH string
var UNISWAP_V2_REMOVE_LIQUIDITY_PERMIT string
var UNISWAP_V2_REMOVE_LIQUIDITY_ETH_PERMIT string
var UNISWAP_V2_REMOVE_LIQUIDITY_ETH_PERMIT_FOTT string

func init() {
	abi, err := abi.JSON(strings.NewReader(uniswapV2AbiJson))
	if err != nil {
		log.Fatalf("Could not parse Uniswap V2 ABI: %s\n", err.Error())
	}
	UniswapV2Abi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "swapExactTokensForTokens":
			UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_TOKENS = selector
		case "swapTokensForExactTokens":
			UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_TOKENS = selector
		case "swapExactETHForTokens":
			UNISWAP_V2_SWAP_EXACT_ETH_FOR_TOKENS = selector
		case "swapTokensForExactETH":
			UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_ETH = selector
		case "swapExactTokensForETH":
			UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_ETH = selector
		case "swapETHForExactTokens":
			UNISWAP_V2_SWAP_ETH_FOR_EXACT_TOKENS = selector
		case "addLiquidity":
			UNISWAP_V2_ADD_LIQUIDITY = selector
		case "addLiquidityETH":
			UNISWAP_V2_ADD_LIQUIDITY_ETH = selector
		case "removeLiquidityETH":
			UNISWAP_V2_REMOVE_LIQUIDITY_ETH = selector
		case "removeLiquidityWithPermit":
			UNISWAP_V2_REMOVE_LIQUIDITY_PERMIT = selector
		case "removeLiquidityETHWithPermit":
			UNISWAP_V2_REMOVE_LIQUIDITY_ETH_PERMIT = selector
		case "removeLiquidityETHWithPermitSupportingFeeOnTransferTokens":
			UNISWAP_V2_REMOVE_LIQUIDITY_ETH_PERMIT_FOTT = selector
		}
	}
}
