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
type transactionReader func(network, hash string) (
	*types.Transaction,
	*types.Receipt,
	*types.Header,
	error,
)
type transactionHandler func(bundle transactionBundle, client *evm.Client, export ctcWriter) error
type transactionBundle struct {
	info    *evm.TxInfo
	tx      *types.Transaction
	receipt *types.Receipt
	block   *types.Header
}

type Private interface {
	HandleTransaction(
		info *evm.TxInfo,
		client *evm.Client,
		readTransaction transactionReader,
		export ctcWriter,
	) (bool, error)
}

// The below code allows the program to compile before implementing the above.

var UnimplementedError = errors.New("Private functions not implemented")

var Implementation Private

type placeholder struct{}

func (p placeholder) HandleTransaction(
	info *evm.TxInfo,
	client *evm.Client,
	readTransaction transactionReader,
	export ctcWriter,
) (bool, error) {
	return false, UnimplementedError
}

func init() {
	Implementation = placeholder(struct{}{})
}
