package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"time"
)

type RateRepository struct {
	db     *pgxpool.Pool
	ctx    context.Context
	logger *zap.Logger
}

func NewRateRepository(db *pgxpool.Pool, ctx context.Context, logger *zap.Logger) RateRepository {
	return RateRepository{
		db:     db,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *RateRepository) UpsertRate(fromAsset string, toAsset string, rate float64, timestamp time.Time) error {
	tx, err := r.db.Begin(r.ctx)
	defer func() {
		var err error
		if p := recover(); p != nil {
			err = tx.Rollback(r.ctx)
		} else if err != nil {
			err = tx.Rollback(r.ctx)
		} else {
			err = tx.Commit(r.ctx)
		}

		if err != nil {
			r.logger.Error("failed to commit / rollback transaction", zap.Error(err))
		}
	}()

	if err != nil {
		return err
	}

	sql := `
		INSERT INTO rate (updated_at, from_asset, to_asset, rate) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (from_asset, to_asset) DO UPDATE SET 
			rate = EXCLUDED.rate,
			updated_at = NOW();
		`

	_, err = tx.Exec(r.ctx, sql, timestamp.UTC(), fromAsset, toAsset, rate)
	if err != nil {
		return err
	}

	sql = `
		INSERT INTO historical_rate (created_at, from_asset, to_asset, rate)
		VALUES ($1, $2, $3, $4)
	`

	_, err = tx.Exec(r.ctx, sql, timestamp.UTC(), fromAsset, toAsset, rate)
	if err != nil {
		return err
	}

	return nil
}

func (r *RateRepository) GetRate(fromAsset string, toAsset string) (float64, error) {
	if fromAsset == toAsset {
		return 1.0, nil
	}

	sql := `
		SELECT rate 
		FROM rate
		WHERE from_asset = $1 AND to_asset = $2
	`

	var rate float64
	err := r.db.QueryRow(r.ctx, sql, fromAsset, toAsset).Scan(&rate)

	if err != nil {
		return 0, err
	}

	return rate, nil
}
