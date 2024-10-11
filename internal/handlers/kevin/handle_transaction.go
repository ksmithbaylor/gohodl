package kevin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ksmithbaylor/gohodl/internal/abis"
	"github.com/ksmithbaylor/gohodl/internal/evm"
	"github.com/ksmithbaylor/gohodl/internal/handlers"
	"golang.org/x/exp/slices"
)

var END_OF_2023 = 1704067199

var NOT_HANDLED = errors.New("transaction not handled")

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
	case info.Method == abis.ERC20_TRANSFER || info.Method == abis.ERC20_TRANSFER_FROM:
		handle = handleErc20Transfer
	case info.Method == abis.ERC20_APPROVE:
		handle = handleErc20Approve
	case info.Method == abis.INSTADAPP_CAST:
		handle = handleInstadapp
	case info.Method == abis.AAVE_SUPPLY:
		handle = handleAaveSupply
	case info.Method == abis.AAVE_BORROW:
		handle = handleAaveBorrow
	case info.Method == abis.AAVE_REPAY:
		handle = handleAaveRepay
	case info.Method == abis.AAVE_REPAY_WITH_A_TOKENS:
		handle = handleAaveRepayWithATokens
	case info.Method == abis.AAVE_DEPOSIT:
		handle = handleAaveDeposit
	case info.Method == abis.AAVE_WITHDRAW:
		handle = handleAaveWithdraw
	case info.Method == abis.AAVE_SET_USER_E_MODE:
		handle = handleAaveSetUserEMode
	case
		info.Method == abis.AAVE_CLAIM_REWARDS,
		info.Method == abis.AAVE_CLAIM_ALL_REWARDS:
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
		info.Method == abis.MOONWELL_NATIVE_MINT && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellMint
	case info.Method == abis.MOONWELL_BORROW && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellBorrow
	case info.Method == abis.MOONWELL_REPAY_BORROW && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellRepayBorrow
	case info.Method == abis.MOONWELL_REDEEM && strings.HasPrefix(info.Network, "moon"):
		handle = handleMoonwellRedeem
	case info.Method == abis.MOONWELL_STAKING_STAKE:
		handle = handleMoonwellStake
	case info.Method == abis.MOONWELL_STAKING_COOLDOWN:
		handle = handleMoonwellStakingCooldown
	case info.Method == abis.MOONWELL_STAKING_REDEEM:
		handle = handleMoonwellStakingRedeem
	case
		info.Method == abis.UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_TOKENS,
		info.Method == abis.UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_TOKENS,
		info.Method == abis.UNISWAP_V2_SWAP_EXACT_ETH_FOR_TOKENS,
		info.Method == abis.UNISWAP_V2_SWAP_TOKENS_FOR_EXACT_ETH,
		info.Method == abis.UNISWAP_V2_SWAP_EXACT_TOKENS_FOR_ETH,
		info.Method == abis.UNISWAP_V2_SWAP_ETH_FOR_EXACT_TOKENS,
		info.Method == abis.UNISWAP_UNIVERSAL_EXECUTE,
		info.Method == abis.UNISWAP_UNIVERSAL_EXECUTE_0,
		info.Method == abis.ONE_INCH_SWAP,
		info.Method == abis.PARASWAP_SIMPLE_BUY,
		info.Method == abis.PARASWAP_SIMPLE_SWAP,
		info.Method == abis.PARASWAP_MEGA_SWAP,
		info.Method == abis.PARASWAP_SWAP_ON_UNISWAP,
		info.Method == abis.PARASWAP_SWAP_ON_UNISWAP_V2_FORK:
		handle = handleTokenSwap
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
	case
		info.Method == abis.UNISWAP_V3_MULTICALL_0,
		info.Method == abis.UNISWAP_V3_MULTICALL_1:
		handle = handleUniswapMulticall
	case info.Method == abis.X_SQUARED_BUY_ITEM:
		handle = handleXSquaredBuyItem
	case info.Method == abis.X_SQUARED_SELL_ITEM:
		handle = handleXSquaredSellItem
	case info.Method == abis.FRIEND_TECH_BUY_SHARES:
		handle = handleFriendTechBuy
	case info.Method == abis.FRIEND_TECH_SELL_SHARES:
		handle = handleFriendTechSell
	case info.Time <= END_OF_2023 && slices.Contains(spamMethods, info.Method):
		// I verified each of these that happened before 2024, so they should just be ignored.
		return true, nil
	case info.Method == abis.WRAPPED_NATIVE_DEPOSIT && slices.Contains(
		wrappedNativeContracts,
		fmt.Sprintf("%s-%s", info.Network, info.To),
	):
		handle = handleWrappedNativeDeposit
	case info.Method == abis.WRAPPED_NATIVE_WITHDRAW && slices.Contains(
		wrappedNativeContracts,
		fmt.Sprintf("%s-%s", info.Network, info.To),
	):
		handle = handleWrappedNativeWithdraw
	case info.Method == "":
		client.OpenTransactionInExplorer(info.Hash)
	}

	if handle != nil {
		return true, readAndThen(handle)
	}

	return false, nil
}

var spamMethods = []string{
	"0x927f59ba", // mintBatch(address[])
}

var wrappedNativeContracts = []string{
	"ethereum-0xC02aaA39b223FE8D0A0e5C4F27eAD9083C756Cc2",
	"polygon-0x0d500B1d8E8eF31E21C99d1Db9A6444d3ADf1270",
	"avalanche-0xB31f66AA3C1e785363F0875A1B74E27b85FD66c7",
}
