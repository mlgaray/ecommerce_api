-- Stored procedure for category update with image replacement
-- Returns deleted storage_ref for Cloudinary cleanup

DROP FUNCTION IF EXISTS update_category(INTEGER, TEXT, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS update_category(INTEGER, INTEGER, TEXT, TEXT, TEXT, TEXT);

CREATE OR REPLACE FUNCTION update_category(
    p_category_id INTEGER,
    p_shop_id INTEGER,       -- Shop ID for ownership validation
    p_name TEXT,
    p_description TEXT,
    p_image_url TEXT,        -- New image URL (NULL if keeping existing)
    p_storage_ref TEXT       -- New storage ref (NULL if keeping existing)
) RETURNS TEXT AS $$
DECLARE
    v_deleted_ref TEXT := '';
    v_has_new_image BOOLEAN;
BEGIN
    -- Check if we have a new image to replace
    v_has_new_image := (p_image_url IS NOT NULL AND trim(p_image_url) <> '');

    -- 1. UPDATE category basic fields (with shop_id validation)
    UPDATE categories
    SET name = p_name,
        description = p_description
    WHERE id = p_category_id AND shop_id = p_shop_id;

    -- Check if category was found and belongs to shop
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Category not found or does not belong to shop: category_id=%, shop_id=%', p_category_id, p_shop_id
        USING ERRCODE = 'P0003';  -- Category not found
    END IF;

    -- 2. Handle image replacement (only if new image is provided)
    IF v_has_new_image THEN
        -- Delete existing image and capture storage_ref in one operation
        DELETE FROM images
        WHERE category_id = p_category_id
        RETURNING storage_ref INTO v_deleted_ref;

        -- Insert new image
        INSERT INTO images (url, storage_ref, category_id)
        VALUES (p_image_url, COALESCE(p_storage_ref, ''), p_category_id);
    END IF;

    -- Return deleted storage_ref for Cloudinary cleanup (empty string if no deletion)
    RETURN COALESCE(v_deleted_ref, '');

EXCEPTION
    WHEN unique_violation THEN
        RAISE;  -- Re-raise with original code 23505 for Go to handle
    -- Other errors propagate naturally with original SQLSTATE
END;
$$ LANGUAGE plpgsql
SET search_path = public;

-- Function documentation
COMMENT ON FUNCTION update_category IS
'Updates a category with optional image replacement.
Parameters:
- p_category_id: ID of the category to update
- p_shop_id: Shop ID for ownership validation (category must belong to this shop)
- p_name: New category name
- p_description: New category description
- p_image_url: URL of new image (NULL to keep existing)
- p_storage_ref: Storage reference of new image (NULL to keep existing)
Returns:
- TEXT: storage_ref of deleted image for Cloudinary cleanup (empty if no image replaced)
Behavior:
- Validates that category belongs to the specified shop
- If p_image_url is NULL or empty, existing image is preserved
- If p_image_url is provided, existing image is deleted and new one inserted
- Uses DELETE ... RETURNING for single-query delete + capture
- Returns old storage_ref for async Cloudinary cleanup
Exception handling included for validation errors.';
