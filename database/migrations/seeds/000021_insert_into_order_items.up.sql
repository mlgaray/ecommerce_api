-- Insert order items for the sample order
-- Big Mac x2 with options: Grande (+1500), Extra bacon (+1200), Coca Cola (+2000)
-- Unit price: 8500 (promo) + 1500 + 1200 + 2000 = 13200
-- Total: 13200 x 2 = 26400
DO $$
DECLARE
    v_order_id bigint;
    v_product_id bigint;
BEGIN
    -- Get the order
    SELECT o.id INTO v_order_id
    FROM orders o
    JOIN shops s ON s.id = o.store_id
    WHERE s.email = 'mcdonalds@mc.com' AND o.order_number = '0001'
    LIMIT 1;

    -- Get the product (Big Mac)
    SELECT id INTO v_product_id
    FROM products WHERE name = 'Big Mac' LIMIT 1;

    -- Insert the order item
    INSERT INTO order_items (
        order_id,
        product_id,
        product_name,
        product_image_url,
        base_price,
        is_promotional,
        promotional_price,
        quantity,
        unit_price,
        total_price
    )
    VALUES (
        v_order_id,
        v_product_id,
        'Big Mac',
        NULL,
        10000,          -- base_price
        true,           -- is_promotional
        8500,           -- promotional_price
        2,              -- quantity
        13200,          -- unit_price (8500 + 1500 + 1200 + 2000)
        26400           -- total_price (13200 x 2)
    );
END $$;
