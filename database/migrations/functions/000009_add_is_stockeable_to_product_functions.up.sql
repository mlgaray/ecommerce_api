-- Update stored procedures to include is_stockeable field
-- Migration: 000009_add_is_stockeable_to_product_functions

-- =============================================================================
-- DROP OLD FUNCTION SIGNATURES
-- =============================================================================

-- Drop old create_product signatures
DROP FUNCTION IF EXISTS create_product(
    VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL,
    INTEGER, INTEGER, TEXT[], JSONB
);

DROP FUNCTION IF EXISTS create_product(
    VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL,
    INTEGER, INTEGER, JSONB, JSONB
);

-- Drop old update_product signatures
DROP FUNCTION IF EXISTS update_product(
    INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

DROP FUNCTION IF EXISTS update_product(
    INTEGER, INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

-- =============================================================================
-- CREATE PRODUCT FUNCTION (with is_stockeable)
-- =============================================================================

CREATE OR REPLACE FUNCTION create_product(
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
    p_shop_id INTEGER,
    p_images JSONB,
    p_variants JSONB
) RETURNS INTEGER AS $$
DECLARE
    v_product_id INTEGER;
    v_variant JSONB;
    v_variant_id INTEGER;
BEGIN
    -- 0. Validate foreign keys exist before inserting
    PERFORM 1 FROM categories WHERE id = p_category_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Category not found: %', p_category_id
        USING ERRCODE = 'P0003';
    END IF;

    PERFORM 1 FROM shops WHERE id = p_shop_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Shop not found: %', p_shop_id
        USING ERRCODE = 'P0004';
    END IF;

    -- 1. Insert product
    INSERT INTO products (
        name, description, price, is_stockeable, stock, minimum_stock,
        is_active, is_highlighted, is_promotional, promotional_price,
        category_id, shop_id
    ) VALUES (
        p_name, p_description, p_price, p_is_stockeable, p_stock, p_minimum_stock,
        p_is_active, p_is_highlighted, p_is_promotional, p_promotional_price,
        p_category_id, p_shop_id
    ) RETURNING id INTO v_product_id;

    -- 2. Insert images (batch with jsonb_array_elements)
    IF p_images IS NOT NULL AND COALESCE(jsonb_array_length(p_images), 0) > 0 THEN
        INSERT INTO images (url, storage_ref, product_id)
        SELECT
            img->>'url',
            COALESCE(img->>'storage_ref', ''),
            v_product_id
        FROM jsonb_array_elements(p_images) img
        WHERE jsonb_typeof(img) = 'object'
          AND trim(COALESCE(img->>'url', '')) <> '';
    END IF;

    -- 3. Insert variants and their options
    IF COALESCE(jsonb_array_length(p_variants), 0) > 0 THEN
        FOR v_variant IN SELECT * FROM jsonb_array_elements(p_variants)
        LOOP
            INSERT INTO product_variants (
                name, "order", selection_type, max_selections, is_required, product_id
            ) VALUES (
                v_variant->>'name',
                COALESCE(NULLIF(v_variant->>'order', '')::INTEGER, 0),
                v_variant->>'selection_type',
                COALESCE(NULLIF(v_variant->>'max_selections', '')::INTEGER, 0),
                COALESCE((v_variant->>'is_required')::BOOLEAN, false),
                v_product_id
            ) RETURNING id INTO v_variant_id;

            IF jsonb_typeof(v_variant->'options') = 'array'
               AND jsonb_array_length(v_variant->'options') > 0 THEN
                INSERT INTO variant_options (name, price, "order", variant_id)
                SELECT
                    opt->>'name',
                    COALESCE(NULLIF(opt->>'price', '')::DECIMAL, 0),
                    COALESCE(NULLIF(opt->>'order', '')::INTEGER, 0),
                    v_variant_id
                FROM jsonb_array_elements(v_variant->'options') opt
                WHERE jsonb_typeof(opt) = 'object'
                  AND trim(COALESCE(opt->>'name', '')) <> '';
            END IF;
        END LOOP;
    END IF;

    RETURN v_product_id;

EXCEPTION
    WHEN unique_violation THEN
        RAISE;
END;
$$ LANGUAGE plpgsql
SET search_path = public;

-- =============================================================================
-- UPDATE PRODUCT FUNCTION (with is_stockeable)
-- =============================================================================

CREATE OR REPLACE FUNCTION update_product(
    p_product_id INTEGER,
    p_shop_id INTEGER,
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
    p_images JSONB,
    p_variants JSONB
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
        USING ERRCODE = 'P0003';
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

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Product not found or does not belong to shop: product_id=%, shop_id=%', p_product_id, p_shop_id
        USING ERRCODE = 'P0002';
    END IF;

    -- 2. UPSERT images
    IF COALESCE(jsonb_array_length(p_images), 0) = 0 THEN
        SELECT array_remove(array_agg(storage_ref), NULL) INTO v_deleted_refs
        FROM images
        WHERE product_id = p_product_id AND storage_ref <> '';

        DELETE FROM images WHERE product_id = p_product_id;
    ELSE
        SELECT array_remove(array_agg(storage_ref), NULL) INTO v_deleted_refs
        FROM images
        WHERE product_id = p_product_id
          AND storage_ref <> ''
          AND NOT (id = ANY(ARRAY(
            SELECT (img->>'id')::INTEGER
            FROM jsonb_array_elements(p_images) img
            WHERE img->>'id' IS NOT NULL AND img->>'id' <> ''
          )));

        DELETE FROM images
        WHERE product_id = p_product_id
          AND NOT (id = ANY(ARRAY(
            SELECT (img->>'id')::INTEGER
            FROM jsonb_array_elements(p_images) img
            WHERE img->>'id' IS NOT NULL AND img->>'id' <> ''
          )));

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
        DELETE FROM product_variants
        WHERE product_id = p_product_id
          AND NOT (id = ANY(ARRAY(
            SELECT (v->>'id')::INTEGER
            FROM jsonb_array_elements(p_variants) v
            WHERE v->>'id' IS NOT NULL AND v->>'id' <> ''
          )));

        FOR v_variant IN SELECT * FROM jsonb_array_elements(p_variants)
        LOOP
            IF v_variant->>'id' IS NOT NULL AND v_variant->>'id' != '' THEN
                v_variant_id := (v_variant->>'id')::INTEGER;

                UPDATE product_variants
                SET name = v_variant->>'name',
                    "order" = (v_variant->>'order')::INTEGER,
                    selection_type = v_variant->>'selection_type',
                    max_selections = (v_variant->>'max_selections')::INTEGER,
                    is_required = COALESCE((v_variant->>'is_required')::BOOLEAN, false)
                WHERE id = v_variant_id;
            ELSE
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

            IF v_variant->'options' IS NOT NULL AND COALESCE(jsonb_array_length(v_variant->'options'), 0) > 0 THEN
                WITH option_data AS (
                    SELECT
                        (opt->>'id')::INTEGER AS id,
                        opt->>'name' AS name,
                        (opt->>'price')::DECIMAL AS price,
                        (opt->>'order')::INTEGER AS "order"
                    FROM jsonb_array_elements(v_variant->'options') opt
                ),
                deleted AS (
                    DELETE FROM variant_options
                    WHERE variant_id = v_variant_id
                      AND NOT (id = ANY(ARRAY(SELECT id FROM option_data WHERE id IS NOT NULL)))
                ),
                updated AS (
                    UPDATE variant_options vo
                    SET name = od.name,
                        price = od.price,
                        "order" = od."order"
                    FROM option_data od
                    WHERE vo.id = od.id AND vo.variant_id = v_variant_id
                )
                INSERT INTO variant_options (name, price, "order", variant_id)
                SELECT name, price, "order", v_variant_id
                FROM option_data
                WHERE id IS NULL;
            ELSE
                DELETE FROM variant_options WHERE variant_id = v_variant_id;
            END IF;
        END LOOP;
    END IF;

    RETURN COALESCE(v_deleted_refs, '{}');

EXCEPTION
    WHEN unique_violation THEN
        RAISE;
END;
$$ LANGUAGE plpgsql
SET search_path = public;
