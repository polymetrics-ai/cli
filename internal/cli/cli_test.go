package cli_test

import (
	"bytes"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
)

func TestHelpIncludesManPageStyleSections(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"help", "credentials"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(help) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "SECURITY", "EXIT STATUS"} {
		if !strings.Contains(out, want) {
			t.Fatalf("help output missing %q:\n%s", want, out)
		}
	}
}

func TestRootHelpAliasesShowManual(t *testing.T) {
	tests := [][]string{
		{"--help"},
		{"-h"},
		{"help"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", args, code, stderr.String())
			}
			out := stdout.String()
			for _, want := range []string{"NAME", "SYNOPSIS", "COMMANDS", "AGENT CONTRACT", "EXIT STATUS"} {
				if !strings.Contains(out, want) {
					t.Fatalf("root help missing %q:\n%s", want, out)
				}
			}
		})
	}
}

func TestRootHelpJSONIsAgentReadable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"--json", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(--json --help) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "CommandManual"`, `"command": "pm"`, `"manual":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("root json manual missing %q:\n%s", want, out)
		}
	}
}

func TestVersionCommandReportsBuildMetadata(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(version) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"pm dev", "commit: none", "built: unknown"} {
		if !strings.Contains(out, want) {
			t.Fatalf("version output missing %q:\n%s", want, out)
		}
	}
}

func TestVersionCommandJSONIsAgentReadable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"version", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(version --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "Version"`, `"version": "dev"`, `"commit": "none"`, `"date": "unknown"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("version json missing %q:\n%s", want, out)
		}
	}
}

func TestBareCommandShowsManualInsteadOfUsageError(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"connectors"}, want: "pm connectors - inspect connector definitions, streams, and write actions"},
		{args: []string{"etl"}, want: "SYNC MODES"},
		{args: []string{"credentials"}, want: "pm credentials - manage encrypted connector credentials"},
		{args: []string{"connections"}, want: "pm connections - configure source-to-destination sync connections"},
		{args: []string{"reverse"}, want: "pm reverse - plan, preview, approve, and execute reverse ETL"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", tt.args, code, stderr.String())
			}
			out := stdout.String()
			if strings.Contains(out, "invalid usage") || strings.Contains(stderr.String(), "invalid usage") {
				t.Fatalf("bare command returned usage error; stdout=%q stderr=%q", out, stderr.String())
			}
			for _, section := range []string{"NAME", "SYNOPSIS", "DESCRIPTION", "EXIT STATUS"} {
				if !strings.Contains(out, section) {
					t.Fatalf("manual missing section %q:\n%s", section, out)
				}
			}
			if !strings.Contains(out, tt.want) {
				t.Fatalf("manual missing %q:\n%s", tt.want, out)
			}
		})
	}
}

func TestBareCommandJSONShowsManualForAgents(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors --json) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "CommandManual"`, `"command": "connectors"`, `"manual":`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json manual missing %q:\n%s", want, out)
		}
	}
}

func TestConnectorsManualDocumentsConnectorArchitectureAndGithubExamples(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"declarative JSON bundles",
		"write=true/false",
		"REVERSE ETL WRITE ACTIONS",
		"218 writable connectors out of 547",
		"GITHUB AUTHENTICATION",
		"public",
		"token",
		"github_app",
		"GITHUB ETL STREAMS",
		"issues",
		"pull_requests",
		"create_pull_request",
		"merge_pull_request",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("connectors manual missing %q:\n%s", want, out)
		}
	}
}

func TestConnectorInspectHumanShowsManualNotRawJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "inspect", "github"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors inspect github) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if strings.HasPrefix(strings.TrimSpace(out), "{") {
		t.Fatalf("human connector inspect returned raw JSON:\n%s", out)
	}
	for _, want := range []string{"NAME", "SYNOPSIS", "AUTHENTICATION", "ETL STREAMS", "REVERSE ETL ACTIONS", "AGENT WORKFLOW"} {
		if !strings.Contains(out, want) {
			t.Fatalf("human connector manual missing %q:\n%s", want, out)
		}
	}
}

func TestDocsGenerateAndValidateConnectorDocs(t *testing.T) {
	dir := t.TempDir()
	cliDir := dir + "/cli"
	connectorsDir := dir + "/connectors"
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"docs", "generate", "--dir", cliDir, "--connectors-dir", connectorsDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs generate code = %d stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = cli.Run([]string{"docs", "validate", "--connectors-dir", connectorsDir}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("docs validate code = %d stderr = %s stdout = %s", code, stderr.String(), stdout.String())
	}
}

func TestConnectorListJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"connectors", "list", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(connectors list) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorList"`, `"name": "sample"`, `"name": "warehouse"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
}

func TestPerfCompareJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"perf", "compare", "--iterations", "1", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(perf compare) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "PerformanceComparison"`, `"mode": "dependency-free"`, `"records": 3`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
}

func TestPerfSyncModesJSON(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"perf", "sync-modes", "--records", "20", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(perf sync-modes) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "SyncModeBenchmark"`, `"full_refresh_append"`, `"incremental_append_deduped"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
}

func TestETLHelpListsAllSyncModes(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"help", "etl"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(help etl) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("etl help missing %q:\n%s", want, out)
		}
	}
}

func TestETLRejectsPlannedCatalogConnectorCommands(t *testing.T) {
	root := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"init", "--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("init code = %d stderr = %s", code, stderr.String())
	}

	tests := [][]string{
		{"etl", "check", "--connector", "source-strava", "--root", root, "--json"},
		{"etl", "catalog", "--connector", "source-strava", "--root", root, "--json"},
		{"etl", "read", "--connector", "source-strava", "--stream", "records", "--limit", "1", "--root", root, "--json"},
	}
	for _, args := range tests {
		t.Run(strings.Join(args[:3], " "), func(t *testing.T) {
			stdout.Reset()
			stderr.Reset()
			code := cli.Run(args, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("Run(%v) code = 0, want planned connector rejection; stdout = %s", args, stdout.String())
			}
			if !strings.Contains(stderr.String()+stdout.String(), `connector "source-strava" not found`) {
				t.Fatalf("Run(%v) did not explain planned connector is unavailable; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
			}
		})
	}
}

func TestRuntimeDoctorJSONDoesNotLeakPostgresPassword(t *testing.T) {
	t.Setenv("POLYMETRICS_POSTGRES_URL", "postgres://user:secret@127.0.0.1:1/db?sslmode=disable")
	t.Setenv("POLYMETRICS_DRAGONFLY_ADDR", "127.0.0.1:1")
	t.Setenv("POLYMETRICS_TEMPORAL_ADDR", "127.0.0.1:1")

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"runtime", "doctor", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(runtime doctor) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	if strings.Contains(out, "secret") {
		t.Fatalf("runtime doctor leaked password:\n%s", out)
	}
	if !strings.Contains(out, `"kind": "RuntimeDoctor"`) {
		t.Fatalf("missing RuntimeDoctor kind:\n%s", out)
	}
}
