-- +migrate Up

CREATE TABLE users (
    id uuid primary key,
    name varchar not null,
    created_at timestamp,
    deleted bool
);

CREATE TABLE wallets (
    id uuid primary key,
    owner uuid references users (id),
    currency varchar,
    balance varchar,
    created_at timestamp,
    deleted bool
);

-- +migrate Down

DROP TABLE wallets, users CASCADE;