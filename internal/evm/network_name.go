package evm

type NetworkName string

const (
	Ethereum  NetworkName = "ethereum"
	Polygon   NetworkName = "polygon"
	Avalanche NetworkName = "avalanche"
	Base      NetworkName = "base"
	Optimism  NetworkName = "optimism"
	Moonbeam  NetworkName = "moonbeam"
	Moonriver NetworkName = "moonriver"
	Fantom    NetworkName = "fantom"
	Evmos     NetworkName = "evmos"
)

func (n NetworkName) String() string {
	return string(n)
}
