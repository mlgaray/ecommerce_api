-- Delete order items for macdonalds orders
-- Note: This will also CASCADE delete order_item_selections
DELETE FROM order_items WHERE order_id IN (
    SELECT o.id FROM orders o
    JOIN shops s ON s.id = o.store_id
    WHERE s.email = 'mcdonalds@mc.com'
);
