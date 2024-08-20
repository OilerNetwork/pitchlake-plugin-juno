package main

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
	var prevUnlockedBalance, prevLockedBalance int64
	var recordExists bool

	// Query to select the latest row for the given address
	querySelect := `
		SELECT unlocked_balance, locked_balance
		FROM public."Liquidity_Providers"
		WHERE address = $1
		ORDER BY block_number DESC
		LIMIT 1;
	`

	row := conn.QueryRow(context.Background(), querySelect, address)
	err = row.Scan(&prevUnlockedBalance, &prevLockedBalance)

	if err == pgx.ErrNoRows {
		// Address does not exist, create a new record with initial values
		prevUnlockedBalance = 0
		prevLockedBalance = 0
		recordExists = false
	} else if err != nil {
		return fmt.Errorf("error fetching previous balance: %v", err)
	} else {
		recordExists = true
	}

	// Update the unlocked balance with the deposit amount
	newUnlockedBalance := prevUnlockedBalance + amount

	// Insert the new balance into the database
	queryInsert := `
		INSERT INTO public."Liquidity_Providers" (address, unlocked_balance, locked_balance, block_number)
		VALUES ($1, $2, $3, $4);
	`
	_, err = conn.Exec(context.Background(), queryInsert, address, newUnlockedBalance, prevLockedBalance, newBlockNumber)
	if err != nil {
		return fmt.Errorf("error inserting new balance: %v", err)
	}

	if recordExists {
		fmt.Printf("Deposit recorded for existing record: Address: %s, Amount: %d, New Unlocked Balance: %d, Block Number: %d\n", address, amount, newUnlockedBalance, newBlockNumber)
	} else {
		fmt.Printf("Deposit recorded for new record: Address: %s, Amount: %d, New Unlocked Balance: %d, Block Number: %d\n", address, amount, newUnlockedBalance, newBlockNumber)
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

	// Query to select the latest row for the given address
	var prevUnlockedBalance, prevLockedBalance int64
	querySelect := `
		SELECT unlocked_balance, locked_balance
		FROM public."Liquidity_Providers"
		WHERE address = $1
		ORDER BY block_number DESC
		LIMIT 1;
	`
	err = conn.QueryRow(context.Background(), querySelect, address).Scan(&prevUnlockedBalance, &prevLockedBalance)
	if err != nil && err != pgx.ErrNoRows {
		return fmt.Errorf("error fetching previous balance: %v", err)
	}

	// Update the unlocked balance with the withdrawal amount
	newUnlockedBalance := prevUnlockedBalance - amount

	// Insert the new balance into the database
	queryInsert := `
		INSERT INTO public."Liquidity_Providers" (address, unlocked_balance, locked_balance, block_number)
		VALUES ($1, $2, $3, $4);
	`
	_, err = conn.Exec(context.Background(), queryInsert, address, newUnlockedBalance, prevLockedBalance, newBlockNumber)
	if err != nil {
		return fmt.Errorf("error inserting new balance: %v", err)
	}

	fmt.Printf("Withdrawal recorded: Address: %s, Amount: %d, New Unlocked Balance: %d, Block Number: %d\n", address, amount, newUnlockedBalance, newBlockNumber)
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
