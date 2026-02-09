-- ============================================================================
-- Migration: Additional indexes for order_items and order_item_selections
-- Optimizations for analytics and reporting
-- ============================================================================

-- =============================================================================
-- ORDER_ITEMS INDEXES
-- =============================================================================

-- 1. Product sales analytics: which products sell most in a period
-- Useful for: "Top 10 products this month"
CREATE INDEX idx_order_items_product_added ON public.order_items (product_id, added_at DESC)
    WHERE product_id IS NOT NULL;

-- 2. Promotional items tracking
-- Useful for: "How many promotional sales this week?"
CREATE INDEX idx_order_items_promotional ON public.order_items (added_at DESC)
    WHERE is_promotional = true;

-- =============================================================================
-- ORDER_ITEM_SELECTIONS INDEXES
-- =============================================================================

-- 1. Variant analytics: which variants are most popular
-- Useful for: "Most selected variants (Size, Color, etc.)"
CREATE INDEX idx_order_item_selections_variant_id
    ON public.order_item_selections (variant_id)
    WHERE variant_id IS NOT NULL;

-- 2. Combined variant+option analytics
-- Useful for: "Size:Large is selected 45% of the time"
CREATE INDEX idx_order_item_selections_variant_option
    ON public.order_item_selections (variant_id, option_id)
    WHERE variant_id IS NOT NULL AND option_id IS NOT NULL;

-- Documentation
COMMENT ON INDEX idx_order_items_product_added IS 'Product sales over time (analytics)';
COMMENT ON INDEX idx_order_items_promotional IS 'Promotional sales tracking (partial)';
COMMENT ON INDEX idx_order_item_selections_variant_id IS 'Variant popularity analytics';
COMMENT ON INDEX idx_order_item_selections_variant_option IS 'Variant+Option combination analytics';
