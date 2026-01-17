ALTER TABLE public.shops ADD COLUMN timezone_id BIGINT NULL;

ALTER TABLE public.shops ADD CONSTRAINT shops_timezone_id_fkey
    FOREIGN KEY (timezone_id) REFERENCES timezones (id)
    ON UPDATE CASCADE ON DELETE RESTRICT;

CREATE INDEX idx_shops_timezone_id ON public.shops (timezone_id);
