package kevin

import (
	"errors"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

var NOT_HANDLED = errors.New("transaction not handled")

var Implementation = personalHandler(struct{}{})

type personalHandler struct{}

func (h personalHandler) HandleTransaction(
	info *evm.TxInfo,
	client *evm.Client,
	readTransactionBundle handlers.TransactionReader,
	export handlers.CTCWriter,
) (bool, error) {
	readAndThen := func(handle handlers.TransactionHandlerFunc) error {
		tx, receipt, block, err := readTransactionBundle(info.Network, info.Hash)
		if err != nil {
			return err
		}

		bundle := handlers.TransactionBundle{
			Info:    info,
			Tx:      tx,
			Receipt: receipt,
			Block:   block,
		}
		return handle(bundle, client, export)
	}

	if info.Method == abis.ERC20_TRANSFER || info.Method == abis.ERC20_TRANSFER_FROM {
		return true, readAndThen(handleErc20Transfer)
	}

	if info.Method == abis.ERC20_APPROVE {
		return true, readAndThen(handleErc20Approve)
	}

	if info.Method == "" {
		return true, readAndThen(handleNoData)
	}

	return false, nil
}
