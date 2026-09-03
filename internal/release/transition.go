package release

import (
	"fmt"
	"time"
)

func (r Release) Transition(
	to Status,
	now time.Time,
) (Release, error) {
	if err := r.Validate(); err != nil {
		return Release{}, fmt.Errorf(
			"validate release before transition: %w",
			err,
		)
	}

	if err := ValidateTransition(
		r.Status,
		to,
	); err != nil {
		return Release{}, err
	}

	updatedAt := now.UTC()

	if updatedAt.Before(r.UpdatedAt) {
		return Release{}, fmt.Errorf(
			"%w: transition time cannot be before current updated_at",
			ErrInvalidRelease,
		)
	}

	next := r

	next.Status = to
	next.Version = r.Version + 1
	next.UpdatedAt = updatedAt

	if err := next.Validate(); err != nil {
		return Release{}, fmt.Errorf(
			"validate release after transition: %w",
			err,
		)
	}

	return next, nil
}
