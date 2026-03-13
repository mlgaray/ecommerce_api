-- Insert a sample order for macdonalds shop
-- Order: Big Mac x2 with Grande size, Extra bacon, and Coca Cola
-- Payment: Cash (efectivo)
-- Delivery: Pickup ($0)
DO $$
DECLARE
    v_shop_id bigint;
    v_shop_name text;
    v_shop_slug text;
    v_payment_method_id bigint;
    v_delivery_method_id bigint;
BEGIN
    -- Get shop data
    SELECT id, name, slug INTO v_shop_id, v_shop_name, v_shop_slug
    FROM public.shops WHERE email = 'mcdonalds@mc.com' LIMIT 1;

    -- Get payment method (cash/efectivo)
    SELECT id INTO v_payment_method_id
    FROM public.payment_methods WHERE code = 'cash' LIMIT 1;

    -- Get delivery method (pickup)
    SELECT id INTO v_delivery_method_id
    FROM public.delivery_methods WHERE code = 'pickup' LIMIT 1;

    -- Insert the order
    -- Unit price calculation: 8500 (promo) + 1500 (Grande) + 1200 (Bacon) + 2000 (Coca) = 13200
    -- Total: 13200 x 2 = 26400
    INSERT INTO orders (
        order_number,
        store_id,
        store_name,
        store_slug,
        status,
        customer_name,
        customer_phone,
        customer_email,
        payment_method_id,
        payment_method_code,
        payment_method_name,
        delivery_method_id,
        delivery_method_code,
        delivery_method_name,
        subtotal,
        shipping_cost,
        total
    )
    VALUES (
        'ORD-260101-0000001',
        v_shop_id,
        v_shop_name,
        v_shop_slug,
        'pending',
        'Juan Pérez',
        '+5491112345678',
        'juan.perez@example.com',
        v_payment_method_id,
        'cash',
        'Efectivo',
        v_delivery_method_id,
        'pickup',
        'Retiro en local',
        26400,
        0,
        26400
    );
END $$;
