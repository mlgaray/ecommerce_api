-- Stored procedure for product update with batch operations
-- Returns deleted storage_refs for Cloudinary cleanup
-- Clean, maintainable, scalable approach with best practices

-- Drop old signature (without shop_id)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

-- Drop old signature (with shop_id, without is_stockeable)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

-- Drop current signature (with shop_id + is_stockeable)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, INTEGER, VARCHAR, TEXT, DECIMAL, BOOLEAN, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

CREATE OR REPLACE FUNCTION update_product(
    p_product_id INTEGER,
    p_shop_id INTEGER,       -- Shop ID for ownership validation
    p_name VARCHAR(255),
    p_description TEXT,
    p_price DECIMAL(10,2),
    p_is_stockeable BOOLEAN,
    p_stock INTEGER,
    p_minimum_stock INTEGER,
    p_is_active BOOLEAN,
    p_is_highlighted BOOLEAN,
    p_is_promotional BOOLEAN,
    p_promotional_price DECIMAL(10,2),
    p_category_id INTEGER,
    p_images JSONB,      -- [{"id": 1, "url": "...", "storage_ref": "..."}, {"url": "new", "storage_ref": "..."}]
    p_variants JSONB     -- [{"id": 1, "name": "...", "options": [...]}, {"name": "new", ...}]
) RETURNS text[] AS $$
DECLARE
    v_variant JSONB;
    v_variant_id INTEGER;
    v_deleted_refs text[];
BEGIN
    -- 0. Validate category exists before updating
    PERFORM 1 FROM categories WHERE id = p_category_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Category not found: %', p_category_id
        USING ERRCODE = 'P0003';  -- Category not found
    END IF;

    -- 1. UPDATE basic product fields (with shop_id validation)
    UPDATE products
    SET name = p_name,
        description = p_description,
        price = p_price,
        is_stockeable = p_is_stockeable,
        stock = p_stock,
        minimum_stock = p_minimum_stock,
        is_active = p_is_active,
        is_highlighted = p_is_highlighted,
        is_promotional = p_is_promotional,
        promotional_price = p_promotional_price,
        category_id = p_category_id
    WHERE id = p_product_id AND shop_id = p_shop_id;

    -- Check if product was found and belongs to shop
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Product not found or does not belong to shop: product_id=%, shop_id=%', p_product_id, p_shop_id
        USING ERRCODE = 'P0002';  -- no_data_found
    END IF;

    -- 2. UPSERT images
    IF COALESCE(jsonb_array_length(p_images), 0) = 0 THEN
        -- Capture refs before deletion for Cloudinary cleanup
        -- Using array_remove to filter out NULL values
        SELECT array_remove(array_agg(storage_ref), NULL) INTO v_deleted_refs
        FROM images
        WHERE product_id = p_product_id AND storage_ref <> '';

        DELETE FROM images WHERE product_id = p_product_id;
    ELSE
        -- Capture refs of images to be deleted (not in incoming list)
        -- Using NOT (id = ANY(...)) to handle empty arrays correctly
        -- Using array_remove to filter out NULL values
        SELECT array_remove(array_agg(storage_ref), NULL) INTO v_deleted_refs
        FROM images
        WHERE product_id = p_product_id
          AND storage_ref <> ''
          AND NOT (id = ANY(ARRAY(
            SELECT (img->>'id')::INTEGER
            FROM jsonb_array_elements(p_images) img
            WHERE img->>'id' IS NOT NULL AND img->>'id' <> ''
          )));

        -- Delete images NOT in the incoming list
        DELETE FROM images
        WHERE product_id = p_product_id
          AND NOT (id = ANY(ARRAY(
            SELECT (img->>'id')::INTEGER
            FROM jsonb_array_elements(p_images) img
            WHERE img->>'id' IS NOT NULL AND img->>'id' <> ''
          )));

        -- Insert new images (batch)
        INSERT INTO images (url, storage_ref, product_id)
        SELECT
            img->>'url',
            COALESCE(img->>'storage_ref', ''),
            p_product_id
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
          AND NOT (id = ANY(ARRAY(
            SELECT (v->>'id')::INTEGER
            FROM jsonb_array_elements(p_variants) v
            WHERE v->>'id' IS NOT NULL AND v->>'id' <> ''
          )));

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
                    max_selections = (v_variant->>'max_selections')::INTEGER,
                    is_required = COALESCE((v_variant->>'is_required')::BOOLEAN, false)
                WHERE id = v_variant_id;
            ELSE
                -- INSERT new variant
                INSERT INTO product_variants (name, "order", selection_type, max_selections, is_required, product_id)
                VALUES (
                    v_variant->>'name',
                    (v_variant->>'order')::INTEGER,
                    v_variant->>'selection_type',
                    (v_variant->>'max_selections')::INTEGER,
                    COALESCE((v_variant->>'is_required')::BOOLEAN, false),
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
                    -- Delete options NOT in the incoming list
                    DELETE FROM variant_options
                    WHERE variant_id = v_variant_id
                      AND NOT (id = ANY(ARRAY(SELECT id FROM option_data WHERE id IS NOT NULL)))
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

    -- Return deleted storage refs for Cloudinary cleanup (fire-and-forget)
    RETURN COALESCE(v_deleted_refs, '{}');

EXCEPTION
    WHEN unique_violation THEN
        RAISE;  -- Re-raise with original code 23505 for Go to handle
    -- Other errors propagate naturally with original SQLSTATE
END;
$$ LANGUAGE plpgsql
SET search_path = public;

-- Function documentation
COMMENT ON FUNCTION update_product IS
'Updates a product with images, variants and options in a single database call.
Parameters:
- p_product_id: ID of the product to update
- p_shop_id: Shop ID for ownership validation (product must belong to this shop)
- p_name, p_description, etc.: Product fields to update
- p_images: JSONB array of images
- p_variants: JSONB array of variants with options
Returns text[] with storage_refs of deleted images for Cloudinary cleanup.
Strategy:
- Product: 1 UPDATE with shop_id validation
- Images: Capture refs + DELETE + batch INSERT (uses unified images table)
- Variants: DELETE + loop for UPDATE/INSERT (N queries for N variants)
- Options: CTE with DELETE + batch UPDATE + batch INSERT per variant (1 query × N variants)
Optimizations applied:
- NOT (id = ANY(ARRAY(...))) instead of != ALL() to handle empty arrays correctly
- CTE for option_data to parse JSONB once (avoids 3x parsing overhead)
- COALESCE() for NULL safety
- Returns deleted storage_refs for async Cloudinary cleanup
Security: Validates product belongs to shop_id before allowing update.
Clean approach: Loop variants (few items), batch options per variant (many items).
Reduces ~10+ Go round trips to 1 stored procedure call.
Exception handling: only captures unique_violation, others propagate with original SQLSTATE.';
