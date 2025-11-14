-- Rollback: Drop the update_product stored procedure

DROP FUNCTION IF EXISTS update_product(
    INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);
