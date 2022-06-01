package http

import (
	"github.com/go-chi/chi/v5"

	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
)

func NewRouter(storage store.Connector, ac *provider.AccrualClient) *chi.Mux {
	r := chi.NewRouter()

	r.Use(LogRequestInfo)
	r.Post("/api/user/register", Register(storage))
	r.Method("POST", "/api/user/login", Login(storage))

	r.Route("/api/user", func(r chi.Router) {
		r.Use(Auth(storage))

		r.Post("/orders", CreateOrder(storage, ac))
		//r.Get("/orders", SelectOrders(storage))
		//r.Route("/balance", func(r chi.Router) {
		//	r.Get("/", GetBalance(storage))
		//})
	})

	return r
}