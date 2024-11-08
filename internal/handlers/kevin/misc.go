package kevin

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleMiscWithLabel(label string) handlers.TransactionHandlerFunc {
	return func(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
		return handleMisc(label, bundle, client, export)
	}
}

func handleMisc(label string, bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	ctcTx := ctc_util.NewFeeTransaction(
		bundle.Block.Time,
		bundle.Info.Network,
		bundle.Info.Hash,
		bundle.Info.From,
		label,
		bundle.Receipt,
	)

	return export(ctcTx.ToCSV())
}

func handleRewardWithLabel(label string) handlers.TransactionHandlerFunc {
	return func(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
		return handleReward(label, bundle, client, export)
	}
}

func handleReward(label string, bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) > 1 {
		panic("Multiple net transfers for rewards tx")
	}

	var rewardAmount *core.Amount
	var receivedTo common.Address

	for _, transfers := range netTransfers {
		if len(transfers) > 1 {
			panic("Multiple net transfers for rewards tx")
		}
		for addr, amount := range transfers {
			if amount.Value.IsPositive() {
				rewardAmount = amount
				receivedTo = addr
			} else {
				panic("Outflow for rewards tx")
			}
		}
	}

	var ctcTx *ctc_util.CTCTransaction

	if rewardAmount == nil {
		if config.Config.IsMyEvmAddressString(bundle.Info.From) {
			ctcTx = ctc_util.NewFeeTransaction(
				bundle.Block.Time,
				bundle.Info.Network,
				bundle.Info.Hash,
				bundle.Info.From,
				label+": rewards, but nothing was claimed",
				bundle.Receipt,
			)
		}
	} else {
		ctcTx = &ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash,
			Type:         ctc_util.CTCIncome,
			BaseCurrency: rewardAmount.Asset.Symbol,
			BaseAmount:   rewardAmount.Value,
			From:         label,
			To:           receivedTo.Hex(),
			Description:  fmt.Sprintf("%s: reward of %s", label, *rewardAmount),
		}
		ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)
	}

	if ctcTx == nil {
		return nil
	}

	return export(ctcTx.ToCSV())
}
