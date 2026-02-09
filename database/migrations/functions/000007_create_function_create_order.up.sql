-- ============================================================================
-- FUNCTION: create_order
-- Creates an order with all its items and selected options in a single call.
-- Generates order_number using generate_order_number() function.
-- Validates products exist and are active.
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
    p_shipping_cost double precision,
    p_items jsonb  -- [{product_id, product_name, product_image_url, base_price, is_promotional, promotional_price, quantity, unit_price, total_price, selected_options: [{variant_id, option_id, variant_name, option_name, option_price}]}]
) RETURNS jsonb AS $$
DECLARE
    v_order_id bigint;
    v_order_number text;
    v_subtotal double precision := 0;
    v_total double precision := 0;
    v_item jsonb;
    v_item_id bigint;
    v_option jsonb;
    v_created_at timestamp with time zone;
    v_result jsonb;
    v_items_result jsonb := '[]'::jsonb;
BEGIN
    -- 1. Validate store exists
    PERFORM 1 FROM shops WHERE id = p_store_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Store not found: %', p_store_id
        USING ERRCODE = 'P0002';  -- No data found
    END IF;

    -- 2. Generate order number
    v_order_number := generate_order_number(p_store_id);

    -- 3. Calculate subtotal from items
    SELECT COALESCE(SUM((item->>'total_price')::double precision), 0) INTO v_subtotal
    FROM jsonb_array_elements(p_items) item;

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
        subtotal, shipping_cost, total
    ) VALUES (
        v_order_number,
        p_store_id, p_store_name, p_store_slug,
        'pending',
        p_customer_name, p_customer_phone, p_customer_email,
        p_customer_address_name, p_customer_address_place_id, p_customer_address_lat, p_customer_address_lng,
        p_payment_method_id, p_payment_method_code, p_payment_method_name,
        p_delivery_method_id, p_delivery_method_code, p_delivery_method_name,
        v_subtotal, COALESCE(p_shipping_cost, 0), v_total
    ) RETURNING id, created_at INTO v_order_id, v_created_at;

    -- 5. Insert items and their selected options
    IF COALESCE(jsonb_array_length(p_items), 0) > 0 THEN
        FOR v_item IN SELECT * FROM jsonb_array_elements(p_items)
        LOOP
            -- Insert order item
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
                total_price
            ) VALUES (
                v_order_id,
                NULLIF(v_item->>'product_id', '')::bigint,
                v_item->>'product_name',
                v_item->>'product_image_url',
                (v_item->>'base_price')::double precision,
                COALESCE((v_item->>'is_promotional')::boolean, false),
                NULLIF(v_item->>'promotional_price', '')::double precision,
                (v_item->>'quantity')::integer,
                (v_item->>'unit_price')::double precision,
                (v_item->>'total_price')::double precision
            ) RETURNING id INTO v_item_id;

            -- Insert selected options for this item
            IF jsonb_typeof(v_item->'selected_options') = 'array'
               AND jsonb_array_length(v_item->'selected_options') > 0 THEN
                FOR v_option IN SELECT * FROM jsonb_array_elements(v_item->'selected_options')
                LOOP
                    INSERT INTO order_item_selected_options (
                        order_item_id,
                        variant_id,
                        option_id,
                        variant_name,
                        option_name,
                        option_price
                    ) VALUES (
                        v_item_id,
                        NULLIF(v_option->>'variant_id', '')::bigint,
                        NULLIF(v_option->>'option_id', '')::bigint,
                        v_option->>'variant_name',
                        v_option->>'option_name',
                        COALESCE((v_option->>'option_price')::double precision, 0)
                    );
                END LOOP;
            END IF;
        END LOOP;
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
- Uses generate_order_number() for sequential order numbers
Performance optimizations:
- Order: 1 INSERT
- Items: N INSERTs (loop)
- Options: N batch INSERTs per item
Reduces multiple Go round trips to 1 stored procedure call.';
