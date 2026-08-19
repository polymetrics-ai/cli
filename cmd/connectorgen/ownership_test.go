package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/boundary"
)

func TestOwnershipCommand_JSONCleanExitZero(t *testing.T) {
	root := newBoundaryCommandFixture(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})

	var stdout, stderr bytes.Buffer
	exit := run([]string{
		"ownership",
		root,
		"--scope-file", "connector-scope.json",
		"--changed-path", "internal/connectors/defs/github/metadata.json",
		"--json",
	}, &stdout, &stderr)
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	var report boundary.OwnershipReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode ownership report: %v\n%s", err, stdout.String())
	}
	if report.Kind != boundary.OwnershipReportKind || report.TargetConnector != "github" || len(report.Findings) != 0 {
		t.Fatalf("unexpected clean ownership report: %+v", report)
	}
}

func TestOwnershipCommand_PolicyViolationExitOne(t *testing.T) {
	root := newBoundaryCommandFixture(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github"]}`,
	})

	var stdout, stderr bytes.Buffer
	exit := run([]string{
		"ownership",
		root,
		"--scope-file", "connector-scope.json",
		"--changed-path", "internal/connectors/defs/github/metadata.json",
		"--changed-path", "internal/connectors/engine/read.go",
		"--json",
	}, &stdout, &stderr)
	if exit != 1 {
		t.Fatalf("exit = %d, want 1\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	var report boundary.OwnershipReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode ownership report: %v\n%s", err, stdout.String())
	}
	if report.Outcome != boundary.OutcomePolicyViolations || len(report.Findings) == 0 {
		t.Fatalf("unexpected violation report: %+v", report)
	}
}

func TestOwnershipCommand_InvalidScopeExitTwo(t *testing.T) {
	root := newBoundaryCommandFixture(t, map[string]string{
		"connector-scope.json": `{"api_version":"polymetrics.ai/v1","kind":"ConnectorImplementationScope","connectors":["github","gong"]}`,
	})

	var stdout, stderr bytes.Buffer
	exit := run([]string{
		"ownership",
		root,
		"--scope-file", "connector-scope.json",
		"--changed-path", "internal/connectors/defs/github/metadata.json",
		"--json",
	}, &stdout, &stderr)
	if exit != 2 {
		t.Fatalf("exit = %d, want 2\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("invalid scope wrote stdout: %s", stdout.String())
	}
	if !strings.Contains(stderr.String(), "exactly one connector") {
		t.Fatalf("stderr missing invalid scope reason: %s", stderr.String())
	}
}

func TestOwnershipCommand_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exit := run([]string{"ownership", ".", "--help"}, &stdout, &stderr)
	if exit != 0 {
		t.Fatalf("exit = %d, want 0\nstderr=%s\nstdout=%s", exit, stderr.String(), stdout.String())
	}
	for _, want := range []string{"connectorgen ownership", "--json", "--scope-file", "ConnectorImplementationScope"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("help missing %q:\n%s", want, stdout.String())
		}
	}
}
