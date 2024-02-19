package types

import "github.com/ethereum/go-ethereum/common"

////////////////////////////////////////////////////////////////////////////////

type EthAddress string

func (a EthAddress) ToGeth() common.Address {
	return common.HexToAddress(string(a))
}

////////////////////////////////////////////////////////////////////////////////

type EthNetworkName string

const (
	Ethereum  EthNetworkName = "ethereum"
	Polygon   EthNetworkName = "polygon"
	Avalanche EthNetworkName = "avalanche"
	Base      EthNetworkName = "base"
	Optimism  EthNetworkName = "optimism"
	Moonbeam  EthNetworkName = "moonbeam"
	Moonriver EthNetworkName = "moonriver"
	Fantom    EthNetworkName = "fantom"
	Evmos     EthNetworkName = "evmos"
)

func (n EthNetworkName) String() string {
	return string(n)
}

////////////////////////////////////////////////////////////////////////////////

type EthNetworkConfig struct {
	ChainID     uint     `mapstructure:"chain_id"`
	NativeAsset string   `mapstructure:"native_asset"`
	RPCs        []string `mapstructure:"rpcs"`
}

////////////////////////////////////////////////////////////////////////////////

type EthNetwork struct {
	Name   EthNetworkName
	Config EthNetworkConfig
}
