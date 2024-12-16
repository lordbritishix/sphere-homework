package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"sphere-homework/app/model"
	"time"
)

type TransferRepository struct {
	db  *pgxpool.Pool
	ctx context.Context
}

func NewTransferRepository(db *pgxpool.Pool, ctx context.Context) TransferRepository {
	return TransferRepository{
		db:  db,
		ctx: ctx,
	}
}

func (t *TransferRepository) InsertOutgoingTransfer(transfer model.Transfer) error {
	sql := `
		INSERT INTO outgoing_transfer (transfer_id, created_at, from_asset, to_asset, requested_amount, fee, net_amount, sender, recipient, status, transfer_type, rate) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := t.db.Exec(t.ctx, sql, uuid.New(), time.Now().UTC(), transfer.FromAsset, transfer.ToAsset, transfer.RequestedAmount, transfer.Fee, transfer.RequestedAmount-transfer.Fee, transfer.Sender, transfer.Recipient, model.UnsentTransferStatus, transfer.TransferType, transfer.Rate)
	if err != nil {
		return err
	}

	return nil
}

func (t *TransferRepository) LockTransfer(transferId uuid.UUID) (*model.Transfer, error) {
	sql := `
		UPDATE outgoing_transfer 
		SET lock_id = uuid_generate_v4()
		WHERE transfer_id = $1
		AND lock_id IS NULL
		RETURNING transfer_id, created_at, from_asset, to_asset, requested_amount, fee, net_amount, sender, recipient, status, transfer_type, rate, lock_id`

	var transfer model.Transfer
	err := t.db.QueryRow(t.ctx, sql, transferId).Scan(
		&transfer.TransferId,
		&transfer.CreatedAt,
		&transfer.FromAsset,
		&transfer.ToAsset,
		&transfer.RequestedAmount,
		&transfer.Fee,
		&transfer.NetAmount,
		&transfer.Sender,
		&transfer.Recipient,
		&transfer.TransferStatus,
		&transfer.TransferType,
		&transfer.Rate,
		&transfer.LockId,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transfer already locked or not found")
	}

	if err != nil {
		return nil, err
	}

	return &transfer, nil
}

func (t *TransferRepository) UnlockAndUpdateTransfer(transfer model.Transfer) (*model.Transfer, error) {
	sql := `
		UPDATE outgoing_transfer 
		SET lock_id = NULL, sent_at = $2, status = $3, sent_amount = $4
		WHERE transfer_id = $1
		AND lock_id IS NOT NULL
		RETURNING transfer_id, created_at, sent_at, from_asset, to_asset, requested_amount, fee, net_amount, rate, sent_amount, sender, recipient, status, failure_reason, transfer_type, lock_id
		`

	var updatedTransfer model.Transfer
	err := t.db.QueryRow(t.ctx, sql, transfer.TransferId, transfer.SentAt, transfer.TransferStatus, transfer.SentAmount).Scan(
		&updatedTransfer.TransferId,
		&updatedTransfer.CreatedAt,
		&updatedTransfer.SentAt,
		&updatedTransfer.FromAsset,
		&updatedTransfer.ToAsset,
		&updatedTransfer.RequestedAmount,
		&updatedTransfer.Fee,
		&updatedTransfer.NetAmount,
		&updatedTransfer.Rate,
		&updatedTransfer.SentAmount,
		&updatedTransfer.Sender,
		&updatedTransfer.Recipient,
		&updatedTransfer.TransferStatus,
		&updatedTransfer.FailureReason,
		&updatedTransfer.TransferType,
		&updatedTransfer.LockId,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("transfer already locked or not found")
	}

	if err != nil {
		return nil, err
	}

	return &updatedTransfer, nil
}

func (t *TransferRepository) GetTransfer(transferId uuid.UUID) (*model.Transfer, error) {
	sql := `
		SELECT * FROM outgoing_transfer
		WHERE transfer_id = $1`

	row := t.db.QueryRow(t.ctx, sql, model.UnsentTransferStatus, transferId)

	var transfer model.Transfer

	err := row.Scan(&transfer.TransferId, &transfer.CreatedAt, &transfer.SentAt,
		&transfer.FromAsset, &transfer.ToAsset,
		&transfer.RequestedAmount, &transfer.Fee, &transfer.NetAmount,
		&transfer.Rate, &transfer.SentAmount, &transfer.Sender, &transfer.Recipient,
		&transfer.TransferStatus, &transfer.FailureReason, &transfer.TransferType,
		&transfer.LockId)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &transfer, nil
}

func (t *TransferRepository) GetUnsentTransfers(limit int) ([]model.Transfer, error) {
	sql := `
		SELECT * FROM outgoing_transfer
		WHERE status = $1
		AND lock_id IS NULL
		ORDER BY created_at LIMIT $2`

	rows, err := t.db.Query(t.ctx, sql, model.UnsentTransferStatus, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var transfers []model.Transfer
	for rows.Next() {
		var transfer model.Transfer
		err := rows.Scan(
			&transfer.TransferId, &transfer.CreatedAt, &transfer.SentAt,
			&transfer.FromAsset, &transfer.ToAsset,
			&transfer.RequestedAmount, &transfer.Fee, &transfer.NetAmount,
			&transfer.Rate, &transfer.SentAmount, &transfer.Sender, &transfer.Recipient,
			&transfer.TransferStatus, &transfer.FailureReason, &transfer.TransferType,
			&transfer.LockId)

		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		transfers = append(transfers, transfer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return transfers, nil
}
