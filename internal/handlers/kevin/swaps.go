package kevin

import (
	"fmt"
	"time"

	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleTokenSwapLabeled(label string) handlers.TransactionHandlerFunc {
	return func(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
		return handleTokenSwap(label, bundle, client, export)
	}
}

func handleTokenSwap(
	label string,
	bundle handlers.TransactionBundle,
	client *evm.Client,
	export handlers.CTCWriter,
) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if bundle.Info.Hash == TX_3 {
		fixTx3NetTransfers(client, netTransfers)
	}

	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for swap")
	}

	var bought *core.Amount
	var sold *core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for swap")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Swap recipient not the transaction sender")
			}
			if amount.Value.IsNegative() {
				if sold != nil {
					panic("Multiple assets sold for swap")
				}
				outflow := amount.Neg()
				sold = &outflow
			} else if amount.Value.IsPositive() {
				if bought != nil {
					panic("Multiple assets bought for swap")
				}
				bought = amount
			} else {
				panic("Zero-value transfer for swap")
			}
		}
	}

	if bought == nil || sold == nil {
		panic("Confusing asset state for swap")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:     time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:    bundle.Info.Network,
		ID:            bundle.Info.Hash,
		Type:          ctc_util.CTCSell,
		BaseCurrency:  sold.Asset.Symbol,
		BaseAmount:    sold.Value,
		QuoteCurrency: bought.Asset.Symbol,
		QuoteAmount:   bought.Value,
		From:          bundle.Info.From,
		To:            bundle.Info.To,
		Description:   fmt.Sprintf("%s: sell %s for %s", label, sold, bought),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
