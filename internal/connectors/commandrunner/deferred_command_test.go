package commandrunner

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/failures"
)

type deferredPreflightConnector struct {
	*fakeConnector
	err error
}

func (c *deferredPreflightConnector) PreflightDeferredCommand(connectors.CommandSurfaceCommand) error {
	return c.err
}

func deferredCommandFixture() connectors.CommandSurfaceCommand {
	return connectors.CommandSurfaceCommand{
		Path: "widgets delete", Intent: "reverse_etl", Availability: "deferred",
		APISurface: []connectors.CommandSurfaceEndpointRef{{Method: "DELETE", Path: "/widgets/{id}"}},
		Foundation: &connectors.CommandFoundation{
			ID: "delete_plan_foundation", Reason: "delete plan compiler is not available",
			Component: connectors.FoundationComponentRuntimeExecutor, Evidence: "runtime_executor_absent",
			Target: connectors.CommandFoundationTarget{Method: "DELETE", Path: "/widgets/{id}"},
		},
	}
}

func TestPreflightDeferredCommandReturnsNamedFoundationAfterExactTargetValidation(t *testing.T) {
	connector := &deferredPreflightConnector{fakeConnector: &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{deferredCommandFixture()}}}}
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

func TestPreflightDeferredCommandFailsClosedWithoutExactTargetValidation(t *testing.T) {
	connector := &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{deferredCommandFixture()}}}
	_, _, err := resolvePreflightCommand(connector, []string{"widgets", "delete"})
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("error = %v, want BlockedCommandError", err)
	}
	if blocked.Failure != nil {
		t.Fatalf("failure = %+v, want no missing_foundation classification without exact target validation", blocked.Failure)
	}
}

func TestPreflightDeferredCommandFailsBeforeMissingFoundationOnInvalidTarget(t *testing.T) {
	connector := &deferredPreflightConnector{
		fakeConnector: &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{deferredCommandFixture()}}},
		err:           fmt.Errorf("target is excluded"),
	}
	_, _, err := resolvePreflightCommand(connector, []string{"widgets", "delete"})
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) {
		t.Fatalf("error = %v, want BlockedCommandError", err)
	}
	if blocked.Failure != nil {
		t.Fatalf("failure = %+v, want no missing_foundation classification for invalid target", blocked.Failure)
	}
}

func TestPreflightDeferredCommandKeepsTypedClassificationForOversizedFoundationText(t *testing.T) {
	command := deferredCommandFixture()
	command.Foundation.ID = strings.Repeat("foundation", 80)
	command.Foundation.Reason = strings.Repeat("runtime foundation is unavailable ", 100)
	connector := &deferredPreflightConnector{fakeConnector: &fakeConnector{surface: &connectors.CommandSurface{Commands: []connectors.CommandSurfaceCommand{command}}}}

	_, _, err := resolvePreflightCommand(connector, []string{"widgets", "delete"})
	var blocked *BlockedCommandError
	if !errors.As(err, &blocked) || blocked.Failure == nil {
		t.Fatalf("oversized foundation error = %v, want typed blocked failure", err)
	}
	if blocked.Failure.Code() != "missing_foundation" || blocked.Failure.Domain() != failures.DomainSystem {
		t.Fatalf("oversized foundation classification = %s/%s, want system/missing_foundation", blocked.Failure.Domain(), blocked.Failure.Code())
	}
}
