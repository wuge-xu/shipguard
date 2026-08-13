package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/wuge-xu/shipguard/internal/config"
	"github.com/wuge-xu/shipguard/internal/httpserver"
)

func main() {
	if err := run(); err != nil {
		slog.Error("shipguard-api stopped with error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	server := httpserver.New(cfg.HTTPAddr, httpserver.NewHandler())

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting shipguard-api", "addr", cfg.HTTPAddr)
		serverErr <- server.ListenAndServe()
	}()

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	select {
	case <-signalCtx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		cfg.ShutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown HTTP server: %w", err)
	}

	if err := <-serverErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP after shutdown: %w", err)
	}

	slog.Info("shipguard-api stopped")
	return nil
}
