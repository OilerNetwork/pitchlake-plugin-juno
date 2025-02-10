package db

import (
	"junoplugin/models"
	"math/big"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return &DB{Conn: gormDB, tx: gormDB}, mock
}

func TestUpdateVaultBalanceAuctionStart(t *testing.T) {
	testDB, mock := setupMockDB(t)
	// Mock the rows that will be updated

	// Mock the expected database interactions
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "VaultStates" SET "latest_block"=\$1,"locked_balance"=unlocked_balance,"unlocked_balance"=\$2 WHERE address=\$3`).
		WithArgs(1000, 0, "0x1234").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Call the method
	err := testDB.UpdateVaultBalanceAuctionStart("0x1234", 1000)
	require.NoError(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
func TestUpdateVaultBalancesAuctionEnd(t *testing.T) {
	testDB, mock := setupMockDB(t)

	// Mock the expected database interactions
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "VaultStates" SET "latest_block"=\$1,"locked_balance"=locked_balance-\$2,"unlocked_balance"=unlocked_balance\+\$3\+\$4 WHERE address=\$5`).
		WithArgs(1000, "100", "100", "50", "0x123").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Call the method
	err := testDB.UpdateVaultBalancesAuctionEnd(
		"0x123",                             // startingLiquidity
		models.BigInt{Int: big.NewInt(100)}, // unsoldLiquidity
		models.BigInt{Int: big.NewInt(50)},  // premiums
		1000,                                // blockNumber
	)
	require.NoError(t, err)

	// Ensure all expectations were met
	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
