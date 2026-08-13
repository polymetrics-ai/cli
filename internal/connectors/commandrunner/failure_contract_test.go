package commandrunner

import (
	"errors"
	"testing"

	"polymetrics.ai/internal/failures"
)

func TestBlockedCommandErrorCarriesEverySharedDispatchClassification(t *testing.T) {
	for _, kind := range []failures.DispatchKind{
		failures.DispatchKindDirectStub,
		failures.DispatchKindHelperDelegatedRefusal,
		failures.DispatchKindWrappedTypedUnsupported,
		failures.DispatchKindDeclaredButUnroutableCommand,
		failures.DispatchKindUnresolvedDynamicTarget,
	} {
		t.Run(string(kind), func(t *testing.T) {
			want, err := failures.New(failures.Input{
				Domain:       failures.DomainSystem,
				Code:         "dispatch_not_executable",
				Message:      "the declared command does not reach an executor",
				DispatchKind: kind,
			}, errors.New("internal dispatch analysis detail"))
			if err != nil {
				t.Fatalf("failures.New() error = %v", err)
			}
			blocked := &BlockedCommandError{Command: "issues list", Failure: want}
			var got *failures.Classification
			if !errors.As(blocked, &got) {
				t.Fatalf("BlockedCommandError does not unwrap shared classification: %v", blocked)
			}
			if got.DispatchKind() != kind {
				t.Fatalf("DispatchKind() = %q, want %q", got.DispatchKind(), kind)
			}
			if got.Retryable() {
				t.Fatal("dispatch refusal is retryable, want false")
			}
		})
	}
}

func TestBlockedCommandErrorWithoutClassificationDoesNotUnwrap(t *testing.T) {
	blocked := &BlockedCommandError{Command: "issues list"}
	var classification *failures.Classification
	if errors.As(blocked, &classification) {
		t.Fatalf("errors.As() classification = %v, want no classification", classification)
	}
}
