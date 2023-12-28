package types

import "github.com/ethereum/go-ethereum/common"

type EthAddress string

func (a EthAddress) ToGeth() common.Address {
	return common.HexToAddress(string(a))
}

type EthNetwork string

const (
	Ethereum  EthNetwork = "ethereum"
	Polygon   EthNetwork = "polygon"
	Avalanche EthNetwork = "avalanche"
	Base      EthNetwork = "base"
	Optimism  EthNetwork = "optimism"
	Moonbeam  EthNetwork = "moonbeam"
	Moonriver EthNetwork = "moonriver"
	Fantom    EthNetwork = "fantom"
	Evmos     EthNetwork = "evmos"
)

func (n EthNetwork) String() string {
	return string(n)
}
