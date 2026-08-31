package release

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidRelease = errors.New("invalid release")

type CreateParams struct {
	ID          string
	Service     string
	Environment string
	SourceSHA   string
	ImageDigest string
}

func NewRelease(
	params CreateParams,
	now time.Time,
) (Release, error) {
	release := Release{
		ID:          strings.TrimSpace(params.ID),
		Service:     strings.TrimSpace(params.Service),
		Environment: strings.TrimSpace(params.Environment),
		SourceSHA:   strings.TrimSpace(params.SourceSHA),
		ImageDigest: strings.TrimSpace(params.ImageDigest),
		Status:      StatusPendingApproval,
		Version:     1,
		CreatedAt:   now.UTC(),
		UpdatedAt:   now.UTC(),
	}

	if err := release.Validate(); err != nil {
		return Release{}, err
	}

	return release, nil
}

func (r Release) Validate() error {
	requiredFields := []struct {
		name  string
		value string
	}{
		{
			name:  "id",
			value: r.ID,
		},
		{
			name:  "service",
			value: r.Service,
		},
		{
			name:  "environment",
			value: r.Environment,
		},
		{
			name:  "source_sha",
			value: r.SourceSHA,
		},
		{
			name:  "image_digest",
			value: r.ImageDigest,
		},
	}

	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			return fmt.Errorf(
				"%w: %s is required",
				ErrInvalidRelease,
				field.name,
			)
		}
	}

	if !r.Status.Valid() {
		return fmt.Errorf(
			"%w: status %q is invalid",
			ErrInvalidRelease,
			r.Status,
		)
	}

	if r.Version < 1 {
		return fmt.Errorf(
			"%w: version must be at least 1",
			ErrInvalidRelease,
		)
	}

	if r.CreatedAt.IsZero() {
		return fmt.Errorf(
			"%w: created_at is required",
			ErrInvalidRelease,
		)
	}

	if r.UpdatedAt.IsZero() {
		return fmt.Errorf(
			"%w: updated_at is required",
			ErrInvalidRelease,
		)
	}

	if r.UpdatedAt.Before(r.CreatedAt) {
		return fmt.Errorf(
			"%w: updated_at cannot be before created_at",
			ErrInvalidRelease,
		)
	}

	return nil
}
