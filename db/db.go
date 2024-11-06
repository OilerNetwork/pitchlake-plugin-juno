package db

import (
	"errors"
	"junoplugin/models"
	"log"
	"math/big"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	Conn *gorm.DB
	tx   *gorm.DB
}

func (db *DB) CreateVault(vault *models.VaultState) error {
	if err := db.tx.Create(vault).Error; err != nil {
		return err
	}
	return nil
}

func Init(dsn string) (*DB, error) {
	log.Printf("connecting to %s", dsn)
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{SkipDefaultTransaction: true})

	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
		return nil, err
	}

	m, err := migrate.New(
		"file://db/migrations",
		dsn)
	if err != nil {
		log.Printf("FAIlED HERE 1")
		log.Fatal(err)
	}
	if err := m.Up(); err != nil {
		if err != migrate.ErrNoChange {
			log.Fatal(err)
		}

	}
	m.Close()
	// Automatically migrate your schema
	// err = conn.AutoMigrate(
	// 	&models.Vault{},
	// 	&models.LiquidityProvider{},
	// 	&models.OptionBuyer{},
	// 	&models.OptionRound{},
	// 	&models.VaultState{},
	// 	&models.Bid{},
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to migrate database schema: %v", err)
	// 	return nil, err
	// }

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

func (db *DB) UpdateAllLiquidityProvidersBalancesAuctionStart(blockNumber uint64) error {
	return db.tx.Model(models.LiquidityProviderState{}).Where("unlocked_balance > 0").Updates(
		map[string]interface{}{
			"locked_balance":   gorm.Expr("unlocked_balance"),
			"unlocked_balance": 0,
			"latest_block":     blockNumber,
		}).Error
}

func (db *DB) UpdateVaultBalanceAuctionStart(vaultAddress string, blockNumber uint64) error {
	return db.tx.Model(models.VaultState{}).Where("address=?", vaultAddress).Updates(
		map[string]interface{}{
			"unlocked_balance": 0,
			"locked_balance":   gorm.Expr("unlocked_balance"),
			"latest_block":     blockNumber,
		}).Error
}

func (db *DB) UpdateAllLiquidityProvidersBalancesAuctionEnd(
	startingLiquidity,
	unsoldLiquidity,
	premiums models.BigInt,
	blockNumber uint64) error {

	return db.tx.Model(models.LiquidityProviderState{}).Where("1=1").Updates(
		map[string]interface{}{
			"locked_balance":   gorm.Expr("locked_balance-FLOOR((locked_balance*?)/?)", unsoldLiquidity, startingLiquidity),
			"unlocked_balance": gorm.Expr("unlocked_balance+FLOOR((locked_balance*?))/?+FLOOR((?*locked_balance)/?)", unsoldLiquidity, startingLiquidity, premiums, startingLiquidity),
			"latest_block":     blockNumber,
		}).Error
}

func (db *DB) UpdateVaultBalancesAuctionEnd(
	vaultAddress string,
	unsoldLiquidity,
	premiums models.BigInt,
	blockNumber uint64) error {

	return db.tx.Model(models.VaultState{}).Where("address=?", vaultAddress).Updates(
		map[string]interface{}{
			"unlocked_balance": gorm.Expr("unlocked_balance+?+?", unsoldLiquidity, premiums),
			"locked_balance":   gorm.Expr("locked_balance-?", unsoldLiquidity),
			"latest_block":     blockNumber,
		}).Error

}

func (db *DB) UpdateOptionRoundAuctionEnd(
	address string,
	clearingPrice,
	optionsSold models.BigInt) error {
	err := db.UpdateOptionRoundFields(
		address,
		map[string]interface{}{
			"clearing_price": clearingPrice,
			"sold_options":   optionsSold,
			"state":          "Running",
		})
	if err != nil {
		return err
	}
	return nil
}
func (db *DB) UpdateBiddersAuctionEnd(
	roundAddress string,
	clearingPrice,
	clearingOptionsSold models.BigInt,
	clearingNonce uint64) error {
	bidsAbove, err := db.GetBidsAboveClearingForRound(roundAddress, clearingPrice, clearingNonce)
	if err != nil {
		return err
	}

	for _, bid := range bidsAbove {
		if true {
			refundableAmount := &models.BigInt{Int: new(big.Int).Mul(new(big.Int).Sub(bid.Amount.Int, clearingOptionsSold.Int), clearingPrice.Int)}
			err := db.UpdateOptionBuyerFields(bid.BuyerAddress, roundAddress, map[string]interface{}{
				"refundable_amount": gorm.Expr("refundable_amount+?", refundableAmount),
				"mintable_options":  gorm.Expr("mintable_options+?", clearingOptionsSold),
			})
			if err != nil {
				return err
			}
			return nil
		} else {
			err := db.UpdateOptionBuyerFields(bid.BuyerAddress, roundAddress, map[string]interface{}{
				"mintable_options": gorm.Expr("mintable_options+?", bid.Amount),
			})
			if err != nil {
				return err
			}
			return nil

		}
	}
	bidsBelow, err := db.GetBidsBelowClearingForRound(roundAddress, clearingPrice, clearingNonce)
	if err != nil {
		return err
	}
	for _, bid := range bidsBelow {
		err := db.UpdateOptionBuyerFields(bid.BuyerAddress, roundAddress, map[string]interface{}{
			"refundable_amount": gorm.Expr("mintable_options+?", bid.Amount),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) UpdateVaultBalancesOptionSettle(remainingLiquidty, remainingLiquidityStashed models.BigInt, blockNumber uint64) error {
	difference := models.BigInt{Int: new(big.Int).Sub(remainingLiquidty.Int, remainingLiquidityStashed.Int)}
	return db.tx.Model(models.VaultState{}).Updates(map[string]interface{}{

		"stashed_balance":  gorm.Expr("stashed_balance+ ? ", remainingLiquidityStashed),
		"unlocked_balance": gorm.Expr("unlocked_balance+?", difference),
		"locked_balance":   0,
		"latest_block":     blockNumber,
	}).Error

}
func (db *DB) UpdateAllLiquidityProvidersBalancesOptionSettle(roundAddress string, startingLiquidity, remainingLiquidty, totalPayout, blockNumber models.BigInt) error {

	db.tx.Model(models.LiquidityProviderState{}).Updates(map[string]interface{}{
		"locked_balance":   0,
		"unlocked_balance": gorm.Expr("unlocked_balance + locked_balance - locked_balance*?/?", totalPayout, startingLiquidity),
		"latest_block":     blockNumber,
	})
	queuedAmounts, err := db.GetAllQueuedLiquidityForRound(roundAddress)
	if err != nil {
		return err
	}
	for _, queuedAmount := range queuedAmounts {

		amountToAdd := &models.BigInt{Int: new(big.Int).Div(new(big.Int).Mul(remainingLiquidty.Int, queuedAmount.QueuedAmount.Int), startingLiquidity.Int)}
		db.tx.Model(models.LiquidityProviderState{}).Where("address = ? AND round_address = ", queuedAmount.Address, roundAddress).
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
func (db *DB) GetVaultByAddress(address string) (*models.VaultState, error) {
	var vault models.VaultState
	if err := db.tx.Where("address = ?", address).First(&vault).Error; err != nil {
		return nil, err
	}
	return &vault, nil
}

func (db *DB) UpsertLiquidityProviderState(lp *models.LiquidityProviderState, blockNumber uint64) error {
	// Log the input for debugging
	// Log the input for debugging
	log.Printf("Upserting LP: %+v, Block Number: %d", lp, blockNumber)

	// Perform upsert using GORM's Clauses with the transaction object
	err := db.tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "address"}},
		DoUpdates: clause.AssignmentColumns([]string{"unlocked_balance", "locked_balance", "latest_block"}),
	}).Create(&models.LiquidityProviderState{
		Address:         lp.Address,
		UnlockedBalance: lp.UnlockedBalance,
		LockedBalance:   lp.LockedBalance,
		LatestBlock:     blockNumber,
	}).Error

	if err != nil {
		log.Printf("Upsert error: %v", err)
		return err
	}

	return nil

}

func (db *DB) UpdateOptionBuyerFields(address string, roundAddress string, updates map[string]interface{}) error {
	return db.tx.Model(models.OptionBuyer{}).Where("address = ? AND round_address = ?", address, roundAddress).Updates(updates).Error
}

func (db *DB) UpdateAllOptionBuyerFields(roundAddress string, updates map[string]interface{}) error {
	return db.tx.Model(models.OptionRound{}).Where("round_address=?", roundAddress).Updates(updates).Error
}

func (db *DB) GetOptionRoundByAddress(address string) (*models.OptionRound, error) {
	var or models.OptionRound
	if err := db.tx.First(&or).Where("address = ?", address).Error; err != nil {
		return nil, err
	}
	return &or, nil
}

func (db *DB) UpdateOptionRoundFields(address string, updates map[string]interface{}) error {
	return db.tx.Model(models.OptionRound{}).Where("address = ?", address).Updates(updates).Error
}

func (db *DB) UpdateVaultFields(address string, updates map[string]interface{}) error {
	return db.tx.Model(models.VaultState{}).Where("address = ?", address).Updates(updates).Error
}
func (db *DB) UpdateLiquidityProviderFields(address string, updates map[string]interface{}) error {
	return db.tx.Model(models.LiquidityProviderState{}).Where("address = ?", address).Updates(updates).Error
}

// DeleteOptionRound deletes an OptionRound record by its ID
func (db *DB) DeleteOptionRound(roundAddress string) error {
	if err := db.tx.Where("address = ?", roundAddress).Delete(&models.OptionRound{}).Error; err != nil {
		return err
	}
	return nil
}

// CreateBid creates a new Bid record in the database
func (db *DB) CreateBid(bid *models.Bid) error {
	if err := db.tx.Create(bid).Error; err != nil {
		return err
	}
	return nil
}
func (db *DB) CreateOptionRound(round *models.OptionRound) error {
	if err := db.tx.Create(round).Error; err != nil {
		return err
	}
	return nil
}

// DeleteBid deletes a Bid record by its ID
func (db *DB) DeleteBid(bidID string, roundAddress string) error {
	if err := db.tx.Model(&models.Bid{}).Where("round_address=? AND bid_id=?", roundAddress, bidID).Error; err != nil {
		return err
	}
	return nil
}
func (db *DB) GetBidsForRound(roundAddress string) ([]models.Bid, error) {
	var bids []models.Bid
	if err := db.Conn.Where("round_address = ?", roundAddress).Order("price DESC").
		Order("tree_nonce ASC").Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

func (db *DB) GetBidsAboveClearingForRound(roundAddress string, clearingPrice models.BigInt, clearingNonce uint64) ([]models.Bid, error) {
	var bids []models.Bid
	if err := db.Conn.Where("round_address = ?", roundAddress).
		Where("price > ? OR (price = ? AND tree_nonce >= ?)", clearingPrice, clearingPrice, clearingNonce).
		Order("price DESC").
		Order("tree_nonce ASC").
		Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

func (db *DB) GetBidsBelowClearingForRound(roundAddress string, clearingPrice models.BigInt, clearingNonce uint64) ([]models.Bid, error) {
	var bids []models.Bid
	if err := db.Conn.Where("round_address = ?", roundAddress).
		Where("NOT(price > ? OR (price = ? AND tree_nonce >= ?))", clearingPrice, clearingPrice, clearingNonce).
		Order("price DESC").
		Order("tree_nonce ASC").
		Find(&bids).Error; err != nil {
		return nil, err
	}
	return bids, nil
}

func (db *DB) GetAllQueuedLiquidityForRound(roundAddress string) ([]models.QueuedLiquidity, error) {

	var queuedAmounts []models.QueuedLiquidity
	if err := db.Conn.Where("roundAddress=?", roundAddress).Find(&queuedAmounts).Error; err != nil {
		return nil, err
	}
	return queuedAmounts, nil
}

// Revert Functions
func (db *DB) RevertVaultState(address string, blockNumber uint64) error {
	var vaultState models.VaultState
	var vaultHistoric, postRevert models.Vault
	if err := db.tx.Where("address = ? AND last_block = ?", address, blockNumber).First(&vaultState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	if err := db.tx.Where("address = ? AND block_number = ?", address, blockNumber).First(&vaultHistoric).Error; err != nil {
		return err
	}

	if err := db.tx.Delete(&vaultHistoric).Error; err != nil {
		return err
	}

	if err := db.tx.Where("address = ?", address).
		Order("latest_block DESC").
		First(&postRevert).Error; err != nil {
		return nil
	}

	if err := db.tx.Where("address = ?").Updates(map[string]interface{}{
		"unlocked_balance": postRevert.UnlockedBalance,
		"locked_balance":   postRevert.LockedBalance,
		"stashed_balance":  postRevert.StashedBalance,
		"latest_block":     postRevert.BlockNumber,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) RevertAllLPState(blockNumber uint64) error {
	var lpStates []models.LiquidityProviderState
	var lpHistoric, postRevert models.LiquidityProvider
	if err := db.tx.Model(models.LiquidityProviderState{}).Where("last_block = ?", blockNumber).Find(&lpStates).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	for _, lpState := range lpStates {
		if err := db.tx.Model(models.LiquidityProvider{}).Where("address = ? AND block_number = ?", lpState.Address, blockNumber).First(&lpHistoric).Error; err != nil {
			return err
		}

		if err := db.tx.Delete(&lpHistoric).Error; err != nil {
			return err
		}

		if err := db.tx.Where("address = ?", lpState.Address).
			Order("latest_block DESC").
			First(&postRevert).Error; err != nil {
			return nil
		}

		if err := db.tx.Where("address = ?").Updates(map[string]interface{}{
			"unlocked_balance": postRevert.UnlockedBalance,
			"locked_balance":   postRevert.LockedBalance,
			"stashed_balance":  postRevert.StashedBalance,
			"latest_block":     postRevert.BlockNumber,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (db *DB) RevertLPState(address string, blockNumber uint64) error {
	var lpState models.LiquidityProviderState
	var lpHistoric, postRevert models.LiquidityProvider
	if err := db.tx.Where("address = ? AND last_block = ?", address, blockNumber).First(&lpState).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		} else {
			return err
		}
	}

	if err := db.tx.Model(models.LiquidityProvider{}).Where("address = ? AND block_number = ?", address, blockNumber).First(&lpHistoric).Error; err != nil {
		return err
	}

	if err := db.tx.Delete(&lpHistoric).Error; err != nil {
		return err
	}

	if err := db.tx.Model(models.LiquidityProvider{}).Where("address = ?", address).
		Order("latest_block DESC").
		First(&postRevert).Error; err != nil {
		return nil
	}

	if err := db.tx.Model(models.LiquidityProviderState{}).Where("address = ?").Updates(map[string]interface{}{
		"unlocked_balance": postRevert.UnlockedBalance,
		"locked_balance":   postRevert.LockedBalance,
		"stashed_balance":  postRevert.StashedBalance,
		"latest_block":     postRevert.BlockNumber,
	}).Error; err != nil {
		return err
	}

	return nil
}

func (db *DB) Begin() {
	tx := db.Conn.Begin()
	db.tx = tx
}

func (db *DB) Commit() {
	db.tx.Commit()
	db.tx = nil
}

func (db *DB) Tx(tx *gorm.DB) {
	db.tx = tx
}
