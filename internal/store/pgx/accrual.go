package pgx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/paramonies/ya-gophermart/internal/store/dto"
)

var (
	ErrAlreadyCreatedByUser      = errors.New("order has already been created by user")
	ErrAlreadyCreatedByOtherUser = errors.New("order has already been created by other user")
	ErrOrdersNotFound            = errors.New("orders not found")
)

type AccrualRepo struct {
	pool         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewAccrualRepo(p *pgxpool.Pool, queryTimeout time.Duration) *AccrualRepo {
	return &AccrualRepo{
		pool:         p,
		queryTimeout: queryTimeout,
	}
}

func (r *AccrualRepo) LoadOrder(orderNumber int, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf("INSERT INTO accruals (order_number, user_id) VALUES ('%d', '%s') RETURNING id", orderNumber, userID)

	_, err := r.pool.Exec(ctx, query)
	if err != nil {
		return err
	}
	return nil
}

func (r *AccrualRepo) GetOrderByOrderNumber(orderNumber int) (*dto.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	var o dto.Order
	query := fmt.Sprintf("SELECT id, order_number, accrual, user_id, order_status, updated_at FROM accruals WHERE order_number='%d'", orderNumber)
	row := r.pool.QueryRow(ctx, query)
	if err := row.Scan(&o.ID, &o.OrderNumber, &o.Accrual, &o.UserID, &o.Status, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *AccrualRepo) GetOrderByUserID(id string) (*[]dto.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.queryTimeout)
	defer cancel()

	query := fmt.Sprintf("SELECT id, order_number, accrual, user_id, order_status, updated_at FROM accruals WHERE user_id ='%s' ORDER BY updated_at ASC", id)
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []dto.Order
	for rows.Next() {
		var o dto.Order
		err := rows.Scan(&o.ID, &o.OrderNumber, &o.Accrual, &o.UserID, &o.Status, &o.UpdatedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	if rows.Err() != nil {
		return nil, err
	}

	return &orders, nil
}

func (r *AccrualRepo) UpdateOrder(order dto.ProviderOrder) error {
	pos := 1
	fields := make([]string, 0)
	values := make([]interface{}, 0)

	status, err := dto.OrderStatusToStore(order.Status)
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

func (r *AccrualRepo) SelectOrders(userID string) ([]dto.Order, error) {
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

	var orders []dto.Order
	for rows.Next() {
		var order dto.Order
		err := rows.Scan(&order.ID, &order.OrderNumber, &order.Accrual, &order.Status, &order.UpdatedAt)
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
