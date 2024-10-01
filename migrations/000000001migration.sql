-- +migrate Up
CREATE TABLE vaults (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    block_number BIGINT NOT NULL,
    unlocked_balance BIGINT NOT NULL,
    locked_balance BIGINT NOT NULL,
    stashed_balance BIGINT NOT NULL
);

CREATE TABLE liquidity_providers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    address TEXT NOT NULL,
    unlocked_balance BIGINT NOT NULL,
    locked_balance BIGINT NOT NULL,
    stashed_balance BIGINT NOT NULL,
    block_number BIGINT NOT NULL
);

CREATE TABLE option_buyers (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    address TEXT NOT NULL,
    round_id BIGINT NOT NULL,
    tokenizable_options BIGINT,
    refundable_balance BIGINT
);

CREATE TABLE option_rounds (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    address TEXT NOT NULL,
    round_id BIGINT NOT NULL,
    bids JSONB,
    cap_level BIGINT,
    starting_block BIGINT NOT NULL,
    ending_block BIGINT NOT NULL,
    settlement_date BIGINT NOT NULL,
    starting_liquidity BIGINT NOT NULL,
    queued_liquidity BIGINT NOT NULL,
    available_options BIGINT NOT NULL,
    settlement_price BIGINT NOT NULL,
    strike_price BIGINT NOT NULL,
    sold_options BIGINT NOT NULL,
    clearing_price BIGINT NOT NULL,
    state TEXT NOT NULL,
    premiums BIGINT NOT NULL
);

-- +migrate Down
DROP TABLE IF EXISTS vaults;
DROP TABLE IF EXISTS liquidity_providers;
DROP TABLE IF EXISTS option_buyers;
DROP TABLE IF EXISTS option_rounds;