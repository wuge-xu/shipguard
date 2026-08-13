package health

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLive(t *testing.T) {
	testOKHandler(t, Live)
}

func TestReady(t *testing.T) {
	testOKHandler(t, Ready)
}

func testOKHandler(t *testing.T, handler http.HandlerFunc) {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}

	if got := strings.TrimSpace(rec.Body.String()); got != `{"status":"ok"}` {
		t.Fatalf("body = %q, want %q", got, `{"status":"ok"}`)
	}
}
