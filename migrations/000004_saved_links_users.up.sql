create table users (
    id serial primary key,
    created_at timestamp not null default now()
);

alter table saved_links add column user_id integer;