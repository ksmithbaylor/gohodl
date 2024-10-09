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

func handleMoonwellEnterMarkets(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	// This just opts an asset into being used as collateral, so the fee is all
	// that needs to be handled
	ctcTx := ctc_util.NewFeeTransaction(
		bundle.Block.Time,
		bundle.Info.Network,
		bundle.Info.Hash,
		bundle.Info.From,
		"moonwell: use an asset as collateral",
		bundle.Receipt,
	)

	return export(ctcTx.ToCSV())
}

func handleMoonwellClaimReward(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) > 1 {
		panic("Multiple net transfers for moonwell rewards claim")
	}

	var rewardAmount *core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) > 1 {
			panic("Multiple net transfers for moonwell rewards claim")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Wrong address received rewards for moonwell rewards claim")
			}
			rewardAmount = amount
		}
	}

	var ctcTx *ctc_util.CTCTransaction

	if rewardAmount == nil {
		ctcTx = ctc_util.NewFeeTransaction(
			bundle.Block.Time,
			bundle.Info.Network,
			bundle.Info.Hash,
			bundle.Info.From,
			"moonwell: claim rewards, but nothing was claimed",
			bundle.Receipt,
		)
	} else {
		ctcTx = &ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash,
			Type:         ctc_util.CTCIncome,
			BaseCurrency: rewardAmount.Asset.Symbol,
			BaseAmount:   rewardAmount.Value,
			From:         "moonwell",
			To:           bundle.Info.From,
			Description: fmt.Sprintf("moonwell: claim reward on %s, %s",
				bundle.Info.Network,
				*rewardAmount,
			),
		}
	}

	return export(ctcTx.ToCSV())
}

func handleMoonwellMint(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleMoonwellBorrow(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleMoonwellRepayBorrow(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleMoonwellRedeem(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleMoonwellRedeemUnderlying(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}
