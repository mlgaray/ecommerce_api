-- ============================================================================
-- FUNCTION: create_order
-- Creates an order with all its items and selected options in a single call.
-- Generates order_number using generate_order_number() function.
-- Uses set-based INSERT...SELECT instead of FOR loops for better performance.
-- Returns JSONB with the created order including all relations.
-- ============================================================================
CREATE OR REPLACE FUNCTION create_order(
    p_store_id bigint,
    p_store_name text,
    p_store_slug text,
    p_customer_name text,
    p_customer_phone text,
    p_customer_email text,
    p_customer_address_name text,
    p_customer_address_place_id text,
    p_customer_address_lat double precision,
    p_customer_address_lng double precision,
    p_payment_method_id bigint,
    p_payment_method_code text,
    p_payment_method_name text,
    p_delivery_method_id bigint,
    p_delivery_method_code text,
    p_delivery_method_name text,
    p_delivery_zone_id bigint,
    p_delivery_zone_name text,
    p_delivery_zone_price double precision,
    p_shipping_cost double precision,
    p_items jsonb  -- [{product_id, product_name, product_image_url, base_price, is_promotional, promotional_price, quantity, unit_price, total_price, notes, selected_options: [{variant_id, option_id, variant_name, option_name, option_price, quantity}]}]
) RETURNS jsonb AS $$
DECLARE
    v_order_id bigint;
    v_order_number text;
    v_subtotal double precision := 0;
    v_total double precision := 0;
    v_created_at timestamp with time zone;
    v_result jsonb;
BEGIN
    -- 1. Validate store exists
    PERFORM 1 FROM shops WHERE id = p_store_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Store not found: %', p_store_id
        USING ERRCODE = 'P0002';  -- No data found
    END IF;

    -- 2. Generate order number (uses global SEQUENCE, no lock needed)
    v_order_number := generate_order_number(p_store_id);

    -- 3. Calculate subtotal from items using jsonb_to_recordset (typed, no casts)
    SELECT COALESCE(SUM(i.total_price), 0) INTO v_subtotal
    FROM jsonb_to_recordset(p_items) AS i(total_price double precision);

    v_total := v_subtotal + COALESCE(p_shipping_cost, 0);

    -- 4. Insert order
    INSERT INTO orders (
        order_number,
        store_id, store_name, store_slug,
        status,
        customer_name, customer_phone, customer_email,
        customer_address_name, customer_address_place_id, customer_address_lat, customer_address_lng,
        payment_method_id, payment_method_code, payment_method_name,
        delivery_method_id, delivery_method_code, delivery_method_name,
        delivery_zone_id, delivery_zone_name, delivery_zone_price,
        subtotal, shipping_cost, total
    ) VALUES (
        v_order_number,
        p_store_id, p_store_name, p_store_slug,
        'pending',
        p_customer_name, p_customer_phone, p_customer_email,
        p_customer_address_name, p_customer_address_place_id, p_customer_address_lat, p_customer_address_lng,
        p_payment_method_id, p_payment_method_code, p_payment_method_name,
        p_delivery_method_id, p_delivery_method_code, p_delivery_method_name,
        p_delivery_zone_id, p_delivery_zone_name, p_delivery_zone_price,
        v_subtotal, COALESCE(p_shipping_cost, 0), v_total
    ) RETURNING id, created_at INTO v_order_id, v_created_at;

    -- 5. Insert items and their selected options using set-based CTE chain
    IF COALESCE(jsonb_array_length(p_items), 0) > 0 THEN
        WITH expanded_items AS (
            -- Use jsonb_array_elements WITH ORDINALITY for guaranteed array position
            SELECT
                ordinality AS item_index,
                (elem->>'product_id')                                   AS product_id,
                (elem->>'product_name')                                 AS product_name,
                (elem->>'product_image_url')                            AS product_image_url,
                (elem->>'base_price')::double precision                 AS base_price,
                COALESCE((elem->>'is_promotional')::boolean, false)     AS is_promotional,
                (elem->>'promotional_price')                            AS promotional_price,
                (elem->>'quantity')::integer                            AS quantity,
                (elem->>'unit_price')::double precision                 AS unit_price,
                (elem->>'total_price')::double precision                AS total_price,
                (elem->>'notes')                                       AS notes,
                (elem->'selected_options')                              AS selected_options
            FROM jsonb_array_elements(p_items) WITH ORDINALITY AS t(elem, ordinality)
        ),
        inserted_items AS (
            -- Batch insert all order items in one INSERT...SELECT
            INSERT INTO order_items (
                order_id,
                product_id,
                product_name,
                product_image_url,
                base_price,
                is_promotional,
                promotional_price,
                quantity,
                unit_price,
                total_price,
                notes
            )
            SELECT
                v_order_id,
                NULLIF(ei.product_id, '')::bigint,
                ei.product_name,
                ei.product_image_url,
                ei.base_price,
                COALESCE(ei.is_promotional, false),
                NULLIF(ei.promotional_price, '')::double precision,
                ei.quantity,
                ei.unit_price,
                ei.total_price,
                ei.notes
            FROM expanded_items ei
            ORDER BY ei.item_index
            RETURNING id
        ),
        numbered_items AS (
            -- Sort by id to recover insertion order (IDENTITY is monotonic within a single INSERT)
            SELECT id, row_number() OVER (ORDER BY id) AS item_index
            FROM inserted_items
        )
        -- Batch insert all order item selections in one INSERT...SELECT
        INSERT INTO order_item_selections (
            order_item_id,
            variant_id,
            option_id,
            variant_name,
            option_name,
            option_price,
            quantity
        )
        SELECT
            ni.id,
            NULLIF(o.variant_id, '')::bigint,
            NULLIF(o.option_id, '')::bigint,
            o.variant_name,
            o.option_name,
            COALESCE(o.option_price, 0),
            COALESCE(o.quantity, 1)
        FROM numbered_items ni
        JOIN expanded_items ei ON ei.item_index = ni.item_index
        CROSS JOIN LATERAL jsonb_to_recordset(ei.selected_options) AS o(
            variant_id text,
            option_id text,
            variant_name text,
            option_name text,
            option_price double precision,
            quantity integer
        )
        WHERE ei.selected_options IS NOT NULL
          AND jsonb_typeof(ei.selected_options) = 'array'
          AND jsonb_array_length(ei.selected_options) > 0;
    END IF;

    -- 6. Build result JSON with order data
    v_result := jsonb_build_object(
        'id', v_order_id,
        'order_number', v_order_number,
        'status', 'pending',
        'subtotal', v_subtotal,
        'shipping_cost', COALESCE(p_shipping_cost, 0),
        'total', v_total,
        'created_at', v_created_at
    );

    RETURN v_result;

EXCEPTION
    WHEN unique_violation THEN
        RAISE;  -- Re-raise with original code 23505 for Go to handle
    WHEN foreign_key_violation THEN
        RAISE;  -- Re-raise for FK violations
END;
$$ LANGUAGE plpgsql
SET search_path = public;

-- Function documentation
COMMENT ON FUNCTION create_order IS
'Creates an order with items and selected options in a single call.
Validations:
- Store existence (P0002 if not found)
- Uses generate_order_number() with global SEQUENCE (lock-free, no contention)
Performance:
- 1 INSERT (order)
- 1 batch INSERT (all items via INSERT...SELECT with jsonb_to_recordset)
- 1 batch INSERT (all selections via INSERT...SELECT with JOIN + LATERAL)
- Uses jsonb_array_elements WITH ORDINALITY for guaranteed array position
- Uses row_number() OVER (ORDER BY id) to recover insertion order from RETURNING
Reduces multiple Go round trips to 1 stored procedure call.';
