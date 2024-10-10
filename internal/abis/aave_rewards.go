package abis

import (
	_ "embed"
	"log"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

var AaveRewardsAbi abi.ABI

//go:embed aave_rewards.json
var aaveRewardsAbiJson string
var AAVE_CLAIM_REWARDS string
var AAVE_CLAIM_ALL_REWARDS string

func init() {
	abi, err := abi.JSON(strings.NewReader(aaveRewardsAbiJson))
	if err != nil {
		log.Fatalf("Could not parse Aave rewards ABI: %s\n", err.Error())
	}
	AaveRewardsAbi = abi
	for _, method := range abi.Methods {
		selector := "0x" + common.Bytes2Hex(method.ID)

		switch method.Name {
		case "claimRewards":
			AAVE_CLAIM_REWARDS = selector
		case "claimAllRewards":
			AAVE_CLAIM_ALL_REWARDS = selector
		}
	}
}
