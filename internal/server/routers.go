package server

import (
	"github.com/go-chi/chi/v5"

	"github.com/paramonies/ya-gophermart/internal/store"
)

type Handlers struct {
	UserHandler    *UserHandler
	OrderHandler   *OrderHandler
	AccrualHandler *AccrualHandler
}

func NewRouter(storage store.Connector, h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(LogRequestInfoMiddleware)
	r.Post("/api/user/register", h.UserHandler.Register())
	r.Method("POST", "/api/user/login", h.UserHandler.Login())

	r.Route("/api/user", func(r chi.Router) {
		r.Use(AuthMiddleware(storage))

		r.Post("/orders", h.OrderHandler.Load())
		r.Get("/orders", h.OrderHandler.ListProcessed())
		r.Route("/balance", func(r chi.Router) {
			r.Get("/", h.OrderHandler.GetBalance())
			r.Post("/withdraw", h.AccrualHandler.PayOrder())
			r.Get("/withdrawals", h.AccrualHandler.GetOrders())
		})
	})

	return r
}
