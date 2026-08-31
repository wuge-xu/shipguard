package config

import (
	"strings"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv(
		"SHIPGUARD_HTTP_ADDR",
		"",
	)
	t.Setenv(
		"SHIPGUARD_SHUTDOWN_TIMEOUT",
		"",
	)
	t.Setenv(
		"SHIPGUARD_DATABASE_URL",
		"",
	)

	got, err := Load()
	if err != nil {
		t.Fatalf(
			"Load() error = %v, want nil",
			err,
		)
	}

	if got.HTTPAddr != ":8080" {
		t.Fatalf(
			"HTTPAddr = %q, want %q",
			got.HTTPAddr,
			":8080",
		)
	}

	if got.ShutdownTimeout != 10*time.Second {
		t.Fatalf(
			"ShutdownTimeout = %v, want %v",
			got.ShutdownTimeout,
			10*time.Second,
		)
	}

	wantDatabaseURL := "postgres://shipguard:shipguard-local@127.0.0.1:15432/shipguard?sslmode=disable"

	if got.DatabaseURL != wantDatabaseURL {
		t.Fatalf(
			"DatabaseURL = %q, want %q",
			got.DatabaseURL,
			wantDatabaseURL,
		)
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv(
		"SHIPGUARD_HTTP_ADDR",
		":9090",
	)
	t.Setenv(
		"SHIPGUARD_SHUTDOWN_TIMEOUT",
		"25s",
	)
	t.Setenv(
		"SHIPGUARD_DATABASE_URL",
		"postgres://custom:secret@db:5432/custom",
	)

	got, err := Load()
	if err != nil {
		t.Fatalf(
			"Load() error = %v, want nil",
			err,
		)
	}

	if got.HTTPAddr != ":9090" {
		t.Fatalf(
			"HTTPAddr = %q, want %q",
			got.HTTPAddr,
			":9090",
		)
	}

	if got.ShutdownTimeout != 25*time.Second {
		t.Fatalf(
			"ShutdownTimeout = %v, want %v",
			got.ShutdownTimeout,
			25*time.Second,
		)
	}

	if got.DatabaseURL != "postgres://custom:secret@db:5432/custom" {
		t.Fatalf(
			"DatabaseURL = %q, want override",
			got.DatabaseURL,
		)
	}
}

func TestLoadRejectsInvalidShutdownTimeout(t *testing.T) {
	t.Setenv(
		"SHIPGUARD_SHUTDOWN_TIMEOUT",
		"not-a-duration",
	)

	_, err := Load()
	if err == nil {
		t.Fatal(
			"Load() error = nil, want non-nil",
		)
	}

	if !strings.Contains(
		err.Error(),
		"SHIPGUARD_SHUTDOWN_TIMEOUT",
	) {
		t.Fatalf(
			"error = %q, want config key",
			err,
		)
	}
}
