-- +migrate Up

CREATE TABLE users (
    id uuid not null primary key,
    name varchar not null,
    created_at timestamp,
    deleted bool not null
);

CREATE TABLE wallets (
    id uuid not null primary key,
    owner uuid references users (id),
    currency varchar not null,
    balance numeric not null DEFAULT 0 check ( balance >= 0 ),
    created_at timestamp,
    updated_at timestamp,
    deleted bool not null
);

CREATE TABLE transactions_history (
    id uuid primary key,
    wallet_id uuid not null,
    target_wallet_id uuid,
    amount numeric,
    converted_amount numeric,
    currency varchar not null,
    transaction_type varchar,
    executed_at timestamp
);
-- +migrate Down

DROP TABLE users, wallets, transactions_history CASCADE;
