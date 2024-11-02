package events

import (
	"fmt"
	"math/big"

	"golang.org/x/crypto/sha3"
)

// Known event names in your contract
var vaultEventNames = []string{
	"Deposit",
	"Withdrawal",
	"WithdrawalQueued",
	"QueuedLiquidityCollected",
	"OptionRoundDeployed",
}
var roundEventNames = []string{
	"AuctionStarted",
	"AuctionEnded",
	"OptionRoundSettled",
	"BidAccepted",
	"BidUpdated",
	"OptionsMinted",
	"UnusedBidsRefunded",
	"OptionsExercised",
}

// keccak256 function to hash the event name
func Keccak256(eventName string) string {
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

// DecodeEventName decodes the event name from the keys of a StarkNet event
func DecodeEventNameRound(eventKey string) (string, error) {
	for _, name := range roundEventNames {
		if Keccak256(name) == eventKey {
			return name, nil
		}
	}
	return "", fmt.Errorf("event name not found for key: %s", eventKey)
}

func DecodeEventNameVault(eventKey string) (string, error) {
	for _, name := range vaultEventNames {
		if Keccak256(name) == eventKey {
			return name, nil
		}
	}
	return "", fmt.Errorf("event name not found for key: %s", eventKey)
}
