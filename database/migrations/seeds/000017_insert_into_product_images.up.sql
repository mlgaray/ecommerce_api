-- Insert shop images (logo and cover)
DO $$
    DECLARE
        v_shop_id bigint;
    BEGIN
        SELECT id INTO v_shop_id FROM public.shops WHERE slug = 'macdonalds' LIMIT 1;

        INSERT INTO images (url, storage_ref, shop_id, type)
        VALUES
            ('https://res.cloudinary.com/dysydflnc/image/upload/v1774413137/3c880c20-c26d-46f6-b84c-78b4385cf158.png', '', v_shop_id, 'logo'),
            ('https://res.cloudinary.com/dysydflnc/image/upload/v1769054698/shop_1/images/oiqp2ggoazsglqhxwdet.png', '', v_shop_id, 'cover');
    END $$;

-- Insert product images
DO $$
    DECLARE
        v_product_id bigint;
    BEGIN
        SELECT id INTO v_product_id FROM public.products WHERE name = 'Big Mac' LIMIT 1;

        INSERT INTO images (url, storage_ref, product_id)
        VALUES
            ('https://i0.wp.com/imgs.hipertextual.com/wp-content/uploads/2016/07/14c33e7aa7e96918d15ac8eedf6dd466_large.jpeg?fit=1200%2C900&quality=55&strip=all&ssl=1', '', v_product_id),
            ('https://www.shutterstock.com/image-photo/lopburithailand-16-august-2023-big-600nw-2348868091.jpg', '', v_product_id),
            ('https://images.openfoodfacts.org/images/products/200/000/000/2603/front_fr.107.full.jpg', '', v_product_id);
    END $$;
