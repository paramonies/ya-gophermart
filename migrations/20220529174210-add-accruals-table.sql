-- +migrate Up
drop table if exists accruals;
drop type if exists order_status;
create type order_status as ENUM ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED');
create table if not exists accruals
(
    id              uuid default gen_random_uuid(),
    order_number    text not null,
    accrual         double precision default 0,
    user_id         uuid,
    order_status    order_status not null default 'NEW',
    updated_at      timestamp default now(),

    constraint accruals_pk primary key (id),
    foreign key (user_id) references users (id),
    constraint number unique (order_number)
    );
-- +migrate Down
drop table accruals;
drop type order_status;