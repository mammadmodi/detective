package main

import (
	"github.com/mammadmodi/detective/internal/config"
	"github.com/mammadmodi/detective/pkg/logger"
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

	l, err = logger.NewZapLogger("detective", c.LoggerConfig)
	if err != nil {
		panic(err)
	}
}
func main() {
	// TODO Run the web server.
}
