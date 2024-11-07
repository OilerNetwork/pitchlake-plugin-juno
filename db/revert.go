package db

import (
	"junoplugin/models"

	"gorm.io/gorm"
)

func (db *DB) DepositOrWithdrawRevert(vaultAddress, lpAddress string, blockNumber uint64) {
	//Map the other parameters as well

	db.RevertVaultState(vaultAddress, blockNumber)
	db.RevertLPState(lpAddress, blockNumber)
}

func (db *DB) RoundDeployedRevert(roundAddress string) {

	db.DeleteOptionRound(roundAddress)
}

func (db *DB) AuctionStartedRevert(vaultAddress, roundAddress string, blockNumber uint64) {
	db.RevertVaultState(vaultAddress, blockNumber)
	db.RevertAllLPState(blockNumber)
	db.UpdateOptionRoundFields(roundAddress, map[string]interface{}{
		"available_options":  0,
		"starting_liquidity": 0,
		"state":              "Open",
	})
}

func (db *DB) AuctionEndedRevert(vaultAddress, roundAddress string, blockNumber uint64) {
	db.RevertVaultState(vaultAddress, blockNumber)
	db.RevertAllLPState(blockNumber)
	db.UpdateOptionRoundFields(roundAddress, map[string]interface{}{
		"clearing_price": nil,
		"options_sold":   nil,
		"state":          "Auctioning",
	})
	db.UpdateAllOptionBuyerFields(roundAddress, map[string]interface{}{
		"tokenizable_options": 0,
		"refundable_options":  0,
	})
}

func (db *DB) RoundSettledRevert(vaultAddress, roundAddress string, blockNumber uint64) {
	db.RevertVaultState(vaultAddress, blockNumber)
	db.RevertAllLPState(blockNumber)
	db.UpdateOptionRoundFields(roundAddress, map[string]interface{}{
		"settlement_price": 0,
		"total_payout":     0,
		"state":            "Running",
	})
}

func (db *DB) BidAcceptedRevert(bidId, roundAddress string) {
	db.DeleteBid(bidId, roundAddress)
}

func (db *DB) BidUpdatedRevert(bidId string, amount models.BigInt, treeNonce uint64) {
	db.tx.Model(models.Bid{}).Where("bid_id = ?", bidId).Updates(map[string]interface{}{
		"amount":     gorm.Expr("amount - ?", amount),
		"tree_nonce": treeNonce,
	})
}
