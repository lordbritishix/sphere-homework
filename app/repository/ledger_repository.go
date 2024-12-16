package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"sphere-homework/app/model"
)

const SystemAccount = "system"

type LedgerRepository struct {
	db     *pgxpool.Pool
	ctx    context.Context
	logger *zap.Logger
}

func NewLedgerRepository(db *pgxpool.Pool, ctx context.Context, logger *zap.Logger) LedgerRepository {
	return LedgerRepository{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (l *LedgerRepository) InsertNewEntryIfNotExists(asset string, accountName string) error {
	query := `
		INSERT INTO ledger(asset, account_name) 
		VALUES ($1, $2)
		ON CONFLICT (asset, account_name) DO NOTHING 
		`
	_, err := l.db.Exec(l.ctx, query, asset, accountName)

	if err != nil {
		return err
	}

	return nil
}

func (l *LedgerRepository) GetBalances(account string) ([]model.LedgerBalance, error) {
	// Get the balance
	query := `
		SELECT asset, balance 
		FROM ledger
		WHERE account_name = $1
	`

	balanceRows, err := l.db.Query(l.ctx, query, account)
	if err != nil {
		return nil, err
	}

	defer balanceRows.Close()
	balances := make(map[string]float64)
	for balanceRows.Next() {
		var asset string
		var balance float64
		if err := balanceRows.Scan(&asset, &balance); err != nil {
			return nil, err
		}
		balances[asset] = balance
	}

	// Sum all the inflows and the outflows for a given time period (1 day) (note to self - query may need some indexing on created_at to be able to perform aggregates faster on volume)
	query = `
		SELECT asset, 
			COALESCE(SUM (CASE WHEN amount > 0 THEN amount ELSE 0 END), 0) AS INFLOW,
			COALESCE(SUM (CASE WHEN amount < 0 THEN -amount ELSE 0 END), 0) AS OUTFLOW
		FROM ledger_history
		WHERE account = $1
		AND created_at >= NOW() - INTERVAL '1 DAY'
		GROUP BY asset
	`

	flowRows, err := l.db.Query(l.ctx, query, account)
	if err != nil {
		return nil, err
	}

	defer flowRows.Close()
	flows := make(map[string]model.LedgerBalance)
	for flowRows.Next() {
		var asset string
		var inflow float64
		var outflow float64
		if err := flowRows.Scan(&asset, &inflow, &outflow); err != nil {
			return nil, err
		}

		balance, ok := balances[asset]
		if ok {
			flows[asset] = model.LedgerBalance{
				Asset:   asset,
				Inflow:  inflow,
				Outflow: outflow,
				Amount:  balance,
			}
		}
	}

	var result []model.LedgerBalance

	for key, balance := range balances {
		flow, ok := flows[key]
		if ok {
			result = append(result, flow)
		} else {
			result = append(result, model.LedgerBalance{
				Asset:   key,
				Amount:  balance,
				Inflow:  0,
				Outflow: 0,
			})
		}
	}

	return result, nil
}

func (l *LedgerRepository) Transfer(transfer *model.Transfer) error {
	tx, err := l.db.Begin(l.ctx)
	if err != nil {
		return err
	}

	defer func() {
		var err error
		if p := recover(); p != nil {
			err = tx.Rollback(l.ctx)
		} else if err != nil {
			err = tx.Rollback(l.ctx)
		} else {
			err = tx.Commit(l.ctx)
		}

		if err != nil {
			l.logger.Error("failed to commit / rollback transaction", zap.Error(err))
		}
	}()

	var sourceBalance float64
	var destBalance float64

	query := `SELECT balance FROM ledger WHERE account_name = $1 AND asset = $2 FOR UPDATE`

	// Lock both sender and receiver entries of the ledger via SELECT FOR UPDATE
	err = tx.QueryRow(l.ctx, query, transfer.Sender, transfer.FromAsset).Scan(&sourceBalance)
	if err != nil {
		return err
	}

	err = tx.QueryRow(l.ctx, query, transfer.Recipient, transfer.ToAsset).Scan(&destBalance)
	if err != nil {
		return err
	}

	// Also lock system account because we need to transfer fees to system account
	if transfer.Sender != SystemAccount {
		var ledgerBalance float64
		err = tx.QueryRow(l.ctx, query, SystemAccount, transfer.FromAsset).Scan(&ledgerBalance)
		if err != nil {
			return err
		}
	}

	var entries []model.LedgerEntry

	// deduct amount is simply the requested amount
	var deductAmount = transfer.RequestedAmount

	// send amount is requested amount less fees, converted to the target asset
	var sendAmount = (transfer.RequestedAmount - transfer.Fee) * transfer.Rate

	if sourceBalance < deductAmount {
		return fmt.Errorf("not enough balance for transfer")
	}

	// Debit deduct amount from sender
	entries = append(entries, model.LedgerEntry{
		TransferId: transfer.TransferId,
		Account:    transfer.Sender,
		Asset:      transfer.FromAsset,
		Amount:     -deductAmount,
		Type:       model.TransferLedgerEntryType,
	})

	if transfer.Fee > 0 {
		// no fees for internal transfers
		if transfer.Sender != SystemAccount || transfer.Recipient != SystemAccount {
			// Credit fees to system account
			// Notes:
			// If sender is user, we deduct the full requested amount from the user, then we add fees to system account and send amount less fees to the recipient
			// If sender is system, we deduct the full requested amount from the system, but credit back the fees to the system account and send amount less fees to the recipient
			entries = append(entries, model.LedgerEntry{
				TransferId: transfer.TransferId,
				Account:    SystemAccount,
				Asset:      transfer.FromAsset,
				Amount:     transfer.Fee,
				Type:       model.FeeLedgerEntryType,
			})
		}
	}

	// Credit send amount to the recipient
	entries = append(entries, model.LedgerEntry{
		TransferId: transfer.TransferId,
		Account:    transfer.Recipient,
		Asset:      transfer.ToAsset,
		Amount:     sendAmount,
		Type:       model.TransferLedgerEntryType,
	})

	// apply the ledger operations
	if err = l.applyLedgerEntries(entries, &tx); err != nil {
		l.logger.Error("failed to apply ledger entries", zap.Error(err))
		return err
	}

	transfer.SentAmount = &sendAmount

	return nil
}

func (l *LedgerRepository) applyLedgerEntries(entries []model.LedgerEntry, tx *pgx.Tx) error {
	query := `UPDATE ledger SET balance = balance + $1 WHERE account_name = $2 AND asset = $3`
	queryHistory := `
		INSERT INTO ledger_history (transfer_id, account, asset, amount, ledger_entry_type) 
		VALUES ($1, $2, $3, $4, $5)`

	for _, entry := range entries {
		// debits or credits entry against the ledger
		if _, err := (*tx).Exec(l.ctx, query, entry.Amount, entry.Account, entry.Asset); err != nil {
			return fmt.Errorf("failed to apply ledger entry: %w", err)
		}

		// record ledger history
		if _, err := (*tx).Exec(l.ctx, queryHistory, entry.TransferId, entry.Account, entry.Asset, entry.Amount, entry.Type); err != nil {
			return fmt.Errorf("failed to apply ledger history entry: %w", err)
		}
	}

	return nil
}
