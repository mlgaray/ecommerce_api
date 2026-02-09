-- ============================================================================
-- Migration: Rename order_item_selected_options to order_item_selections
-- Reason: Cleaner name, reflects that it stores variant+option selections
-- ============================================================================

-- 1. Rename the table
ALTER TABLE order_item_selected_options RENAME TO order_item_selections;

-- 2. Rename constraints
ALTER TABLE order_item_selections
    RENAME CONSTRAINT order_item_selected_options_pkey TO order_item_selections_pkey;

ALTER TABLE order_item_selections
    RENAME CONSTRAINT order_item_selected_options_order_item_id_fkey TO order_item_selections_order_item_id_fkey;

ALTER TABLE order_item_selections
    RENAME CONSTRAINT order_item_selected_options_variant_id_fkey TO order_item_selections_variant_id_fkey;

ALTER TABLE order_item_selections
    RENAME CONSTRAINT order_item_selected_options_option_id_fkey TO order_item_selections_option_id_fkey;

ALTER TABLE order_item_selections
    RENAME CONSTRAINT order_item_selected_options_price_not_negative TO order_item_selections_price_not_negative;

-- 3. Rename indexes
ALTER INDEX idx_order_item_options_order_item_id RENAME TO idx_order_item_selections_order_item_id;
ALTER INDEX idx_order_item_options_option_id RENAME TO idx_order_item_selections_option_id;

-- 4. Update comments
COMMENT ON TABLE order_item_selections IS 'Variant and option selections for each order item. One row per selected option. Snapshots preserve historical data.';
COMMENT ON COLUMN order_item_selections.option_price IS 'Additional price of the option at order time. Adds to the item unit_price.';
