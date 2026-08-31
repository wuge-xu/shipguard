package database

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestOpenRejectsInvalidDatabaseURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		time.Second,
	)
	defer cancel()

	_, err := Open(
		ctx,
		"://invalid-url",
	)

	if err == nil {
		t.Fatal(
			"Open() error = nil, want non-nil",
		)
	}

	if !strings.Contains(
		err.Error(),
		"parse PostgreSQL config",
	) {
		t.Fatalf(
			"error = %q, want parse PostgreSQL config",
			err,
		)
	}
}

func TestOpenIntegration(t *testing.T) {
	databaseURL := os.Getenv(
		"SHIPGUARD_TEST_DATABASE_URL",
	)

	if databaseURL == "" {
		t.Skip(
			"SHIPGUARD_TEST_DATABASE_URL is not set",
		)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	pool, err := Open(
		ctx,
		databaseURL,
	)
	if err != nil {
		t.Fatalf(
			"Open() error = %v, want nil",
			err,
		)
	}
	defer pool.Close()

	var database string
	var user string

	err = pool.QueryRow(
		ctx,
		`
		SELECT
			current_database(),
			current_user
		`,
	).Scan(
		&database,
		&user,
	)
	if err != nil {
		t.Fatalf(
			"query database identity: %v",
			err,
		)
	}

	if database != "shipguard" {
		t.Fatalf(
			"database = %q, want %q",
			database,
			"shipguard",
		)
	}

	if user != "shipguard" {
		t.Fatalf(
			"user = %q, want %q",
			user,
			"shipguard",
		)
	}
}
