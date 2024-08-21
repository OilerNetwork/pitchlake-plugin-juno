package events

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func onDeposit(address string, amount int64, newBlockNumber int64) error {
	// Connect to the database
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// Variables to store previous values
	var prevUnlockedBalanceLP, prevLockedBalanceLP int64
	var prevUnlockedBalanceVault, prevLockedBalanceVault int64

	// Query to select the latest row for the given address in Liquidity_Providers
	querySelectLP := `
		SELECT unlocked_balance, locked_balance
		FROM public."Liquidity_Providers"
		WHERE address = $1
		ORDER BY block_number DESC
		LIMIT 1;
	`
	rowLP := conn.QueryRow(context.Background(), querySelectLP, address)
	err = rowLP.Scan(&prevUnlockedBalanceLP, &prevLockedBalanceLP)

	if err == pgx.ErrNoRows {
		// Address does not exist in Liquidity_Providers, create a new record with initial values
		prevUnlockedBalanceLP = 0
		prevLockedBalanceLP = 0

	} else if err != nil {
		return fmt.Errorf("error fetching previous balance LP: %v", err)
	}

	// Query to select the latest row in Vault
	querySelectVault := `
		SELECT unlocked_balance, locked_balance
		FROM public."Vault"
		ORDER BY block_number DESC
		LIMIT 1;
	`
	rowVault := conn.QueryRow(context.Background(), querySelectVault)
	err = rowVault.Scan(&prevUnlockedBalanceVault, &prevLockedBalanceVault)

	if err == pgx.ErrNoRows {
		// No record exists in Vault, create a new record with initial values
		prevUnlockedBalanceVault = 0
		prevLockedBalanceVault = 0
	} else if err != nil {
		return fmt.Errorf("error fetching previous balance Vault: %v", err)
	}

	// Update the unlocked balance with the deposit amount
	newUnlockedBalanceLP := prevUnlockedBalanceLP + amount
	newUnlockedBalanceVault := prevUnlockedBalanceVault + amount

	// Insert or update Liquidity_Providers
	queryInsertLP := `
		INSERT INTO public."Liquidity_Providers" (address, unlocked_balance, locked_balance, block_number)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (address, block_number) 
		DO UPDATE SET unlocked_balance = EXCLUDED.unlocked_balance, locked_balance = EXCLUDED.locked_balance;
	`
	_, err = conn.Exec(context.Background(), queryInsertLP, address, newUnlockedBalanceLP, prevLockedBalanceLP, newBlockNumber)
	if err != nil {
		return fmt.Errorf("error inserting or updating balance in Liquidity_Providers: %v", err)
	}

	// Insert or update Vault
	queryInsertVault := `
		INSERT INTO public."Vault" (block_number, unlocked_balance, locked_balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (block_number) 
		DO UPDATE SET unlocked_balance = EXCLUDED.unlocked_balance, locked_balance = EXCLUDED.locked_balance;
	`
	_, err = conn.Exec(context.Background(), queryInsertVault, newBlockNumber, newUnlockedBalanceVault, prevLockedBalanceVault)
	if err != nil {
		return fmt.Errorf("error inserting or updating balance in Vault: %v", err)
	}
	return nil
}
func onWithdrawal(address string, amount int64, newBlockNumber int64) error {
	// Connect to the database
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		return fmt.Errorf("unable to connect to database: %v", err)
	}
	defer conn.Close(context.Background())

	// Variables to store previous values
	var prevUnlockedBalanceLP, prevLockedBalanceLP int64
	var prevUnlockedBalanceVault, prevLockedBalanceVault, prevBlockVault int64

	// Query to select the latest row for the given address in Liquidity_Providers
	querySelectLP := `
		SELECT unlocked_balance, locked_balance
		FROM public."Liquidity_Providers"
		WHERE address = $1
		ORDER BY block_number DESC
		LIMIT 1;
	`
	rowLP := conn.QueryRow(context.Background(), querySelectLP, address)
	err = rowLP.Scan(&prevUnlockedBalanceLP, &prevLockedBalanceLP)

	if err == pgx.ErrNoRows {
		// Address does not exist in Liquidity_Providers, create a new record with initial values
		prevUnlockedBalanceLP = 0
		prevLockedBalanceLP = 0
	} else if err != nil {
		return fmt.Errorf("error fetching previous balance LP: %v", err)
	}
	// Query to select the latest row in Vault
	querySelectVault := `
		SELECT unlocked_balance, locked_balance, block_number
		FROM public."Vault"
		ORDER BY block_number DESC
		LIMIT 1;
	`
	rowVault := conn.QueryRow(context.Background(), querySelectVault)
	err = rowVault.Scan(&prevUnlockedBalanceVault, &prevLockedBalanceVault, &prevBlockVault)

	if err == pgx.ErrNoRows {
		// No record exists in Vault, create a new record with initial values
		prevUnlockedBalanceVault = 0
		prevLockedBalanceVault = 0
	} else if err != nil {
		return fmt.Errorf("error fetching previous balance Vault: %v", err)
	}

	// Update the unlocked balance with the withdrawal amount
	newUnlockedBalanceLP := prevUnlockedBalanceLP - amount
	newUnlockedBalanceVault := prevUnlockedBalanceVault - amount

	// Insert or update Liquidity_Providers
	queryInsertLP := `
		INSERT INTO public."Liquidity_Providers" (address, unlocked_balance, locked_balance, block_number)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (address, block_number) 
		DO UPDATE SET unlocked_balance = EXCLUDED.unlocked_balance, locked_balance = EXCLUDED.locked_balance;
	`
	_, err = conn.Exec(context.Background(), queryInsertLP, address, newUnlockedBalanceLP, prevLockedBalanceLP, newBlockNumber)
	if err != nil {
		return fmt.Errorf("error inserting or updating balance in Liquidity_Providers: %v", err)
	}

	// Insert or update Vault
	queryInsertVault := `
		INSERT INTO public."Vault" (block_number, unlocked_balance, locked_balance)
		VALUES ($1, $2, $3)
		ON CONFLICT (block_number) 
		DO UPDATE SET unlocked_balance = EXCLUDED.unlocked_balance, locked_balance = EXCLUDED.locked_balance;
	`
	if newBlockNumber == prevBlockVault {
		// Update the existing row in Vault if the block number is the same
		queryUpdateVault := `
			UPDATE public."Vault"
			SET unlocked_balance = $1, locked_balance = $2
			WHERE block_number = $3;
		`
		_, err = conn.Exec(context.Background(), queryUpdateVault, newUnlockedBalanceVault, prevLockedBalanceVault, newBlockNumber)
		if err != nil {
			return fmt.Errorf("error updating balance in Vault: %v", err)
		}
	} else {
		_, err = conn.Exec(context.Background(), queryInsertVault, newBlockNumber, newUnlockedBalanceVault, prevLockedBalanceVault)
		if err != nil {
			return fmt.Errorf("error inserting new balance in Vault: %v", err)
		}
	}

	return nil
}

func onDeployTransaction() error {

	return nil
}

func onWithdrawalQueued() error {
	return nil
}

func onQueuedLiquidityCollected() error {
	return nil
}

func onOptionRoundDeployed() error {
	return nil
}

func onAuctionStarted() error {
	return nil
}
func onBidAccepted() error {
	return nil
}
func onBidUpdated() error {
	return nil
}
func onAuctionEnded() error {
	return nil
}
func onOptionRoundSettled() error {
	return nil
}
func onOptionsExercised() error {
	return nil
}
func onUnusedBidsRefunded() error {
	return nil
}
func onOptionsMinted() error {
	return nil
}
