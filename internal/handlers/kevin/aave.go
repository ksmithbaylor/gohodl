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

	if supplyEvent.Data["reserve"].(common.Address).Hex() != deposited.Asset.Identifier {
		panic("Different asset supplied than token movements would suggest for aave supply")
	}

	if supplyEvent.Data["amount"].(*big.Int).String() != deposited.Value.Shift(int32(deposited.Asset.Decimals)).String() {
		panic("Different amount supplied than token movements would suggest for aave supply")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0),
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
	return NOT_HANDLED
}

func handleAaveRepay(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveRepayWithATokens(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	return NOT_HANDLED
}

func handleAaveDeposit(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	// fmt.Printf("--------------------------- %s - %s\n", bundle.Info.Hash, bundle.Info.Network)
	// fmt.Println("aave deposit")
	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	// fmt.Printf("netTransfersOnlyMine: \n%v\n", netTransfersOnlyMine)

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

	// fmt.Println("deposited:", deposited)
	// fmt.Println("received: ", received)

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0),
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
	return NOT_HANDLED
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
