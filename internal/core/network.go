package core

type Network interface {
	GetKind() NetworkKind
	GetName() string
	GetDeprecated() bool
	NativeAsset() Asset
}
