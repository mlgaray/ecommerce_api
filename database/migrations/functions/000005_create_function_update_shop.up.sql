-- Stored procedure for shop update with all relations
-- Returns deleted storage_refs for Cloudinary cleanup
-- OPTIMIZED V2: Single JSON parse + shop_id validation for security

DROP FUNCTION IF EXISTS update_shop(
    INTEGER, TEXT, TEXT, TEXT, TEXT, TEXT,
    JSONB, JSONB, JSONB, JSONB, JSONB
);

CREATE OR REPLACE FUNCTION update_shop(
    p_shop_id INTEGER,
    p_name TEXT,
    p_slug TEXT,
    p_email TEXT,
    p_phone TEXT,
    p_instagram TEXT,
    p_images JSONB,              -- [{"id": 1, "url": "...", "type": "logo"}, {"url": "new", "type": "cover", "storage_ref": "..."}]
    p_address JSONB,             -- {"id": 1, "name": "...", "place_id": "...", "lat": -34.6, "lng": -58.4} or NULL
    p_payment_methods JSONB,     -- [{"id": 10, "is_active": true, "transfer_config": {...}, "mercadopago_config": {...}}]
    p_delivery_methods JSONB,    -- [{"id": 20, "is_active": true, "delivery_config": {...}, "pickup_config": {...}, "delivery_zones": [...]}]
    p_operating_schedules JSONB  -- [{"day_of_week": 1, "open_time": "09:00", "close_time": "13:00"}]
) RETURNS TEXT[] AS $$
DECLARE
    v_deleted_refs TEXT[];
BEGIN
    -- =============================================
    -- 1. UPDATE shop basic fields
    -- =============================================
    UPDATE shops
    SET name = p_name,
        slug = p_slug,
        email = p_email,
        phone = p_phone,
        instagram = p_instagram
    WHERE id = p_shop_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Shop not found: %', p_shop_id
        USING ERRCODE = 'P0002';
    END IF;

    -- =============================================
    -- 2. UPSERT images (logo, cover) - SET-BASED
    -- NULL = don't touch, [] = delete all
    -- =============================================
    IF p_images IS NULL THEN
        -- NULL: don't modify images (skip section)
        NULL;
    ELSIF jsonb_array_length(p_images) = 0 THEN
        -- Empty array: delete all images, capture storage_refs
        SELECT array_remove(array_agg(storage_ref), NULL) INTO v_deleted_refs
        FROM images
        WHERE shop_id = p_shop_id AND storage_ref <> '';

        DELETE FROM images WHERE shop_id = p_shop_id;
    ELSE
        -- Parse images once, delete old, insert new - all in one CTE chain
        WITH img_input AS (
            SELECT
                (img->>'id')::INTEGER AS id,
                img->>'url' AS url,
                COALESCE(img->>'storage_ref', '') AS storage_ref,
                img->>'type' AS type
            FROM jsonb_array_elements(p_images) img
        ),
        deleted_refs AS (
            SELECT storage_ref
            FROM images i
            WHERE i.shop_id = p_shop_id
              AND i.storage_ref <> ''
              AND NOT EXISTS (SELECT 1 FROM img_input ii WHERE ii.id = i.id)
        ),
        deleted AS (
            DELETE FROM images i
            WHERE i.shop_id = p_shop_id
              AND NOT EXISTS (SELECT 1 FROM img_input ii WHERE ii.id = i.id)
        )
        SELECT array_remove(array_agg(storage_ref), NULL) INTO v_deleted_refs
        FROM deleted_refs;

        -- Insert new images (where id IS NULL)
        WITH img_input AS (
            SELECT
                (img->>'id')::INTEGER AS id,
                img->>'url' AS url,
                COALESCE(img->>'storage_ref', '') AS storage_ref,
                img->>'type' AS type
            FROM jsonb_array_elements(p_images) img
        )
        INSERT INTO images (url, storage_ref, type, shop_id)
        SELECT ii.url, ii.storage_ref, ii.type, p_shop_id
        FROM img_input ii
        WHERE ii.id IS NULL;
    END IF;

    -- =============================================
    -- 3. UPSERT address (0-1)
    -- =============================================
    IF p_address IS NULL OR p_address = 'null'::jsonb OR p_address = '{}'::jsonb THEN
        DELETE FROM addresses WHERE shop_id = p_shop_id;
    ELSE
        DELETE FROM addresses WHERE shop_id = p_shop_id;
        INSERT INTO addresses (name, place_id, ltd, lng, shop_id)
        VALUES (
            p_address->>'name',
            p_address->>'place_id',
            (p_address->>'lat')::DOUBLE PRECISION,
            (p_address->>'lng')::DOUBLE PRECISION,
            p_shop_id
        );
    END IF;

    -- =============================================
    -- 4. UPDATE payment methods - OPTIMIZED
    -- Single JSON parse + shop_id validation via JOIN
    -- =============================================
    IF p_payment_methods IS NOT NULL AND jsonb_array_length(p_payment_methods) > 0 THEN
        -- Create temp table to materialize JSON ONCE
        CREATE TEMP TABLE _pm_input ON COMMIT DROP AS
        SELECT
            (j->>'id')::INTEGER AS spm_id,
            (j->>'is_active')::BOOLEAN AS is_active,
            j->'transfer_config' AS transfer_config,
            j->'mercadopago_config' AS mercadopago_config
        FROM jsonb_array_elements(p_payment_methods) j;

        -- UPDATE is_active (validated by shop_id)
        UPDATE shop_payment_methods spm
        SET is_active = pm.is_active
        FROM _pm_input pm
        WHERE spm.id = pm.spm_id AND spm.shop_id = p_shop_id;

        -- UPSERT transfer_configs (validated by JOIN + required fields check)
        INSERT INTO transfer_configs (shop_payment_method_id, cbu, cuil, alias, owner_name)
        SELECT
            pm.spm_id,
            pm.transfer_config->>'cbu',
            pm.transfer_config->>'cuil',
            pm.transfer_config->>'alias',
            pm.transfer_config->>'owner_name'
        FROM _pm_input pm
        JOIN shop_payment_methods spm ON spm.id = pm.spm_id AND spm.shop_id = p_shop_id
        WHERE pm.transfer_config IS NOT NULL
          AND pm.transfer_config != 'null'::jsonb
          AND pm.transfer_config != '{}'::jsonb
          AND (pm.transfer_config->>'cbu') IS NOT NULL
          AND (pm.transfer_config->>'cuil') IS NOT NULL
          AND (pm.transfer_config->>'owner_name') IS NOT NULL
        ON CONFLICT (shop_payment_method_id) DO UPDATE SET
            cbu = EXCLUDED.cbu,
            cuil = EXCLUDED.cuil,
            alias = EXCLUDED.alias,
            owner_name = EXCLUDED.owner_name,
            updated_at = NOW();

        -- DELETE transfer_configs where null/empty or missing required fields (validated by JOIN)
        DELETE FROM transfer_configs tc
        WHERE tc.shop_payment_method_id IN (
            SELECT pm.spm_id
            FROM _pm_input pm
            JOIN shop_payment_methods spm ON spm.id = pm.spm_id AND spm.shop_id = p_shop_id
            WHERE pm.transfer_config IS NULL
               OR pm.transfer_config = 'null'::jsonb
               OR pm.transfer_config = '{}'::jsonb
               OR (pm.transfer_config->>'cbu') IS NULL
               OR (pm.transfer_config->>'cuil') IS NULL
               OR (pm.transfer_config->>'owner_name') IS NULL
        );

        -- UPSERT mercadopago_configs (validated by JOIN + required fields check)
        INSERT INTO mercadopago_configs (shop_payment_method_id, access_token, public_key, user_id)
        SELECT
            pm.spm_id,
            pm.mercadopago_config->>'access_token',
            pm.mercadopago_config->>'public_key',
            pm.mercadopago_config->>'user_id'
        FROM _pm_input pm
        JOIN shop_payment_methods spm ON spm.id = pm.spm_id AND spm.shop_id = p_shop_id
        WHERE pm.mercadopago_config IS NOT NULL
          AND pm.mercadopago_config != 'null'::jsonb
          AND pm.mercadopago_config != '{}'::jsonb
          AND (pm.mercadopago_config->>'access_token') IS NOT NULL
          AND (pm.mercadopago_config->>'public_key') IS NOT NULL
        ON CONFLICT (shop_payment_method_id) DO UPDATE SET
            access_token = EXCLUDED.access_token,
            public_key = EXCLUDED.public_key,
            user_id = EXCLUDED.user_id,
            updated_at = NOW();

        -- DELETE mercadopago_configs where null/empty or missing required fields (validated by JOIN)
        DELETE FROM mercadopago_configs mc
        WHERE mc.shop_payment_method_id IN (
            SELECT pm.spm_id
            FROM _pm_input pm
            JOIN shop_payment_methods spm ON spm.id = pm.spm_id AND spm.shop_id = p_shop_id
            WHERE pm.mercadopago_config IS NULL
               OR pm.mercadopago_config = 'null'::jsonb
               OR pm.mercadopago_config = '{}'::jsonb
               OR (pm.mercadopago_config->>'access_token') IS NULL
               OR (pm.mercadopago_config->>'public_key') IS NULL
        );
        -- _pm_input dropped automatically by ON COMMIT DROP
    END IF;

    -- =============================================
    -- 5. UPDATE delivery methods - OPTIMIZED
    -- Single JSON parse + shop_id validation via JOIN
    -- =============================================
    IF p_delivery_methods IS NOT NULL AND jsonb_array_length(p_delivery_methods) > 0 THEN
        -- Create temp table to materialize JSON ONCE
        CREATE TEMP TABLE _dm_input ON COMMIT DROP AS
        SELECT
            (j->>'id')::INTEGER AS sdm_id,
            (j->>'is_active')::BOOLEAN AS is_active,
            j->'delivery_config' AS delivery_config,
            j->'pickup_config' AS pickup_config,
            j->'delivery_zones' AS delivery_zones
        FROM jsonb_array_elements(p_delivery_methods) j;

        -- UPDATE is_active (validated by shop_id)
        UPDATE shop_delivery_methods sdm
        SET is_active = dm.is_active
        FROM _dm_input dm
        WHERE sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id;

        -- UPSERT delivery_configs (validated by JOIN)
        -- Note: fixed_price can be NULL (means use zones) or 0 (free shipping) or > 0 (fixed price)
        INSERT INTO delivery_configs (shop_delivery_method_id, fixed_price)
        SELECT
            dm.sdm_id,
            CASE
                WHEN dm.delivery_config->>'fixed_price' IS NULL OR dm.delivery_config->>'fixed_price' = '' THEN NULL
                ELSE (dm.delivery_config->>'fixed_price')::DECIMAL(10,2)
            END
        FROM _dm_input dm
        JOIN shop_delivery_methods sdm ON sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id
        WHERE dm.delivery_config IS NOT NULL
          AND dm.delivery_config != 'null'::jsonb
          AND dm.delivery_config != '{}'::jsonb
        ON CONFLICT (shop_delivery_method_id) DO UPDATE SET
            fixed_price = EXCLUDED.fixed_price,
            updated_at = NOW();

        -- DELETE delivery_configs where null/empty (validated by JOIN)
        DELETE FROM delivery_configs dc
        WHERE dc.shop_delivery_method_id IN (
            SELECT dm.sdm_id
            FROM _dm_input dm
            JOIN shop_delivery_methods sdm ON sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id
            WHERE dm.delivery_config IS NULL
               OR dm.delivery_config = 'null'::jsonb
               OR dm.delivery_config = '{}'::jsonb
        );

        -- UPSERT pickup_configs (validated by JOIN + required fields check)
        INSERT INTO pickup_configs (shop_delivery_method_id, address, city, province, postal_code, instructions)
        SELECT
            dm.sdm_id,
            dm.pickup_config->>'address',
            dm.pickup_config->>'city',
            dm.pickup_config->>'province',
            dm.pickup_config->>'postal_code',
            dm.pickup_config->>'instructions'
        FROM _dm_input dm
        JOIN shop_delivery_methods sdm ON sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id
        WHERE dm.pickup_config IS NOT NULL
          AND dm.pickup_config != 'null'::jsonb
          AND dm.pickup_config != '{}'::jsonb
          AND (dm.pickup_config->>'address') IS NOT NULL
          AND (dm.pickup_config->>'city') IS NOT NULL
        ON CONFLICT (shop_delivery_method_id) DO UPDATE SET
            address = EXCLUDED.address,
            city = EXCLUDED.city,
            province = EXCLUDED.province,
            postal_code = EXCLUDED.postal_code,
            instructions = EXCLUDED.instructions,
            updated_at = NOW();

        -- DELETE pickup_configs where null/empty or missing required fields (validated by JOIN)
        DELETE FROM pickup_configs pc
        WHERE pc.shop_delivery_method_id IN (
            SELECT dm.sdm_id
            FROM _dm_input dm
            JOIN shop_delivery_methods sdm ON sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id
            WHERE dm.pickup_config IS NULL
               OR dm.pickup_config = 'null'::jsonb
               OR dm.pickup_config = '{}'::jsonb
               OR (dm.pickup_config->>'address') IS NULL
               OR (dm.pickup_config->>'city') IS NULL
        );

        -- =============================================
        -- 5b. delivery_zones - OPTIMIZED with shop_id validation
        -- =============================================
        -- Expand zones from temp table (already validated sdm_ids)
        CREATE TEMP TABLE _zones_input ON COMMIT DROP AS
        SELECT
            dm.sdm_id,
            (z->>'id')::INTEGER AS zone_id,
            z->>'name' AS name,
            (z->>'price')::DECIMAL AS price
        FROM _dm_input dm
        JOIN shop_delivery_methods sdm ON sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id
        CROSS JOIN LATERAL jsonb_array_elements(COALESCE(dm.delivery_zones, '[]'::jsonb)) z;

        -- Get valid sdm_ids for this shop
        CREATE TEMP TABLE _valid_sdm_ids ON COMMIT DROP AS
        SELECT DISTINCT dm.sdm_id
        FROM _dm_input dm
        JOIN shop_delivery_methods sdm ON sdm.id = dm.sdm_id AND sdm.shop_id = p_shop_id;

        -- DELETE zones NOT in incoming list
        DELETE FROM delivery_zones dz
        WHERE dz.shop_delivery_method_id IN (SELECT sdm_id FROM _valid_sdm_ids)
          AND NOT EXISTS (
              SELECT 1 FROM _zones_input zi
              WHERE zi.sdm_id = dz.shop_delivery_method_id AND zi.zone_id = dz.id
          );

        -- UPDATE existing zones
        UPDATE delivery_zones dz
        SET name = zi.name, price = zi.price
        FROM _zones_input zi
        WHERE dz.id = zi.zone_id AND dz.shop_delivery_method_id = zi.sdm_id;

        -- INSERT new zones (where zone_id IS NULL)
        INSERT INTO delivery_zones (shop_delivery_method_id, name, price)
        SELECT zi.sdm_id, zi.name, zi.price
        FROM _zones_input zi
        WHERE zi.zone_id IS NULL;
        -- All temp tables dropped automatically by ON COMMIT DROP
    END IF;

    -- =============================================
    -- 6. REPLACE operating schedules (delete all + recreate)
    -- =============================================
    DELETE FROM operating_schedules WHERE shop_id = p_shop_id;

    IF p_operating_schedules IS NOT NULL AND jsonb_array_length(p_operating_schedules) > 0 THEN
        INSERT INTO operating_schedules (shop_id, day_of_week, open_time, close_time)
        SELECT
            p_shop_id,
            (os->>'day_of_week')::SMALLINT,
            (os->>'open_time')::TIME,
            (os->>'close_time')::TIME
        FROM jsonb_array_elements(p_operating_schedules) os;
    END IF;

    -- Return deleted storage refs for Cloudinary cleanup
    RETURN COALESCE(v_deleted_refs, '{}');

EXCEPTION
    WHEN unique_violation THEN
        RAISE;  -- Re-raise with original code 23505 for Go to handle
END;
$$ LANGUAGE plpgsql;

-- Function documentation
COMMENT ON FUNCTION update_shop IS
'Updates a shop with all related entities in a single database call.
OPTIMIZED V2:
- Temp tables to parse JSON ONCE per section
- JOIN with shop_id validation for multi-tenant security
- SET-BASED operations (no loops)

Parameters:
- p_shop_id: ID of the shop to update
- p_name, p_slug, p_email, p_phone, p_instagram: Shop basic fields
- p_images: JSONB array of images (logo, cover). IDs indicate existing, no ID = new.
- p_address: JSONB object for address (null to delete)
- p_payment_methods: JSONB array with id (shop_payment_method_id), is_active, and nested configs
- p_delivery_methods: JSONB array with id, is_active, and nested configs + zones
- p_operating_schedules: JSONB array of schedules (delete all + recreate)

Returns TEXT[] with storage_refs of deleted images for Cloudinary cleanup.

Security:
- All config operations JOIN with shop_payment_methods/shop_delivery_methods
- Validates that spm_id/sdm_id belongs to p_shop_id
- Prevents cross-shop data modification

Performance:
- JSON parsed once into temp tables
- Temp tables with ON COMMIT DROP (auto-cleanup)
- Batch operations for all updates/inserts/deletes';
