-- Stored procedure for product update with batch operations
-- Clean, maintainable, scalable approach with best practices

DROP FUNCTION IF EXISTS update_product(
    INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

CREATE OR REPLACE FUNCTION update_product(
    p_product_id INTEGER,
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
    p_images JSONB,      -- [{"id": 1, "url": "..."}, {"url": "new"}]
    p_variants JSONB     -- [{"id": 1, "name": "...", "options": [...]}, {"name": "new", ...}]
) RETURNS VOID AS $$
DECLARE
    v_variant JSONB;
    v_variant_id INTEGER;
BEGIN
    -- 1. UPDATE basic product fields
    UPDATE products
    SET name = p_name,
        description = p_description,
        price = p_price,
        stock = p_stock,
        minimum_stock = p_minimum_stock,
        is_active = p_is_active,
        is_highlighted = p_is_highlighted,
        is_promotional = p_is_promotional,
        promotional_price = p_promotional_price,
        category_id = p_category_id
    WHERE id = p_product_id;

    -- 2. UPSERT images (optimized with != ALL for better index usage)
    IF COALESCE(jsonb_array_length(p_images), 0) = 0 THEN
        DELETE FROM product_images WHERE product_id = p_product_id;
    ELSE
        -- Delete images NOT in the incoming list (using != ALL for performance)
        DELETE FROM product_images
        WHERE product_id = p_product_id
          AND id != ALL (
            SELECT (img->>'id')::INTEGER
            FROM jsonb_array_elements(p_images) img
            WHERE img->>'id' IS NOT NULL AND img->>'id' != ''
          );

        -- Insert new images (batch)
        INSERT INTO product_images (url, product_id)
        SELECT img->>'url', p_product_id
        FROM jsonb_array_elements(p_images) img
        WHERE img->>'id' IS NULL OR img->>'id' = '';
    END IF;

    -- 3. UPSERT variants and options
    IF COALESCE(jsonb_array_length(p_variants), 0) = 0 THEN
        DELETE FROM product_variants WHERE product_id = p_product_id;
    ELSE
        -- 3a. Delete variants NOT in the incoming list (cascade deletes options)
        DELETE FROM product_variants
        WHERE product_id = p_product_id
          AND id != ALL (
            SELECT (v->>'id')::INTEGER
            FROM jsonb_array_elements(p_variants) v
            WHERE v->>'id' IS NOT NULL AND v->>'id' != ''
          );

        -- 3b. Update or insert each variant (loop needed - typically 2-5 variants)
        FOR v_variant IN SELECT * FROM jsonb_array_elements(p_variants)
        LOOP
            IF v_variant->>'id' IS NOT NULL AND v_variant->>'id' != '' THEN
                -- UPDATE existing variant
                v_variant_id := (v_variant->>'id')::INTEGER;

                UPDATE product_variants
                SET name = v_variant->>'name',
                    "order" = (v_variant->>'order')::INTEGER,
                    selection_type = v_variant->>'selection_type',
                    max_selections = (v_variant->>'max_selections')::INTEGER
                WHERE id = v_variant_id;
            ELSE
                -- INSERT new variant
                INSERT INTO product_variants (name, "order", selection_type, max_selections, product_id)
                VALUES (
                    v_variant->>'name',
                    (v_variant->>'order')::INTEGER,
                    v_variant->>'selection_type',
                    (v_variant->>'max_selections')::INTEGER,
                    p_product_id
                ) RETURNING id INTO v_variant_id;
            END IF;

            -- 3c. UPSERT options for this variant (optimized with CTE to avoid reparsing JSONB)
            IF v_variant->'options' IS NOT NULL AND COALESCE(jsonb_array_length(v_variant->'options'), 0) > 0 THEN
                -- Parse JSONB once and reuse (avoids 3x parsing overhead)
                WITH option_data AS (
                    SELECT
                        (opt->>'id')::INTEGER AS id,
                        opt->>'name' AS name,
                        (opt->>'price')::DECIMAL AS price,
                        (opt->>'order')::INTEGER AS "order"
                    FROM jsonb_array_elements(v_variant->'options') opt
                ),
                deleted AS (
                    -- Delete options NOT in the incoming list (using != ALL for performance)
                    DELETE FROM variant_options
                    WHERE variant_id = v_variant_id
                      AND id != ALL (SELECT id FROM option_data WHERE id IS NOT NULL)
                ),
                updated AS (
                    -- Batch UPDATE existing options
                    UPDATE variant_options vo
                    SET name = od.name,
                        price = od.price,
                        "order" = od."order"
                    FROM option_data od
                    WHERE vo.id = od.id AND vo.variant_id = v_variant_id
                )
                -- Batch INSERT new options
                INSERT INTO variant_options (name, price, "order", variant_id)
                SELECT name, price, "order", v_variant_id
                FROM option_data
                WHERE id IS NULL;
            ELSE
                -- No options provided, delete all for this variant
                DELETE FROM variant_options WHERE variant_id = v_variant_id;
            END IF;
        END LOOP;
    END IF;

EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Error updating product (ID: %): %', p_product_id, SQLERRM;
END;
$$ LANGUAGE plpgsql;

-- Function documentation
COMMENT ON FUNCTION update_product IS
'Updates a product with images, variants and options in a single database call.
Strategy:
- Product: 1 UPDATE
- Images: DELETE + batch INSERT (2 queries)
- Variants: DELETE + loop for UPDATE/INSERT (N queries for N variants)
- Options: CTE with DELETE + batch UPDATE + batch INSERT per variant (1 query × N variants)
Optimizations applied:
- != ALL() instead of NOT IN for better index usage and NULL handling
- CTE for option_data to parse JSONB once (avoids 3x parsing overhead)
- COALESCE() for NULL safety
Clean approach: Loop variants (few items), batch options per variant (many items).
Reduces ~10+ Go round trips to 1 stored procedure call.
Exception handling included for validation errors.';
