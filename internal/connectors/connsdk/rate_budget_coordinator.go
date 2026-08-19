package connsdk

import (
	"context"
	"time"
)

// ReservationKey is the only identity sent to the run-local budget
// coordinator. Both fields are opaque values; no raw account subject,
// credential, request, URL, or header crosses this seam.
type ReservationKey struct {
	PolicyFingerprint string
	Scope             string
}

// ReservationPolicy is one selected consumptive policy in an atomic admission
// batch. Its fingerprint binds an opaque identity to its declared budget shape.
type ReservationPolicy struct {
	Key     ReservationKey
	Budgets []RateLimitBudget
}

// ReservationBatch is all policy state that one logical request must reserve.
// A coordinator grants every policy plus one lease or charges none of them.
type ReservationBatch struct {
	Policies []ReservationPolicy
}

// RateBudgetLease is an opaque, one-shot reservation handle. It must never be
// emitted in user-facing output or delivery evidence.
type RateBudgetLease string

// AdmissionDecision is a non-network coordinator result. A non-grant never
// consumes budget; Wait and NotBefore are advisory owner-clock values.
type AdmissionDecision struct {
	Granted   bool
	Lease     RateBudgetLease
	NotBefore time.Time
	Wait      time.Duration
}

// CompletionObservation reuses the requester's already secret-free response
// vocabulary, preventing a second provider-observation contract from drifting.
type CompletionObservation = RateLimitObservation

// BudgetCoordinator reserves a complete policy batch and later finishes its
// opaque lease. Finish is idempotent so an uncertain transport result remains
// conservatively charged rather than authorizing a replay.
type BudgetCoordinator interface {
	Decide(context.Context, ReservationBatch) (AdmissionDecision, error)
	Finish(context.Context, RateBudgetLease, CompletionObservation) error
}
