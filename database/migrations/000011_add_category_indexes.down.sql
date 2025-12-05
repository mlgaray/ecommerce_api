-- Rollback: Remove category search indexes

DROP INDEX IF EXISTS idx_categories_shop_created_at;
DROP INDEX IF EXISTS idx_categories_name_trgm;

-- Note: We don't drop pg_trgm extension as it may be used by other indexes
