-- ============================================================================
-- Rollback: Remove additional order_items and order_item_selections indexes
-- ============================================================================

-- Order items indexes
DROP INDEX IF EXISTS idx_order_items_product_added;
DROP INDEX IF EXISTS idx_order_items_promotional;

-- Order item selections indexes
DROP INDEX IF EXISTS idx_order_item_selections_variant_id;
DROP INDEX IF EXISTS idx_order_item_selections_variant_option;
