package config

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mammadmodi/detective/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// setConfigOsEnvVariables creates a template AppConfig and sets it's related env variables to OS.
func setConfigOsEnvVariables() *AppConfig {
	c := &AppConfig{
		LoggerConfig: &logger.Config{
			Enabled:             true,
			Level:               "WARN",
			Pretty:              true,
			FileRedirectEnabled: true,
			FileRedirectPath:    "/var/log",
			FileRedirectPrefix:  "detective",
		},
		Addr:        "10.0.0.1:8080",
		HTTPTimeout: 25 * time.Second,
	}

	_ = os.Setenv("DETECTIVE_LOGGER_ENABLED", fmt.Sprint(c.LoggerConfig.Enabled))
	_ = os.Setenv("DETECTIVE_LOGGER_LEVEL", c.LoggerConfig.Level)
	_ = os.Setenv("DETECTIVE_LOGGER_PRETTY", fmt.Sprint(c.LoggerConfig.Pretty))
	_ = os.Setenv("DETECTIVE_LOGGER_FILE_REDIRECT_ENABLED", fmt.Sprint(c.LoggerConfig.FileRedirectEnabled))
	_ = os.Setenv("DETECTIVE_LOGGER_FILE_REDIRECT_PATH", c.LoggerConfig.FileRedirectPath)
	_ = os.Setenv("DETECTIVE_LOGGER_FILE_REDIRECT_PREFIX", c.LoggerConfig.FileRedirectPrefix)
	_ = os.Setenv("DETECTIVE_ADDR", c.Addr)
	_ = os.Setenv("DETECTIVE_HTTP_TIMEOUT", c.HTTPTimeout.String())

	return c
}

func TestNewConfigurationSuccess(t *testing.T) {
	expectedConfigs := setConfigOsEnvVariables()
	actualConfigs, err := NewAppConfig()
	if !assert.NoError(t, err) {
		t.Error(fmt.Errorf("error while loading configurations, error: %v", err))
		return
	}

	assert.NotNil(t, actualConfigs)
	assert.EqualValues(t, expectedConfigs, actualConfigs)
}

func TestNewConfigurationFailures(t *testing.T) {
	t.Run("error when logger config is not valid", func(t *testing.T) {
		// Enabled field must be boolean.
		_ = os.Setenv("DETECTIVE_LOGGER_ENABLED", "invalid_type")
		c, err := NewAppConfig()

		assert.Nil(t, c)
		assert.Error(t, err)
	})

	t.Run("error when http client timeout value is not valid", func(t *testing.T) {
		// Port must be integer
		_ = os.Setenv("DETECTIVE_HTTP_TIMEOUT", "invalid_duration")
		c, err := NewAppConfig()

		assert.Nil(t, c)
		assert.Error(t, err)
	})
}
