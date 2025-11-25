-- ============================================================================
-- ROLLBACK: Drop all indexes, extension, and products table
-- ============================================================================

-- Drop all product indexes
DROP INDEX IF EXISTS idx_products_search_spanish;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_shop_ordering;
DROP INDEX IF EXISTS idx_products_shop_id;
DROP INDEX IF EXISTS idx_products_name_trgm;

-- Drop table (CASCADE will handle any remaining dependencies)
DROP TABLE IF EXISTS products CASCADE;

-- Note: pg_trgm extension is NOT dropped as it may be used by other tables