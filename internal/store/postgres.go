package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/paramonies/ya-gophermart/pkg/log"
)

var (
	MigDirName = "migrations"
)

type PostgresDB struct {
	Conn *pgxpool.Pool
}

func NewPostgresDB(dsn string, conTimeout time.Duration) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conTimeout)
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

	MigDirPath := fmt.Sprintf("%s/%s", rootDir, MigDirName)
	migrations := &migrate.FileMigrationSource{
		Dir: MigDirPath,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %v", err)
	}

	log.Debug(context.Background(), fmt.Sprintf("Applied %d migrations!", n))

	return &PostgresDB{Conn: conn}, nil
}
