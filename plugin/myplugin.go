package main

import (
	"junoplugin/adaptors"
	"junoplugin/db"
	"junoplugin/models"
	"log"
	"os"
	"strconv"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	junoplugin "github.com/NethermindEth/juno/plugin"
)

// Todo: push this stuff to a config file / cmd line

//go:generate go build -buildmode=plugin -o ../../build/plugin.so ./example.go
type pitchlakePlugin struct {
	vaultHash         string
	vaultAddressesMap map[string]struct{}
	roundAddressesMap map[string]struct{}
	deployer          string
	udcAddress        string
	db                *db.DB
	log               *log.Logger
	junoAdaptor       *adaptors.JunoAdaptor
	cursor            uint64
}

// Important: "JunoPluginInstance" needs to be exported for Juno to load the plugin correctly
var JunoPluginInstance = pitchlakePlugin{}

// Ensure the plugin and Juno client follow the same interface
var _ junoplugin.JunoPlugin = (*pitchlakePlugin)(nil)

func (p *pitchlakePlugin) Init() error {
	dbUrl := os.Getenv("DB_URL")
	udcAddress := os.Getenv("UDC_ADDRESS")
	p.udcAddress = udcAddress
	p.vaultAddressesMap = make(map[string]struct{})
	p.roundAddressesMap = make(map[string]struct{})
	dbClient, err := db.Init(dbUrl)
	if err != nil {
		return err
	}
	p.db = dbClient
	vaultAddresses, err := p.db.GetVaultAddresses()
	if err != nil {
		return err
	}

	//Map
	for _, vaultAddress := range vaultAddresses {
		p.vaultAddressesMap[vaultAddress] = struct{}{}
		roundAddresses, err := p.db.GetRoundAddressess(vaultAddress)
		if err != nil {
			return err
		}
		//Round Address Map
		for _, roundAddress := range *roundAddresses {
			p.roundAddressesMap[roundAddress] = struct{}{}
		}
	}

	p.junoAdaptor = &adaptors.JunoAdaptor{}
	p.vaultHash = os.Getenv("VAULT_HASH")
	p.deployer = os.Getenv("DEPLOYER")
	cursor := os.Getenv("CURSOR")
	if cursor != "" {
		p.cursor, err = strconv.ParseUint(cursor, 10, 64)
		if err != nil {
			return err
		}
	}
	p.log = log.Default()

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
	if block.Number < p.cursor {
		log.Printf("Pre-cursor block")
		return nil
	}

	var err error
	for _, receipt := range block.Receipts {
		for i, event := range receipt.Events {
			fromAddress := event.From.String()

			if fromAddress == p.udcAddress {
				err = p.processUDC(receipt.Events, event, i, block.Number, block.Timestamp)
			} else {

				//HashMap processing
				if _, exists := p.vaultAddressesMap[fromAddress]; exists {
					err = p.processVaultEvent(fromAddress, event, block.Number, block.Timestamp)
				} else if _, exists := p.roundAddressesMap[fromAddress]; exists {
					err = p.processRoundEvent(fromAddress, event, block.Number)
				}
				if err != nil {
					log.Fatal(err)
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
	p.db.Begin()
	length := len(from.Block.Receipts)
	var err error
	for i := length - 1; i >= 0; i-- {
		receipt := from.Block.Receipts[i]
		for _, event := range receipt.Events {

			fromAddress := event.From.String()

			//HashMap
			if _, exists := p.vaultAddressesMap[fromAddress]; exists {
				err = p.revertVaultEvent(fromAddress, event, from.Block.Number)
			} else if _, exists := p.roundAddressesMap[fromAddress]; exists {
				err = p.revertRoundEvent(fromAddress, event, from.Block.Number)
			}
			if err != nil {
				log.Fatal(err)
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
	timestamp uint64,
) error {

	eventHash := adaptors.Keccak256("ContractDeployed")
	if eventHash == event.Keys[0].String() {
		address := adaptors.FeltToHexString(event.Data[0].Bytes())
		deployer := adaptors.FeltToHexString(event.Data[1].Bytes())
		classHash := adaptors.FeltToHexString(event.Data[3].Bytes())

		//ClassHash and deployer filter, may use other filters here

		if classHash == p.vaultHash && deployer == p.deployer {
			fossilClientAddress, ethAddress, optionRoundClassHash, alpha, strikeLevel, roundTransitionDuration, auctionDuration, roundDuration := p.junoAdaptor.ContractDeployed(*event)
			p.vaultAddressesMap[address] = struct{}{}
			vault := models.VaultState{
				CurrentRound:          *models.NewBigInt("1"),
				UnlockedBalance:       *models.NewBigInt("0"),
				LockedBalance:         *models.NewBigInt("0"),
				StashedBalance:        *models.NewBigInt("0"),
				Address:               address,
				LatestBlock:           blockNumber,
				FossilClientAddress:   fossilClientAddress,
				EthAddress:            ethAddress,
				OptionRoundClassHash:  optionRoundClassHash,
				Alpha:                 alpha,
				StrikeLevel:           strikeLevel,
				RoundTransitionPeriod: roundTransitionDuration,
				AuctionDuration:       auctionDuration,
				RoundDuration:         roundDuration,
				DeploymentDate:        timestamp,
			}
			if err := p.db.CreateVault(&vault); err != nil {
				return err
			}
			log.Printf("index %v", index)
			if err := p.processVaultEvent(address, events[index-1], blockNumber, timestamp); err != nil {
				return err
			}
		}

	}
	return nil
}

func (p *pitchlakePlugin) processVaultEvent(
	vaultAddress string,
	event *core.Event,
	blockNumber uint64,
	timestamp uint64,
) error {

	eventName, err := adaptors.DecodeEventNameVault(event.Keys[0].String())
	if err != nil {
		log.Printf("Unknown Event")
		return nil
	}
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
		optionRound.DeploymentDate = timestamp
		err = p.db.RoundDeployedIndex(optionRound)
		p.roundAddressesMap[optionRound.Address] = struct{}{}
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
	prevStateOptionRound, err := p.db.GetOptionRoundByAddress(roundAddress)
	if err != nil {
		return err
	}

	eventName, err := adaptors.DecodeEventNameRound(event.Keys[0].String())
	if err != nil {
		return nil
	}
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
			*prevStateOptionRound,
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
			*prevStateOptionRound,
			roundAddress,
			blockNumber,
			settlementPrice,
			prevStateOptionRound.SoldOptions,
			payoutPerOption,
		); err != nil {
			return err
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
		return err
	}
	switch eventName {
	case "Deposit", "Withdraw",
		"StashWithdrawn": //Add withdraw queue

		lpAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		err = p.db.DepositOrWithdrawRevert(vaultAddress, lpAddress, blockNumber)
	case "WithdrawalQueued":
		lpAddress,
			bps,
			roundId,
			accountQueuedBefore,
			accountQueuedNow,
			vaultQueuedNow := p.junoAdaptor.WithdrawalQueued(*event)

		err = p.db.WithdrawalQueuedRevertIndex(
			lpAddress,
			vaultAddress,
			roundId,
			bps,
			accountQueuedBefore,
			accountQueuedNow,
			vaultQueuedNow,
			blockNumber,
		)
	case "OptionRoundDeployed":
		roundAddress := adaptors.FeltToHexString(event.Data[2].Bytes())
		err = p.db.DeleteOptionRound(roundAddress)
	}
	if err != nil {
		return err
	}

	return nil
}

func (p *pitchlakePlugin) revertRoundEvent(roundAddress string, event *core.Event, blockNumber uint64) error {
	eventName, err := adaptors.DecodeEventNameRound(event.Keys[0].String())
	if err != nil {
		return err
	}
	prevStateOptionRound, err := p.db.GetOptionRoundByAddress(roundAddress)
	if err != nil {
		return err
	}
	switch eventName {
	case "AuctionStarted":
		err = p.db.AuctionStartedRevert(prevStateOptionRound.VaultAddress, roundAddress, blockNumber)
	case "AuctionEnded":
		err = p.db.AuctionEndedRevert(prevStateOptionRound.VaultAddress, roundAddress, blockNumber)

	case "OptionRoundSettled":
		err = p.db.RoundSettledRevert(prevStateOptionRound.VaultAddress, roundAddress, blockNumber)
	case "BidAccepted":
		id := event.Data[1].String()
		err = p.db.BidAcceptedRevert(id, roundAddress)
	case "BidUpdated":
		bidId, amount, treeNonceOld, _ := p.junoAdaptor.BidUpdated(*event)
		p.db.BidUpdatedRevert(bidId, amount, treeNonceOld)
	case "OptionsMinted", "OptionsExercised":
		buyerAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		err = p.db.UpdateOptionBuyerFields(
			buyerAddress,
			roundAddress,
			map[string]interface{}{
				"has_minted": false,
			})
	case "UnusedBidsRefunded":
		buyerAddress := adaptors.FeltToHexString(event.Keys[1].Bytes())
		err = p.db.UpdateOptionBuyerFields(
			buyerAddress,
			roundAddress,
			map[string]interface{}{
				"has_refunded": false,
			})

	case "Transfer":
	}
	if err != nil {
		log.Fatal(err)
	}
	return nil
}
