package main

import (
	"junoplugin/adaptors"
	"junoplugin/db"
	"junoplugin/models"
	"log"
	"os"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
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
	junoAdaptor          *adaptors.JunoAdaptor
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{}

// Ensure the plugin and Juno client follow the same interface
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	dbUrl := os.Getenv("DB_URL")
	udcAddress := os.Getenv("UDC_ADDRESS")
	p.udcAddress = udcAddress
	dbClient, err := db.Init(dbUrl)
	if err != nil {
		log.Fatalf("Failed to initialise db: %v", err)
		return err
	}
	p.junoAdaptor = &adaptors.JunoAdaptor{}
	p.db = dbClient
	p.vaultHash = os.Getenv("VAULT_HASH")
	p.log = log.Default()

	p.log.Println("STARTED, UDC %v %v %v", p.udcAddress, dbUrl, udcAddress)
	//Add function to catch up on vaults/rounds that are not synced to currentBlock
	return nil
}

func (p *pitchlakePlugin) Shutdown() error {
	p.log.Println("Calling Shutdown() in plugin")
	p.db.Close()
	return nil
}

func (p *pitchlakePlugin) NewBlock(block *core.Block, stateUpdate *core.StateUpdate, newClasses map[felt.Felt]core.Class) error {

	p.db.Begin()
	p.log.Println("ExamplePlugin NewBlock called")
	for _, receipt := range block.Receipts {
		for i, event := range receipt.Events {
			fromAddress := event.From.String()

			if fromAddress == p.udcAddress {
				log.Printf("UDC", event.From.String())
				p.processUDC(receipt.Events[i-1], event, block.Number)
			} else if fromAddress == p.vaultAddress {
				p.processVaultEvent(fromAddress, event, block.Number)
			} else {
				for _, roundAddress := range p.roundAddresses {
					if fromAddress == roundAddress {
						p.processRoundEvent(roundAddress, event, block.Number)
					}
				}
			}
		}
	}
	p.db.Commit()
	return nil
}

func (p *pitchlakePlugin) RevertBlock(from, to *junoplugin.BlockAndStateUpdate, reverseStateDiff *core.StateDiff) error {
	p.log.Println("ExamplePlugin NewBlock called")
	p.db.Begin()
	length := len(from.Block.Receipts)
	for i := length - 1; i >= 0; i-- {
		receipt := from.Block.Receipts[i]
		for _, event := range receipt.Events {

			fromAddress := event.From.String()
			if fromAddress == p.vaultAddress {
				p.revertVaultEvent(fromAddress, event, from.Block.Number)
			} else {

				for _, roundAddress := range p.roundAddresses {
					if fromAddress == roundAddress {
						p.revertRoundEvent(roundAddress, event, from.Block.Number)
					}
				}
			}
		}
	}
	p.db.Commit()
	return nil
}

func (p *pitchlakePlugin) processUDC(previousevent, event *core.Event, blockNumber uint64) error {
	eventHash := adaptors.Keccak256("ContractDeployed")
	address := adaptors.FeltToHexString(event.Data[0].Bytes())
	classHash := adaptors.FeltToHexString(event.Data[3].Bytes())
	vaultAddress := os.Getenv("VAULT_ADDRESS")
	log.Printf("address", address, p.vaultAddress)
	if address == vaultAddress {
		log.Printf("MATCHED")
		p.vaultAddress = address
		vault := models.VaultState{
			CurrentRound:    *models.NewBigInt("1"),
			UnlockedBalance: *models.NewBigInt("0"),
			LockedBalance:   *models.NewBigInt("0"),
			StashedBalance:  *models.NewBigInt("0"),
			Address:         address,
			LatestBlock:     blockNumber,
		}
		p.db.CreateVault(&vault)
		p.processVaultEvent(vaultAddress, previousevent, blockNumber)

	}
	if eventHash == event.Keys[0].String() {

		if classHash == p.vaultHash {

		}

	}
	return nil
}

func (p *pitchlakePlugin) processVaultEvent(vaultAddress string, event *core.Event, blockNumber uint64) error {
	eventName, err := adaptors.DecodeEventNameVault(event.Keys[0].String())
	if err != nil {
		log.Fatalf("Failed to decode event: %v", err)
		return err
	}
	switch eventName {
	case "Deposit", "Withdraw": //Add withdrawQueue and collect queue case based on event
		lpAddress, lpUnlocked, vaultUnlocked := p.junoAdaptor.DepositOrWithdraw(*event)

		p.db.DepositOrWithdrawIndex(vaultAddress, lpAddress, lpUnlocked, vaultUnlocked, blockNumber)
		//Map the other parameters as well

	case "OptionRoundDeployed":

		optionRound := p.junoAdaptor.RoundDeployed(*event)

		p.db.RoundDeployedIndex(optionRound)
		p.roundAddresses = append(p.roundAddresses, optionRound.Address)
	}
	return nil
}

func (p *pitchlakePlugin) processRoundEvent(roundAddress string, event *core.Event, blockNumber uint64) error {
	var err error
	p.prevStateOptionRound, err = p.db.GetOptionRoundByAddress(roundAddress)
	if err != nil {
		return err
	}
	eventName, err := adaptors.DecodeEventNameRound(event.Keys[0].String())
	if err != nil {
		log.Fatalf("Failed to decode event: %v", err)
		return err
	}
	switch eventName {
	case "PricingDataSet":
		strikePrice, capLevel, reservePrice := p.junoAdaptor.PricingDataSet(*event)
		p.db.PricingDataSetIndex(roundAddress, strikePrice, capLevel, reservePrice)
	case "AuctionStarted":
		availableOptions, startingLiquidity := p.junoAdaptor.AuctionStarted(*event)
		p.db.AuctionStartedIndex(roundAddress, blockNumber, availableOptions, startingLiquidity)
	case "AuctionEnded":
		optionsSold, clearingPrice, clearingNonce, premiums := p.junoAdaptor.AuctionEnded(*event)
		p.db.AuctionEndedIndex(*p.prevStateOptionRound, roundAddress, blockNumber, optionsSold, clearingPrice, clearingNonce, premiums)
	case "OptionRoundSettled":
		totalPayout, settlementPrice := p.junoAdaptor.RoundSettled(*event)
		p.db.RoundSettledIndex(*p.prevStateOptionRound, roundAddress, blockNumber, totalPayout, settlementPrice)
	case "BidAccepted":
		bid := p.junoAdaptor.BidAccepted(*event)
		p.db.BidAcceptedIndex(bid)
	case "BidUpdated":
		bidId, amount, _, treeNonceNew := p.junoAdaptor.BidUpdated(*event)
		p.db.BidUpdatedIndex(bidId, amount, treeNonceNew)
	case "OptionsMinted":
		p.db.UpdateOptionBuyerFields(event.Keys[0].String(), roundAddress, map[string]interface{}{
			"has_minted": true,
		})
	case "UnusedBidsRefunded":

		p.db.UpdateOptionBuyerFields(event.Keys[0].String(), roundAddress, map[string]interface{}{
			"has_refunded": true,
		})
	case "OptionsExercised":

		p.db.UpdateOptionBuyerFields(event.Keys[0].String(), roundAddress, map[string]interface{}{
			"has_minted": true,
		})
	case "Transfer":
	}

	return nil
}

func (p *pitchlakePlugin) revertVaultEvent(vaultAddress string, event *core.Event, blockNumber uint64) error {
	eventName, err := adaptors.DecodeEventNameVault(event.Keys[0].String())
	if err != nil {
		log.Fatalf("Failed to decode event: %v", err)
		return err
	}
	switch eventName {
	case "Deposit", "Withdraw",
		"QueuedLiquidityCollected": //Add withdraw queue

		lpAddress := event.Keys[1].String()
		p.db.DepositOrWithdrawRevert(vaultAddress, lpAddress, blockNumber)
	case "OptionRoundDeployed":
		roundAddress := adaptors.FeltToHexString(event.Data[2].Bytes())
		p.db.DeleteOptionRound(roundAddress)
	}

	return nil
}

func (p *pitchlakePlugin) revertRoundEvent(roundAddress string, event *core.Event, blockNumber uint64) error {
	eventName, err := adaptors.DecodeEventNameRound(event.Keys[0].String())
	if err != nil {
		log.Fatalf("Failed to decode event: %v", err)
		return err
	}
	p.prevStateOptionRound, err = p.db.GetOptionRoundByAddress(roundAddress)
	if err != nil {
		return err
	}
	switch eventName {
	case "AuctionStarted":
		p.db.AuctionStartedRevert(p.prevStateOptionRound.VaultAddress, roundAddress, blockNumber)
	case "AuctionEnded":
		p.db.AuctionEndedRevert(p.prevStateOptionRound.VaultAddress, roundAddress, blockNumber)

	case "OptionRoundSettled":
		p.db.RoundSettledRevert(p.prevStateOptionRound.VaultAddress, roundAddress, blockNumber)
	case "BidAccepted":
		id := event.Data[1].String()
		p.db.BidAcceptedRevert(id, roundAddress)
	case "BidUpdated":
		bidId, amount, treeNonceOld, _ := p.junoAdaptor.BidUpdated(*event)
		p.db.BidUpdatedRevert(bidId, amount, treeNonceOld)
	case "OptionsMinted":
		p.db.UpdateOptionBuyerFields(event.Keys[0].String(), roundAddress, map[string]interface{}{
			"has_minted": false,
		})
	case "UnusedBidsRefunded":
		p.db.UpdateOptionBuyerFields(event.Keys[0].String(), roundAddress, map[string]interface{}{
			"has_refunded": false,
		})
	case "OptionsExercised":

		p.db.UpdateOptionBuyerFields(event.Keys[0].String(), roundAddress, map[string]interface{}{
			"has_minted": false,
		})
	case "Transfer":
	}

	return nil
}
