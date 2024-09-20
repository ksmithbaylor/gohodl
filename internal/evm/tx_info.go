package evm

type TxInfo struct {
	Network   string
	Hash      string
	BlockHash string
	From      string
	To        string
	Method    string
	Value     string
	Success   bool
}
