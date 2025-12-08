-- Rollback: Drop the update_product stored procedure

-- Drop the current version (14 parameters with p_shop_id)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

-- Also drop the old version (13 parameters without p_shop_id) for clean rollback
DROP FUNCTION IF EXISTS update_product(
    INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);
