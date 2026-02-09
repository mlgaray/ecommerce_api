-- Activate delivery methods for macdonalds shop
-- Delivery methods: delivery (with zones), pickup (with config)

DO $$
    DECLARE
        v_shop_id bigint;
        v_delivery_dm_id bigint;
        v_pickup_dm_id bigint;
        v_sdm_delivery_id bigint;
        v_sdm_pickup_id bigint;
    BEGIN
        -- Get macdonalds shop_id
        SELECT id INTO v_shop_id FROM public.shops WHERE slug = 'macdonalds' LIMIT 1;

        IF v_shop_id IS NULL THEN
            RAISE NOTICE 'Shop macdonalds not found, skipping delivery methods setup';
            RETURN;
        END IF;

        -- Get delivery method IDs
        SELECT id INTO v_delivery_dm_id FROM public.delivery_methods WHERE code = 'delivery' LIMIT 1;
        SELECT id INTO v_pickup_dm_id FROM public.delivery_methods WHERE code = 'pickup' LIMIT 1;

        -- Activate delivery method
        UPDATE public.shop_delivery_methods
        SET is_active = true
        WHERE shop_id = v_shop_id AND delivery_method_id = v_delivery_dm_id
        RETURNING id INTO v_sdm_delivery_id;

        -- Activate pickup method
        UPDATE public.shop_delivery_methods
        SET is_active = true
        WHERE shop_id = v_shop_id AND delivery_method_id = v_pickup_dm_id
        RETURNING id INTO v_sdm_pickup_id;

        -- Add delivery zones for macdonalds (zone-based pricing)
        INSERT INTO public.delivery_zones (shop_delivery_method_id, name, price)
        VALUES
            (v_sdm_delivery_id, 'Centro', 200.00),
            (v_sdm_delivery_id, 'Zona Norte', 350.00),
            (v_sdm_delivery_id, 'Zona Sur', 350.00),
            (v_sdm_delivery_id, 'Zona Oeste', 400.00)
        ON CONFLICT DO NOTHING;

        -- Add pickup config for macdonalds
        INSERT INTO public.pickup_configs (shop_delivery_method_id, address, city, province, postal_code, instructions)
        VALUES (
            v_sdm_pickup_id,
            'Av. Corrientes 1234',
            'Buenos Aires',
            'CABA',
            'C1043',
            'Retirar en mostrador. Presentar número de pedido. Horario: Lun-Dom 10:00-23:00'
        )
        ON CONFLICT (shop_delivery_method_id) DO UPDATE
        SET address = EXCLUDED.address,
            city = EXCLUDED.city,
            province = EXCLUDED.province,
            postal_code = EXCLUDED.postal_code,
            instructions = EXCLUDED.instructions,
            updated_at = now();

        RAISE NOTICE 'Delivery methods activated for macdonalds (shop_id: %)', v_shop_id;
    END $$;
