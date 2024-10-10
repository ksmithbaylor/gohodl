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

func handleTokenSwap(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
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
		Description:   fmt.Sprintf("uniswap (or fork): sell %s for %s", sold, bought),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleUniswapAddLiquidity(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 3 {
		panic("Unexpected net transfers for uniswap liquidity add")
	}

	var tokenA *core.Amount
	var tokenB *core.Amount
	var lpToken *core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for uniswap liquidity add")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for uniswap liquidity add")
			}
			if amount.Value.IsPositive() {
				if lpToken != nil {
					panic("Multiple token inflows for uniswap liquidity add")
				}
				lpToken = amount
			} else if amount.Value.IsNegative() {
				outflow := amount.Neg()
				if tokenA == nil {
					tokenA = &outflow
				} else if tokenB == nil {
					tokenB = &outflow
				} else {
					panic("More than two tokens provided for uniswap liquidity add")
				}
			} else {
				panic("Zero-value transfer for uniswap liquidity add")
			}
		}
	}

	if tokenA == nil || tokenB == nil || lpToken == nil {
		panic("Confusing asset state for uniswap liquidity add")
	}

	ctcTxs := []ctc_util.CTCTransaction{
		*ctc_util.NewFeeTransaction(
			bundle.Block.Time,
			bundle.Info.Network,
			bundle.Info.Hash+"-1",
			bundle.Info.From,
			"uniswap add liquidity: transaction fee",
			bundle.Receipt,
		),
		{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash + "-2",
			Type:         ctc_util.CTCAddLiquidity,
			BaseCurrency: tokenA.Asset.Symbol,
			BaseAmount:   tokenA.Value,
			Description:  fmt.Sprintf("uniswap add liquidity: deposit %s", tokenA),
		},
		{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash + "-3",
			Type:         ctc_util.CTCAddLiquidity,
			BaseCurrency: tokenB.Asset.Symbol,
			BaseAmount:   tokenB.Value,
			Description:  fmt.Sprintf("uniswap add liquidity: deposit %s", tokenB),
		},
		{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash + "-3",
			Type:         ctc_util.CTCReceiveLPToken,
			BaseCurrency: fmt.Sprintf("%s-%s", lpToken.Asset.Identifier, lpToken.Asset.Symbol),
			BaseAmount:   lpToken.Value,
			Description:  fmt.Sprintf("uniswap add liquidity: receive lp token %s", lpToken),
		},
	}

	err = nil
	for _, ctcTx := range ctcTxs {
		err = combineErrs(err, export(ctcTx.ToCSV()))
	}

	return err
}

func handleUniswapRemoveLiquidity(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if bundle.Info.Hash == TX_4 {
		fixTx4NetTransfers(client, netTransfers)
	}

	if len(netTransfers) != 3 {
		panic("Unexpected net transfers for uniswap liquidity remove")
	}

	var tokenA *core.Amount
	var tokenB *core.Amount
	var lpToken *core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for uniswap liquidity remove")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for uniswap liquidity remove")
			}
			if amount.Value.IsNegative() {
				if lpToken != nil {
					panic("Multiple token outflows for uniswap liquidity remove")
				}
				outflow := amount.Neg()
				lpToken = &outflow
			} else if amount.Value.IsPositive() {
				if tokenA == nil {
					tokenA = amount
				} else if tokenB == nil {
					tokenB = amount
				} else {
					panic("More than two tokens received for uniswap liquidity remove")
				}
			} else {
				panic("Zero-value transfer for uniswap liquidity remove")
			}
		}
	}

	if tokenA == nil || tokenB == nil || lpToken == nil {
		panic("Confusing asset state for uniswap liquidity remove")
	}

	ctcTxs := []ctc_util.CTCTransaction{
		*ctc_util.NewFeeTransaction(
			bundle.Block.Time,
			bundle.Info.Network,
			bundle.Info.Hash+"-1",
			bundle.Info.From,
			"uniswap remove liquidity: transaction fee",
			bundle.Receipt,
		),
		{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash + "-2",
			Type:         ctc_util.CTCRemoveLiquidity,
			BaseCurrency: tokenA.Asset.Symbol,
			BaseAmount:   tokenA.Value,
			Description:  fmt.Sprintf("uniswap remove liquidity: receive %s", tokenA),
		},
		{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash + "-3",
			Type:         ctc_util.CTCRemoveLiquidity,
			BaseCurrency: tokenB.Asset.Symbol,
			BaseAmount:   tokenB.Value,
			Description:  fmt.Sprintf("uniswap remove liquidity: receive %s", tokenB),
		},
		{
			Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
			Blockchain:   bundle.Info.Network,
			ID:           bundle.Info.Hash + "-3",
			Type:         ctc_util.CTCReturnLPToken,
			BaseCurrency: fmt.Sprintf("%s-%s", lpToken.Asset.Identifier, lpToken.Asset.Symbol),
			BaseAmount:   lpToken.Value,
			Description:  fmt.Sprintf("uniswap remove liquidity: burn lp token %s", lpToken),
		},
	}

	err = nil
	for _, ctcTx := range ctcTxs {
		err = combineErrs(err, export(ctcTx.ToCSV()))
	}

	return err
}
