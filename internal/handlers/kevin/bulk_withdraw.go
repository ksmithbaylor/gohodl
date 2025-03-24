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

func handleBulkWithdrawFrom(label string) handlers.TransactionHandlerFunc {
	return func(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
		return handleBulkWithdraw(label, bundle, client, export)
	}
}

func handleBulkWithdraw(label string, bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	if config.Config.IsMyEvmAddressString(bundle.Info.From) {
		panic("Unexpected bulk withdraw from my own address")
	}

	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for bulk withdraw")
	}

	var received core.Amount
	var receivedTo common.Address

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for bulk withdraw")
		}
		for addr, amount := range transfers {
			if !amount.Value.IsPositive() {
				panic("Outflow for bulk withdraw")
			}
			received = *amount
			receivedTo = addr
		}
	}

	if received.Asset.Symbol == "" {
		panic("Nothing received for bulk withdraw")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCReceive,
		BaseCurrency: received.Asset.Symbol,
		BaseAmount:   received.Value,
		From:         label,
		To:           receivedTo.Hex(),
		Description:  fmt.Sprintf("withdraw %s from %s", received, label),
	}

	return export(ctcTx.ToCSV())
}
