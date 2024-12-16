BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE outgoing_transfer_status AS ENUM ('UNSENT', 'SENT', 'COMPLETED', 'FAILED', 'CANCELLED');
CREATE TYPE outgoing_transfer_type AS ENUM ('INTERNAL', 'EXTERNAL');
CREATE TYPE ledger_entry_type AS ENUM ('FEE', 'TRANSFER');

CREATE TABLE IF NOT EXISTS outgoing_transfer (
    transfer_id UUID NOT NULL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    sent_at TIMESTAMP WITH TIME ZONE,
    from_asset VARCHAR NOT NULL,
    to_asset VARCHAR NOT NULL,
    requested_amount NUMERIC(40, 30) NOT NULL,
    fee NUMERIC(40, 30) NOT NULL,
    net_amount NUMERIC(40, 30) NOT NULL,
    rate NUMERIC(40, 30) NOT NULL,
    sent_amount NUMERIC(40, 30),
    sender VARCHAR NOT NULL,
    recipient VARCHAR NOT NULL,
    status outgoing_transfer_status NOT NULL DEFAULT 'UNSENT',
    failure_reason VARCHAR,
    transfer_type outgoing_transfer_type NOT NULL DEFAULT 'EXTERNAL',
    lock_id UUID
);

CREATE TABLE IF NOT EXISTS transfer_history (
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    event_type VARCHAR NOT NULL,
    sender VARCHAR NOT NULL,
    event JSONB NOT NULL
);

-- store current rate for fast lookup
CREATE TABLE IF NOT EXISTS rate (
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    from_asset VARCHAR NOT NULL,
    to_asset VARCHAR NOT NULL,
    rate NUMERIC(40, 30) NOT NULL,

    PRIMARY KEY (from_asset, to_asset)
);

CREATE TABLE IF NOT EXISTS historical_rate (
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    from_asset VARCHAR NOT NULL,
    to_asset VARCHAR NOT NULL,
    rate NUMERIC(40, 30) NOT NULL
);

CREATE TABLE IF NOT EXISTS ledger (
    account_name VARCHAR NOT NULL,
    balance NUMERIC(40, 30) NOT NULL DEFAULT 0,
    asset VARCHAR NOT NULL,

    PRIMARY KEY (account_name, asset)
);

CREATE TABLE IF NOT EXISTS ledger_history (
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    account VARCHAR NOT NULL,
    asset VARCHAR NOT NULL,
    amount NUMERIC(40, 30),
    transfer_id UUID NOT NULL,
    ledger_entry_type ledger_entry_type NOT NULL
);

CREATE TABLE IF NOT EXISTS fee (
    to_asset VARCHAR NOT NULL,
    fee NUMERIC(40, 30) NOT NULL
);

CREATE INDEX ledger__account_name ON ledger(account_name);

COMMIT;