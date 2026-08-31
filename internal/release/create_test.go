package release

import (
	"errors"
	"testing"
	"time"
)

func TestNewRelease(t *testing.T) {
	now := time.Date(
		2026,
		time.August,
		31,
		18,
		30,
		0,
		0,
		time.FixedZone("CST", 8*60*60),
	)

	got, err := NewRelease(
		CreateParams{
			ID:          " release-123 ",
			Service:     " demo-service ",
			Environment: " production ",
			SourceSHA:   " abc123 ",
			ImageDigest: " sha256:deadbeef ",
		},
		now,
	)
	if err != nil {
		t.Fatalf(
			"NewRelease() error = %v, want nil",
			err,
		)
	}

	if got.ID != "release-123" {
		t.Fatalf(
			"ID = %q, want %q",
			got.ID,
			"release-123",
		)
	}

	if got.Service != "demo-service" {
		t.Fatalf(
			"Service = %q, want %q",
			got.Service,
			"demo-service",
		)
	}

	if got.Environment != "production" {
		t.Fatalf(
			"Environment = %q, want %q",
			got.Environment,
			"production",
		)
	}

	if got.SourceSHA != "abc123" {
		t.Fatalf(
			"SourceSHA = %q, want %q",
			got.SourceSHA,
			"abc123",
		)
	}

	if got.ImageDigest != "sha256:deadbeef" {
		t.Fatalf(
			"ImageDigest = %q, want %q",
			got.ImageDigest,
			"sha256:deadbeef",
		)
	}

	if got.GitOpsSHA != "" {
		t.Fatalf(
			"GitOpsSHA = %q, want empty",
			got.GitOpsSHA,
		)
	}

	if got.Status != StatusPendingApproval {
		t.Fatalf(
			"Status = %q, want %q",
			got.Status,
			StatusPendingApproval,
		)
	}

	if got.Version != 1 {
		t.Fatalf(
			"Version = %d, want 1",
			got.Version,
		)
	}

	wantTime := now.UTC()

	if !got.CreatedAt.Equal(wantTime) {
		t.Fatalf(
			"CreatedAt = %v, want %v",
			got.CreatedAt,
			wantTime,
		)
	}

	if !got.UpdatedAt.Equal(wantTime) {
		t.Fatalf(
			"UpdatedAt = %v, want %v",
			got.UpdatedAt,
			wantTime,
		)
	}

	if got.CreatedAt.Location() != time.UTC {
		t.Fatalf(
			"CreatedAt location = %v, want UTC",
			got.CreatedAt.Location(),
		)
	}

	if got.UpdatedAt.Location() != time.UTC {
		t.Fatalf(
			"UpdatedAt location = %v, want UTC",
			got.UpdatedAt.Location(),
		)
	}
}

func TestNewReleaseRejectsMissingRequiredFields(t *testing.T) {
	tests := []struct {
		name   string
		params CreateParams
	}{
		{
			name: "missing id",
			params: CreateParams{
				Service:     "demo-service",
				Environment: "production",
				SourceSHA:   "abc123",
				ImageDigest: "sha256:deadbeef",
			},
		},
		{
			name: "missing service",
			params: CreateParams{
				ID:          "release-123",
				Environment: "production",
				SourceSHA:   "abc123",
				ImageDigest: "sha256:deadbeef",
			},
		},
		{
			name: "missing environment",
			params: CreateParams{
				ID:          "release-123",
				Service:     "demo-service",
				SourceSHA:   "abc123",
				ImageDigest: "sha256:deadbeef",
			},
		},
		{
			name: "missing source sha",
			params: CreateParams{
				ID:          "release-123",
				Service:     "demo-service",
				Environment: "production",
				ImageDigest: "sha256:deadbeef",
			},
		},
		{
			name: "missing image digest",
			params: CreateParams{
				ID:          "release-123",
				Service:     "demo-service",
				Environment: "production",
				SourceSHA:   "abc123",
			},
		},
	}

	now := time.Now()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRelease(
				tt.params,
				now,
			)

			if !errors.Is(err, ErrInvalidRelease) {
				t.Fatalf(
					"error = %v, want ErrInvalidRelease",
					err,
				)
			}
		})
	}
}

func TestReleaseValidate(t *testing.T) {
	base := Release{
		ID:          "release-123",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      StatusPendingApproval,
		Version:     1,
		CreatedAt:   time.Date(2026, time.August, 31, 10, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, time.August, 31, 10, 0, 0, 0, time.UTC),
	}

	t.Run("valid release", func(t *testing.T) {
		if err := base.Validate(); err != nil {
			t.Fatalf(
				"Validate() error = %v, want nil",
				err,
			)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		release := base
		release.Status = Status("unknown")

		if err := release.Validate(); !errors.Is(err, ErrInvalidRelease) {
			t.Fatalf(
				"error = %v, want ErrInvalidRelease",
				err,
			)
		}
	})

	t.Run("invalid version", func(t *testing.T) {
		release := base
		release.Version = 0

		if err := release.Validate(); !errors.Is(err, ErrInvalidRelease) {
			t.Fatalf(
				"error = %v, want ErrInvalidRelease",
				err,
			)
		}
	})

	t.Run("missing created at", func(t *testing.T) {
		release := base
		release.CreatedAt = time.Time{}

		if err := release.Validate(); !errors.Is(err, ErrInvalidRelease) {
			t.Fatalf(
				"error = %v, want ErrInvalidRelease",
				err,
			)
		}
	})

	t.Run("missing updated at", func(t *testing.T) {
		release := base
		release.UpdatedAt = time.Time{}

		if err := release.Validate(); !errors.Is(err, ErrInvalidRelease) {
			t.Fatalf(
				"error = %v, want ErrInvalidRelease",
				err,
			)
		}
	})

	t.Run("updated at before created at", func(t *testing.T) {
		release := base
		release.UpdatedAt = release.CreatedAt.Add(-time.Second)

		if err := release.Validate(); !errors.Is(err, ErrInvalidRelease) {
			t.Fatalf(
				"error = %v, want ErrInvalidRelease",
				err,
			)
		}
	})
}
