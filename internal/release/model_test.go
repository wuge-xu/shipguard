package release

import (
	"errors"
	"testing"
)

func TestStatusValid(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{
			name:   "pending approval",
			status: StatusPendingApproval,
			want:   true,
		},
		{
			name:   "approved",
			status: StatusApproved,
			want:   true,
		},
		{
			name:   "deploying",
			status: StatusDeploying,
			want:   true,
		},
		{
			name:   "verifying",
			status: StatusVerifying,
			want:   true,
		},
		{
			name:   "succeeded",
			status: StatusSucceeded,
			want:   true,
		},
		{
			name:   "failed",
			status: StatusFailed,
			want:   true,
		},
		{
			name:   "rolling back",
			status: StatusRollingBack,
			want:   true,
		},
		{
			name:   "rolled back",
			status: StatusRolledBack,
			want:   true,
		},
		{
			name:   "canceled",
			status: StatusCanceled,
			want:   true,
		},
		{
			name:   "unknown status",
			status: Status("unknown"),
			want:   false,
		},
		{
			name:   "empty status",
			status: Status(""),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.status.Valid(); got != tt.want {
				t.Fatalf(
					"Valid() = %v, want %v",
					got,
					tt.want,
				)
			}
		})
	}
}

func TestCanTransition(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{
			name: "pending approval to approved",
			from: StatusPendingApproval,
			to:   StatusApproved,
			want: true,
		},
		{
			name: "pending approval to canceled",
			from: StatusPendingApproval,
			to:   StatusCanceled,
			want: true,
		},
		{
			name: "approved to deploying",
			from: StatusApproved,
			to:   StatusDeploying,
			want: true,
		},
		{
			name: "deploying to verifying",
			from: StatusDeploying,
			to:   StatusVerifying,
			want: true,
		},
		{
			name: "deploying to failed",
			from: StatusDeploying,
			to:   StatusFailed,
			want: true,
		},
		{
			name: "deploying to rolling back",
			from: StatusDeploying,
			to:   StatusRollingBack,
			want: true,
		},
		{
			name: "verifying to succeeded",
			from: StatusVerifying,
			to:   StatusSucceeded,
			want: true,
		},
		{
			name: "verifying to failed",
			from: StatusVerifying,
			to:   StatusFailed,
			want: true,
		},
		{
			name: "succeeded to rolling back",
			from: StatusSucceeded,
			to:   StatusRollingBack,
			want: true,
		},
		{
			name: "failed to rolling back",
			from: StatusFailed,
			to:   StatusRollingBack,
			want: true,
		},
		{
			name: "rolling back to rolled back",
			from: StatusRollingBack,
			to:   StatusRolledBack,
			want: true,
		},
		{
			name: "pending approval cannot skip to deploying",
			from: StatusPendingApproval,
			to:   StatusDeploying,
			want: false,
		},
		{
			name: "approved cannot skip to succeeded",
			from: StatusApproved,
			to:   StatusSucceeded,
			want: false,
		},
		{
			name: "succeeded cannot return to verifying",
			from: StatusSucceeded,
			to:   StatusVerifying,
			want: false,
		},
		{
			name: "rolled back is terminal",
			from: StatusRolledBack,
			to:   StatusDeploying,
			want: false,
		},
		{
			name: "canceled is terminal",
			from: StatusCanceled,
			to:   StatusApproved,
			want: false,
		},
		{
			name: "same status is not a transition",
			from: StatusDeploying,
			to:   StatusDeploying,
			want: false,
		},
		{
			name: "invalid source",
			from: Status("unknown"),
			to:   StatusApproved,
			want: false,
		},
		{
			name: "invalid destination",
			from: StatusApproved,
			to:   Status("unknown"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanTransition(tt.from, tt.to); got != tt.want {
				t.Fatalf(
					"CanTransition(%q, %q) = %v, want %v",
					tt.from,
					tt.to,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestValidateTransition(t *testing.T) {
	t.Run("valid transition", func(t *testing.T) {
		err := ValidateTransition(
			StatusPendingApproval,
			StatusApproved,
		)

		if err != nil {
			t.Fatalf(
				"ValidateTransition() error = %v, want nil",
				err,
			)
		}
	})

	t.Run("invalid source status", func(t *testing.T) {
		err := ValidateTransition(
			Status("unknown"),
			StatusApproved,
		)

		if !errors.Is(err, ErrInvalidStatus) {
			t.Fatalf(
				"error = %v, want ErrInvalidStatus",
				err,
			)
		}
	})

	t.Run("invalid destination status", func(t *testing.T) {
		err := ValidateTransition(
			StatusApproved,
			Status("unknown"),
		)

		if !errors.Is(err, ErrInvalidStatus) {
			t.Fatalf(
				"error = %v, want ErrInvalidStatus",
				err,
			)
		}
	})

	t.Run("invalid transition", func(t *testing.T) {
		err := ValidateTransition(
			StatusPendingApproval,
			StatusSucceeded,
		)

		if !errors.Is(err, ErrInvalidTransition) {
			t.Fatalf(
				"error = %v, want ErrInvalidTransition",
				err,
			)
		}
	})
}
