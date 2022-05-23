package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var configPath string

const (
	defaultServerAPIAddr = ":8090"
	defaultLoggingLevel  = "debug"

	defaultDatabaseURI          = "postgresql://postgres:123456@localhost/ya-gophermart?connect_timeout=10&sslmode=disable"
	defaultDatabaseQueryTimeout = 1 * time.Second

	defaultAccrualSystemAddress = ":9000"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"db"`
	ExtApp   ExtAppConfig   `mapstructure:"ext_app"`
}

func (cfg *Config) Validate() error {
	err := cfg.App.validate()
	if err != nil {
		return fmt.Errorf("bad app configuration: %s", err)
	}

	return nil
}

type AppConfig struct {
	RunAddress string `mapstructure:"run_address"`
	LogLevel   string `mapstructure:"log_level"`
}

type DatabaseConfig struct {
	DatabaseURI  string        `mapstructure:"database_uri"`
	QueryTimeout time.Duration `mapstructure:"query_timeout"`
}

type ExtAppConfig struct {
	AccrualSystemAddress string `mapstructure:"accrual_system_address"`
}

var once = new(sync.Once)

func InitConfig() {
	once.Do(initViper)
}

func initViper() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind command line arguments.
	pflag.StringVar(&configPath, "config", "", "the configuration file path")

	pflag.String("a", defaultServerAPIAddr, "address of the server API to listen (env: RUN_ADDRESS)")
	pflag.String("log-level", defaultLoggingLevel, "the application log level: debug, info, warn, error (env: APP_LOG_LEVEL)")

	pflag.String("d", defaultDatabaseURI, "the database connection URL (env: DATABASE_URI)")
	pflag.Duration("db-query-timeout", defaultDatabaseQueryTimeout, "the database query timeout (env: DB_QUERY_TIMEOUT)")

	pflag.String("r", defaultAccrualSystemAddress, "address of the external accrual system (env: ACCRUAL_SYSTEM_ADDRESS)")

	pflag.Parse()

	_ = viper.BindPFlag("app.run_address", pflag.Lookup("a"))
	_ = viper.BindPFlag("app.log_level", pflag.Lookup("log-level"))

	_ = viper.BindPFlag("db.database_uri", pflag.Lookup("d"))
	_ = viper.BindPFlag("db.query_timeout", pflag.Lookup("db-query-timeout"))

	_ = viper.BindPFlag("ext_app.accrual_system_address", pflag.Lookup("r"))
}

func LoadConfig() (*Config, error) {
	if configPath != "" {
		viper.SetConfigFile(configPath)
		err := viper.ReadInConfig()
		if err != nil {
			return nil, fmt.Errorf("error reading config file: %s", err)
		}
	}

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %s", err)
	}

	return &config, nil
}
