-- Insert order item selections (variant options) for the sample order
-- Selections: Grande (+1500), Extra bacon (+1200), Coca Cola (+2000)
DO $$
DECLARE
    v_order_item_id bigint;
    v_tamano_variant_id bigint;
    v_ingredientes_variant_id bigint;
    v_bebida_variant_id bigint;
    v_grande_option_id bigint;
    v_bacon_option_id bigint;
    v_coca_option_id bigint;
BEGIN
    -- Get the order item
    SELECT oi.id INTO v_order_item_id
    FROM order_items oi
    JOIN orders o ON o.id = oi.order_id
    JOIN shops s ON s.id = o.store_id
    WHERE s.email = 'mcdonalds@mc.com' AND o.order_number = '0001'
    LIMIT 1;

    -- Get variant IDs
    SELECT id INTO v_tamano_variant_id FROM product_variants WHERE name = 'Tamaño' LIMIT 1;
    SELECT id INTO v_ingredientes_variant_id FROM product_variants WHERE name = 'Ingredientes adicionales' LIMIT 1;
    SELECT id INTO v_bebida_variant_id FROM product_variants WHERE name = 'Bebida' LIMIT 1;

    -- Get option IDs
    SELECT id INTO v_grande_option_id FROM variant_options WHERE name = 'Grande' AND variant_id = v_tamano_variant_id LIMIT 1;
    SELECT id INTO v_bacon_option_id FROM variant_options WHERE name = 'Extra bacon' AND variant_id = v_ingredientes_variant_id LIMIT 1;
    SELECT id INTO v_coca_option_id FROM variant_options WHERE name = 'Coca Cola' AND variant_id = v_bebida_variant_id LIMIT 1;

    -- Insert selections
    INSERT INTO order_item_selections (
        order_item_id,
        variant_id,
        option_id,
        variant_name,
        option_name,
        option_price
    )
    VALUES
        -- Tamaño: Grande (+1500)
        (v_order_item_id, v_tamano_variant_id, v_grande_option_id, 'Tamaño', 'Grande', 1500),
        -- Ingrediente: Extra bacon (+1200)
        (v_order_item_id, v_ingredientes_variant_id, v_bacon_option_id, 'Ingredientes adicionales', 'Extra bacon', 1200),
        -- Bebida: Coca Cola (+2000)
        (v_order_item_id, v_bebida_variant_id, v_coca_option_id, 'Bebida', 'Coca Cola', 2000);
END $$;
