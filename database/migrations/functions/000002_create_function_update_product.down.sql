-- Rollback: Drop the update_product stored procedure

-- Drop current version (with is_stockeable + shop_id)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, INTEGER, VARCHAR, TEXT, DECIMAL, BOOLEAN, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

-- Drop old version (with shop_id, without is_stockeable)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);

-- Drop old version (without shop_id)
DROP FUNCTION IF EXISTS update_product(
    INTEGER, VARCHAR, TEXT, DECIMAL, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL, INTEGER,
    JSONB, JSONB
);
