-- Activate payment methods for macdonalds shop
-- Payment methods: transfer (with config), cash

DO $$
    DECLARE
        v_shop_id bigint;
        v_transfer_pm_id bigint;
        v_cash_pm_id bigint;
        v_spm_transfer_id bigint;
    BEGIN
        -- Get macdonalds shop_id
        SELECT id INTO v_shop_id FROM public.shops WHERE slug = 'macdonalds' LIMIT 1;

        IF v_shop_id IS NULL THEN
            RAISE NOTICE 'Shop macdonalds not found, skipping payment methods setup';
            RETURN;
        END IF;

        -- Get payment method IDs
        SELECT id INTO v_transfer_pm_id FROM public.payment_methods WHERE code = 'transfer' LIMIT 1;
        SELECT id INTO v_cash_pm_id FROM public.payment_methods WHERE code = 'cash' LIMIT 1;

        -- Activate transfer payment method
        UPDATE public.shop_payment_methods
        SET is_active = true
        WHERE shop_id = v_shop_id AND payment_method_id = v_transfer_pm_id
        RETURNING id INTO v_spm_transfer_id;

        -- Activate cash payment method
        UPDATE public.shop_payment_methods
        SET is_active = true
        WHERE shop_id = v_shop_id AND payment_method_id = v_cash_pm_id;

        -- Add transfer config for macdonalds
        INSERT INTO public.transfer_configs (shop_payment_method_id, cbu, cuil, alias, owner_name)
        VALUES (v_spm_transfer_id, '0000003100000000000001', '20-12345678-9', 'MCDONALDS.PAGOS', 'McDonald''s Argentina S.A.')
        ON CONFLICT (shop_payment_method_id) DO UPDATE
        SET cbu = EXCLUDED.cbu,
            cuil = EXCLUDED.cuil,
            alias = EXCLUDED.alias,
            owner_name = EXCLUDED.owner_name,
            updated_at = now();

        RAISE NOTICE 'Payment methods activated for macdonalds (shop_id: %)', v_shop_id;
    END $$;
