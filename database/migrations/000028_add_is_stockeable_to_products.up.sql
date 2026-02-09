-- ============================================================================
-- MIGRATION: Add is_stockeable column to products
-- Indicates whether the product manages stock or has unlimited availability
-- ============================================================================

ALTER TABLE products
ADD COLUMN is_stockeable boolean NOT NULL DEFAULT false;

COMMENT ON COLUMN products.is_stockeable IS 'true = product manages stock (validate availability), false = unlimited availability (no stock validation)';
