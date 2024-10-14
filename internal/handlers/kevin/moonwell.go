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
			if amount.Value.IsPositive() {
				rewardAmount = amount
			} else {
				panic("Outflow for moonwell rewards claim")
			}
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
		ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)
	}

	return export(ctcTx.ToCSV())
}

func handleMoonwellMint(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for moonwell mint")
	}

	var deposited core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for moonwell mint")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for moonwell mint")
			}
			if amount.Value.IsNegative() {
				deposited = amount.Neg()
			}
		}
	}

	if deposited.Asset.Symbol == "" {
		panic("No asset deposited for moonwell mint")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralDeposit,
		BaseCurrency: deposited.Asset.Symbol,
		BaseAmount:   deposited.Value,
		From:         bundle.Info.From,
		To:           "moonwell",
		Description:  fmt.Sprintf("moonwell: supply %s", deposited),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleMoonwellBorrow(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for moonwell borrow")
	}

	var borrowed core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for moonwell borrow")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for moonwell borrow")
			}
			if amount.Value.IsPositive() {
				borrowed = *amount
			}
		}
	}

	if borrowed.Asset.Symbol == "" {
		panic("No asset received for moonwell borrow")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBorrow,
		BaseCurrency: borrowed.Asset.Symbol,
		BaseAmount:   borrowed.Value,
		From:         "moonwell",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("moonwell: borrow %s", borrowed),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleMoonwellRepayBorrow(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for moonwell borrow")
	}

	var repaid core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for moonwell repay")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for moonwell repay")
			}
			if amount.Value.IsNegative() {
				repaid = amount.Neg()
			}
		}
	}

	if repaid.Asset.Symbol == "" {
		panic("No asset sent for moonwell repay")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCLoanRepayment,
		BaseCurrency: repaid.Asset.Symbol,
		BaseAmount:   repaid.Value,
		From:         bundle.Info.From,
		To:           "moonwell",
		Description:  fmt.Sprintf("moonwell: repay %s", repaid),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleMoonwellRedeem(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for moonwell redeem")
	}

	var withdrawn core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for moonwell redeem")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for moonwell redeem")
			}
			if amount.Value.IsPositive() {
				withdrawn = *amount
			}
		}
	}

	if withdrawn.Asset.Symbol == "" {
		panic("No asset withdrawn for moonwell redeem")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralWithdrawal,
		BaseCurrency: withdrawn.Asset.Symbol,
		BaseAmount:   withdrawn.Value,
		From:         "moonwell",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("moonwell: withdraw %s", withdrawn),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleMoonwellStake(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for moonwell stake")
	}

	var staked core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for moonwell stake")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for moonwell stake")
			}
			if amount.Value.IsNegative() {
				staked = amount.Neg()
			}
		}
	}

	if staked.Asset.Symbol == "" {
		panic("No asset sent for moonwell stake")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCStakingDeposit,
		BaseCurrency: staked.Asset.Symbol,
		BaseAmount:   staked.Value,
		From:         bundle.Info.From,
		To:           "moonwell",
		Description:  fmt.Sprintf("moonwell: stake %s", staked),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleMoonwellStakingCooldown(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	// This just starts a timer for unstaking, so the fee is the only thing that needs to be handled
	ctcTx := ctc_util.NewFeeTransaction(
		bundle.Block.Time,
		bundle.Info.Network,
		bundle.Info.Hash,
		bundle.Info.From,
		"moonwell: start unstaking cooldown",
		bundle.Receipt,
	)

	return export(ctcTx.ToCSV())
}

func handleMoonwellStakingRedeem(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for moonwell staking redeem")
	}

	var withdrawn core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for moonwell staking redeem")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for moonwell staking redeem")
			}
			if amount.Value.IsPositive() {
				withdrawn = *amount
			}
		}
	}

	if withdrawn.Asset.Symbol == "" {
		panic("No asset withdrawn for moonwell staking redeem")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCStakingWithdrawal,
		BaseCurrency: withdrawn.Asset.Symbol,
		BaseAmount:   withdrawn.Value,
		From:         "moonwell",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("moonwell: withdraw stake of %s", withdrawn),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
