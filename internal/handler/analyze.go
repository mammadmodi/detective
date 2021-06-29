package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *HTTPHandler) Form(c *gin.Context) {
	c.String(http.StatusNoContent, "")
}

func (h *HTTPHandler) AnalyzeURL(c *gin.Context) {
	c.String(http.StatusNoContent, "")
}
