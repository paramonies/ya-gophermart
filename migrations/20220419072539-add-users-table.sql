-- +migrate Up
create table if not exists users
(
    id              uuid default gen_random_uuid(),
    user_name       text not null,
    password_hash   text not null,
    token           text,
    created_at      timestamp default now(),

    constraint users_pk primary key (id),
    constraint user_name unique (user_name)
);
-- +migrate Down
drop table users;

