-- Rollback: Drop the stored procedure for product creation

-- Drop current signature (with is_stockeable)
DROP FUNCTION IF EXISTS create_product(
    VARCHAR, TEXT, DECIMAL, BOOLEAN, INTEGER, INTEGER,
    BOOLEAN, BOOLEAN, BOOLEAN, DECIMAL,
    INTEGER, INTEGER, JSONB, JSONB
);

-- Drop old signature (TEXT[] for images)
DROP FUNCTION IF EXISTS create_product(
    VARCHAR(255),
    TEXT,
    DECIMAL(10,2),
    INTEGER,
    INTEGER,
    BOOLEAN,
    BOOLEAN,
    BOOLEAN,
    DECIMAL(10,2),
    INTEGER,
    INTEGER,
    TEXT[],
    JSONB
);
