package middlewares

import (
	"context"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/utils"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

func VerifyCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := r.Cookie("token"); err != nil {
			log.Info(context.Background(), "user is unauthorized")
			utils.WriteErrorAsJSON(w, "user is unauthorized", err, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
