-- Delete orders for macdonalds shop
-- CASCADE will automatically delete order_items and order_item_selections
DELETE FROM orders WHERE store_id IN (
    SELECT id FROM shops WHERE email = 'mcdonalds@mc.com'
);
