package main

import (
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestMondaySensitiveAdminPolicy(t *testing.T) {
	b, err := engine.Load(os.DirFS("../../internal/connectors/defs"), "monday")
	if err != nil {
		t.Fatalf("Load(monday): %v", err)
	}

	mutations := 0
	for _, op := range b.Operations {
		if op.Kind != "graphql_mutation" {
			continue
		}
		mutations++
		if !strings.Contains(op.Approval, "plan, preview, approval, execute") {
			t.Fatalf("mutation %q approval = %q, want reverse ETL approval flow", op.ID, op.Approval)
		}
		if op.Risk == "high" || op.Risk == "critical" || op.MutationClass == "admin" || op.MutationClass == "secret" || op.MutationClass == "delete" {
			if !strings.Contains(op.Approval, "typed confirmation") {
				t.Fatalf("sensitive/admin mutation %q approval = %q, want typed confirmation", op.ID, op.Approval)
			}
		}
		if op.SecretSensitive {
			if op.SensitivePolicy == nil {
				t.Fatalf("secret-sensitive mutation %q has no sensitive_policy", op.ID)
			}
			if op.SensitivePolicy.InputMode == "" || op.SensitivePolicy.InputMode == "inline" {
				t.Fatalf("secret-sensitive mutation %q input_mode = %q, want non-inline", op.ID, op.SensitivePolicy.InputMode)
			}
			if len(op.SensitivePolicy.RedactFields) == 0 || op.SensitivePolicy.ApprovalMode != "typed_confirmation" {
				t.Fatalf("secret-sensitive mutation %q policy = %+v, want redact fields + typed confirmation", op.ID, op.SensitivePolicy)
			}
		}
	}
	if mutations != 280 {
		t.Fatalf("mutation operations = %d, want 280", mutations)
	}

	for _, ep := range b.Surface.Endpoints {
		if strings.EqualFold(ep.Method, "POST") {
			if ep.Operation == nil || !ep.Operation.BlockedByDefault || ep.Operation.Status != "blocked" {
				t.Fatalf("mutation endpoint %s is not blocked operation metadata: %+v", ep.Path, ep)
			}
		}
	}

	for _, cmd := range b.CLISurface.Commands {
		if cmd.Intent == "raw_api" || cmd.Intent == "direct_write" {
			t.Fatalf("command %q exposes forbidden intent %q", cmd.Path, cmd.Intent)
		}
		if cmd.Intent == "reverse_etl" && cmd.Availability == "implemented" {
			t.Fatalf("reverse ETL command %q is executable; Monday writes must stay blocked", cmd.Path)
		}
	}

	for _, want := range []string{
		"Sensitive/admin mutation policy",
		"env_or_stdin",
		"typed confirmation",
		"No Monday mutation is executable",
	} {
		if !strings.Contains(b.Docs, want) {
			t.Fatalf("docs.md missing %q", want)
		}
	}
}
