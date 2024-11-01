package models

import (
	"database/sql/driver"
	"fmt"
	"math/big"

	"gorm.io/gorm"
)

type BigInt struct {
	*big.Int
}

// NewBigInt creates a new BigInt from a string
func NewBigInt(s string) *BigInt {
	i := new(big.Int)
	i.SetString(s, 10)
	return &BigInt{i}
}

// Scan implements the sql.Scanner interface for BigInt
func (b *BigInt) Scan(value interface{}) error {
	if b.Int == nil {
		b.Int = new(big.Int)
	}
	switch v := value.(type) {
	case string:
		_, ok := b.Int.SetString(v, 10)
		if !ok {
			return fmt.Errorf("failed to scan BigInt: invalid string %s", v)
		}
	case []byte:
		_, ok := b.Int.SetString(string(v), 10)
		if !ok {
			return fmt.Errorf("failed to scan BigInt: invalid bytes %s", v)
		}
	case int64:
		b.Int.SetInt64(v)
	default:
		return fmt.Errorf("unsupported scan type for BigInt: %T", value)
	}
	return nil
}

// Value implements the driver.Valuer interface for BigInt
func (b BigInt) Value() (driver.Value, error) {
	if b.Int == nil {
		return "0", nil
	}
	return b.Int.String(), nil
}

type Vault struct {
	gorm.Model             // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	BlockNumber     uint64 `gorm:"column:block_number;type:numeric(78,0);not null"`
	UnlockedBalance BigInt `gorm:"column:unlocked_balance;not null"`
	LockedBalance   BigInt `gorm:"column:locked_balance;not null"`
	StashedBalance  BigInt `gorm:"column:stashed_balance;not null"`
}

type LiquidityProvider struct {
	gorm.Model             // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address         string `gorm:"column:address;not null"`
	UnlockedBalance BigInt `gorm:"column:unlocked_balance;not null"`
	LockedBalance   BigInt `gorm:"column:locked_balance;not null"`
	StashedBalance  BigInt `gorm:"column:stashed_balance;not null"`
	BlockNumber     uint64 `gorm:"column:block_number;not null"`
}

type OptionBuyer struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address    string `gorm:"column:address;not null"`
	//Maybe this is not required and can be directly fetched as a view/index on the bids table
	//Bids       string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	RoundAddress     string `gorm:"column:round_id;not null"`
	MintableOptions  BigInt `gorm:"column:mintable_options;"`
	HasMinted        bool   `gorm:"column:has_minted;"`
	RefundableAmount BigInt `gorm:"column:refundable_amount;"`
	HasRefunded      bool   `gorm:"column:has_refunded;"`
}

type OptionRound struct {
	gorm.Model
	VaultAddress      string `gorm:"column:vault_address;"`
	Address           string `gorm:"column:address;not null"`
	RoundID           BigInt `gorm:"column:round_id;not null"`
	Bids              string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	CapLevel          BigInt `gorm:"column:cap_level"`
	StartingBlock     uint64 `gorm:"column:starting_block;not null"`
	EndingBlock       uint64 `gorm:"column:ending_block;not null"`
	SettlementDate    uint64 `gorm:"column:settlement_date;not null"`
	StartingLiquidity BigInt `gorm:"column:starting_liquidity;not null"`
	QueuedLiquidity   BigInt `gorm:"column:queued_liquidity;not null"`
	AvailableOptions  BigInt `gorm:"column:available_options;not null"`
	SettlementPrice   BigInt `gorm:"column:settlement_price;not null"`
	StrikePrice       BigInt `gorm:"column:strike_price;not null"`
	SoldOptions       BigInt `gorm:"column:sold_options;not null"`
	ReservePrice      BigInt `gorm:"column:reserve_price"`
	ClearingPrice     BigInt `gorm:"column:clearing_price"`
	State             string `gorm:"column:state;not null"`
	Premiums          BigInt `gorm:"column:premiums;not null"`
	PayoutPerOption   BigInt `gorm:"column:payout_per_option;"`
}

type VaultState struct {
	gorm.Model
	CurrentRound        BigInt `gorm:"column:current_round;not null"`
	CurrentRoundAddress string `gorm:"column:current_round_address;not null"`
	UnlockedBalance     BigInt `gorm:"column:unlocked_balance;not null"`
	LockedBalance       BigInt `gorm:"column:locked_balance;not null"`
	StashedBalance      BigInt `gorm:"column:stashed_balance;not null"`
	Address             string `gorm:"column:address;not null"`
	LatestBlock         uint64 `gorm:"column:latest_block;"`
}

type LiquidityProviderState struct {
	gorm.Model
	Address         string `gorm:"column:address;not null;primaryKey"`
	UnlockedBalance BigInt `gorm:"column:unlocked_balance;not null"`
	LockedBalance   BigInt `gorm:"column:locked_balance;not null"`
	StashedBalance  BigInt `gorm:"column:stashed_balance;"`
	QueuedBalance   BigInt `gorm:"column:queued_balance;"`
	LatestBlock     uint64 `gorm:"column:latest_block;"`
}

type QueuedLiquidity struct {
	gorm.Model
	Address        string `gorm:"column:address;not null"`
	RoundAddress   BigInt `gorm:"column:round_address;not null"`
	StartingAmount BigInt `gorm:"column:starting_amount;not null"`
	QueuedAmount   BigInt `gorm:"column:amount;not null"`
}
type Bid struct {
	gorm.Model
	BuyerAddress string `gorm:"column:buyer_address;not null"`
	RoundAddress string `gorm:"column:round_address;not null"`
	BidID        string `gorm:"column:bid_id;not null"`
	TreeNonce    uint64 `gorm:"column:tree_nonce;not null"`
	Amount       BigInt `gorm:"column:amount;not null"`
	Price        BigInt `gorm:"column:price;not null"`
}

func (VaultState) TableName() string {
	return "VaultStates"
}
func (LiquidityProviderState) TableName() string {
	return "Liquidity_Providers"
}

func (OptionRound) TableName() string {
	return "Option_Rounds"
}

func (QueuedLiquidity) TableName() string {
	return "Queued_Liquidity"
}

func (Bid) TableName() string {
	return "Bids"
}

func (OptionBuyer) TableName() string {
	return "Option_Buyers"
}
