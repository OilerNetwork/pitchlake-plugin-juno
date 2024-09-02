package main

import (
	"fmt"
	"junoplugin/adaptors"
	"junoplugin/db"
	"junoplugin/events"
	"junoplugin/models"
	"log"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
)

// Todo: push this stuff to a config file / cmd line
var dsn = "INSERT_DETAILS_HERE"
var vaultAddress = new(felt.Felt).SetUint64(1)
var previousState = int64(1)

//go:generate go build -buildmode=plugin -o ../../build/plugin.so ./example.go
type pitchlakePlugin struct {
	vaultAddress         *felt.Felt
	roundAddresses       []*felt.Felt
	prevStateVault       *models.VaultState
	prevStateOptionRound *models.OptionRound
	db                   *db.DB
	log                  *log.Logger
	adaptors             *adaptors.PostgresAdapter
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{}

// Ensure the plugin and Juno client follow the same interface
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	db, err := db.Init(dsn)
	if err != nil {
		log.Fatalf("Failed to initialise db: %v", err)
		return err
	}
	p.db = db
	p.vaultAddress = vaultAddress
	p.log = log.Default()
	return nil
}

func (p *pitchlakePlugin) Shutdown() error {
	p.log.Println("Calling Shutdown() in plugin")
	p.db.Close()
	return nil
}

func (p *pitchlakePlugin) NewBlock(block *core.Block, stateUpdate *core.StateUpdate, newClasses map[felt.Felt]core.Class) error {

	p.log.Println("ExamplePlugin NewBlock called")
	for _, receipt := range block.Receipts {
		for _, event := range receipt.Events {
			if event.From.Equal(p.vaultAddress) {
				eventName, err := events.DecodeEventNameVault(event.Keys[0].String())
				if err != nil {
					log.Fatalf("Failed to decode event: %v", err)
					return err
				}
				switch eventName {
				case "Deposit", "Withdraw", "WithdrawalQueued",
					"QueuedLiquidityCollected":

					//Map the other parameters as well
					var newLPState = &(models.LiquidityProviderState{Address: event.Data[1].String()})
					var newVaultState = &(models.VaultState{Address: p.vaultAddress.String()})
					p.db.UpsertLiquidityProviderState(newLPState)
					p.db.UpdateVaultState(newVaultState)
				case "OptionRoundDeployed":
				}

			} else {

				for _, roundAddress := range p.roundAddresses {
					if event.From.Equal(roundAddress) {
						eventName, err := events.DecodeEventNameRound(event.Keys[0].String())
						if err != nil {
							log.Fatalf("Failed to decode event: %v", err)
							return err
						}
						switch eventName {
						case "AuctionStarted":
							p.db.UpdateOptionRoundFields(roundAddress.String(), map[string]interface{}{
								"available_options":  event.Data[0],
								"starting_liquidity": event.Data[1],
							})
							p.db.UpdateAllLiquidityProvidersBalancesAuctionStart()
							p.db.UpdateVaultBalancesAuctionStart()
						case "AuctionEnded":
							var optionsSold, clearingPrice, clearingNonce, clearingOptionsSold uint64
							event.Data[0].SetUint64(optionsSold)
							event.Data[3].SetUint64(clearingPrice)
							event.Data[1].SetUint64(clearingNonce)
							event.Data[2].SetUint64(clearingOptionsSold)
							premiums := optionsSold * clearingPrice

							var unsoldLiquidity = p.prevStateOptionRound.StartingLiquidity - p.prevStateOptionRound.StartingLiquidity*p.prevStateOptionRound.SoldOptions/p.prevStateOptionRound.AvailableOptions
							p.db.UpdateAllLiquidityProvidersBalancesAuctionEnd(p.prevStateOptionRound.StartingLiquidity, unsoldLiquidity, premiums)
							p.db.UpdateVaultBalancesAuctionEnd(unsoldLiquidity, premiums)
							p.db.UpdateBiddersAuctionEnd(clearingPrice, optionsSold, p.prevStateVault.CurrentRound, clearingOptionsSold)
							p.db.UpdateOptionRoundAuctionEnd(roundAddress.String(), clearingPrice, optionsSold)
						case "OptionRoundSettled":
							var totalPayout uint64
							event.Data[0].SetUint64(totalPayout)
							p.db.UpdateVaultBalancesOptionSettle(p.prevStateOptionRound.StartingLiquidity, p.prevStateOptionRound.QueuedLiquidity)
							p.db.UpdateAllLiquidityProvidersBalancesOptionSettle(p.prevStateOptionRound.RoundID, p.prevStateOptionRound.StartingLiquidity, p.prevStateOptionRound.QueuedLiquidity, totalPayout)
						case "BidAccepted":
							var bid models.Bid
							var bidAmount, bidPrice, bidNonce uint64
							event.Data[0].SetUint64(bidNonce)

							event.Data[2].SetUint64(bidAmount)
							event.Data[3].SetUint64(bidPrice)
							bid.Address = event.Keys[0].String()
							bid.BidID = event.Data[1].String()
							bid.RoundID = p.prevStateOptionRound.RoundID
							p.db.CreateBid(&bid)
						case "BidUpdated":

						case "OptionsMinted":
							optionRound, err := p.db.GetOptionRoundByAddress(roundAddress.String())
							if err != nil {
								return err
							}

							p.db.UpdateOptionBuyerFields(event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"tokenizable_options": 0,
							})
						case "UnusedBidsRefunded":
							optionRound, err := p.db.GetOptionRoundByAddress(roundAddress.String())
							if err != nil {
								return err
							}

							p.db.UpdateOptionBuyerFields(event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"refundable_balance": 0,
							})
						case "OptionsExercised":
							optionRound, err := p.db.GetOptionRoundByAddress(roundAddress.String())
							if err != nil {
								return err
							}

							p.db.UpdateOptionBuyerFields(event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"tokenizable_options": 0,
							})
						case "Transfer":
						}

					}
				}
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
	//get previous value

	//insert new value

	var query = `INSERT INTO public."Liquidity_Providers"(
	address, unlocked_balance, locked_balance, block_number)
	VALUES (?, ?, ?, ?);`
	fmt.Println(query)
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
