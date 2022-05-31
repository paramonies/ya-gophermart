package pgx

import (
	"context"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type WithdrawnRepo struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewUWithdrawnRepo(p *pgxpool.Pool, queryTimeout time.Duration) *WithdrawnRepo {
	return &WithdrawnRepo{
		pool:         p,
		queryTimeout: queryTimeout,
	}
}

func (r *WithdrawnRepo) GetUserWithdrawnSum(userID int) (*float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := `SELECT COALESCE(SUM(sum), 0) FROM withdrawals WHERE user_id = $1`
	var sum float64
	row := r.pool.QueryRow(ctx, query, userID)
	if err := row.Scan(&sum); err != nil {
		return nil, err
	}

	return &sum, nil
}
