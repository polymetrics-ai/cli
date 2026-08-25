package app

import (
	"context"
	"errors"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

func TestPlanConnectorCommandPreflightsDeferredCommandBeforeCredentialResolution(t *testing.T) {
	bundle := engine.Bundle{
		Name: "deferred-plan-fixture",
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{
			Method: "POST", Path: "/widgets",
			Operation: &engine.SurfaceOperation{Model: "sensitive_reverse_etl", Status: "blocked", BlockedByDefault: true, Reason: "runtime executor is pending"},
		}}},
		CLISurface: &engine.CLISurface{Commands: []engine.CLICommand{{
			Path: "widget create", Intent: "reverse_etl", Availability: "deferred",
			APISurface: []engine.CLISurfaceEndpointRef{{Method: "POST", Path: "/widgets"}},
			Foundation: &engine.CommandFoundation{
				ID: "runtime_executor_foundation", Reason: "runtime executor is pending",
				Component: "runtime_executor", Evidence: "runtime_executor_absent",
				Target: engine.CommandFoundationTarget{Method: "POST", Path: "/widgets"},
			},
		}}},
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(engine.New(bundle, nil))
	instance := &App{registry: registry}

	_, _, err := instance.PlanConnectorCommand(context.Background(), PlanConnectorCommandRequest{
		Connector: bundle.Name,
		Path:      []string{"widget", "create"},
	})
	var blocked *commandrunner.BlockedCommandError
	if !errors.As(err, &blocked) || blocked.Failure == nil {
		t.Fatalf("PlanConnectorCommand error = %v, want typed deferred preflight refusal before missing credential", err)
	}
	if blocked.Failure.Code() != "missing_foundation" {
		t.Fatalf("PlanConnectorCommand code = %q, want missing_foundation", blocked.Failure.Code())
	}
}
