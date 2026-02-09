-- ============================================================================
-- ROLLBACK: Remove is_stockeable column from products
-- ============================================================================

ALTER TABLE products
DROP COLUMN IF EXISTS is_stockeable;
