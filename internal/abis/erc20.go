package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

var Erc20Abi abi.ABI

//go:embed erc20.json
var erc20AbiJson string

func init() {
	abi, err := abi.JSON(strings.NewReader(erc20AbiJson))
	if err != nil {
		log.Fatalf("Could not parse ERC-20 ABI: %s\n", err.Error())
	}
	Erc20Abi = abi
}
