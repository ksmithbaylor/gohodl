package kevin

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/k0kubun/pp/v3"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

type instadappEvent struct {
	origin    common.Address
	sender    common.Address
	subEvents []instadappSubEvent
}

type instadappSubEvent struct {
	selector   string
	name       string
	data       []uint8
	args       []any
	target     common.Address
	targetName string
}

type instadappTargetHandlerArgs struct {
	totalSubEvents       int
	subEventNumber       int
	events               []instadappEvent
	event                instadappEvent
	subEvent             instadappSubEvent
	netTransfers         evm_util.NetTransfers
	netTransfersOnlyMine evm_util.NetTransfers
	bundle               handlers.TransactionBundle
	client               *evm.Client
	export               handlers.CTCWriter
}

func (args instadappTargetHandlerArgs) Print() {
	fmt.Printf("--------- %s: %s -> %s on %s\n",
		args.bundle.Info.Hash,
		args.bundle.Info.From,
		args.bundle.Info.To,
		args.bundle.Info.Network,
	)

	for _, event := range args.events {
		fmt.Println("- Event: ")
		fmt.Printf("    Origin: %s\n", event.origin)
		fmt.Printf("    Sender: %s\n", event.sender)
		fmt.Println("    Sub-events:")
		for _, subEvent := range event.subEvents {
			fmt.Printf("      - Target: %s (%s)\n", subEvent.targetName, subEvent.target)
			fmt.Printf("        Selector: %s\n", subEvent.selector)
			fmt.Println("        Args:")
			for _, arg := range subEvent.args {
				switch arg := arg.(type) {
				case common.Address:
					fmt.Printf("          - address %s\n", arg)
				case []common.Address:
					fmt.Println("          - addresses:")
					for _, addr := range arg {
						fmt.Printf("            - %s\n", addr)
					}
				case *big.Int:
					fmt.Printf("          - numeric %s\n", arg)
				case []*big.Int:
					fmt.Println("          - numerics:")
					for _, num := range arg {
						fmt.Printf("            - %s\n", num)
					}
				case bool:
					fmt.Printf("          - bool %t\n", arg)
				default:
					pp.Println(arg)
					panic("Unknown instadapp sub-event arg type")
				}
			}
		}
	}

	fmt.Println("Net transfers:")
	fmt.Println(args.netTransfers)

	fmt.Println("Net transfers (only mine):")
	fmt.Println(args.netTransfersOnlyMine)
}

func handleInstadapp(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.InstadappAbi)
	if err != nil {
		return err
	}

	instadappEvents := make([]instadappEvent, len(events))

	for i, event := range events {
		if event.Name != "LogCast" && event.Name != "LogCastMigrate" {
			panic("Unexpected event emitted from instadapp call")
		}

		origin := event.Data["origin"].(common.Address)
		sender := event.Data["sender"].(common.Address)
		value := event.Data["value"].(*big.Int)
		eventNames := event.Data["eventNames"].([]string)
		eventParams := event.Data["eventParams"].([][]uint8)
		targets := event.Data["targets"].([]common.Address)
		targetNames := event.Data["targetsNames"].([]string)

		if value.String() != "0" {
			panic("Unexpected value in instadapp cast log")
		}

		if len(eventNames) != len(targets) || len(eventParams) != len(targets) || len(targetNames) != len(targets) {
			panic("Mismatched sub-events in instadapp call")
		}

		subEvents := make([]instadappSubEvent, len(eventNames))

		for i := 0; i < len(eventNames); i++ {
			name := eventNames[i]
			args := make([]any, 0)
			if name != "" {
				name, args, err = abis.DecodeAdhoc(eventNames[i], eventParams[i])
				if err != nil {
					return err
				}
			}
			subEvents[i] = instadappSubEvent{
				selector:   eventNames[i],
				data:       eventParams[i],
				name:       name,
				args:       args,
				target:     targets[i],
				targetName: targetNames[i],
			}
		}

		instadappEvents[i] = instadappEvent{
			origin:    origin,
			sender:    sender,
			subEvents: subEvents,
		}
	}

	if len(instadappEvents) == 0 {
		panic("No instadapp events in instadapp transaction")
	}

	return handleInstadappEvents(instadappEvents, bundle, client, export)
}

// TODO
// - Figure out if I need to re-order txs containing flash loans, such as
//   0x3ea65f97fe4aa224ce0561e81c18a65cf41bb65452b7892b18aa726ea180097d on avalanche
// - Figure out event sender not matching signer
// - Figure out what the origin means, if anything

func handleInstadappEvents(
	events []instadappEvent,
	bundle handlers.TransactionBundle,
	client *evm.Client,
	export handlers.CTCWriter,
) error {
	netTransfers, err := evm_util.NetTokenTransfers(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	totalSubEvents := 0
	for _, event := range events {
		totalSubEvents += len(event.subEvents)
	}

	subEventNumber := 0

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:   time.Unix(int64(bundle.Block.Time), 0),
		Blockchain:  bundle.Info.Network,
		ID:          fmt.Sprintf("%s-%d", bundle.Info.Hash, subEventNumber),
		From:        bundle.Info.From,
		Type:        ctc_util.CTCFee,
		Description: "instadapp: record network fee separately from individual events",
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)
	err = combineErrs(err, export(ctcTx.ToCSV()))

	// TODO: remove and handle the rest of the more complicated ones
	if len(events) > 1 {
		return NOT_HANDLED
	}

	for _, event := range events {
		for _, subEvent := range event.subEvents {
			subEventNumber++

			args := instadappTargetHandlerArgs{
				totalSubEvents,
				subEventNumber,
				events,
				event,
				subEvent,
				netTransfers,
				netTransfersOnlyMine,
				bundle,
				client,
				export,
			}

			if bundle.Info.Hash == TX_2 {
				return handleTx2(args)
			}

			switch subEvent.targetName {
			case "BASIC-A":
				err = combineErrs(err, handleInstadappTargetBasicA(args))
			case "AUTHORITY-A":
				err = combineErrs(err, handleInstadappTargetAuthorityA(args))
			case "AAVE-V2-A":
				err = combineErrs(err, handleInstadappTargetAaveV2A(args))
			case "AAVE-CLAIM-A":
				err = combineErrs(err, handleInstadappTargetAaveClaimA(args))
			case "AAVE-CLAIM-B":
				err = combineErrs(err, handleInstadappTargetAaveClaimB(args))
			case "AAVE-V2-IMPORT-A":
				err = combineErrs(err, handleInstadappTargetAaveV2ImportA(args))
			case "INSTAPOOL-A":
				err = combineErrs(err, handleInstadappTargetInstapoolA(args))
			case "INSTAPOOL-B":
				err = combineErrs(err, handleInstadappTargetInstapoolB(args))
			case "INSTAPOOL-C":
				err = combineErrs(err, handleInstadappTargetInstapoolC(args))
			case "1INCH-A":
				err = combineErrs(err, handleInstadappTarget1inchA(args))
			case "1INCH-V4-A":
				err = combineErrs(err, handleInstadappTarget1inchV4A(args))
			case "PARASWAP-A":
				err = combineErrs(err, handleInstadappTargetParaswapA(args))
			case "PARASWAP-V5-A":
				err = combineErrs(err, handleInstadappTargetParaswapV5A(args))
			default:
				panic("Unknown instadapp target: " + subEvent.targetName)
			}
		}
	}

	return err
}

func combineErrs(a, b error) error {
	if a == NOT_HANDLED || b == NOT_HANDLED {
		return NOT_HANDLED
	}

	return errors.Join(a, b)
}

func instadappTxID(args instadappTargetHandlerArgs) string {
	if args.totalSubEvents == 1 {
		return args.bundle.Info.Hash
	}

	return fmt.Sprintf("%s-%d", args.bundle.Info.Hash, args.subEventNumber)
}

func handleInstadappTargetBasicA(args instadappTargetHandlerArgs) error {
	if args.subEvent.selector != "LogWithdraw(address,uint256,address,uint256,uint256)" {
		panic("Unknown BASIC-A selector: " + args.subEvent.selector)
	}

	token := args.subEvent.args[0].(common.Address)
	value := args.subEvent.args[1].(*big.Int)
	dest := args.subEvent.args[2].(common.Address)
	dsa := args.bundle.Info.To

	asset, err := args.client.TokenAsset(token, true)
	if err != nil {
		return err
	}

	amount, err := asset.WithAtomicStringValue(value.String())
	if err != nil {
		return err
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(args.bundle.Block.Time), 0),
		Blockchain:   args.bundle.Info.Network,
		ID:           instadappTxID(args),
		Type:         ctc_util.CTCSend,
		BaseCurrency: asset.Symbol,
		BaseAmount:   amount.Value,
		From:         dsa,
		To:           dest.Hex(),
		Description: fmt.Sprintf("instadapp: withdraw %s to %s from dsa %s on %s",
			amount,
			dsa,
			dest,
			args.bundle.Info.Network,
		),
	}

	return args.export(ctcTx.ToCSV())
}

func handleInstadappTargetAuthorityA(args instadappTargetHandlerArgs) error {
	if args.subEvent.selector != "LogAddAuth(address,address)" {
		panic("Unknown AUTHORITY-A selector: " + args.subEvent.selector)
	}

	// Nothing to do here tax-wise, and transaction fee is already handled
	return nil
}

func handleInstadappTargetAaveV2A(args instadappTargetHandlerArgs) error {
	var ctcType ctc_util.CTCTransactionType
	var description string
	var from string
	var to string

	// This doesn't impact anything tax-wise, and the transaction fee is already handled
	if args.subEvent.selector == "LogEnableCollateral(address[])" {
		return nil
	}

	if !slices.Contains(
		[]string{
			"LogDeposit(address,uint256,uint256,uint256)",
			"LogBorrow(address,uint256,uint256,uint256,uint256)",
			"LogPayback(address,uint256,uint256,uint256,uint256)",
			"LogWithdraw(address,uint256,uint256,uint256)",
		},
		args.subEvent.selector,
	) {
		panic("Unknown AAVE-V2-A selector: " + args.subEvent.selector)
	}

	// All four of the above sub-events have the token as the first argument, and
	// the value as the second argument
	token := args.subEvent.args[0].(common.Address)
	value := args.subEvent.args[1].(*big.Int)
	dsa := args.bundle.Info.To
	aaveConnector := args.subEvent.target.Hex()

	asset, err := args.client.TokenAsset(token, true)
	if err != nil {
		return err
	}

	amount, err := asset.WithAtomicStringValue(value.String())
	if err != nil {
		return err
	}

	switch args.subEvent.selector {
	case "LogDeposit(address,uint256,uint256,uint256)":
		ctcType = ctc_util.CTCCollateralDeposit
		description = fmt.Sprintf("instadapp: deposit %s as collateral from %s to aave", amount, dsa)
		from = dsa
		to = aaveConnector
	case "LogBorrow(address,uint256,uint256,uint256,uint256)":
		ctcType = ctc_util.CTCBorrow
		description = fmt.Sprintf("instadapp: borrow %s from aave to %s", amount, dsa)
		from = aaveConnector
		to = dsa
	case "LogPayback(address,uint256,uint256,uint256,uint256)":
		ctcType = ctc_util.CTCLoanRepayment
		description = fmt.Sprintf("instadapp: pay back %s to aave from %s", amount, dsa)
		from = dsa
		to = aaveConnector
	case "LogWithdraw(address,uint256,uint256,uint256)":
		ctcType = ctc_util.CTCCollateralWithdrawal
		description = fmt.Sprintf("instadapp: withdraw %s as collateral from aave to %s", amount, dsa)
		from = aaveConnector
		to = dsa
	default:
		panic("Unknown AAVE-V2-A selector: " + args.subEvent.selector)
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(args.bundle.Block.Time), 0),
		Blockchain:   args.bundle.Info.Network,
		ID:           instadappTxID(args),
		Type:         ctcType,
		BaseCurrency: asset.Symbol,
		BaseAmount:   amount.Value,
		From:         from,
		To:           to,
		Description:  description,
	}

	return args.export(ctcTx.ToCSV())
}

func handleInstadappTargetAaveClaimA(args instadappTargetHandlerArgs) error {
	if args.totalSubEvents > 1 {
		panic("Unexpected multiple instadapp events for AAVE-CLAIM-A")
	}
	if args.subEvent.selector != "LogClaimed(address[],uint256,uint256,uint256)" {
		panic("Unknown AAVE-CLAIM-A selector: " + args.subEvent.selector)
	}
	if len(args.netTransfersOnlyMine) > 1 {
		panic("Multiple assets claimed for instadapp")
	}
	if len(args.netTransfersOnlyMine) == 0 {
		ctcTx := ctc_util.CTCTransaction{
			Timestamp:  time.Unix(int64(args.bundle.Block.Time), 0),
			Blockchain: args.bundle.Info.Network,
			ID:         instadappTxID(args),
			From:       args.bundle.Info.From,
			Type:       ctc_util.CTCFee,
			Description: fmt.Sprintf("instadapp: claim aave rewards for dsa %s on %s, but nothing claimed",
				args.bundle.Info.To,
				args.bundle.Info.Network,
			),
		}

		return args.export(ctcTx.ToCSV())
	}

	for asset, transfers := range args.netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("Extra net transfer in instadapp claim")
		}

		dsaInflow, ok := transfers[common.HexToAddress(args.bundle.Info.To)]
		if !ok {
			panic("No DSA inflow in instadapp claim")
		}

		ctcTx := ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(args.bundle.Block.Time), 0),
			Blockchain:   args.bundle.Info.Network,
			ID:           instadappTxID(args),
			Type:         ctc_util.CTCIncome,
			BaseCurrency: asset.Symbol,
			BaseAmount:   dsaInflow.Value,
			From:         "Instadapp Aave",
			To:           args.bundle.Info.To,
			Description: fmt.Sprintf("instadapp: claim %s in rewards for dsa %s on %s",
				dsaInflow,
				args.bundle.Info.To,
				args.bundle.Info.Network,
			),
		}

		return args.export(ctcTx.ToCSV())
	}

	return NOT_HANDLED
}

func handleInstadappTargetAaveClaimB(args instadappTargetHandlerArgs) error {
	if args.totalSubEvents > 1 {
		panic("Unexpected multiple instadapp events for AAVE-CLAIM-B")
	}
	if args.subEvent.selector != "LogAaveV2Claim(address,address[],address[],uint256[],uint256[])" {
		panic("Unknown AAVE-CLAIM-B selector: " + args.subEvent.selector)
	}
	if len(args.netTransfersOnlyMine) != 1 {
		panic("Multiple assets claimed for instadapp")
	}

	for asset, transfers := range args.netTransfersOnlyMine {
		if len(transfers) != 1 {
			panic("Extra net transfer in instadapp claim")
		}

		dsaInflow, ok := transfers[common.HexToAddress(args.bundle.Info.To)]
		if !ok {
			panic("No DSA inflow in instadapp claim")
		}

		ctcTx := ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(args.bundle.Block.Time), 0),
			Blockchain:   args.bundle.Info.Network,
			ID:           instadappTxID(args),
			Type:         ctc_util.CTCIncome,
			BaseCurrency: asset.Symbol,
			BaseAmount:   dsaInflow.Value,
			From:         "Instadapp Aave",
			To:           args.bundle.Info.To,
			Description: fmt.Sprintf("instadapp: claim %s in rewards for dsa %s on %s",
				dsaInflow,
				args.bundle.Info.To,
				args.bundle.Info.Network,
			),
		}

		return args.export(ctcTx.ToCSV())
	}

	return NOT_HANDLED
}

func handleInstadappTargetAaveV2ImportA(args instadappTargetHandlerArgs) error {
	if args.subEvent.selector != "LogAaveV2Import(address,bool,address[],address[],uint256[],uint256[],uint256[])" {
		panic("Unknown AAVE-V2-IMPORT-A selector: " + args.subEvent.selector)
	}

	// Nothing to do here from a tax perspective, just moving a position around
	return nil
}

func handleInstadappTargetInstapoolA(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetInstapoolB(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetInstapoolC(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTarget1inchA(args instadappTargetHandlerArgs) error {
	if args.subEvent.selector != "LogSell(address,address,uint256,uint256,uint256,uint256)" {
		panic("Unknown 1INCH-A selector: " + args.subEvent.selector)
	}

	boughtToken := args.subEvent.args[0].(common.Address)
	soldToken := args.subEvent.args[1].(common.Address)
	boughtValue := args.subEvent.args[2].(*big.Int)
	soldValue := args.subEvent.args[3].(*big.Int)
	connector := args.subEvent.target
	dsa := args.bundle.Info.To

	boughtAsset, err := args.client.TokenAsset(boughtToken, true)
	if err != nil {
		return err
	}

	soldAsset, err := args.client.TokenAsset(soldToken, true)
	if err != nil {
		return err
	}

	boughtAmount, err := boughtAsset.WithAtomicStringValue(boughtValue.String())
	if err != nil {
		return err
	}

	soldAmount, err := soldAsset.WithAtomicStringValue(soldValue.String())
	if err != nil {
		return err
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:     time.Unix(int64(args.bundle.Block.Time), 0),
		Blockchain:    args.bundle.Info.Network,
		ID:            instadappTxID(args),
		Type:          ctc_util.CTCSell,
		BaseCurrency:  soldAsset.Symbol,
		BaseAmount:    soldAmount.Value,
		QuoteCurrency: boughtAsset.Symbol,
		QuoteAmount:   boughtAmount.Value,
		From:          dsa,
		To:            connector.Hex(),
		Description:   fmt.Sprintf("instadapp: sell %s for %s", soldAmount, boughtAmount),
	}
	ctcTx.Print()

	return args.export(ctcTx.ToCSV())
}

func handleInstadappTarget1inchV4A(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetParaswapA(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetParaswapV5A(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}
