
CREATE TABLE "Liquidity_Providers"
(
    address character varying COLLATE pg_catalog."default" NOT NULL,
    stashed_balance numeric(78,0),
    locked_balance numeric(78,0),
    unlocked_balance numeric(78,0),
    latest_block numeric(78,0),
    CONSTRAINT "Liquidity_Providers_pkey" PRIMARY KEY (address)
);

CREATE TABLE "Option_Rounds"
(
    address character varying COLLATE pg_catalog."default" NOT NULL,
    available_options numeric(78,0) DEFAULT 0,
    clearing_price numeric(78,0),
    settlement_price numeric(78,0),
    reserve_price numeric(78,0),
    strike_price numeric(78,0),
    sold_options numeric(78,0),
    state character varying(10) COLLATE pg_catalog."default",
    premiums numeric(78,0),
    vault_address character varying(67) COLLATE pg_catalog."default",
    round_id numeric(78,0),
    cap_level numeric(78,0),
    starting_liquidity numeric(78,0),
    queued_liquidity numeric(78,0),
    payout_per_option numeric(78,0),
    start_date numeric(78,0),
    end_date numeric(78,0),
    settlement_date numeric(78,0),
    CONSTRAINT "Option_Rounds_pkey" PRIMARY KEY (address)
);


ALTER TABLE "Option_Rounds"
    OWNER to pitchlake_user;


-- Table: public.Queued_Liquidity

CREATE TABLE "Queued_Liquidity"
(
    address character varying(67) COLLATE pg_catalog."default" NOT NULL,
    starting_amount numeric(78,0) NOT NULL,
    queued_amount numeric(78,0) NOT NULL,
    round_address character varying(67) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT lp_round_address PRIMARY KEY (address, round_address),
    CONSTRAINT lp_address FOREIGN KEY (address)
        REFERENCES public."Liquidity_Providers" (address) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID,
    CONSTRAINT round_address FOREIGN KEY (round_address)
        REFERENCES public."Option_Rounds" (address) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);

ALTER TABLE IF EXISTS "Queued_Liquidity"
    OWNER to pitchlake_user;


-- Table: public.VaultStates

CREATE TABLE "VaultStates"
(
    unlocked_balance numeric(78,0),
    locked_balance numeric(78,0),
    current_round_address character varying(67) COLLATE pg_catalog."default",
    stashed_balance numeric(78,0),
    address character varying(67) COLLATE pg_catalog."default" NOT NULL,
    latest_block numeric(78,0),
    current_round numeric(78,0),
    CONSTRAINT "VaultState_pkey" PRIMARY KEY (address)
);


ALTER TABLE IF EXISTS "VaultStates"
    OWNER to pitchlake_user;


-- Table: public.Option_Buyers


CREATE TABLE "Option_Buyers"
(
    address character varying COLLATE pg_catalog."default" NOT NULL,
    round_address character varying COLLATE pg_catalog."default" NOT NULL,
    has_minted boolean NOT NULL DEFAULT false,
    has_refunded boolean NOT NULL DEFAULT false,
    mintable_options numeric(78,0),
    refundable_options numeric(78,0),
    CONSTRAINT buyer_round PRIMARY KEY (address, round_address)
);


ALTER TABLE "Option_Buyers"
    OWNER to pitchlake_user;


-- Table: public.Bids


CREATE TABLE "Bids"
(
    buyer_address character varying(67) COLLATE pg_catalog."default",
    round_address character varying(67) COLLATE pg_catalog."default" NOT NULL,
    bid_id character varying(67) COLLATE pg_catalog."default" NOT NULL,
    tree_nonce numeric(78,0),
    amount numeric(78,0),
    price numeric(78,0),
    CONSTRAINT round_address_bid_id PRIMARY KEY (round_address, bid_id)
);


ALTER TABLE "Bids"
    OWNER to pitchlake_user;









