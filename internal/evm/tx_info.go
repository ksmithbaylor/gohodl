package evm

type TxInfo struct {
	Time      int
	Network   string
	Hash      string
	BlockHash string
	From      string
	To        string
	Method    string
	Value     string
	Success   bool
}
