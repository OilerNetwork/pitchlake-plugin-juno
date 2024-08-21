package db_client

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func getConnection() (*pgx.Conn, error) {
	conn, err := pgx.Connect(context.Background(), os.Getenv("DB_URL"))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}
	return conn, nil
}
func getLPCurrent(address string) (pgx.Row, error) {

	conn, err := getConnection()
	if err != nil {
		return nil, err
	}
	defer conn.Close(context.Background())
	var prevUnlockedBalanceLP, prevLockedBalanceLP int64
	querySelectLP := `
		SELECT unlocked_balance, locked_balance
		FROM public."Liquidity_Providers"
		WHERE address = $1
		ORDER BY block_number DESC
		LIMIT 1;
	`
	rowLP := conn.QueryRow(context.Background(), querySelectLP, address)
	err = rowLP.Scan(&prevUnlockedBalanceLP, &prevLockedBalanceLP)
	return rowLP, err
}

func insertLP(address string, unlockedBalance, lockedBalance, blockNumber int64) error {
	conn, err := getConnection()
	if err != nil {
		return err
	}
	defer conn.Close(context.Background())
	queryInsertLP := `
		INSERT INTO public."Liquidity_Providers" (address, unlocked_balance, locked_balance, block_number)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (address, block_number) 
		DO UPDATE SET unlocked_balance = EXCLUDED.unlocked_balance, locked_balance = EXCLUDED.locked_balance;
	`
	_, err = conn.Exec(context.Background(), queryInsertLP, address, unlockedBalance, lockedBalance, blockNumber)
	if err != nil {
		return fmt.Errorf("error inserting or updating balance in Liquidity_Providers: %v", err)
	}

	return nil
}
