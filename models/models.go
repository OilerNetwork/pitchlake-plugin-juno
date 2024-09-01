package models

import (
	"gorm.io/gorm"
)

type Vault struct {
	gorm.Model          // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	BlockNumber     int `gorm:"column:block_number;not null;primaryKey"`
	UnlockedBalance int `gorm:"column:unlocked_balance;not null"`
	LockedBalance   int `gorm:"column:locked_balance;not null"`
}

type LiquidityProvider struct {
	gorm.Model             // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address         string `gorm:"column:address;not null"`
	UnlockedBalance uint64 `gorm:"column:unlocked_balance;not null"`
	LockedBalance   uint64 `gorm:"column:locked_balance;not null"`
	StashedBalance  uint64 `gorm:"column:stashed_balance;not null"`
	BlockNumber     uint64 `gorm:"column:block_number;not null"`
}

type OptionBuyer struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address    string `gorm:"column:address;not null"`
	//Maybe this is not required and can be directly fetched as a view/index on the bids table
	//Bids       string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	RoundID            int    `gorm:"column:round_id;not null"`
	TokenizableOptions int    `gorm:"column:tokenizable_options;"`
	RefundableBalance  uint64 `gorm:"column:refundable_balance;"`
}

type OptionRound struct {
	gorm.Model
	Address           string `gorm:"column:address;not null"`
	RoundID           uint64 `gorm:"column:round_id;not null"`
	Bids              string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	CapLevel          uint64 `gorm:"column:cap_level"`
	StartingBlock     uint64 `gorm:"column:starting_block;not null"`
	EndingBlock       uint64 `gorm:"column:ending_block;not null"`
	SettlementDate    uint64 `gorm:"column:settlement_date;not null"`
	StartingLiquidity uint64 `gorm:"column:starting_liquidity;not null"`
	QueuedLiquidity   uint64 `gorm:"column:queued_liquidity;not null"`
	AvailableOptions  uint64 `gorm:"column:available_options;not null"`
	SettlementPrice   uint64 `gorm:"column:settlement_price;not null"`
	StrikePrice       uint64 `gorm:"column:strike_price;not null"`
	SoldOptions       uint64 `gorm:"column:sold_options;not null"`
	ClearingPrice     uint64 `gorm:"column:clearing_price;not null"`
	State             string `gorm:"column:state;not null"`
	Premiums          uint64 `gorm:"column:premiums;not null"`
	PayoutPerOption   uint64 `gorm:"column:payout_per_option;"`
}

type VaultState struct {
	gorm.Model
	CurrentRound        uint64 `gorm:"column:current_round;not null"`
	CurrentRoundAddress string `gorm:"column:current_round_address;not null"`
	UnlockedBalance     uint64 `gorm:"column:unlocked_balance;not null"`
	LockedBalance       uint64 `gorm:"column:locked_balance;not null"`
	Address             string `gorm:"column:address;not null"`
}

type LiquidityProviderState struct {
	gorm.Model
	Address         string `gorm:"column:address;not null;primaryKey"`
	UnlockedBalance int    `gorm:"column:unlocked_balance;not null"`
	LockedBalance   int    `gorm:"column:locked_balance;not null"`
	StashedBalance  int    `gorm:"column:stashed_balance;"`
	QueuedBalance   int    `gorm:"column:queued_balance;"`
}

type QueuedLiquidity struct {
	gorm.Model
	Address        string `gorm:"column:address;not null"`
	RoundID        uint64 `gorm:"column:round_id;not null"`
	StartingAmount uint64 `gorm:"column:starting_amount;not null"`
	QueuedAmount   uint64 `gorm:"column:amount;not null"`
}
type Bid struct {
	gorm.Model
	Address   string `gorm:"column:address;not null"`
	RoundID   uint64 `gorm:"column:round_id;not null"`
	BidID     string `gorm:"column:bid_id;not null"`
	TreeNonce uint64 `gorm:"column:tree_nonce;not null"`
	Amount    uint64 `gorm:"column:amount;not null"`
	Price     uint64 `gorm:"column:price;not null"`
}

type Position struct {
}
