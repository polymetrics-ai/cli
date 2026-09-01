package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	deniedGitHubFlowTokenEnv = "PM_TEST_GITHUB_DENIED_TOKEN"
)

func assertNoCredentialMaterialInTree(t *testing.T, root, token string) {
	t.Helper()
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		payload, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if bytes.Contains(payload, []byte(token)) {
			return fmt.Errorf("credential material reached a persisted artifact")
		}
		return nil
	}); err != nil {
		t.Fatalf("persisted artifact credential audit: %v", err)
	}
}

type githubFlowRoundTripTarget struct {
	owner    string
	repo     string
	baseURL  string
	tokenEnv string
	token    string
	comments func() ([]string, error)
	provider string
}

type githubFlowRoundTripEvidence struct {
	Action                    string `json:"action"`
	Provider                  string `json:"provider"`
	FlowSyncRecords           int    `json:"flow_sync_records"`
	FlowActionRecords         int    `json:"flow_action_records"`
	ProviderCommentsAfterFlow int    `json:"provider_comments_after_flow"`
	WarehouseQueryRecords     int    `json:"warehouse_query_records"`
	CheckpointCommitted       bool   `json:"checkpoint_committed"`
	FlowReceiptPersisted      bool   `json:"flow_receipt_persisted"`
	ReplayRefused             bool   `json:"replay_refused"`
	ReplayRefusalType         string `json:"replay_refusal_type"`
	UnapprovedRefused         bool   `json:"unapproved_refused"`
	UnapprovedRefusalType     string `json:"unapproved_refusal_type"`
	AuthRefused               bool   `json:"auth_refused"`
	AuthRefusalType           string `json:"auth_refusal_type"`
	AuthCheckpointUnchanged   bool   `json:"auth_checkpoint_unchanged"`
	UnsafePreviewsOnly        bool   `json:"unsafe_previews_only"`
	ZeroProviderResidue       bool   `json:"zero_provider_residue"`
}

// TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip is the hermetic
// execution proof. It uses the real compiled
// GitHub definition, real durable warehouse, real reverse plan, and real flow
// engine; only GitHub's HTTPS boundary is replaced with a faithful local API.
func runFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip(t *testing.T) {
	const tokenEnv = "PM_TEST_GITHUB_LOCAL_FLOW_TOKEN"
	const token = "github-flow-local-canary"
	server, comments := newGitHubFlowRoundTripServer(t, token)
	t.Setenv(tokenEnv, token)
	target := githubFlowRoundTripTarget{
		owner:    "acme",
		repo:     "widgets",
		baseURL:  server.URL,
		tokenEnv: tokenEnv,
		token:    token,
		comments: comments,
		provider: "faithful_local_github_api",
	}
	evidence := executeFreshBinaryGitHubFlowRoundTrip(t, target)
	logGitHubFlowEvidence(t, evidence)
}

func executeFreshBinaryGitHubFlowRoundTrip(t *testing.T, target githubFlowRoundTripTarget) githubFlowRoundTripEvidence {
	t.Helper()
	binary := buildTransportPM(t)
	sha, size := transportBinaryIdentity(t, binary)
	t.Logf("fresh GitHub flow pm binary sha256=%s size_bytes=%d", sha, size)
	root := filepath.Join(t.TempDir(), "project")
	t.Setenv(deniedGitHubFlowTokenEnv, "deliberately-invalid-github-flow-token")

	credentialArgs := []string{
		"credentials", "add", "github-flow",
		"--connector", "github",
		"--config", "owner=" + target.owner,
		"--config", "repo=" + target.repo,
		"--config", "auth_type=token",
		"--config", "max_pages=1",
		"--config", "rate_limit_account=" + target.owner,
	}
	if target.baseURL != "" {
		credentialArgs = append(credentialArgs, "--config", "base_url="+target.baseURL)
	}
	credentialArgs = append(credentialArgs, "--from-env", "token="+target.tokenEnv, "--root", root, "--json")
	mustRunGitHubFlowPM(t, binary, "", []string{target.token}, "init", "--root", root, "--json")
	mustRunGitHubFlowPM(t, binary, "", []string{target.token}, credentialArgs...)
	mustRunGitHubFlowPM(t, binary, "", []string{target.token},
		"credentials", "add", "warehouse-flow", "--connector", "warehouse",
		"--config", "path="+filepath.Join(root, ".polymetrics", "warehouse"),
		"--root", root, "--json",
	)
	mustRunGitHubFlowPM(t, binary, "", []string{target.token},
		"connections", "create", "github-flow-job",
		"--source", "github:github-flow",
		"--destination", "warehouse:warehouse-flow",
		"--stream", "issues",
		"--sync-mode", "full_refresh_overwrite",
		"--primary-key", "node_id",
		"--cursor", "updated_at",
		"--table", "flow_issues",
		"--root", root, "--json",
	)
	// Bootstrap the durable source table so the approval authority can bind a
	// real reverse-ETL job before that job is composed into the flow.
	mustRunGitHubFlowPM(t, binary, "", []string{target.token},
		"etl", "run", "--connection", "github-flow-job", "--stream", "issues",
		"--root", root, "--json",
	)

	planOutput := mustRunGitHubFlowPM(t, binary, "", []string{target.token},
		"reverse", "plan", "flow_issue_comment_target",
		"--source-table", "flow_issues",
		"--connection", "github-flow-job",
		"--destination", "github:github-flow",
		"--action", "comment_issue",
		"--map", "number:issue_number",
		"--map", "title:body",
		"--root", root,
	)
	planID := githubFlowOutputField(t, planOutput.stdout, `Created reverse plan (\S+)`)
	approvalToken := githubFlowOutputField(t, planOutput.stdout, `Approval token: (\S+)`)
	mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"reverse", "preview", planID, "--root", root, "--json",
	)
	mustRunGitHubFlowPM(t, binary, approvalToken+"\n", []string{target.token, approvalToken},
		"reverse", "run", planID, "--approval-token-stdin", "--root", root, "--json",
	)
	flowPath := writeGitHubFlowManifest(t, root, planID)
	mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"flow", "plan", "--file", flowPath, "--root", root, "--json",
	)
	mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"flow", "preview", "--file", flowPath, "--root", root, "--json",
	)
	beforeFlow, err := target.comments()
	if err != nil {
		t.Fatalf("read provider comments before flow: %v", err)
	}
	flowOutput := mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"flow", "run", "--file", flowPath, "--root", root, "--json",
	)
	flowSyncRecords, flowActionRecords := assertGitHubFlowRunResult(t, flowOutput.stdout)
	afterFlow, err := target.comments()
	if err != nil {
		t.Fatalf("read provider comments after flow: %v", err)
	}
	if len(afterFlow) != len(beforeFlow)+1 {
		t.Fatalf("provider comments after composed flow = %d, want %d", len(afterFlow), len(beforeFlow)+1)
	}
	if afterFlow[len(afterFlow)-1] != "transport execution issue" {
		t.Fatalf("provider readback did not contain the issue title written by the flow")
	}

	queryOutput := mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"query", "run", "--table", "flow_issues", "--connection", "github-flow-job",
		"--limit", "10", "--root", root, "--json",
	)
	warehouseRecords := assertGitHubFlowWarehouseQuery(t, queryOutput.stdout)
	checkpointCommitted := assertGitHubFlowCheckpointCommitted(t, root)
	flowReceiptPersisted := assertGitHubFlowReceiptPersisted(t, root)

	providerCount := len(afterFlow)
	replay := runGitHubFlowPM(t, binary, approvalToken+"\n", []string{target.token, approvalToken},
		"reverse", "run", planID, "--approval-token-stdin", "--root", root, "--json",
	)
	replayRefusalType := assertGitHubFlowTypedRefusal(t, replay, "validation", "validation_error")
	assertGitHubFlowProviderCount(t, target.comments, providerCount, "approval replay")

	unapprovedPlan := mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"reverse", "plan", "unapproved_issue_comment_target",
		"--source-table", "flow_issues", "--connection", "github-flow-job",
		"--destination", "github:github-flow", "--action", "comment_issue",
		"--map", "number:issue_number", "--map", "title:body", "--root", root,
	)
	unapprovedPlanID := githubFlowOutputField(t, unapprovedPlan.stdout, `Created reverse plan (\S+)`)
	unapproved := runGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"reverse", "run", unapprovedPlanID, "--root", root, "--json",
	)
	unapprovedRefusalType := assertGitHubFlowTypedRefusal(t, unapproved, "usage", "usage_error")
	assertGitHubFlowProviderCount(t, target.comments, providerCount, "unapproved plan")

	deniedCredentialArgs := []string{
		"credentials", "add", "github-denied", "--connector", "github",
		"--config", "owner=" + target.owner, "--config", "repo=" + target.repo,
		"--config", "auth_type=token", "--config", "rate_limit_account=" + target.owner,
		"--from-env", "token=" + deniedGitHubFlowTokenEnv,
	}
	if target.baseURL != "" {
		deniedCredentialArgs = append(deniedCredentialArgs, "--config", "base_url="+target.baseURL)
	}
	deniedCredentialArgs = append(deniedCredentialArgs, "--root", root, "--json")
	mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken}, deniedCredentialArgs...)
	deniedPlan := mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken},
		"reverse", "plan", "denied_issue_comment_target",
		"--source-table", "flow_issues", "--connection", "github-flow-job",
		"--destination", "github:github-denied", "--action", "comment_issue",
		"--map", "number:issue_number", "--map", "title:body", "--root", root,
	)
	deniedPlanID := githubFlowOutputField(t, deniedPlan.stdout, `Created reverse plan (\S+)`)
	deniedToken := githubFlowOutputField(t, deniedPlan.stdout, `Approval token: (\S+)`)
	mustRunGitHubFlowPM(t, binary, "", []string{target.token, approvalToken, deniedToken},
		"reverse", "preview", deniedPlanID, "--root", root, "--json",
	)
	authCheckpointBefore := githubFlowCheckpointSnapshot(t, root)
	denied := runGitHubFlowPM(t, binary, deniedToken+"\n", []string{target.token, approvalToken, deniedToken},
		"reverse", "run", deniedPlanID, "--approval-token-stdin", "--root", root, "--json",
	)
	// A provider-verified 401 is invalid credential input, not an internal
	// program fault. This real binary path also proves the rejection leaves the
	// provider and durable checkpoint untouched.
	authRefusalType := assertGitHubFlowTerminalReverseRun(t, denied)
	assertGitHubFlowProviderCount(t, target.comments, providerCount, "provider authentication refusal")
	assertGitHubFlowCheckpointUnchanged(t, root, authCheckpointBefore, "provider authentication refusal")

	unsafePreviewsOnly := true
	for _, unsafe := range []struct {
		name    string
		action  string
		mapping []string
	}{
		{name: "merge_pull_request_preview", action: "merge_pull_request", mapping: []string{"number:pull_number"}},
		{name: "delete_file_preview", action: "delete_file", mapping: []string{"title:path", "state:message", "node_id:sha"}},
	} {
		assertGitHubFlowUnsafePreviewOnly(t, binary, root, target, unsafe.name, unsafe.action, unsafe.mapping, providerCount, approvalToken, deniedToken)
	}

	assertNoCredentialMaterialInTree(t, filepath.Join(root, ".polymetrics"), target.token)
	for _, credentialEnv := range []string{"PM_CERT_GITHUB_TOKEN", "PM_SCALE_GITHUB_TOKEN", "GITHUB_TOKEN"} {
		if credential := os.Getenv(credentialEnv); credential != "" && credential != target.token {
			assertNoCredentialMaterialInTree(t, filepath.Join(root, ".polymetrics"), credential)
		}
	}
	assertNoCredentialMaterialInTree(t, filepath.Join(root, ".polymetrics"), approvalToken)
	assertNoCredentialMaterialInTree(t, filepath.Join(root, ".polymetrics"), deniedToken)
	return githubFlowRoundTripEvidence{
		Action:                    "comment_issue",
		Provider:                  target.provider,
		FlowSyncRecords:           flowSyncRecords,
		FlowActionRecords:         flowActionRecords,
		ProviderCommentsAfterFlow: len(afterFlow),
		WarehouseQueryRecords:     warehouseRecords,
		CheckpointCommitted:       checkpointCommitted,
		FlowReceiptPersisted:      flowReceiptPersisted,
		ReplayRefused:             true,
		ReplayRefusalType:         replayRefusalType,
		UnapprovedRefused:         true,
		UnapprovedRefusalType:     unapprovedRefusalType,
		AuthRefused:               true,
		AuthRefusalType:           authRefusalType,
		AuthCheckpointUnchanged:   true,
		UnsafePreviewsOnly:        unsafePreviewsOnly,
	}
}

type githubFlowPMResult struct {
	stdout string
	stderr string
	code   int
}

func runGitHubFlowPM(t *testing.T, binary, stdin string, secrets []string, args ...string) githubFlowPMResult {
	t.Helper()
	command := exec.Command(binary, args...)
	command.Stdin = strings.NewReader(stdin)
	command.Env = os.Environ()
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	code := 0
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("start fresh pm command: %v", err)
		}
		code = exitErr.ExitCode()
	}
	for _, credentialEnv := range []string{"PM_CERT_GITHUB_TOKEN", "PM_SCALE_GITHUB_TOKEN", "GITHUB_TOKEN"} {
		if credential := os.Getenv(credentialEnv); credential != "" {
			secrets = append(secrets, credential)
		}
	}
	for _, secret := range secrets {
		if secret != "" && (strings.Contains(stdout.String(), secret) || strings.Contains(stderr.String(), secret)) {
			t.Fatal("fresh pm command exposed protected material in output")
		}
	}
	return githubFlowPMResult{stdout: stdout.String(), stderr: stderr.String(), code: code}
}

func mustRunGitHubFlowPM(t *testing.T, binary, stdin string, secrets []string, args ...string) githubFlowPMResult {
	t.Helper()
	result := runGitHubFlowPM(t, binary, stdin, secrets, args...)
	if result.code != 0 {
		t.Fatalf("fresh pm %s failed with exit %d (captured output withheld)", githubFlowCommandName(args), result.code)
	}
	return result
}

func githubFlowCommandName(args []string) string {
	visible := make([]string, 0, 3)
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") {
			break
		}
		visible = append(visible, arg)
		if len(visible) == 3 {
			break
		}
	}
	return strings.Join(visible, " ")
}

func githubFlowOutputField(t *testing.T, output, pattern string) string {
	t.Helper()
	match := regexp.MustCompile(pattern).FindStringSubmatch(output)
	if len(match) != 2 || strings.TrimSpace(match[1]) == "" {
		t.Fatal("fresh pm output omitted a required bounded field")
	}
	return match[1]
}

func writeGitHubFlowManifest(t *testing.T, root, jobReference string) string {
	t.Helper()
	dir := filepath.Join(root, ".polymetrics", "flows")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("create flow manifest directory: %v", err)
	}
	manifest := map[string]any{
		"version":     1,
		"name":        "github_warehouse_comment_roundtrip",
		"description": "GitHub issue to durable warehouse to GitHub issue comment",
		"steps": []any{
			map[string]any{
				"id": "extract_issues", "kind": "sync", "connection": "github-flow-job",
				"streams": []string{"issues"}, "in": []string{}, "out": []string{"flow_issues"},
			},
			map[string]any{
				"id": "comment_extracted_issue", "kind": "action", "in": []string{"flow_issues"}, "out": []string{},
				"job": jobReference,
				"action_cfg": map[string]any{
					"read_back_stream": "issue_comments",
				},
			},
		},
	}
	payload, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		t.Fatalf("encode flow manifest: %v", err)
	}
	path := filepath.Join(dir, "github_warehouse_comment_roundtrip.json")
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		t.Fatalf("write flow manifest: %v", err)
	}
	return path
}

func assertGitHubFlowRunResult(t *testing.T, payload string) (int, int) {
	t.Helper()
	var result struct {
		Status string `json:"status"`
		Steps  []struct {
			ID             string `json:"id"`
			Status         string `json:"status"`
			RecordsRead    int    `json:"records_read"`
			RecordsWritten int    `json:"records_written"`
		} `json:"steps"`
	}
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		t.Fatalf("decode flow result: %v", err)
	}
	if result.Status != "ok" || len(result.Steps) != 2 {
		t.Fatalf("composed flow status = %q steps=%d, want ok/2", result.Status, len(result.Steps))
	}
	if result.Steps[0].ID != "extract_issues" || result.Steps[0].Status != "ok" || result.Steps[0].RecordsWritten < 1 {
		t.Fatalf("flow sync step = id:%q status:%q read:%d written:%d, want extracted issue", result.Steps[0].ID, result.Steps[0].Status, result.Steps[0].RecordsRead, result.Steps[0].RecordsWritten)
	}
	if result.Steps[1].ID != "comment_extracted_issue" || result.Steps[1].Status != "ok" || result.Steps[1].RecordsWritten != 1 {
		t.Fatalf("flow action step did not acknowledge exactly one GitHub comment")
	}
	return result.Steps[0].RecordsWritten, result.Steps[1].RecordsWritten
}

func assertGitHubFlowWarehouseQuery(t *testing.T, payload string) int {
	t.Helper()
	var result struct {
		Kind  string           `json:"kind"`
		Rows  []map[string]any `json:"rows"`
		Count int              `json:"count"`
	}
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		t.Fatalf("decode warehouse query: %v", err)
	}
	if result.Kind != "QueryResult" || result.Count != 1 || len(result.Rows) != 1 {
		t.Fatalf("warehouse query count = %d rows=%d, want one", result.Count, len(result.Rows))
	}
	return result.Count
}

func assertGitHubFlowCheckpointCommitted(t *testing.T, root string) bool {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read durable state after flow: %v", err)
	}
	var state struct {
		StreamStates map[string]struct {
			Checkpoint json.RawMessage `json:"checkpoint"`
		} `json:"stream_states"`
	}
	if err := json.Unmarshal(payload, &state); err != nil {
		t.Fatalf("decode durable state after flow: %v", err)
	}
	for _, streamState := range state.StreamStates {
		if len(streamState.Checkpoint) > 0 && string(streamState.Checkpoint) != "null" {
			return true
		}
	}
	t.Fatal("freshly reopened project contains no committed stream checkpoint")
	return false
}

func githubFlowCheckpointSnapshot(t *testing.T, root string) []byte {
	t.Helper()
	var state struct {
		Checkpoints  map[string]map[string]string `json:"checkpoints"`
		StreamStates map[string]struct {
			Checkpoint json.RawMessage `json:"checkpoint"`
		} `json:"stream_states"`
	}
	readGitHubFlowState(t, root, &state)
	payload, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("encode durable checkpoint snapshot: %v", err)
	}
	return payload
}

func assertGitHubFlowCheckpointUnchanged(t *testing.T, root string, before []byte, stage string) {
	t.Helper()
	after := githubFlowCheckpointSnapshot(t, root)
	if !bytes.Equal(after, before) {
		t.Fatalf("durable checkpoint advanced after %s", stage)
	}
}

func assertGitHubFlowReceiptPersisted(t *testing.T, root string) bool {
	t.Helper()
	var state struct {
		FlowActionReceipts []struct {
			ID string `json:"id"`
		} `json:"flow_action_receipts"`
	}
	readGitHubFlowState(t, root, &state)
	if len(state.FlowActionReceipts) != 1 || strings.TrimSpace(state.FlowActionReceipts[0].ID) == "" {
		t.Fatalf("durable flow action receipt count = %d, want one", len(state.FlowActionReceipts))
	}
	return true
}

func readGitHubFlowState(t *testing.T, root string, destination any) {
	t.Helper()
	payload, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read durable flow state: %v", err)
	}
	if err := json.Unmarshal(payload, destination); err != nil {
		t.Fatalf("decode durable flow state: %v", err)
	}
}

func assertGitHubFlowTypedRefusal(t *testing.T, result githubFlowPMResult, wantCategory, wantCode string) string {
	t.Helper()
	if result.code == 0 {
		t.Fatal("refusal command unexpectedly succeeded")
	}
	var envelope struct {
		Kind  string `json:"kind"`
		Error struct {
			Category string `json:"category"`
			Code     string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &envelope); err != nil {
		t.Fatalf("decode typed refusal envelope: %v", err)
	}
	if envelope.Kind != "Error" || envelope.Error.Category == "" || envelope.Error.Code == "" {
		t.Fatalf("refusal did not return a typed Error envelope")
	}
	if wantCategory != "" && envelope.Error.Category != wantCategory {
		t.Fatalf("refusal category = %q, want %q", envelope.Error.Category, wantCategory)
	}
	if wantCode != "" && envelope.Error.Code != wantCode {
		t.Fatalf("refusal code = %q, want %q", envelope.Error.Code, wantCode)
	}
	return envelope.Error.Category + "/" + envelope.Error.Code
}

func assertGitHubFlowTerminalReverseRun(t *testing.T, result githubFlowPMResult) string {
	t.Helper()
	if result.code == 0 {
		t.Fatal("failed reverse run unexpectedly exited zero")
	}
	var envelope struct {
		Kind string `json:"kind"`
		Run  struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"run"`
	}
	if err := json.Unmarshal([]byte(result.stdout), &envelope); err != nil {
		t.Fatalf("decode terminal reverse-run envelope: %v", err)
	}
	if envelope.Kind != "ReverseRun" || envelope.Run.ID == "" || envelope.Run.Status != "failed" {
		t.Fatalf("terminal reverse-run envelope = %#v, want persisted failed run", envelope)
	}
	return envelope.Kind + "/" + envelope.Run.Status
}

func assertGitHubFlowProviderCount(t *testing.T, snapshot func() ([]string, error), want int, stage string) {
	t.Helper()
	comments, err := snapshot()
	if err != nil {
		t.Fatalf("read provider state after %s: %v", stage, err)
	}
	if len(comments) != want {
		t.Fatalf("provider mutations after %s = %d comments, want %d", stage, len(comments), want)
	}
}

func assertGitHubFlowUnsafePreviewOnly(t *testing.T, binary, root string, target githubFlowRoundTripTarget, name, action string, mappings []string, providerCount int, secrets ...string) {
	t.Helper()
	args := []string{
		"reverse", "plan", name, "--source-table", "flow_issues", "--connection", "github-flow-job",
		"--destination", "github:github-flow", "--action", action,
	}
	for _, mapping := range mappings {
		args = append(args, "--map", mapping)
	}
	args = append(args, "--root", root, "--json")
	plan := runGitHubFlowPM(t, binary, "", append([]string{target.token}, secrets...), args...)
	if plan.code == 0 {
		var envelope struct {
			Plan struct {
				ID string `json:"id"`
			} `json:"plan"`
		}
		if err := json.Unmarshal([]byte(plan.stdout), &envelope); err != nil || envelope.Plan.ID == "" {
			t.Fatal("unsafe action plan did not return a bounded plan id")
		}
		preview := runGitHubFlowPM(t, binary, "", append([]string{target.token}, secrets...),
			"reverse", "preview", envelope.Plan.ID, "--root", root, "--json",
		)
		if preview.code != 0 {
			assertGitHubFlowTypedRefusal(t, preview, "", "")
		}
	} else {
		assertGitHubFlowTypedRefusal(t, plan, "", "")
	}
	assertGitHubFlowProviderCount(t, target.comments, providerCount, action+" preview/refusal")
}

func logGitHubFlowEvidence(t *testing.T, evidence githubFlowRoundTripEvidence) {
	t.Helper()
	payload, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("encode GitHub flow evidence: %v", err)
	}
	t.Logf("github_flow_evidence=%s", payload)
}

func newGitHubFlowRoundTripServer(t *testing.T, token string) (*httptest.Server, func() ([]string, error)) {
	t.Helper()
	var mu sync.Mutex
	comments := []string{}
	now := time.Now().UTC().Format(time.RFC3339)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/widgets/issues":
			_ = json.NewEncoder(w).Encode([]map[string]any{{
				"id": 1, "node_id": "I_flow_4166", "number": 1,
				"title": "transport execution issue", "state": "open",
				"created_at": now, "updated_at": now, "labels": []any{},
			}})
		case request.Method == http.MethodPost && request.URL.Path == "/repos/acme/widgets/issues/1/comments":
			var body struct {
				Body string `json:"body"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil || body.Body == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mu.Lock()
			comments = append(comments, body.Body)
			id := len(comments)
			mu.Unlock()
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": id, "node_id": fmt.Sprintf("IC_%d", id), "body": body.Body, "created_at": now, "updated_at": now})
		case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/widgets/issues/comments":
			mu.Lock()
			rows := make([]map[string]any, 0, len(comments))
			for i, body := range comments {
				rows = append(rows, map[string]any{"id": i + 1, "node_id": fmt.Sprintf("IC_%d", i+1), "body": body, "created_at": now, "updated_at": now})
			}
			mu.Unlock()
			_ = json.NewEncoder(w).Encode(rows)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server, func() ([]string, error) {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), comments...), nil
	}
}
