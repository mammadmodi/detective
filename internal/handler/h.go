package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// HTTPHandler handles http requests.
type HTTPHandler struct {
	// HTTPClient is used for performing Get http requests to entered urls.
	HTTPClient *http.Client
	Logger     *zap.Logger
}

// New is a factory function for HTTPHandler.
func New(logger *zap.Logger, client *http.Client) *HTTPHandler {
	return &HTTPHandler{
		HTTPClient: client,
		Logger:     logger,
	}
}

// GetRouter returns a http.Handler implementation for analyzing urls.
func (h *HTTPHandler) GetRouter() http.Handler {
	r := gin.New()
	r.GET("/analyze-url", h.Form)
	r.POST("/analyze-url", h.AnalyzeURL)

	return r
}
