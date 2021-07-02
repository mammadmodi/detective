package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/mammadmodi/detective/pkg/htmlanalysis"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		HTTPClient: http.DefaultClient,
		Logger:     zap.NewNop(),
	}
}

func TestHTTPHandler_AnalyzeURLIncorrectBody(t *testing.T) {
	h := newTestHTTPHandler()
	gin.SetMode(gin.TestMode)

	res := httptest.NewRecorder()
	ginCtx, r := gin.CreateTestContext(res)
	r.POST("/analyze-url", h.AnalyzeURL)

	ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader("invalid_request_json"))
	r.ServeHTTP(res, ginCtx.Request)

	var actualResponse Response
	_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

	expectedResponse := Response{
		Error: "cannot parse request body",
		Code:  http.StatusNotAcceptable,
	}
	assert.Equal(t, expectedResponse, actualResponse)
	assert.Equal(t, http.StatusNotAcceptable, res.Code)
}

func TestHTTPHandler_AnalyzeURLBadRequest(t *testing.T) {
	h := newTestHTTPHandler()
	gin.SetMode(gin.TestMode)

	ur := URLRequest{
		URL: "invalid_url",
	}
	b, err := json.Marshal(ur)
	if err != nil {
		t.Logf("error while marshalling url request, err: %v", err)
		return
	}

	res := httptest.NewRecorder()
	ginCtx, r := gin.CreateTestContext(res)
	r.POST("/analyze-url", h.AnalyzeURL)

	ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader(string(b)))
	r.ServeHTTP(res, ginCtx.Request)

	var actualResponse Response
	_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

	expectedResponse := Response{
		Error: "entered url is not valid",
		Code:  http.StatusBadRequest,
	}
	assert.Equal(t, expectedResponse, actualResponse)
	assert.Equal(t, http.StatusBadRequest, res.Code)
}

func TestHTTPHandler_AnalyzeURLInaccessibleURL(t *testing.T) {
	h := newTestHTTPHandler()
	gin.SetMode(gin.TestMode)

	t.Run("when url is inaccessible", func(t *testing.T) {
		ur := URLRequest{
			URL: "http://localhost:22222/",
		}
		b, err := json.Marshal(ur)
		if err != nil {
			t.Logf("error while marshalling url request, err: %v", err)
			return
		}

		res := httptest.NewRecorder()
		ginCtx, r := gin.CreateTestContext(res)
		r.POST("/analyze-url", h.AnalyzeURL)

		ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader(string(b)))
		r.ServeHTTP(res, ginCtx.Request)

		var actualResponse Response
		_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

		expectedResponse := Response{
			Error: "could not retrieve html body of url",
			Code:  http.StatusPreconditionFailed,
		}
		assert.Equal(t, expectedResponse, actualResponse)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("when url doesn't return 2xx response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
			res.WriteHeader(http.StatusNotFound)
		}))
		serverURL, _ := url.Parse(server.URL)
		ur := URLRequest{
			URL: serverURL.String(),
		}
		defer server.Close()

		b, err := json.Marshal(ur)
		if err != nil {
			t.Logf("error while marshalling url request, err: %v", err)
			return
		}

		res := httptest.NewRecorder()
		ginCtx, r := gin.CreateTestContext(res)
		r.POST("/analyze-url", h.AnalyzeURL)

		ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader(string(b)))
		r.ServeHTTP(res, ginCtx.Request)

		var actualResponse Response
		_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

		expectedResponse := Response{
			Error: "could not retrieve html body of url",
			Code:  http.StatusPreconditionFailed,
		}
		assert.Equal(t, expectedResponse, actualResponse)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})

	t.Run("when url returns non html response in body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusOK)
		}))
		serverURL, _ := url.Parse(server.URL)
		ur := URLRequest{
			URL: serverURL.String(),
		}
		defer server.Close()

		b, err := json.Marshal(ur)
		if err != nil {
			t.Logf("error while marshalling url request, err: %v", err)
			return
		}

		res := httptest.NewRecorder()
		ginCtx, r := gin.CreateTestContext(res)
		r.POST("/analyze-url", h.AnalyzeURL)

		ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader(string(b)))
		r.ServeHTTP(res, ginCtx.Request)

		var actualResponse Response
		_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

		expectedResponse := Response{
			Error: "could not retrieve html body of url",
			Code:  http.StatusPreconditionFailed,
		}
		assert.Equal(t, expectedResponse, actualResponse)
		assert.Equal(t, http.StatusPreconditionFailed, res.Code)
	})
}

func TestHTTPHandler_AnalyzeURLNotParsableHTML(t *testing.T) {
	h := newTestHTTPHandler()
	gin.SetMode(gin.TestMode)

	h.HTMLAnalyzeFunc = func(_ context.Context, _ *url.URL, _ string) (*htmlanalysis.Result, error) {
		return nil, errors.New("cannot parse html")
	}
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.Header().Set("Content-Type", "text/html")
		res.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(res, "<!DOCTYPE html>")
	}))
	serverURL, _ := url.Parse(server.URL)
	ur := URLRequest{
		URL: serverURL.String(),
	}
	defer server.Close()

	b, err := json.Marshal(ur)
	if err != nil {
		t.Logf("error while marshalling url request, err: %v", err)
		return
	}

	res := httptest.NewRecorder()
	ginCtx, r := gin.CreateTestContext(res)
	r.POST("/analyze-url", h.AnalyzeURL)

	ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader(string(b)))
	r.ServeHTTP(res, ginCtx.Request)

	var actualResponse Response
	_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

	expectedResponse := Response{
		Error: "error while parsing html",
		Code:  http.StatusPreconditionFailed,
	}
	assert.Equal(t, expectedResponse, actualResponse)
	assert.Equal(t, http.StatusPreconditionFailed, res.Code)
}

func TestHTTPHandler_AnalyzeURLSuccess(t *testing.T) {
	h := newTestHTTPHandler()
	gin.SetMode(gin.TestMode)

	expectedResult := htmlanalysis.Result{
		HTMLVersion: "HTML 5",
		PageTitle:   "Detective",
	}
	h.HTMLAnalyzeFunc = func(_ context.Context, _ *url.URL, _ string) (*htmlanalysis.Result, error) {
		return &expectedResult, nil
	}
	server := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, _ *http.Request) {
		res.Header().Set("Content-Type", "text/html")
		res.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(res, "<!DOCTYPE html>")
	}))
	serverURL, _ := url.Parse(server.URL)
	ur := URLRequest{
		URL: serverURL.String(),
	}
	defer server.Close()

	b, err := json.Marshal(ur)
	if err != nil {
		t.Logf("error while marshalling url request, err: %v", err)
		return
	}

	res := httptest.NewRecorder()
	ginCtx, r := gin.CreateTestContext(res)
	r.POST("/analyze-url", h.AnalyzeURL)

	ginCtx.Request, _ = http.NewRequest(http.MethodPost, "/analyze-url", strings.NewReader(string(b)))
	r.ServeHTTP(res, ginCtx.Request)

	var actualResponse Response
	_ = json.Unmarshal(res.Body.Bytes(), &actualResponse)

	assert.Equal(t, expectedResult, *actualResponse.Result)
	assert.Equal(t, http.StatusOK, actualResponse.Code)
	assert.Empty(t, actualResponse.Error)
	assert.Equal(t, http.StatusOK, res.Code)
}
