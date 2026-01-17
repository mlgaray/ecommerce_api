INSERT INTO public.timezones (name, identifier, utc_offset, is_active)
VALUES ('Buenos Aires', 'America/Buenos_Aires', '-03:00', true)
ON CONFLICT (identifier) DO NOTHING;
