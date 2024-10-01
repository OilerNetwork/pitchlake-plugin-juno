package models

import (
	"database/sql/driver"
	"fmt"
	"math/big"

	"gorm.io/gorm"
)

type BigInt struct {
	big.Int
}

// Scan implements the sql.Scanner interface for BigInt
func (b *BigInt) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		b.SetString(v, 10)
	case []byte:
		b.SetString(string(v), 10)
	case int64:
		b.SetInt64(v)
	default:
		return fmt.Errorf("unsupported scan type for BigInt: %T", value)
	}
	return nil
}

// Value implements the driver.Valuer interface for BigInt
func (b BigInt) Value() (driver.Value, error) {
	return b.String(), nil
}

type Vault struct {
	gorm.Model             // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	BlockNumber     BigInt `gorm:"column:block_number;type:numeric(78,0);not null"`
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
	BlockNumber     BigInt `gorm:"column:block_number;not null"`
}

type OptionBuyer struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address    string `gorm:"column:address;not null"`
	//Maybe this is not required and can be directly fetched as a view/index on the bids table
	//Bids       string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	RoundID            BigInt `gorm:"column:round_id;not null"`
	TokenizableOptions BigInt `gorm:"column:tokenizable_options;"`
	RefundableBalance  BigInt `gorm:"column:refundable_balance;"`
}

type OptionRound struct {
	gorm.Model
	Address           string `gorm:"column:address;not null"`
	RoundID           BigInt `gorm:"column:round_id;not null"`
	Bids              string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	CapLevel          BigInt `gorm:"column:cap_level"`
	StartingBlock     BigInt `gorm:"column:starting_block;not null"`
	EndingBlock       BigInt `gorm:"column:ending_block;not null"`
	SettlementDate    BigInt `gorm:"column:settlement_date;not null"`
	StartingLiquidity BigInt `gorm:"column:starting_liquidity;not null"`
	QueuedLiquidity   BigInt `gorm:"column:queued_liquidity;not null"`
	AvailableOptions  BigInt `gorm:"column:available_options;not null"`
	SettlementPrice   BigInt `gorm:"column:settlement_price;not null"`
	StrikePrice       BigInt `gorm:"column:strike_price;not null"`
	SoldOptions       BigInt `gorm:"column:sold_options;not null"`
	ClearingPrice     BigInt `gorm:"column:clearing_price;not null"`
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
	LastBlock           BigInt `gorm:"column:last_block;"`
}

type LiquidityProviderState struct {
	gorm.Model
	Address         string `gorm:"column:address;not null;primaryKey"`
	UnlockedBalance BigInt `gorm:"column:unlocked_balance;not null"`
	LockedBalance   BigInt `gorm:"column:locked_balance;not null"`
	StashedBalance  BigInt `gorm:"column:stashed_balance;"`
	QueuedBalance   BigInt `gorm:"column:queued_balance;"`
	LastBlock       BigInt `gorm:"column:last_block;"`
}

type QueuedLiquidity struct {
	gorm.Model
	Address        string `gorm:"column:address;not null"`
	RoundID        BigInt `gorm:"column:round_id;not null"`
	StartingAmount BigInt `gorm:"column:starting_amount;not null"`
	QueuedAmount   BigInt `gorm:"column:amount;not null"`
}
type Bid struct {
	gorm.Model
	Address   string `gorm:"column:address;not null"`
	RoundID   BigInt `gorm:"column:round_id;not null"`
	BidID     string `gorm:"column:bid_id;not null"`
	TreeNonce BigInt `gorm:"column:tree_nonce;not null"`
	Amount    BigInt `gorm:"column:amount;not null"`
	Price     BigInt `gorm:"column:price;not null"`
}

type Position struct {
}
