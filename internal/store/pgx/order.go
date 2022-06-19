package pgx

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/paramonies/ya-gophermart/internal/store/dto"
)

type OrderRepo struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewOrderRepo(p *pgxpool.Pool, queryTimeout time.Duration) *OrderRepo {
	return &OrderRepo{
		pool:         p,
		queryTimeout: queryTimeout,
	}
}

func (r OrderRepo) GetOrdersPriceForUser(userID string) (*float64, error) {
	var totalPrice float64
	query := fmt.Sprintf("SELECT COALESCE(SUM(price),0) FROM orders WHERE user_id = '%s';", userID)
	err := r.pool.QueryRow(context.Background(), query).Scan(&totalPrice)
	if err != nil {
		return nil, err
	}

	return &totalPrice, nil
}

func (r OrderRepo) Register(userID string, orderNumber string, price float64) error {
	query := fmt.Sprintf("INSERT INTO orders (user_id, order_number, price) VALUES ('%s', '%s', %f);", userID, orderNumber, price)
	_, err := r.pool.Exec(context.Background(), query)
	if err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgerr.SQLState()) {
				return ErrConstraintViolationOrder
			}
		}
		return err
	}

	return nil
}

func (r OrderRepo) GetOrdersByUserID(userID string) (*[]dto.Order, error) {
	query := fmt.Sprintf("SELECT id, user_id, order_number, price, updated_at FROM orders WHERE user_id = '%s' ORDER BY updated_at ASC", userID)
	rows, err := r.pool.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make([]dto.Order, 0)
	for rows.Next() {
		var or dto.Order
		err := rows.Scan(&or.ID, &or.UserID, &or.OrderNumber, &or.Price, &or.UpdatedAt)
		if err != nil {
			return nil, nil
		}
		res = append(res, or)
	}

	return &res, nil
}
