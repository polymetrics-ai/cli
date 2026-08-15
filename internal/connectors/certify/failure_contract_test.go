package certify

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/failures"
)

func TestCapabilityResultSerializesSharedUntestableReasonWithoutCause(t *testing.T) {
	cause := errors.New("internal dispatch graph analysis found unresolved interface target")
	reason, err := failures.New(failures.Input{
		Domain:       failures.DomainSystem,
		Code:         "dispatch_not_executable",
		Message:      "the declared command cannot be certified as executable",
		DispatchKind: failures.DispatchKindUnresolvedDynamicTarget,
		References: []failures.Reference{
			{Kind: failures.ReferenceKindCommand, Value: "issues/list"},
		},
	}, cause)
	if err != nil {
		t.Fatalf("failures.New() error = %v", err)
	}

	raw, err := json.Marshal(CapabilityResult{Result: "untestable", UntestableReason: reason})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	text := string(raw)
	if !strings.Contains(text, `"untestable_reason":{"domain":"system","code":"dispatch_not_executable"`) {
		t.Fatalf("serialized untestable reason = %s", text)
	}
	if strings.Contains(text, cause.Error()) || strings.Contains(text, "cause") {
		t.Fatalf("serialized untestable reason leaked cause: %s", text)
	}
}
