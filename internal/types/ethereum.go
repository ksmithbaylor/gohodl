package types

type EthAddress string

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
