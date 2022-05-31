package pgx

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrConstraintViolation = errors.New("login has already occupied")
	ErrUserNotFound        = errors.New("user not found")
)

type UserRepo struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewUserRepo(p *pgxpool.Pool, queryTimeout time.Duration) *UserRepo {
	return &UserRepo{
		pool:         p,
		queryTimeout: queryTimeout,
	}
}

func (r *UserRepo) CreateUser(userName, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := `
INSERT INTO users 
(
    user_name,
    password_hash
)
VALUES ($1, $2)
RETURNING id
`

	var id string
	row := r.pool.QueryRow(ctx, query, userName, hash)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("failed to create user")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.SQLState()) {
				return ErrConstraintViolation
			}
		}
		return err
	}
	return nil
}

func (r *UserRepo) GetHashedPassword(login string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := "SELECT password_hash From users WHERE user_name=$1"
	var hashedPassword string
	row := r.pool.QueryRow(ctx, query, login)
	if err := row.Scan(&hashedPassword); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &hashedPassword, nil
}
