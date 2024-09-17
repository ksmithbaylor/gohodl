package private

import (
	"errors"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

// This module is meant to contain code specific to each user's transactions. It
// should export the functions below, and is used by the `ctc` module in the
// export step to translate the raw transaction data to the CSV format accepted
// by CTC. To customize this for your own usage, simply implement the functions.
// The file is named how it is to ensure its `init` function runs first,
// allowing it to be overridden by the other files in the directory.

type ctcWriter func([]string) error
type transactionReader func(network, hash string) (*types.Transaction, *types.Receipt, error)
type transactionHandler func(bundle transactionBundle, export ctcWriter) error
type transactionBundle struct {
	info    *evm.TxInfo
	tx      *types.Transaction
	receipt *types.Receipt
}

type Private interface {
	HandleTransaction(info *evm.TxInfo, readTransaction transactionReader, export ctcWriter) error
}

// The below code allows the program to compile before implementing the above.

var UnimplementedError = errors.New("Private functions not implemented")

var Implementation Private

type placeholder struct{}

func (p placeholder) HandleTransaction(info *evm.TxInfo, readTransaction transactionReader, export ctcWriter) error {
	return UnimplementedError
}

func init() {
	Implementation = placeholder(struct{}{})
}
