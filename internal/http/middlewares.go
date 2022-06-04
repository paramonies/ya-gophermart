package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

func LoadAccruals(ac *provider.AccrualClient, storage store.Connector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info(context.Background(), "try to get accruals for orders")

			user, err := getUser(r.Context())
			if err != nil {
				utils.WriteErrorAsJSON(w, "unauthorized", "failed to get user from context", err, http.StatusUnauthorized)
				return
			}

			list, err := storage.Accruals().GetPendingOrdersByUserID(user.ID)
			if err != nil {
				log.Info(context.Background(), "failed to get pending orders for user from db")
			}

			if len(*list) != 0 && err == nil {
				go func() {
					for _, or := range *list {
						err := ac.UpdateAccrual(or.OrderNumber)
						if err != nil {
							log.Error(context.Background(), fmt.Sprintf("failed to update %s order for user %s", or.ID, user.Login), err)
						}
					}
				}()
			}

			next.ServeHTTP(w, r)
		})
	}
}

func Auth(storage store.Connector) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newReq := r
			cookie, err := r.Cookie("token")
			if err != nil {
				log.Info(context.Background(), "user is unauthorized")
				utils.WriteErrorAsJSON(w, "user is unauthorized", "failed to authorize user", err, http.StatusUnauthorized)
				return
			} else {
				user, err := storage.Users().GetByToken(cookie.Value)
				if err != nil {
					log.Info(context.Background(), "user is unauthorized")
					utils.WriteErrorAsJSON(w, "user is unauthorized", "failed to authorize user", err, http.StatusUnauthorized)
					return
				}
				newReq = r.WithContext(context.WithValue(r.Context(), User, user))
			}

			next.ServeHTTP(w, newReq)
		})
	}
}

func LogRequestInfo(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "=======", "request URL", r.URL, "method", r.Method)
		next.ServeHTTP(w, r)
	})
}
