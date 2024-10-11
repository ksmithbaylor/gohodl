package kevin

import (
	"github.com/ksmithbaylor/gohodl/internal/evm"
	// "github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleBenqi(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	// printHeader(bundle)
	// netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	// if err != nil {
	//   return err
	// }
	// netTransfers.Print()
	return NOT_HANDLED
}
