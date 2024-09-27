package kevin

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/k0kubun/pp/v3"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

const INSTADAPP_ORIGIN = "0x03d70891b8994feB6ccA7022B25c32be92ee3725"

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

	fmt.Println("Event: ")
	fmt.Printf("  Origin: %s\n", args.event.origin)
	fmt.Printf("  Sender: %s\n", args.event.sender)
	fmt.Println("  Sub-events:")
	for _, subEvent := range args.event.subEvents {
		fmt.Printf("    - Target: %s (%s)\n", subEvent.targetName, subEvent.target)
		fmt.Printf("      Selector: %s\n", subEvent.selector)
		fmt.Println("      Args:")
		for _, arg := range args.subEvent.args {
			switch arg := arg.(type) {
			case common.Address:
				fmt.Printf("        - address %s\n", arg)
			case []common.Address:
				fmt.Println("        - addresses")
				for _, addr := range arg {
					fmt.Printf("          - %s\n", addr)
				}
			case *big.Int:
				fmt.Printf("        - numeric %s\n", arg)
			default:
				pp.Println(arg)
				panic("Unknown instadapp sub-event arg type")
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

	if len(instadappEvents) == 1 {
		return handleSingleInstadappEvent(instadappEvents[0], bundle, client, export)
	}

	// TODO handle transactions with multiple instadapp events

	return NOT_HANDLED
}

func handleSingleInstadappEvent(
	event instadappEvent,
	bundle handlers.TransactionBundle,
	client *evm.Client,
	export handlers.CTCWriter,
) error {
	if event.origin.Hex() != INSTADAPP_ORIGIN {
		panic("Unknown origin for instadapp event")
	}

	if bundle.Info.From != event.sender.Hex() {
		panic("Instadapp sender does not match transaction signer")
	}

	netTransfers, err := evm_util.NetTokenTransfers(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	netTransfersOnlyMine, err := evm_util.NetTokenTransfersOnlyMine(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	err = nil

	for _, subEvent := range event.subEvents {
		args := instadappTargetHandlerArgs{
			event,
			subEvent,
			netTransfers,
			netTransfersOnlyMine,
			bundle,
			client,
			export,
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

	return err
}

func combineErrs(a, b error) error {
	if a == NOT_HANDLED || b == NOT_HANDLED {
		return NOT_HANDLED
	}

	return errors.Join(a, b)
}

func handleInstadappTargetBasicA(args instadappTargetHandlerArgs) error {
	if len(args.event.subEvents) > 1 {
		panic("Unexpected multiple instadapp events")
	}
	if args.subEvent.selector != "LogWithdraw(address,uint256,address,uint256,uint256)" {
		panic("Unknown BASIC-A selector: " + args.subEvent.selector)
	}
	if len(args.netTransfers) == 0 {
		panic("No transfers in instadapp withdrawal")
	}
	if len(args.netTransfers) > 1 {
		panic("Multiple transfers in instadapp withdrawal")
	}

	for asset, transfers := range args.netTransfers {
		if len(transfers) != 2 {
			panic("Extra net transfer in instadapp withdrawal")
		}

		dsaOutflow, ok := transfers[common.HexToAddress(args.bundle.Info.To)]
		if !ok {
			panic("No DSA outflow in instadapp withdrawal")
		}

		myInflow, ok := transfers[common.HexToAddress(args.bundle.Info.From)]
		if !ok {
			panic("No self-inflow in instadapp withdrawal")
		}

		if dsaOutflow.Value.Abs().Cmp(myInflow.Value.Abs()) != 0 {
			panic("Unbalanced instadapp withdrawal flows")
		}

		ctcTx := ctc_util.CTCTransaction{
			Timestamp:    time.Unix(int64(args.bundle.Block.Time), 0),
			Blockchain:   args.bundle.Info.Network,
			ID:           args.bundle.Info.Hash,
			Type:         ctc_util.CTCSend,
			BaseCurrency: asset.Symbol,
			BaseAmount:   myInflow.Value,
			From:         args.bundle.Info.To,
			To:           args.bundle.Info.From,
			Description: fmt.Sprintf("instadapp: withdraw %s to %s from dsa %s on %s",
				myInflow.String(),
				args.bundle.Info.From,
				args.bundle.Info.To,
				args.bundle.Info.Network,
			),
		}

		ctcTx.AddTransactionFeeIfMine(args.bundle.Info.From, args.bundle.Info.Network, args.bundle.Receipt)

		return args.export(ctcTx.ToCSV())
	}

	return NOT_HANDLED
}

func handleInstadappTargetAuthorityA(args instadappTargetHandlerArgs) error {
	if len(args.event.subEvents) > 1 {
		panic("Unexpected multiple instadapp events")
	}
	if args.subEvent.selector != "LogAddAuth(address,address)" {
		panic("Unknown AUTHORITY-A selector: " + args.subEvent.selector)
	}

	authorized := args.subEvent.args[1].(common.Address).Hex()
	authorizor := args.subEvent.args[0].(common.Address).Hex()
	if authorizor != args.bundle.Info.From {
		panic("Unexpected instadapp authorizor")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:  time.Unix(int64(args.bundle.Block.Time), 0),
		Blockchain: args.bundle.Info.Network,
		ID:         args.bundle.Info.Hash,
		From:       authorizor,
		To:         authorized,
		Type:       ctc_util.CTCApproval,
		Description: fmt.Sprintf("instadapp: authorize %s for dsa %s on %s",
			authorized,
			args.bundle.Info.To,
			args.bundle.Info.Network,
		),
	}

	ctcTx.AddTransactionFeeIfMine(args.bundle.Info.From, args.bundle.Info.Network, args.bundle.Receipt)

	return args.export(ctcTx.ToCSV())
}

func handleInstadappTargetAaveV2A(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetAaveClaimA(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetAaveClaimB(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTargetAaveV2ImportA(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
}

func handleInstadappTarget1inchA(args instadappTargetHandlerArgs) error {
	return NOT_HANDLED
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
