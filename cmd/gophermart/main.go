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
	"github.com/paramonies/ya-gophermart/internal/middlewares"
	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/store"
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

	db, err := store.NewPostgresDB(cfg.Database.DatabaseURI, cfg.Database.QueryTimeout)
	if err != nil {
		log.Error(context.Background(), "failed to create postgres DB connection", err)
		os.Exit(errorExitCode)
	}
	log.Info(context.Background(), "create connection to postgres DB")

	ac := provider.NewAccrualClient(cfg.ExtApp.AccrualSystemAddress)
	log.Info(context.Background(), "create accrual service client")

	addr := cfg.App.RunAddress
	log.Info(context.Background(), "start listening API server", "address", addr)

	var srv = http.Server{
		Addr:    addr,
		Handler: newRouter(db, ac),
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

func newRouter(db *store.PostgresDB, ac *provider.AccrualClient) *chi.Mux {
	r := chi.NewRouter()

	r.Post("/api/user/register", handlers.Register(db))
	r.Method("POST", "/api/user/login", handlers.Login(db))

	r.Route("/api/user", func(r chi.Router) {
		r.Use(middlewares.VerifyCookie)

		r.Post("/orders", handlers.CreateOrder(db, ac))
		r.Get("/orders", handlers.SelectOrders(db))
	})
	return r
}
