package types

import "fmt"

type NetworkID fmt.Stringer

type NetworkKind string

const (
	UtxoNetworkKind   NetworkKind = "utxo"
	EvmNetworkKind    NetworkKind = "evm"
	CosmosNetworkKind NetworkKind = "cosmos"
)

func (n NetworkKind) String() string {
	return string(n)
}
