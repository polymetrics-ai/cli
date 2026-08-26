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

func TestGitHubLabelDeleteValidatesRequiredInputBeforeCredentialResolution(t *testing.T) {
	root := t.TempDir()
	var initOut, initErr bytes.Buffer
	if code := Run([]string{"init", "--root", root, "--json"}, &initOut, &initErr); code != 0 {
		t.Fatalf("init code=%d stdout=%s stderr=%s", code, initOut.String(), initErr.String())
	}

	var helpOut, helpErr bytes.Buffer
	if code := Run([]string{"github", "label", "delete", "--help", "--root", root, "--json"}, &helpOut, &helpErr); code != 0 {
		t.Fatalf("label delete help code=%d stdout=%s stderr=%s", code, helpOut.String(), helpErr.String())
	}
	if !strings.Contains(helpOut.String(), "--name") {
		t.Fatalf("label delete help omitted required input: %s", helpOut.String())
	}

	var missingOut, missingErr bytes.Buffer
	code := Run([]string{"github", "label", "delete", "--root", root, "--json"}, &missingOut, &missingErr)
	if code == 0 || !strings.Contains(missingOut.String()+missingErr.String(), "missing required flag --name") {
		t.Fatalf("bare label delete code=%d stdout=%s stderr=%s", code, missingOut.String(), missingErr.String())
	}
	if strings.Contains(missingOut.String()+missingErr.String(), "missing --credential") {
		t.Fatalf("bare label delete reached credential resolution before input validation: stdout=%s stderr=%s", missingOut.String(), missingErr.String())
	}

	var validOut, validErr bytes.Buffer
	code = Run([]string{"github", "label", "delete", "--name", "bug", "--root", root, "--json"}, &validOut, &validErr)
	if code == 0 || !strings.Contains(validOut.String()+validErr.String(), "missing --credential") {
		t.Fatalf("complete label delete code=%d stdout=%s stderr=%s", code, validOut.String(), validErr.String())
	}
}

func TestConnectorCommandInputDefectsFailBeforeWithApp(t *testing.T) {
	minimum := connectors.ExactNumber("2")
	bundle := engine.Bundle{
		Name:    "input-ordering",
		Streams: []engine.StreamSpec{{Name: "widgets", Method: "GET", Path: "/widgets"}},
		Surface: &engine.APISurface{Endpoints: []engine.SurfaceEndpoint{{Method: "GET", Path: "/widgets"}}},
		CLISurface: &engine.CLISurface{
			Tagline: "Input ordering fixture", Usage: "pm input-ordering widgets list",
			Commands: []engine.CLICommand{{
				Path: "widgets list", Summary: "List widgets", Intent: "etl", Availability: "implemented", Stream: "widgets",
				APISurface: []engine.CLISurfaceEndpointRef{{Method: "GET", Path: "/widgets"}},
				Flags: []engine.CLIFlag{
					{Name: "state", Type: "enum", Values: []string{"open", "closed"}, MapsTo: "query.state", Required: true},
					{Name: "batch", Type: "integer", MapsTo: "query.batch", Minimum: &minimum},
					{Name: "secret-input", Type: "string", MapsTo: "query.secret_input", EnvOnly: true},
				},
			}},
		},
	}
	registry := connectors.NewEmptyRegistry()
	registry.Register(engine.New(bundle, nil))
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "missing required", args: []string{"widgets", "list"}, want: "missing required flag --state"},
		{name: "unknown", args: []string{"widgets", "list", "--state", "open", "--other", "value"}, want: "unknown flag --other"},
		{name: "enum", args: []string{"widgets", "list", "--state", "pending"}, want: "want one of"},
		{name: "minimum", args: []string{"widgets", "list", "--state", "open", "--batch", "1"}, want: "at least 2"},
		{name: "env only direct carrier", args: []string{"widgets", "list", "--state", "open", "--secret-input", "must-not-enter-argv"}, want: "--secret-input must be supplied through --from-env"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "input-ordering", testCase.args, &stdout, &stderr, true, registry)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("input defect error=%v stdout=%s stderr=%s; want %q before app open", err, stdout.String(), stderr.String(), testCase.want)
			}
		})
	}

	var helpOut, helpErr bytes.Buffer
	if err := runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "input-ordering", []string{"widgets", "list", "--help"}, &helpOut, &helpErr, false, registry); err != nil {
		t.Fatalf("help should return before request validation: %v", err)
	}
}
