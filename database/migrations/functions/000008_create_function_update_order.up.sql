-- ============================================================================
-- FUNCTION: update_order
-- Updates an order with all its items and selected options in a single call.
-- Validates order exists AND belongs to the store.
-- Recalculates subtotal and total from items.
-- Deletes all existing items (CASCADE handles selections) and inserts new ones.
-- Uses set-based INSERT...SELECT instead of FOR loops for better performance.
-- Returns JSONB with updated_at, subtotal, total.
-- ============================================================================
CREATE OR REPLACE FUNCTION update_order(
    p_order_id bigint,
    p_store_id bigint,
    p_customer_name text,
    p_customer_phone text,
    p_customer_email text,
    p_address_name text,
    p_address_place_id text,
    p_address_lat double precision,
    p_address_lng double precision,
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
    p_discount double precision,
    p_coupon_code text,
    p_coupon_type text,
    p_coupon_value double precision,
    p_coupon_discount_amount double precision,
    p_coupon_min_order_amount double precision,
    p_coupon_id bigint,
    p_coupon_phone text,
    p_items jsonb  -- [{product_id, product_name, product_image_url, base_price, is_promotional, promotional_price, quantity, unit_price, total_price, notes, selected_options: [{variant_id, option_id, variant_name, option_name, option_price, quantity}]}]
) RETURNS jsonb AS $$
DECLARE
    v_subtotal double precision := 0;
    v_total double precision := 0;
    v_updated_at timestamp with time zone;
    v_result jsonb;
    v_usage_limit integer;
    v_max_uses_per_phone integer;
    v_current_usages integer;
BEGIN
    -- 1. Calculate subtotal from items using jsonb_to_recordset (typed, no casts)
    SELECT COALESCE(SUM(i.total_price), 0) INTO v_subtotal
    FROM jsonb_to_recordset(p_items) AS i(total_price double precision);

    v_total := v_subtotal - COALESCE(p_discount, 0) + COALESCE(p_shipping_cost, 0);

    -- 2. Update order fields (customer, payment, delivery, delivery_zone, coupon, totals)
    UPDATE orders SET
        customer_name = p_customer_name,
        customer_phone = p_customer_phone,
        customer_email = p_customer_email,
        customer_address_name = p_address_name,
        customer_address_place_id = p_address_place_id,
        customer_address_lat = p_address_lat,
        customer_address_lng = p_address_lng,
        payment_method_id = p_payment_method_id,
        payment_method_code = p_payment_method_code,
        payment_method_name = p_payment_method_name,
        delivery_method_id = p_delivery_method_id,
        delivery_method_code = p_delivery_method_code,
        delivery_method_name = p_delivery_method_name,
        delivery_zone_id = p_delivery_zone_id,
        delivery_zone_name = p_delivery_zone_name,
        delivery_zone_price = p_delivery_zone_price,
        subtotal = v_subtotal,
        shipping_cost = COALESCE(p_shipping_cost, 0),
        discount = COALESCE(p_discount, 0),
        total = v_total,
        coupon_code = p_coupon_code,
        coupon_type = p_coupon_type,
        coupon_value = p_coupon_value,
        coupon_discount_amount = p_coupon_discount_amount,
        coupon_min_order_amount = p_coupon_min_order_amount,
        updated_at = now()
    WHERE id = p_order_id AND store_id = p_store_id
    RETURNING updated_at INTO v_updated_at;

    -- 3. Verify the order exists and belongs to the store
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Order not found: order_id=%, store_id=%', p_order_id, p_store_id
        USING ERRCODE = 'P0002';  -- No data found
    END IF;

    -- 3b. Manage coupon_usages atomically with limit enforcement
    -- Lock coupon FIRST to serialize concurrent usage attempts, then clean + re-insert.
    IF p_coupon_id IS NOT NULL THEN
        SELECT usage_limit, max_uses_per_phone INTO v_usage_limit, v_max_uses_per_phone
        FROM coupons WHERE id = p_coupon_id FOR UPDATE;
    END IF;

    DELETE FROM coupon_usages WHERE order_id = p_order_id;

    IF p_coupon_id IS NOT NULL THEN

        -- Enforce global usage limit (after DELETE, so re-applying same coupon doesn't double-count)
        IF v_usage_limit IS NOT NULL THEN
            SELECT COUNT(*) INTO v_current_usages
            FROM coupon_usages WHERE coupon_id = p_coupon_id;

            IF v_current_usages >= v_usage_limit THEN
                RAISE EXCEPTION 'coupon_usage_limit_reached'
                USING ERRCODE = 'P0003';
            END IF;
        END IF;

        -- Enforce per-phone usage limit
        IF v_max_uses_per_phone IS NOT NULL AND p_coupon_phone IS NOT NULL THEN
            SELECT COUNT(*) INTO v_current_usages
            FROM coupon_usages WHERE coupon_id = p_coupon_id AND phone = p_coupon_phone;

            IF v_current_usages >= v_max_uses_per_phone THEN
                RAISE EXCEPTION 'coupon_phone_usage_limit_reached'
                USING ERRCODE = 'P0003';
            END IF;
        END IF;

        INSERT INTO coupon_usages (coupon_id, order_id, phone)
        VALUES (p_coupon_id, p_order_id, p_coupon_phone);
    END IF;

    -- 4. Delete all existing order_items (CASCADE handles order_item_selections)
    DELETE FROM order_items WHERE order_id = p_order_id;

    -- 5. Insert new items and their selected options using set-based CTE chain
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
                p_order_id,
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

    -- 6. Build result JSON
    v_result := jsonb_build_object(
        'updated_at', v_updated_at,
        'subtotal', v_subtotal,
        'discount', COALESCE(p_discount, 0),
        'total', v_total
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
COMMENT ON FUNCTION update_order IS
'Updates an order with items and selected options in a single call.
Parameters:
- p_order_id: ID of the order to update
- p_store_id: Store ID for ownership validation (order must belong to this store)
- Customer fields: name, phone, email, address (name, place_id, lat, lng)
- Payment method fields: id, code, name
- Delivery method fields: id, code, name
- Delivery zone fields: id, name, price
- p_shipping_cost: Shipping cost
- p_items: JSONB array of items with selected_options
Validations:
- Order existence and store ownership (P0002 if not found)
Performance:
- 1 UPDATE (order)
- 1 DELETE (existing items, CASCADE handles selections)
- 1 batch INSERT (all items via INSERT...SELECT with jsonb_to_recordset)
- 1 batch INSERT (all selections via INSERT...SELECT with JOIN + LATERAL)
- Uses jsonb_array_elements WITH ORDINALITY for guaranteed array position
- Uses row_number() OVER (ORDER BY id) to recover insertion order from RETURNING
Returns JSONB with updated_at, subtotal, total.';
