package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeeRepository struct {
	db  *pgxpool.Pool
	ctx context.Context
}

func NewFeeRepository(db *pgxpool.Pool, ctx context.Context) FeeRepository {
	return FeeRepository{
		db:  db,
		ctx: ctx,
	}
}

func (f *FeeRepository) GetFee(toAsset string) (float64, error) {
	query := `
		SELECT fee 
		FROM fee
		WHERE to_asset = $1
	`

	var fee float64

	err := f.db.QueryRow(f.ctx, query, toAsset).Scan(&fee)
	if err != nil {
		return 0, err
	}

	return fee, nil
}
