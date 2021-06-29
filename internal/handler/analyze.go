package handler

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type URLRequest struct {
	URL string `json:"url"`
}

type Headings struct {
	H1, H2, H3, H4, H5, H6 uint
}

type Result struct {
	HTMLVersion       string    `json:"html_version"`
	PageTitle         string    `json:"page_title"`
	HeadingsCount     *Headings `json:"headings"`
	InternalLinks     uint      `json:"internal_links"`
	ExternalLinks     uint      `json:"external_links"`
	InaccessibleLinks uint      `json:"inaccessible_links"`
	HasLoginForm      bool      `json:"has_login_form"`
}

type Response struct {
	Result Result `json:"result"`
	Error  string `json:"error"`
	Code   int    `json:"code"`
}

// AnalyzeURL gets an url in request and analyzes the content of the html returned by url.
func (h *HTTPHandler) AnalyzeURL(c *gin.Context) {
	req := URLRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.Logger.With(zap.Error(err)).Error("error while binding request body")
		c.AbortWithStatusJSON(http.StatusBadRequest, &Response{
			Error: "cannot parse request body",
			Code:  http.StatusNotAcceptable,
		})
		return
	}
	h.Logger.With(zap.Any("request", req)).Info("request bound successfully")

	u, err := url.ParseRequestURI(req.URL)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("requested url is not valid")
		c.AbortWithStatusJSON(http.StatusBadRequest, &Response{
			Error: "entered url is not valid",
			Code:  http.StatusBadRequest,
		})
		return
	}
	h.Logger.With(zap.Any("entered_url", u)).Info("entered url parsed successfully")

	res := Response{}
	// TODO load res by the html of the url.
	c.JSON(http.StatusOK, res)
}
