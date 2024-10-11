package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var WonderlandAbi abi.ABI
var WONDERLAND_DEPOSIT string
var WONDERLAND_REDEEM string

//go:embed wonderland.json
var wonderlandAbiJson string

func init() {
	abi, err := abi.JSON(strings.NewReader(wonderlandAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Wonderland ABI: %s\n", err.Error())
	}
	WonderlandAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "deposit":
			WONDERLAND_DEPOSIT = selector
		case "redeem":
			WONDERLAND_REDEEM = selector
		}
	}
}
