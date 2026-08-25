package cli

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestShippedDeclarationTargetReachesPublicMissingFoundationEnvelopeWithoutAPISurface(t *testing.T) {
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load shipped GitHub bundle: %v", err)
	}
	if bundle.Surface != nil {
		t.Fatal("shipped GitHub bundle unexpectedly retained api_surface")
	}
	targets := engine.DeclarationAdmissionTargets(bundle)
	if len(targets) != 1 {
		t.Fatalf("shipped declaration targets = %+v, want one compact admitted source row", targets)
	}
	commands := map[string]engine.CLICommand{}
	for _, candidate := range bundle.CLISurface.Commands {
		switch candidate.Path {
		case "label delete", "discussion list", "discussion create":
			commands[candidate.Path] = candidate
		}
	}
	if len(commands) != 3 {
		t.Fatalf("shipped GitHub stale-deferral controls = %v, want delete and GraphQL read/write", commands)
	}
	target := targets[0]
	controls := []struct {
		name       string
		command    engine.CLICommand
		target     engine.CommandFoundationTarget
		invocation []string
	}{
		{
			name: "implemented delete", command: commands["label delete"], invocation: []string{"label", "delete"},
			target: engine.CommandFoundationTarget{
				SourceID: target.SourceID, ProviderOperationID: target.ProviderOperationID,
				Binding:         engine.CommandBindingIdentity{Kind: target.Binding.Kind, ID: target.Binding.ID},
				DestructiveKind: target.DestructiveKind, Method: target.Method, Path: target.Path,
			},
		},
		{
			name: "GraphQL ETL read", command: commands["discussion list"], invocation: []string{"discussion", "list"},
			target: engine.CommandFoundationTarget{
				SourceID: "github.graphql.list-discussions", ProviderOperationID: "ListDiscussions",
				Binding:         engine.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: "discussions"},
				DestructiveKind: "none", Method: "GRAPHQL", Path: "ListDiscussions",
			},
		},
		{
			name: "GraphQL direct write", command: commands["discussion create"], invocation: []string{"discussion", "create"},
			target: engine.CommandFoundationTarget{
				SourceID: "github.graphql.create-discussion", ProviderOperationID: "GitHubMutationCreateDiscussion",
				Binding:         engine.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: "github.graphql.mutation.create-discussion"},
				DestructiveKind: "none", Method: "POST", Path: "/graphql",
			},
		},
	}
	foundations := []struct{ component, evidence string }{
		{connectors.FoundationComponentTypedWriteAction, "write_action_absent"},
		{connectors.FoundationComponentTypedRecordSchema, "record_schema_absent"},
		{connectors.FoundationComponentTypedRequestBody, "request_body_schema_absent"},
		{connectors.FoundationComponentTypedResponseDescriptor, "response_descriptor_absent"},
		{connectors.FoundationComponentBinaryTransferBinding, "binary_transfer_binding_absent"},
		{connectors.FoundationComponentSourceImporter, "source_importer_absent"},
		{connectors.FoundationComponentRuntimeExecutor, "runtime_executor_absent"},
		{"idempotency_contract", "idempotency_contract_absent"},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			for _, foundation := range foundations {
				t.Run(foundation.component, func(t *testing.T) {
					deferred := control.command
					deferred.Availability = "deferred"
					deferred.Foundation = &engine.CommandFoundation{
						ID: foundation.component + "_foundation", Reason: "the command is incorrectly relabelled deferred",
						Component: foundation.component, Evidence: foundation.evidence, Target: control.target,
					}
					bundleCopy := bundle
					bundleCopy.CLISurface = &engine.CLISurface{Commands: []engine.CLICommand{deferred}}
					registry := connectors.NewEmptyRegistry()
					registry.Register(engine.New(bundleCopy, nil))

					var stdout, stderr bytes.Buffer
					runErr := runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "github", control.invocation, &stdout, &stderr, true, registry)
					var blocked *commandrunner.BlockedCommandError
					if !errors.As(runErr, &blocked) {
						t.Fatalf("deferred shipped command error = %v, want blocked preflight", runErr)
					}
					if blocked.Failure != nil && blocked.Failure.Code() == "missing_foundation" {
						t.Fatalf("runnable GitHub command accepted stale %s deferral: %+v", foundation.component, blocked)
					}
					if !strings.Contains(blocked.Reason, "passes implemented runtime preflight") {
						t.Fatalf("stale deferral reason = %q, want real implemented-preflight proof", blocked.Reason)
					}
					if stdout.Len() != 0 || stderr.Len() != 0 {
						t.Fatalf("preflight wrote output before the public error boundary: stdout=%q stderr=%q", stdout.String(), stderr.String())
					}
				})
			}
		})
	}
}
