UPDATE public.shops
SET timezone_id = (SELECT id FROM public.timezones WHERE identifier = 'America/Buenos_Aires' LIMIT 1)
WHERE timezone_id IS NULL;
