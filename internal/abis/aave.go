package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var AaveAbi abi.ABI

//go:embed aave.json
var aaveAbiJson string
var AAVE_SET_USER_E_MODE string
var AAVE_DEPOSIT string
var AAVE_WITHDRAW string
var AAVE_REPAY_WITH_A_TOKENS string
var AAVE_REPAY string
var AAVE_SUPPLY string
var AAVE_BORROW string

func init() {
	abi, err := abi.JSON(strings.NewReader(aaveAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Aave ABI: %s\n", err.Error())
	}
	AaveAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "setUserEMode":
			AAVE_SET_USER_E_MODE = selector
		case "deposit":
			AAVE_DEPOSIT = selector
		case "withdraw":
			AAVE_WITHDRAW = selector
		case "repayWithATokens":
			AAVE_REPAY_WITH_A_TOKENS = selector
		case "repay":
			AAVE_REPAY = selector
		case "supply":
			AAVE_SUPPLY = selector
		case "borrow":
			AAVE_BORROW = selector
		}
	}
}
