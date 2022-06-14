package server

import (
	"github.com/go-chi/chi/v5"

	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
)

func NewRouter(storage store.Connector, ac *provider.AccrualClient, userHandler *UserHandler, orderHandler *OrderHandler, accrualHandler *AccrualHandler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(LogRequestInfoMiddleware)
	//r.Post("/api/user/register", Register(storage))
	r.Post("/api/user/register", userHandler.Register())
	//r.Method("POST", "/api/user/login", Login(storage))
	r.Method("POST", "/api/user/login", userHandler.Login())

	r.Route("/api/user", func(r chi.Router) {
		r.Use(AuthMiddleware(storage))
		r.Use(LoadAccrualsMiddleware(ac, storage))

		//r.Post("/orders", LoadOrder(storage))
		r.Post("/orders", orderHandler.Load())
		//r.Get("/orders", ListProcessedOrders(storage))
		r.Get("/orders", orderHandler.ListProcessed())
		r.Route("/balance", func(r chi.Router) {
			//r.Get("/", GetBalance(storage))
			r.Get("/", orderHandler.GetBalance())
			//r.Post("/withdraw", PayOrder(storage))
			r.Post("/withdraw", accrualHandler.PayOrder())
			//r.Get("/withdrawals", GetOrders(storage))
			r.Get("/withdrawals", accrualHandler.GetOrders())
		})
	})

	return r
}
