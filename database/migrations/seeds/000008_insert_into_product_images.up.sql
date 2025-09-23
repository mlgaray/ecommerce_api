-- Obtén el id del producto
DO $$
    DECLARE
        product_id bigint;
    BEGIN
        SELECT id INTO product_id FROM public.products WHERE name = 'Big Mac' LIMIT 1;

        -- Inserta los datos en la tabla product_images
        INSERT INTO product_images (url, product_id)
        VALUES
            ('https://i0.wp.com/imgs.hipertextual.com/wp-content/uploads/2016/07/14c33e7aa7e96918d15ac8eedf6dd466_large.jpeg?fit=1200%2C900&quality=55&strip=all&ssl=1', product_id),
            ('https://d1fd34dzzl09j.cloudfront.net/Images/MCDONALDS/Products/Desktop/11112_BigMac_832x472.jpg', product_id),
            ('https://cache-backend-mcd.mcdonaldscupones.com/media/image/product$kqX95Raf/200/200/original?country=ar', product_id);
    END $$;