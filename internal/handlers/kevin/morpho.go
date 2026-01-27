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

func handleMorphoClaimRewards(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	claimed := make([]core.Amount, 0)

	for _, transfers := range netTransfers {
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for morpho claim rewards")
			}
			if !amount.Value.IsPositive() {
				panic("Outflow for morpho claim rewards")
			}
			claimed = append(claimed, *amount)
		}
	}

	if len(claimed) == 0 {
		panic("Nothing claimed for morpho claim rewards")
	}

	ctcTx := ctc_util.NewFeeTransaction(
		bundle.Block.Time,
		bundle.Info.Network,
		bundle.Info.Hash+"-1",
		bundle.Info.From,
		"Fee for Morpho rewards claim",
		bundle.Receipt,
	)
	err = export(ctcTx.ToCSV())

	for i, amount := range claimed {
		ctcTx = &ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           fmt.Sprintf("%s-%d", bundle.Info.Hash, i+2),
			Type:         ctc_util.CTCInterest,
			BaseCurrency: amount.Asset.Symbol,
			BaseAmount:   amount.Value,
			From:         "unknown",
			To:           bundle.Info.From,
			Description:  fmt.Sprintf("morpho: claim %s in rewards", amount),
		}
		err = combineErrs(err, export(ctcTx.ToCSV()))
	}

	return err
}
