package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/paramonies/ya-gophermart/internal/dto"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

var (
	MigDirName             = "migrations"
	ErrConstraintViolation = errors.New("login has already occupied")
	ErrUserNotFound        = errors.New("user not found")

	ErrAlreadyCreatedByUser      = errors.New("order has already been created by user")
	ErrAlreadyCreatedByOtherUser = errors.New("order has already been created by other user")
	ErrOrdersNotFound            = errors.New("orders not found")
)

type PostgresDB struct {
	conn         *pgxpool.Pool
	queryTimeout time.Duration
}

func NewPostgresDB(dsn string, queryTimeout time.Duration) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), queryTimeout)
	defer cancel()

	conn, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to establish a connection with a PostgreSQL server: %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %v", err)
	}

	rootDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	fmt.Println("!!! rootDir ", rootDir)
	MigDirPath := fmt.Sprintf("%s/%s", rootDir, MigDirName)
	migrations := &migrate.FileMigrationSource{
		Dir: MigDirPath,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %v", err)
	}

	log.Debug(context.Background(), fmt.Sprintf("Applied %d migrations!", n))

	return &PostgresDB{
		conn:         conn,
		queryTimeout: queryTimeout}, nil
}

func (db *PostgresDB) CreateUser(userName, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
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
	row := db.conn.QueryRow(ctx, query, userName, hash)
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

func (db *PostgresDB) GetHashedPassword(login string) (*string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
	defer cancel()

	query := "SELECT password_hash From users WHERE user_name=$1"
	var hashedPassword string
	row := db.conn.QueryRow(ctx, query, login)
	if err := row.Scan(&hashedPassword); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &hashedPassword, nil
}

func (db *PostgresDB) CreateOrder(orderNumber int, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
	defer cancel()

	query := `
SELECT users.id
FROM orders
LEFT JOIN users on orders.user_id = users.id
WHERE number = $1
`

	var checkingUserID string
	row := db.conn.QueryRow(ctx, query, orderNumber)
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

	ctx2, cancel2 := context.WithTimeout(context.Background(), db.queryTimeout)
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
	row = db.conn.QueryRow(ctx2, query, orderNumber, userID)
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("failed to create order")
		}
		return err
	}
	return nil
}

func (db *PostgresDB) UpdateOrder(order dto.ProviderOrder) error {
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

	ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
	defer cancel()

	tag, err := db.conn.Exec(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("error updating order: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrOrdersNotFound
	}

	return nil
}

func (db *PostgresDB) SelectOrders(userID string) ([]dto.Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
	defer cancel()

	query := `
SELECT id,number, accrual, status, updated_at
FROM orders
LEFT JOIN users on orders.user_id = users.id
WHERE users.id = $1
`

	rows, err := db.conn.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []dto.Order
	for rows.Next() {
		var order dto.Order
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
