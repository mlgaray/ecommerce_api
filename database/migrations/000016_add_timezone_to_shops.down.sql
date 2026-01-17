DROP INDEX IF EXISTS idx_shops_timezone_id;

ALTER TABLE public.shops DROP CONSTRAINT IF EXISTS shops_timezone_id_fkey;

ALTER TABLE public.shops DROP COLUMN IF EXISTS timezone_id;
