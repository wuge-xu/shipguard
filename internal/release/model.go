package release

import (
	"errors"
	"fmt"
	"time"
)

type Status string

const (
	StatusPendingApproval Status = "pending_approval"
	StatusApproved        Status = "approved"
	StatusDeploying       Status = "deploying"
	StatusVerifying       Status = "verifying"
	StatusSucceeded       Status = "succeeded"
	StatusFailed          Status = "failed"
	StatusRollingBack     Status = "rolling_back"
	StatusRolledBack      Status = "rolled_back"
	StatusCanceled        Status = "canceled"
)

var (
	ErrInvalidStatus     = errors.New("invalid release status")
	ErrInvalidTransition = errors.New("invalid release status transition")
)

type Release struct {
	ID          string
	Service     string
	Environment string
	SourceSHA   string
	ImageDigest string
	GitOpsSHA   string
	Status      Status
	Version     int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

var allowedTransitions = map[Status]map[Status]struct{}{
	StatusPendingApproval: {
		StatusApproved: {},
		StatusCanceled: {},
	},
	StatusApproved: {
		StatusDeploying: {},
		StatusCanceled:  {},
	},
	StatusDeploying: {
		StatusVerifying:   {},
		StatusFailed:      {},
		StatusRollingBack: {},
	},
	StatusVerifying: {
		StatusSucceeded:   {},
		StatusFailed:      {},
		StatusRollingBack: {},
	},
	StatusSucceeded: {
		StatusRollingBack: {},
	},
	StatusFailed: {
		StatusRollingBack: {},
	},
	StatusRollingBack: {
		StatusRolledBack: {},
		StatusFailed:     {},
	},
}

func (s Status) Valid() bool {
	switch s {
	case StatusPendingApproval,
		StatusApproved,
		StatusDeploying,
		StatusVerifying,
		StatusSucceeded,
		StatusFailed,
		StatusRollingBack,
		StatusRolledBack,
		StatusCanceled:
		return true
	default:
		return false
	}
}

func CanTransition(from, to Status) bool {
	if !from.Valid() || !to.Valid() {
		return false
	}

	nextStatuses, ok := allowedTransitions[from]
	if !ok {
		return false
	}

	_, ok = nextStatuses[to]
	return ok
}

func ValidateTransition(from, to Status) error {
	if !from.Valid() {
		return fmt.Errorf(
			"%w: %q",
			ErrInvalidStatus,
			from,
		)
	}

	if !to.Valid() {
		return fmt.Errorf(
			"%w: %q",
			ErrInvalidStatus,
			to,
		)
	}

	if !CanTransition(from, to) {
		return fmt.Errorf(
			"%w: %s -> %s",
			ErrInvalidTransition,
			from,
			to,
		)
	}

	return nil
}
