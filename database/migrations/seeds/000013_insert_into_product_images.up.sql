-- Insert shop images (logo and cover)
DO $$
    DECLARE
        v_shop_id bigint;
    BEGIN
        SELECT id INTO v_shop_id FROM public.shops WHERE slug = 'macdonalds' LIMIT 1;

        INSERT INTO images (url, storage_ref, shop_id, type)
        VALUES
            ('https://upload.wikimedia.org/wikipedia/commons/thumb/4/4b/McDonald%27s_logo.svg/2560px-McDonald%27s_logo.svg.png', '', v_shop_id, 'logo'),
            ('https://cdn.britannica.com/67/231567-050-DB21B46A/McDonalds-restaurant-Taiwan.jpg', '', v_shop_id, 'cover');
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
            ('https://d1fd34dzzl09j.cloudfront.net/Images/MCDONALDS/Products/Desktop/11112_BigMac_832x472.jpg', '', v_product_id),
            ('https://cache-backend-mcd.mcdonaldscupones.com/media/image/product$kqX95Raf/200/200/original?country=ar', '', v_product_id);
    END $$;
