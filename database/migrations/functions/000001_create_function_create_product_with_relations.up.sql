-- Stored procedure for product creation with batch inserts
-- Reduces 5 round trips to 1, and eliminates loops with batch INSERTs
-- Example: 3 variants × 5 options = 18 individual INSERTs → 2 batch INSERTs

DROP FUNCTION IF EXISTS create_product(
    VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL,
    INTEGER, INTEGER, TEXT[], JSONB
);

CREATE OR REPLACE FUNCTION create_product(
    p_name VARCHAR(255),
    p_description TEXT,
    p_price DECIMAL(10,2),
    p_stock INTEGER,
    p_minimum_stock INTEGER,
    p_is_active BOOLEAN,
    p_is_highlighted BOOLEAN,
    p_is_promotional BOOLEAN,
    p_promotional_price DECIMAL(10,2),
    p_category_id INTEGER,
    p_shop_id INTEGER,
    p_images TEXT[],
    p_variants JSONB
) RETURNS INTEGER AS $$
DECLARE
    v_product_id INTEGER;
    v_variant JSONB;
    v_variant_id INTEGER;
BEGIN
    -- 1. Insert product
    INSERT INTO products (
        name, description, price, stock, minimum_stock,
        is_active, is_highlighted, is_promotional, promotional_price,
        category_id, shop_id
    ) VALUES (
        p_name, p_description, p_price, p_stock, p_minimum_stock,
        p_is_active, p_is_highlighted, p_is_promotional, p_promotional_price,
        p_category_id, p_shop_id
    ) RETURNING id INTO v_product_id;

    -- 2. Insert images (batch with UNNEST)
    IF p_images IS NOT NULL AND COALESCE(array_length(p_images, 1), 0) > 0 THEN
        INSERT INTO product_images (url, product_id)
        SELECT unnest(p_images), v_product_id;
    END IF;

    -- 3. Insert variants and their options (avoid O(n²) array concatenation)
    -- Strategy: Loop variants (few items), batch insert options per variant
    -- Example: 3 variants + 15 options = 3 variant INSERTs + 3 batch option INSERTs
    IF COALESCE(jsonb_array_length(p_variants), 0) > 0 THEN
        -- Loop through variants (typically 2-5 items)
        FOR v_variant IN SELECT * FROM jsonb_array_elements(p_variants)
        LOOP
            -- Insert variant and get ID
            INSERT INTO product_variants (
                name, "order", selection_type, max_selections, product_id
            ) VALUES (
                v_variant->>'name',
                (v_variant->>'order')::INTEGER,
                v_variant->>'selection_type',
                (v_variant->>'max_selections')::INTEGER,
                v_product_id
            ) RETURNING id INTO v_variant_id;

            -- Batch insert options for THIS variant immediately (avoids array accumulation)
            IF COALESCE(jsonb_array_length(v_variant->'options'), 0) > 0 THEN
                INSERT INTO variant_options (name, price, "order", variant_id)
                SELECT
                    opt->>'name',
                    (opt->>'price')::DECIMAL,
                    (opt->>'order')::INTEGER,
                    v_variant_id
                FROM jsonb_array_elements(v_variant->'options') opt
                WHERE opt->>'name' IS NOT NULL  -- Validation: skip options without name
                  AND opt->>'price' IS NOT NULL  -- Validation: skip options without price
                  AND jsonb_typeof(opt) = 'object';
            END IF;
        END LOOP;
    END IF;

    RETURN v_product_id;

EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Error creating product: %', SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Function documentation
COMMENT ON FUNCTION create_product IS
'Creates a product with images, variants and options in a single call.
Performance optimizations:
- Product: 1 INSERT
- Images: 1 batch INSERT (all at once)
- Variants: N INSERTs (loop - typically 2-5 items)
- Options: N batch INSERTs (one per variant, avoids O(n²) array concatenation)
- COALESCE() for NULL safety on all length checks
- Data validation: skips options without name or price
- Exception handling for validation errors
Example: 3 variants + 15 options = 6 total INSERTs (3 variants + 3 batch option INSERTs)
Avoids O(n²) array_cat() overhead with 100+ options.
Reduces 5+ Go round trips to 1 stored procedure call.';
