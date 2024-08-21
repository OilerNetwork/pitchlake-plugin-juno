package db

import (
	"database/sql"
	"junoplugin/models"
	"log"
)

type PostgresDB struct {
	conn *sql.DB
}

// NewPostgresDB initializes a new database connection using a connection string and returns a PostgresDB instance
func NewPostgresDB(connString string) (*PostgresDB, error) {
	// Open a connection to the PostgreSQL database
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return nil, err
	}

	// Verify connection by pinging the database
	if err := db.Ping(); err != nil {
		log.Printf("Error verifying connection to the database: %v", err)
		return nil, err
	}

	log.Println("Successfully connected to the database")

	// Return the PostgresDB instance
	return &PostgresDB{conn: db}, nil
}

// Close closes the database connection
func (db *PostgresDB) Close() error {
	return db.conn.Close()
}

// InsertVault inserts a new record into the vault table
func (db *PostgresDB) InsertVault(vault *models.Vault) error {
	query := `INSERT INTO vault (block_number, unlocked_balance, locked_balance) VALUES ($1, $2, $3)`
	_, err := db.conn.Exec(query, vault.BlockNumber, vault.UnlockedBalance, vault.LockedBalance)
	if err != nil {
		log.Printf("Failed to insert vault: %v", err)
		return err
	}
	return nil
}

// GetVault retrieves a record from the vault table by block_number
func (db *PostgresDB) GetVault(blockNumber int) (*models.Vault, error) {
	var vault models.Vault
	query := `SELECT block_number, unlocked_balance, locked_balance FROM vault WHERE block_number = $1`
	err := db.conn.QueryRow(query, blockNumber).Scan(&vault.BlockNumber, &vault.UnlockedBalance, &vault.LockedBalance)
	if err != nil {
		log.Printf("Failed to get vault: %v", err)
		return nil, err
	}
	return &vault, nil
}
