package handlers

import (
	"context"
	"io"
	"net/http"

	"github.com/paramonies/ya-gophermart/pkg/log"
)

func Auth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "auth handler")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "auth method")
	}
}

func Login() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug(context.Background(), "login handler")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "login method")
	})
}
