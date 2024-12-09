-- FUNCTION: public.notify_lp_update()

CREATE FUNCTION public.notify_lp_update()
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



-- FUNCTION: public.notify_or_update()

CREATE FUNCTION public.notify_or_update()
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




-- FUNCTION: public.notify_vault_update()

CREATE FUNCTION public.notify_vault_update()
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


-- FUNCTION: public.notify_ob_update()


CREATE FUNCTION public.notify_ob_update()
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
    

CREATE FUNCTION public.notify_bids_channel()
    RETURNS trigger
    LANGUAGE 'plpgsql'
    COST 100
    VOLATILE NOT LEAKPROOF
AS $BODY$
DECLARE
    notification_payload JSON;
BEGIN
    -- Build the payload with operation type and the new or updated row
    IF (TG_OP = 'INSERT') THEN
        notification_payload := json_build_object(
            'operation', 'INSERT',
            'data', row_to_json(NEW)
        );
    ELSIF (TG_OP = 'UPDATE') THEN
        notification_payload := json_build_object(
            'operation', 'UPDATE',
            'data', row_to_json(NEW)
        );
    END IF;

    -- Notify the 'bids_channel' channel with the payload
    PERFORM pg_notify('bids_channel', notification_payload::text);

    RETURN NEW;
END;
$BODY$;

