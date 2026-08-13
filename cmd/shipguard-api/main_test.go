package main

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/wuge-xu/shipguard/internal/httpserver"
)

func TestServeStopsWhenContextCanceled(t *testing.T) {
	addr := reserveAddr(t)

	server := httpserver.New(
		addr,
		httpserver.NewHandler(),
	)

	ctx, cancel := context.WithCancel(context.Background())

	result := make(chan error, 1)
	go func() {
		result <- serve(ctx, server, time.Second)
	}()

	waitForHealthy(t, "http://"+addr+"/health/live")

	cancel()

	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("serve() error = %v, want nil", err)
		}

	case <-time.After(2 * time.Second):
		t.Fatal("serve() did not stop after context cancellation")
	}

	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	resp, err := client.Get("http://" + addr + "/health/live")
	if err == nil {
		resp.Body.Close()
		t.Fatal("server still accepts requests after shutdown")
	}
}

func TestServeReturnsListenError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}
	defer listener.Close()

	server := httpserver.New(
		listener.Addr().String(),
		httpserver.NewHandler(),
	)

	err = serve(
		context.Background(),
		server,
		time.Second,
	)

	if err == nil {
		t.Fatal("serve() error = nil, want listen error")
	}

	if !strings.Contains(err.Error(), "serve HTTP") {
		t.Fatalf("serve() error = %q, want serve HTTP error", err)
	}
}

func reserveAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("net.Listen() error = %v", err)
	}

	addr := listener.Addr().String()

	if err := listener.Close(); err != nil {
		t.Fatalf("listener.Close() error = %v", err)
	}

	return addr
}

func waitForHealthy(t *testing.T, url string) {
	t.Helper()

	client := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	deadline := time.Now().Add(2 * time.Second)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return
			}
		}

		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("server did not become healthy at %s", url)
}
