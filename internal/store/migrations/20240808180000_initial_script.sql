-- +migrate Up

CREATE TABLE users (
    id uuid not null primary key,
    name varchar not null,
    email varchar not null,
    phone varchar not null,
    password varchar not null,
    created_at timestamp not null,
    updated_at timestamp,
    deleted bool not null
);

CREATE TABLE wallets (
    id uuid not null primary key,
    owner uuid references users (id),
    currency varchar not null,
    balance numeric not null DEFAULT 0 check ( balance >= 0 ),
    created_at timestamp not null,
    updated_at timestamp,
    deleted bool not null
);

CREATE TABLE transactions_history (
    id uuid primary key,
    wallet_id uuid not null,
    target_wallet_id uuid,
    amount numeric not null,
    converted_amount numeric,
    currency varchar not null,
    transaction_type varchar not null,
    executed_at timestamp not null
);
-- +migrate Down

DROP TABLE users, wallets, transactions_history CASCADE;
