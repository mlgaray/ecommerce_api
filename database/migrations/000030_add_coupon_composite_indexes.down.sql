-- Restore original single-column indexes
CREATE INDEX IF NOT EXISTS idx_coupons_shop_id ON public.coupons (shop_id);
CREATE INDEX IF NOT EXISTS idx_coupons_code ON public.coupons (code);
CREATE INDEX IF NOT EXISTS idx_coupon_usages_phone ON public.coupon_usages (phone);

-- Drop composite indexes
DROP INDEX IF EXISTS idx_coupons_shop_created_id;
DROP INDEX IF EXISTS idx_coupons_shop_code_id;
DROP INDEX IF EXISTS idx_coupon_usages_coupon_phone;
