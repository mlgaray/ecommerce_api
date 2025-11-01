-- Obtén los ids de las variantes del producto
DO $$
    DECLARE
        tamano_variant_id bigint;
        ingredientes_variant_id bigint;
        bebida_variant_id bigint;
    BEGIN
        SELECT id INTO tamano_variant_id FROM public.product_variants WHERE name = 'Tamaño' LIMIT 1;
        SELECT id INTO ingredientes_variant_id FROM public.product_variants WHERE name = 'Ingredientes adicionales' LIMIT 1;
        SELECT id INTO bebida_variant_id FROM public.product_variants WHERE name = 'Bebida' LIMIT 1;

        -- Inserta los datos en la tabla variant_options
        INSERT INTO variant_options (name, price, "order", variant_id)
        VALUES
            -- Opciones para Tamaño
            ('Mediano', 0.0, 1, tamano_variant_id),
            ('Grande', 1500.0, 2, tamano_variant_id),
            ('Extra Grande', 2500.0, 3, tamano_variant_id),

            -- Opciones para Ingredientes adicionales
            ('Extra queso', 800.0, 1, ingredientes_variant_id),
            ('Extra bacon', 1200.0, 2, ingredientes_variant_id),
            ('Extra cebolla', 500.0, 3, ingredientes_variant_id),
            ('Pepinillos extra', 300.0, 4, ingredientes_variant_id),

            -- Opciones para Bebida
            ('Coca Cola', 2000.0, 1, bebida_variant_id),
            ('Sprite', 2000.0, 2, bebida_variant_id),
            ('Fanta', 2000.0, 3, bebida_variant_id),
            ('Agua', 1500.0, 4, bebida_variant_id);
    END $$;