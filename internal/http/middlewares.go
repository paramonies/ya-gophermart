package http

import (
	"context"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

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
