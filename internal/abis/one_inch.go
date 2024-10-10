package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var OneInchAbi abi.ABI

//go:embed one_inch.json
var oneInchAbiJson string
var ONE_INCH_SWAP string

func init() {
	abi, err := abi.JSON(strings.NewReader(oneInchAbiJson))
	if err != nil {
		log.Fatalf("Could not parse 1inch ABI: %s\n", err.Error())
	}
	OneInchAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "swap":
			ONE_INCH_SWAP = selector
		}
	}
}
