package myplugin

import (
	"context"
	"fmt"
	"junoplugin/db"
	"log"
	"os"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
	"github.com/jackc/pgx/v5"
)

// Todo: push this stuff to a config file / cmd line
var PostgresConnString = "INSERT_DETAILS_HERE"
var vaultAddress = new(felt.Felt).SetUint64(1)
var previousState = int64(1)

//go:generate go build -buildmode=plugin -o ../../build/plugin.so ./example.go
type pitchlakePlugin struct {
	vaultAddress *felt.Felt
	prevState    int64
	db           *db.PostgresDB
	log          *log.Logger
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{}

// Ensure the plugin and Juno client follow the same interface
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	postgresDB, err := db.NewPostgresDB(PostgresConnString)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
		return err
	}
	p.db = postgresDB
	p.prevState = previousState
	p.vaultAddress = vaultAddress
	p.log = log.Default()
	return nil
}

func (p *pitchlakePlugin) Shutdown() error {
	p.log.Println("Calling Shutdown() in plugin")
	err := p.db.Close()
	if err != nil {
		p.log.Fatalf("Failed to close connection to Postgres: %v", err)
		return err
	}
	return nil
}

func (p *pitchlakePlugin) NewBlock(block *core.Block, stateUpdate *core.StateUpdate, newClasses map[felt.Felt]core.Class) error {
	p.log.Println("ExamplePlugin NewBlock called")
	for _, receipt := range block.Receipts {
		for _, event := range receipt.Events {
			if event.From.Equal(p.vaultAddress) {
				//Event is from the contract, perform actions here

				//If the event is a state transition, update locked unlocked balances

				//If the event is deposit/withdraw update lp balance
			}
		}
	}
	return nil
}

func (p *pitchlakePlugin) RevertBlock(from, to *junoplugin.BlockAndStateUpdate, reverseStateDiff *core.StateDiff) error {
	p.log.Println("ExamplePlugin RevertBlock called")
	return nil
}

func onDeposit([]any) error {

	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	return nil
}

func onWithdrawal() error {
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
