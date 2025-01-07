package db

import (
	"junoplugin/models"
	"log"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupInMemoryDB(t *testing.T) *DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite: %v", err)
	}

	// Auto-migrate the VaultState model
	err = db.AutoMigrate(&models.VaultState{})
	if err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	return &DB{Conn: db, tx: db}
}
func TestUpdateVaultBalanceAuctionStart(t *testing.T) {
	// Setup in-memory SQLite DB
	testDB := setupInMemoryDB(t)

	// Create a mock VaultState for testing
	vault := models.VaultState{
		Address:         "0x1234",
		UnlockedBalance: models.BigInt{Int: big.NewInt(1000)},
		LockedBalance:   models.BigInt{Int: big.NewInt(0)},
		LatestBlock:     0,
	}
	err := testDB.Conn.Create(&vault).Error
	require.NoError(t, err)

	// Test input data
	vaultAddress := "0x1234"
	blockNumber := uint64(1000)

	// Call the method
	log.Printf("HERE")
	err = testDB.UpdateVaultBalanceAuctionStart(vaultAddress, blockNumber)
	require.NoError(t, err)

	// Retrieve the updated VaultState and verify the update
	var updatedVault models.VaultState
	err = testDB.Conn.Where("address = ?", vaultAddress).First(&updatedVault).Error
	require.NoError(t, err)

	// Assert that the balances are reset to zero
	require.Equal(t, uint64(1000), updatedVault.LatestBlock)
	require.Equal(t, big.NewInt(1000), updatedVault.LockedBalance.Int)
	require.Equal(t, big.NewInt(0), updatedVault.UnlockedBalance.Int)
}

func TestUpdateVaultBalancesAuctionEnd(t *testing.T) {
	db := setupInMemoryDB(t)

	// Seed initial data
	initialVault := models.VaultState{
		Address:         "0x123",
		UnlockedBalance: models.BigInt{Int: big.NewInt(1000)},
		LockedBalance:   models.BigInt{Int: big.NewInt(500)},
		LatestBlock:     900,
	}
	err := db.Conn.Create(&initialVault).Error
	require.NoError(t, err)

	// Call the function to update the vault balances
	err = db.UpdateVaultBalancesAuctionEnd("0x123", models.BigInt{Int: big.NewInt(100)}, models.BigInt{Int: big.NewInt(50)}, 1000)
	require.NoError(t, err)

	// Verify the updates
	var updatedVault models.VaultState
	err = db.Conn.First(&updatedVault, "address = ?", "0x123").Error
	require.NoError(t, err)

	require.Equal(t, models.BigInt{Int: big.NewInt(1150)}, updatedVault.UnlockedBalance, "unlocked balance should be updated")
	require.Equal(t, models.BigInt{Int: big.NewInt(400)}, updatedVault.LockedBalance, "locked balance should be updated")
	require.Equal(t, uint64(1000), updatedVault.LatestBlock, "latest block should be updated")
}
