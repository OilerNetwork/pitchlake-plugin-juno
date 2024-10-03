package events

import (
	"encoding/hex"
	"fmt"

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
func keccak256(input string) string {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil))
}

// DecodeEventName decodes the event name from the keys of a StarkNet event
func DecodeEventNameRound(eventKey string) (string, error) {
	for _, name := range roundEventNames {
		if keccak256(name) == eventKey {
			return name, nil
		}
	}
	return "", fmt.Errorf("event name not found for key: %s", eventKey)
}

func DecodeEventNameVault(eventKey string) (string, error) {
	for _, name := range vaultEventNames {
		if keccak256(name) == eventKey {
			return name, nil
		}
	}
	return "", fmt.Errorf("event name not found for key: %s", eventKey)
}
