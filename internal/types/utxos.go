package types

type UtxoWalletScheme string

const (
	Segwit       UtxoWalletScheme = "segwit"
	NestedSegwit UtxoWalletScheme = "nested-segwit"
)
