-- Stored procedure for category creation with image
-- Creates category and image atomically in a single transaction

DROP FUNCTION IF EXISTS create_category(TEXT, TEXT, INTEGER, TEXT, TEXT);

CREATE OR REPLACE FUNCTION create_category(
    p_name TEXT,
    p_description TEXT,
    p_shop_id INTEGER,
    p_image_url TEXT,
    p_storage_ref TEXT
) RETURNS INTEGER AS $$
DECLARE
    v_category_id INTEGER;
BEGIN
    -- 1. Create category
    INSERT INTO categories (name, description, shop_id)
    VALUES (p_name, p_description, p_shop_id)
    RETURNING id INTO v_category_id;

    -- 2. Create image if provided
    IF p_image_url IS NOT NULL AND p_image_url != '' THEN
        INSERT INTO images (url, storage_ref, category_id)
        VALUES (p_image_url, COALESCE(p_storage_ref, ''), v_category_id);
    END IF;

    RETURN v_category_id;

EXCEPTION
    WHEN unique_violation THEN
        RAISE;  -- Re-raise with original code 23505 for Go to handle
    -- Other errors propagate naturally with original SQLSTATE
END;
$$ LANGUAGE plpgsql;

-- Function documentation
COMMENT ON FUNCTION create_category IS
'Creates a category with its image in a single atomic transaction.
Parameters:
- p_name: Category name
- p_description: Category description
- p_shop_id: Shop ID that owns this category
- p_image_url: URL of the uploaded image (from Cloudinary)
- p_storage_ref: Storage reference for deletion (Cloudinary public_id)
Returns the created category ID.
Image is optional - if p_image_url is NULL or empty, no image record is created.';
