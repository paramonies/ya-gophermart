package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

func Register(db *store.PostgresDB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "register handler", "request URL", r.URL, "method", r.Method)

		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Error(context.Background(), "failed to read request body", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		log.Debug(context.Background(), "request body", string(b))

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(b)

		log.Debug(context.Background(), "user registered and authenticated ")
	}
}

func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "login handler")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "login method")
	})
}
