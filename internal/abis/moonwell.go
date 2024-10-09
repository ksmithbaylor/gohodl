package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var MoonwellAbi abi.ABI

//go:embed moonwell.json
var moonwellAbiJson string
var MOONWELL_ENTER_MARKETS string
var MOONWELL_CLAIM_REWARD string
var MOONWELL_CLAIM_REWARD_0 string

func init() {
	abi, err := abi.JSON(strings.NewReader(moonwellAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Moonwell ABI: %s\n", err.Error())
	}
	MoonwellAbi = abi
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
