package main

import (
	"github.com/mammadmodi/webpage-analyzer/internal/config"
	"github.com/mammadmodi/webpage-analyzer/pkg/logger"
	"go.uber.org/zap"
)

var c *config.AppConfig
var l *zap.Logger

func init() {
	var err error
	c, err = config.NewAppConfig()
	if err != nil {
		panic(err)
	}

	l, err = logger.NewZapLogger("webpage_analyzer", c.LoggerConfig)
	if err != nil {
		panic(err)
	}
}
func main() {
	// TODO Run the web server.
}
