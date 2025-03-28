package kevin

import (
	"time"

	"github.com/ksmithbaylor/gohodl/internal/ctc_util"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
)

var spamMethods = []string{
	"0x110bcd45", // mintItem(address,string)
	"0x12514bba", // transfer_iABlJaxlyCyqFbft((uint8,address,address,address,uint256)[])
	"0x12d94235", // batchTransferToken_10001(address[],uint256)
	"0x15270ace", // distribute(address,address[],uint256[])
	"0x163e1e61", // gift(address[])
	"0x2c10c112", // unknown
	"0x327ca788", // airDropBulk(address[],uint256)
	"0x3fe561cf", // transfer(address[],address)
	"0x41ed24a2", // unknown
	"0x441ff998", // unknown
	"0x4ee51a27", // airdropTokens(address[])
	"0x4f61d102", // unknown
	"0x512d7cfd", // batchTransferToken(address[],uint256)
	"0x520f3e69", // unknown
	"0x588d826a", // unknown
	"0x5c45079a", // dropToken(address,address[],uint256[])
	"0x62b74da5", // unknown
	"0x67243482", // airdrop(address[],uint256[])
	"0x6c6c9c84", // multisendTokenWithSignature(address,address[],uint256[],uint256,address,bytes,uint256)
	"0x6c9d713d", // unknown
	"0x6d244f2f", // unknown
	"0x6e56cd92", // unknown
	"0x729ad39e", // airdrop(address[])
	"0x74a72e41", // registerAddressesValue(address[],uint256)
	"0x7c8255db", // sendGifts(address[])
	"0x7f4d683a", // unknown
	"0x82947abe", // airdropERC20(address,address[],uint256[],uint256)
	"0x927f59ba", // mintBatch(address[])
	"0x9c96eec5", // Rewards(address _from,address[] _to,uint256 amount)
	"0xa8c6551f", // doAirDrop(address[],uint256)
	"0xb8ae5a2c", // adminMintAirdrop(address[])
	"0xbd075b84", // mint(address[])
	"0xc01ae5d3", // drop(address[],uint256[])
	"0xc204642c", // airdrop(address[],uint256)
	"0xc73a2d60", // disperseToken(address,address[],uint256[])
	"0xd43a632f", // reward(address[])
	"0xd57498ea", // test(address[])
	"0xeeb9052f", // AirDrop(address[],uint256)
	"0xfaf67b43", // unknown
}

func handleSpam(bundle handlers.TransactionBundle, client *evm.Client, export handlers.CTCWriter) error {
	ctcTx := &ctc_util.CTCTransaction{
		Timestamp:   time.Unix(int64(bundle.Block.Time), 0).UTC(),
		Blockchain:  bundle.Info.Network,
		ID:          bundle.Info.Hash,
		Type:        ctc_util.CTCSpam,
		Description: "spam transaction",
	}
	ctcTx.AddTransactionFeeIfMine(bundle.Info.From, bundle.Info.Network, bundle.Receipt)

	return export(ctcTx.ToCSV())
}
