package adaptors

import (
	"testing"

	"github.com/NethermindEth/juno/core"
	"github.com/stretchr/testify/assert"
)

func TestPricingDataSet(t *testing.T) {
	// Prepare mock data

	event := core.Event{
		// Data: []*felt.Felt{
		// 	felt.NewFelt("12345"),  // event.Data[0]
		// 	core.NewFeltFromInt(67890),  // event.Data[1]
		// 	core.NewFeltFromInt(112233), // event.Data[2]
		// 	core.NewFeltFromInt(445566), // event.Data[3]
		// 	core.NewFeltFromInt(778899), // event.Data[4]
		// },
	}

	// Instantiate JunoAdaptor
	junoAdaptor := &JunoAdaptor{}

	// Call the method
	strikePrice, capLevel, reservePrice := junoAdaptor.PricingDataSet(event)

	// Check the output using the expected values
	expectedStrikePrice := CombineFeltToBigInt(event.Data[1].Bytes(), event.Data[0].Bytes())
	expectedCapLevel := FeltToBigInt(event.Data[2].Bytes())
	expectedReservePrice := CombineFeltToBigInt(event.Data[4].Bytes(), event.Data[3].Bytes())

	assert.Equal(t, expectedStrikePrice, strikePrice, "strikePrice should match the expected value")
	assert.Equal(t, expectedCapLevel, capLevel, "capLevel should match the expected value")
	assert.Equal(t, expectedReservePrice, reservePrice, "reservePrice should match the expected value")
}
