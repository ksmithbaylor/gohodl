package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var MoonwellComptrollerAbi abi.ABI

//go:embed moonwell_comptroller.json
var moonwellComptrollerAbiJson string
var MOONWELL_ENTER_MARKETS string
var MOONWELL_CLAIM_REWARD string
var MOONWELL_CLAIM_REWARD_0 string

func init() {
	abi, err := abi.JSON(strings.NewReader(moonwellComptrollerAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Moonwell comptroller ABI: %s\n", err.Error())
	}
	MoonwellComptrollerAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "enterMarkets":
			MOONWELL_ENTER_MARKETS = selector
		case "claimReward":
			MOONWELL_CLAIM_REWARD = selector
		case "claimReward0":
			MOONWELL_CLAIM_REWARD_0 = selector
		}
	}
}
