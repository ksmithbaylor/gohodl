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

func handleBenqi(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	switch bundle.Info.Method {
	case "0xa0712d68": // mint(uint256)
		return handleBenqiMint(bundle, client, export, netTransfers)
	case "0xc5ebeaec": // borrow(uint256)
		return handleBenqiBorrow(bundle, client, export, netTransfers)
	case "0x0e752702": // repayBorrow(uint256)
		return handleBenqiRepay(bundle, client, export, netTransfers)
	case "0x852a12e3": // redeemUnderlying(uint256)
		return handleBenqiRedeem(bundle, client, export, netTransfers)
	}

	return NOT_HANDLED
}

// The below is mostly copied from the moonwell handlers

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
	printHeader(bundle)
	fmt.Println(bundle.Info.Method)
	netTransfers.Print()

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
	ctcTx.Print()

	return export(ctcTx.ToCSV())
}
