package cli

import (
	"bytes"
	"context"
	"testing"

	"polymetrics.ai/internal/connectors"
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
	var command engine.CLICommand
	for _, candidate := range bundle.CLISurface.Commands {
		if candidate.Path == "label delete" {
			command = candidate
			break
		}
	}
	if command.Path == "" {
		t.Fatal("shipped GitHub label delete command is missing")
	}
	target := targets[0]
	command.Availability = "deferred"
	command.Foundation = &engine.CommandFoundation{
		ID: "idempotency_contract_foundation", Reason: "the admitted delete action has no idempotency contract",
		Component: "idempotency_contract", Evidence: "idempotency_contract_absent",
		Target: engine.CommandFoundationTarget{
			SourceID: target.SourceID, ProviderOperationID: target.ProviderOperationID,
			Binding:         engine.CommandBindingIdentity{Kind: target.Binding.Kind, ID: target.Binding.ID},
			DestructiveKind: target.DestructiveKind, Method: target.Method, Path: target.Path,
		},
	}
	bundle.CLISurface = &engine.CLISurface{Commands: []engine.CLICommand{command}}
	registry := connectors.NewEmptyRegistry()
	registry.Register(engine.New(bundle, nil))

	var stdout, stderr bytes.Buffer
	err = runMaybeConnectorCommandWithRegistry(context.Background(), t.TempDir(), "github", []string{"label", "delete"}, &stdout, &stderr, true, registry)
	if err == nil {
		t.Fatal("deferred shipped command returned success")
	}
	public := publicErrorEnvelope(err)
	if public["code"] != "missing_foundation" || public["category"] != string(categoryInternal) {
		t.Fatalf("public error = %#v, want internal/missing_foundation", public)
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("preflight wrote output before the public error boundary: stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}
