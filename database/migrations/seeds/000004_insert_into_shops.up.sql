-- Obtén el id del usuario
DO $$
    DECLARE
        user_id bigint;
    BEGIN
        SELECT id INTO user_id FROM public.users WHERE email = 'r.macdonalds@mc.com' LIMIT 1;

        -- Inserta los datos en la tabla shops
        INSERT INTO public.shops (name, user_id, slug, email, phone, image, instagram)
        VALUES ('McDonald´s', user_id, 'macdonalds','mcdonalds@mc.com', '123456789', 'https://upload.wikimedia.org/wikipedia/commons/thumb/4/4b/McDonald%27s_logo.svg/2560px-McDonald%27s_logo.svg.png', 'https://www.instagram.com/mcdonalds');
    END $$;