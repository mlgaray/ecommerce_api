-- Rollback: Drop the stored procedure for product creation

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
