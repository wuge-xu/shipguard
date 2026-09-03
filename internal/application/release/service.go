package releaseapp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/wuge-xu/shipguard/internal/release"
)

type Repository interface {
	Create(
		ctx context.Context,
		item release.Release,
	) error

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

type CreateReleaseInput struct {
	Service     string
	Environment string
	SourceSHA   string
	ImageDigest string
}

type Service struct {
	repository Repository
	now        func() time.Time
	newID      func() string
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
		newID:      newReleaseID,
	}
}

func (s *Service) CreateRelease(
	ctx context.Context,
	input CreateReleaseInput,
) (release.Release, error) {
	item, err := release.NewRelease(
		release.CreateParams{
			ID:          s.newID(),
			Service:     input.Service,
			Environment: input.Environment,
			SourceSHA:   input.SourceSHA,
			ImageDigest: input.ImageDigest,
		},
		s.now(),
	)
	if err != nil {
		return release.Release{}, fmt.Errorf(
			"create release domain model: %w",
			err,
		)
	}

	if err := s.repository.Create(
		ctx,
		item,
	); err != nil {
		return release.Release{}, fmt.Errorf(
			"persist release %q: %w",
			item.ID,
			err,
		)
	}

	return item, nil
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

func newReleaseID() string {
	var buffer [16]byte

	if _, err := rand.Read(buffer[:]); err == nil {
		return "rel-" + hex.EncodeToString(buffer[:])
	}

	return fmt.Sprintf(
		"rel-%d",
		time.Now().UnixNano(),
	)
}
