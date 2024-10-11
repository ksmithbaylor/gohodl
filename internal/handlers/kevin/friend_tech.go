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

func handleFriendTechBuy(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	var received core.Amount
	var receivedTo common.Address
	var paid core.Amount

	if len(netTransfers) > 1 {
		panic("Unexpected net transfers for friend.tech buy")
	}

	if len(netTransfers) == 0 {
		nativeAsset, err := client.NativeAsset()
		if err != nil {
			return err
		}
		paid = nativeAsset.WithAtomicValue(0)
	}

	for asset, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for friend.tech buy")
		}
		if asset.Kind != core.EvmNative {
			panic("Non-native asset transfer for friend.tech buy")
		}
		for addr, amount := range transfers {
			if amount.Value.IsNegative() {
				if addr.Hex() != bundle.Info.From {
					panic("Unexpected net transfers for friend.tech buy")
				}
				if received.Asset.Symbol != "" {
					panic("Buy and receive in same friend.tech buy transaction")
				}
				paid = amount.Neg()
			} else {
				if paid.Asset.Symbol != "" {
					panic("Buy and receive in same friend.tech buy transaction")
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
			From:         "friend.tech",
			To:           receivedTo.Hex(),
			Description:  "friend.tech: someone bought my shares",
		}

		if config.Config.IsMyEvmAddressString(bundle.Info.From) {
			panic("Bought my own shares for friend.tech buy")
		}

		return export(ctcTx.ToCSV())
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.FriendTechAbi)
	if err != nil {
		return err
	}

	if len(events) != 1 || events[0].Name != "Trade" {
		panic("Unexpected events for friend.tech buy")
	}

	event := events[0]

	isBuy := event.Data["isBuy"].(bool)
	if !isBuy {
		panic("Sell event for friend.tech buy")
	}

	quantityString := event.Data["shareAmount"].(*big.Int).String()
	quantity, err := decimal.NewFromString(quantityString)
	if err != nil {
		panic("Invalid shareAmount string for friend.tech buy")
	}

	subject, ok := event.Data["subject"].(common.Address)
	if !ok {
		panic("Invalid subject for friend.tech buy")
	}

	description := friendTechDescription(subject)

	ctcTx := &ctc_util.CTCTransaction{
		Timestamp:     time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:    bundle.Info.Network,
		ID:            bundle.Info.Hash,
		Type:          ctc_util.CTCBuy,
		BaseCurrency:  description,
		BaseAmount:    quantity,
		QuoteCurrency: paid.Asset.Symbol,
		QuoteAmount:   paid.Value,
		From:          "friend.tech",
		To:            bundle.Info.From,
		Description: fmt.Sprintf("friend.tech: bought %s shares of %s with %s",
			quantity,
			subject.Hex(),
			paid,
		),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleFriendTechSell(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for friend.tech sell")
	}

	var received core.Amount
	var receivedTo common.Address
	var saleProceeds core.Amount

	for asset, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for friend.tech sell")
		}
		if asset.Kind != core.EvmNative {
			panic("Non-native asset transfer for friend.tech sell")
		}
		for addr, amount := range transfers {
			if !amount.Value.IsPositive() {
				panic("Unexpected outflow for friend.tech sell")
			}
			if addr.Hex() == bundle.Info.From {
				if received.Asset.Symbol != "" {
					panic("Sell and receive in same friend.tech sell transaction")
				}
				saleProceeds = *amount
			} else {
				if saleProceeds.Asset.Symbol != "" {
					panic("Sell and receive in same friend.tech sell transaction")
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
			From:         "friend.tech",
			To:           receivedTo.Hex(),
			Description:  "friend.tech: someone sold my shares",
		}

		if config.Config.IsMyEvmAddressString(bundle.Info.From) {
			panic("Sold my own shares for friend.tech sell")
		}

		return export(ctcTx.ToCSV())
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.FriendTechAbi)
	if err != nil {
		return err
	}

	if len(events) != 1 || events[0].Name != "Trade" {
		panic("Unexpected events for friend.tech sell")
	}

	event := events[0]

	isBuy := event.Data["isBuy"].(bool)
	if isBuy {
		panic("Buy event for friend.tech sell")
	}

	quantityString := event.Data["shareAmount"].(*big.Int).String()
	quantity, err := decimal.NewFromString(quantityString)
	if err != nil {
		panic("Invalid shareAmount string for friend.tech sell")
	}

	subject, ok := event.Data["subject"].(common.Address)
	if !ok {
		panic("Invalid subject for friend.tech buy")
	}

	description := friendTechDescription(subject)

	ctcTx := &ctc_util.CTCTransaction{
		Timestamp:     time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:    bundle.Info.Network,
		ID:            bundle.Info.Hash,
		Type:          ctc_util.CTCSell,
		BaseCurrency:  description,
		BaseAmount:    quantity,
		QuoteCurrency: saleProceeds.Asset.Symbol,
		QuoteAmount:   saleProceeds.Value,
		From:          "friend.tech",
		To:            bundle.Info.From,
		Description: fmt.Sprintf("friend.tech: sold %s shares of %s for %s",
			quantity,
			subject.Hex(),
			saleProceeds,
		),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func friendTechDescription(subject common.Address) string {
	hash := md5.Sum(subject[:])
	return fmt.Sprintf("friend.tech-%x", hash[:4])
}
