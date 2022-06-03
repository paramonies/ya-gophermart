-- +migrate Up
create table if not exists orders
(
    id              uuid default gen_random_uuid(),
    user_id         uuid,
    order_number    text not null,
    price           double precision default 0,
    updated_at      timestamp default now(),

    constraint orders_pk primary key (id),
    foreign key (user_id) references users (id),
    constraint order_number unique (order_number)
    );
-- +migrate Down
drop table orders;
