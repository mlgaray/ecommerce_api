-- Insert shop_delivery_methods for all existing shops
-- Each shop gets all delivery methods with is_active = false by default
INSERT INTO shop_delivery_methods (shop_id, delivery_method_id, is_active)
SELECT s.id, dm.id, false
FROM shops s
CROSS JOIN delivery_methods dm
ON CONFLICT DO NOTHING;
