-- +migrate Up
create table if not exists withdrawals
(
    id              uuid default gen_random_uuid(),
    user_id         uuid,
    orderNumber     text not null,
    sum             double precision default 0,
    processed_at    timestamp default now(),

    constraint withdrawals_pk primary key (id),
    foreign key (user_id) references users (id),
    constraint orderNumber unique (orderNumber)
    );
-- +migrate Down
drop table withdrawals;
