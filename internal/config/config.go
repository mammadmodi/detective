package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/mammadmodi/detective/pkg/logger"
)

// AppConfig is a struct which contains configuration of the application.
type AppConfig struct {
	LoggerConfig *logger.Config
	Addr         string        `default:":8000"`
	HTTPTimeout  time.Duration `split_words:"true" default:"30s"`
}

// NewAppConfig creates an AppConfig object based on the environment variables of the OS.
func NewAppConfig() (*AppConfig, error) {
	// Try to load env variables to AppConfig struct.
	c := &AppConfig{}
	err := envconfig.Process("detective", c)
	if err != nil {
		return nil, fmt.Errorf("error while processing env variables for root configs, error: %s", err.Error())
	}

	// Try to load env variables to logger.Config struct.
	loggerConfig := &logger.Config{}
	if err := envconfig.Process("detective_logger", loggerConfig); err != nil {
		return nil, fmt.Errorf("error while processing env variables for logger configs, error: %s", err.Error())
	}
	c.LoggerConfig = loggerConfig

	return c, nil
}
