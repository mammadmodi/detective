package app

import (
	"net/http"

	"go.uber.org/zap"
)

// App is the struct which holds main dependencies of the application.
type App struct {
	logger     *zap.Logger
	httpClient *http.Client
	server     *http.Server
}
