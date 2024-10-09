package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var MoonwellStakingAbi abi.ABI

//go:embed moonwell_staking.json
var moonwellStakingAbiJson string
var MOONWELL_STAKING_STAKE string
var MOONWELL_STAKING_COOLDOWN string
var MOONWELL_STAKING_CLAIM string
var MOONWELL_STAKING_REDEEM string

func init() {
	abi, err := abi.JSON(strings.NewReader(moonwellStakingAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Moonwell token ABI: %s\n", err.Error())
	}
	MoonwellStakingAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)
		printIfUsed("moonwell staking", method.Sig, selector)

		switch method.Name {
		case "stake":
			MOONWELL_STAKING_STAKE = selector
		case "cooldown":
			MOONWELL_STAKING_COOLDOWN = selector
		case "claimRewards":
			MOONWELL_STAKING_CLAIM = selector
		case "redeem":
			MOONWELL_STAKING_REDEEM = selector
		}
	}
}
