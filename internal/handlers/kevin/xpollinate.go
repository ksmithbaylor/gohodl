package kevin

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleXpollinateBridgeOut(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for xpollinate bridge out")
	}

	var bridged core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for xpollinate bridge out")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for xpollinate bridge out")
			}
			if !amount.Value.IsNegative() {
				panic("Unexpected net transfers for xpollinate bridge out")
			}
			bridged = amount.Neg()
		}
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBridgeOut,
		BaseCurrency: bridged.Asset.Symbol,
		BaseAmount:   bridged.Value,
		From:         bundle.Info.From,
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("xpollinate: bridge out %s", bridged),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleXpollinateBridgeIn(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for xpollinate bridge in")
	}

	var bridged core.Amount
	var bridgedTo common.Address

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for xpollinate bridge in")
		}
		for addr, amount := range transfers {
			if !amount.Value.IsPositive() {
				panic("Unexpected net transfers for xpollinate bridge in")
			}
			bridged = *amount
			bridgedTo = addr
		}
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBridgeIn,
		BaseCurrency: bridged.Asset.Symbol,
		BaseAmount:   bridged.Value,
		From:         bridgedTo.Hex(),
		To:           bridgedTo.Hex(),
		Description:  fmt.Sprintf("xpollinate: bridge in %s", bridged),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	ctcTx.Print()

	return export(ctcTx.ToCSV())
}
