package main

import (
	"os"
	"strings"
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
	for path, cmd := range commands {
		op := ops[cmd.Operation]
		if op.Kind != "graphql_query" || op.GraphQL == nil {
			t.Fatalf("direct read command %q operation = %+v, want graphql_query with document", path, op)
		}
		doc := strings.TrimSpace(op.GraphQL.Document)
		if !strings.HasPrefix(doc, "query "+op.GraphQL.OperationName) {
			t.Fatalf("direct read command %q document starts %q, want named fixed query %q", path, firstLine(doc), op.GraphQL.OperationName)
		}
		if strings.Contains(doc, "__typename") {
			t.Fatalf("direct read command %q still uses placeholder __typename document", path)
		}
		if strings.Contains(doc, "$") && len(cmd.Flags) == 0 {
			t.Fatalf("direct read command %q requires GraphQL variables but declares no CLI flags", path)
		}
	}
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}
