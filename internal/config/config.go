package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"127.0.0.1:8081"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"server://127.0.0.1:8080"`
	//DatabaseURI          string `env:"DATABASE_URI" envDefault:"postgres://postgres:123456@localhost:5432/ya-gophermart?sslmode=disable"`
	DatabaseURI string `env:"DATABASE_URI"`
}

func Init() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "api server host and port")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "address of external accrual system")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "path to DB-file on disk")

	flag.Parse()

	return &cfg, nil
}
