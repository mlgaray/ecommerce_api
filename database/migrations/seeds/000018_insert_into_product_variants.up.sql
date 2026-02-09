-- Obtén el id del producto
DO $$
    DECLARE
        product_id bigint;
    BEGIN
        SELECT id INTO product_id FROM public.products WHERE name = 'Big Mac' LIMIT 1;

        -- Inserta los datos en la tabla product_variants
        INSERT INTO product_variants (name, "order", selection_type, max_selections, is_required, product_id)
        VALUES
            ('Tamaño', 1, 'single', 1, true, product_id),
            ('Ingredientes adicionales', 2, 'multiple', 3, false, product_id),
            ('Bebida', 3, 'single', 1, false, product_id);
    END $$;