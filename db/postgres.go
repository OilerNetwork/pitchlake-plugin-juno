package db

import (
	"database/sql"
	"fmt"
	"junoplugin/models"
	"log"

	"github.com/lib/pq"
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

// InsertLiquidityProvider inserts a new record into the liquidity_providers table
func (db *PostgresDB) InsertLiquidityProvider(lp *models.LiquidityProvider) error {
	query := `INSERT INTO liquidity_providers (address, unlocked_balance, locked_balance, block_number) VALUES ($1, $2, $3, $4)`
	_, err := db.conn.Exec(query, lp.Address, lp.UnlockedBalance, lp.LockedBalance, lp.BlockNumber)
	if err != nil {
		log.Printf("Failed to insert liquidity provider: %v", err)
		return fmt.Errorf("insert liquidity provider failed: %w", err)
	}
	return nil
}

// GetLiquidityProvider retrieves a record from the liquidity_providers table by address
func (db *PostgresDB) GetLiquidityProvider(address string) (*models.LiquidityProvider, error) {
	var lp models.LiquidityProvider
	query := `SELECT address, unlocked_balance, locked_balance, block_number FROM liquidity_providers WHERE address = $1`
	err := db.conn.QueryRow(query, address).Scan(&lp.Address, &lp.UnlockedBalance, &lp.LockedBalance, &lp.BlockNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No liquidity provider found with address: %s", address)
			return nil, nil
		}
		log.Printf("Failed to get liquidity provider: %v", err)
		return nil, fmt.Errorf("get liquidity provider failed: %w", err)
	}
	return &lp, nil
}

// InsertOptionBuyer inserts a new record into the option_buyers table
func (db *PostgresDB) InsertOptionBuyer(ob *models.OptionBuyer) error {
	query := `INSERT INTO option_buyers (address, bids, round_id, options_won) VALUES ($1, $2, $3, $4)`
	_, err := db.conn.Exec(query, ob.Address, pq.Array(ob.Bids), ob.RoundID, ob.OptionsWon)
	if err != nil {
		log.Printf("Failed to insert option buyer: %v", err)
		return fmt.Errorf("insert option buyer failed: %w", err)
	}
	return nil
}

// InsertOptionRound inserts a new record into the option_rounds table
func (db *PostgresDB) InsertOptionRound(or *models.OptionRound) error {
	query := `INSERT INTO option_rounds (address, round_id, bids, starting_block, ending_block, settlement_date, 
                starting_liquidity, available_options, settlement_price, strike_price, sold_options, clearing_price, state, premiums) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`
	_, err := db.conn.Exec(query, or.Address, or.RoundID, pq.Array(or.Bids), or.StartingBlock, or.EndingBlock,
		or.SettlementDate, or.StartingLiquidity, or.AvailableOptions, or.SettlementPrice,
		or.StrikePrice, or.SoldOptions, or.ClearingPrice, or.State, or.Premiums)
	if err != nil {
		log.Printf("Failed to insert option round: %v", err)
		return fmt.Errorf("insert option round failed: %w", err)
	}
	return nil
}

// GetOptionRound retrieves a record from the option_rounds table by round_id
func (db *PostgresDB) GetOptionRound(roundID int) (*models.OptionRound, error) {
	var or models.OptionRound
	query := `SELECT address, round_id, bids, starting_block, ending_block, settlement_date, 
                    starting_liquidity, available_options, settlement_price, strike_price, 
                    sold_options, clearing_price, state, premiums 
              FROM option_rounds WHERE round_id = $1`
	err := db.conn.QueryRow(query, roundID).Scan(&or.Address, &or.RoundID, pq.Array(&or.Bids), &or.StartingBlock, &or.EndingBlock,
		&or.SettlementDate, &or.StartingLiquidity, &or.AvailableOptions, &or.SettlementPrice,
		&or.StrikePrice, &or.SoldOptions, &or.ClearingPrice, &or.State, &or.Premiums)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No option round found with round ID: %d", roundID)
			return nil, nil
		}
		log.Printf("Failed to get option round: %v", err)
		return nil, fmt.Errorf("get option round failed: %w", err)
	}
	return &or, nil
}

// InsertVaultState inserts a new record into the vault_state table
func (db *PostgresDB) InsertVaultState(vs *models.VaultState) error {
	query := `INSERT INTO vault_state (current_round, current_round_address, unlocked_balance, locked_balance, address) 
              VALUES ($1, $2, $3, $4, $5)`
	_, err := db.conn.Exec(query, vs.CurrentRound, vs.CurrentRoundAddress, vs.UnlockedBalance, vs.LockedBalance, vs.Address)
	if err != nil {
		log.Printf("Failed to insert vault state: %v", err)
		return fmt.Errorf("insert vault state failed: %w", err)
	}
	return nil
}

// GetVaultState retrieves a record from the vault_state table by address
func (db *PostgresDB) GetVaultState(address string) (*models.VaultState, error) {
	var vs models.VaultState
	query := `SELECT current_round, current_round_address, unlocked_balance, locked_balance, address FROM vault_state WHERE address = $1`
	err := db.conn.QueryRow(query, address).Scan(&vs.CurrentRound, &vs.CurrentRoundAddress, &vs.UnlockedBalance, &vs.LockedBalance, &vs.Address)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No vault state found with address: %s", address)
			return nil, nil
		}
		log.Printf("Failed to get vault state: %v", err)
		return nil, fmt.Errorf("get vault state failed: %w", err)
	}
	return &vs, nil
}

// InsertBid inserts a new record into the bids table
func (db *PostgresDB) InsertBid(bid *models.Bid) error {
	query := `INSERT INTO bids (round_id, bid_id, amount, price) VALUES ($1, $2, $3, $4)`
	_, err := db.conn.Exec(query, bid.RoundID, bid.BidID, bid.Amount, bid.Price)
	if err != nil {
		log.Printf("Failed to insert bid: %v", err)
		return fmt.Errorf("insert bid failed: %w", err)
	}
	return nil
}

// GetBid retrieves a record from the bids table by bid_id
func (db *PostgresDB) GetBid(bidID string) (*models.Bid, error) {
	var bid models.Bid
	query := `SELECT round_id, bid_id, amount, price FROM bids WHERE bid_id = $1`
	err := db.conn.QueryRow(query, bidID).Scan(&bid.RoundID, &bid.BidID, &bid.Amount, &bid.Price)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No bid found with bid ID: %s", bidID)
			return nil, nil
		}
		log.Printf("Failed to get bid: %v", err)
		return nil, fmt.Errorf("get bid failed: %w", err)
	}
	return &bid, nil
}
