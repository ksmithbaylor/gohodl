package kevin

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/k0kubun/pp/v3"

	"github.com/ksmithbaylor/gohodl/internal/abis"
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

func handleInstadapp(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	fmt.Printf("--------- %s - %s\n", bundle.Info.Hash, bundle.Info.Network)

	events, err := evm.ParseKnownEvents(bundle.Info.Network, bundle.Receipt.Logs, abis.InstadappAbi)
	if err != nil {
		return err
	}

	instadappEvents := make([]instadappEvent, len(events))

	for i, event := range events {
		if event.Name != "LogCast" {
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
					pp.Println(event)
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

	dsa := common.HexToAddress(bundle.Info.To)
	signer := common.HexToAddress(bundle.Info.From)
	fmt.Printf("%s -> %s\n", signer, dsa)

	_, err := evm_util.NetTokenTransfers(client, bundle.Info, bundle.Receipt.Logs)
	if err != nil {
		return err
	}

	err = nil

	for _, subEvent := range event.subEvents {
		switch subEvent.targetName {
		case "BASIC-A":
			err = errors.Join(err, NOT_HANDLED)
		case "AUTHORITY-A":
			err = errors.Join(err, NOT_HANDLED)
		case "WETH-A":
			err = errors.Join(err, NOT_HANDLED)
		case "HOP-A":
			err = errors.Join(err, NOT_HANDLED)
		case "INSTAPOOL-A":
			err = errors.Join(err, NOT_HANDLED)
		case "AAVE-V2-A":
			err = errors.Join(err, NOT_HANDLED)
		case "AAVE-V2-IMPORT-A":
			err = errors.Join(err, NOT_HANDLED)
		case "AAVE-CLAIM-A":
			err = errors.Join(err, NOT_HANDLED)
		case "AAVE-CLAIM-B":
			err = errors.Join(err, NOT_HANDLED)
		case "AAVE-V3-A":
			err = errors.Join(err, NOT_HANDLED)
		case "AAVE-V3-CLAIM-A":
			err = errors.Join(err, NOT_HANDLED)
		case "UNISWAP-V3-A":
			err = errors.Join(err, NOT_HANDLED)
		case "1INCH-A":
			err = errors.Join(err, NOT_HANDLED)
		case "1INCH-V4-A":
			err = errors.Join(err, NOT_HANDLED)
		case "PARASWAP-A":
			err = errors.Join(err, NOT_HANDLED)
		case "PARASWAP-V5-A":
			err = errors.Join(err, NOT_HANDLED)
		case "SWAP-AGGREGATOR-A":
			err = errors.Join(err, NOT_HANDLED)
		default:
			panic("Unknown instadapp target: " + subEvent.targetName)
		}
	}

	return NOT_HANDLED
}
