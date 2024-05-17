CREATE TABLE public.actor
(
    actor_id integer NOT NULL GENERATED ALWAYS AS IDENTITY,
    actor_name character varying NOT NULL,
    actor_sex "char" NOT NULL,
    actor_birthday date NOT NULL,
    PRIMARY KEY (actor_id)
);

ALTER TABLE IF EXISTS public.actor
    OWNER to "user";
