package db

import (
	"errors"
	"junoplugin/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Conn *gorm.DB
}

func Init(dsn string) (*DB, error) {
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}

	// Automatically migrate your schema
	err = conn.AutoMigrate(
		&models.Vault{},
		&models.LiquidityProvider{},
		&models.OptionBuyer{},
		&models.OptionRound{},
		&models.VaultState{},
		&models.Bid{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schema: %v", err)
		return nil, err
	}

	return &DB{Conn: conn}, nil
}

func (db *DB) Close() error {
	//Close the DB connection
	sqlDB, err := db.Conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (db *DB) UpdateAllLiquidityProvidersBalancesAuctionStart(tx *gorm.DB, blockNumber uint64) error {
	return tx.Model(models.LiquidityProviderState{}).Updates(map[string]interface{}{
		"locked_balance":   gorm.Expr("unlocked_balance"),
		"unlocked_balance": 0,
		"last_block":       blockNumber,
	}).Error
}

func (db *DB) UpdateVaultBalancesAuctionStart(tx *gorm.DB, blockNumber uint64) error {
	return tx.Model(models.VaultState{}).Updates(map[string]interface{}{
		"unlocked_balance": 0,
		"locked_balance":   gorm.Expr("unlocked_balance"),
		"last_block":       blockNumber,
	}).Error
}

func (db *DB) UpdateAllLiquidityProvidersBalancesAuctionEnd(tx *gorm.DB, startingLiquidity, unsoldLiquidity, premiums, blockNumber uint64) error {

	return tx.Model(models.LiquidityProviderState{}).Updates(map[string]interface{}{
		"locked_balance":   gorm.Expr("locked_balance-FLOOR((locked_balance*?)/?)", unsoldLiquidity, startingLiquidity),
		"unlocked_balance": gorm.Expr("unlocked_balance-FLOOR((locked_balance*?))/?+FLOOR((?*locked_balance)/?)", unsoldLiquidity, startingLiquidity, premiums, startingLiquidity),
		"last_block":       blockNumber,
	}).Error
}

func (db *DB) UpdateVaultBalancesAuctionEnd(tx *gorm.DB, unsoldLiquidity, premiums, blockNumber uint64) error {

	return tx.Model(models.VaultState{}).Updates(map[string]interface{}{
		"unlocked_balance": gorm.Expr("unlocked_balance+?+?", unsoldLiquidity, premiums),
		"locked_balance":   gorm.Expr("locked_balance-?", unsoldLiquidity),
		"last_block":       blockNumber,
	}).Error

}

func (db *DB) UpdateOptionRoundAuctionEnd(tx *gorm.DB, address string, clearingPrice, optionsSold uint64) error {
	err := db.UpdateOptionRoundFields(tx, address, map[string]interface{}{
		"clearing_price": clearingPrice,
		"options_sold":   optionsSold,
		"state":          "Running",
	})
	if err != nil {
		return err
	}
	return nil
}
func (db *DB) UpdateBiddersAuctionEnd(tx *gorm.DB, clearingPrice, clearingOptionsSold, roundID, clearingNonce uint64) error {
	bidsAbove, err := db.GetBidsAboveClearingForRound(tx, roundID, clearingPrice, clearingNonce)
	if err != nil {
		return err
	}

	for _, bid := range bidsAbove {
		if clearingNonce == bid.TreeNonce {

			err := db.UpdateOptionBuyerFields(tx, bid.Address, roundID, map[string]interface{}{
				"refundable_amount": gorm.Expr("refundable_amount+?", (bid.Amount-clearingOptionsSold)*clearingPrice),
				"mintable_options":  gorm.Expr("mintable_options+?", clearingOptionsSold),
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			err := db.UpdateOptionBuyerFields(tx, bid.Address, roundID, map[string]interface{}{
				"mintable_options": gorm.Expr("mintable_options+?", bid.Amount),
			})
			if err != nil {
				return err
			}
			return nil

		}
	}
	bidsBelow, err := db.GetBidsBelowClearingForRound(tx, roundID, clearingPrice, clearingNonce)
	if err != nil {
		return err
	}
	for _, bid := range bidsBelow {
		err := db.UpdateOptionBuyerFields(tx, bid.Address, roundID, map[string]interface{}{
			"refundable_amount": gorm.Expr("mintable_options+?", bid.Amount),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) UpdateVaultBalancesOptionSettle(tx *gorm.DB, remainingLiquidty, remainingLiquidityStashed, blockNumber uint64) error {

	return tx.Model(models.VaultState{}).Updates(map[string]interface{}{
		"stashed_balance":  gorm.Expr("stashed_balance+ ? ", remainingLiquidityStashed),
		"unlocked_balance": gorm.Expr("unlocked_balance+?", remainingLiquidty-remainingLiquidityStashed),
		"locked_balance":   0,
		"last_block":       blockNumber,
	}).Error

}
func (db *DB) UpdateAllLiquidityProvidersBalancesOptionSettle(tx *gorm.DB, roundID, startingLiquidity, remainingLiquidty, totalPayout, blockNumber uint64) error {

	tx.Model(models.LiquidityProviderState{}).Updates(map[string]interface{}{
		"locked_balance":   0,
		"unlocked_balance": gorm.Expr("unlocked_balance + locked_balance - locked_balance*?/?", totalPayout, startingLiquidity),
		"last_block":       blockNumber,
	})
	queuedAmounts, err := db.GetAllQueuedLiquidityForRound(roundID)
	if err != nil {
		return err
	}
	for _, queuedAmount := range queuedAmounts {
		amountToAdd := remainingLiquidty * queuedAmount.QueuedAmount / startingLiquidity
		tx.Model(models.LiquidityProviderState{}).Where("address = ? AND round_id = ", queuedAmount.Address, roundID).
			Updates(map[string]interface{}{
				"stashed_balance":  gorm.Expr("stashed_balance + ?", amountToAdd),
				"unlocked_balance": gorm.Expr("unlocked_balance - ?", amountToAdd),
			})
	}

	/* Use this JOIN query to update this without creating 2 entries on the historic table
	// Perform the update in a single query using JOINs and subqueries
	err := db.Conn.Exec(`
		UPDATE liquidity_provider_states lps
		JOIN (
			SELECT
				address,
				queued_amount,
				remaining_liquidity * queued_amount / ? AS amount_to_add
			FROM queued_liquidity
			WHERE round_id = ?
		) ql ON lps.address = ql.address AND lps.round_id = ?
		SET
			lps.locked_balance = 0,
			lps.unlocked_balance = lps.unlocked_balance + lps.locked_balance - lps.locked_balance * ? / ? - ql.amount_to_add,
			lps.stashed_balance = lps.stashed_balance + ql.amount_to_add
	`, startingLiquidity, roundID, roundID, totalPayout, startingLiquidity).Error
	*/
	return nil
}
func (db *DB) GetVaultByAddress(tx *gorm.DB, address string) (*models.Vault, error) {
	var vault models.Vault
	if err := tx.Where("address = ?", address).First(&vault).Error; err != nil {
		return nil, err
	}
	return &vault, nil
}

func (db *DB) UpsertLiquidityProviderState(tx *gorm.DB, lp *models.LiquidityProviderState, blockNumber uint64) error {
	// Attempt to update the record based on the composite key (address and block_number)
	if err := tx.Model(&models.LiquidityProvider{}).
		Where("address = ?", lp.Address).
		Updates(map[string]interface{}{
			"unlocked_balance": lp.UnlockedBalance,
			"locked_balance":   lp.LockedBalance,
			"last_block":       blockNumber,
		}).Error; err != nil {

		// Handle the case where the record was not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Record not found, so create a new one
			if createErr := tx.Create(lp).Error; createErr != nil {
				return createErr // Handle any errors during the creation process
			}
		} else {
			// Handle other errors (e.g., connection failure)
			return err
		}
	}

	return nil
}

func (db *DB) UpdateOptionBuyerFields(tx *gorm.DB, address string, roundID uint64, updates map[string]interface{}) error {
	return tx.Model(models.OptionRound{}).Where("address = ? AND round_id = ?", address, roundID).Updates(updates).Error
}

func (db *DB) UpdateAllOptionBuyerFields(tx *gorm.DB, roundID uint64, updates map[string]interface{}) error {
	return tx.Model(models.OptionRound{}).Where("round_id=?", roundID).Updates(updates).Error
}

func (db *DB) GetOptionRoundByAddress(tx *gorm.DB, address string) (*models.OptionRound, error) {
	var or models.OptionRound
	if err := tx.First(&or).Where("address = ?", address).Error; err != nil {
		return nil, err
	}
	return &or, nil
}

func (db *DB) UpdateOptionRoundFields(tx *gorm.DB, address string, updates map[string]interface{}) error {
	return tx.Model(models.OptionRound{}).Where("address = ?", address).Updates(updates).Error
}

func (db *DB) UpdateVaultFields(tx *gorm.DB, updates map[string]interface{}) error {
	return tx.Model(models.OptionRound{}).Updates(updates).Error
}
func (db *DB) UpdateLiquidityProviderFields(tx *gorm.DB, address string, updates map[string]interface{}) error {
	return tx.Model(models.LiquidityProviderState{}).Where("address = ?", address).Updates(updates).Error
}

// DeleteOptionRound deletes an OptionRound record by its ID
func (db *DB) DeleteOptionRound(tx *gorm.DB, id uint) error {
	if err := tx.Delete(&models.OptionRound{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateBid creates a new Bid record in the database
func (db *DB) CreateBid(tx *gorm.DB, bid *models.Bid) error {
	if err := tx.Create(bid).Error; err != nil {
		return err
	}
	return nil
}

// DeleteBid deletes a Bid record by its ID
func (db *DB) DeleteBid(tx *gorm.DB, bidID string, roundID uint64) error {
	if err := tx.Model(&models.Bid{}).Where("round_id=? AND bid_id=?", roundID, bidID).Error; err != nil {
		return err
	}
	return nil
}
func (db *DB) GetBidsForRound(tx *gorm.DB, roundID uint64) ([]models.Bid, error) {
	var bids []models.Bid
	if err := tx.Where("round_id = ?", roundID).Order("price DESC").
		Order("tree_nonce ASC").Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

func (db *DB) GetBidsAboveClearingForRound(tx *gorm.DB, roundID uint64, clearingPrice uint64, clearingNonce uint64) ([]models.Bid, error) {
	var bids []models.Bid
	if err := tx.Where("round_id = ?", roundID).
		Where("price > ? OR (price = ? AND tree_nonce >= ?)", clearingPrice, clearingPrice, clearingNonce).
		Order("price DESC").
		Order("tree_nonce ASC").
		Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

func (db *DB) GetBidsBelowClearingForRound(roundID uint64, clearingPrice uint64, clearingNonce uint64) ([]models.Bid, error) {
	var bids []models.Bid
	if err := db.Conn.Where("round_id = ?", roundID).
		Where("NOT(price > ? OR (price = ? AND tree_nonce >= ?))", clearingPrice, clearingPrice, clearingNonce).
		Order("price DESC").
		Order("tree_nonce ASC").
		Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

func (db *DB) GetAllQueuedLiquidityForRound(roundID uint64) ([]models.QueuedLiquidity, error) {

	var queuedAmounts []models.QueuedLiquidity
	if err := db.Conn.Where("round_id=?", roundID).Find(&queuedAmounts).Error; err != nil {
		return nil, err
	}
	return queuedAmounts, nil
}

// Revert Functions
func (db *DB) RevertVaultState(tx *gorm.DB, address string, blockNumber uint64) error {
	var vaultState models.VaultState
	var vaultHistoric, postRevert models.Vault
	if err := tx.Where("address = ? AND last_block = ?", address, blockNumber).First(&vaultState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	if err := tx.Where("address = ? AND block_number = ?", address, blockNumber).First(&vaultHistoric).Error; err != nil {
		return err
	}

	if err := tx.Delete(&vaultHistoric).Error; err != nil {
		return err
	}

	if err := tx.Where("address = ?", address).
		Order("latest_block DESC").
		First(&postRevert).Error; err != nil {
		return nil
	}

	if err := tx.Where("address = ?").Updates(map[string]interface{}{
		"unlocked_balance": postRevert.UnlockedBalance,
		"locked_balance":   postRevert.LockedBalance,
		"stashed_balance":  postRevert.StashedBalance,
		"last_block":       postRevert.BlockNumber,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) RevertAllLPState(tx *gorm.DB, blockNumber uint64) error {
	var lpStates []models.LiquidityProviderState
	var lpHistoric, postRevert models.LiquidityProvider
	if err := tx.Model(models.LiquidityProviderState{}).Where("last_block = ?", blockNumber).Find(&lpStates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	for _, lpState := range lpStates {
		if err := tx.Model(models.LiquidityProvider{}).Where("address = ? AND block_number = ?", lpState.Address, blockNumber).First(&lpHistoric).Error; err != nil {
			return err
		}

		if err := tx.Delete(&lpHistoric).Error; err != nil {
			return err
		}

		if err := tx.Where("address = ?", lpState.Address).
			Order("latest_block DESC").
			First(&postRevert).Error; err != nil {
			return nil
		}

		if err := tx.Where("address = ?").Updates(map[string]interface{}{
			"unlocked_balance": postRevert.UnlockedBalance,
			"locked_balance":   postRevert.LockedBalance,
			"stashed_balance":  postRevert.StashedBalance,
			"last_block":       postRevert.BlockNumber,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) RevertLPState(tx *gorm.DB, address string, blockNumber uint64) error {
	var lpState models.LiquidityProviderState
	var lpHistoric, postRevert models.LiquidityProvider
	if err := tx.Where("address = ? AND last_block = ?", address, blockNumber).First(&lpState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	if err := tx.Model(models.LiquidityProvider{}).Where("address = ? AND block_number = ?", address, blockNumber).First(&lpHistoric).Error; err != nil {
		return err
	}

	if err := tx.Delete(&lpHistoric).Error; err != nil {
		return err
	}

	if err := tx.Model(models.LiquidityProvider{}).Where("address = ?", address).
		Order("latest_block DESC").
		First(&postRevert).Error; err != nil {
		return nil
	}

	if err := tx.Model(models.LiquidityProviderState{}).Where("address = ?").Updates(map[string]interface{}{
		"unlocked_balance": postRevert.UnlockedBalance,
		"locked_balance":   postRevert.LockedBalance,
		"stashed_balance":  postRevert.StashedBalance,
		"last_block":       postRevert.BlockNumber,
	}).Error; err != nil {
		return err
	}

	return nil
}
