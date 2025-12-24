-- Obtén el id de la tienda
DO $$
    DECLARE
        shop_id bigint;
    BEGIN
        SELECT id INTO shop_id FROM public.shops WHERE email = 'mcdonalds@mc.com' LIMIT 1;

        -- Inserta los datos en la tabla addresses
        INSERT INTO addresses (text, place_id, ltd, lng, shop_id)
        VALUES ('Address Text', 'PlaceID', 40.712776, -74.005974, shop_id);
    END $$;