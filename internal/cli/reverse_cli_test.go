package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/warehouse"
)

func TestReverseETLCLIWorkflowIsScriptableAndApprovalBounded(t *testing.T) {
	root := setupReverseCLIProject(t)

	var planStdout, planStderr bytes.Buffer
	code := cli.Run([]string{
		"reverse", "plan", "customers_to_outbox",
		"--source-table", "sample_customers",
		"--destination", "outbox:outbox-local",
		"--map", "id:external_id",
		"--map", "name:full_name",
		"--map", "email:email",
		"--root", root,
	}, &planStdout, &planStderr)
	if code != 0 {
		t.Fatalf("reverse plan code = %d stderr = %s", code, planStderr.String())
	}
	planID := extractReverseField(t, planStdout.String(), `Created reverse plan (\S+)`)
	token := extractReverseField(t, planStdout.String(), `Approval token: (\S+)`)
	if planID == "" || token == "" {
		t.Fatalf("missing plan id or approval token:\n%s", planStdout.String())
	}

	var previewStdout, previewStderr bytes.Buffer
	code = cli.Run([]string{"reverse", "preview", planID, "--root", root, "--json"}, &previewStdout, &previewStderr)
	if code != 0 {
		t.Fatalf("reverse preview code = %d stderr = %s", code, previewStderr.String())
	}
	var preview struct {
		Kind string `json:"kind"`
		Plan struct {
			ID            string           `json:"id"`
			RecordCount   int              `json:"record_count"`
			ApprovalToken string           `json:"approval_token"`
			Sample        []map[string]any `json:"sample"`
		} `json:"plan"`
	}
	if err := json.Unmarshal(previewStdout.Bytes(), &preview); err != nil {
		t.Fatalf("preview JSON decode: %v\n%s", err, previewStdout.String())
	}
	if preview.Kind != "ReversePlanPreview" || preview.Plan.ID != planID || preview.Plan.RecordCount != 3 {
		t.Fatalf("unexpected preview: %+v", preview)
	}
	if preview.Plan.ApprovalToken != "" {
		t.Fatalf("preview leaked approval token: %+v", preview.Plan)
	}
	if len(preview.Plan.Sample) == 0 || preview.Plan.Sample[0]["external_id"] != "cus_001" || preview.Plan.Sample[0]["id"] != nil {
		t.Fatalf("preview sample should show mapped destination payload, got %+v", preview.Plan.Sample)
	}

	var deniedStdout, deniedStderr bytes.Buffer
	code = cli.Run([]string{"reverse", "run", planID, "--root", root, "--json"}, &deniedStdout, &deniedStderr)
	if code == 0 {
		t.Fatalf("reverse run without approval unexpectedly succeeded: stdout=%s", deniedStdout.String())
	}
	if !strings.Contains(deniedStderr.String(), "requires --approval-token-stdin") {
		t.Fatalf("missing approval error: stderr=%s stdout=%s", deniedStderr.String(), deniedStdout.String())
	}

	var runStdout, runStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{"reverse", "run", planID, "--approval-token-stdin", "--root", root, "--json"}, token+"\n", &runStdout, &runStderr)
	if code != 0 {
		t.Fatalf("reverse run code = %d stderr = %s stdout = %s", code, runStderr.String(), runStdout.String())
	}
	var runResult struct {
		Kind string `json:"kind"`
		Run  struct {
			ID               string `json:"id"`
			Status           string `json:"status"`
			RecordsSucceeded int    `json:"records_succeeded"`
		} `json:"run"`
	}
	if err := json.Unmarshal(runStdout.Bytes(), &runResult); err != nil {
		t.Fatalf("run JSON decode: %v\n%s", err, runStdout.String())
	}
	if runResult.Kind != "ReverseRun" || runResult.Run.Status != "completed" || runResult.Run.RecordsSucceeded != 3 {
		t.Fatalf("unexpected run result: %+v", runResult)
	}

	var statusStdout, statusStderr bytes.Buffer
	code = cli.Run([]string{"reverse", "status", runResult.Run.ID, "--root", root, "--json"}, &statusStdout, &statusStderr)
	if code != 0 {
		t.Fatalf("reverse status code = %d stderr = %s", code, statusStderr.String())
	}
	if !strings.Contains(statusStdout.String(), `"kind": "ReverseRun"`) || !strings.Contains(statusStdout.String(), `"status": "completed"`) {
		t.Fatalf("unexpected status output:\n%s", statusStdout.String())
	}

	var listStdout, listStderr bytes.Buffer
	code = cli.Run([]string{"reverse", "list", "--root", root, "--json"}, &listStdout, &listStderr)
	if code != 0 {
		t.Fatalf("reverse list code = %d stderr = %s", code, listStderr.String())
	}
	if !strings.Contains(listStdout.String(), `"kind": "ReversePlanList"`) || !strings.Contains(listStdout.String(), planID) {
		t.Fatalf("unexpected list output:\n%s", listStdout.String())
	}

	outboxPath := filepath.Join(root, ".polymetrics", "outbox", "customers_to_outbox.jsonl")
	if info, err := os.Stat(outboxPath); err != nil || info.Size() == 0 {
		t.Fatalf("expected outbox file %s: info=%v err=%v", outboxPath, info, err)
	}
}

func TestReverseETLRejectsApprovalCarriersOutsideRun(t *testing.T) {
	root := setupReverseCLIProject(t)
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "plan legacy argv carrier",
			args: []string{
				"reverse", "plan", "rejected-plan",
				"--source-table", "sample_customers",
				"--destination", "outbox:outbox-local",
				"--map", "id:external_id",
				"--approve", "carrier-value",
				"--root", root, "--json",
			},
			want: "approval tokens must be supplied with --approval-token-stdin",
		},
		{
			name: "preview valued stdin marker",
			args: []string{
				"reverse", "preview", "rplan_missing",
				"--approval-token-stdin=carrier-value",
				"--root", root, "--json",
			},
			want: "--approval-token-stdin must be a bare stdin marker",
		},
		{
			name: "list bare stdin marker",
			args: []string{
				"reverse", "list", "--approval-token-stdin",
				"--root", root, "--json",
			},
			want: "--approval-token-stdin is only valid with reverse run",
		},
		{
			name: "status legacy argv carrier",
			args: []string{
				"reverse", "status", "rrun_missing",
				"--approve", "carrier-value",
				"--root", root, "--json",
			},
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
		})
	}

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"reverse", "list", "--root", root, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("reverse list after rejected carriers code = %d stderr = %s", code, stderr.String())
	}
	var result struct {
		Plans []json.RawMessage `json:"plans"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		t.Fatalf("decode reverse list after rejected carriers: %v\n%s", err, stdout.String())
	}
	if len(result.Plans) != 0 {
		t.Fatalf("rejected approval carrier persisted reverse plans: %s", stdout.String())
	}
}

func TestReverseHelpRejectsApprovalCarriers(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "help flag valued stdin marker",
			args: []string{"reverse", "--help", "--approval-token-stdin=carrier-value", "--json"},
			want: "--approval-token-stdin must be a bare stdin marker",
		},
		{
			name: "help command retired argv carrier",
			args: []string{"reverse", "help", "--approve", "carrier-value", "--json"},
			want: "approval tokens must be supplied with --approval-token-stdin",
		},
		{
			name: "help alias bare stdin marker",
			args: []string{"help", "reverse", "--approval-token-stdin", "--json"},
			want: "--approval-token-stdin is only valid with reverse run",
		},
		{
			name: "man alias retired argv carrier",
			args: []string{"man", "reverse", "--approve", "carrier-value", "--json"},
			want: "approval tokens must be supplied with --approval-token-stdin",
		},
		{
			name: "root help retired argv carrier",
			args: []string{"--help", "--approve", "carrier-value", "--json"},
			want: "approval tokens must be supplied with --approval-token-stdin",
		},
		{
			name: "connector help alias valued stdin marker",
			args: []string{"help", "github", "issue", "close", "--approval-token-stdin=carrier-value", "--json"},
			want: "--approval-token-stdin must be a bare stdin marker",
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
			if strings.Contains(out, "CommandManual") {
				t.Fatalf("Run(%v) rendered a manual before rejecting the approval carrier: %s", tc.args, out)
			}
		})
	}
}

func TestReverseApprovalCarrierRejectionUsesConfiguredJSONOutput(t *testing.T) {
	for _, tc := range []struct {
		name     string
		env      string
		fromFile bool
	}{
		{name: "primary environment", env: "POLYMETRICS_JSON"},
		{name: "compatibility environment", env: "PM_JSON"},
		{name: "project configuration", fromFile: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			if tc.env != "" {
				t.Setenv(tc.env, "true")
			}
			if tc.fromFile {
				configPath := filepath.Join(root, ".polymetrics", "config.yaml")
				if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
					t.Fatalf("create config directory: %v", err)
				}
				if err := os.WriteFile(configPath, []byte("json: true\n"), 0o600); err != nil {
					t.Fatalf("write config: %v", err)
				}
			}

			var stdout, stderr bytes.Buffer
			code := cli.Run([]string{"--root", root, "--help", "--approve", "carrier-value"}, &stdout, &stderr)
			if code != 2 {
				t.Fatalf("Run() code = %d, want usage error; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
			}
			var result struct {
				Kind string `json:"kind"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("decode JSON error: %v\n%s", err, stdout.String())
			}
			if result.Kind != "Error" {
				t.Fatalf("JSON result kind = %q, want Error\n%s", result.Kind, stdout.String())
			}
			if strings.Contains(stdout.String()+stderr.String(), "carrier-value") {
				t.Fatalf("carrier rejection echoed its value: stdout=%s stderr=%s", stdout.String(), stderr.String())
			}
		})
	}
}

func TestRawApprovalCarrierRejectionPrecedesGlobalRootDispatch(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "separate valued stdin marker",
			args: []string{"--root", "--approval-token-stdin=carrier-value", "credentials", "list", "--json"},
			want: "--approval-token-stdin must be a bare stdin marker",
		},
		{
			name: "inline valued stdin marker",
			args: []string{"--root=--approval-token-stdin=carrier-value", "credentials", "list", "--json"},
			want: "--approval-token-stdin must be a bare stdin marker",
		},
		{
			name: "inline retired argv carrier",
			args: []string{"--root=--approve=carrier-value", "credentials", "list", "--json"},
			want: "approval tokens must be supplied with --approval-token-stdin",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := cli.Run(tc.args, &stdout, &stderr)
			combined := stdout.String() + stderr.String()
			if code != 2 {
				t.Fatalf("Run(%v) code = %d, want usage error; stdout=%s stderr=%s", tc.args, code, stdout.String(), stderr.String())
			}
			if !strings.Contains(combined, tc.want) {
				t.Fatalf("Run(%v) output = %q, want %q", tc.args, combined, tc.want)
			}
			if strings.Contains(combined, "carrier-value") {
				t.Fatalf("Run(%v) echoed the approval carrier value: %s", tc.args, combined)
			}
			if strings.Contains(combined, "open project") {
				t.Fatalf("Run(%v) opened a project before rejecting the approval carrier: %s", tc.args, combined)
			}
			var result struct {
				Kind string `json:"kind"`
			}
			if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
				t.Fatalf("decode JSON error: %v\n%s", err, stdout.String())
			}
			if result.Kind != "Error" {
				t.Fatalf("JSON result kind = %q, want Error\n%s", result.Kind, stdout.String())
			}
		})
	}
}

func TestReverseApprovalInputRejectsBeforeOpeningLegacyProject(t *testing.T) {
	t.Run("reverse execution", func(t *testing.T) {
		root := setupReverseCLIProject(t)
		planID, _ := planReverseApprovalInputTest(t, root, "reverse-legacy-state")
		state := removeWorkspaceIdentity(t, root)

		var stdout, stderr bytes.Buffer
		code := runCLIWithApprovalStdin(t, []string{
			"reverse", "run", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		}, "", &stdout, &stderr)
		assertApprovalInputRefusedWithoutStateWrite(t, code, stdout.String()+stderr.String(), state, root)
	})

	t.Run("connector execution", func(t *testing.T) {
		root := t.TempDir()
		runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
		runCLIForReverseTest(t, []string{
			"credentials", "add", "github-local",
			"--connector", "github",
			"--config", "owner=acme",
			"--config", "repo=widgets",
			"--config", "public_access=true",
			"--root", root,
			"--json",
		})

		var planStdout, planStderr bytes.Buffer
		code := cli.Run([]string{
			"github", "issue", "close",
			"--issue-number", "101",
			"--credential", "github-local",
			"--root", root,
		}, &planStdout, &planStderr)
		if code != 0 {
			t.Fatalf("github issue close plan code = %d stdout=%s stderr=%s", code, planStdout.String(), planStderr.String())
		}
		planID := extractReverseField(t, planStdout.String(), `Created connector command plan (\S+)`)
		state := removeWorkspaceIdentity(t, root)

		var stdout, stderr bytes.Buffer
		code = runCLIWithApprovalStdin(t, []string{
			"github", "issue", "close",
			"--plan", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		}, "", &stdout, &stderr)
		assertApprovalInputRefusedWithoutStateWrite(t, code, stdout.String()+stderr.String(), state, root)
	})
}

func TestReverseApprovalReplayRejectsBeforeOpeningLegacyProject(t *testing.T) {
	t.Run("reverse execution", func(t *testing.T) {
		root := setupReverseCLIProject(t)
		planID, token := planReverseApprovalInputTest(t, root, "reverse-replay-state")

		var completedStdout, completedStderr bytes.Buffer
		code := runCLIWithApprovalStdin(t, []string{
			"reverse", "run", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		}, token+"\n", &completedStdout, &completedStderr)
		if code != 0 {
			t.Fatalf("reverse run code=%d stdout=%s stderr=%s", code, completedStdout.String(), completedStderr.String())
		}
		state := removeWorkspaceIdentity(t, root)

		var stdout, stderr bytes.Buffer
		code = runCLIWithApprovalStdin(t, []string{
			"reverse", "run", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		}, token+"\n", &stdout, &stderr)
		assertApprovalReplayRefusedWithoutStateWrite(t, code, stdout.String()+stderr.String(), state, root)
	})

	t.Run("connector execution", func(t *testing.T) {
		calls := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			if r.Method != http.MethodPatch || r.URL.Path != "/repos/acme/widgets/issues/101" {
				t.Errorf("request = %s %s, want PATCH /repos/acme/widgets/issues/101", r.Method, r.URL.Path)
				http.NotFound(w, r)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 101, "state": "closed"})
		}))
		defer server.Close()

		root := t.TempDir()
		runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
		runCLIForReverseTest(t, []string{
			"credentials", "add", "github-local",
			"--connector", "github",
			"--config", "owner=acme",
			"--config", "repo=widgets",
			"--config", "public_access=true",
			"--config", "base_url=" + server.URL,
			"--root", root,
			"--json",
		})

		var planStdout, planStderr bytes.Buffer
		code := cli.Run([]string{
			"github", "issue", "close",
			"--issue-number", "101",
			"--credential", "github-local",
			"--root", root,
		}, &planStdout, &planStderr)
		if code != 0 {
			t.Fatalf("github issue close plan code=%d stdout=%s stderr=%s", code, planStdout.String(), planStderr.String())
		}
		planID := extractReverseField(t, planStdout.String(), `Created connector command plan (\S+)`)
		token := extractReverseField(t, planStdout.String(), `Approval token: (\S+)`)

		var completedStdout, completedStderr bytes.Buffer
		code = runCLIWithApprovalStdin(t, []string{
			"github", "issue", "close",
			"--plan", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		}, token+"\n", &completedStdout, &completedStderr)
		if code != 0 {
			t.Fatalf("github issue close code=%d stdout=%s stderr=%s", code, completedStdout.String(), completedStderr.String())
		}
		if calls != 1 {
			t.Fatalf("github issue close calls=%d, want 1", calls)
		}
		state := removeWorkspaceIdentity(t, root)

		var stdout, stderr bytes.Buffer
		code = runCLIWithApprovalStdin(t, []string{
			"github", "issue", "close",
			"--plan", planID,
			"--approval-token-stdin",
			"--root", root,
			"--json",
		}, token+"\n", &stdout, &stderr)
		assertApprovalReplayRefusedWithoutStateWrite(t, code, stdout.String()+stderr.String(), state, root)
		if calls != 1 {
			t.Fatalf("replayed github issue close calls=%d, want 1", calls)
		}
	})
}

func planReverseApprovalInputTest(t *testing.T, root, name string) (string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"reverse", "plan", name,
		"--source-table", "sample_customers",
		"--destination", "outbox:outbox-local",
		"--map", "id:external_id",
		"--root", root,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("reverse plan code = %d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	return extractReverseField(t, stdout.String(), `Created reverse plan (\S+)`), extractReverseField(t, stdout.String(), `Approval token: (\S+)`)
}

func removeWorkspaceIdentity(t *testing.T, root string) []byte {
	t.Helper()
	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	contents, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state before legacy rewrite: %v", err)
	}
	var state map[string]any
	if err := json.Unmarshal(contents, &state); err != nil {
		t.Fatalf("decode state before legacy rewrite: %v", err)
	}
	delete(state, "workspace_id")
	legacyState, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("encode legacy state: %v", err)
	}
	if err := os.WriteFile(statePath, legacyState, 0o600); err != nil {
		t.Fatalf("write legacy state: %v", err)
	}
	return legacyState
}

func assertApprovalInputRefusedWithoutStateWrite(t *testing.T, code int, output string, wantState []byte, root string) {
	assertApprovalRefusedWithoutStateWrite(t, code, output, "approval token stdin must contain one bounded line", wantState, root)
}

func assertApprovalReplayRefusedWithoutStateWrite(t *testing.T, code int, output string, wantState []byte, root string) {
	assertApprovalRefusedWithoutStateWrite(t, code, output, "reverse plan approval has already been consumed", wantState, root)
}

func assertApprovalRefusedWithoutStateWrite(t *testing.T, code int, output, want string, wantState []byte, root string) {
	t.Helper()
	if code == 0 || !strings.Contains(output, want) {
		t.Fatalf("approval rejection code=%d output=%s", code, output)
	}
	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	gotState, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state after rejected approval: %v", err)
	}
	if !bytes.Equal(gotState, wantState) {
		t.Fatal("rejected approval opened and rewrote legacy project state")
	}
	if _, err := os.Stat(statePath + ".lock"); err == nil || !os.IsNotExist(err) {
		t.Fatalf("rejected approval created a state lock: %v", err)
	}
}

func TestReverseETLToGitHubCreatesPullRequestAfterApproval(t *testing.T) {
	type seenRequest struct {
		Method string
		Path   string
		Body   map[string]any
	}
	var seen []seenRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode GitHub request body: %v", err)
		}
		seen = append(seen, seenRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		switch r.URL.Path {
		case "/repos/acme/widgets/pulls":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 101, "html_url": "https://github.test/acme/widgets/pull/101"})
		case "/repos/acme/widgets/issues/101":
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 101})
		case "/repos/acme/widgets/pulls/101/requested_reviewers":
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"number": 101})
		default:
			t.Fatalf("unexpected GitHub path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	root := t.TempDir()
	t.Setenv("PM_GITHUB_TOKEN", "secret-token")
	runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
	runCLIForReverseTest(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=acme",
		"--config", "repo=widgets",
		"--config", "auth_type=token",
		"--config", "rate_limit_account=reverse-cli-test",
		"--config", "base_url=" + server.URL,
		"--from-env", "token=PM_GITHUB_TOKEN",
		"--root", root,
		"--json",
	})
	seedWarehouseTableFixture(t, root, "github_pr_candidates", map[string]any{
		"title":     "Ship connector writes",
		"body":      "Created by approved reverse ETL",
		"head":      "feature/github-writes",
		"base":      "main",
		"labels":    []any{"agentic", "reverse-etl"},
		"reviewers": []any{"ada", "grace"},
	})

	var planStdout, planStderr bytes.Buffer
	code := cli.Run([]string{
		"reverse", "plan", "github_prs",
		"--source-table", "github_pr_candidates",
		"--destination", "github:github-local",
		"--action", "create_pull_request",
		"--map", "title:title",
		"--map", "body:body",
		"--map", "head:head",
		"--map", "base:base",
		"--map", "labels:labels",
		"--map", "reviewers:reviewers",
		"--root", root,
	}, &planStdout, &planStderr)
	if code != 0 {
		t.Fatalf("reverse plan code = %d stderr = %s", code, planStderr.String())
	}
	planID := extractReverseField(t, planStdout.String(), `Created reverse plan (\S+)`)
	token := extractReverseField(t, planStdout.String(), `Approval token: (\S+)`)

	var runStdout, runStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{"reverse", "run", planID, "--approval-token-stdin", "--root", root, "--json"}, token+"\n", &runStdout, &runStderr)
	if code != 0 {
		t.Fatalf("reverse run code = %d stderr = %s stdout = %s", code, runStderr.String(), runStdout.String())
	}
	if len(seen) != 3 {
		t.Fatalf("GitHub request count = %d, want 3: %+v", len(seen), seen)
	}
	if seen[0].Method != http.MethodPost || seen[0].Path != "/repos/acme/widgets/pulls" {
		t.Fatalf("create PR request = %+v", seen[0])
	}
	if seen[0].Body["title"] != "Ship connector writes" || seen[0].Body["head"] != "feature/github-writes" || seen[0].Body["base"] != "main" {
		t.Fatalf("create PR body = %+v", seen[0].Body)
	}
	if seen[1].Method != http.MethodPatch || seen[1].Path != "/repos/acme/widgets/issues/101" {
		t.Fatalf("metadata request = %+v", seen[1])
	}
	if seen[2].Method != http.MethodPost || seen[2].Path != "/repos/acme/widgets/pulls/101/requested_reviewers" {
		t.Fatalf("reviewer request = %+v", seen[2])
	}
	if !strings.Contains(runStdout.String(), `"records_succeeded": 1`) {
		t.Fatalf("unexpected run output:\n%s", runStdout.String())
	}
}

func TestGitHubCommandWriteUsesReversePlanApproval(t *testing.T) {
	type seenRequest struct {
		Method string
		Path   string
		Body   map[string]any
	}
	var seen []seenRequest
	serverErrors := make(chan string, 1)
	reportServerError := func(format string, args ...any) {
		select {
		case serverErrors <- fmt.Sprintf(format, args...):
		default:
		}
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			reportServerError("decode GitHub request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		seen = append(seen, seenRequest{Method: r.Method, Path: r.URL.Path, Body: body})
		if r.Method != http.MethodPatch || r.URL.Path != "/repos/acme/widgets/issues/101" {
			reportServerError("unexpected GitHub request %s %s", r.Method, r.URL.Path)
			http.Error(w, "unexpected request", http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"number": 101, "state": "closed"})
	}))
	defer server.Close()
	assertNoServerError := func() {
		t.Helper()
		select {
		case msg := <-serverErrors:
			t.Fatal(msg)
		default:
		}
	}

	root := t.TempDir()
	runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
	runCLIForReverseTest(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=acme",
		"--config", "repo=widgets",
		"--config", "public_access=true",
		"--config", "base_url=" + server.URL,
		"--root", root,
		"--json",
	})

	var planStdout, planStderr bytes.Buffer
	code := cli.Run([]string{
		"github", "issue", "close",
		"--issue-number", "101",
		"--credential", "github-local",
		"--root", root,
	}, &planStdout, &planStderr)
	if code != 0 {
		t.Fatalf("github issue close plan code = %d stderr = %s stdout = %s", code, planStderr.String(), planStdout.String())
	}
	if len(seen) != 0 {
		t.Fatalf("plan made HTTP requests: %+v", seen)
	}
	planID := extractReverseField(t, planStdout.String(), `Created connector command plan (\S+)`)
	token := extractReverseField(t, planStdout.String(), `Approval token: (\S+)`)

	var previewStdout, previewStderr bytes.Buffer
	code = cli.Run([]string{
		"github", "issue", "close",
		"--plan", planID,
		"--preview",
		"--root", root,
		"--json",
	}, &previewStdout, &previewStderr)
	if code != 0 {
		t.Fatalf("github issue close preview code = %d stderr = %s stdout = %s", code, previewStderr.String(), previewStdout.String())
	}
	assertNoServerError()
	if len(seen) != 0 {
		t.Fatalf("preview made HTTP requests: %+v", seen)
	}
	for _, want := range []string{`"kind": "ConnectorCommandWritePreview"`, `"connector_command": "issue close"`, `"action": "close_issue"`} {
		if !strings.Contains(previewStdout.String(), want) {
			t.Fatalf("preview missing %q:\n%s", want, previewStdout.String())
		}
	}

	var wrongPathStdout, wrongPathStderr bytes.Buffer
	code = cli.Run([]string{
		"github", "issue", "create",
		"--plan", planID,
		"--preview",
		"--root", root,
		"--json",
	}, &wrongPathStdout, &wrongPathStderr)
	if code == 0 || !strings.Contains(wrongPathStdout.String()+wrongPathStderr.String(), "targets command") {
		t.Fatalf("wrong command path result code=%d stdout=%s stderr=%s", code, wrongPathStdout.String(), wrongPathStderr.String())
	}
	assertNoServerError()
	if len(seen) != 0 {
		t.Fatalf("wrong command path made HTTP requests: %+v", seen)
	}

	var deniedStdout, deniedStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "issue", "close",
		"--plan", planID,
		"--approval-token-stdin",
		"--root", root,
		"--json",
	}, "wrong-token\n", &deniedStdout, &deniedStderr)
	if code == 0 || !strings.Contains(deniedStderr.String(), "approval token is invalid") {
		t.Fatalf("bad approval result code=%d stdout=%s stderr=%s", code, deniedStdout.String(), deniedStderr.String())
	}
	assertNoServerError()
	if len(seen) != 0 {
		t.Fatalf("bad approval made HTTP requests: %+v", seen)
	}

	runCLIForReverseTest(t, []string{
		"credentials", "add", "outbox-local",
		"--connector", "outbox",
		"--config", "path=" + filepath.Join(root, ".polymetrics", "outbox"),
		"--root", root,
		"--json",
	})
	seedWarehouseTableFixture(t, root, "not_command", map[string]any{"id": "row-1"})
	var normalPlanStdout, normalPlanStderr bytes.Buffer
	code = cli.Run([]string{
		"reverse", "plan", "not_command",
		"--source-table", "not_command",
		"--destination", "outbox:outbox-local",
		"--map", "id:id",
		"--root", root,
	}, &normalPlanStdout, &normalPlanStderr)
	if code != 0 {
		t.Fatalf("normal reverse plan code=%d stdout=%s stderr=%s", code, normalPlanStdout.String(), normalPlanStderr.String())
	}
	normalPlanID := extractReverseField(t, normalPlanStdout.String(), `Created reverse plan (\S+)`)
	normalToken := extractReverseField(t, normalPlanStdout.String(), `Approval token: (\S+)`)
	var normalRunStdout, normalRunStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "issue", "close",
		"--plan", normalPlanID,
		"--approval-token-stdin",
		"--root", root,
		"--json",
	}, normalToken+"\n", &normalRunStdout, &normalRunStderr)
	if code == 0 || !strings.Contains(normalRunStdout.String()+normalRunStderr.String(), "not a connector command plan") {
		t.Fatalf("normal plan via provider command result code=%d stdout=%s stderr=%s", code, normalRunStdout.String(), normalRunStderr.String())
	}
	assertNoServerError()
	if len(seen) != 0 {
		t.Fatalf("normal reverse plan via provider command made HTTP requests: %+v", seen)
	}

	var runStdout, runStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "issue", "close",
		"--plan", planID,
		"--approval-token-stdin",
		"--root", root,
		"--json",
	}, token+"\n", &runStdout, &runStderr)
	if code != 0 {
		t.Fatalf("github issue close run code = %d stderr = %s stdout = %s", code, runStderr.String(), runStdout.String())
	}
	assertNoServerError()
	if len(seen) != 1 {
		t.Fatalf("run request count = %d, want 1: %+v", len(seen), seen)
	}
	if seen[0].Body["state"] != "closed" {
		t.Fatalf("run body = %+v, want state=closed", seen[0].Body)
	}
	if !strings.Contains(runStdout.String(), `"records_succeeded": 1`) {
		t.Fatalf("unexpected run output:\n%s", runStdout.String())
	}
}

func TestGitHubDestructiveCommandRequiresTypedConfirmation(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodDelete {
			t.Fatalf("request method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	root := t.TempDir()
	runCLIForReverseTest(t, []string{"init", "--root", root, "--json"})
	runCLIForReverseTest(t, []string{
		"credentials", "add", "github-local",
		"--connector", "github",
		"--config", "owner=acme",
		"--config", "repo=widgets",
		"--config", "public_access=true",
		"--config", "base_url=" + server.URL,
		"--root", root,
		"--json",
	})

	var planStdout, planStderr bytes.Buffer
	code := cli.Run([]string{
		"github", "repo", "deploy-key", "delete",
		"--key-id", "42",
		"--credential", "github-local",
		"--root", root,
	}, &planStdout, &planStderr)
	if code != 0 {
		t.Fatalf("github repo delete plan code=%d stdout=%s stderr=%s", code, planStdout.String(), planStderr.String())
	}
	planID := extractReverseField(t, planStdout.String(), `Created connector command plan (\S+)`)
	if !strings.Contains(planStdout.String(), "Preview required before an approval token is issued.") {
		t.Fatalf("destructive plan did not require preview before approval:\n%s", planStdout.String())
	}
	if calls != 0 {
		t.Fatalf("plan dispatched destructive request; calls=%d", calls)
	}

	var previewStdout, previewStderr bytes.Buffer
	code = cli.Run([]string{
		"github", "repo", "deploy-key", "delete",
		"--plan", planID,
		"--preview",
		"--root", root,
	}, &previewStdout, &previewStderr)
	if code != 0 {
		t.Fatalf("github repo delete preview code=%d stdout=%s stderr=%s", code, previewStdout.String(), previewStderr.String())
	}
	token := extractReverseField(t, previewStdout.String(), `Approval token: (\S+)`)
	if calls != 0 {
		t.Fatalf("preview dispatched destructive request; calls=%d", calls)
	}

	var deniedStdout, deniedStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "repo", "deploy-key", "delete",
		"--plan", planID,
		"--approval-token-stdin",
		"--root", root,
		"--json",
	}, token+"\n", &deniedStdout, &deniedStderr)
	if code == 0 || !strings.Contains(strings.ToLower(deniedStdout.String()+deniedStderr.String()), "confirmation") {
		t.Fatalf("missing confirmation result code=%d stdout=%s stderr=%s", code, deniedStdout.String(), deniedStderr.String())
	}
	if calls != 0 {
		t.Fatalf("missing confirmation dispatched destructive request; calls=%d", calls)
	}

	var runStdout, runStderr bytes.Buffer
	code = runCLIWithApprovalStdin(t, []string{
		"github", "repo", "deploy-key", "delete",
		"--plan", planID,
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root,
		"--json",
	}, token+"\n", &runStdout, &runStderr)
	if code != 0 {
		t.Fatalf("confirmed destructive run code=%d stdout=%s stderr=%s", code, runStdout.String(), runStderr.String())
	}
	if calls != 1 {
		t.Fatalf("confirmed destructive request count=%d, want 1", calls)
	}
}

func TestReverseETLRejectsPlannedCatalogDestination(t *testing.T) {
	root := setupReverseCLIProject(t)

	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{
		"credentials", "add", "postgres-native",
		"--connector", "destination-postgres",
		"--config", "mode=fixture",
		"--root", root,
		"--json",
	}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("credentials add destination-postgres code = 0, want planned connector rejection; stdout=%s", stdout.String())
	}
	if !strings.Contains(stdout.String()+stderr.String(), `connector "destination-postgres" uses a legacy source-/destination- prefix; use bare connector name "postgres"`) {
		t.Fatalf("planned destination rejection missing connector message: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
}

func TestReverseManualHasGithubCLIStyleDiscoverability(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"reverse"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("reverse manual code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"USAGE",
		"COMMANDS",
		"FLAGS",
		"EXAMPLES",
		"LEARN MORE",
		"pm reverse status <run-id>",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("reverse manual missing %q:\n%s", want, out)
		}
	}
}

func TestReverseManualExplainsConnectorCommandContentPolicy(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"reverse"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("reverse manual code = %d stderr = %s", code, stderr.String())
	}
	manual := stdout.String()
	normalizedManual := strings.Join(strings.Fields(manual), " ")
	for _, want := range []string{
		"The connector command runner does not mask",
		"This runner policy does not change source-table output or other execution paths.",
		"does not automatically retry a failed dispatch",
		"JSON plan and preview output omit tokens",
	} {
		if !strings.Contains(normalizedManual, want) {
			t.Fatalf("reverse manual missing %q:\n%s", want, manual)
		}
	}
	for _, obsolete := range []string{
		"Connector-command content is complete",
		"request, response, error, and preview content is complete",
		"mask those fields in plan samples",
		"connector-declared fields masked in sample rows",
		"masks connector-declared sensitive record fields",
	} {
		if strings.Contains(normalizedManual, obsolete) {
			t.Fatalf("reverse manual retained obsolete masking claim %q:\n%s", obsolete, manual)
		}
	}
}

func TestReverseManualDistinguishesEnvironmentOnlyWithheldFields(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := cli.Run([]string{"reverse"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("reverse manual code = %d stderr = %s", code, stderr.String())
	}
	out := stdout.String()
	for _, want := range []string{
		"--<withheld-flag> <value>",
		"--from-env <flag>=ENV",
		"A field declared env_only must instead",
		"without placing it in argv",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("reverse manual missing %q:\n%s", want, out)
		}
	}
}

func setupReverseCLIProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	t.Setenv("PM_SAMPLE_TOKEN", "sample-token")
	commands := [][]string{
		{"init", "--root", root, "--json"},
		{"credentials", "add", "sample-local", "--connector", "sample", "--from-env", "token=PM_SAMPLE_TOKEN", "--root", root, "--json"},
		{"credentials", "add", "warehouse-local", "--connector", "warehouse", "--config", "path=" + filepath.Join(root, ".polymetrics", "warehouse"), "--root", root, "--json"},
		{"credentials", "add", "outbox-local", "--connector", "outbox", "--config", "path=" + filepath.Join(root, ".polymetrics", "outbox"), "--root", root, "--json"},
		{"connections", "create", "sample_to_warehouse", "--source", "sample:sample-local", "--destination", "warehouse:warehouse-local", "--stream", "customers", "--primary-key", "id", "--cursor", "updated_at", "--table", "sample_customers", "--root", root, "--json"},
		{"etl", "run", "--connection", "sample_to_warehouse", "--stream", "customers", "--root", root, "--json"},
	}
	for _, args := range commands {
		var stdout, stderr bytes.Buffer
		code := cli.Run(args, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("setup command %v code = %d stderr = %s stdout = %s", args, code, stderr.String(), stdout.String())
		}
	}
	return root
}

func runCLIForReverseTest(t *testing.T, args []string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("command %v code = %d stderr = %s stdout = %s", args, code, stderr.String(), stdout.String())
	}
}

func runCLIWithApprovalStdin(t *testing.T, args []string, stdin string, stdout, stderr *bytes.Buffer) int {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create reverse approval stdin pipe: %v", err)
	}
	original := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = original
		_ = reader.Close()
	}()
	if _, err := io.WriteString(writer, stdin); err != nil {
		_ = writer.Close()
		t.Fatalf("write reverse approval stdin: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close reverse approval stdin: %v", err)
	}
	return cli.Run(args, stdout, stderr)
}

func extractReverseField(t *testing.T, text, pattern string) string {
	t.Helper()
	match := regexp.MustCompile(pattern).FindStringSubmatch(text)
	if len(match) != 2 {
		t.Fatalf("pattern %q not found in:\n%s", pattern, text)
	}
	return match[1]
}

// seedWarehouseTableFixture materializes an unattributed root-level warehouse
// table the way pm itself would, through the real Parquet writer. Hand-writing
// the file would put a fixture in a format the binary under test refuses, and
// would drift the moment the format changes again.
func seedWarehouseTableFixture(t *testing.T, root, table string, rows ...warehouse.Row) {
	t.Helper()
	path := filepath.Join(root, ".polymetrics", "warehouse", table+warehouse.TableFileExt)
	if err := warehouse.WriteTable(context.Background(), path, rows); err != nil {
		t.Fatalf("seed warehouse table %s: %v", table, err)
	}
}
