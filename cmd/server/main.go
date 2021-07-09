package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mammadmodi/detective/internal/config"
	"github.com/mammadmodi/detective/internal/handler"
	"github.com/mammadmodi/detective/pkg/htmlanalysis"
	"github.com/mammadmodi/detective/pkg/logger"
	"go.uber.org/zap"
)

const AsciiArt = `
 ______   _______ _________ _______  _______ __________________          _______ 
(  __  \ (  ____ \\__   __/(  ____ \(  ____ \\__   __/\__   __/|\     /|(  ____ \
| (  \  )| (    \/   ) (   | (    \/| (    \/   ) (      ) (   | )   ( || (    \/
| |   ) || (__       | |   | (__    | |         | |      | |   | |   | || (__    
| |   | ||  __)      | |   |  __)   | |         | |      | |   ( (   ) )|  __)   
| |   ) || (         | |   | (      | |         | |      | |    \ \_/ / | (      
| (__/  )| (____/\   | |   | (____/\| (____/\   | |   ___) (___  \   /  | (____/\
(______/ (_______/   )_(   (_______/(_______/   )_(   \_______/   \_/   (_______/

Version: __commit_ref_name__ (__commit_sha__)
Build Date: __build_date__
`

// Following variables must be loaded in build time.
var (
	CommitSHA     string
	CommitRefName string
	BuildDate     string
)

var c *config.AppConfig
var l *zap.Logger
var hc *http.Client
var h *handler.HTTPHandler
var r *gin.Engine

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
	// We should reduce the IdleConnTimeout because the requests that are being performed
	// by this HTTPClient target different hosts and there is no meaning to have an idle connection
	// for a long time.
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

	// Create application router.
	r = gin.New()
	r.Static("/", "./web/static/")
	r.POST("/analyze-url", h.AnalyzeURL)

	l.With(zap.Any("configs", c)).Info("application initialized successfully")
}

func main() {
	fmt.Println(
		strings.NewReplacer(
			"__commit_ref_name__", CommitRefName,
			"__commit_sha__", CommitSHA,
			"__build_date__", BuildDate,
		).Replace(AsciiArt))

	// Launch server and listen to application port.
	if err := http.ListenAndServe(c.Addr, r); err != nil && err != http.ErrServerClosed {
		l.With(zap.Error(err)).Panic("error while running gin http server")
	}

	os.Exit(0)
}
