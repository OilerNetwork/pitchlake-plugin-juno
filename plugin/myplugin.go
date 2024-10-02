package main

import (
	"junoplugin/adaptors"
	"junoplugin/db"
	"junoplugin/events"
	"junoplugin/models"
	"log"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
	"gorm.io/gorm"
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

	tx := p.db.Conn.Begin()
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
				case "Deposit", "Withdraw": //Add withdrawQueue and collect queue case based on event
					var lpLocked, lpUnlocked, vaultLocked, vaultUnlocked uint64
					event.Data[0].SetUint64(lpLocked)
					event.Data[1].SetUint64(lpUnlocked)
					event.Data[2].SetUint64(vaultLocked)
					event.Data[3].SetUint64(vaultUnlocked)

					//Map the other parameters as well
					var newLPState = &(models.LiquidityProviderState{Address: event.Data[1].String()})
					p.db.UpsertLiquidityProviderState(tx, newLPState, block.Number)
					p.db.UpdateVaultFields(tx, map[string]interface{}{"unlocked_balance": vaultUnlocked, "locked_balance": vaultLocked, "latest_block": block.Number})
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
							p.db.UpdateOptionRoundFields(tx, roundAddress.String(), map[string]interface{}{
								"available_options":  event.Data[0],
								"starting_liquidity": event.Data[1],
								"state":              "Auctioning",
							})
							p.db.UpdateAllLiquidityProvidersBalancesAuctionStart(tx, block.Number)
							p.db.UpdateVaultBalancesAuctionStart(tx, block.Number)
						case "AuctionEnded":
							var optionsSold, clearingPrice, clearingNonce, clearingOptionsSold uint64
							event.Data[0].SetUint64(optionsSold)
							event.Data[3].SetUint64(clearingPrice)
							event.Data[1].SetUint64(clearingNonce)
							event.Data[2].SetUint64(clearingOptionsSold)
							premiums := optionsSold * clearingPrice

							var unsoldLiquidity = p.prevStateOptionRound.StartingLiquidity - p.prevStateOptionRound.StartingLiquidity*p.prevStateOptionRound.SoldOptions/p.prevStateOptionRound.AvailableOptions
							p.db.UpdateAllLiquidityProvidersBalancesAuctionEnd(tx, p.prevStateOptionRound.StartingLiquidity, unsoldLiquidity, premiums, block.Number)
							p.db.UpdateVaultBalancesAuctionEnd(tx, unsoldLiquidity, premiums, block.Number)
							p.db.UpdateBiddersAuctionEnd(tx, clearingPrice, optionsSold, p.prevStateVault.CurrentRound, clearingOptionsSold)
							p.db.UpdateOptionRoundAuctionEnd(tx, roundAddress.String(), clearingPrice, optionsSold)
						case "OptionRoundSettled":
							var totalPayout, settlementPrice uint64
							event.Data[0].SetUint64(totalPayout)
							event.Data[2].SetUint64(settlementPrice)
							p.db.UpdateVaultBalancesOptionSettle(tx, p.prevStateOptionRound.StartingLiquidity, p.prevStateOptionRound.QueuedLiquidity, block.Number)
							p.db.UpdateAllLiquidityProvidersBalancesOptionSettle(tx, p.prevStateOptionRound.RoundID, p.prevStateOptionRound.StartingLiquidity, p.prevStateOptionRound.QueuedLiquidity, totalPayout, block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"settlement_price": settlementPrice,
								"total_payout":     totalPayout,
								"state":            "Settled",
							})
						case "BidAccepted":
							var bid models.Bid
							var bidAmount, bidPrice, bidNonce uint64
							event.Data[0].SetUint64(bidNonce)

							event.Data[2].SetUint64(bidAmount)
							event.Data[3].SetUint64(bidPrice)
							bid.Address = event.Keys[0].String()
							bid.BidID = event.Data[1].String()
							bid.RoundID = p.prevStateOptionRound.RoundID
							p.db.CreateBid(tx, &bid)
						case "BidUpdated":
							tx.Model(models.Bid{}).Where("bid_id = ?", event.Data[0].String()).Update("amount", gorm.Expr("amount + ?", event.Data[1]))

						case "OptionsMinted":
							optionRound, err := p.db.GetOptionRoundByAddress(tx, roundAddress.String())
							if err != nil {
								return err
							}

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"has_minted": true,
							})
						case "UnusedBidsRefunded":
							optionRound, err := p.db.GetOptionRoundByAddress(tx, roundAddress.String())
							if err != nil {
								return err
							}

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"has_refunded": true,
							})
						case "OptionsExercised":
							optionRound, err := p.db.GetOptionRoundByAddress(tx, roundAddress.String())
							if err != nil {
								return err
							}

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"has_minted": true,
							})
						case "Transfer":
						}

					}
				}
			}
		}
	}
	tx.Commit()
	return nil
}

func (p *pitchlakePlugin) RevertBlock(from, to *junoplugin.BlockAndStateUpdate, reverseStateDiff *core.StateDiff) error {
	p.log.Println("ExamplePlugin NewBlock called")
	tx := p.db.Conn.Begin()
	length := len(from.Block.Receipts)
	for i := length - 1; i >= 0; i-- {
		receipt := from.Block.Receipts[i]
		for _, event := range receipt.Events {
			if event.From.Equal(p.vaultAddress) {
				eventName, err := events.DecodeEventNameVault(event.Keys[0].String())
				if err != nil {
					log.Fatalf("Failed to decode event: %v", err)
					return err
				}
				switch eventName {
				case "Deposit", "Withdraw",
					"QueuedLiquidityCollected": //Add withdraw queue

					//Map the other parameters as well
					// var newLPState = &(models.LiquidityProviderState{Address: event.Data[1].String()})
					// var newVaultState = &(models.VaultState{Address: p.vaultAddress.String()})
					// p.db.UpsertLiquidityProviderState(newLPState)
					// p.db.UpdateVaultState(newVaultState)
					p.db.RevertVaultState(tx, p.vaultAddress.String(), from.Block.Number)
					p.db.RevertLPState(tx, event.Keys[0].String(), from.Block.Number)
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
							p.db.RevertVaultState(tx, p.vaultAddress.String(), from.Block.Number)
							p.db.RevertAllLPState(tx, from.Block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"available_options":  0,
								"starting_liquidity": 0,
								"state":              "Open",
							})
						case "AuctionEnded":
							p.db.RevertVaultState(tx, p.vaultAddress.String(), from.Block.Number)
							p.db.RevertAllLPState(tx, from.Block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"clearing_price": nil,
								"options_sold":   nil,
								"state":          "Auctioning",
							})
							p.db.UpdateAllOptionBuyerFields(tx, p.prevStateOptionRound.RoundID, map[string]interface{}{
								"tokenizable_options": 0,
								"refundable_amount":   0,
							})

						case "OptionRoundSettled":
							p.db.RevertVaultState(tx, p.vaultAddress.String(), from.Block.Number)
							p.db.RevertAllLPState(tx, from.Block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"settlement_price": 0,
								"total_payout":     0,
								"state":            "Running",
							})
						case "BidAccepted":
							id := event.Data[1].String()
							p.db.DeleteBid(id, p.prevStateOptionRound.RoundID)
						case "BidUpdated":
							tx.Model(models.Bid{}).Where("bid_id = ?", event.Data[0].String()).Update("amount", gorm.Expr("amount - ?", event.Data[1]))
						case "OptionsMinted":
							optionRound, err := p.db.GetOptionRoundByAddress(tx, roundAddress.String())
							if err != nil {
								return err
							}
							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"has_minted": false,
							})
							// optionRound, err := p.db.GetOptionRoundByAddress(roundAddress.String())
							// if err != nil {
							// 	return err
							// }

							// p.db.UpdateOptionBuyerFields(event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
							// 	"tokenizable_options": 0,
							// })
						case "UnusedBidsRefunded":
							optionRound, err := p.db.GetOptionRoundByAddress(tx, roundAddress.String())
							if err != nil {
								return err
							}
							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"has_refunded": false,
							})
							// optionRound, err := p.db.GetOptionRoundByAddress(roundAddress.String())
							// if err != nil {
							// 	return err
							// }

							// p.db.UpdateOptionBuyerFields(event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
							// 	"refundable_amount": 0,
							// })
						case "OptionsExercised":
							optionRound, err := p.db.GetOptionRoundByAddress(tx, roundAddress.String())
							if err != nil {
								return err
							}
							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
								"has_minted": false,
							})
							// optionRound, err := p.db.GetOptionRoundByAddress(roundAddress.String())
							// if err != nil {
							// 	return err
							// }

							// p.db.UpdateOptionBuyerFields(event.Keys[0].String(), optionRound.RoundID, map[string]interface{}{
							// 	"tokenizable_options": 0,
							// })
						case "Transfer":
						}

					}
				}
			}
		}
	}
	tx.Commit()
	return nil
}
