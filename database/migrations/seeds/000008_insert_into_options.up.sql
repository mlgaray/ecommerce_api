DO $$
    DECLARE
product_id bigint;

BEGIN
SELECT id INTO product_id FROM products WHERE name = 'Big Mac' LIMIT 1;

INSERT INTO options (name, price, product_id)
VALUES ('Extra: Cheddar',2500,product_id),
    ('Extra: Bacon', 3000, product_id);


END $$;
