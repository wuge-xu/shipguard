package release

import (
	"errors"
	"testing"
	"time"
)

func TestReleaseTransition(t *testing.T) {
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

	current := Release{
		ID:          "release-001",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      StatusPendingApproval,
		Version:     1,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}

	transitionAt := createdAt.Add(
		5 * time.Minute,
	)

	next, err := current.Transition(
		StatusApproved,
		transitionAt,
	)
	if err != nil {
		t.Fatalf(
			"Transition() error = %v, want nil",
			err,
		)
	}

	if next.Status != StatusApproved {
		t.Fatalf(
			"Status = %q, want %q",
			next.Status,
			StatusApproved,
		)
	}

	if next.Version != 2 {
		t.Fatalf(
			"Version = %d, want 2",
			next.Version,
		)
	}

	if !next.UpdatedAt.Equal(transitionAt) {
		t.Fatalf(
			"UpdatedAt = %v, want %v",
			next.UpdatedAt,
			transitionAt,
		)
	}

	if current.Status != StatusPendingApproval {
		t.Fatalf(
			"original Status = %q, want unchanged",
			current.Status,
		)
	}

	if current.Version != 1 {
		t.Fatalf(
			"original Version = %d, want 1",
			current.Version,
		)
	}

	if !current.UpdatedAt.Equal(createdAt) {
		t.Fatalf(
			"original UpdatedAt = %v, want %v",
			current.UpdatedAt,
			createdAt,
		)
	}
}

func TestReleaseTransitionRejectsInvalidTransition(t *testing.T) {
	now := time.Date(
		2026,
		time.September,
		3,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	current := Release{
		ID:          "release-001",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      StatusPendingApproval,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := current.Transition(
		StatusSucceeded,
		now.Add(time.Minute),
	)

	if !errors.Is(
		err,
		ErrInvalidTransition,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidTransition",
			err,
		)
	}
}

func TestReleaseTransitionRejectsTimeGoingBackwards(t *testing.T) {
	now := time.Date(
		2026,
		time.September,
		3,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	current := Release{
		ID:          "release-001",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      StatusPendingApproval,
		Version:     1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err := current.Transition(
		StatusApproved,
		now.Add(-time.Second),
	)

	if !errors.Is(
		err,
		ErrInvalidRelease,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidRelease",
			err,
		)
	}
}

func TestReleaseTransitionRejectsTerminalState(t *testing.T) {
	now := time.Date(
		2026,
		time.September,
		3,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	current := Release{
		ID:          "release-001",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      StatusRolledBack,
		Version:     8,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now,
	}

	_, err := current.Transition(
		StatusDeploying,
		now.Add(time.Minute),
	)

	if !errors.Is(
		err,
		ErrInvalidTransition,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidTransition",
			err,
		)
	}
}

func TestReleaseTransitionNormalizesTimeToUTC(t *testing.T) {
	location := time.FixedZone(
		"CST",
		8*60*60,
	)

	createdAt := time.Date(
		2026,
		time.September,
		3,
		18,
		0,
		0,
		0,
		location,
	)

	current := Release{
		ID:          "release-001",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      StatusPendingApproval,
		Version:     1,
		CreatedAt:   createdAt.UTC(),
		UpdatedAt:   createdAt.UTC(),
	}

	next, err := current.Transition(
		StatusApproved,
		createdAt.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf(
			"Transition() error = %v, want nil",
			err,
		)
	}

	if next.UpdatedAt.Location() != time.UTC {
		t.Fatalf(
			"UpdatedAt location = %v, want UTC",
			next.UpdatedAt.Location(),
		)
	}
}
