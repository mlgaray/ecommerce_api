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
        VALUES ('https://res.cloudinary.com/dysydflnc/image/upload/v1773535384/shop_1/categories/kwhqbfm4fxjvej4n9mub.png', '', v_category_id);
    END $$;