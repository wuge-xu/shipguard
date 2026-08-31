package httpserver

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthRoutes(t *testing.T) {
	handler := NewHandler()

	for _, path := range []string{
		"/health/live",
		"/health/ready",
	} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodGet,
				path,
				nil,
			)

			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf(
					"status = %d, want %d",
					rec.Code,
					http.StatusOK,
				)
			}
		})
	}
}

func TestUnknownRouteReturnsNotFound(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/not-found",
		nil,
	)

	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusNotFound,
		)
	}
}

func TestHealthRouteRejectsPost(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodPost,
		"/health/live",
		nil,
	)

	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusMethodNotAllowed,
		)
	}
}

func TestRouterAddsRequestID(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"/health/live",
		nil,
	)

	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got == "" {
		t.Fatal("X-Request-ID response header is empty")
	}
}

func TestRouterPreservesRequestID(t *testing.T) {
	const requestID = "demo-123"

	req := httptest.NewRequest(
		http.MethodGet,
		"/health/live",
		nil,
	)

	req.Header.Set(
		"X-Request-ID",
		requestID,
	)

	rec := httptest.NewRecorder()

	NewHandler().ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != requestID {
		t.Fatalf(
			"X-Request-ID = %q, want %q",
			got,
			requestID,
		)
	}
}

func TestMiddlewareChainRecoversPanic(t *testing.T) {
	var buffer bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(
			&buffer,
			nil,
		),
	)

	panicHandler := http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) {
			panic("router boom")
		},
	)

	handler := withMiddleware(
		logger,
		panicHandler,
	)

	req := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)

	req.Header.Set(
		"X-Request-ID",
		"router-panic-123",
	)

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}

	if got := rec.Header().Get("X-Request-ID"); got != "router-panic-123" {
		t.Fatalf(
			"X-Request-ID = %q, want %q",
			got,
			"router-panic-123",
		)
	}

	logOutput := buffer.String()

	if !strings.Contains(
		logOutput,
		`"request_id":"router-panic-123"`,
	) {
		t.Fatalf(
			"log = %q, want request ID",
			logOutput,
		)
	}

	if !strings.Contains(
		logOutput,
		`"status":500`,
	) {
		t.Fatalf(
			"log = %q, want access log status 500",
			logOutput,
		)
	}

	if !strings.Contains(
		logOutput,
		`"panic":"router boom"`,
	) {
		t.Fatalf(
			"log = %q, want panic evidence",
			logOutput,
		)
	}
}
