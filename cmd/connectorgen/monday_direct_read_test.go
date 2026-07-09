package main

import (
	"os"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestMondayDirectReadMetadataAndSafety(t *testing.T) {
	findings, _ := validateBundleDir(os.DirFS("../../internal/connectors/defs"), "monday")
	if len(findings) != 0 {
		t.Fatalf("monday validate findings = %+v, want none", findings)
	}

	b, err := engine.Load(os.DirFS("../../internal/connectors/defs"), "monday")
	if err != nil {
		t.Fatalf("Load(monday): %v", err)
	}
	commands := map[string]engine.CLICommand{}
	for _, cmd := range b.CLISurface.Commands {
		if cmd.Intent == "raw_api" || cmd.Intent == "direct_write" {
			t.Fatalf("command %q exposes forbidden intent %q", cmd.Path, cmd.Intent)
		}
		if cmd.Availability == "implemented" && cmd.Intent == "direct_read" {
			commands[cmd.Path] = cmd
		}
	}

	want := map[string]string{
		"me view":      "monday.me.get_me",
		"account view": "monday.account.get_account",
	}
	for path, operation := range want {
		cmd, ok := commands[path]
		if !ok {
			t.Fatalf("implemented direct reads = %+v, missing %q", commands, path)
		}
		if cmd.Operation != operation || cmd.OutputPolicy != "graphql_json" || len(cmd.APISurface) != 1 {
			t.Fatalf("command %q = %+v, want operation %s graphql_json one api_surface ref", path, cmd, operation)
		}
	}
	if len(commands) != 82 {
		t.Fatalf("implemented direct reads = %d, want 82 non-stream query operations", len(commands))
	}

	covered := map[string]bool{}
	for _, ep := range b.Surface.Endpoints {
		if ep.CoveredBy == nil {
			continue
		}
		for _, direct := range append(ep.CoveredBy.DirectReads, ep.CoveredBy.DirectRead) {
			if direct != "" {
				covered[direct] = true
			}
		}
	}
	for path := range commands {
		if !covered[path] {
			t.Fatalf("api surface direct read coverage missing %q", path)
		}
	}
	if len(covered) != 82 {
		t.Fatalf("api surface direct read coverage = %d, want 82", len(covered))
	}

	ops := map[string]engine.OperationSpec{}
	for _, op := range b.Operations {
		ops[op.ID] = op
	}
	if op := ops["monday.me.get_me"]; op.Kind != "graphql_query" || op.GraphQL == nil || op.GraphQL.Document == "query MondayMeGetMe { __typename }" {
		t.Fatalf("monday.me.get_me operation = %+v, want fixed real query document", op)
	}
	if op := ops["monday.account.get_account"]; op.Kind != "graphql_query" || op.GraphQL == nil || op.GraphQL.Document == "query MondayAccountGetAccount { __typename }" {
		t.Fatalf("monday.account.get_account operation = %+v, want fixed real query document", op)
	}
}
