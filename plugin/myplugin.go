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
	roundAddresses []string
	udcAddress     string
	db             *db.DB
	log            *log.Logger
	junoAdaptor    *adaptors.JunoAdaptor
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{}

// Ensure the plugin and Juno client follow the same interface
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	dbUrl := os.Getenv("DB_URL")
	udcAddress := os.Getenv("UDC_ADDRESS")
	p.udcAddress = udcAddress
	p.roundAddresses = make([]string, 0)
	dbClient, err := db.Init(dbUrl)
	if err != nil {
		log.Fatalf("Failed to initialise db: %v", err)
		return err
	}
	vaultAddresses, err := p.db.GetVaultAddresses()
	if err != nil {
		return err
	}
	if len(*vaultAddresses) > 0 {
		p.vaultAddress = (*vaultAddresses)[0]
	}

	//Make vault address multiple and add loop here to fetch rounds for each vault

	if p.vaultAddress != "" {

		roundAddresses, err := p.db.GetRoundAddressess(p.vaultAddress)
		if err != nil {
			return err
		}
		p.roundAddresses = *roundAddresses
	}

	p.junoAdaptor = &adaptors.JunoAdaptor{}
	p.db = dbClient
	p.vaultHash = os.Getenv("VAULT_HASH")
	p.log = log.Default()

	p.log.Printf("STARTED, UDC %v %v %v", p.udcAddress, dbUrl, udcAddress)
	//Add function to catch up on vaults/rounds that are not synced to currentBlock
	return nil
}

func (p *pitchlakePlugin) Shutdown() error {
	p.log.Println("Calling Shutdown() in plugin")
	p.db.Close()
	return nil
}

func (p *pitchlakePlugin) NewBlock(
	block *core.Block,
	stateUpdate *core.StateUpdate,
	newClasses map[felt.Felt]core.Class,
) error {

	p.db.Begin()
	p.log.Println("ExamplePlugin NewBlock called")
	for _, receipt := range block.Receipts {
		for i, event := range receipt.Events {
			fromAddress := event.From.String()

			if fromAddress == p.udcAddress {
				log.Printf("UDC", event.From.String())
				p.processUDC(receipt.Events, event, i, block.Number)
			} else if fromAddress == p.vaultAddress {
				p.processVaultEvent(fromAddress, event, block.Number)
			} else {
				for _, roundAddress := range p.roundAddresses {
					if fromAddress == roundAddress {
						p.processRoundEvent(roundAddress, event, block.Number)
						break
					}
				}
			}
		}
	}
	p.db.Commit()
	return nil
}

func (p *pitchlakePlugin) RevertBlock(
	from,
	to *junoplugin.BlockAndStateUpdate,
	reverseStateDiff *core.StateDiff,
) error {
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

func (p *pitchlakePlugin) processUDC(
	events []*core.Event,
	event *core.Event,
	index int,
	blockNumber uint64,
) error {
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
		if err := p.db.CreateVault(&vault); err != nil {
			log.Fatal(err)
			return err
		}
		log.Printf("index %v", index)
		p.processVaultEvent(vaultAddress, events[index-1], blockNumber)

	}
	if eventHash == event.Keys[0].String() {

		if classHash == p.vaultHash {

		}

	}
	return nil
}

func (p *pitchlakePlugin) processVaultEvent(
	vaultAddress string,
	event *core.Event,
	blockNumber uint64,
) error {

	eventName, err := adaptors.DecodeEventNameVault(event.Keys[0].String())
	if err != nil {
		log.Printf("Unknown Event")
		return nil
	}
	log.Printf("NAME %v", eventName)
	switch eventName {
	case "Deposit": //Add withdrawQueue and collect queue case based on event
		lpAddress,
			lpUnlocked,
			vaultUnlocked := p.junoAdaptor.DepositOrWithdraw(*event)

		err = p.db.DepositIndex(vaultAddress, lpAddress, lpUnlocked, vaultUnlocked, blockNumber)
		//Map the other parameters as well
	case "Withdrawal":
		lpAddress,
			lpUnlocked,
			vaultUnlocked := p.junoAdaptor.DepositOrWithdraw(*event)

		err = p.db.WithdrawIndex(vaultAddress, lpAddress, lpUnlocked, vaultUnlocked, blockNumber)
	case "WithdrawalQueued":
		lpAddress,
			bps,
			roundId,
			accountQueuedBefore,
			accountQueuedNow,
			vaultQueuedNow := p.junoAdaptor.WithdrawalQueued(*event)

		err = p.db.WithdrawalQueuedIndex(
			lpAddress,
			vaultAddress,
			roundId,
			bps,
			accountQueuedBefore,
			accountQueuedNow,
			vaultQueuedNow,
		)

	case "StashWithdrawn":
		lpAddress, amount, vaultStashed := p.junoAdaptor.StashWithdrawn(*event)
		err = p.db.StashWithdrawnIndex(
			vaultAddress,
			lpAddress,
			amount,
			vaultStashed,
			blockNumber,
		)
	case "OptionRoundDeployed":

		optionRound := p.junoAdaptor.RoundDeployed(*event)
		err = p.db.RoundDeployedIndex(optionRound)
		p.roundAddresses = append(p.roundAddresses, optionRound.Address)
	}
	if err != nil {
		return err
	}
	return nil
}

func (p *pitchlakePlugin) processRoundEvent(
	roundAddress string,
	event *core.Event,
	blockNumber uint64,
) error {
	var err error
	prevStateOptionRound := p.db.GetOptionRoundByAddress(roundAddress)
	if err != nil {
		return err
	}

	eventName, err := adaptors.DecodeEventNameRound(event.Keys[0].String())
	if err != nil {
		return nil
	}
	log.Printf("ADDRESSES %v", roundAddress)
	log.Printf("ROUND STATE PREV %v %v", prevStateOptionRound, eventName)
	switch eventName {
	case "PricingDataSet":
		strikePrice, capLevel, reservePrice := p.junoAdaptor.PricingDataSet(*event)
		err = p.db.PricingDataSetIndex(roundAddress, strikePrice, capLevel, reservePrice)
	case "AuctionStarted":
		availableOptions, startingLiquidity := p.junoAdaptor.AuctionStarted(*event)
		err = p.db.AuctionStartedIndex(
			prevStateOptionRound.VaultAddress,
			roundAddress,
			blockNumber,
			availableOptions,
			startingLiquidity,
		)

	case "AuctionEnded":
		optionsSold,
			clearingPrice,
			unsoldLiquidity,
			clearingNonce,
			premiums := p.junoAdaptor.AuctionEnded(*event)

		err = p.db.AuctionEndedIndex(
			prevStateOptionRound,
			roundAddress,
			blockNumber,
			clearingNonce,
			optionsSold,
			clearingPrice,
			premiums,
			unsoldLiquidity,
		)
	case "OptionRoundSettled":
		settlementPrice, payoutPerOption := p.junoAdaptor.RoundSettled(*event)
		if err := p.db.RoundSettledIndex(
			prevStateOptionRound,
			roundAddress,
			blockNumber,
			settlementPrice,
			prevStateOptionRound.SoldOptions,
			payoutPerOption,
		); err != nil {
			log.Fatal(err)
		}
	case "BidPlaced":
		bid, buyer := p.junoAdaptor.BidPlaced(*event)
		err = p.db.BidPlacedIndex(bid, buyer)
	case "BidUpdated":
		bidId, price, _, treeNonceNew := p.junoAdaptor.BidUpdated(*event)
		err = p.db.BidUpdatedIndex(event.From.String(), bidId, price, treeNonceNew)
	case "OptionsMinted", "OptionsExercised":
		buyerAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		err = p.db.UpdateOptionBuyerFields(
			buyerAddress,
			roundAddress,
			map[string]interface{}{
				"has_minted": true,
			})
	case "UnusedBidsRefunded":
		buyerAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		err = p.db.UpdateOptionBuyerFields(
			buyerAddress,
			roundAddress,
			map[string]interface{}{
				"has_refunded": true,
			})
	case "Transfer":
	}

	if err != nil {
		return err
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
		"StashWithdrawn": //Add withdraw queue

		lpAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
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
	prevStateOptionRound := p.db.GetOptionRoundByAddress(roundAddress)
	switch eventName {
	case "AuctionStarted":
		p.db.AuctionStartedRevert(prevStateOptionRound.VaultAddress, roundAddress, blockNumber)
	case "AuctionEnded":
		p.db.AuctionEndedRevert(prevStateOptionRound.VaultAddress, roundAddress, blockNumber)

	case "OptionRoundSettled":
		p.db.RoundSettledRevert(prevStateOptionRound.VaultAddress, roundAddress, blockNumber)
	case "BidAccepted":
		id := event.Data[1].String()
		p.db.BidAcceptedRevert(id, roundAddress)
	case "BidUpdated":
		bidId, amount, treeNonceOld, _ := p.junoAdaptor.BidUpdated(*event)
		p.db.BidUpdatedRevert(bidId, amount, treeNonceOld)
	case "OptionsMinted", "OptionsExercised":
		buyerAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		p.db.UpdateOptionBuyerFields(
			buyerAddress,
			roundAddress,
			map[string]interface{}{
				"has_minted": false,
			})
	case "UnusedBidsRefunded":
		buyerAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		p.db.UpdateOptionBuyerFields(
			buyerAddress,
			roundAddress,
			map[string]interface{}{
				"has_refunded": false,
			})

	case "Transfer":
	}

	return nil
}
