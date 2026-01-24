-- Obtén el id del usuario
DO $$
    DECLARE
        user_id bigint;
    BEGIN
        SELECT id INTO user_id FROM public.users WHERE email = 'r.macdonalds@mc.com' LIMIT 1;

        -- Inserta los datos en la tabla shops
        INSERT INTO public.shops (name, user_id, slug, email, phone, instagram, primary_color)
        VALUES ('McDonald´s', user_id, 'macdonalds','mcdonalds@mc.com', '123456789', 'https://www.instagram.com/mcdonalds', '#F30808');
    END $$;