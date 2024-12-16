BEGIN;

-- Print monies representing initial balance to the omni-bus wallet
INSERT INTO ledger (account_name, balance, asset)
VALUES
    ('system', '1000000', 'USD'),
    ('system', '921658', 'EUR'),
    ('system', '109890110', 'JPY'),
    ('system', '750000', 'GBP'),
    ('system', '1349528', 'AUD');

-- Create test users with free monies
INSERT INTO ledger (account_name, balance, asset)
VALUES
    -- jim user
    ('jim', '3500', 'USD'),
    ('jim', '0', 'EUR'),
    ('jim', '0', 'JPY'),
    ('jim', '210', 'GBP'),
    ('jim', '0', 'AUD'),

    -- jacob user
    ('jacob', '2000', 'USD'),
    ('jacob', '0', 'EUR'),
    ('jacob', '0', 'JPY'),
    ('jacob', '100', 'GBP'),
    ('jacob', '0', 'AUD')
;

-- Create test fees
INSERT INTO fee (to_asset, fee)
VALUES
    ('USD', 0.01),
    ('EUR', 0.02),
    ('JPY', 0.03),
    ('GBP', 0.0125),
    ('AUD', 0.013);

COMMIT;