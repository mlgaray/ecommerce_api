-- Drop update_category stored procedure

-- Drop the current version (6 parameters with p_shop_id)
DROP FUNCTION IF EXISTS update_category(INTEGER, INTEGER, TEXT, TEXT, TEXT, TEXT);

-- Also drop the old version (5 parameters without p_shop_id) for clean rollback
DROP FUNCTION IF EXISTS update_category(INTEGER, TEXT, TEXT, TEXT, TEXT);
