package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var MoonwellTokenAbi abi.ABI

//go:embed moonwell_token.json
var moonwellTokenAbiJson string
var MOONWELL_MINT string
var MOONWELL_BORROW string
var MOONWELL_REPAY_BORROW string
var MOONWELL_REDEEM string
var MOONWELL_REDEEM_UNDERLYING string

func init() {
	abi, err := abi.JSON(strings.NewReader(moonwellTokenAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Moonwell token ABI: %s\n", err.Error())
	}
	MoonwellTokenAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "mint":
			MOONWELL_MINT = selector
		case "borrow":
			MOONWELL_BORROW = selector
		case "repayBorrow":
			MOONWELL_REPAY_BORROW = selector
		case "redeem":
			MOONWELL_REDEEM = selector
		case "redeemUnderlying":
			MOONWELL_REDEEM_UNDERLYING = selector
		}
	}
}
