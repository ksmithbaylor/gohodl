package kevin

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

func handleAaveSupply(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfersOnlyMine) != 2 {
		panic("More than 2 net transfers for aave supply")
	}

	var deposited core.Amount
	var received core.Amount

	for _, transfers := range netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("More than 1 transfer for an asset for aave supply")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Transfer to/from the wrong address for aave supply")
			}
			if amount.Value.IsNegative() {
				deposited = amount.Neg()
			} else if amount.Value.IsPositive() {
				received = *amount
			} else {
				panic("Zero-value transfers for aave supply")
			}
		}
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.AaveAbi)
	if err != nil {
		return err
	}

	var supplyEvent evm.ParsedEvent
	for _, event := range events {
		if event.Name == "Supply" {
			supplyEvent = event
		}
	}

	if supplyEvent.Name == "" {
		panic("No supply event for aave supply")
	}

	if supplyEvent.Data["reserve"].(common.Address).Hex() != deposited.Asset.Identifier {
		panic("Different asset supplied than token movements would suggest for aave supply")
	}

	if supplyEvent.Data["amount"].(*big.Int).String() != deposited.Value.Shift(int32(deposited.Asset.Decimals)).String() {
		panic("Different amount supplied than token movements would suggest for aave supply")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralDeposit,
		BaseCurrency: deposited.Asset.Symbol,
		BaseAmount:   deposited.Value,
		From:         bundle.Info.From,
		To:           "aave",
		Description: fmt.Sprintf("aave: supply %s, receive receipt token (%s)",
			deposited,
			received,
		),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleAaveBorrow(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfersOnlyMine) != 2 {
		panic("More than 2 net transfers for aave borrow")
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.AaveAbi)
	if err != nil {
		return err
	}

	var borrowEvent evm.ParsedEvent
	for _, event := range events {
		if event.Name == "Borrow" || event.Name == "Borrow0" {
			borrowEvent = event
		}
	}

	if borrowEvent.Name == "" {
		panic("No borrow event for aave borrow")
	}

	borrowedTokenAddress := borrowEvent.Data["reserve"].(common.Address).Hex()
	borrowedAmount := borrowEvent.Data["amount"].(*big.Int).String()

	var borrowed core.Amount

	for _, transfers := range netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("More than 1 transfer for an asset for aave borrow")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Transfer to/from the wrong address for aave borrow")
			}
			if amount.Asset.Identifier == borrowedTokenAddress {
				borrowed = *amount
			}
		}
	}

	if borrowedAmount != borrowed.Value.Shift(int32(borrowed.Asset.Decimals)).String() {
		panic("Different amount borrowed vs received for aave borrow")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBorrow,
		BaseCurrency: borrowed.Asset.Symbol,
		BaseAmount:   borrowed.Value,
		From:         "aave",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("aave: borrow %s", borrowed),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleAaveRepay(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfersOnlyMine) != 2 {
		panic("More than 2 net transfers for aave repay")
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.AaveAbi)
	if err != nil {
		return err
	}

	var repayEvent evm.ParsedEvent
	for _, event := range events {
		if event.Name == "Repay" || event.Name == "Repay0" {
			repayEvent = event
		}
	}

	if repayEvent.Name == "" {
		panic("No repay event for aave repay")
	}

	repaidTokenAddress := repayEvent.Data["reserve"].(common.Address).Hex()
	repaidAmount := repayEvent.Data["amount"].(*big.Int).String()

	var repaid core.Amount

	for _, transfers := range netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("More than 1 transfer for an asset for aave repay")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Transfer to/from the wrong address for aave repay")
			}
			if amount.Asset.Identifier == repaidTokenAddress {
				repaid = amount.Neg()
			}
		}
	}

	if repaidAmount != repaid.Value.Shift(int32(repaid.Asset.Decimals)).String() {
		panic("Different amount repaid vs sent for aave repay")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCLoanRepayment,
		BaseCurrency: repaid.Asset.Symbol,
		BaseAmount:   repaid.Value,
		From:         bundle.Info.From,
		To:           "aave",
		Description:  fmt.Sprintf("aave: repay %s", repaid),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleAaveRepayWithATokens(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfersOnlyMine) != 2 {
		panic("More than 2 net transfers for aave repay with atokens")
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.AaveAbi)
	if err != nil {
		return err
	}

	var repayEvent evm.ParsedEvent
	for _, event := range events {
		if event.Name == "Repay" || event.Name == "Repay0" {
			repayEvent = event
		}
	}

	if repayEvent.Name == "" {
		panic("No repay event for aave repay with atokens")
	}

	repaidTokenAddress := repayEvent.Data["reserve"].(common.Address)
	repaidAmount := repayEvent.Data["amount"].(*big.Int).String()

	repaidAsset, err := client.TokenAsset(repaidTokenAddress)
	if err != nil {
		return err
	}

	repaid, err := repaidAsset.WithAtomicStringValue(repaidAmount)
	if err != nil {
		return err
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCLoanRepayment,
		BaseCurrency: repaid.Asset.Symbol,
		BaseAmount:   repaid.Value,
		From:         bundle.Info.From,
		To:           "aave",
		Description:  fmt.Sprintf("aave: repay %s", repaid),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleAaveDeposit(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfersOnlyMine) != 2 {
		panic("More than 2 net transfers for aave deposit")
	}

	var deposited core.Amount
	var received core.Amount

	for _, transfers := range netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("More than 1 transfer for an asset for aave deposit")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Transfer to/from the wrong address for aave deposit")
			}
			if amount.Value.IsNegative() {
				deposited = amount.Neg()
			} else if amount.Value.IsPositive() {
				received = *amount
			} else {
				panic("Zero-value transfers for aave deposit")
			}
		}
	}

	if !deposited.Value.Equal(received.Value) {
		panic("Different amount deposited vs received for aave deposit")
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.AaveAbi)
	if err != nil {
		return err
	}

	var depositEvent evm.ParsedEvent
	for _, event := range events {
		if event.Name == "Deposit" {
			depositEvent = event
		}
	}

	if depositEvent.Name == "" {
		panic("No deposit event for aave deposit")
	}

	if depositEvent.Data["reserve"].(common.Address).Hex() != deposited.Asset.Identifier {
		panic("Different asset deposited than token movements would suggest for aave deposit")
	}

	if depositEvent.Data["amount"].(*big.Int).String() != deposited.Value.Shift(int32(deposited.Asset.Decimals)).String() {
		panic("Different amount deposited than token movements would suggest for aave deposit")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralDeposit,
		BaseCurrency: deposited.Asset.Symbol,
		BaseAmount:   deposited.Value,
		From:         bundle.Info.From,
		To:           "aave",
		Description: fmt.Sprintf("aave: deposit %s, receive receipt token (%s)",
			deposited,
			received,
		),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleAaveWithdraw(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	if len(netTransfersOnlyMine) != 2 {
		panic("More than 2 net transfers for aave withdrawal")
	}

	var withdrawn core.Amount

	for _, transfers := range netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("More than 1 transfer for an asset for aave withdrawal")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Transfer to/from the wrong address for aave withdrawal")
			}
			if amount.Value.IsPositive() {
				withdrawn = *amount
			}
		}
	}

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.AaveAbi)
	if err != nil {
		return err
	}

	var withdrawEvent evm.ParsedEvent
	for _, event := range events {
		if event.Name == "Withdraw" {
			withdrawEvent = event
		}
	}

	if withdrawEvent.Name == "" {
		panic("No withdraw event for aave withdrawal")
	}

	if withdrawEvent.Data["reserve"].(common.Address).Hex() != withdrawn.Asset.Identifier {
		panic("Different asset withdrawn than token movements would suggest for aave withdrawal")
	}

	if withdrawEvent.Data["amount"].(*big.Int).String() != withdrawn.Value.Shift(int32(withdrawn.Asset.Decimals)).String() {
		panic("Different amount withdrawn than token movements would suggest for aave withdrawal")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralWithdrawal,
		BaseCurrency: withdrawn.Asset.Symbol,
		BaseAmount:   withdrawn.Value,
		From:         "aave",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("aave: withdraw %s", withdrawn),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleAaveSetUserEMode(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	ctcTx := ctc_util.NewFeeTransaction(
		bundle.Block.Time,
		bundle.Info.Network,
		bundle.Info.Hash,
		bundle.Info.From,
		"aave: set user e-mode",
		bundle.Receipt,
	)
	return export(ctcTx.ToCSV())
}
