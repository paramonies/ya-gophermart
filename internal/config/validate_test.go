package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate_AppConfig(t *testing.T) {
	t.Run("EmptyBootstrapAPIAddr", func(t *testing.T) {
		cfg := &AppConfig{
			RunAddress: "",
		}
		err := cfg.validate()
		if assert.Error(t, err) {
			assert.ErrorAs(t, err, new(ErrMissingOption))
			assert.Contains(t, err.Error(), "run_address")
		}
	})
}
