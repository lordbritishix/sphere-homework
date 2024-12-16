BEGIN;

DROP TABLE IF EXISTS outgoing_transfer;
DROP TABLE IF EXISTS ledger;
DROP TYPE IF EXISTS outgoing_transfer_status;
DROP TYPE IF EXISTS outgoing_transfer_type;

COMMIT;