package models

// Vault represents the vault table
type Vault struct {
	BlockNumber     int `db:"block_number"`
	UnlockedBalance int `db:"unlocked_balance"`
	LockedBalance   int `db:"locked_balance"`
}

// LiquidityProvider represents the liquidity_providers table
type LiquidityProvider struct {
	Address         int `db:"address"`
	UnlockedBalance int `db:"unlocked_balance"`
	LockedBalance   int `db:"locked_balance"`
	BlockNumber     int `db:"block_number"`
}

// OptionBuyer represents the option_buyers table
type OptionBuyer struct {
	Address    string `db:"address"`
	Bids       []int  `db:"bids"` // Assuming bids are stored as integer array
	RoundID    int    `db:"round_id"`
	OptionsWon int    `db:"options_won"`
}

// OptionRound represents the option_rounds table
type OptionRound struct {
	Address           string `db:"address"`
	RoundID           int    `db:"round_id"`
	Bids              []int  `db:"bids"` // Assuming bids are stored as integer array
	StartingBlock     int    `db:"starting_block"`
	EndingBlock       int    `db:"ending_block"`
	SettlementDate    int    `db:"settlement_date"`
	StartingLiquidity int    `db:"starting_liquidity"`
	AvailableOptions  int    `db:"available_options"`
	SettlementPrice   int    `db:"settlement_price"`
	StrikePrice       int    `db:"strike_price"`
	SoldOptions       int    `db:"sold_options"`
	ClearingPrice     int    `db:"clearing_price"`
	State             string `db:"state"`
	Premiums          int    `db:"premiums"`
}

// VaultState represents the vault_state table
type VaultState struct {
	CurrentRound        int    `db:"current_round"`
	CurrentRoundAddress string `db:"current_round_address"`
	UnlockedBalance     int    `db:"unlocked_balance"`
	LockedBalance       int    `db:"locked_balance"`
	Address             string `db:"address"`
}

// Bid represents the bids table
type Bid struct {
	RoundID int    `db:"round_id"`
	BidID   string `db:"bid_id"`
	Amount  int    `db:"amount"`
	Price   int    `db:"price"`
}
