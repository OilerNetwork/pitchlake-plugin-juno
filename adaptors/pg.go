package adaptors

import (
	"junoplugin/models"
	"math/big"

	"github.com/NethermindEth/juno/core"
)

type JunoAdaptor struct {
}

func (p *JunoAdaptor) DepositOrWithdraw(event core.Event) (string, models.BigInt, models.BigInt) {
	lpAddress := FeltToHexString(event.Keys[0].Bytes())
	lpUnlocked := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[3].Bytes())
	vaultUnlocked := CombineFeltToBigInt(event.Data[4].Bytes(), event.Data[5].Bytes())
	return lpAddress, lpUnlocked, vaultUnlocked
}

func (p *JunoAdaptor) RoundDeployed(event core.Event) models.OptionRound {

	vaultAddress :=
		event.From.String()
	roundId := CombineFeltToBigInt(event.Data[0].Bytes(), event.Data[1].Bytes())
	roundAddress := FeltToHexString(event.Data[2].Bytes())
	startingBlock := event.Data[3].Uint64()
	endingBlock := event.Data[4].Uint64()
	settlementDate := event.Data[5].Uint64()
	strikePrice := CombineFeltToBigInt(event.Data[6].Bytes(), event.Data[7].Bytes())
	capLevel := FeltToBigInt(event.Data[8].Bytes())
	reservePrice := CombineFeltToBigInt(event.Data[9].Bytes(), event.Data[10].Bytes())
	optionRound := models.OptionRound{
		RoundID:        roundId,
		Address:        roundAddress,
		VaultAddress:   vaultAddress,
		StartingBlock:  startingBlock,
		EndingBlock:    endingBlock,
		SettlementDate: settlementDate,
		StrikePrice:    strikePrice,
		CapLevel:       capLevel,
		ReservePrice:   reservePrice,
		State:          "Open",
	}
	return optionRound

}

func (p *JunoAdaptor) AuctionStarted(event core.Event) (models.BigInt, models.BigInt) {

	availableOptions := CombineFeltToBigInt(event.Data[0].Bytes(), event.Data[1].Bytes())
	startingLiquidity := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[3].Bytes())
	return availableOptions, startingLiquidity
}

func (p *JunoAdaptor) AuctionEnded(event core.Event) (models.BigInt, models.BigInt, models.BigInt, models.BigInt) {
	optionsSold := CombineFeltToBigInt(event.Data[1].Bytes(), event.Data[0].Bytes())
	clearingPrice := CombineFeltToBigInt(event.Data[3].Bytes(), event.Data[2].Bytes())
	clearingNonce := CombineFeltToBigInt(event.Data[5].Bytes(), event.Data[4].Bytes())
	premiums := models.BigInt{Int: new(big.Int).Mul(optionsSold.Int, clearingPrice.Int)}

	return optionsSold, clearingPrice, clearingNonce, premiums
}

func (p *JunoAdaptor) RoundSettled(event core.Event) (models.BigInt, models.BigInt) {
	totalPayout := CombineFeltToBigInt(event.Data[0].Bytes(), event.Data[1].Bytes())
	settlementPrice := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[3].Bytes())
	return totalPayout, settlementPrice
}

func (p *JunoAdaptor) BidAccepted(event core.Event) models.Bid {
	bidAmount := CombineFeltToBigInt(event.Data[2].Bytes(), event.Data[1].Bytes())
	bidPrice := CombineFeltToBigInt(event.Data[4].Bytes(), event.Data[3].Bytes())
	treeNonce := event.Data[5].Uint64()

	var bid models.Bid
	bid.BuyerAddress = event.Keys[0].String()
	bid.BidID = event.Data[0].String()
	bid.RoundAddress = event.From.String()
	bid.Amount = bidAmount
	bid.Price = bidPrice
	bid.TreeNonce = treeNonce
	return bid
}

func (p *JunoAdaptor) BidUpdated(event core.Event) (string, models.BigInt, uint64, uint64) {
	bidId := event.Data[0].String()
	amount := CombineFeltToBigInt(event.Data[1].Bytes(), event.Data[2].Bytes())
	treeNonceOld := event.Data[3].Uint64()
	treeNonceNew := event.Data[4].Uint64()
	return bidId, amount, treeNonceOld, treeNonceNew
}
