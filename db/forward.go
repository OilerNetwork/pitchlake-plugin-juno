package db

import (
	"junoplugin/models"
	"log"
	"math/big"

	"gorm.io/gorm"
)

func (db *DB) DepositOrWithdrawIndex(
	vaultAddress,
	lpAddress string,
	lpUnlocked, vaultUnlocked models.BigInt,
	blockNumber uint64) error {
	//Map the other parameters as well
	var newLPState = &(models.LiquidityProviderState{
		Address:         lpAddress,
		UnlockedBalance: lpUnlocked,
		LatestBlock:     blockNumber,
	})
	if err := db.UpsertLiquidityProviderState(newLPState, blockNumber); err != nil {
		return err
	}
	if err := db.UpdateVaultFields(vaultAddress, map[string]interface{}{
		"unlocked_balance": vaultUnlocked,
		"latest_block":     blockNumber,
	}); err != nil {
		return err
	}
	return nil
}

func (db *DB) WithdrawalQueuedIndex(
	lpAddress, vaultAddress string,
	roundId uint64,
	bps, accountQueuedBefore, accountQueuedNow, vaultQueuedNow models.BigInt,
) error {
	vault, err := db.GetVaultByAddress(vaultAddress)
	if err != nil {
		return err
	}
	queuedLiquidity := models.QueuedLiquidity{
		Address:      lpAddress,
		RoundAddress: vault.CurrentRoundAddress,
		Bps:          bps,
		QueuedAmount: accountQueuedNow,
	}
	if err := db.UpsertQueuedLiquidity(&queuedLiquidity); err != nil {
		return err
	}
	if err := db.UpdateOptionRoundFields(vault.CurrentRoundAddress, map[string]interface{}{
		"queued_liquidity": vaultQueuedNow,
	}); err != nil {
		return err
	}
	return nil
}

func (db *DB) StashWithdrawnIndex(
	vaultAddress, lpAddress string,
	amount, vaultBalanceNow models.BigInt,
	blockNumber uint64) error {
	if err := db.UpdateLiquidityProviderFields(lpAddress, map[string]interface{}{
		"stashed_balance": 0,
		"latest_block":    blockNumber,
	}); err != nil {

		return err
	}
	if err := db.UpdateVaultFields(vaultAddress, map[string]interface{}{
		"stashed_balance": vaultBalanceNow,
		"latest_block":    blockNumber,
	}); err != nil {
		return err
	}
	return nil
}

func (db *DB) RoundDeployedIndex(optionRound models.OptionRound) error {

	if err := db.CreateOptionRound(&optionRound); err != nil {
		return err
	}
	if err := db.UpdateVaultFields(optionRound.VaultAddress, map[string]interface{}{
		"current_round":         optionRound.RoundID,
		"current_round_address": optionRound.Address,
	}); err != nil {
		return err
	}
	return nil
}

func (dbc *DB) PricingDataSetIndex(
	roundAddress string,
	strikePrice, capLevel, reservePrice models.BigInt) error {
	err := dbc.UpdateOptionRoundFields(roundAddress, map[string]interface{}{
		"strike_price":  strikePrice,
		"cap_level":     capLevel,
		"reserve_price": reservePrice,
	})
	if err != nil {
		return err
	}
	return nil

}
func (dbc *DB) AuctionStartedIndex(
	vaultAddress, roundAddress string,
	blockNumber uint64,
	availableOptions, startingLiquidity models.BigInt) {
	dbc.UpdateOptionRoundFields(roundAddress, map[string]interface{}{
		"available_options":  availableOptions,
		"starting_liquidity": startingLiquidity,
		"state":              "Auctioning",
	})
	dbc.UpdateAllLiquidityProvidersBalancesAuctionStart(blockNumber)
	dbc.UpdateVaultBalanceAuctionStart(vaultAddress, blockNumber)
}

func (dbc *DB) AuctionEndedIndex(
	prevStateOptionRound models.OptionRound,
	roundAddress string,
	blockNumber, clearingNonce uint64,
	optionsSold, clearingPrice, premiums, unsoldData models.BigInt) {
	unsoldLiquidity := models.BigInt{Int: new(big.Int).Sub(
		prevStateOptionRound.StartingLiquidity.Int,
		new(big.Int).Div(
			new(big.Int).Mul(
				prevStateOptionRound.StartingLiquidity.Int,
				optionsSold.Int,
			),
			prevStateOptionRound.AvailableOptions.Int,
		),
	)}
	log.Printf("DATA %v %v %v ", unsoldLiquidity, unsoldData, optionsSold)
	dbc.UpdateAllLiquidityProvidersBalancesAuctionEnd(prevStateOptionRound.VaultAddress, prevStateOptionRound.StartingLiquidity, unsoldLiquidity, premiums, blockNumber)
	dbc.UpdateVaultBalancesAuctionEnd(prevStateOptionRound.VaultAddress, unsoldLiquidity, premiums, blockNumber)
	dbc.UpdateBiddersAuctionEnd(roundAddress, clearingPrice, optionsSold, clearingNonce)
	dbc.UpdateOptionRoundAuctionEnd(roundAddress, clearingPrice, optionsSold)
}

func (dbc *DB) RoundSettledIndex(prevStateOptionRound models.OptionRound, roundAddress string, blockNumber uint64, settlementPrice, optionsSold, payoutPerOption models.BigInt) {
	dbc.UpdateVaultBalancesOptionSettle(prevStateOptionRound.VaultAddress, prevStateOptionRound.StartingLiquidity, prevStateOptionRound.QueuedLiquidity, blockNumber)
	dbc.UpdateAllLiquidityProvidersBalancesOptionSettle(roundAddress, prevStateOptionRound.StartingLiquidity, prevStateOptionRound.QueuedLiquidity, payoutPerOption, optionsSold, models.BigInt{Int: new(big.Int).SetUint64(blockNumber)})
	dbc.UpdateOptionRoundFields(prevStateOptionRound.Address, map[string]interface{}{
		"settlement_price":  settlementPrice,
		"payout_per_option": payoutPerOption,
		"state":             "Settled",
	})
}

func (db *DB) BidPlacedIndex(bid models.Bid, buyer models.OptionBuyer) {
	log.Printf("NEW BID %v", bid)
	db.CreateOptionBuyer(&buyer)
	db.CreateBid(&bid)
}

func (dbc *DB) BidUpdatedIndex(bidId string, amount models.BigInt, treeNonce uint64) {
	dbc.tx.Model(models.Bid{}).Where("bid_id = ?", bidId).Updates(map[string]interface{}{
		"amount":     gorm.Expr("amount + ?", amount),
		"tree_nonce": treeNonce,
	})
}
