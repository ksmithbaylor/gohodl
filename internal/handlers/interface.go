package handlers

import (
	"errors"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/evm"
)

// This module is meant to contain code specific to each user's transactions. It
// should export the functions below, and is used by the `ctc` module in the
// export step to translate the raw transaction data to the CSV format accepted
// by CTC. To customize this for your own usage, simply implement the interface
// in a new nested module and use that instead of the `kevin` one. Mine can be
// used as an example. Private constants and other values are in private.go,
// which is protected by git-crypt for my own personal implementation.

type CTCWriter func([]string) error
type TransactionReader func(network, hash string) (
	*types.Transaction,
	*types.Receipt,
	*types.Header,
	error,
)
type TransactionHandlerFunc func(bundle TransactionBundle, client *evm.Client, export CTCWriter) error
type TransactionBundle struct {
	Info    *evm.TxInfo
	Tx      *types.Transaction
	Receipt *types.Receipt
	Block   *types.Header
}

type TransactionHander interface {
	HandleTransaction(
		info *evm.TxInfo,
		client *evm.Client,
		readTransaction TransactionReader,
		export CTCWriter,
	) (bool, error)
}

// The below code allows the program to compile before implementing the real
// handling logic

type placeholder struct{}

var Implementation = placeholder(struct{}{})

func (p placeholder) HandleTransaction(
	info *evm.TxInfo,
	client *evm.Client,
	readTransaction TransactionReader,
	export CTCWriter,
) (bool, error) {
	return false, errors.New("TransactionHander functions not implemented")
}
