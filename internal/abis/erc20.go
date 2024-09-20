package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var Erc20Abi abi.ABI
var ERC20_TRANSFER string
var ERC20_TRANSFER_FROM string
var ERC20_APPROVE string

//go:embed erc20.json
var erc20AbiJson string

func init() {
	abi, err := abi.JSON(strings.NewReader(erc20AbiJson))
	if err != nil {
		log.Fatalf("Could not parse ERC-20 ABI: %s\n", err.Error())
	}
	Erc20Abi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "transfer":
			ERC20_TRANSFER = selector
		case "transferFrom":
			ERC20_TRANSFER_FROM = selector
		case "approve":
			ERC20_APPROVE = selector
		}
	}
}
