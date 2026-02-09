-- Deactivate payment methods for macdonalds shop

DO $$
    DECLARE
        v_shop_id bigint;
        v_transfer_pm_id bigint;
        v_spm_transfer_id bigint;
    BEGIN
        -- Get macdonalds shop_id
        SELECT id INTO v_shop_id FROM public.shops WHERE slug = 'macdonalds' LIMIT 1;

        IF v_shop_id IS NULL THEN
            RETURN;
        END IF;

        -- Get transfer payment method ID
        SELECT id INTO v_transfer_pm_id FROM public.payment_methods WHERE code = 'transfer' LIMIT 1;

        -- Get shop_payment_method_id for transfer
        SELECT id INTO v_spm_transfer_id FROM public.shop_payment_methods
        WHERE shop_id = v_shop_id AND payment_method_id = v_transfer_pm_id;

        -- Delete transfer config
        DELETE FROM public.transfer_configs WHERE shop_payment_method_id = v_spm_transfer_id;

        -- Deactivate all payment methods for macdonalds
        UPDATE public.shop_payment_methods
        SET is_active = false
        WHERE shop_id = v_shop_id;
    END $$;
