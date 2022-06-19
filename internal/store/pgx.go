package store

import (
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/paramonies/ya-gophermart/internal/store/pgx"
)

type pgxConnector struct {
	UserRepo    *pgx.UserRepo
	AccrualRepo *pgx.AccrualRepo
	OrderRepo   *pgx.OrderRepo
}

func NewPgxConnector(p *pgxpool.Pool, queryTimeout time.Duration) Connector {
	return &pgxConnector{
		UserRepo:    pgx.NewUserRepo(p, queryTimeout),
		AccrualRepo: pgx.NewAccrualRepo(p, queryTimeout),
		OrderRepo:   pgx.NewOrderRepo(p, queryTimeout),
	}
}

func (c *pgxConnector) Users() UserRepoIf {
	return c.UserRepo
}

func (c *pgxConnector) Accruals() AccrualRepoIf {
	return c.AccrualRepo
}

func (c *pgxConnector) Orders() OrderRepoIf {
	return c.OrderRepo
}
