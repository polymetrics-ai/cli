package cli_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/bundleregistry"
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

func TestDynamicConnectorHelpAndBareNamespace(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "help topic", args: []string{"help", "gong"}, want: []string{"pm connectors inspect gong", "calls transcript", "Gong"}},
		{name: "bare connector", args: []string{"gong"}, want: []string{"pm gong - Gong command surface", "COMMAND GROUPS", "calls"}},
		{name: "connector help flag", args: []string{"gong", "--help"}, want: []string{"pm gong - Gong command surface", "COMMAND GROUPS", "calls"}},
		{name: "command help flag", args: []string{"gong", "calls", "transcript", "--help"}, want: []string{"pm gong calls transcript", "INTENT", "direct_read", "FLAGS"}},
		{name: "flag only namespace", args: []string{"gong", "--credential", "gong-local"}, want: []string{"pm gong - Gong command surface", "COMMAND GROUPS", "calls"}},
		{name: "false preview is passive", args: []string{"gong", "--preview=false"}, want: []string{"pm gong - Gong command surface", "COMMAND GROUPS", "calls"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr = %s", tt.args, code, stderr.String())
			}
			out := stdout.String()
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Fatalf("Run(%v) help missing %q:\n%s", tt.args, want, out)
				}
			}
		})
	}
}

func TestDynamicConnectorDeepHelpPathsResolveOrReportUsage(t *testing.T) {
	t.Run("real deep command", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"gong", "calls", "transcript", "--help"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(gong calls transcript --help) code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "pm gong calls transcript") {
			t.Fatalf("real deep command help missing command manual:\n%s", stdout.String())
		}
	})

	t.Run("unknown deep command", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"gong", "calls", "definitely-not-real", "--help"}, &stdout, &stderr)
		if code == 0 {
			t.Fatalf("Run(gong calls definitely-not-real --help) code = 0, want usage error; stdout=%s stderr=%s", stdout.String(), stderr.String())
		}
		out := stdout.String() + stderr.String()
		for _, want := range []string{`unknown command "calls definitely-not-real"`} {
			if !strings.Contains(out, want) {
				t.Fatalf("unknown deep command output missing %q:\nstdout=%s\nstderr=%s", want, stdout.String(), stderr.String())
			}
		}
		if strings.Contains(out, "pm gong - Gong command surface") {
			t.Fatalf("unknown deep command rendered connector root help:\nstdout=%s\nstderr=%s", stdout.String(), stderr.String())
		}
	})

	t.Run("unknown deep command JSON", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"gong", "calls", "definitely-not-real", "--help", "--json"}, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("Run(gong calls definitely-not-real --help --json) code = %d, want usage error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		var env struct {
			Error struct {
				Category string `json:"category"`
				Code     string `json:"code"`
				Message  string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
			t.Fatalf("decode JSON error: %v\nstdout=%s", err, stdout.String())
		}
		if env.Error.Category != "usage" || env.Error.Code != "usage_error" || !strings.Contains(env.Error.Message, `unknown command "calls definitely-not-real"`) {
			t.Fatalf("error = %+v, want usage_error for unresolved path", env.Error)
		}
	})

	t.Run("unknown group with trailing help value JSON", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"gong", "definitely-not-real", "--help", "trailing", "--json"}, &stdout, &stderr)
		if code != 2 {
			t.Fatalf("Run(gong definitely-not-real --help trailing --json) code = %d, want usage error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		var env struct {
			Error struct {
				Category string `json:"category"`
				Code     string `json:"code"`
				Message  string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
			t.Fatalf("decode JSON error: %v\nstdout=%s", err, stdout.String())
		}
		if env.Error.Category != "usage" || env.Error.Code != "usage_error" || !strings.Contains(env.Error.Message, `unknown command "definitely-not-real"`) {
			t.Fatalf("error = %+v, want usage_error for unresolved path", env.Error)
		}
	})

	t.Run("valid group with trailing help value", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := cli.Run([]string{"gong", "calls", "--help", "trailing"}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(gong calls --help trailing) code = %d, want success; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
		}
		if !strings.Contains(stdout.String(), "pm gong calls - Gong calls commands") {
			t.Fatalf("valid group help missing group manual:\n%s", stdout.String())
		}
	})
}

func TestGongCallsListHelpDocumentsDateFlagsAndLimitOutputCap(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"gong", "calls", "list", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(gong calls list --help) code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"pm gong calls list",
		"--from (string): Inclusive ISO-8601 lower bound mapped to Gong fromDateTime",
		"--to (string): Exclusive ISO-8601 upper bound mapped to Gong toDateTime",
		"--limit (integer): Maximum PM ETL records to emit; does not control Gong page size or total provider-side results",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("calls list help missing %q:\n%s", want, out)
		}
	}
}

func TestDynamicConnectorSharedPassiveFlagRendersHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"github", "--credential", "github-local"}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "pm github") {
		t.Fatalf("Run(github --credential) code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestDynamicConnectorInvalidFlagOnlyInvocationsAreUsageErrors(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{args: []string{"gong", "--bogus"}, want: "missing connector command path"},
		{args: []string{"gong", "--plan", "rplan_fixture", "--preview"}, want: "missing connector command path"},
		{args: []string{"gong", "--plan="}, want: "missing connector command path"},
		{args: []string{"gong", "--approval-token-stdin=value"}, want: "--approval-token-stdin must be a bare stdin marker"},
		{args: []string{"gong", "--confirm="}, want: "missing connector command path"},
	} {
		t.Run(strings.Join(tc.args[1:], "_"), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tc.args, &stdout, &stderr)
			if code != 2 || !strings.Contains(stdout.String()+stderr.String(), tc.want) {
				t.Fatalf("Run(%v) code = %d stdout=%s stderr=%s", tc.args, code, stdout.String(), stderr.String())
			}
		})
	}
}

func TestDynamicConnectorValuedApprovalStdinMarkerDoesNotRenderGroupHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"gong", "calls", "--approval-token-stdin=value"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(gong calls --approval-token-stdin=value) code = %d, want usage error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String()+stderr.String(), "--approval-token-stdin must be a bare stdin marker") {
		t.Fatalf("valued approval stdin marker did not return its validation error: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	if strings.Contains(stdout.String(), "pm gong calls - Gong calls commands") {
		t.Fatalf("valued approval stdin marker rendered passive group help: stdout=%s", stdout.String())
	}
}

func TestDynamicConnectorWriteHelpDocumentsApprovalStdinMarker(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"github", "issue", "close", "--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(github issue close --help) code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "--approval-token-stdin") {
		t.Fatalf("GitHub write help omitted the approval stdin marker: stdout=%s", stdout.String())
	}
}

func TestDynamicConnectorHelpRejectsApprovalCarrierArguments(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "valued stdin marker",
			args: []string{"github", "issue", "close", "--help", "--approval-token-stdin=carrier-value"},
			want: "--approval-token-stdin must be a bare stdin marker",
		},
		{
			name: "retired argv carrier",
			args: []string{"github", "issue", "close", "--help", "--approve", "carrier-value"},
			want: "approval tokens must be supplied with --approval-token-stdin",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tc.args, &stdout, &stderr)
			out := stdout.String() + stderr.String()
			if code != 2 {
				t.Fatalf("Run(%v) code = %d, want usage error; stdout=%s stderr=%s", tc.args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(out, tc.want) {
				t.Fatalf("Run(%v) output = %q, want %q", tc.args, out, tc.want)
			}
			if strings.Contains(out, "carrier-value") {
				t.Fatalf("Run(%v) echoed the approval carrier value: %s", tc.args, out)
			}
			if strings.Contains(stdout.String(), "pm github issue close") {
				t.Fatalf("Run(%v) rendered help before rejecting the approval carrier: %s", tc.args, stdout.String())
			}
		})
	}
}

func TestDynamicConnectorUnknownPathIsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"amazon-sqs", "not-a-command", "--json"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(amazon-sqs not-a-command --json) code = %d, want usage error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}

	var env struct {
		Error struct {
			Category string `json:"category"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode error JSON: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if env.Error.Category != "usage" || env.Error.Code != "usage_error" {
		t.Fatalf("error = %+v, want usage_error; stdout=%s stderr=%s", env.Error, stdout.String(), stderr.String())
	}
	if !strings.Contains(env.Error.Message, `unknown command "not-a-command"`) {
		t.Fatalf("message = %q, want unknown command", env.Error.Message)
	}
	if strings.Contains(stdout.String()+stderr.String(), "connector_command_blocked") {
		t.Fatalf("unknown command returned policy block output: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestDynamicConnectorEmptyLifecycleFlagsWithCommandAreUsageErrors(t *testing.T) {
	for _, flag := range []string{"--plan=", "--confirm=", "--plan", "--confirm"} {
		t.Run(flag, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run([]string{"github", "issue", "create", flag}, &stdout, &stderr)
			if code != 2 || !strings.Contains(stdout.String()+stderr.String(), "requires a value") {
				t.Fatalf("Run(github issue create %s) code = %d stdout=%s stderr=%s", flag, code, stdout.String(), stderr.String())
			}
		})
	}

	t.Run("bare flag before repeated value", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		args := []string{"github", "issue", "create", "--plan", "--plan", "rplan_fixture"}
		code := cli.Run(args, &stdout, &stderr)
		if code != 2 || !strings.Contains(stdout.String()+stderr.String(), "requires a value") {
			t.Fatalf("Run(%v) code = %d stdout=%s stderr=%s", args, code, stdout.String(), stderr.String())
		}
	})
}

func TestDynamicConnectorHelpJSONIsAgentReadable(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"help", "gong", "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("Run(help gong --json) code = %d stderr = %s", code, stderr.String())
	}
	var env struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Manual  string `json:"manual"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &env); err != nil {
		t.Fatalf("decode JSON help: %v\n%s", err, stdout.String())
	}
	if env.Kind != "CommandManual" || env.Command != "gong" || !strings.Contains(env.Manual, "calls transcript") {
		t.Fatalf("help envelope = %+v", env)
	}
}

func TestGongTranscriptCommandAllowsDeclaredResponseCap(t *testing.T) {
	response := `{"transcript":"` + strings.Repeat("x", (1<<20)+1024) + `"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v2/calls/transcript" {
			t.Errorf("request = %s %s, want POST /v2/calls/transcript", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	t.Setenv("PM_TEST_GONG_ACCESS_KEY", "fixture-access-key")
	t.Setenv("PM_TEST_GONG_ACCESS_KEY_SECRET", "fixture-access-key-secret")
	runCLI(t, []string{
		"credentials", "add", "gong-local",
		"--connector", "gong",
		"--from-env", "access_key=PM_TEST_GONG_ACCESS_KEY",
		"--from-env", "access_key_secret=PM_TEST_GONG_ACCESS_KEY_SECRET",
		"--config", "base_url=" + server.URL,
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"gong", "calls", "transcript",
		"--credential", "gong-local",
		"--call-id", "call-fixture",
		"--root", root,
		"--json",
	})
	if !strings.Contains(stdout, `"kind": "ConnectorCommandDirectRead"`) {
		t.Fatalf("output missing direct-read envelope: %.200s", stdout)
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
		"DECLARATION-BOUND STRUCTURED WRITE INPUTS",
		"There is no raw\n  --body flag",
		"pm connectors catalog --capability write --json",
		"pm connectors certify <connector> [--full | --direct-read-only | --write-only] [--resume] [--external-proof] [--full-parity] [--from-env field=ENV | --value-stdin field] [--json]",
		"legacy_unverified",
		"provider-artifact",
		"provenance evidence",
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
	for _, want := range []string{"NAME", "SYNOPSIS", "AUTHENTICATION", "ETL STREAMS", "REVERSE ETL ACTIONS", "AGENT WORKFLOW", "CERTIFICATION", "COMMUNITY BUILD, UNCERTIFIED"} {
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
	for _, want := range []string{`"kind": "ConnectorList"`, `"name": "sample"`, `"name": "warehouse"`, `"name": "akeneo"`, `"name": "github"`, `"name": "postgres"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
	for _, forbidden := range []string{`"name": "source-github"`, `"name": "destination-postgres"`} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("json output contains legacy slug %q:\n%s", forbidden, out)
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
	for _, want := range []string{`"kind": "SyncModeBenchmark"`, `"full_refresh_append"`, `"incremental_append"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("json output missing %q:\n%s", want, out)
		}
	}
	for _, forbidden := range []string{`"full_refresh_overwrite_deduped"`, `"incremental_append_deduped"`} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("json output contains typed-only compatibility name %q:\n%s", forbidden, out)
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
		"Compatibility name for typed full_overwrite admission",
		"incremental_append",
		"incremental_append_deduped",
		"Compatibility name for typed incremental_dedupe admission",
		"incremental_dedupe",
		"incremental_dedupe_history",
		"retains deduplicated source versions with _valid_from, _valid_to, and _is_current fields",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("etl help missing %q:\n%s", want, out)
		}
	}
}

func TestETLRejectsLegacyPrefixedConnectorCommands(t *testing.T) {
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
			if !strings.Contains(stderr.String()+stdout.String(), `connector "source-strava" uses a legacy source-/destination- prefix; use bare connector name "strava"`) {
				t.Fatalf("Run(%v) did not explain legacy prefix migration; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
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

func TestGitHubCommandSurfaceRunsStreamBackedIssueList(t *testing.T) {
	var gotPath, gotState string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotState = r.URL.Query().Get("state")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{
				"id": 101,
				"node_id": "I_kwDOAA",
				"number": 101,
				"state": "closed",
				"title": "closed issue",
				"user": {"login": "octocat", "id": 1},
				"updated_at": "2026-07-06T00:00:00Z"
			}
		]`))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"github", "issue", "list",
		"--credential", "github-local",
		"--state", "closed",
		"--limit", "1",
		"--root", root,
		"--json",
	})
	if gotPath != "/repos/octocat/hello-world/issues" {
		t.Fatalf("request path = %q, want /repos/octocat/hello-world/issues", gotPath)
	}
	if gotState != "closed" {
		t.Fatalf("request state = %q, want closed", gotState)
	}

	var env struct {
		Kind    string `json:"kind"`
		Command string `json:"command"`
		Stream  string `json:"stream"`
		Count   int    `json:"count"`
		Records []struct {
			NodeID     string `json:"node_id"`
			State      string `json:"state"`
			Repository string `json:"repository"`
			UserLogin  string `json:"user_login"`
		} `json:"records"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandRead" || env.Command != "issue list" || env.Stream != "issues" || env.Count != 1 {
		t.Fatalf("envelope = %+v, want kind ConnectorCommandRead command issue list stream issues count 1", env)
	}
	if len(env.Records) != 1 || env.Records[0].State != "closed" || env.Records[0].Repository != "octocat/hello-world" || env.Records[0].UserLogin != "octocat" {
		t.Fatalf("records = %+v, want projected GitHub issue record", env.Records)
	}
}

func TestGitHubCommandSurfaceClampsOversizedLimit(t *testing.T) {
	const wantLimit = 10000
	var body strings.Builder
	body.WriteByte('[')
	for i := 1; i <= wantLimit+1; i++ {
		if i > 1 {
			body.WriteByte(',')
		}
		fmt.Fprintf(&body, `{
			"id": %d,
			"node_id": "I_%d",
			"number": %d,
			"state": "open",
			"title": "issue %d",
			"user": {"login": "octocat", "id": 1},
			"updated_at": "2026-07-06T00:00:00Z"
		}`, i, i, i, i)
	}
	body.WriteByte(']')

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body.String()))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"github", "issue", "list",
		"--credential", "github-local",
		"--limit", fmt.Sprint(wantLimit + 1),
		"--root", root,
		"--json",
	})

	var env struct {
		Kind  string `json:"kind"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandRead" || env.Count != wantLimit {
		t.Fatalf("envelope = %+v, want clamped ConnectorCommandRead count %d", env, wantLimit)
	}
}

func TestGitHubCommandSurfaceRunsDirectReadFile(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"README.md","type":"file","encoding":"base64","content":"SGVsbG8=","download_url":"https://raw.example.test/README.md"}`))
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	stdout, _ := runCLI(t, []string{
		"github", "repo", "read-file",
		"--credential", "github-local",
		"--path", "README.md",
		"--root", root,
		"--json",
	})
	if gotPath != "/repos/octocat/hello-world/contents/README.md" {
		t.Fatalf("request path = %q, want contents file path", gotPath)
	}

	var env struct {
		Kind     string         `json:"kind"`
		Command  string         `json:"command"`
		Method   string         `json:"method"`
		Path     string         `json:"path"`
		Status   int            `json:"status"`
		Response map[string]any `json:"response"`
	}
	if err := json.Unmarshal([]byte(stdout), &env); err != nil {
		t.Fatalf("decode json: %v\n%s", err, stdout)
	}
	if env.Kind != "ConnectorCommandDirectRead" || env.Command != "repo read-file" || env.Method != "GET" || env.Status != http.StatusOK {
		t.Fatalf("envelope = %+v, want direct-read README result", env)
	}
	if env.Response["name"] != "README.md" || env.Response["type"] != "file" {
		t.Fatalf("response = %+v, want README file metadata", env.Response)
	}
	if _, ok := env.Response["content"]; ok {
		t.Fatalf("response leaked content: %+v", env.Response)
	}
	if _, ok := env.Response["download_url"]; ok {
		t.Fatalf("response leaked download_url: %+v", env.Response)
	}
	if env.Response["content_redacted"] != true || env.Response["download_url_redacted"] != true {
		t.Fatalf("response redaction markers = %+v, want content and download_url redacted", env.Response)
	}

	gotPath = ""
	runCLI(t, []string{
		"github", "repo", "read-file",
		"--credential", "github-local",
		"--path", "help",
		"--root", root,
		"--json",
	})
	if gotPath != "/repos/octocat/hello-world/contents/help" {
		t.Fatalf("request path for help-valued flag = %q, want contents/help", gotPath)
	}
}

// TestGitHubDirectReadParametersAndPageContextReachWire drives the real CLI,
// embedded GitHub bundle, command runner, and direct-read executor against a
// known-larger fixture. It asserts returned records and the server-observed
// query, never a successful exit code alone.
func TestGitHubDirectReadParametersAndPageContextReachWire(t *testing.T) {
	const since = "2026-01-02T03:04:05Z"
	type observedRequest struct {
		path  string
		query string
	}
	var observed []observedRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed = append(observed, observedRequest{path: r.URL.Path, query: r.URL.RawQuery})
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/repos/octocat/hello-world/notifications":
			if got := r.URL.Query().Get("since"); got != since {
				t.Errorf("notifications since = %q, want %q", got, since)
			}
			if got := r.URL.Query().Get("per_page"); got != "100" {
				t.Errorf("notifications per_page = %q, want declared 100", got)
			}
			_, _ = w.Write([]byte(`[{"id":"thread-1"},{"id":"thread-2"}]`))
		case "/repos/octocat/hello-world/pulls/42/files":
			if got := r.URL.Query().Get("per_page"); got != "100" {
				t.Errorf("pull files per_page = %q, want declared 100", got)
			}
			page := r.URL.Query().Get("page")
			if page == "" {
				page = "1"
			}
			count := 100
			if page == "2" {
				count = 20
			}
			if page != "1" && page != "2" {
				t.Errorf("pull files page = %q, want 1 or 2", page)
			}
			rows := make([]map[string]any, count)
			for i := range rows {
				rows[i] = map[string]any{"filename": fmt.Sprintf("file-%03d", i)}
			}
			_ = json.NewEncoder(w).Encode(rows)
		default:
			t.Errorf("unexpected request %s", r.URL.Path)
		}
	}))
	t.Cleanup(srv.Close)

	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "base_url=" + srv.URL,
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	// Both rejections occur before the handler is reached, and the enum
	// rejection names every accepted option instead of deferring to GitHub.
	for _, tc := range []struct {
		args         []string
		want         string
		usageRefusal bool
	}{
		{args: []string{"github", "issue", "list", "--credential", "github-local", "--state", "impossible", "--root", root, "--json"}, want: "all|closed|open"},
		{args: []string{"github", "pulls", "files", "view", "--credential", "github-local", "--root", root, "--json"}, want: "missing required flag --pull-number", usageRefusal: true},
	} {
		var stdout, stderr bytes.Buffer
		code := cli.Run(tc.args, &stdout, &stderr)
		if code == 0 {
			t.Fatalf("Run(%v) code = 0, want parser refusal", tc.args)
		}
		if got := stdout.String() + stderr.String(); !strings.Contains(got, tc.want) {
			t.Fatalf("Run(%v) output = %q, want %q", tc.args, got, tc.want)
		}
		if len(observed) != 0 {
			t.Fatalf("Run(%v) reached the provider: %#v", tc.args, observed)
		}
		if tc.usageRefusal {
			if code != 2 {
				t.Fatalf("Run(%v) exit code = %d, want 2 usage error", tc.args, code)
			}
			var envelope struct {
				Error struct {
					Category string `json:"category"`
					Code     string `json:"code"`
					Message  string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
				t.Fatalf("decode usage refusal: %v\nstdout=%s", err, stdout.String())
			}
			if envelope.Error.Category != "usage" || envelope.Error.Code != "usage_error" || !strings.Contains(envelope.Error.Message, tc.want) {
				t.Fatalf("usage refusal = %+v, want category=usage code=usage_error message containing %q", envelope.Error, tc.want)
			}
		}
	}

	notificationOut, _ := runCLI(t, []string{
		"github", "notifications", "view", "--credential", "github-local",
		"--since", since, "--root", root, "--json",
	})
	var notification struct {
		Response []map[string]any `json:"response"`
		Page     struct {
			Records  int  `json:"records"`
			Size     int  `json:"size"`
			Complete bool `json:"complete"`
		} `json:"page"`
	}
	if err := json.Unmarshal([]byte(notificationOut), &notification); err != nil {
		t.Fatalf("decode notifications output: %v\n%s", err, notificationOut)
	}
	if len(notification.Response) != 2 || notification.Page.Records != 2 || notification.Page.Size != 100 || !notification.Page.Complete {
		t.Fatalf("notifications response/page = %#v/%+v, want two real rows and a complete 100-size page", notification.Response, notification.Page)
	}

	readPage := func(page string) struct {
		Response []map[string]any `json:"response"`
		Page     struct {
			Records    int  `json:"records"`
			Size       int  `json:"size"`
			Number     int  `json:"number"`
			NextNumber int  `json:"next_number"`
			HasMore    bool `json:"has_more"`
			Complete   bool `json:"complete"`
		} `json:"page"`
	} {
		args := []string{"github", "pulls", "files", "view", "--credential", "github-local", "--pull-number", "42", "--root", root, "--json"}
		if page != "" {
			args = append(args, "--page", page)
		}
		stdout, _ := runCLI(t, args)
		var out struct {
			Response []map[string]any `json:"response"`
			Page     struct {
				Records    int  `json:"records"`
				Size       int  `json:"size"`
				Number     int  `json:"number"`
				NextNumber int  `json:"next_number"`
				HasMore    bool `json:"has_more"`
				Complete   bool `json:"complete"`
			} `json:"page"`
		}
		if err := json.Unmarshal([]byte(stdout), &out); err != nil {
			t.Fatalf("decode pull-files page %q: %v\n%s", page, err, stdout)
		}
		return out
	}

	first := readPage("")
	if len(first.Response) != 100 || first.Page.Records != 100 || first.Page.Size != 100 || first.Page.Number != 1 || first.Page.NextNumber != 2 || !first.Page.HasMore || first.Page.Complete {
		t.Fatalf("first pull-files page = %#v/%+v, want 100 rows and an addressable incomplete page", first.Response, first.Page)
	}
	second := readPage("2")
	if len(second.Response) != 20 || second.Page.Records != 20 || second.Page.Number != 2 || second.Page.HasMore || !second.Page.Complete {
		t.Fatalf("second pull-files page = %#v/%+v, want final 20-row page", second.Response, second.Page)
	}
	if total := len(first.Response) + len(second.Response); total != 120 {
		t.Fatalf("records reached across pages = %d, want 120", total)
	}
}

func TestGitHubCommandSurfacePlansReverseETLCommand(t *testing.T) {
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	runCLI(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=octocat",
		"--config", "repo=hello-world",
		"--config", "public_access=true",
		"--root", root,
		"--json",
	})

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"github", "issue", "create",
		"--title", "Ship connector command plans",
		"--credential", "github-local",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("issue create code = %d stderr=%s stdout=%s", code, stderr.String(), stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{`"kind": "ConnectorCommandWritePlan"`, `"connector_command": "issue create"`, `"action": "create_issue"`, `"approval_required": true`} {
		if !strings.Contains(out, want) {
			t.Fatalf("planned command output missing %q:\nstdout=%s\nstderr=%s", want, out, stderr.String())
		}
	}
	if strings.Contains(out, "approval_token") || strings.Contains(out, "approval_token_hash") ||
		strings.Contains(out, "connector_command_record") {
		t.Fatalf("plan JSON leaked approval or raw command payload:\n%s", out)
	}
}

func TestGitHubCommandSurfaceIssueDeleteReachesCredentialResolution(t *testing.T) {
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"github", "issue", "delete",
		"--input", `{"issueId":"I_kwDOA"}`,
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("issue delete code = 0, want missing credential; stdout=%s", stdout.String())
	}
	out := stdout.String()
	for _, want := range []string{`"category": "internal"`, `"code": "internal_error"`, "missing --credential"} {
		if !strings.Contains(out, want) {
			t.Fatalf("reachable operation output missing %q:\nstdout=%s\nstderr=%s", want, out, stderr.String())
		}
	}
	if strings.Contains(out, "connector_command_blocked") || strings.Contains(stderr.String(), "connector_command_blocked") {
		t.Fatalf("issue delete stayed blocked instead of reaching its typed declared operation:\nstdout=%s\nstderr=%s", out, stderr.String())
	}
}

func TestRootHelpListsDynamicConnectorCommands(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"--help"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("root help code = %d stderr=%s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{"CONNECTOR COMMANDS", "pm bahmni <command>", "pm github <command>", "pm gong <command>"} {
		if !strings.Contains(out, want) {
			t.Fatalf("root help missing %q:\n%s", want, out)
		}
	}
}

func TestBahmniCommandSurfaceHelpScopes(t *testing.T) {
	tests := []struct {
		name   string
		args   []string
		want   []string
		forbid []string
	}{
		{
			name:   "connector root",
			args:   []string{"bahmni", "--help"},
			want:   []string{"NAME", "pm bahmni - Bahmni command surface", "COMMAND GROUPS", "appointments", "pm bahmni appointments --help"},
			forbid: []string{"CAPABILITIES", "REVERSE ETL ACTIONS"},
		},
		{
			name:   "appointment group",
			args:   []string{"bahmni", "appointments", "--help"},
			want:   []string{"NAME", "pm bahmni appointments - Bahmni appointments commands", "appointments list", "appointments create", "appointments status-change", "appointments provider-response", "appointments reschedule", "appointment_date only"},
			forbid: []string{"patients create", "drug_orders create"},
		},
		{
			name:   "appointment group bare",
			args:   []string{"bahmni", "appointments"},
			want:   []string{"NAME", "pm bahmni appointments - Bahmni appointments commands", "appointments list", "appointments create", "appointments status-change", "appointments provider-response", "appointments reschedule", "appointment_date only"},
			forbid: []string{"patients create", "drug_orders create"},
		},
		{
			name:   "appointment group passive flags",
			args:   []string{"bahmni", "appointments", "--credential", "bahmni-local", "--preview=false"},
			want:   []string{"NAME", "pm bahmni appointments - Bahmni appointments commands", "appointments list", "appointments create", "appointments status-change", "appointments provider-response", "appointments reschedule", "appointment_date only"},
			forbid: []string{"patients create", "drug_orders create", `credential "bahmni-local" not found`},
		},
		{
			name:   "appointment group short help",
			args:   []string{"bahmni", "appointments", "-h"},
			want:   []string{"NAME", "pm bahmni appointments - Bahmni appointments commands", "appointments list", "appointments create", "appointments status-change", "appointments provider-response", "appointments reschedule", "appointment_date only"},
			forbid: []string{"patients create", "drug_orders create"},
		},
		{
			name:   "appointment create command",
			args:   []string{"bahmni", "appointments", "create", "--help"},
			want:   []string{"NAME", "pm bahmni appointments create", "INTENT", "reverse_etl", "AVAILABILITY", "implemented", "APPROVAL", "plan -> preview -> approval -> execute", "FLAGS", "--patient-uuid", "--service-uuid", "--start-date-time"},
			forbid: []string{"patients create", "drug_orders create"},
		},
		{
			name:   "appointment create command short help",
			args:   []string{"bahmni", "appointments", "create", "-h"},
			want:   []string{"NAME", "pm bahmni appointments create", "INTENT", "reverse_etl", "AVAILABILITY", "implemented", "APPROVAL", "plan -> preview -> approval -> execute", "FLAGS", "--patient-uuid", "--service-uuid", "--start-date-time"},
			forbid: []string{"patients create", "drug_orders create"},
		},
		{
			name:   "appointment list command",
			args:   []string{"bahmni", "appointments", "list", "--help"},
			want:   []string{"NAME", "pm bahmni appointments list", "INTENT", "etl", "AVAILABILITY", "implemented", "STREAM", "appointments", "appointment_date only"},
			forbid: []string{"patient_uuid", "patients create"},
		},
		{
			name:   "patient search redaction help",
			args:   []string{"bahmni", "bahmnicore", "patient-search", "--help"},
			want:   []string{"NAME", "pm bahmni bahmnicore patient-search", "declared patient-search sensitive fields", "identifier", "addressFieldValue"},
			forbid: []string{"not field-redacted"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tt.args, &stdout, &stderr)
			if code != 0 {
				t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", tt.args, code, stderr.String(), stdout.String())
			}
			out := stdout.String()
			for _, want := range tt.want {
				if !strings.Contains(out, want) {
					t.Fatalf("help for %v missing %q:\n%s", tt.args, want, out)
				}
			}
			for _, forbidden := range tt.forbid {
				if strings.Contains(out, forbidden) {
					t.Fatalf("help for %v unexpectedly included %q:\n%s", tt.args, forbidden, out)
				}
			}
		})
	}
}

func TestBahmniBareCommandGroupInvalidMultiPartPathIsNotHelp(t *testing.T) {
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"bahmni", "appointments", "bogus", "--credential", "absent", "--root", root, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("invalid command path code = 0, want usage error")
	}
	out := stdout.String() + stderr.String()
	for _, want := range []string{"unknown command", "appointments bogus"} {
		if !strings.Contains(out, want) {
			t.Fatalf("invalid command path output missing %q:\nstdout=%s\nstderr=%s", want, stdout.String(), stderr.String())
		}
	}
	for _, forbidden := range []string{"Bahmni appointments commands", `credential "absent" not found`} {
		if strings.Contains(out, forbidden) {
			t.Fatalf("invalid command path unexpectedly included %q:\nstdout=%s\nstderr=%s", forbidden, stdout.String(), stderr.String())
		}
	}
}

func TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked(t *testing.T) {
	surface := bahmniCommandSurface(t)
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})

	for _, cmd := range surface.Commands {
		cmd := cmd
		t.Run(strings.ReplaceAll(cmd.Path, " ", "_"), func(t *testing.T) {
			helpArgs := append([]string{"bahmni"}, strings.Fields(cmd.Path)...)
			helpArgs = append(helpArgs, "--help")
			var helpOut, helpErr bytes.Buffer
			if code := cli.Run(helpArgs, &helpOut, &helpErr); code != 0 {
				t.Fatalf("help Run(%v) code = %d stderr=%s stdout=%s", helpArgs, code, helpErr.String(), helpOut.String())
			}
			if !strings.Contains(helpOut.String(), "pm bahmni "+cmd.Path) {
				t.Fatalf("command help for %q missing exact command path:\n%s", cmd.Path, helpOut.String())
			}

			shortHelpArgs := append([]string{"bahmni"}, strings.Fields(cmd.Path)...)
			shortHelpArgs = append(shortHelpArgs, "-h")
			var shortHelpOut, shortHelpErr bytes.Buffer
			if code := cli.Run(shortHelpArgs, &shortHelpOut, &shortHelpErr); code != 0 {
				t.Fatalf("short help Run(%v) code = %d stderr=%s stdout=%s", shortHelpArgs, code, shortHelpErr.String(), shortHelpOut.String())
			}
			if !strings.Contains(shortHelpOut.String(), "pm bahmni "+cmd.Path) {
				t.Fatalf("short command help for %q missing exact command path:\n%s", cmd.Path, shortHelpOut.String())
			}

			runArgs := append([]string{"bahmni"}, strings.Fields(cmd.Path)...)
			runArgs = append(runArgs, "--credential", "absent", "--root", root, "--json")
			var stdout, stderr bytes.Buffer
			code := cli.Run(runArgs, &stdout, &stderr)
			if code == 0 {
				t.Fatalf("Run(%v) code = 0, want credential lookup or explicit block", runArgs)
			}
			out := stdout.String() + stderr.String()
			if cmd.Availability == "implemented" {
				if !strings.Contains(out, `credential "absent" not found`) {
					t.Fatalf("implemented command %q did not reach credential lookup:\nstdout=%s\nstderr=%s", cmd.Path, stdout.String(), stderr.String())
				}
				if strings.Contains(out, "unknown command") || strings.Contains(out, "connector_command_blocked") {
					t.Fatalf("implemented command %q looked unregistered or blocked:\nstdout=%s\nstderr=%s", cmd.Path, stdout.String(), stderr.String())
				}
				return
			}
			for _, want := range []string{`"code": "connector_command_blocked"`, cmd.Path, cmd.Availability} {
				if !strings.Contains(out, want) {
					t.Fatalf("blocked command %q output missing %q:\nstdout=%s\nstderr=%s", cmd.Path, want, stdout.String(), stderr.String())
				}
			}
			if strings.Contains(out, `credential "absent" not found`) {
				t.Fatalf("blocked command %q attempted credential lookup first:\nstdout=%s\nstderr=%s", cmd.Path, stdout.String(), stderr.String())
			}
		})
	}
}

func TestBahmniAppointmentAliasSuggestion(t *testing.T) {
	root := t.TempDir()
	runCLI(t, []string{"init", "--root", root, "--json"})
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"bahmni", "appoint", "list", "--credential", "absent", "--root", root, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("appoint alias code = 0, want actionable usage error")
	}
	out := stdout.String() + stderr.String()
	for _, want := range []string{"unknown command", "did you mean \"appointments list\""} {
		if !strings.Contains(out, want) {
			t.Fatalf("appoint alias output missing %q:\nstdout=%s\nstderr=%s", want, stdout.String(), stderr.String())
		}
	}
}

func bahmniCommandSurface(t *testing.T) *connectors.CommandSurface {
	t.Helper()
	connector, ok := bundleregistry.New().Get("bahmni")
	if !ok {
		t.Fatal("bahmni connector not found")
	}
	provider, ok := connector.(connectors.CommandSurfaceProvider)
	if !ok || provider.CommandSurface() == nil {
		t.Fatal("bahmni command surface not found")
	}
	return provider.CommandSurface()
}

func runCLI(t *testing.T, args []string) (stdout string, stderr string) {
	t.Helper()
	var outBuf, errBuf bytes.Buffer
	code := cli.Run(args, &outBuf, &errBuf)
	if code != 0 {
		t.Fatalf("Run(%v) code = %d stderr=%s stdout=%s", args, code, errBuf.String(), outBuf.String())
	}
	return outBuf.String(), errBuf.String()
}
