package core

type Network interface {
	GetName() string
	NativeAsset() Asset
}
