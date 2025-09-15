DO $$
    DECLARE
product_id bigint;

BEGIN
SELECT id INTO product_id FROM products WHERE name = 'Big Mac' LIMIT 1;

INSERT INTO variants (name, product_id)
VALUES ('Combo chico',product_id),
       ('Combo mediano', product_id),
        ('Combo grande', product_id);


END $$;