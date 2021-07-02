package handler

import (
	"context"
	"net/http"
	"net/url"

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
