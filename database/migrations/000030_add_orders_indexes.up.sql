-- ============================================================================
-- Migration: Additional indexes for orders table
-- Optimizations for high-traffic multi-tenant e-commerce
-- ============================================================================

-- 1. Search by order number (customer support, tracking)
-- Note: order_number is unique per store (via orders_order_number_store_unique constraint)
-- This index helps when searching across all stores (admin) or with partial match
CREATE INDEX idx_orders_order_number ON public.orders (order_number);

-- 2. Search by customer phone (common in e-commerce, customers call/message)
CREATE INDEX idx_orders_customer_phone ON public.orders (customer_phone)
    WHERE customer_phone IS NOT NULL;

-- 3. Dashboard: pending orders requiring attention (filtered by status)
-- Partial index for active orders only - much smaller than full table
CREATE INDEX idx_orders_store_pending ON public.orders (store_id, created_at DESC)
    WHERE status IN ('pending', 'confirmed');

-- 4. Date range queries without store filter (admin reports)
CREATE INDEX idx_orders_created_at ON public.orders (created_at DESC);

-- 5. Completed orders for revenue reports (partial index)
CREATE INDEX idx_orders_store_completed ON public.orders (store_id, created_at DESC)
    WHERE status = 'completed';

-- Documentation
COMMENT ON INDEX idx_orders_order_number IS 'Search orders by order number (customer support)';
COMMENT ON INDEX idx_orders_customer_phone IS 'Search orders by customer phone';
COMMENT ON INDEX idx_orders_store_pending IS 'Partial index for active orders dashboard';
COMMENT ON INDEX idx_orders_created_at IS 'Date range queries across all stores (admin)';
COMMENT ON INDEX idx_orders_store_completed IS 'Partial index for revenue reports (completed only)';
