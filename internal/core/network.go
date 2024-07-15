package core

type Network interface {
	GetKind() NetworkKind
	GetName() string
	NativeAsset() Asset
}
