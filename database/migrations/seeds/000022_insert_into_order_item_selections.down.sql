-- Delete order item selections for macdonalds orders
DELETE FROM order_item_selections WHERE order_item_id IN (
    SELECT oi.id FROM order_items oi
    JOIN orders o ON o.id = oi.order_id
    JOIN shops s ON s.id = o.store_id
    WHERE s.email = 'mcdonalds@mc.com'
);
