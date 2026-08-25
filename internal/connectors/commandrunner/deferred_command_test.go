package commandrunner

import (
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/failures"
)

func TestPreflightDeferredCommandReturnsNamedFoundationBeforeExecutor(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{{
		Path: "widgets delete", Intent: "reverse_etl", Availability: "deferred",
		Foundation: &connectors.CommandFoundation{ID: "delete_plan_foundation", Reason: "delete plan compiler is not available"},
	}}}}
	_, _, err := resolvePreflightCommand(connector, []string{"widgets", "delete"})
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("error = %v, want BlockedCommandError", err)
	}
	if blocked.Availability != "deferred" || blocked.Failure == nil {
		t.Fatalf("blocked = %+v, want deferred typed refusal", blocked)
	}
	if blocked.Failure.Code() != "missing_foundation" || blocked.Failure.Domain() != failures.DomainSystem {
		t.Fatalf("failure = code=%q domain=%q, want missing_foundation/system", blocked.Failure.Code(), blocked.Failure.Domain())
	}
}
