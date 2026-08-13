package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
}

func Load() (Config, error) {
	timeoutText := envOrDefault("SHIPGUARD_SHUTDOWN_TIMEOUT", "10s")

	shutdownTimeout, err := time.ParseDuration(timeoutText)
	if err != nil {
		return Config{}, fmt.Errorf("parse SHIPGUARD_SHUTDOWN_TIMEOUT: %w", err)
	}

	return Config{
		HTTPAddr:        envOrDefault("SHIPGUARD_HTTP_ADDR", ":8080"),
		ShutdownTimeout: shutdownTimeout,
	}, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
