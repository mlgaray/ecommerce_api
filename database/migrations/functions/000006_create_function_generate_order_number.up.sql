
-- ============================================================================
-- SEQUENCE: order_number_seq
-- Global sequence for order number generation. Avoids MAX(order_number) + regex
-- parsing which breaks with padding overflow (>99,999).
-- ============================================================================
CREATE SEQUENCE IF NOT EXISTS order_number_seq;

-- Seed sequence from existing orders to avoid collisions.
-- With data: setval(MAX, true)  → next nextval() returns MAX+1
-- Empty DB:  setval(1, false)   → next nextval() returns 1 (not 2)
WITH max_id AS (
    SELECT COALESCE(MAX(id), 0) AS val FROM orders
)
SELECT setval('order_number_seq', GREATEST(val, 1), val > 0)
FROM max_id;

-- ============================================================================
-- FUNCTION: generate_order_number
-- Genera un número de orden usando SEQUENCE global + fecha.
-- Formato: ORD-YYMMDD-0000001 (ej: ORD-260302-0000014)
--   - YYMMDD: año corto, orden cronológico lexicográfico (250101 < 260101)
--   - Zero-padded a 7 dígitos (hasta 9.999.999 órdenes)
-- ============================================================================
CREATE OR REPLACE FUNCTION generate_order_number(p_store_id bigint)
RETURNS text AS $$
BEGIN
    RETURN 'ORD-' || to_char(now(), 'YYMMDD') || '-' || lpad(nextval('order_number_seq')::text, 7, '0');
END;
$$ LANGUAGE plpgsql
SET search_path = public;

COMMENT ON FUNCTION generate_order_number(bigint) IS 'Genera número de orden con SEQUENCE global + fecha YYMMDD. Formato: ORD-YYMMDD-0000001. 7 dígitos, orden lexicográfico = cronológico.';
