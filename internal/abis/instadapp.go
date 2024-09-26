package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var InstadappAbi abi.ABI
var INSTADAPP_CAST string

//go:embed instadapp.json
var instadappAbiJson string

func init() {
	abi, err := abi.JSON(strings.NewReader(instadappAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Instadapp ABI: %s\n", err.Error())
	}
	InstadappAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "cast":
			INSTADAPP_CAST = selector
		}
	}
}
