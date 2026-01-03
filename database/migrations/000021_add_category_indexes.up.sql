-- Migration: Add indexes for category search and filtering
-- Purpose: Optimize category listing with search and pagination

-- Enable pg_trgm extension for trigram-based text search (if not already enabled)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Index for text search on category name using trigrams
-- Enables fast ILIKE queries: WHERE name ILIKE '%search_term%'
CREATE INDEX IF NOT EXISTS idx_categories_name_trgm
ON categories USING gin (name gin_trgm_ops);

-- Composite index for sorting by created_at with shop_id filter
-- Supports: ORDER BY created_at DESC/ASC WHERE shop_id = ?
CREATE INDEX IF NOT EXISTS idx_categories_shop_created_at
ON categories (shop_id, created_at DESC);
