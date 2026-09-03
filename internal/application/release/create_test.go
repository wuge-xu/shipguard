package releaseapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wuge-xu/shipguard/internal/release"
)

var errFakeCreate = errors.New("fake create error")

func (f *fakeRepository) Create(
	_ context.Context,
	_ release.Release,
) error {
	return nil
}

type fakeCreateRepository struct {
	created     release.Release
	createErr   error
	createCalls int
}

func (f *fakeCreateRepository) Create(
	_ context.Context,
	item release.Release,
) error {
	f.createCalls++
	f.created = item

	return f.createErr
}

func (f *fakeCreateRepository) GetByID(
	_ context.Context,
	_ string,
) (release.Release, error) {
	return release.Release{}, nil
}

func (f *fakeCreateRepository) UpdateTransition(
	_ context.Context,
	_ release.Release,
	_ release.Release,
) error {
	return nil
}

func TestCreateRelease(t *testing.T) {
	now := time.Date(
		2026,
		time.September,
		3,
		15,
		0,
		0,
		0,
		time.UTC,
	)

	repository := &fakeCreateRepository{}

	service := NewService(
		repository,
		func() time.Time {
			return now
		},
	)

	service.newID = func() string {
		return "rel-test-001"
	}

	got, err := service.CreateRelease(
		context.Background(),
		CreateReleaseInput{
			Service:     "demo-service",
			Environment: "production",
			SourceSHA:   "abc123",
			ImageDigest: "sha256:deadbeef",
		},
	)
	if err != nil {
		t.Fatalf(
			"CreateRelease() error = %v, want nil",
			err,
		)
	}

	if got.ID != "rel-test-001" {
		t.Fatalf(
			"ID = %q, want %q",
			got.ID,
			"rel-test-001",
		)
	}

	if got.Status != release.StatusPendingApproval {
		t.Fatalf(
			"Status = %q, want %q",
			got.Status,
			release.StatusPendingApproval,
		)
	}

	if got.Version != 1 {
		t.Fatalf(
			"Version = %d, want 1",
			got.Version,
		)
	}

	if !got.CreatedAt.Equal(now) {
		t.Fatalf(
			"CreatedAt = %v, want %v",
			got.CreatedAt,
			now,
		)
	}

	if repository.createCalls != 1 {
		t.Fatalf(
			"Create calls = %d, want 1",
			repository.createCalls,
		)
	}

	if repository.created != got {
		t.Fatalf(
			"persisted release = %#v, want %#v",
			repository.created,
			got,
		)
	}
}

func TestCreateReleaseRejectsInvalidInput(t *testing.T) {
	repository := &fakeCreateRepository{}

	service := NewService(
		repository,
		nil,
	)

	service.newID = func() string {
		return "rel-test-invalid"
	}

	_, err := service.CreateRelease(
		context.Background(),
		CreateReleaseInput{
			Service:     "",
			Environment: "production",
			SourceSHA:   "abc123",
			ImageDigest: "sha256:deadbeef",
		},
	)

	if !errors.Is(
		err,
		release.ErrInvalidRelease,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidRelease",
			err,
		)
	}

	if repository.createCalls != 0 {
		t.Fatalf(
			"Create calls = %d, want 0",
			repository.createCalls,
		)
	}
}

func TestCreateReleasePropagatesRepositoryError(t *testing.T) {
	repository := &fakeCreateRepository{
		createErr: errFakeCreate,
	}

	service := NewService(
		repository,
		nil,
	)

	service.newID = func() string {
		return "rel-test-error"
	}

	_, err := service.CreateRelease(
		context.Background(),
		CreateReleaseInput{
			Service:     "demo-service",
			Environment: "production",
			SourceSHA:   "abc123",
			ImageDigest: "sha256:deadbeef",
		},
	)

	if !errors.Is(
		err,
		errFakeCreate,
	) {
		t.Fatalf(
			"error = %v, want errFakeCreate",
			err,
		)
	}

	if repository.createCalls != 1 {
		t.Fatalf(
			"Create calls = %d, want 1",
			repository.createCalls,
		)
	}
}
