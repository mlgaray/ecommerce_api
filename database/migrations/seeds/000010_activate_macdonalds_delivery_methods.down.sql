-- Deactivate delivery methods for macdonalds shop

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
            RETURN;
        END IF;

        -- Get delivery method IDs
        SELECT id INTO v_delivery_dm_id FROM public.delivery_methods WHERE code = 'delivery' LIMIT 1;
        SELECT id INTO v_pickup_dm_id FROM public.delivery_methods WHERE code = 'pickup' LIMIT 1;

        -- Get shop_delivery_method_ids
        SELECT id INTO v_sdm_delivery_id FROM public.shop_delivery_methods
        WHERE shop_id = v_shop_id AND delivery_method_id = v_delivery_dm_id;

        SELECT id INTO v_sdm_pickup_id FROM public.shop_delivery_methods
        WHERE shop_id = v_shop_id AND delivery_method_id = v_pickup_dm_id;

        -- Delete delivery zones
        DELETE FROM public.delivery_zones WHERE shop_delivery_method_id = v_sdm_delivery_id;

        -- Delete pickup config
        DELETE FROM public.pickup_configs WHERE shop_delivery_method_id = v_sdm_pickup_id;

        -- Deactivate all delivery methods for macdonalds
        UPDATE public.shop_delivery_methods
        SET is_active = false
        WHERE shop_id = v_shop_id;
    END $$;
