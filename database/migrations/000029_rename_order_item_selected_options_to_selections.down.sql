-- ============================================================================
-- Rollback: Rename order_item_selections back to order_item_selected_options
-- ============================================================================

-- 1. Rename the table back
ALTER TABLE order_item_selections RENAME TO order_item_selected_options;

-- 2. Rename constraints back
ALTER TABLE order_item_selected_options
    RENAME CONSTRAINT order_item_selections_pkey TO order_item_selected_options_pkey;

ALTER TABLE order_item_selected_options
    RENAME CONSTRAINT order_item_selections_order_item_id_fkey TO order_item_selected_options_order_item_id_fkey;

ALTER TABLE order_item_selected_options
    RENAME CONSTRAINT order_item_selections_variant_id_fkey TO order_item_selected_options_variant_id_fkey;

ALTER TABLE order_item_selected_options
    RENAME CONSTRAINT order_item_selections_option_id_fkey TO order_item_selected_options_option_id_fkey;

ALTER TABLE order_item_selected_options
    RENAME CONSTRAINT order_item_selections_price_not_negative TO order_item_selected_options_price_not_negative;

-- 3. Rename indexes back
ALTER INDEX idx_order_item_selections_order_item_id RENAME TO idx_order_item_options_order_item_id;
ALTER INDEX idx_order_item_selections_option_id RENAME TO idx_order_item_options_option_id;

-- 4. Restore original comments
COMMENT ON TABLE order_item_selected_options IS 'Opciones de variantes seleccionadas por el cliente. Una fila por cada opción. Snapshot para preservar datos históricos.';
COMMENT ON COLUMN order_item_selected_options.option_price IS 'Precio adicional de la opción al momento de la orden. Suma al unit_price del item.';
