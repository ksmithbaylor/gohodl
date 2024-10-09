package kevin

import (
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/core"
	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/evm_util"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

const BASE_BRIDGE = "0x49048044D57e1C92A77f79988d21Fa8fAF74E97e"
const AVAX_OLD_BRIDGE = "0x50Ff3B278fCC70ec7A9465063d68029AB460eA04"
const AVAX_BITCOIN_BRIDGE = "0xF5163f69F97B221d50347Dd79382F11c6401f1a1"

func handleNoData(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	switch bundle.Tx.Type() {
	case types.LegacyTxType, types.DynamicFeeTxType:
		return handleRegularNoDataTx(bundle, client, export)
	case types.DepositTxType:
		return handleDepositTx(bundle, client, export)
	case types.AccessListTxType:
		panic("Access list transactions not implemented")
	case types.BlobTxType:
		panic("Blob transactions not implemented")
	default:
		panic(fmt.Sprintf("Unimplemented transaction type: %d\n", bundle.Tx.Type()))
	}
}

func handleDepositTx(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	nativeAsset, err := client.NativeAsset()
	if err != nil {
		panic("No native asset found")
	}

	amount, err := nativeAsset.WithAtomicStringValue(bundle.Tx.Value().String())
	if err != nil {
		panic("Could not parse value")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		Type:         ctc_util.CTCBridgeIn,
		BaseCurrency: nativeAsset.Symbol,
		BaseAmount:   amount.Value,
		From:         bundle.Info.From,
		To:           bundle.Info.To,
		Description:  fmt.Sprintf("bridge %s to %s", amount.String(), bundle.Info.Network),
	}

	return export(ctcTx.ToCSV())
}

func handleRegularNoDataTx(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	nativeAsset, err := client.NativeAsset()
	if err != nil {
		panic("No native asset found")
	}

	amount, err := nativeAsset.WithAtomicStringValue(bundle.Tx.Value().String())
	if err != nil {
		panic("Could not parse value")
	}

	ctcTx := ctc_util.CTCTransaction{
		Timestamp:    time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:   bundle.Info.Network,
		ID:           bundle.Info.Hash,
		BaseCurrency: nativeAsset.Symbol,
		BaseAmount:   amount.Value,
		From:         bundle.Info.From,
		To:           bundle.Info.To,
		Description: fmt.Sprintf("transfer %s from %s to %s on %s",
			amount.String(),
			bundle.Info.From,
			bundle.Info.To,
			bundle.Info.Network,
		),
	}

	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)
	fromMe := config.Config.IsMyEvmAddressString(bundle.Info.From)
	toMe := config.Config.IsMyEvmAddressString(bundle.Info.To)

	switch {
	case bundle.Info.To == BASE_BRIDGE:
		ctcTx.Type = ctc_util.CTCBridgeOut
		ctcTx.Description = fmt.Sprintf("bridge %s to base", amount.String())
	case bundle.Info.From == AVAX_OLD_BRIDGE || bundle.Info.From == AVAX_BITCOIN_BRIDGE:
		if amount.Asset.Symbol != "AVAX" || amount.Asset.Kind != core.AssetKind("evm_native") {
			panic("Non-airdrop transfer from avalanche bridge")
		}
		ctcTx.Type = ctc_util.CTCIncome
		ctcTx.Description = "airdrop from avalanche bridge"
	case bundle.Info.From == evm.ZERO_ADDRESS && bundle.Info.To == evm.ZERO_ADDRESS:
		err := handleBatchBridge(bundle, client, &ctcTx)
		if err != nil {
			return err
		}
	case bundle.Info.To == TO_1:
		ctcTx.Type = TYPE_1
		ctcTx.Description = DESCRIPTION_1
	case toMe && !fromMe:
		ctcTx.Type = ctc_util.CTCReceive
	case fromMe:
		ctcTx.Type = ctc_util.CTCSend
	default:
		panic("Found irrelevant transaction, not from/to any of my addresses")
	}

	return export(ctcTx.ToCSV())
}

func handleBatchBridge(bundle handlers.TransactionBundle, client *evm.Client, ctcTx *ctc_util.CTCTransaction) error {
	netTransfers, err := evm_util.NetTokenTransfersOnlyMine(
		client,
		bundle.Info,
		bundle.Receipt.Logs,
	)
	if err != nil {
		return err
	}

	if len(netTransfers) == 0 {
		panic("No net transfers to my addresses found in bridge transaction")
	}
	if len(netTransfers) > 1 {
		panic("Multiple assets transferred in bridge transaction")
	}

	for _, transfers := range netTransfers {
		if len(transfers) == 0 {
			panic("No net transfers to my addresses found in bridge transaction")
		}
		if len(transfers) > 1 {
			panic("Multiple of my addresses had net transfers in bridge transaction")
		}
		for addr, amount := range transfers {
			ctcTx.Type = ctc_util.CTCBridgeIn
			ctcTx.BaseCurrency = amount.Asset.Symbol
			ctcTx.BaseAmount = amount.Value
			ctcTx.From = addr.Hex()
			ctcTx.To = addr.Hex()
			ctcTx.Description = fmt.Sprintf("bridge %s to %s on %s",
				amount.String(),
				addr.Hex(),
				bundle.Info.Network,
			)
		}
	}

	return nil
}
