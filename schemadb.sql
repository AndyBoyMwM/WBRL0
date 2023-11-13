create table orders (
                        id bigserial not NULL primary key,
                        order_uid text not NULL,
                        track_number text,
                        entry text,
                        delivery_id_fk integer not NULL,
                        payment_id_fk integer not NULL,
                        locale text,
                        internal_signature text,
                        customer_id text not NULL,
                        delivery_service text,
                        shardkey text,
                        sm_id integer,
                        date_created text DEFAULT CAST(NOW() AS text) not NULL,
                        oof_shard text
);


create table deliverys (
                           id bigserial not NULL primary key,
                           name text not NULL,
                           phone text,
                           zip text not NULL,
                           city text not NULL,
                           address text not NULL,
                           region text,
                           email text
);

create table items (
                       id bigserial not NULL primary key,
                       chrt_id bigint,
                       track_number text,
                       price integer not NULL,
                       rid text not NULL,
                       name text not NULL,
                       sale integer not NULL,
                       size text not NULL,
                       total_price integer not NULL,
                       nm_id integer not NULL,
                       brand text,
                       status integer not NULL
);

create table payments (
                          id bigserial not NULL primary key,
                          transaction text not NULL,
                          request_id text,
                          currency text not NULL,
                          provider text not NULL,
                          amount integer,
                          payment_dt bigint not NULL,
                          bank text,
                          delivery_cost integer DEFAULT 0 not NULL,
                          goods_total integer,
                          custom_fee integer
);

create table "order_items" (
                               id bigserial not NULL primary key,
                               order_id_fk bigserial,
                               item_id_fk bigserial
);


create table "cache" (
                         id bigserial not NULL primary key,
                         order_id bigserial,
                         cache_stream text
);

ALTER TABLE public.orders ADD CONSTRAINT delivery_id_fkey FOREIGN KEY (delivery_id_fk) REFERENCES public.deliverys(id) on update no action on delete no action not valid;
ALTER TABLE public.orders ADD CONSTRAINT payment_id_fkey FOREIGN KEY (payment_id_fk) REFERENCES public.payments(id) on update no action on delete no action not valid;
ALTER TABLE public.order_items ADD CONSTRAINT order_id_fkey FOREIGN KEY (order_id_fk) REFERENCES public.orders(id) match simple on update no action on delete no action not valid;
