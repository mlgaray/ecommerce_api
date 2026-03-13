-- ============================================================================
-- Composite indexes for coupon pagination and lookup patterns
-- ============================================================================

-- Keyset pagination: ORDER BY created_at DESC, id DESC WHERE shop_id = $1
-- Covers the default coupon listing query
CREATE INDEX idx_coupons_shop_created_id ON public.coupons (shop_id, created_at DESC, id DESC);

-- Keyset pagination: ORDER BY code ASC/DESC, id ASC/DESC WHERE shop_id = $1
-- Covers coupon listing sorted by code
CREATE INDEX idx_coupons_shop_code_id ON public.coupons (shop_id, code, id);

-- Coupon validation lookup: WHERE code = $1 AND shop_id = $2
-- Already covered by coupons_code_shop_id_unique (UNIQUE creates implicit index)
-- No additional index needed

-- Usage count by coupon: WHERE coupon_id = $1
-- Already covered by idx_coupon_usages_coupon_id

-- Usage count by coupon + phone: WHERE coupon_id = $1 AND phone = $2
-- Composite index for per-phone usage limit validation
CREATE INDEX idx_coupon_usages_coupon_phone ON public.coupon_usages (coupon_id, phone);

-- Drop redundant single-column indexes now covered by composites
DROP INDEX IF EXISTS idx_coupons_shop_id;
DROP INDEX IF EXISTS idx_coupons_code;
DROP INDEX IF EXISTS idx_coupon_usages_phone;
