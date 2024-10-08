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

	var handle handlers.TransactionHandlerFunc

	switch {
	case !info.Success:
		handle = handleFailed
	case info.Method == "":
		handle = handleNoData
	case info.Method == abis.ERC20_TRANSFER || info.Method == abis.ERC20_TRANSFER_FROM:
		handle = handleErc20Transfer
	case info.Method == abis.ERC20_APPROVE:
		handle = handleErc20Approve
	case info.Method == abis.INSTADAPP_CAST:
		handle = handleInstadapp
	case info.Method == abis.AAVE_SUPPLY:
		return true, NOT_HANDLED
	case info.Method == abis.AAVE_BORROW:
		return true, NOT_HANDLED
	case info.Method == abis.AAVE_REPAY:
		return true, NOT_HANDLED
	case info.Method == abis.AAVE_REPAY_WITH_A_TOKENS:
		return true, NOT_HANDLED
	case info.Method == abis.AAVE_DEPOSIT:
		return true, NOT_HANDLED
	case info.Method == abis.AAVE_WITHDRAW:
		return true, NOT_HANDLED
	case info.Method == abis.AAVE_SET_USER_E_MODE:
		return true, NOT_HANDLED
	}

	if handle != nil {
		return true, readAndThen(handle)
	}

	return false, nil
}
