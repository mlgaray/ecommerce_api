-- Insert shop without user_id (ownership via staff)
INSERT INTO public.shops (name, slug, email, phone, instagram, primary_color)
VALUES ('McDonald´s', 'macdonalds', 'mcdonalds@mc.com', '123456789', 'https://www.instagram.com/mcdonalds', '#F30808');

-- Create staff entry for the owner of the shop
DO $$
    DECLARE
        v_user_id bigint;
        v_shop_id bigint;
        v_owner_role_id bigint;
        v_staff_id bigint;
    BEGIN
        SELECT id INTO v_user_id FROM public.users WHERE email = 'r.macdonalds@mc.com' LIMIT 1;
        SELECT id INTO v_shop_id FROM public.shops WHERE slug = 'macdonalds' LIMIT 1;
        SELECT id INTO v_owner_role_id FROM public.roles WHERE name = 'owner' LIMIT 1;

        INSERT INTO public.staff (user_id, shop_id, is_active)
        VALUES (v_user_id, v_shop_id, true)
        RETURNING id INTO v_staff_id;

        INSERT INTO public.staff_roles (staff_id, role_id)
        VALUES (v_staff_id, v_owner_role_id);
    END $$;
