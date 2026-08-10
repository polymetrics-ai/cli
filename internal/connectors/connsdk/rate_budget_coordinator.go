package connsdk

import (
	"context"
	"time"
)

// RateBudgetBackendMode selects where declared rate-limit budgets are
// coordinated. The zero value retains the dependency-free process-local
// registry. RequireShared never falls back to that registry.
type RateBudgetBackendMode string

const (
	RateBudgetBackendProcessLocal  RateBudgetBackendMode = "process_local"
	RateBudgetBackendRequireShared RateBudgetBackendMode = "require_shared"
)

// ReservationKey is the only identity sent to a budget coordinator. The
// policy fingerprint and Scope are both opaque values: no connector name,
// policy ID, credential, binding preimage, raw provider account value, URL,
// header, variable, or request body crosses this seam.
type ReservationKey struct {
	PolicyFingerprint string
	Scope             string
}

// ReservationPolicy is one already-selected consumptive policy in an atomic
// admission batch. Fingerprint binds a key to its declared budget shape so a
// later caller cannot silently reinterpret a live shared budget.
type ReservationPolicy struct {
	Key     ReservationKey
	Budgets []RateLimitBudget
}

// ReservationBatch contains every policy that must be charged by one logical
// request. A coordinator either reserves every policy and one opaque lease or
// reserves none of them.
type ReservationBatch struct {
	Policies []ReservationPolicy
}

// RateBudgetLease is an opaque, one-shot reservation handle. Its value has no
// operator meaning and must never be included in user-visible evidence.
type RateBudgetLease string

// AdmissionDecision is a non-network decision from a BudgetCoordinator. A
// non-grant with NotBefore set has charged no consumptive policy. Wait is an
// owner-clock duration provided for local scheduling; callers must still honor
// their own context deadline before waiting.
type AdmissionDecision struct {
	Granted   bool
	Lease     RateBudgetLease
	NotBefore time.Time
	Wait      time.Duration
}

// CompletionObservation intentionally has exactly the secret-free response
// facts already accepted by RateLimitObserver. Keeping it an alias prevents a
// second, drifting provider-observation vocabulary at the coordinator seam.
type CompletionObservation = RateLimitObservation

// BudgetCoordinator decides a complete policy batch and later finishes its
// opaque lease. Finish is idempotent: callers may safely try it after a
// transport error, while a coordinator crash leaves a charged reservation
// conservative rather than allowing an uncertain request to be replayed.
type BudgetCoordinator interface {
	Decide(context.Context, ReservationBatch) (AdmissionDecision, error)
	Finish(context.Context, RateBudgetLease, CompletionObservation) error
}

// RateBudgetRefusalReason classifies a local admission refusal without
// retaining request or credential material.
type RateBudgetRefusalReason string

const (
	RateBudgetRefusalNotBefore                    RateBudgetRefusalReason = "not_before"
	RateBudgetRefusalDeadlineTooShort             RateBudgetRefusalReason = "deadline_too_short"
	RateBudgetRefusalSharedCoordinatorUnavailable RateBudgetRefusalReason = "shared_coordinator_unavailable"
)

// RateBudgetRefusalError is returned before transport when a coordinator must
// fail closed or the caller cannot wait until a safe retry time.
type RateBudgetRefusalError struct {
	Reason    RateBudgetRefusalReason
	NotBefore time.Time
}

func (e *RateBudgetRefusalError) Error() string {
	if e == nil {
		return "rate-budget admission refused"
	}
	switch e.Reason {
	case RateBudgetRefusalDeadlineTooShort:
		return "rate-budget admission refused: caller deadline is too short"
	case RateBudgetRefusalSharedCoordinatorUnavailable:
		return "rate-budget admission refused: shared coordinator is unavailable"
	default:
		return "rate-budget admission refused"
	}
}
