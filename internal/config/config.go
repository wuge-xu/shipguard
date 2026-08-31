package config

import (
	"fmt"
	"os"
	"time"
)

const (
	defaultHTTPAddr        = ":8080"
	defaultShutdownTimeout = "10s"
	defaultDatabaseURL     = "postgres://shipguard:shipguard-local@127.0.0.1:15432/shipguard?sslmode=disable"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
	DatabaseURL     string
}

func Load() (Config, error) {
	shutdownTimeoutValue := envOrDefault(
		"SHIPGUARD_SHUTDOWN_TIMEOUT",
		defaultShutdownTimeout,
	)

	shutdownTimeout, err := time.ParseDuration(
		shutdownTimeoutValue,
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"parse SHIPGUARD_SHUTDOWN_TIMEOUT: %w",
			err,
		)
	}

	return Config{
		HTTPAddr: envOrDefault(
			"SHIPGUARD_HTTP_ADDR",
			defaultHTTPAddr,
		),
		ShutdownTimeout: shutdownTimeout,
		DatabaseURL: envOrDefault(
			"SHIPGUARD_DATABASE_URL",
			defaultDatabaseURL,
		),
	}, nil
}

func envOrDefault(
	key string,
	defaultValue string,
) string {
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return defaultValue
	}

	return value
}
