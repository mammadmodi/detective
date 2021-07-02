package main

import (
	"github.com/gin-gonic/gin"
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
var h *handler.HTTPHandler

func init() {
	var err error
	// Initialize application configuration.
	c, err = config.NewAppConfig()
	if err != nil {
		panic(err)
	}

	// Initialize application logger.
	l, err = logger.NewZapLogger("detective", c.LoggerConfig)
	if err != nil {
		panic(err)
	}

	// Initialize application HTTP client.
	hc = &http.Client{
		Timeout: c.HTTPTimeout,
		Transport: &http.Transport{
			IdleConnTimeout: 5 * time.Second,
		},
	}

	h = &handler.HTTPHandler{
		HTTPClient:      hc,
		Logger:          l.Named("http_handler"),
		HTMLAnalyzeFunc: htmlanalysis.Analyze,
	}

	// Setup package level dependencies.
	hcClone := *hc
	htmlanalysis.SetGlobalLogger(l.Named("html_analyzer"))
	htmlanalysis.SetGlobalHTTPClient(&hcClone)

	l.With(zap.Any("configs", c)).Info("application initialized successfully")
}

func main() {
	// Create http server.
	r := gin.New()
	r.GET("/analyze-url", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./web/static/form.html")
	})
	r.POST("/analyze-url", h.AnalyzeURL)

	// Launch server and listen to application port.
	if err := http.ListenAndServe(c.Addr, r); err != nil && err != http.ErrServerClosed {
		l.With(zap.Error(err)).Panic("error while running gin http server")
	}

	os.Exit(0)
}
