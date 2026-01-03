-- Insert shop_payment_methods for all existing shops
-- Each shop gets all payment methods with is_active = false by default
INSERT INTO shop_payment_methods (shop_id, payment_method_id, is_active)
SELECT s.id, pm.id, false
FROM shops s
CROSS JOIN payment_methods pm
ON CONFLICT DO NOTHING;
