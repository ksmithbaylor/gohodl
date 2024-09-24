package evm

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/core"
)

type TokenTransfers map[common.Address]*core.Amount
type NetTransfers map[core.Asset]TokenTransfers

type transfer struct {
	from   common.Address
	to     common.Address
	amount core.Amount
}

func NetTokenTransfers(client *Client, info *TxInfo, logs []*types.Log) (NetTransfers, error) {
	transfers := make([]transfer, 0)
	netTransfers := make(NetTransfers)

	nativeAsset, err := client.NativeAsset()
	if err != nil {
		return nil, err
	}

	if info.Value != "0" {
		amount, err := nativeAsset.WithAtomicStringValue(info.Value)
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer{
			common.HexToAddress(info.From),
			common.HexToAddress(info.To),
			amount,
		})
	}

	erc20Logs, err := ParseKnownEvents(info.Network, logs, abis.Erc20Abi)
	if err != nil {
		return nil, err
	}

	for _, log := range erc20Logs {
		if log.Name != "Transfer" {
			continue
		}
		from, ok := log.Data["from"]
		if !ok {
			continue
		}
		to, ok := log.Data["to"]
		if !ok {
			continue
		}
		value, ok := log.Data["value"]
		if !ok {
			continue
		}
		fromAddr, ok := from.(common.Address)
		if !ok {
			continue
		}
		toAddr, ok := to.(common.Address)
		if !ok {
			continue
		}
		valueBig, ok := value.(*big.Int)
		if !ok {
			continue
		}

		asset, err := client.TokenAsset(log.Contract)
		if err != nil {
			return nil, err
		}

		amount, err := asset.WithAtomicStringValue(valueBig.String())
		if err != nil {
			return nil, err
		}

		transfers = append(transfers, transfer{fromAddr, toAddr, amount})
	}

	internalTxs, _, err := client.GetInternalTransactions(info.Hash)
	if err != nil {
		return nil, err
	}

	for _, tx := range internalTxs {
		amount, err := nativeAsset.WithAtomicStringValue(tx.Value.Int().String())
		if err != nil {
			return nil, err
		}
		transfers = append(transfers, transfer{
			common.HexToAddress(tx.From),
			common.HexToAddress(tx.To),
			amount,
		})
	}

	for _, transfer := range transfers {
		asset := transfer.amount.Asset

		if netTransfers[asset] == nil {
			netTransfers[asset] = make(TokenTransfers)
		}

		if netTransfers[asset][transfer.from] == nil {
			zero := asset.WithAtomicValue(0)
			netTransfers[asset][transfer.from] = &zero
		}

		if netTransfers[asset][transfer.to] == nil {
			zero := asset.WithAtomicValue(0)
			netTransfers[asset][transfer.to] = &zero
		}

		newFromAmount, err := netTransfers[asset][transfer.from].Sub(transfer.amount)
		if err != nil {
			return nil, err
		}
		netTransfers[asset][transfer.from] = &newFromAmount

		newToAmount, err := netTransfers[asset][transfer.to].Add(transfer.amount)
		if err != nil {
			return nil, err
		}
		netTransfers[asset][transfer.to] = &newToAmount
	}

	return netTransfers, nil
}
