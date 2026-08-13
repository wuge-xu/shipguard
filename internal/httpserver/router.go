package httpserver

import (
	"net/http"

	"github.com/wuge-xu/shipguard/internal/health"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", health.Live)
	mux.HandleFunc("GET /health/ready", health.Ready)

	return mux
}
