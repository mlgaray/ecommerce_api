
-- ============================================================================
-- FUNCTION: generate_order_number
-- Genera un número de orden legible y secuencial por store
-- Formato: ORD-YYYY-NNNNN (ej: ORD-2024-00001)
-- Reinicia la secuencia cada año por store
-- ============================================================================
CREATE OR REPLACE FUNCTION generate_order_number(p_store_id bigint)
RETURNS text AS $$
DECLARE
    v_count bigint;
    v_year text;
BEGIN
    v_year := to_char(now(), 'YYYY');

    SELECT COUNT(*) + 1 INTO v_count
    FROM orders
    WHERE store_id = p_store_id
      AND created_at >= date_trunc('year', now());

    RETURN 'ORD-' || v_year || '-' || lpad(v_count::text, 5, '0');
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION generate_order_number(bigint) IS 'Genera número de orden secuencial por store. Formato: ORD-YYYY-NNNNN. Reinicia cada año.';
