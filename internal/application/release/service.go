package releaseapp

import (
	"context"
	"fmt"
	"time"

	"github.com/wuge-xu/shipguard/internal/release"
)

type Repository interface {
	GetByID(
		ctx context.Context,
		id string,
	) (release.Release, error)

	UpdateTransition(
		ctx context.Context,
		current release.Release,
		next release.Release,
	) error
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(
	repository Repository,
	now func() time.Time,
) *Service {
	if now == nil {
		now = time.Now
	}

	return &Service{
		repository: repository,
		now:        now,
	}
}

func (s *Service) ApproveRelease(
	ctx context.Context,
	id string,
) (release.Release, error) {
	return s.transition(
		ctx,
		id,
		release.StatusApproved,
	)
}

func (s *Service) transition(
	ctx context.Context,
	id string,
	to release.Status,
) (release.Release, error) {
	current, err := s.repository.GetByID(
		ctx,
		id,
	)
	if err != nil {
		return release.Release{}, fmt.Errorf(
			"get release %q: %w",
			id,
			err,
		)
	}

	next, err := current.Transition(
		to,
		s.now(),
	)
	if err != nil {
		return release.Release{}, fmt.Errorf(
			"transition release %q: %w",
			id,
			err,
		)
	}

	if err := s.repository.UpdateTransition(
		ctx,
		current,
		next,
	); err != nil {
		return release.Release{}, fmt.Errorf(
			"persist release %q transition: %w",
			id,
			err,
		)
	}

	return next, nil
}
