-- Obtén el id de la tienda
DO $$
    DECLARE
shop_id bigint;
category_id bigint;
BEGIN
SELECT id INTO shop_id FROM public.shops WHERE email = 'mcdonalds@mc.com' LIMIT 1;
SELECT id INTO category_id FROM public.categories WHERE name = 'Hamburguesas' LIMIT 1;

-- Inserta los datos en la tabla categories
INSERT INTO products (name, description, image, price, is_active, category_id,shop_id)
VALUES ('Big Mac',
        'La perfección hecha hamburguesa que te hace agua la boca comienza con dos patties de 100% carne y la salsa Big Mac.',
        'https://i0.wp.com/imgs.hipertextual.com/wp-content/uploads/2016/07/14c33e7aa7e96918d15ac8eedf6dd466_large.jpeg?fit=1200%2C900&quality=55&strip=all&ssl=1',
        10000,
        true,
        category_id,
        shop_id);
END $$;