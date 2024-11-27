package db

import (
	"junoplugin/models"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpdateVaultBalanceAuctionStart(t *testing.T) {
	// Initialize SQLite in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	// Migrate the models
	err = db.AutoMigrate(&models.VaultState{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	// Initialize DB struct with transaction scope
	testDB := &DB{Conn: db, tx: db}

	// Create a mock VaultState for testing
	vault := models.VaultState{
		Address:         "0x1234",
		UnlockedBalance: models.BigInt{Int: big.NewInt(1000)}, // BigInt initialized
		LockedBalance:   models.BigInt{Int: big.NewInt(500)},  // BigInt initialized
		LatestBlock:     0,
	}
	db.Create(&vault)

	// Test input data
	vaultAddress := "0x1234"
	blockNumber := uint64(1000)

	// Call the method
	err = testDB.UpdateVaultBalanceAuctionStart(vaultAddress, blockNumber)

	// Assert that there was no error
	assert.NoError(t, err)

	// Retrieve the updated VaultState and verify the update
	var updatedVault models.VaultState
	err = db.Where("address = ?", vaultAddress).First(&updatedVault).Error
	assert.NoError(t, err)

	// Compare using BigInt's string representation
	assert.Equal(t, uint64(1000), updatedVault.LatestBlock)
	assert.True(t, updatedVault.LockedBalance.Int.Cmp(updatedVault.UnlockedBalance.Int) == 0) // locked_balance should equal unlocked_balance (both 0)
	assert.True(t, updatedVault.UnlockedBalance.Int.Cmp(big.NewInt(0)) == 0)                  // unlocked_balance should be 0
}
