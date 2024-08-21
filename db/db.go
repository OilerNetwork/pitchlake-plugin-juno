package db

import (
	"errors"
	"junoplugin/models"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	conn *gorm.DB
}

func Init(dsn string) (*DB, error) {
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

	return &DB{conn: conn}, nil
}

func (db *DB) Close() error {
	//Close the DB connection
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateVault creates a new Vault record in the database
func (db *DB) CreateVault(vault *models.Vault) error {
	if err := db.conn.Create(vault).Error; err != nil {
		return err
	}
	return nil
}

// GetVault retrieves a Vault record by its ID
func (db *DB) GetVault(id uint) (*models.Vault, error) {
	var vault models.Vault
	if err := db.conn.First(&vault, id).Error; err != nil {
		return nil, err
	}
	return &vault, nil
}

// UpdateVault updates an existing Vault record
func (db *DB) UpdateVault(vault *models.Vault) error {
	if err := db.conn.Save(vault).Error; err != nil {
		return err
	}
	return nil
}

// DeleteVault deletes a Vault record by its ID
func (db *DB) DeleteVault(id uint) error {
	if err := db.conn.Delete(&models.Vault{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateLiquidityProvider creates a new LiquidityProvider record in the database
func (db *DB) CreateLiquidityProvider(lp *models.LiquidityProvider) error {
	if err := db.conn.Create(lp).Error; err != nil {
		return err
	}
	return nil
}

// GetLiquidityProvider retrieves a LiquidityProvider record by its ID
func (db *DB) GetLiquidityProvider(id uint) (*models.LiquidityProvider, error) {
	var lp models.LiquidityProvider
	if err := db.conn.First(&lp, id).Error; err != nil {
		return nil, err
	}
	return &lp, nil
}

// UpdateLiquidityProvider updates an existing LiquidityProvider record
func (db *DB) UpdateLiquidityProvider(lp *models.LiquidityProvider) error {
	if err := db.conn.Save(lp).Error; err != nil {
		return err
	}
	return nil
}

func (db *DB) UpsertLiquidityProviderState(lp *models.LiquidityProviderState) error {
	// Attempt to update the record based on the composite key (address and block_number)
	if err := db.conn.Model(&models.LiquidityProvider{}).
		Where("address = ?", lp.Address).
		Updates(map[string]interface{}{
			"unlocked_balance": lp.UnlockedBalance,
			"locked_balance":   lp.LockedBalance,
		}).Error; err != nil {

		// Handle the case where the record was not found
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Record not found, so create a new one
			if createErr := db.conn.Create(lp).Error; createErr != nil {
				return createErr // Handle any errors during the creation process
			}
		} else {
			// Handle other errors (e.g., connection failure)
			return err
		}
	}

	return nil
}

// DeleteLiquidityProvider deletes a LiquidityProvider record by its ID
func (db *DB) DeleteLiquidityProvider(id uint) error {
	if err := db.conn.Delete(&models.LiquidityProvider{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateOptionBuyer creates a new OptionBuyer record in the database
func (db *DB) CreateOptionBuyer(ob *models.OptionBuyer) error {
	if err := db.conn.Create(ob).Error; err != nil {
		return err
	}
	return nil
}

// GetOptionBuyer retrieves an OptionBuyer record by its ID
func (db *DB) GetOptionBuyer(id uint) (*models.OptionBuyer, error) {
	var ob models.OptionBuyer
	if err := db.conn.First(&ob, id).Error; err != nil {
		return nil, err
	}
	return &ob, nil
}

// UpdateOptionBuyer updates an existing OptionBuyer record
func (db *DB) UpdateOptionBuyer(ob *models.OptionBuyer) error {
	if err := db.conn.Save(ob).Error; err != nil {
		return err
	}
	return nil
}

// DeleteOptionBuyer deletes an OptionBuyer record by its ID
func (db *DB) DeleteOptionBuyer(id uint) error {
	if err := db.conn.Delete(&models.OptionBuyer{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateOptionRound creates a new OptionRound record in the database
func (db *DB) CreateOptionRound(or *models.OptionRound) error {
	if err := db.conn.Create(or).Error; err != nil {
		return err
	}
	return nil
}

// GetOptionRound retrieves an OptionRound record by its ID
func (db *DB) GetOptionRound(id uint) (*models.OptionRound, error) {
	var or models.OptionRound
	if err := db.conn.First(&or, id).Error; err != nil {
		return nil, err
	}
	return &or, nil
}

// UpdateOptionRound updates an existing OptionRound record
func (db *DB) UpdateOptionRound(or *models.OptionRound) error {
	if err := db.conn.Save(or).Error; err != nil {
		return err
	}
	return nil
}

// DeleteOptionRound deletes an OptionRound record by its ID
func (db *DB) DeleteOptionRound(id uint) error {
	if err := db.conn.Delete(&models.OptionRound{}, id).Error; err != nil {
		return err
	}
	return nil
}

// CreateBid creates a new Bid record in the database
func (db *DB) CreateBid(bid *models.Bid) error {
	if err := db.conn.Create(bid).Error; err != nil {
		return err
	}
	return nil
}

// GetBid retrieves a Bid record by its ID
func (db *DB) GetBid(id uint) (*models.Bid, error) {
	var bid models.Bid
	if err := db.conn.First(&bid, id).Error; err != nil {
		return nil, err
	}
	return &bid, nil
}

// UpdateBid updates an existing Bid record
func (db *DB) UpdateBid(bid *models.Bid) error {
	if err := db.conn.Save(bid).Error; err != nil {
		return err
	}
	return nil
}

// DeleteBid deletes a Bid record by its ID
func (db *DB) DeleteBid(id uint) error {
	if err := db.conn.Delete(&models.Bid{}, id).Error; err != nil {
		return err
	}
	return nil
}
