package httpserver

import (
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	handler := http.NewServeMux()

	server := New(":9090", handler)

	if server.Addr != ":9090" {
		t.Fatalf("Addr = %q, want %q", server.Addr, ":9090")
	}
	if server.Handler != handler {
		t.Fatal("Handler was not preserved")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("ReadHeaderTimeout = %s, want %s", server.ReadHeaderTimeout, 5*time.Second)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("IdleTimeout = %s, want %s", server.IdleTimeout, 60*time.Second)
	}
}
