package kevin

import (
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleWrappedNativeWithdraw(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return handleTokenSwap(bundle, client, export)
}

func handleWrappedNativeDeposit(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return handleTokenSwap(bundle, client, export)
}
