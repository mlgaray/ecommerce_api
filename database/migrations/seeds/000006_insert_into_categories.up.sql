-- Obtén el id de la tienda e inserta la categoría con su imagen
DO $$
    DECLARE
        v_shop_id bigint;
        v_category_id bigint;
    BEGIN
        SELECT id INTO v_shop_id FROM public.shops WHERE email = 'mcdonalds@mc.com' LIMIT 1;

        -- Inserta la categoría (sin columna image)
        INSERT INTO categories (name, description, shop_id)
        VALUES ('Hamburguesas', 'Deliciosas hamburguesas con ingredientes frescos', v_shop_id)
        RETURNING id INTO v_category_id;

        -- Inserta la imagen en la tabla images
        INSERT INTO images (url, storage_ref, category_id)
        VALUES ('https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcSgTOxxZToIChqO4GH5ZeQ7_lXpSngK4HVBUg&s', '', v_category_id);
    END $$;