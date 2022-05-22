package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/paramonies/ya-gophermart/internal/config"
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

	config.InitConfig()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Error(context.Background(), "failed to load config", err)
		os.Exit(errorExitCode)
	}

	err = cfg.Validate()
	if err != nil {
		log.Error(context.Background(), "failed to validate config", err)
		os.Exit(errorExitCode)
	}

	log.Debug(context.Background(), "config params", "run_address", cfg.App.RunAddress,
		"log_level", cfg.App.LogLevel, "database_uri", cfg.Database.DatabaseURI, "query_timeout",
		cfg.Database.QueryTimeout, "accrual_system_address", cfg.ExtApp.AccrualSystemAddress)

	logLevel := convertLogLevel(cfg.App.LogLevel)
	log.SetGlobalLevel(logLevel)
	log.Info(context.Background(), "updated global logging level", "newLevel", logLevel)

	addr := cfg.App.RunAddress
	log.Info(context.Background(), "start listening API server", "address", addr)

	var srv http.Server = http.Server{
		Addr:    addr,
		Handler: newRouter(),
	}
	done := make(chan struct{})
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer func() {
			cancel()
			close(sigCh)
		}()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error(context.Background(), "failed to shut down server gracefully", err)
			os.Exit(errorExitCode)
		}
		close(done)
	}()

	if err = srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error(context.Background(), "failed to run API server", err)
		os.Exit(errorExitCode)
	}

	<-done
	log.Info(context.Background(), "server was shut down gracefully")
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
