package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRecoveryReturnsInternalServerError(t *testing.T) {
	var buffer bytes.Buffer

	logger := slog.New(
		slog.NewJSONHandler(&buffer, nil),
	)

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})

	handler := RequestID(
		Recovery(logger, next),
	)

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	req.Header.Set(RequestIDHeader, "panic-request-123")

	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusInternalServerError,
		)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf(
			"Content-Type = %q, want %q",
			got,
			"application/json",
		)
	}

	if got := strings.TrimSpace(rec.Body.String()); got != `{"error":"internal server error"}` {
		t.Fatalf(
			"body = %q, want internal server error JSON",
			got,
		)
	}

	logLine := buffer.String()

	if !strings.Contains(logLine, `"request_id":"panic-request-123"`) {
		t.Fatalf(
			"log = %q, want request ID",
			logLine,
		)
	}

	if !strings.Contains(logLine, `"panic":"boom"`) {
		t.Fatalf(
			"log = %q, want panic value",
			logLine,
		)
	}
}

func TestRecoveryAllowsNormalRequests(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(
			&bytes.Buffer{},
			nil,
		),
	)

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	handler := Recovery(logger, next)

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"status = %d, want %d",
			rec.Code,
			http.StatusNoContent,
		)
	}
}

func TestRecoveryDoesNotBreakSubsequentRequests(t *testing.T) {
	logger := slog.New(
		slog.NewTextHandler(
			&bytes.Buffer{},
			nil,
		),
	)

	shouldPanic := true

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if shouldPanic {
			shouldPanic = false
			panic("first request failed")
		}

		w.WriteHeader(http.StatusNoContent)
	})

	handler := Recovery(logger, next)

	firstReq := httptest.NewRequest(http.MethodGet, "/panic", nil)
	firstRec := httptest.NewRecorder()

	handler.ServeHTTP(firstRec, firstReq)

	if firstRec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"first status = %d, want %d",
			firstRec.Code,
			http.StatusInternalServerError,
		)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/ok", nil)
	secondRec := httptest.NewRecorder()

	handler.ServeHTTP(secondRec, secondReq)

	if secondRec.Code != http.StatusNoContent {
		t.Fatalf(
			"second status = %d, want %d",
			secondRec.Code,
			http.StatusNoContent,
		)
	}
}
