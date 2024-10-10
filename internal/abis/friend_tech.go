package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var FriendTechAbi abi.ABI

//go:embed friend_tech.json
var friendTechAbiJson string
var FRIEND_TECH_BUY_SHARES string
var FRIEND_TECH_SELL_SHARES string

func init() {
	abi, err := abi.JSON(strings.NewReader(friendTechAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Friend Tech ABI: %s\n", err.Error())
	}
	FriendTechAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "buyShares":
			FRIEND_TECH_BUY_SHARES = selector
		case "sellShares":
			FRIEND_TECH_SELL_SHARES = selector
		}
	}
}
