package kevin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/config"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
	"golang.org/x/exp/slices"
)

var _ fmt.Stringer // Allow commenting and uncommenting printlns

var WRAPPED_NATIVE_CONTRACTS = []string{
	"ethereum-0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
	"base-0x4200000000000000000000000000000000000006",
	"polygon-0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
	"avalanche-0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7",
	"fantom-0x21be370D5312f44cB42ce377BC9b8a0cEF1A4C83",
}

var NOT_HANDLED = errors.New("transaction not handled")
var END_OF_2023 = 1704067199
var END_OF_2024 = 1735689599
var END_OF_2025 = 1767225599

var Implementation = personalHandler(struct{}{})

type personalHandler struct{}

func (h personalHandler) HandleTransaction(
	info *evm.TxInfo,
	client *evm.Client,
	readTransactionBundle handlers.TransactionReader,
	export handlers.CTCWriter,
) (bool, error) {
	readAndThen := func(handle handlers.TransactionHandlerFunc) error {
		tx, receipt, block, err := readTransactionBundle(info.Network, info.Hash)
		if err != nil {
			return err
		}

		bundle := handlers.TransactionBundle{
			Info:    info,
			Tx:      tx,
			Receipt: receipt,
			Block:   block,
		}
		return handle(bundle, client, export)
	}

	var handle handlers.TransactionHandlerFunc

	switch {
	case !info.Success:
		handle = handleFailed
	case info.Method == "":
		handle = handleNoData
	case slices.Contains(spamContracts, info.To):
		return true, nil
	case
		slices.Contains(spamMethods, info.Method) &&
			!config.Config.IsMyEvmAddressString(info.From):
		if info.Time > END_OF_2023 && info.Time <= END_OF_2025 {
			return true, nil // Verified all through 2025, spam
		} else {
			client.OpenTransactionInExplorer(info.Hash, true)
			return false, nil
		}
	case info.Method == abis.ERC20_APPROVE:
		handle = handleErc20Approve
	case info.Method == abis.ERC20_TRANSFER || info.Method == abis.ERC20_TRANSFER_FROM:
		handle = handleErc20Transfer
	case
		info.Method == "0x1a1da075",
		info.Method == "0xca350aa6":
		handle = handleBulkWithdrawFrom("coinbase")
	case
		info.Method == abis.UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_TOKENS,
		info.Method == abis.UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_TOKENS,
		info.Method == abis.UNISWAP_V2_SWAP_EXACT_ETH_FOR_TOKENS,
		info.Method == abis.UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_ETH,
		info.Method == abis.UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_ETH,
		info.Method == abis.UNISWAP_V2_SWAP_ETH_FOR_EXACT_TOKENS,
		info.Method == abis.UNISWAP_UNIVERSAL_EXECUTE,
		info.Method == abis.UNISWAP_UNIVERSAL_EXECUTE_0:
		handle = handleTokenSwapLabeled("uniswap")
	case
		info.Method == abis.AAVE_SUPPLY,
		info.Method == "0x474cf53d": // depositETH(address,address,uint16)
		handle = handleAaveSupply
	case info.Method == abis.AAVE_DEPOSIT:
		handle = handleAaveDeposit
	case info.Method == abis.AAVE_BORROW:
		handle = handleAaveBorrow
	case info.Method == abis.AAVE_REPAY,
		info.Method == "0x02c5fcf8": // repayETH(address,uint256,uint256,address)
		handle = handleAaveRepay
	case info.Method == abis.AAVE_REPAY_WITH_A_TOKENS:
		handle = handleAaveRepayWithATokens
	case info.Method == abis.AAVE_SET_USER_E_MODE:
		handle = handleAaveSetUserEMode
	case
		info.Method == abis.AAVE_WITHDRAW,
		info.Method == "0x80500d20": // withdrawETH(address,uint256,address)
		handle = handleAaveWithdraw
	case
		info.Method == abis.AAVE_CLAIM_REWARDS,
		info.Method == abis.AAVE_CLAIM_ALL_REWARDS,
		info.Method == "0x3111e7b3": // claimRewards(address[],uint256,address)
		handle = handleAaveClaimRewards
	case info.Method == abis.MOONWELL_ENTER_MARKETS:
		handle = handleMoonwellEnterMarkets
	case
		info.Method == abis.MOONWELL_CLAIM_REWARD,
		info.Method == abis.MOONWELL_CLAIM_REWARD_0,
		info.Method == abis.MOONWELL_STAKING_CLAIM:
		handle = handleMoonwellClaimReward
	case
		info.Method == abis.MOONWELL_MINT && strings.HasPrefix(info.Network, "moon"),
		info.Method == abis.MOONWELL_NATIVE_MINT && strings.HasPrefix(info.Network, "moon"),
		info.Method == "0x6a627842" && info.Network == "base":
		handle = handleMoonwellMint
	case
		info.Method == abis.MOONWELL_BORROW && strings.HasPrefix(info.Network, "moon"),
		info.Method == "0xc5ebeaec" && info.Network == "base":
		handle = handleMoonwellBorrow
	case
		info.Method == abis.MOONWELL_REPAY_BORROW && strings.HasPrefix(info.Network, "moon"),
		info.Method == "0x4e4d9fea" && strings.HasPrefix(info.Network, "moon"), // repayBorrow()
		info.Method == "0x0e752702" && strings.HasPrefix(info.Network, "base"): // repayBorrow(uint256)
		handle = handleMoonwellRepayBorrow
	case
		info.Method == abis.MOONWELL_REDEEM && strings.HasPrefix(info.Network, "moon"),
		info.Method == "0xdb006a75" && info.Network == "base",
		info.Method == "0x7bde82f2" && info.Network == "base":
		handle = handleMoonwellRedeem
	case info.Method == abis.MOONWELL_STAKING_STAKE:
		handle = handleMoonwellStake
	case info.Method == abis.MOONWELL_STAKING_COOLDOWN:
		handle = handleMoonwellStakingCooldown
	case info.Method == abis.MOONWELL_STAKING_REDEEM:
		handle = handleMoonwellStakingRedeem
	case slices.Contains(
		WRAPPED_NATIVE_CONTRACTS,
		fmt.Sprintf("%s-%s", info.Network, info.To),
	) && (info.Method == abis.WRAPPED_NATIVE_DEPOSIT || info.Method == abis.WRAPPED_NATIVE_WITHDRAW):
		handle = handleTokenSwapLabeled("wrapped native")
	case
		info.Method == abis.UNISWAP_V2_ADD_LIQUIDITY,
		info.Method == abis.UNISWAP_V2_ADD_LIQUIDITY_ETH:
		handle = handleUniswapAddLiquidity
	case
		info.Method == abis.UNISWAP_V2_REMOVE_LIQUIDITY_ETH,
		info.Method == abis.UNISWAP_V2_REMOVE_LIQUIDITY_PERMIT,
		info.Method == abis.UNISWAP_V2_REMOVE_LIQUIDITY_ETH_PERMIT,
		info.Method == abis.UNISWAP_V2_REMOVE_LIQUIDITY_ETH_PERMIT_FOTT:
		handle = handleUniswapRemoveLiquidity
	case info.Method == "0x65b2489b":
		handle = handleTokenSwapLabeled("curve")
	case info.Method == "0xe21fd0e9":
		handle = handleTokenSwapLabeled("kyberswap")
	case info.Method == "0x999b6464":
		handle = handleTokenSwapLabeled("rainbow")
	case info.Method == "0x12aa3caf":
		handle = handleTokenSwapLabeled("1inch")
	case info.Method == "0x415565b0":
		handle = handleTokenSwapLabeled("0x")
	case info.Method == abis.FRIEND_TECH_BUY_SHARES:
		handle = handleFriendTechBuy
	case info.Method == abis.FRIEND_TECH_SELL_SHARES:
		handle = handleFriendTechSell
	case info.Method == "0x52c7f8dc":
		handle = handleRewardWithLabel("XEN Crypto")
	case info.Method == "0x56781388":
		handle = handleMiscWithLabel("moonwell governance vote")
	case info.Method == "0x2046d075":
		handle = handleRewardWithLabel("misc reward")
	default:
		handle = handleOneOff
	}

	if handle != nil {
		return true, readAndThen(handle)
	}

	return false, nil
}
