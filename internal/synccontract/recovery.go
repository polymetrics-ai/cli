package synccontract

import (
	"errors"
	"fmt"
)

// ErrRebootstrapRequired is the sentinel parent for every recovery outcome
// that must stop resumption. Callers can use errors.Is while still inspecting
// the typed outcome with errors.As.
var ErrRebootstrapRequired = errors.New("sync rebootstrap required")

// RecoveryOutcome is a closed reason why an existing checkpoint cannot be
// resumed. None of these outcomes permit a caller to clear state and scan
// again implicitly.
type RecoveryOutcome string

const (
	RecoveryOutcomeInvalidCheckpoint          RecoveryOutcome = "invalid_checkpoint"
	RecoveryOutcomeRetentionGap               RecoveryOutcome = "retention_gap"
	RecoveryOutcomeInvalidatedSlot            RecoveryOutcome = "invalidated_slot"
	RecoveryOutcomeExpiredToken               RecoveryOutcome = "expired_token"
	RecoveryOutcomeSourceGenerationChanged    RecoveryOutcome = "source_generation_changed"
	RecoveryOutcomeSourceIdentityIncompatible RecoveryOutcome = "source_identity_incompatible"
)

func (o RecoveryOutcome) valid() bool {
	switch o {
	case RecoveryOutcomeInvalidCheckpoint,
		RecoveryOutcomeRetentionGap,
		RecoveryOutcomeInvalidatedSlot,
		RecoveryOutcomeExpiredToken,
		RecoveryOutcomeSourceGenerationChanged,
		RecoveryOutcomeSourceIdentityIncompatible:
		return true
	default:
		return false
	}
}

// RebootstrapRequiredError is the non-retriable-by-resume error returned when
// a source has invalidated its durable state. Detail is diagnostic context; it
// must never contain an opaque provider token.
type RebootstrapRequiredError struct {
	Outcome RecoveryOutcome
	Detail  string
}

func (e *RebootstrapRequiredError) Error() string {
	if e == nil {
		return ErrRebootstrapRequired.Error()
	}
	if e.Detail == "" {
		return fmt.Sprintf("%s: %s", ErrRebootstrapRequired, e.Outcome)
	}
	return fmt.Sprintf("%s: %s: %s", ErrRebootstrapRequired, e.Outcome, e.Detail)
}

func (e *RebootstrapRequiredError) Unwrap() error { return ErrRebootstrapRequired }

// RequireRebootstrap constructs a typed recovery outcome. Unknown outcomes
// are treated as an invalid checkpoint rather than creating an unhandled
// fallback that might resume unsafely.
func RequireRebootstrap(outcome RecoveryOutcome, detail string) error {
	if !outcome.valid() {
		outcome = RecoveryOutcomeInvalidCheckpoint
	}
	return &RebootstrapRequiredError{Outcome: outcome, Detail: detail}
}
