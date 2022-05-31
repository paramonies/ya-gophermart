package pgx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	dto2 "github.com/paramonies/ya-gophermart/internal/store/dto"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

var (
	ErrAlreadyCreatedByUser      = errors.New("order has already been created by user")
	ErrAlreadyCreatedByOtherUser = errors.New("order has already been created by other user")
	ErrOrdersNotFound            = errors.New("orders not found")
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

func (r *OrderRepo) CreateOrder(orderNumber int, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := `
SELECT users.id
FROM orders
LEFT JOIN users on orders.user_id = users.id
WHERE number = $1
`

	var checkingUserID string
	row := r.pool.QueryRow(ctx, query, orderNumber)
	if err := row.Scan(&checkingUserID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	if checkingUserID != "" {
		if checkingUserID == userID {
			return ErrAlreadyCreatedByOtherUser
		} else {
			return ErrAlreadyCreatedByOtherUser
		}
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel2()

	query = `
INSERT INTO orders
(
    number,
    user_id
)
VALUES ($1, $2)
RETURNING id
`

	var id string
	row = r.pool.QueryRow(ctx2, query, orderNumber, userID)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("failed to create order")
		}
		return err
	}
	return nil
}

func (r *OrderRepo) UpdateOrder(order dto2.ProviderOrder) error {
	pos := 1
	fields := make([]string, 0)
	values := make([]interface{}, 0)

	status, err := dto2.OrderStatusToStore(order.Status)
	if err != nil {
		return nil
	}

	fields = append(fields, fmt.Sprintf("status=$%d", pos))
	values = append(values, string(status))
	pos++

	fields = append(fields, fmt.Sprintf("accrual=$%d", pos))
	values = append(values, order.Accrual)
	pos++

	values = append(values, order.Number)
	query := fmt.Sprintf("UPDATE orders SET %s WHERE number=$%d", strings.Join(fields, ","), pos)

	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	tag, err := r.pool.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("error updating order: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrOrdersNotFound
	}

	return nil
}

func (r *OrderRepo) SelectOrders(userID string) ([]dto2.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := `
SELECT id,number, accrual, status, updated_at
FROM orders
LEFT JOIN users on orders.user_id = users.id
WHERE users.id = $1
`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []dto2.Order
	for rows.Next() {
		var order dto2.Order
		err := rows.Scan(&order.ID, &order.Number, &order.Accural, &order.Status, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}

	if len(orders) == 0 {
		return nil, ErrOrdersNotFound
	}

	return orders, nil
}
