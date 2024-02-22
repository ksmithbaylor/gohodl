package evm

type EvmNetworkName string

const (
	Ethereum  EvmNetworkName = "ethereum"
	Polygon   EvmNetworkName = "polygon"
	Avalanche EvmNetworkName = "avalanche"
	Base      EvmNetworkName = "base"
	Optimism  EvmNetworkName = "optimism"
	Moonbeam  EvmNetworkName = "moonbeam"
	Moonriver EvmNetworkName = "moonriver"
	Fantom    EvmNetworkName = "fantom"
	Evmos     EvmNetworkName = "evmos"
)

func (n EvmNetworkName) String() string {
	return string(n)
}
