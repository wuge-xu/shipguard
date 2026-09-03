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

func TestReleaseRepositoryIntegration(t *testing.T) {
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

	now := time.Date(
		2026,
		time.August,
		31,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	item, err := release.NewRelease(
		release.CreateParams{
			ID:          "release-repository-001",
			Service:     "demo-service",
			Environment: "production",
			SourceSHA:   "abc123",
			ImageDigest: "sha256:deadbeef",
		},
		now,
	)
	if err != nil {
		t.Fatalf("NewRelease() error = %v", err)
	}

	t.Run("create and get by id", func(t *testing.T) {
		err := repo.Create(ctx, item)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		got, err := repo.GetByID(
			ctx,
			item.ID,
		)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}

		assertReleaseEqual(
			t,
			got,
			item,
		)
	})

	t.Run("not found", func(t *testing.T) {
		_, err := repo.GetByID(
			ctx,
			"release-does-not-exist",
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
	})

	t.Run("duplicate id is rejected", func(t *testing.T) {
		err := repo.Create(
			ctx,
			item,
		)

		if err == nil {
			t.Fatal(
				"Create() error = nil, want duplicate ID error",
			)
		}
	})
}

func TestReleaseRepositoryRejectsInvalidRelease(t *testing.T) {
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

	repo := NewReleaseRepository(pool)

	item := release.Release{
		ID: "invalid-release",
	}

	err = repo.Create(
		ctx,
		item,
	)
	if err == nil {
		t.Fatal(
			"Create() error = nil, want validation error",
		)
	}

	if !errors.Is(
		err,
		release.ErrInvalidRelease,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidRelease",
			err,
		)
	}
}

func assertReleaseEqual(
	t *testing.T,
	got release.Release,
	want release.Release,
) {
	t.Helper()

	if got.ID != want.ID {
		t.Fatalf(
			"ID = %q, want %q",
			got.ID,
			want.ID,
		)
	}

	if got.Service != want.Service {
		t.Fatalf(
			"Service = %q, want %q",
			got.Service,
			want.Service,
		)
	}

	if got.Environment != want.Environment {
		t.Fatalf(
			"Environment = %q, want %q",
			got.Environment,
			want.Environment,
		)
	}

	if got.SourceSHA != want.SourceSHA {
		t.Fatalf(
			"SourceSHA = %q, want %q",
			got.SourceSHA,
			want.SourceSHA,
		)
	}

	if got.ImageDigest != want.ImageDigest {
		t.Fatalf(
			"ImageDigest = %q, want %q",
			got.ImageDigest,
			want.ImageDigest,
		)
	}

	if got.GitOpsSHA != want.GitOpsSHA {
		t.Fatalf(
			"GitOpsSHA = %q, want %q",
			got.GitOpsSHA,
			want.GitOpsSHA,
		)
	}

	if got.Status != want.Status {
		t.Fatalf(
			"Status = %q, want %q",
			got.Status,
			want.Status,
		)
	}

	if got.Version != want.Version {
		t.Fatalf(
			"Version = %d, want %d",
			got.Version,
			want.Version,
		)
	}

	if !got.CreatedAt.Equal(want.CreatedAt) {
		t.Fatalf(
			"CreatedAt = %v, want %v",
			got.CreatedAt,
			want.CreatedAt,
		)
	}

	if !got.UpdatedAt.Equal(want.UpdatedAt) {
		t.Fatalf(
			"UpdatedAt = %v, want %v",
			got.UpdatedAt,
			want.UpdatedAt,
		)
	}
}
