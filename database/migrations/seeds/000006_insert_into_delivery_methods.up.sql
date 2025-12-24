INSERT INTO public.delivery_methods (name, code, description, is_active)
VALUES
    ('Envío a domicilio', 'delivery', 'Envío a domicilio con costo por zona de cobertura', true),
    ('Retiro en tienda', 'pickup', 'Retiro en la dirección indicada por la tienda', true)
ON CONFLICT (code) DO NOTHING;
