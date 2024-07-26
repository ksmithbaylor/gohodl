package core

type Indexer interface {
	GetAllTransactionHashes(address string, startBlock, endBlock *int) ([]string, error)
}
