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

var BENQI_CONTRACTS = []string{
	"0xe194c4c5aC32a3C9ffDb358d9Bfd523a0B6d1568",
	"0xc9e5999b8e75C3fEB117F6f73E664b9f3C8ca65C",
	"0x35Bd6aedA81a7E5FC7A7832490e71F757b0cD9Ce",
	"0xBEb5d47A3f720Ec0a390d04b4d41ED7d9688bC7F",
	"0x755c78D3bC25e297e8E010A2D1FCf49Cc43ADa21",
}

func handleBenqi(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	switch bundle.Info.Method {
	case "0xa0712d68", "0xb6b55f25": // mint(uint256), deposit(uint256)
		return handleBenqiMint(bundle, client, export, netTransfers)
	case "0xc5ebeaec": // borrow(uint256)
		return handleBenqiBorrow(bundle, client, export, netTransfers)
	case "0x0e752702": // repayBorrow(uint256)
		return handleBenqiRepay(bundle, client, export, netTransfers)
	case "0x852a12e3", "0x2e1a7d4d": // redeemUnderlying(uint256), withdraw(uint256)
		return handleBenqiRedeem(bundle, client, export, netTransfers)
	case "0xfdb5a03e": // reinvest()
		// Not quite accurate, but close enough
		return handleBenqiClaimRewards(bundle, client, export, netTransfers)
	}

	return NOT_HANDLED
}

// The below is mostly copied from the moonwell/aave handlers

func handleBenqiMint(bundle handlers.TransactionBundle, _ *evm.Client, export handlers.CTCWriter, netTransfers evm_util.NetTransfers) error {
	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for benqi mint")
	}

	var deposited core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for benqi mint")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for benqi mint")
			}
			if amount.Value.IsNegative() {
				deposited = amount.Neg()
			}
		}
	}

	if deposited.Asset.Symbol == "" {
		panic("No asset deposited for benqi mint")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralDeposit,
		BaseCurrency: deposited.Asset.Symbol,
		BaseAmount:   deposited.Value,
		From:         bundle.Info.From,
		To:           "benqi",
		Description:  fmt.Sprintf("benqi: supply %s", deposited),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleBenqiBorrow(bundle handlers.TransactionBundle, _ *evm.Client, export handlers.CTCWriter, netTransfers evm_util.NetTransfers) error {
	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for benqi borrow")
	}

	var borrowed core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for benqi borrow")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for benqi borrow")
			}
			if amount.Value.IsPositive() {
				borrowed = *amount
			}
		}
	}

	if borrowed.Asset.Symbol == "" {
		panic("No asset deposited for benqi mint")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBorrow,
		BaseCurrency: borrowed.Asset.Symbol,
		BaseAmount:   borrowed.Value,
		From:         "benqi",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("benqi: borrow %s", borrowed),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleBenqiRepay(bundle handlers.TransactionBundle, _ *evm.Client, export handlers.CTCWriter, netTransfers evm_util.NetTransfers) error {
	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for benqi borrow")
	}

	var repaid core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for benqi repay")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for benqi repay")
			}
			if amount.Value.IsNegative() {
				repaid = amount.Neg()
			}
		}
	}

	if repaid.Asset.Symbol == "" {
		panic("No asset sent for benqi repay")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCLoanRepayment,
		BaseCurrency: repaid.Asset.Symbol,
		BaseAmount:   repaid.Value,
		From:         bundle.Info.From,
		To:           "benqi",
		Description:  fmt.Sprintf("benqi: repay %s", repaid),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleBenqiRedeem(bundle handlers.TransactionBundle, _ *evm.Client, export handlers.CTCWriter, netTransfers evm_util.NetTransfers) error {
	if len(netTransfers) != 2 {
		panic("Unexpected net transfers for benqi redeem")
	}

	var withdrawn core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for benqi redeem")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for benqi redeem")
			}
			if amount.Value.IsPositive() {
				withdrawn = *amount
			}
		}
	}

	if withdrawn.Asset.Symbol == "" {
		panic("No asset withdrawn for benqi redeem")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCCollateralWithdrawal,
		BaseCurrency: withdrawn.Asset.Symbol,
		BaseAmount:   withdrawn.Value,
		From:         "benqi",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("benqi: withdraw %s", withdrawn),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}

func handleBenqiClaimRewards(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter, netTransfers evm_util.NetTransfers) error {
	if len(netTransfers) != 1 {
		panic("Unexpected net transfers for benqi claim rewards")
	}

	var claimed core.Amount

	for _, transfers := range netTransfers {
		if len(transfers) != 1 {
			panic("Unexpected net transfers for benqi claim rewards")
		}
		for addr, amount := range transfers {
			if addr.Hex() != bundle.Info.From {
				panic("Unexpected net transfers for benqi claim rewards")
			}
			if !amount.Value.IsPositive() {
				panic("Outflow for benqi claim rewards")
			}
			claimed = *amount
		}
	}

	if claimed.Asset.Symbol == "" {
		panic("Nothing claimed for benqi claim rewards")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCIncome,
		BaseCurrency: claimed.Asset.Symbol,
		BaseAmount:   claimed.Value,
		From:         "benqi",
		To:           bundle.Info.From,
		Description:  fmt.Sprintf("benqi: claim %s in rewards", claimed),
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
