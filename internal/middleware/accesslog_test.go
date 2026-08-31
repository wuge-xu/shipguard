package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAccessLogRecordsExplicitStatus(t *testing.T) {
	var buffer bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(&buffer, nil),
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := RequestID(
		AccessLog(logger, next),
	)

	req := httptest.NewRequest(http.MethodPost, "/releases", nil)
	req.Header.Set(RequestIDHeader, "request-123")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logLine := buffer.String()

	if !strings.Contains(logLine, `"status":201`) {
		t.Fatalf("log = %q, want status 201", logLine)
	}

	if !strings.Contains(logLine, `"request_id":"request-123"`) {
		t.Fatalf("log = %q, want request ID", logLine)
	}
}

func TestAccessLogDefaultsStatusToOK(t *testing.T) {
	var buffer bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(&buffer, nil),
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	handler := RequestID(
		AccessLog(logger, next),
	)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !strings.Contains(buffer.String(), `"status":200`) {
		t.Fatalf(
			"log = %q, want status 200",
			buffer.String(),
		)
	}
}

func TestAccessLogRecordsMethodAndPath(t *testing.T) {
	var buffer bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(&buffer, nil),
	)

	next := http.NotFoundHandler()

	handler := RequestID(
		AccessLog(logger, next),
	)

	req := httptest.NewRequest(http.MethodGet, "/missing", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	logLine := buffer.String()

	if !strings.Contains(logLine, `"method":"GET"`) {
		t.Fatalf("log = %q, want method GET", logLine)
	}

	if !strings.Contains(logLine, `"path":"/missing"`) {
		t.Fatalf("log = %q, want path /missing", logLine)
	}

	if !strings.Contains(logLine, `"status":404`) {
		t.Fatalf("log = %q, want status 404", logLine)
	}
}
