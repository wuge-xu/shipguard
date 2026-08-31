package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/wuge-xu/shipguard/internal/health"
	"github.com/wuge-xu/shipguard/internal/middleware"
)

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health/live", health.Live)
	mux.HandleFunc("GET /health/ready", health.Ready)

	return withMiddleware(
		slog.Default(),
		mux,
	)
}

func withMiddleware(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	handler := middleware.Recovery(
		logger,
		next,
	)

	handler = middleware.AccessLog(
		logger,
		handler,
	)

	return middleware.RequestID(handler)
}
