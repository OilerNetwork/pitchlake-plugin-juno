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
	gorm.Model          // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address         int `gorm:"column:address;not null"`
	UnlockedBalance int `gorm:"column:unlocked_balance;not null"`
	LockedBalance   int `gorm:"column:locked_balance;not null"`
	BlockNumber     int `gorm:"column:block_number;not null"`
}

type OptionBuyer struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address    string `gorm:"column:address;not null"`
	//Maybe this is not required and can be directly fetched as a view/index on the bids table
	//Bids       string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	RoundID    int `gorm:"column:round_id;not null"`
	OptionsWon int `gorm:"column:options_won;not null"`
}

type OptionRound struct {
	gorm.Model               // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address           string `gorm:"column:address;not null"`
	RoundID           int    `gorm:"column:round_id;not null"`
	Bids              string `gorm:"column:bids;type:jsonb"` // Store bids as JSON in PostgreSQL
	StartingBlock     int    `gorm:"column:starting_block;not null"`
	EndingBlock       int    `gorm:"column:ending_block;not null"`
	SettlementDate    int    `gorm:"column:settlement_date;not null"`
	StartingLiquidity int    `gorm:"column:starting_liquidity;not null"`
	AvailableOptions  int    `gorm:"column:available_options;not null"`
	SettlementPrice   int    `gorm:"column:settlement_price;not null"`
	StrikePrice       int    `gorm:"column:strike_price;not null"`
	SoldOptions       int    `gorm:"column:sold_options;not null"`
	ClearingPrice     int    `gorm:"column:clearing_price;not null"`
	State             string `gorm:"column:state;not null"`
	Premiums          int    `gorm:"column:premiums;not null"`
}

type VaultState struct {
	gorm.Model                 // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	CurrentRound        int    `gorm:"column:current_round;not null"`
	CurrentRoundAddress string `gorm:"column:current_round_address;not null"`
	UnlockedBalance     int    `gorm:"column:unlocked_balance;not null"`
	LockedBalance       int    `gorm:"column:locked_balance;not null"`
	Address             string `gorm:"column:address;not null"`
}

type LiquidityProviderState struct {
	gorm.Model          // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address         int `gorm:"column:address;not null;primaryKey"`
	UnlockedBalance int `gorm:"column:unlocked_balance;not null"`
	LockedBalance   int `gorm:"column:locked_balance;not null"`
	StashedBalance  int `gorm:"column:stashed_balance;"`
	QueuedBalance   int `gorm:"column:queued_balance;"`
}

type Bid struct {
	gorm.Model        // Adds ID, CreatedAt, UpdatedAt, DeletedAt fields
	Address    string `gorm:"column:address;not null"`
	RoundID    int    `gorm:"column:round_id;not null"`
	BidID      string `gorm:"column:bid_id;not null;unique"`
	Amount     int    `gorm:"column:amount;not null"`
	Price      int    `gorm:"column:price;not null"`
}
