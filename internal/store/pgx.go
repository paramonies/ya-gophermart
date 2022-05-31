package store

import (
	"time"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/paramonies/ya-gophermart/internal/store/pgx"
)

type pgxConnector struct {
	UserRepo      *pgx.UserRepo
	OrderRepo     *pgx.OrderRepo
	WithdrawnRepo *pgx.WithdrawnRepo
}

func NewPgxConnector(p *pgxpool.Pool, queryTimeout time.Duration) Connector {
	return &pgxConnector{
		UserRepo:      pgx.NewUserRepo(p, queryTimeout),
		OrderRepo:     pgx.NewOrderRepo(p, queryTimeout),
		WithdrawnRepo: pgx.NewUWithdrawnRepo(p, queryTimeout),
	}
}

func (c *pgxConnector) Users() UserRepoIf {
	return c.UserRepo
}

func (c *pgxConnector) Orders() OrderRepoIf {
	return c.OrderRepo
}

func (c *pgxConnector) Withdrawns() WithdrawnRepoIf {
	return c.WithdrawnRepo
}
