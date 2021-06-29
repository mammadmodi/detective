package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mammadmodi/webpage-analyzer/pkg/logger"
)

// AppConfig is a struct which contains configuration of the application.
type AppConfig struct {
	LoggerConfig *logger.Config
	Host         string        `default:"127.0.0.1"`
	Port         int           `default:"8000"`
	HTTPTimeout  time.Duration `split_words:"true" default:"10s"`
}

// NewAppConfig creates an AppConfig object based on the environment variables of the OS.
func NewAppConfig() (*AppConfig, error) {
	// Try to load env variables to AppConfig struct
	c := &AppConfig{}
	err := envconfig.Process("webpage_analyzer", c)
	if err != nil {
		return nil, fmt.Errorf("error while processing env variables for root configs, error: %s", err.Error())
	}

	// Try to load env variables to logger.Config struct
	loggerConfig := &logger.Config{}
	if err := envconfig.Process("webpage_analyzer_logger", loggerConfig); err != nil {
		return nil, fmt.Errorf("error while processing env variables for logger configs, error: %s", err.Error())
	}
	c.LoggerConfig = loggerConfig

	return c, nil
}
