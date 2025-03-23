package kevin

import (
	"fmt"
	"time"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleSpamDrop(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if config.Config.IsMyEvmAddressString(bundle.Info.From) {
		panic("I sent spamdrop?")
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for spamdrop")
	}

	var to string
	var received *core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for spamdrop")
		}
		for addr, amount := range transfers {
			if !config.Config.IsMyEvmAddress(addr) {
				panic("Irrelevant transfer for spamdrop")
			}
			if amount.Value.IsNegative() {
				panic("Negative transfer for spamdrop")
			}
			to = addr.Hex()
			received = amount
		}
	}

	if to == "" || received == nil {
		panic("Nothing received for spamdrop")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCAirdrop,
		BaseCurrency: "spam-" + received.Asset.Symbol,
		BaseAmount:   received.Value,
		From:         "Unknown",
		To:           to,
		Description:  fmt.Sprintf("spamdrop: %s received %s", to, received),
	}

	return export(ctcTx.ToCSV())
}
