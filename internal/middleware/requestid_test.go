package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDGeneratesID(t *testing.T) {
	var contextRequestID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextRequestID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	responseRequestID := rec.Header().Get(RequestIDHeader)

	if responseRequestID == "" {
		t.Fatal("response request ID is empty")
	}

	if contextRequestID == "" {
		t.Fatal("context request ID is empty")
	}

	if responseRequestID != contextRequestID {
		t.Fatalf(
			"response request ID = %q, context request ID = %q",
			responseRequestID,
			contextRequestID,
		)
	}
}

func TestRequestIDPreservesExistingID(t *testing.T) {
	const requestID = "client-request-123"

	var contextRequestID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextRequestID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, requestID)

	rec := httptest.NewRecorder()

	RequestID(next).ServeHTTP(rec, req)

	if got := rec.Header().Get(RequestIDHeader); got != requestID {
		t.Fatalf(
			"response request ID = %q, want %q",
			got,
			requestID,
		)
	}

	if contextRequestID != requestID {
		t.Fatalf(
			"context request ID = %q, want %q",
			contextRequestID,
			requestID,
		)
	}
}

func TestRequestIDFromContextReturnsEmptyWhenMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := RequestIDFromContext(req.Context()); got != "" {
		t.Fatalf(
			"RequestIDFromContext() = %q, want empty string",
			got,
		)
	}
}
