package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/wuge-xu/shipguard/internal/health"
	httpHandler "github.com/wuge-xu/shipguard/internal/httpserver/handler"
	"github.com/wuge-xu/shipguard/internal/middleware"
)

func NewHandler(
	releaseHandlers ...*httpHandler.ReleaseHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health/live",
		health.Live,
	)

	mux.HandleFunc(
		"GET /health/ready",
		health.Ready,
	)

	if len(releaseHandlers) > 0 &&
		releaseHandlers[0] != nil {
		mux.HandleFunc(
			"POST /api/releases",
			releaseHandlers[0].CreateRelease,
		)
	}

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
