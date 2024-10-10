-- -- +migrate Up
-- CREATE TABLE vaults (
--     id BIGSERIAL PRIMARY KEY,
--     created_at TIMESTAMPTZ,
--     updated_at TIMESTAMPTZ,
--     deleted_at TIMESTAMPTZ,
--     block_number BIGINT NOT NULL,
--     unlocked_balance BIGINT NOT NULL,
--     locked_balance BIGINT NOT NULL,
--     stashed_balance BIGINT NOT NULL
-- );

-- CREATE TABLE liquidity_providers (
--     id BIGSERIAL PRIMARY KEY,
--     created_at TIMESTAMPTZ,
--     updated_at TIMESTAMPTZ,
--     deleted_at TIMESTAMPTZ,
--     address TEXT NOT NULL,
--     unlocked_balance BIGINT NOT NULL,
--     locked_balance BIGINT NOT NULL,
--     stashed_balance BIGINT NOT NULL,
--     block_number BIGINT NOT NULL
-- );

-- CREATE TABLE option_buyers (
--     id BIGSERIAL PRIMARY KEY,
--     created_at TIMESTAMPTZ,
--     updated_at TIMESTAMPTZ,
--     deleted_at TIMESTAMPTZ,
--     address TEXT NOT NULL,
--     round_id BIGINT NOT NULL,
--     tokenizable_options BIGINT,
--     refundable_balance BIGINT
-- );

-- CREATE TABLE option_rounds (
--     id BIGSERIAL PRIMARY KEY,
--     created_at TIMESTAMPTZ,
--     updated_at TIMESTAMPTZ,
--     deleted_at TIMESTAMPTZ,
--     address TEXT NOT NULL,
--     round_id BIGINT NOT NULL,
--     bids JSONB,
--     cap_level BIGINT,
--     starting_block BIGINT NOT NULL,
--     ending_block BIGINT NOT NULL,
--     settlement_date BIGINT NOT NULL,
--     starting_liquidity BIGINT NOT NULL,
--     queued_liquidity BIGINT NOT NULL,
--     available_options BIGINT NOT NULL,
--     settlement_price BIGINT NOT NULL,
--     strike_price BIGINT NOT NULL,
--     sold_options BIGINT NOT NULL,
--     clearing_price BIGINT NOT NULL,
--     state TEXT NOT NULL,
--     premiums BIGINT NOT NULL
-- );

-- -- +migrate Down
-- DROP TABLE IF EXISTS vaults;
-- DROP TABLE IF EXISTS liquidity_providers;
-- DROP TABLE IF EXISTS option_buyers;
-- DROP TABLE IF EXISTS option_rounds;




-- Table: public.Liquidity_Providers

-- DROP TABLE IF EXISTS public."Liquidity_Providers";

CREATE TABLE IF NOT EXISTS public."Liquidity_Providers"
(
    address character varying COLLATE pg_catalog."default" NOT NULL,
    stashed_balance numeric(78,0),
    locked_balance numeric(78,0),
    unlocked_balance numeric(78,0),
    latest_block numeric(78,0),
    CONSTRAINT "Liquidity_Providers_pkey" PRIMARY KEY (address)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public."Liquidity_Providers"
    OWNER to postgres;


-- FUNCTION: public.notify_lp_update()

-- DROP FUNCTION IF EXISTS public.notify_lp_update();

CREATE OR REPLACE FUNCTION public.notify_lp_update()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE NOT LEAKPROOF
AS $BODY$
DECLARE
    updated_row JSON;
BEGIN
    -- Convert the NEW row to JSON (if you want to notify the entire row)
    updated_row := row_to_json(NEW);

    -- Notify with the 'lp_update' channel and send the updated row as JSON
    PERFORM pg_notify('lp_update', updated_row::text);

    RETURN NEW;
END;
$BODY$;

ALTER FUNCTION public.notify_lp_update()
    OWNER TO postgres;


-- Trigger: lp_update

-- DROP TRIGGER IF EXISTS lp_update ON public."Liquidity_Providers";

CREATE OR REPLACE TRIGGER lp_update
    AFTER UPDATE 
    ON public."Liquidity_Providers"
    FOR EACH ROW
    EXECUTE FUNCTION public.notify_lp_update();



-- Table: public.Option_Rounds

-- DROP TABLE IF EXISTS public."Option_Rounds";

CREATE TABLE IF NOT EXISTS public."Option_Rounds"
(
    address character varying COLLATE pg_catalog."default" NOT NULL,
    available_options numeric(78,0) DEFAULT 0,
    clearing_price numeric(78,0),
    settlement_price numeric(78,0),
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
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public."Option_Rounds"
    OWNER to postgres;


-- FUNCTION: public.notify_or_update()

-- DROP FUNCTION IF EXISTS public.notify_or_update();

CREATE OR REPLACE FUNCTION public.notify_or_update()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE NOT LEAKPROOF
AS $BODY$
DECLARE
    updated_row JSON;
BEGIN
    -- Convert the NEW row to JSON (if you want to notify the entire row)
    updated_row := row_to_json(NEW);

    -- Notify with the 'or_update' channel and send the updated row as JSON
    PERFORM pg_notify('or_update', updated_row::text);

    RETURN NEW;
END;
$BODY$;

ALTER FUNCTION public.notify_or_update()
    OWNER TO postgres;


-- Trigger: or_update

-- DROP TRIGGER IF EXISTS or_update ON public."Option_Rounds";

CREATE OR REPLACE TRIGGER or_update
    AFTER UPDATE 
    ON public."Option_Rounds"
    FOR EACH ROW
    EXECUTE FUNCTION public.notify_or_update();



-- Table: public.Queued_Liquidity

-- DROP TABLE IF EXISTS public."Queued_Liquidity";

CREATE TABLE IF NOT EXISTS public."Queued_Liquidity"
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
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public."Queued_Liquidity"
    OWNER to postgres;


-- Table: public.VaultStates

-- DROP TABLE IF EXISTS public."VaultStates";

CREATE TABLE IF NOT EXISTS public."VaultStates"
(
    unlocked_balance numeric(78,0),
    locked_balance numeric(78,0),
    current_round_address character varying(67) COLLATE pg_catalog."default",
    stashed_balance numeric(78,0),
    address character varying(67) COLLATE pg_catalog."default" NOT NULL,
    latest_block numeric(78,0),
    current_round numeric(78,0),
    CONSTRAINT "VaultState_pkey" PRIMARY KEY (address)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public."VaultStates"
    OWNER to postgres;



-- FUNCTION: public.notify_vault_update()

-- DROP FUNCTION IF EXISTS public.notify_vault_update();

CREATE OR REPLACE FUNCTION public.notify_vault_update()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE NOT LEAKPROOF
AS $BODY$
DECLARE
    updated_row JSON;
BEGIN
    -- Convert the NEW row to JSON (if you want to notify the entire row)
    updated_row := row_to_json(NEW);

    -- Notify with the 'vault_update' channel and send the updated row as JSON
    PERFORM pg_notify('vault_update', updated_row::text);

    RETURN NEW;
END;
$BODY$;

ALTER FUNCTION public.notify_vault_update()
    OWNER TO postgres;

-- Trigger: vault_update

-- DROP TRIGGER IF EXISTS vault_update ON public."VaultStates";

CREATE OR REPLACE TRIGGER vault_update
    AFTER UPDATE 
    ON public."VaultStates"
    FOR EACH ROW
    EXECUTE FUNCTION public.notify_vault_update();


-- Table: public.Option_Buyers

-- DROP TABLE IF EXISTS public."Option_Buyers";

CREATE TABLE IF NOT EXISTS public."Option_Buyers"
(
    address character varying COLLATE pg_catalog."default" NOT NULL,
    round_id numeric(78,0) NOT NULL,
    has_minted boolean NOT NULL DEFAULT false,
    has_refunded boolean NOT NULL DEFAULT false,
    mintable_options numeric(78,0),
    refundable_options numeric(78,0),
    CONSTRAINT buyer_round PRIMARY KEY (address, round_id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public."Option_Buyers"
    OWNER to postgres;


-- FUNCTION: public.notify_ob_update()

-- DROP FUNCTION IF EXISTS public.notify_ob_update();


CREATE OR REPLACE FUNCTION public.notify_ob_update()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE NOT LEAKPROOF
AS $BODY$
DECLARE
    updated_row JSON;
BEGIN
    -- Convert the NEW row to JSON (if you want to notify the entire row)
    updated_row := row_to_json(NEW);

    -- Notify with the 'ob_update' channel and send the updated row as JSON
    PERFORM pg_notify('ob_update', updated_row::text);

    RETURN NEW;
END;
$BODY$;

ALTER FUNCTION public.notify_ob_update()
    OWNER TO postgres;


-- Trigger: ob_update

-- DROP TRIGGER IF EXISTS ob_update ON public."Option_Buyers";

CREATE OR REPLACE TRIGGER ob_update
    AFTER UPDATE 
    ON public."Option_Buyers"
    FOR EACH ROW
    EXECUTE FUNCTION public.notify_ob_update();


    


-- Table: public.Bids

-- DROP TABLE IF EXISTS public."Bids";

CREATE TABLE IF NOT EXISTS public."Bids"
(
    buyer_address character varying(67) COLLATE pg_catalog."default",
    round_address character varying(67) COLLATE pg_catalog."default" NOT NULL,
    bid_id character varying(67) COLLATE pg_catalog."default" NOT NULL,
    tree_nonce numeric(78,0),
    amount numeric(78,0),
    price numeric(78,0),
    CONSTRAINT round_address_bid_id PRIMARY KEY (round_address, bid_id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public."Bids"
    OWNER to postgres;