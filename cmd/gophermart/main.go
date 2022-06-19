package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/paramonies/ya-gophermart/internal/job"

	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"

	"github.com/paramonies/ya-gophermart/internal/app"
	"github.com/paramonies/ya-gophermart/internal/config"
	"github.com/paramonies/ya-gophermart/internal/managers"
	"github.com/paramonies/ya-gophermart/internal/provider"
	"github.com/paramonies/ya-gophermart/internal/server"
	"github.com/paramonies/ya-gophermart/internal/store"
	"github.com/paramonies/ya-gophermart/pkg/log"
)

const (
	errorExitCode = 1
)

var (
	dbConnectTimeout = 1 * time.Second
	migDirName       = "migrations"
)

var appManagers managers.AppManagers

func main() {
	log.Init(os.Stdout, &log.Config{
		WithCaller: true,
		WithStack:  true,
	})

	log.Info(context.Background(), "start server")

	cfg, err := config.Init()
	if err != nil {
		log.Error(context.Background(), "failed to load config", err)
		os.Exit(errorExitCode)
	}

	log.Debug(context.Background(), "config params", "run_address", cfg.RunAddress,
		"database_uri", cfg.DatabaseURI, "query_timeout",
		"accrual_system_address", cfg.AccrualSystemAddress)

	logLevel := convertLogLevel("debug")
	log.SetGlobalLevel(logLevel)
	log.Info(context.Background(), "updated global logging level", "newLevel", logLevel)

	dbPool, err := initDatabaseConnection(cfg.DatabaseURI)
	if err != nil {
		log.Error(context.Background(), "failed to init database connection", err)
		os.Exit(errorExitCode)
	}
	dbConn := store.NewPgxConnector(dbPool, dbConnectTimeout)
	log.Info(context.Background(), "create connection to postgres DB")

	ac := provider.NewAccrualClient(cfg.AccrualSystemAddress, dbConn)
	log.Info(context.Background(), "create accrual server client")

	addr := cfg.RunAddress
	log.Info(context.Background(), "start listening API server", "address", addr)

	handlers := initHandlers(dbConn)

	var srv = http.Server{
		Addr:    addr,
		Handler: server.NewRouter(dbConn, ac, handlers),
	}
	done := make(chan struct{})

	loadAccrualJob := job.InitJob(ac, dbConn, done)
	loadAccrualJob.Run()

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer func() {
			cancel()
			close(sigCh)
		}()

		dbPool.Close()
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

func initDatabaseConnection(databaseURI string) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbConnectTimeout)
	defer cancel()

	db, err := sql.Open("postgres", databaseURI)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB: %v", err)
	}

	rootDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	MigDirPath := fmt.Sprintf("%s/%s", rootDir, migDirName)
	migrations := &migrate.FileMigrationSource{
		Dir: MigDirPath,
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %v", err)
	}

	log.Debug(context.Background(), fmt.Sprintf("Applied %d migrations!", n))

	pool, err := pgxpool.Connect(ctx, databaseURI)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func initHandlers(storage store.Connector) *server.Handlers {
	appManagers := app.NewAppManagers(storage)
	userHandler := server.NewUserHandler(appManagers.UserManager())
	orderHandler := server.NewOrderHandler(appManagers.OrderManager())
	accrualHandler := server.NewAccrualHandler(appManagers.AccrualManager())
	return &server.Handlers{
		UserHandler:    userHandler,
		OrderHandler:   orderHandler,
		AccrualHandler: accrualHandler,
	}
}
