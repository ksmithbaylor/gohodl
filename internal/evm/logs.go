package evm

import (
	"fmt"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/k0kubun/pp/v3"
	"github.com/ksmithbaylor/gohodl/internal/util"
)

type ParsedEvent struct {
	Contract common.Address
	Name     string
	Data     map[string]any
}

func (pe ParsedEvent) Print() {
	pp.Println(pe)
}

var KNOWN_STRANGE_CONTRACTS = []string{
	"0xC36442b4a4522E871399CD717aBDD847Ab11FE88",
}

func ParseKnownEvents(network string, logs []*types.Log, contractAbi abi.ABI) ([]ParsedEvent, error) {
	events := make([]ParsedEvent, 0)

	for _, log := range logs {
		tokenID := fmt.Sprintf("%s-%s", network, log.Address.Hex())
		if slices.Contains(SPAM_TOKENS, tokenID) {
			util.Debugf("SPAM FOUND, skipping (%s)\n", tokenID)
			continue
		}

		event, err := contractAbi.EventByID(log.Topics[0])
		if err != nil {
			if strings.Contains(err.Error(), "no event with id") {
				continue
			}
			return nil, fmt.Errorf("Could not parse event signature: %w", err)
		}

		indexedArgs := make([]abi.Argument, 0)
		for _, input := range event.Inputs {
			if input.Indexed {
				indexedArgs = append(indexedArgs, input)
			}
		}

		eventData := make(map[string]any)

		err = abi.ParseTopicsIntoMap(eventData, indexedArgs, log.Topics[1:])
		if err != nil {
			if slices.Contains(KNOWN_STRANGE_CONTRACTS, log.Address.Hex()) {
				continue
			}
			return nil, fmt.Errorf("Could not unpack indexed log topics: %w", err)
		}

		err = contractAbi.UnpackIntoMap(eventData, event.Name, log.Data)
		if err != nil {
			return nil, fmt.Errorf("Could not unpack unindexed log data: %w", err)
		}

		events = append(events, ParsedEvent{
			Contract: log.Address,
			Name:     event.Name,
			Data:     eventData,
		})
	}

	return events, nil
}

var SPAM_TOKENS = []string{
	"base-0x0F510911e1f53B1A53893649007d6f3B9B6860D6",
}
