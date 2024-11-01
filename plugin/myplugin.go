package main

import (
	"fmt"
	"junoplugin/db"
	"junoplugin/events"
	"junoplugin/models"
	"log"
	"math/big"
	"os"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
)

// Todo: push this stuff to a config file / cmd line

//go:generate go build -buildmode=plugin -o ../../build/plugin.so ./example.go
type pitchlakePlugin struct {
	vaultHash    string
	vaultAddress string
	//vaultAddresses       []string
	roundAddresses       []string
	udcAddress           string
	prevStateOptionRound *models.OptionRound
	db                   *db.DB
	log                  *log.Logger
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{}

// Ensure the plugin and Juno client follow the same interface
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	dbUrl := os.Getenv("DB_URL")
	p.udcAddress = os.Getenv("UDC")
	db, err := db.Init(dbUrl)
	if err != nil {
		log.Fatalf("Failed to initialise db: %v", err)
		return err
	}
	p.db = db
	p.vaultHash = os.Getenv("VAULT_HASH")
	p.log = log.Default()
	return nil
}

func (p *pitchlakePlugin) Shutdown() error {
	p.log.Println("Calling Shutdown() in plugin")
	p.db.Close()
	return nil
}

func keccak256(eventName string) string {
	hasher := sha3.NewLegacyKeccak256()

	// Write the event name as bytes to the hasher
	hasher.Write([]byte(eventName))

	// Compute the full 256-bit hash
	hashBytes := hasher.Sum(nil)

	// Convert the hash to a big integer
	hashInt := new(big.Int).SetBytes(hashBytes)

	// Apply a 250-bit mask to fit StarkNet's felt requirements
	mask := new(big.Int).Lsh(big.NewInt(1), 250)
	mask.Sub(mask, big.NewInt(1))
	hashInt.And(hashInt, mask)

	// Convert the masked hash to a hexadecimal string with "0x" prefix
	return "0x" + hashInt.Text(16)
}

func (p *pitchlakePlugin) NewBlock(block *core.Block, stateUpdate *core.StateUpdate, newClasses map[felt.Felt]core.Class) error {

	tx := p.db.Conn.Begin()
	p.log.Println("ExamplePlugin NewBlock called")
	for _, receipt := range block.Receipts {
		for _, event := range receipt.Events {
			fromAddress := event.From.String()
			log.Printf("EVENT.FROM %s", event.From.String())
			if fromAddress == p.udcAddress {
				eventHash := keccak256("ContractDeployed")
				address := FeltToHexString(event.Data[0].Bytes())
				classHash := FeltToHexString(event.Data[3].Bytes())

				if eventHash == event.Keys[0].String() {

					if classHash == p.vaultHash {
						p.vaultAddress = address
					}
				}
			} else if fromAddress == p.vaultAddress {
				eventName, err := events.DecodeEventNameVault(event.Keys[0].String())
				if err != nil {
					log.Fatalf("Failed to decode event: %v", err)
					return err
				}
				switch eventName {
				case "Deposit", "Withdraw": //Add withdrawQueue and collect queue case based on event
					lpUnlocked := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[3].Bytes())
					vaultUnlocked := CombineFeltToBigInt(event.Data[4].Bytes(), event.Data[5].Bytes())

					//Map the other parameters as well
					var newLPState = &(models.LiquidityProviderState{Address: event.Data[1].String(), UnlockedBalance: lpUnlocked, LatestBlock: block.Number})
					p.db.UpsertLiquidityProviderState(tx, newLPState, block.Number)
					p.db.UpdateVaultFields(tx, p.vaultAddress, map[string]interface{}{"unlocked_balance": vaultUnlocked, "latest_block": block.Number})
				case "OptionRoundDeployed":
				}

			} else {

				for _, roundAddress := range p.roundAddresses {
					if fromAddress == roundAddress {
						eventName, err := events.DecodeEventNameRound(event.Keys[0].String())
						if err != nil {
							log.Fatalf("Failed to decode event: %v", err)
							return err
						}
						switch eventName {
						case "AuctionStarted":
							availableOptions := CombineFeltToBigInt(event.Data[0].Bytes(), event.Data[1].Bytes())
							startingLiquidity := CombineFeltToBigInt(event.Data[0].Bytes(), event.Data[1].Bytes())
							p.db.UpdateOptionRoundFields(tx, roundAddress, map[string]interface{}{
								"available_options":  availableOptions,
								"starting_liquidity": startingLiquidity,
								"state":              "Auctioning",
							})
							p.db.UpdateAllLiquidityProvidersBalancesAuctionStart(tx, block.Number)
							p.db.UpdateVaultBalancesAuctionStart(tx, block.Number)
						case "AuctionEnded":

							optionsSold := CombineFeltToBigInt(event.Data[1].Bytes(), event.Data[0].Bytes())
							clearingPrice := CombineFeltToBigInt(event.Data[3].Bytes(), event.Data[2].Bytes())
							clearingNonce := CombineFeltToBigInt(event.Data[5].Bytes(), event.Data[4].Bytes())
							premiums := models.BigInt{Int: new(big.Int).Mul(optionsSold.Int, clearingPrice.Int)}

							unsoldLiquidity := models.BigInt{Int: new(big.Int).Sub(
								p.prevStateOptionRound.StartingLiquidity.Int,
								new(big.Int).Div(
									new(big.Int).Mul(
										p.prevStateOptionRound.StartingLiquidity.Int,
										p.prevStateOptionRound.SoldOptions.Int,
									),
									p.prevStateOptionRound.AvailableOptions.Int,
								),
							)}
							p.db.UpdateAllLiquidityProvidersBalancesAuctionEnd(tx, p.prevStateOptionRound.StartingLiquidity, unsoldLiquidity, premiums, block.Number)
							p.db.UpdateVaultBalancesAuctionEnd(tx, unsoldLiquidity, premiums, block.Number)
							p.db.UpdateBiddersAuctionEnd(tx, roundAddress, clearingPrice, optionsSold, clearingNonce)
							p.db.UpdateOptionRoundAuctionEnd(tx, roundAddress, clearingPrice, optionsSold)
						case "OptionRoundSettled":

							totalPayout := CombineFeltToBigInt(event.Data[0].Bytes(), event.Data[1].Bytes())
							settlementPrice := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[3].Bytes())
							p.db.UpdateVaultBalancesOptionSettle(tx, p.prevStateOptionRound.StartingLiquidity, p.prevStateOptionRound.QueuedLiquidity, block.Number)
							p.db.UpdateAllLiquidityProvidersBalancesOptionSettle(tx, roundAddress, p.prevStateOptionRound.StartingLiquidity, p.prevStateOptionRound.QueuedLiquidity, totalPayout, models.BigInt{Int: new(big.Int).SetUint64(block.Number)})
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"settlement_price": settlementPrice,
								"total_payout":     totalPayout,
								"state":            "Settled",
							})
						case "BidAccepted":
							bidAmount := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[1].Bytes())
							bidPrice := CombineFeltToBigInt(event.Data[4].Bytes(), event.Data[3].Bytes())
							treeNonce := event.Data[5].Uint64()

							var bid models.Bid
							bid.BuyerAddress = event.Keys[0].String()
							bid.BidID = event.Data[0].String()
							bid.RoundAddress = roundAddress
							bid.Amount = bidAmount
							bid.Price = bidPrice
							bid.TreeNonce = treeNonce
							p.db.CreateBid(tx, &bid)
						case "BidUpdated":
							tx.Model(models.Bid{}).Where("bid_id = ?", event.Data[0].String()).Updates(map[string]interface{}{
								"amount":     gorm.Expr("amount + ?", CombineFeltToBigInt(event.Data[1].Bytes(), event.Data[2].Bytes())),
								"tree_nonce": event.Data[3].Uint64(),
							})

						case "OptionsMinted":

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), roundAddress, map[string]interface{}{
								"has_minted": true,
							})
						case "UnusedBidsRefunded":

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), roundAddress, map[string]interface{}{
								"has_refunded": true,
							})
						case "OptionsExercised":

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), roundAddress, map[string]interface{}{
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

			fromAddress := event.From.String()
			if fromAddress == p.vaultAddress {
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
					p.db.RevertVaultState(tx, p.vaultAddress, from.Block.Number)
					p.db.RevertLPState(tx, event.Keys[0].String(), from.Block.Number)
				case "OptionRoundDeployed":
				}

			} else {

				for _, roundAddress := range p.roundAddresses {
					if fromAddress == roundAddress {
						eventName, err := events.DecodeEventNameRound(event.Keys[0].String())
						if err != nil {
							log.Fatalf("Failed to decode event: %v", err)
							return err
						}
						switch eventName {
						case "AuctionStarted":
							p.db.RevertVaultState(tx, p.vaultAddress, from.Block.Number)
							p.db.RevertAllLPState(tx, from.Block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"available_options":  0,
								"starting_liquidity": 0,
								"state":              "Open",
							})
						case "AuctionEnded":
							p.db.RevertVaultState(tx, p.vaultAddress, from.Block.Number)
							p.db.RevertAllLPState(tx, from.Block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"clearing_price": nil,
								"options_sold":   nil,
								"state":          "Auctioning",
							})
							p.db.UpdateAllOptionBuyerFields(tx, roundAddress, map[string]interface{}{
								"tokenizable_options": 0,
								"refundable_amount":   0,
							})

						case "OptionRoundSettled":
							p.db.RevertVaultState(tx, p.vaultAddress, from.Block.Number)
							p.db.RevertAllLPState(tx, from.Block.Number)
							p.db.UpdateOptionRoundFields(tx, p.prevStateOptionRound.Address, map[string]interface{}{
								"settlement_price": 0,
								"total_payout":     0,
								"state":            "Running",
							})
						case "BidAccepted":
							id := event.Data[1].String()
							p.db.DeleteBid(tx, id, roundAddress)
						case "BidUpdated":
							tx.Model(models.Bid{}).Where("bid_id = ?", event.Data[0].String()).Updates(map[string]interface{}{
								"amount":     gorm.Expr("amount - ?", CombineFeltToBigInt(event.Data[1].Bytes(), event.Data[2].Bytes())),
								"tree_nonce": event.Data[4].Uint64(),
							})
						case "OptionsMinted":
							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), roundAddress, map[string]interface{}{
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
							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), roundAddress, map[string]interface{}{
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

							p.db.UpdateOptionBuyerFields(tx, event.Keys[0].String(), roundAddress, map[string]interface{}{
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

func CombineFeltToBigInt(highFelt, lowFelt [32]byte) models.BigInt {
	combinedBytes := make([]byte, 64) // 32 bytes for highFelt and 32 bytes for lowFelt

	// Copy highFelt into the first 32 bytes
	copy(combinedBytes[0:32], highFelt[:])

	// Copy lowFelt into the next 32 bytes
	copy(combinedBytes[32:64], lowFelt[:])

	// Convert the combined bytes to a big.Int
	combinedInt := models.BigInt{Int: new(big.Int).SetBytes(combinedBytes)}

	return combinedInt
}

func FeltToHexString(felt [32]byte) string {

	combinedInt := models.BigInt{Int: new(big.Int).SetBytes(felt[:])}
	// Assuming `f.Value` holds the *big.Int representation of the felt
	return "0x" + combinedInt.Text(16)
}

func BigIntToHexString(f big.Int) string {

	// Assuming `f.Value` holds the *big.Int representation of the felt
	return "0x" + f.Text(16)
}

func DecimalStringToHexString(decimalString string) (string, error) {
	// Create a new big.Int and set its value from the decimal string
	num := new(big.Int)
	_, success := num.SetString(decimalString, 10)
	if !success {
		return "", fmt.Errorf("invalid decimal string")
	}

	// Convert the big.Int to a hexadecimal string
	hexString := num.Text(16)

	// Add "0x" prefix for clarity
	return "0x" + hexString, nil
}
