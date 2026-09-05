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
	"time"

	releaseapp "github.com/wuge-xu/shipguard/internal/application/release"
	"github.com/wuge-xu/shipguard/internal/config"
	"github.com/wuge-xu/shipguard/internal/database"
	"github.com/wuge-xu/shipguard/internal/httpserver"
	httpHandler "github.com/wuge-xu/shipguard/internal/httpserver/handler"
	"github.com/wuge-xu/shipguard/internal/repository"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	slog.SetDefault(logger)

	signalCtx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	if err := run(signalCtx); err != nil {
		slog.Error(
			"shipguard-api stopped with error",
			"error",
			err,
		)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf(
			"load config: %w",
			err,
		)
	}

	pool, err := database.Open(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		return fmt.Errorf(
			"open database: %w",
			err,
		)
	}
	defer pool.Close()

	releaseRepository := repository.NewReleaseRepository(
		pool,
	)

	releaseService := releaseapp.NewService(
		releaseRepository,
		nil,
	)

	releaseHandler := httpHandler.NewReleaseHandler(
		releaseService,
	)

	server := httpserver.New(
		cfg.HTTPAddr,
		httpserver.NewHandler(
			releaseHandler,
		),
	)

	return serve(
		ctx,
		server,
		cfg.ShutdownTimeout,
	)
}

func serve(
	ctx context.Context,
	server *http.Server,
	shutdownTimeout time.Duration,
) error {
	serverErr := make(chan error, 1)

	go func() {
		slog.Info(
			"starting shipguard-api",
			"addr",
			server.Addr,
		)

		serverErr <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown requested")

	case err := <-serverErr:
		if err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf(
				"serve HTTP: %w",
				err,
			)
		}

		return nil
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		shutdownTimeout,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf(
			"shutdown HTTP server: %w",
			err,
		)
	}

	if err := <-serverErr; err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf(
			"serve HTTP after shutdown: %w",
			err,
		)
	}

	slog.Info("shipguard-api stopped")

	return nil
}
