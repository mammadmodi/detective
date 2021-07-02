package main

import (
	"net/http"
	"os"
	"time"

	"github.com/mammadmodi/detective/internal/config"
	"github.com/mammadmodi/detective/internal/handler"
	"github.com/mammadmodi/detective/pkg/htmlanalysis"
	"github.com/mammadmodi/detective/pkg/logger"
	"go.uber.org/zap"
)

var c *config.AppConfig
var l *zap.Logger
var hc *http.Client

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

	hc = &http.Client{
		Timeout: c.HTTPTimeout,
		Transport: &http.Transport{
			IdleConnTimeout: 5 * time.Second,
		},
	}

	l.With(zap.Any("configs", c)).Info("application initialized successfully")
}

func main() {
	// Setup package level dependencies.
	hcClone := *hc
	htmlanalysis.SetGlobalLogger(l.Named("html_analyzer"))
	htmlanalysis.SetGlobalHTTPClient(&hcClone)

	// Create http server.
	h := handler.New(l.Named("http_handler"), hc, htmlanalysis.Analyze)
	server := &http.Server{
		Addr:    c.Addr,
		Handler: h.GetRouter(),
	}

	// Launch server and listen to application port.
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.With(zap.Error(err)).Panic("error while running gin http server")
	}

	os.Exit(0)
}
