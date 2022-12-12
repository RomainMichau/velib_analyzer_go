-- public.run definition

-- Drop table

-- DROP TABLE public.run;

CREATE TABLE public.run (
                            id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
                            run_time timestamptz NOT NULL,
                            minor_issues_count int8 NULL DEFAULT 0,
                            success bool NULL DEFAULT false,
                            CONSTRAINT run_pk PRIMARY KEY (id),
                            CONSTRAINT run_time_un UNIQUE (run_time)
);


-- public.stations definition

-- Drop table

-- DROP TABLE public.stations;

CREATE TABLE public.stations (
                                 id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
                                 station_name varchar NOT NULL,
                                 long numeric NOT NULL,
                                 lat numeric NOT NULL,
                                 station_code int8 NOT NULL,
                                 run int8 NOT NULL,
                                 CONSTRAINT code_un UNIQUE (station_code),
                                 CONSTRAINT name_un UNIQUE (station_name),
                                 CONSTRAINT station_pk PRIMARY KEY (id),
                                 CONSTRAINT run_fk FOREIGN KEY (run) REFERENCES public.run(id)
);


-- public.velibs definition

-- Drop table

-- DROP TABLE public.velibs;

CREATE TABLE public.velibs (
                               id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
                               velib_code int8 NOT NULL,
                               electric bool NOT NULL,
                               run int8 NOT NULL,
                               CONSTRAINT velib_code_un UNIQUE (velib_code),
                               CONSTRAINT velib_pk PRIMARY KEY (id),
                               CONSTRAINT run_fk FOREIGN KEY (run) REFERENCES public.run(id)
);


-- public.rating definition

-- Drop table

-- DROP TABLE public.rating;

CREATE TABLE public.rating (
                               id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
                               velib_code int8 NOT NULL,
                               rate int8 NOT NULL,
                               rate_time timestamptz NOT NULL,
                               run int8 NOT NULL,
                               CONSTRAINT rating_pk PRIMARY KEY (id),
                               CONSTRAINT velib_code UNIQUE (velib_code, rate_time),
                               CONSTRAINT run_fk FOREIGN KEY (run) REFERENCES public.run(id),
                               CONSTRAINT velib_code_fk FOREIGN KEY (velib_code) REFERENCES public.velibs(velib_code)
);


-- public.velib_docked definition

-- Drop table

-- DROP TABLE public.velib_docked;

CREATE TABLE public.velib_docked (
                                     id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
                                     velib_code int8 NOT NULL,
                                     "timestamp" timestamptz NOT NULL,
                                     station_code int8 NOT NULL,
                                     run int8 NOT NULL,
                                     available bool NOT NULL,
                                     CONSTRAINT velib_docked_pk PRIMARY KEY (id),
                                     CONSTRAINT run_fk FOREIGN KEY (run) REFERENCES public.run(id),
                                     CONSTRAINT velib_code_fk FOREIGN KEY (velib_code) REFERENCES public.velibs(velib_code),
                                     CONSTRAINT velib_docked_fk FOREIGN KEY (station_code) REFERENCES public.stations(station_code)
);
CREATE UNIQUE INDEX velib_docked_velib_code_idx ON public.velib_docked USING btree (velib_code, "timestamp");


-- public.avg_velib_per_station_dow_hr source

CREATE MATERIALIZED VIEW public.avg_velib_per_station_dow_hr
    TABLESPACE pg_default
AS SELECT avg(stuff.nb_docked) AS avg,
          stuff.station_code,
          stuff.dow,
          stuff.hour
   FROM ( SELECT date_trunc('hour'::text, vd."timestamp") AS td,
                 count(vd.velib_code) AS nb_docked,
                 vd.station_code,
                 date_part('isodow'::text, vd."timestamp") AS dow,
                 date_part('hour'::text, vd."timestamp") AS hour
          FROM velib_docked vd
          GROUP BY vd.station_code, (date_trunc('hour'::text, vd."timestamp")), (date_part('isodow'::text, vd."timestamp")), (date_part('hour'::text, vd."timestamp"))
          ORDER BY vd.station_code, (date_trunc('hour'::text, vd."timestamp"))) stuff
   GROUP BY stuff.station_code, stuff.dow, stuff.hour
WITH DATA;

CREATE UNIQUE INDEX avg_velib_per_station_dow_hr_station_code_idx ON public.avg_velib_per_station_dow_hr (station_code,dow,"hour");
