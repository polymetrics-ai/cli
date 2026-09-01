package app

import (
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

func TestPlanConnectorCommandPreflightsDeferredCommandBeforeCredentialResolution(t *testing.T) {
	bundle := engine.Bundle{
		Name: "deferred-plan-fixture",
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
	if !errors.As(err, &blocked) {
		t.Fatalf("PlanConnectorCommand error = %v, want typed deferred preflight refusal before missing credential", err)
	}
	if blocked.Failure != nil {
		t.Fatalf("PlanConnectorCommand legacy classified failure = %v, want execution-binding error only", blocked.Failure)
	}
	if !strings.Contains(blocked.Error(), "deferred command foundation target requires one execution binding") {
		t.Fatalf("PlanConnectorCommand error = %v, want missing execution-binding refusal", blocked)
	}
}

func TestPlanConnectorCommandValidatesRequiredInputBeforeVaultResolution(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	if err := InitProject(root); err != nil {
		t.Fatalf("InitProject: %v", err)
	}
	instance, err := Open(root)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if _, err := instance.AddCredential(ctx, AddCredentialRequest{
		Name:      "github-local",
		Connector: "github",
		Config: map[string]string{
			"owner": "acme", "repo": "widgets", "public_access": "true", "base_url": "https://provider.example.test",
		},
	}); err != nil {
		t.Fatalf("AddCredential: %v", err)
	}

	// A nil vault turns an accidental credential read into an observable panic.
	// Valid request validation must therefore return before this boundary.
	instance.vault = nil
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("PlanConnectorCommand reached vault resolution for invalid input: %v", recovered)
		}
	}()

	_, _, err = instance.PlanConnectorCommand(ctx, PlanConnectorCommandRequest{
		Connector: "github", Credential: "github-local", Path: []string{"label", "delete"},
	})
	if err == nil || !strings.Contains(err.Error(), "missing required flag --name") {
		t.Fatalf("PlanConnectorCommand invalid input error = %v, want required-input validation before credential resolution", err)
	}
}
