-- Obtén el id de la tienda
DO $$
    DECLARE
        shop_id bigint;
    BEGIN
        SELECT id INTO shop_id FROM public.shops WHERE email = 'mcdonalds@mc.com' LIMIT 1;

        -- Inserta los datos en la tabla categories
        INSERT INTO categories (name, description, image, shop_id)
        VALUES ('Hamburguesas', 'Deliciosas hamburguesas con ingredientes frescos', 'https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSgTOxxZToIChqO4GH5ZeQ7_lXpSngK4HVBUg&s', shop_id);
    END $$;