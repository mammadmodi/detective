package handler

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mammadmodi/detective/pkg/htmlanalysis"
	"go.uber.org/zap"
)

// URLRequest is a struct which entry requests bind to it.
type URLRequest struct {
	URL string `json:"url"`
}

// Response is a struct which is returned to user on the analyze request.
type Response struct {
	Result *htmlanalysis.Result `json:"result"`
	Error  string               `json:"error"`
	Code   int                  `json:"code"`
}

// AnalyzeURL gets an URLRequest and analyzes the content of the html returned by url.
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
	h.Logger.With(zap.Any("request", req)).Info("request body bound successfully")

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

	htmlDoc, err := h.performGetRequest(c.Request.Context(), u)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("error while performing request")
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, &Response{
			Error: "could not retrieve html body of url",
			Code:  http.StatusPreconditionFailed,
		})
		return
	}
	h.Logger.Info("request performed successfully")

	res, err := htmlanalysis.Analyze(c.Request.Context(), u, htmlDoc)
	if err != nil {
		h.Logger.With(zap.Error(err)).Error("error while parsing html")
		c.AbortWithStatusJSON(http.StatusPreconditionFailed, &Response{
			Error: "error while parsing html",
			Code:  http.StatusPreconditionFailed,
		})
		return
	}
	h.Logger.With(zap.Any("result", res)).Info("html analyzed successfully")

	c.JSON(http.StatusOK, Response{
		Result: res,
		Code:   http.StatusOK,
	})
}

// performGetRequest performs a GET request to url returns a html string if it has.
func (h *HTTPHandler) performGetRequest(ctx context.Context, u *url.URL) (html string, err error) {
	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return "", errors.New("error while creating HTTP request")
	}

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return "", errors.New("error while performing HTTP request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("could not get a response from url")
	}

	t := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(t, "text/html") {
		return "", errors.New("response content type wasn't text/html")
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("could not read the response body")
	}

	return string(b), nil
}
