package core

type NodeClient interface {
	LatestBlock() (uint64, error)
}
