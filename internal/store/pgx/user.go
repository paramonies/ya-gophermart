package pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/paramonies/ya-gophermart/internal/store/dto"
)

var (
	ErrConstraintViolation = errors.New("login has already exist")
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

func (r *UserRepo) GetByName(name string) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	var u dto.User
	query := fmt.Sprintf("SELECT id, user_name, password_hash, token FROM users WHERE user_name='%s'", name)
	row := r.pool.QueryRow(ctx, query)
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Token); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Create(userName, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf("INSERT INTO users (user_name, password_hash) VALUES ('%s', '%s') RETURNING id", userName, hash)

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) SetToken(userName, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf("UPDATE users SET token='%s' WHERE user_name='%s'", token, userName)

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepo) GetByToken(token string) (*dto.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	var u dto.User
	query := fmt.Sprintf("SELECT id, user_name, password_hash, token FROM users WHERE token='%s'", token)
	row := r.pool.QueryRow(ctx, query)
	if err := row.Scan(&u.ID, &u.Login, &u.PasswordHash, &u.Token); err != nil {
		return nil, err
	}
	return &u, nil
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
