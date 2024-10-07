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
