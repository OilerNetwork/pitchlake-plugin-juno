package db

import (
	"junoplugin/models"
	"math/big"

	"gorm.io/gorm"
)

func (dbc *DB) DepositOrWithdrawIndex(vaultAddress, lpAddress string, lpUnlocked, vaultUnlocked models.BigInt, blockNumber uint64) {
	//Map the other parameters as well
	var newLPState = &(models.LiquidityProviderState{Address: lpAddress, UnlockedBalance: lpUnlocked, LatestBlock: blockNumber})
	dbc.UpsertLiquidityProviderState(newLPState, blockNumber)
	dbc.UpdateVaultFields(vaultAddress, map[string]interface{}{"unlocked_balance": vaultUnlocked, "latest_block": blockNumber})
}

func (dbc *DB) RoundDeployedIndex(optionRound models.OptionRound) {

	dbc.CreateOptionRound(&optionRound)
}

func (dbc *DB) AuctionStartedIndex(roundAddress string, blockNumber uint64, availableOptions, startingLiquidity models.BigInt) {
	dbc.UpdateOptionRoundFields(roundAddress, map[string]interface{}{
		"available_options":  availableOptions,
		"starting_liquidity": startingLiquidity,
		"state":              "Auctioning",
	})
	dbc.UpdateAllLiquidityProvidersBalancesAuctionStart(blockNumber)
	dbc.UpdateVaultBalancesAuctionStart(blockNumber)
}

func (dbc *DB) AuctionEndedIndex(prevStateOptionRound models.OptionRound, roundAddress string, blockNumber uint64, optionsSold, clearingPrice, clearingNonce, premiums models.BigInt) {
	unsoldLiquidity := models.BigInt{Int: new(big.Int).Sub(
		prevStateOptionRound.StartingLiquidity.Int,
		new(big.Int).Div(
			new(big.Int).Mul(
				prevStateOptionRound.StartingLiquidity.Int,
				prevStateOptionRound.SoldOptions.Int,
			),
			prevStateOptionRound.AvailableOptions.Int,
		),
	)}
	dbc.UpdateAllLiquidityProvidersBalancesAuctionEnd(prevStateOptionRound.StartingLiquidity, unsoldLiquidity, premiums, blockNumber)
	dbc.UpdateVaultBalancesAuctionEnd(unsoldLiquidity, premiums, blockNumber)
	dbc.UpdateBiddersAuctionEnd(roundAddress, clearingPrice, optionsSold, clearingNonce)
	dbc.UpdateOptionRoundAuctionEnd(roundAddress, clearingPrice, optionsSold)
}

func (dbc *DB) RoundSettledIndex(prevStateOptionRound models.OptionRound, roundAddress string, blockNumber uint64, settlementPrice, totalPayout models.BigInt) {
	dbc.UpdateVaultBalancesOptionSettle(prevStateOptionRound.StartingLiquidity, prevStateOptionRound.QueuedLiquidity, blockNumber)
	dbc.UpdateAllLiquidityProvidersBalancesOptionSettle(roundAddress, prevStateOptionRound.StartingLiquidity, prevStateOptionRound.QueuedLiquidity, totalPayout, models.BigInt{Int: new(big.Int).SetUint64(blockNumber)})
	dbc.UpdateOptionRoundFields(prevStateOptionRound.Address, map[string]interface{}{
		"settlement_price": settlementPrice,
		"total_payout":     totalPayout,
		"state":            "Settled",
	})
}

func (dbc *DB) BidAcceptedIndex(bid models.Bid) {
	dbc.CreateBid(&bid)
}

func (dbc *DB) BidUpdatedIndex(bidId string, amount models.BigInt, treeNonce uint64) {
	dbc.tx.Model(models.Bid{}).Where("bid_id = ?", bidId).Updates(map[string]interface{}{
		"amount":     gorm.Expr("amount + ?", amount),
		"tree_nonce": treeNonce,
	})
}
