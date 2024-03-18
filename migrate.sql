CREATE TABLE IF NOT EXISTS public.actor
(
    actor_id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    actor_name character varying COLLATE pg_catalog."default" NOT NULL,
    actor_sex character varying(1) COLLATE pg_catalog."default" NOT NULL,
    actor_birthday date NOT NULL,
    CONSTRAINT actor_pkey PRIMARY KEY (actor_id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.actor
    OWNER to "user";

CREATE TABLE IF NOT EXISTS public.film
(
    film_id integer NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 2147483647 CACHE 1 ),
    film_name character varying(150) COLLATE pg_catalog."default" NOT NULL,
    film_description character varying(1000) COLLATE pg_catalog."default" NOT NULL,
    film_release_date date NOT NULL,
    film_rating integer NOT NULL,
    film_actor_list integer[] NOT NULL,
    CONSTRAINT film_pkey PRIMARY KEY (film_id)
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.film
    OWNER to "user";


CREATE TABLE IF NOT EXISTS public.actor_film
(
    actor_id integer NOT NULL,
    film_id integer NOT NULL,
    CONSTRAINT unique_rows UNIQUE (actor_id, film_id)
        INCLUDE(actor_id, film_id),
    CONSTRAINT actor_fkey FOREIGN KEY (actor_id)
        REFERENCES public.actor (actor_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID,
    CONSTRAINT film_fkey FOREIGN KEY (film_id)
        REFERENCES public.film (film_id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE CASCADE
        NOT VALID
)

TABLESPACE pg_default;

ALTER TABLE IF EXISTS public.actor_film
    OWNER to "user";
-- Index: fki_actor_fkey

-- DROP INDEX IF EXISTS public.fki_actor_fkey;

CREATE INDEX IF NOT EXISTS fki_actor_fkey
    ON public.actor_film USING btree
    (actor_id ASC NULLS LAST)
    TABLESPACE pg_default;
-- Index: fki_film_fkey

-- DROP INDEX IF EXISTS public.fki_film_fkey;

CREATE INDEX IF NOT EXISTS fki_film_fkey
    ON public.actor_film USING btree
    (film_id ASC NULLS LAST)
    TABLESPACE pg_default;

