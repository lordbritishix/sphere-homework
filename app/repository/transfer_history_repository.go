package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"sphere-homework/app/event"
	"time"
)

type TransferHistoryRepository struct {
	db  *pgxpool.Pool
	ctx context.Context
}

func NewTransferHistoryRepository(db *pgxpool.Pool, ctx context.Context) TransferHistoryRepository {
	return TransferHistoryRepository{
		db:  db,
		ctx: ctx,
	}
}

func (t *TransferHistoryRepository) InsertTransferHistory(event event.BaseEvent) error {
	sql := `
		INSERT INTO transfer_history (created_at, event_type, sender, event)
		VALUES ($1, $2, $3, $4)
	`

	timestamp := time.UnixMilli(event.Timestamp)
	_, err := t.db.Exec(t.ctx, sql, timestamp, event.EventType, event.Sender, string(event.Payload))

	return err
}
