package core

type UtxoWalletScheme string

type UtxoNetwork string

const (
	Segwit       UtxoWalletScheme = "segwit"
	NestedSegwit UtxoWalletScheme = "nested-segwit"
	Legacy       UtxoWalletScheme = "legacy"

	Bitcoin  UtxoNetwork = "bitcoin"
	Litecoin UtxoNetwork = "litecoin"
	Dogecoin UtxoNetwork = "dogecoin"
)

func (n UtxoNetwork) String() string {
	return string(n)
}
