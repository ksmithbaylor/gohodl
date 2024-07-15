package core

type Indexer interface {
	GetAllTransactionHashes(address string) ([]string, error)
}
