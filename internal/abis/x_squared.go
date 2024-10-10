package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var XSquaredAbi abi.ABI

//go:embed x_squared.json
var xSquaredAbiJson string
var X_SQUARED_BUY_ITEM string
var X_SQUARED_SELL_ITEM string

func init() {
	abi, err := abi.JSON(strings.NewReader(xSquaredAbiJson))
	if err != nil {
		log.Fatalf("Could not parse XSquared ABI: %s\n", err.Error())
	}
	XSquaredAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "buyItem":
			X_SQUARED_BUY_ITEM = selector
		case "sellItem":
			X_SQUARED_SELL_ITEM = selector
		}
	}
}
