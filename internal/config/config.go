package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// AppConfig is a struct which contains configuration of the application.
type AppConfig struct {
	Host        int           `default:"127.0.0.1"`
	Port        int           `default:"8000"`
	LogLevel    string        `split_words:"true" default:"info"`
	HTTPTimeout time.Duration `split_words:"true" default:"10s"`
}

// NewAppConfig creates an AppConfig object based on the environment variables of the OS.
func NewAppConfig() (*AppConfig, error) {
	c := &AppConfig{}

	err := envconfig.Process("webpage_analyzer", c)
	if err != nil {
		return nil, fmt.Errorf("error while processing env variables, error: %s", err.Error())
	}

	return c, nil
}
