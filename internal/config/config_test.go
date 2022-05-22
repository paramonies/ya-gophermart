package config

import (
	"os"
	"testing"
	"time"

	"github.com/magiconair/properties/assert"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	InitConfig()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, ":8090", cfg.App.RunAddress)
	assert.Equal(t, "debug", cfg.App.LogLevel)

	assert.Equal(t, "postgresql://postgres:123456@localhost/ya-gophermart", cfg.Database.DatabaseURI)
	assert.Equal(t, 1*time.Second, cfg.Database.QueryTimeout)

	assert.Equal(t, ":9000", cfg.ExtApp.AccrualSystemAddress)
}

func TestLoadConfig_FromFile(t *testing.T) {
	os.Args = append(os.Args, "--config=testdata/test.yaml")

	InitConfig()
	pflag.Parse()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "localhost:9090", cfg.App.RunAddress)
	assert.Equal(t, "info", cfg.App.LogLevel)

	assert.Equal(t, "postgresql://postgres:123456@localhost/ya-gophermart-test?connect_timeout=10&sslmode=disable", cfg.Database.DatabaseURI)
	assert.Equal(t, 2*time.Second, cfg.Database.QueryTimeout)

	assert.Equal(t, "localhost:9091", cfg.ExtApp.AccrualSystemAddress)
}

func TestLoadConfig_Envs(t *testing.T) {
	err := os.Setenv("RUN_ADDRESS", ":80")
	require.NoError(t, err)

	err = os.Setenv("APP_LOG_LEVEL", "error")
	require.NoError(t, err)

	err = os.Setenv("DATABASE_URI", "postgresql://192.168.1.3")
	require.NoError(t, err)

	err = os.Setenv("DB_QUERY_TIMEOUT", "2s")
	require.NoError(t, err)

	err = os.Setenv("ACCRUAL_SYSTEM_ADDRESS", ":90")
	require.NoError(t, err)

	InitConfig()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, ":80", cfg.App.RunAddress)
	assert.Equal(t, "error", cfg.App.LogLevel)
	assert.Equal(t, "postgresql://192.168.1.3", cfg.Database.DatabaseURI)
	assert.Equal(t, "2s", cfg.Database.QueryTimeout)
	assert.Equal(t, ":90", cfg.ExtApp.AccrualSystemAddress)
}
