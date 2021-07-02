package handler

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/mammadmodi/detective/pkg/htmlanalysis"
	"go.uber.org/zap"
)

// HTMLAnalyzeFunc is type of function which analyzes an html doc and returns a Result object.
type HTMLAnalyzeFunc func(ctx context.Context, url *url.URL, htmlDoc string) (*htmlanalysis.Result, error)

// HTTPHandler handles http requests.
// HTTPClient is used for performing Get http requests to entered urls.
type HTTPHandler struct {
	HTTPClient      *http.Client
	Logger          *zap.Logger
	HTMLAnalyzeFunc HTMLAnalyzeFunc
}

// New is a factory function for HTTPHandler.
func New(logger *zap.Logger, client *http.Client, analyzeFunc HTMLAnalyzeFunc) *HTTPHandler {
	return &HTTPHandler{
		HTTPClient:      client,
		Logger:          logger,
		HTMLAnalyzeFunc: analyzeFunc,
	}
}

// GetRouter returns a http.Handler implementation for analyzing urls.
func (h *HTTPHandler) GetRouter() http.Handler {
	r := gin.New()
	r.GET("/analyze-url", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "./web/static/form.html")
	})
	r.POST("/analyze-url", h.AnalyzeURL)

	return r
}
