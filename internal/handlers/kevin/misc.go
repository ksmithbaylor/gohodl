package kevin

import (
	"time"

	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleMiscWithLabel(label string) handlers.TransactionHandlerFunc {
	return func(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
		return handleMisc(label, bundle, client, export)
	}
}

func handleMisc(label string, bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	ctcTx := &ctc_util.CTCTransaction{
		Timestamp:   time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:  bundle.Info.Network,
		ID:          bundle.Info.Hash,
		Type:        ctc_util.CTCSpam,
		Description: label,
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
