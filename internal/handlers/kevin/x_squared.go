package kevin

import (
	"crypto/md5"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleXSquaredBuyItem(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for xsquared buy")
	}

	var received core.Amount
	var receivedTo common.Address
	var paid core.Amount

	for asset, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for xsquared buy")
		}
		if asset.Kind != core.EvmNative {
			panic("Non-native asset transfer for xsquared buy")
		}
		for addr, amount := range transfers {
			if amount.Value.IsNegative() {
				if addr.Hex() != bundle.Info.From {
					panic("Unexpected net transfers for xsquared buy")
				}
				if received.Asset.Symbol != "" {
					panic("Buy and receive in same xsquared buy transaction")
				}
				paid = amount.Neg()
			} else {
				if paid.Asset.Symbol != "" {
					panic("Buy and receive in same xsquared buy transaction")
				}
				received = *amount
				receivedTo = addr
			}
		}
	}

	if received.Asset.Symbol != "" {
		ctcTx := &ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash,
			Type:         ctc_util.CTCIncome,
			BaseCurrency: received.Asset.Symbol,
			BaseAmount:   received.Value,
			From:         "xsquared",
			To:           receivedTo.Hex(),
			Description:  "xsquared: someone bought my item",
		}

		if config.Config.IsMyEvmAddressString(bundle.Info.From) {
			panic("Bought my own item for xsquared buy")
		}

		return export(ctcTx.ToCSV())
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.XSquaredAbi)
	if err != nil {
		return err
	}

	if len(events) != 1 {
		panic("Unexpected events for xsquared buy")
	}

	event := events[0]

	isBuy := event.Data["isBuy"].(bool)
	if !isBuy {
		panic("Sell event for xsquared buy")
	}

	quantityString := event.Data["quantity"].(*big.Int).String()
	quantity, err := decimal.NewFromString(quantityString)
	if err != nil {
		panic("Invalid quantity string for xsquared buy")
	}

	collectionBytes, ok := event.Data["collection"].([32]uint8)
	if !ok {
		panic("Invalid collection for xsquared buy")
	}
	collection := common.Bytes2Hex(collectionBytes[:])

	itemBytes, ok := event.Data["item"].([32]uint8)
	if !ok {
		panic("Invalid item for xsquared buy")
	}
	item := common.Bytes2Hex(itemBytes[:])

	description := xsquaredDescription(collection, item)

	ctcTx := &ctc_util.CTCTransaction{
		Timestamp:     time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:    bundle.Info.Network,
		ID:            bundle.Info.Hash,
		Type:          ctc_util.CTCBuy,
		BaseCurrency:  description,
		BaseAmount:    quantity,
		QuoteCurrency: paid.Asset.Symbol,
		QuoteAmount:   paid.Value,
		From:          "xsquared",
		To:            bundle.Info.From,
		Description: fmt.Sprintf("xsquared: bought %s %s with %s",
			quantity,
			description,
			paid,
		),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleXSquaredSellItem(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for xsquared sell")
	}

	var received core.Amount
	var receivedTo common.Address
	var saleProceeds core.Amount

	for asset, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for xsquared sell")
		}
		if asset.Kind != core.EvmNative {
			panic("Non-native asset transfer for xsquared sell")
		}
		for addr, amount := range transfers {
			if !amount.Value.IsPositive() {
				panic("Unexpected outflow for xsquared sell")
			}
			if addr.Hex() == bundle.Info.From {
				if received.Asset.Symbol != "" {
					panic("Sell and receive in same xsquared sell transaction")
				}
				saleProceeds = *amount
			} else {
				if saleProceeds.Asset.Symbol != "" {
					panic("Sell and receive in same xsquared sell transaction")
				}
				received = *amount
				receivedTo = addr
			}
		}
	}

	if received.Asset.Symbol != "" {
		ctcTx := &ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash,
			Type:         ctc_util.CTCIncome,
			BaseCurrency: received.Asset.Symbol,
			BaseAmount:   received.Value,
			From:         "xsquared",
			To:           receivedTo.Hex(),
			Description:  "xsquared: someone sold my item",
		}

		if config.Config.IsMyEvmAddressString(bundle.Info.From) {
			panic("Sold my own item for xsquared sell")
		}

		return export(ctcTx.ToCSV())
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.XSquaredAbi)
	if err != nil {
		return err
	}

	if len(events) != 1 {
		panic("Unexpected events for xsquared sell")
	}

	event := events[0]

	isBuy := event.Data["isBuy"].(bool)
	if isBuy {
		panic("Buy event for xsquared sell")
	}

	quantityString := event.Data["quantity"].(*big.Int).String()
	quantity, err := decimal.NewFromString(quantityString)
	if err != nil {
		panic("Invalid quantity string for xsquared sell")
	}

	collectionBytes, ok := event.Data["collection"].([32]uint8)
	if !ok {
		panic("Invalid collection for xsquared sell")
	}
	collection := common.Bytes2Hex(collectionBytes[:])

	itemBytes, ok := event.Data["item"].([32]uint8)
	if !ok {
		panic("Invalid item for xsquared sell")
	}
	item := common.Bytes2Hex(itemBytes[:])

	description := xsquaredDescription(collection, item)

	ctcTx := &ctc_util.CTCTransaction{
		Timestamp:     time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:    bundle.Info.Network,
		ID:            bundle.Info.Hash,
		Type:          ctc_util.CTCSell,
		BaseCurrency:  description,
		BaseAmount:    quantity,
		QuoteCurrency: saleProceeds.Asset.Symbol,
		QuoteAmount:   saleProceeds.Value,
		From:          "xsquared",
		To:            bundle.Info.From,
		Description: fmt.Sprintf("xsquared: sold %s %s for %s",
			quantity,
			description,
			saleProceeds,
		),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func xsquaredDescription(collection, item string) string {
	hash := md5.Sum([]byte(collection + item))
	return fmt.Sprintf("xsquared-%x", hash[:4])
}
