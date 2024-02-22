package core

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
