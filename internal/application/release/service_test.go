package releaseapp

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wuge-xu/shipguard/internal/release"
)

var (
	errFakeGet    = errors.New("fake get error")
	errFakeUpdate = errors.New("fake update error")
)

type fakeRepository struct {
	getItem release.Release
	getErr  error

	updateCurrent release.Release
	updateNext    release.Release
	updateErr     error

	getCalls    int
	updateCalls int
}

func (f *fakeRepository) GetByID(
	_ context.Context,
	_ string,
) (release.Release, error) {
	f.getCalls++

	if f.getErr != nil {
		return release.Release{}, f.getErr
	}

	return f.getItem, nil
}

func (f *fakeRepository) UpdateTransition(
	_ context.Context,
	current release.Release,
	next release.Release,
) error {
	f.updateCalls++
	f.updateCurrent = current
	f.updateNext = next

	return f.updateErr
}

func TestApproveRelease(t *testing.T) {
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

	approveAt := createdAt.Add(
		5 * time.Minute,
	)

	current := newTestRelease(
		release.StatusPendingApproval,
		1,
		createdAt,
		createdAt,
	)

	repository := &fakeRepository{
		getItem: current,
	}

	service := NewService(
		repository,
		func() time.Time {
			return approveAt
		},
	)

	got, err := service.ApproveRelease(
		context.Background(),
		current.ID,
	)
	if err != nil {
		t.Fatalf(
			"ApproveRelease() error = %v, want nil",
			err,
		)
	}

	if got.Status != release.StatusApproved {
		t.Fatalf(
			"Status = %q, want %q",
			got.Status,
			release.StatusApproved,
		)
	}

	if got.Version != 2 {
		t.Fatalf(
			"Version = %d, want 2",
			got.Version,
		)
	}

	if !got.UpdatedAt.Equal(approveAt) {
		t.Fatalf(
			"UpdatedAt = %v, want %v",
			got.UpdatedAt,
			approveAt,
		)
	}

	if repository.getCalls != 1 {
		t.Fatalf(
			"GetByID calls = %d, want 1",
			repository.getCalls,
		)
	}

	if repository.updateCalls != 1 {
		t.Fatalf(
			"UpdateTransition calls = %d, want 1",
			repository.updateCalls,
		)
	}

	if repository.updateCurrent.Status != release.StatusPendingApproval {
		t.Fatalf(
			"persist current Status = %q, want %q",
			repository.updateCurrent.Status,
			release.StatusPendingApproval,
		)
	}

	if repository.updateCurrent.Version != 1 {
		t.Fatalf(
			"persist current Version = %d, want 1",
			repository.updateCurrent.Version,
		)
	}

	if repository.updateNext.Status != release.StatusApproved {
		t.Fatalf(
			"persist next Status = %q, want %q",
			repository.updateNext.Status,
			release.StatusApproved,
		)
	}

	if repository.updateNext.Version != 2 {
		t.Fatalf(
			"persist next Version = %d, want 2",
			repository.updateNext.Version,
		)
	}
}

func TestApproveReleasePropagatesGetError(t *testing.T) {
	repository := &fakeRepository{
		getErr: errFakeGet,
	}

	service := NewService(
		repository,
		nil,
	)

	_, err := service.ApproveRelease(
		context.Background(),
		"release-001",
	)

	if !errors.Is(
		err,
		errFakeGet,
	) {
		t.Fatalf(
			"error = %v, want errFakeGet",
			err,
		)
	}

	if repository.updateCalls != 0 {
		t.Fatalf(
			"UpdateTransition calls = %d, want 0",
			repository.updateCalls,
		)
	}
}

func TestApproveReleaseRejectsInvalidTransition(t *testing.T) {
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

	current := newTestRelease(
		release.StatusSucceeded,
		5,
		now.Add(-time.Hour),
		now,
	)

	repository := &fakeRepository{
		getItem: current,
	}

	service := NewService(
		repository,
		func() time.Time {
			return now.Add(time.Minute)
		},
	)

	_, err := service.ApproveRelease(
		context.Background(),
		current.ID,
	)

	if !errors.Is(
		err,
		release.ErrInvalidTransition,
	) {
		t.Fatalf(
			"error = %v, want ErrInvalidTransition",
			err,
		)
	}

	if repository.updateCalls != 0 {
		t.Fatalf(
			"UpdateTransition calls = %d, want 0",
			repository.updateCalls,
		)
	}
}

func TestApproveReleasePropagatesUpdateError(t *testing.T) {
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

	current := newTestRelease(
		release.StatusPendingApproval,
		1,
		now,
		now,
	)

	repository := &fakeRepository{
		getItem:   current,
		updateErr: errFakeUpdate,
	}

	service := NewService(
		repository,
		func() time.Time {
			return now.Add(time.Minute)
		},
	)

	_, err := service.ApproveRelease(
		context.Background(),
		current.ID,
	)

	if !errors.Is(
		err,
		errFakeUpdate,
	) {
		t.Fatalf(
			"error = %v, want errFakeUpdate",
			err,
		)
	}

	if repository.updateCalls != 1 {
		t.Fatalf(
			"UpdateTransition calls = %d, want 1",
			repository.updateCalls,
		)
	}
}

func newTestRelease(
	status release.Status,
	version int64,
	createdAt time.Time,
	updatedAt time.Time,
) release.Release {
	return release.Release{
		ID:          "release-001",
		Service:     "demo-service",
		Environment: "production",
		SourceSHA:   "abc123",
		ImageDigest: "sha256:deadbeef",
		Status:      status,
		Version:     version,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}
