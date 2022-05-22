package main

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"

	"github.com/paramonies/ya-gophermart/internal/handlers"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

const (
	errorExitCode = 1
)

func main() {
	log.Init(os.Stdout, &log.Config{
		WithCaller: true,
		WithStack:  true,
	})

	log.Info(context.Background(), "start service")

	logLevel := convertLogLevel("debug")
	log.SetGlobalLevel(logLevel)
	log.Info(context.Background(), "updated global logging level", "newLevel", logLevel)

	err := http.ListenAndServe("localhost:8080", newRouter())
	if err != nil {
		log.Error(context.Background(), "failed to run API server", err)
		os.Exit(errorExitCode)
	}
}

func convertLogLevel(lvl string) log.Level {
	parsed, err := log.ParseLevel(lvl)
	if err != nil {
		log.Warning(context.Background(), "unknown level string, defaulting to debug level", "input", lvl)
		parsed = log.DebugLevel
	}

	return parsed
}

func newRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Get("/auth", handlers.Auth())
	r.Method("GET", "/login", handlers.Login())
	return r
}
