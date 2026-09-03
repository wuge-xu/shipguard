package repository

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/wuge-xu/shipguard/internal/database"
	"github.com/wuge-xu/shipguard/internal/release"
)

func TestReleaseRepositoryOptimisticLockIntegration(t *testing.T) {
	databaseURL := os.Getenv("SHIPGUARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SHIPGUARD_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(
		ctx,
		"TRUNCATE releases",
	)
	if err != nil {
		t.Fatalf("truncate releases: %v", err)
	}

	repo := NewReleaseRepository(pool)

	createdAt := time.Date(
		2026,
		time.September,
		3,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	item, err := release.NewRelease(
		release.CreateParams{
			ID:          "release-lock-001",
			Service:     "demo-service",
			Environment: "production",
			SourceSHA:   "abc123",
			ImageDigest: "sha256:deadbeef",
		},
		createdAt,
	)
	if err != nil {
		t.Fatalf("NewRelease() error = %v", err)
	}

	if err := repo.Create(ctx, item); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	workerA, err := repo.GetByID(
		ctx,
		item.ID,
	)
	if err != nil {
		t.Fatalf("worker A GetByID() error = %v", err)
	}

	workerB, err := repo.GetByID(
		ctx,
		item.ID,
	)
	if err != nil {
		t.Fatalf("worker B GetByID() error = %v", err)
	}

	if workerA.Version != 1 || workerB.Version != 1 {
		t.Fatalf(
			"worker versions = (%d, %d), want (1, 1)",
			workerA.Version,
			workerB.Version,
		)
	}

	nextA, err := workerA.Transition(
		release.StatusApproved,
		createdAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("worker A Transition() error = %v", err)
	}

	nextB, err := workerB.Transition(
		release.StatusApproved,
		createdAt.Add(2*time.Minute),
	)
	if err != nil {
		t.Fatalf("worker B Transition() error = %v", err)
	}

	if err := repo.UpdateTransition(
		ctx,
		workerA,
		nextA,
	); err != nil {
		t.Fatalf(
			"worker A UpdateTransition() error = %v",
			err,
		)
	}

	err = repo.UpdateTransition(
		ctx,
		workerB,
		nextB,
	)
	if !errors.Is(
		err,
		ErrReleaseVersionConflict,
	) {
		t.Fatalf(
			"worker B error = %v, want ErrReleaseVersionConflict",
			err,
		)
	}

	stored, err := repo.GetByID(
		ctx,
		item.ID,
	)
	if err != nil {
		t.Fatalf("final GetByID() error = %v", err)
	}

	if stored.Status != release.StatusApproved {
		t.Fatalf(
			"final Status = %q, want %q",
			stored.Status,
			release.StatusApproved,
		)
	}

	if stored.Version != 2 {
		t.Fatalf(
			"final Version = %d, want 2",
			stored.Version,
		)
	}

	if !stored.UpdatedAt.Equal(nextA.UpdatedAt) {
		t.Fatalf(
			"final UpdatedAt = %v, want worker A value %v",
			stored.UpdatedAt,
			nextA.UpdatedAt,
		)
	}
}

func TestReleaseRepositoryUpdateTransitionNotFound(t *testing.T) {
	databaseURL := os.Getenv("SHIPGUARD_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("SHIPGUARD_TEST_DATABASE_URL is not set")
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer pool.Close()

	_, err = pool.Exec(
		ctx,
		"TRUNCATE releases",
	)
	if err != nil {
		t.Fatalf("truncate releases: %v", err)
	}

	repo := NewReleaseRepository(pool)

	createdAt := time.Date(
		2026,
		time.September,
		3,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	current, err := release.NewRelease(
		release.CreateParams{
			ID:          "release-missing-001",
			Service:     "demo-service",
			Environment: "production",
			SourceSHA:   "abc123",
			ImageDigest: "sha256:deadbeef",
		},
		createdAt,
	)
	if err != nil {
		t.Fatalf("NewRelease() error = %v", err)
	}

	next, err := current.Transition(
		release.StatusApproved,
		createdAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("Transition() error = %v", err)
	}

	err = repo.UpdateTransition(
		ctx,
		current,
		next,
	)

	if !errors.Is(
		err,
		ErrReleaseNotFound,
	) {
		t.Fatalf(
			"error = %v, want ErrReleaseNotFound",
			err,
		)
	}
}
