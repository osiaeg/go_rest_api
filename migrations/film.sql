CREATE TABLE public.film
(
    film_id integer NOT NULL,
    film_name character varying(150) NOT NULL,
    film_description character varying(1000) NOT NULL,
    film_release_date date NOT NULL,
    film_rating integer NOT NULL,
    film_actor_list integer[] NOT NULL,
    PRIMARY KEY (film_id)
);

ALTER TABLE IF EXISTS public.film
    OWNER to "user";
