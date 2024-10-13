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

var WONDERLAND_CONTRACT = "0x694738E0A438d90487b4a549b201142c1a97B556"

// Mostly copied from moonwell

func handleWonderlandDeposit(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for wonderland deposit")
	}

	var deposited core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for wonderland deposit")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for wonderland deposit")
			}
			if amount.Value.IsNegative() {
				deposited = amount.Neg()
			}
		}
	}

	if deposited.Asset.Symbol == "" {
		panic("No asset deposited for wonderland deposit")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralDeposit,
		BaseCurrency: deposited.Asset.Symbol,
		BaseAmount:   deposited.Value,
		From:         bundle.Info.From,
		To:           "wonderland",
		Description:  fmt.Sprintf("wonderland: deposit %s", deposited),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleWonderlandRedeem(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for wonderland redeem/bond")
	}

	var borrowed core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for wonderland redeem/bond")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for wonderland redeem/bond")
			}
			if amount.Value.IsPositive() {
				borrowed = *amount
			}
		}
	}

	if borrowed.Asset.Symbol == "" {
		panic("No asset received for wonderland redeem/bond")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBorrow,
		BaseCurrency: borrowed.Asset.Symbol,
		BaseAmount:   borrowed.Value,
		From:         "wonderland",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("wonderland: redeem/bond %s", borrowed),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
