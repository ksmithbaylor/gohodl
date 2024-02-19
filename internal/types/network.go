package types

type NetworkKind string

const (
	EvmNetworkKind    NetworkKind = "evm"
	UtxoNetworkKind   NetworkKind = "utxo"
	SolanaNetworkKind NetworkKind = "solana"
	CosmosNetworkKind NetworkKind = "cosmos"
)

func (n NetworkKind) String() string {
	return string(n)
}

type Network struct {
	Kind    NetworkKind
	Details any
}
