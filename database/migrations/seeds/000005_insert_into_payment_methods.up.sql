INSERT INTO public.payment_methods (name, code, description, is_active)
VALUES
    ('Transferencia Bancaria', 'transfer', 'Pago mediante transferencia bancaria o CBU/CVU', true),
    ('MercadoPago', 'mercadopago', 'Pago mediante MercadoPago (tarjetas, dinero en cuenta)', true),
    ('Efectivo', 'cash', 'Pago en efectivo al momento de la entrega', true)
ON CONFLICT (code) DO NOTHING;
