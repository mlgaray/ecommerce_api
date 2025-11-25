-- ============================================================================
-- BULK INSERT: 500,000 COMPLETE Products (with images and variants)
-- ============================================================================
-- Purpose: Performance testing with realistic large dataset
-- Execution time: ~10-20 minutes (optimized with batch inserts)
--
-- ⚠️ IMPORTANT: Run this OUTSIDE of migrations
-- This script is too heavy for golang-migrate (causes timeout)
--
-- Usage:
--   psql -U your_user -d your_database -f bulk_insert_500k_products.sql
--
-- Or with connection string:
--   psql "postgresql://user:pass@host:5432/dbname" -f bulk_insert_500k_products.sql
--
-- Strategy:
-- 1. Insert 500k products using optimized INSERT ... SELECT (fast)
-- 2. Insert images using batch INSERT (3 images per product)
-- 3. Insert variants using batch INSERT (2 variants per product)
-- 4. Insert options using batch INSERT (4 options per variant)
-- 5. ANALYZE tables for query planner
--
-- Total objects created:
-- - 500,000 products
-- - 1,500,000 images (3 per product)
-- - 1,000,000 variants (2 per product)
-- - 4,000,000 options (4 per variant)
-- ============================================================================

-- Configuración para máxima velocidad
SET synchronous_commit = OFF; -- Acelera escrituras (safe para bulk insert)
SET maintenance_work_mem = '1GB'; -- Más memoria para índices
SET work_mem = '256MB';

BEGIN;

DO $$
DECLARE
    v_shop_id INT;
    v_shop_name TEXT;
    v_category_ids INT[];
    v_category_count INT;
    v_total_products INT := 500000;
    v_batch_size INT := 10000; -- Batch commits cada 10k
    v_start_time TIMESTAMP;
    v_elapsed_interval INTERVAL;
    v_min_product_id INT;
    v_max_product_id INT;
    v_products_inserted INT;
    v_images_inserted BIGINT;
    v_variants_inserted BIGINT;
    v_options_inserted BIGINT;

    -- Data arrays for variety
    v_name_prefixes TEXT[] := ARRAY[
        'Producto', 'Artículo', 'Item', 'Hamburguesa', 'Pizza', 'Bebida',
        'Laptop', 'Mouse', 'Teclado', 'Monitor', 'Auriculares', 'Cámara',
        'Teléfono', 'Tablet', 'Reloj', 'Zapatos', 'Remera', 'Pantalón'
    ];

    v_name_suffixes TEXT[] := ARRAY[
        'Clásico', 'Premium', 'Pro', 'Max', 'Ultra', 'Plus', 'Lite',
        'Gaming', 'Professional', 'Deluxe', 'Special', 'Limited'
    ];

    v_descriptions TEXT[] := ARRAY[
        'Producto de alta calidad con características premium.',
        'Diseño moderno y funcional para uso diario.',
        'Fabricado con materiales de primera calidad.',
        'Perfecta combinación de estilo y rendimiento.',
        'Tecnología de punta al mejor precio.',
        'Ideal para profesionales y entusiastas.',
        'Garantía de satisfacción incluida.',
        'El mejor de su categoría en el mercado.'
    ];
BEGIN
    v_start_time := clock_timestamp();

    RAISE NOTICE '';
    RAISE NOTICE '========================================================================';
    RAISE NOTICE '                  BULK INSERT - 500K PRODUCTS';
    RAISE NOTICE '========================================================================';
    RAISE NOTICE 'Start time: %', v_start_time;
    RAISE NOTICE '';

    -- Buscar primer shop disponible
    SELECT id, name INTO v_shop_id, v_shop_name
    FROM shops
    ORDER BY id
    LIMIT 1;

    IF v_shop_id IS NULL THEN
        RAISE EXCEPTION 'No shops found in database. Please run: make migrate-up';
    END IF;

    -- Obtener categorías
    SELECT array_agg(id) INTO v_category_ids
    FROM categories
    LIMIT 20; -- Usar máximo 20 categorías

    v_category_count := array_length(v_category_ids, 1);

    IF v_category_count IS NULL OR v_category_count = 0 THEN
        RAISE EXCEPTION 'No categories found. Run category seeds first.';
    END IF;

    RAISE NOTICE 'Configuration:';
    RAISE NOTICE '  Shop ID: % (%)', v_shop_id, v_shop_name;
    RAISE NOTICE '  Categories: %', v_category_count;
    RAISE NOTICE '  Products to insert: %', v_total_products;
    RAISE NOTICE '  Batch size: %', v_batch_size;
    RAISE NOTICE '';

    -- ========================================================================
    -- STEP 1: Insert 500k products (FAST bulk insert)
    -- ========================================================================
    RAISE NOTICE '[1/4] Inserting % products...', v_total_products;

    INSERT INTO products (
        name,
        description,
        price,
        stock,
        minimum_stock,
        is_active,
        is_promotional,
        promotional_price,
        is_highlighted,
        category_id,
        shop_id,
        created_at
    )
    SELECT
        -- Name: prefix + number + suffix
        v_name_prefixes[1 + floor(random() * array_length(v_name_prefixes, 1))::int] || ' ' ||
        seq::text || ' ' ||
        v_name_suffixes[1 + floor(random() * array_length(v_name_suffixes, 1))::int],

        -- Description
        v_descriptions[1 + floor(random() * array_length(v_descriptions, 1))::int] ||
        ' SKU: ' || lpad(seq::text, 8, '0'),

        -- Price: $10 - $5000
        (10 + random() * 4990)::NUMERIC(10,2),

        -- Stock: 0-500
        floor(random() * 501)::int,

        -- Minimum stock: 5-20
        (5 + floor(random() * 16))::int,

        -- is_active: 80% active
        random() > 0.2,

        -- is_promotional: 25% promotional
        random() < 0.25,

        -- promotional_price: 10-30% discount
        CASE
            WHEN random() < 0.25 THEN
                ((10 + random() * 4990) * (0.7 + random() * 0.2))::NUMERIC(10,2)
            ELSE NULL
        END,

        -- is_highlighted: 10% highlighted
        random() < 0.1,

        -- category_id: random from available
        v_category_ids[1 + floor(random() * v_category_count)::int],

        -- shop_id
        v_shop_id,

        -- created_at: distributed over last 365 days
        NOW() - (random() * INTERVAL '365 days')

    FROM generate_series(1, v_total_products) seq;

    GET DIAGNOSTICS v_products_inserted = ROW_COUNT;

    v_elapsed_interval := clock_timestamp() - v_start_time;
    RAISE NOTICE '  ✓ Inserted % products in %', v_products_inserted, v_elapsed_interval;
    RAISE NOTICE '  Speed: % products/second',
        (v_products_inserted / GREATEST(EXTRACT(EPOCH FROM v_elapsed_interval), 0.1))::int;
    RAISE NOTICE '';

    -- Get product ID range for next steps
    SELECT MIN(id), MAX(id) INTO v_min_product_id, v_max_product_id
    FROM products
    WHERE shop_id = v_shop_id;

    -- ========================================================================
    -- STEP 2: Insert images (3 per product = 1.5M images)
    -- ========================================================================
    RAISE NOTICE '[2/4] Inserting product images (3 per product)...';

    INSERT INTO product_images (product_id, url)
    SELECT
        product_id,
        'https://picsum.photos/800/600?random=' || product_id || '-' || img_num
    FROM (
        SELECT id as product_id
        FROM products
        WHERE shop_id = v_shop_id
    ) p
    CROSS JOIN generate_series(1, 3) img_num;

    GET DIAGNOSTICS v_images_inserted = ROW_COUNT;

    v_elapsed_interval := clock_timestamp() - v_start_time;
    RAISE NOTICE '  ✓ Inserted % images in %', v_images_inserted, v_elapsed_interval;
    RAISE NOTICE '';

    -- ========================================================================
    -- STEP 3: Insert variants (2 per product = 1M variants)
    -- ========================================================================
    RAISE NOTICE '[3/4] Inserting product variants (2 per product)...';

    INSERT INTO product_variants (product_id, name, "order", selection_type, max_selections)
    SELECT
        product_id,
        CASE variant_num
            WHEN 1 THEN 'Color'
            WHEN 2 THEN 'Tamaño'
        END,
        variant_num,
        'single',
        1
    FROM (
        SELECT id as product_id
        FROM products
        WHERE shop_id = v_shop_id
    ) p
    CROSS JOIN generate_series(1, 2) variant_num;

    GET DIAGNOSTICS v_variants_inserted = ROW_COUNT;

    v_elapsed_interval := clock_timestamp() - v_start_time;
    RAISE NOTICE '  ✓ Inserted % variants in %', v_variants_inserted, v_elapsed_interval;
    RAISE NOTICE '';

    -- ========================================================================
    -- STEP 4: Insert variant options (4 per variant = 4M options)
    -- ========================================================================
    RAISE NOTICE '[4/4] Inserting variant options (4 per variant)...';

    INSERT INTO variant_options (variant_id, name, price, "order")
    SELECT
        v.id as variant_id,
        CASE
            WHEN v.name = 'Color' THEN
                CASE opt_num
                    WHEN 1 THEN 'Negro'
                    WHEN 2 THEN 'Blanco'
                    WHEN 3 THEN 'Rojo'
                    WHEN 4 THEN 'Azul'
                END
            WHEN v.name = 'Tamaño' THEN
                CASE opt_num
                    WHEN 1 THEN 'S'
                    WHEN 2 THEN 'M'
                    WHEN 3 THEN 'L'
                    WHEN 4 THEN 'XL'
                END
        END,
        (random() * 50)::NUMERIC(10,2), -- Extra price $0-$50
        opt_num
    FROM product_variants v
    CROSS JOIN generate_series(1, 4) opt_num
    WHERE v.product_id IN (
        SELECT id FROM products WHERE shop_id = v_shop_id
    );

    GET DIAGNOSTICS v_options_inserted = ROW_COUNT;

    v_elapsed_interval := clock_timestamp() - v_start_time;
    RAISE NOTICE '  ✓ Inserted % options in %', v_options_inserted, v_elapsed_interval;
    RAISE NOTICE '';

    -- ========================================================================
    -- STEP 5: Update statistics (CRITICAL for query performance)
    -- ========================================================================
    RAISE NOTICE '[5/5] Updating table statistics (ANALYZE)...';
    ANALYZE products;
    ANALYZE product_images;
    ANALYZE product_variants;
    ANALYZE variant_options;
    RAISE NOTICE '  ✓ Statistics updated';
    RAISE NOTICE '';

    -- ========================================================================
    -- FINAL REPORT
    -- ========================================================================
    v_elapsed_interval := clock_timestamp() - v_start_time;

    RAISE NOTICE '========================================================================';
    RAISE NOTICE '                         BULK INSERT COMPLETED';
    RAISE NOTICE '========================================================================';
    RAISE NOTICE 'Total time: %', v_elapsed_interval;
    RAISE NOTICE '';
    RAISE NOTICE 'Objects created:';
    RAISE NOTICE '  Products:        %', v_products_inserted;
    RAISE NOTICE '  Images:          %', v_images_inserted;
    RAISE NOTICE '  Variants:        %', v_variants_inserted;
    RAISE NOTICE '  Options:         %', v_options_inserted;
    RAISE NOTICE '  TOTAL:           %', v_products_inserted + v_images_inserted + v_variants_inserted + v_options_inserted;
    RAISE NOTICE '';
    RAISE NOTICE 'Average insertion speed:';
    RAISE NOTICE '  % objects/second',
        ((v_products_inserted + v_images_inserted + v_variants_inserted + v_options_inserted) /
         GREATEST(EXTRACT(EPOCH FROM v_elapsed_interval), 0.1))::int;
    RAISE NOTICE '========================================================================';
    RAISE NOTICE '';

END $$;

COMMIT;

-- Restore settings
RESET synchronous_commit;
RESET maintenance_work_mem;
RESET work_mem;

-- Show final statistics
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '=== Final Table Sizes ===';
    RAISE NOTICE 'products: % rows, %',
        (SELECT count(*) FROM products WHERE shop_id = 1),
        pg_size_pretty(pg_total_relation_size('products'));
    RAISE NOTICE 'product_images: % rows, %',
        (SELECT count(*) FROM product_images),
        pg_size_pretty(pg_total_relation_size('product_images'));
    RAISE NOTICE 'product_variants: % rows, %',
        (SELECT count(*) FROM product_variants),
        pg_size_pretty(pg_total_relation_size('product_variants'));
    RAISE NOTICE 'variant_options: % rows, %',
        (SELECT count(*) FROM variant_options),
        pg_size_pretty(pg_total_relation_size('variant_options'));
    RAISE NOTICE '';
    RAISE NOTICE '✓ Script completed successfully!';
    RAISE NOTICE '  You can now test search performance with 500k products';
    RAISE NOTICE '';
END $$;
