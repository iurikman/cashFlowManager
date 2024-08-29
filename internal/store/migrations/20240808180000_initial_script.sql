-- +migrate Up

CREATE TABLE users (
    id uuid primary key,
    name varchar not null,
    created_at timestamp,
    deleted bool
);

CREATE TABLE wallets (
    id uuid primary key,
    owner uuid,
    currency varchar,
    balance float,
    created_at timestamp,
    updated_at timestamp,
    deleted bool
);

CREATE TABLE transactions_history (
    id uuid primary key,
    wallet_id uuid,
    target_wallet_id uuid,
    amount float,
    currency varchar,
    transaction_type varchar,
    executed_at timestamp
);
-- +migrate Down

DROP TABLE wallets, users, transactions_history CASCADE;
