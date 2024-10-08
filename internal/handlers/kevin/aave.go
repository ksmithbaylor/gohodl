package kevin

import (
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleAaveSupply(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveBorrow(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveRepay(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveRepayWithATokens(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveDeposit(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveWithdraw(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveSetUserEMode(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	ctcTx := ctc_util.NewFeeTransaction(
		bundle.Block.Time,
		bundle.Info.Network,
		bundle.Info.Hash,
		bundle.Info.From,
		"aave: set user e-mode",
		bundle.Receipt,
	)
	return export(ctcTx.ToCSV())
}
