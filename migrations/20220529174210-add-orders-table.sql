-- +migrate Up
drop table if exists orders;
drop type if exists order_status;
create type order_status as ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
create table if not exists orders
(
    id              uuid default gen_random_uuid(),
    number          text not null,
    accrual         double precision default 0,
    user_id         uuid,
    status          order_status not null default 'NEW',
    updated_at      timestamp default now(),

    constraint orders_pk primary key (id),
    foreign key (user_id) references users (id),
    constraint number unique (number)
    );
-- +migrate Down
drop table orders;
drop type order_status;