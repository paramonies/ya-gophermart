package server

import (
	"github.com/go-chi/chi/v5"

	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
)

type Handlers struct {
	UserHandler    *UserHandler
	OrderHandler   *OrderHandler
	AccrualHandler *AccrualHandler
}

func NewRouter(storage store.Connector, ac *provider.AccrualClient, h *Handlers) *chi.Mux {
	r := chi.NewRouter()

	r.Use(LogRequestInfoMiddleware)
	//r.Post("/api/user/register", Register(storage))
	r.Post("/api/user/register", h.UserHandler.Register())
	//r.Method("POST", "/api/user/login", Login(storage))
	r.Method("POST", "/api/user/login", h.UserHandler.Login())

	r.Route("/api/user", func(r chi.Router) {
		r.Use(AuthMiddleware(storage))
		//r.Use(LoadAccrualsMiddleware(ac, storage))

		//r.Post("/orders", LoadOrder(storage))
		r.Post("/orders", h.OrderHandler.Load())
		//r.Get("/orders", ListProcessedOrders(storage))
		r.Get("/orders", h.OrderHandler.ListProcessed())
		r.Route("/balance", func(r chi.Router) {
			//r.Get("/", GetBalance(storage))
			r.Get("/", h.OrderHandler.GetBalance())
			//r.Post("/withdraw", PayOrder(storage))
			r.Post("/withdraw", h.AccrualHandler.PayOrder())
			//r.Get("/withdrawals", GetOrders(storage))
			r.Get("/withdrawals", h.AccrualHandler.GetOrders())
		})
	})

	return r
}
