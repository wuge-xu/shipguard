package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func Recovery(
	logger *slog.Logger,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			logger.ErrorContext(
				r.Context(),
				"http handler panic",
				"method", r.Method,
				"path", r.URL.Path,
				"request_id", RequestIDFromContext(r.Context()),
				"panic", recovered,
			)

			w.Header().Set(
				"Content-Type",
				"application/json",
			)

			w.WriteHeader(
				http.StatusInternalServerError,
			)

			_ = json.NewEncoder(w).Encode(
				errorResponse{
					Error: "internal server error",
				},
			)
		}()

		next.ServeHTTP(w, r)
	})
}
