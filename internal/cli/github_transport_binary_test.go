package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle is the executable
// proof for #4081. Every lifecycle boundary goes through a separately invoked
// fresh pm binary: a closed plan and human preview issue an ephemeral token;
// the ordinary ETL command receives that token on stdin and owns the durable
// source -> warehouse -> reopen -> provider -> read-back -> checkpoint path.
// Cleanup has its own closed plan, preview, and one-time approval.
func TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle(t *testing.T) {
	server := newFaithfulIssueLabelTransportServer(t)
	binary := buildTransportPM(t)
	if repeated := buildTransportPM(t); repeated != binary {
		t.Fatalf("package pm fixture path changed between callers: first=%q repeated=%q", binary, repeated)
	}
	sha, size := transportBinaryIdentity(t, binary)
	t.Logf("fresh pm binary sha256=%s size_bytes=%d", sha, size)
	root := filepath.Join(t.TempDir(), "project")

	var sanitizedOutputs []string
	mustRun := func(args ...string) string {
		t.Helper()
		output, err := runTransportPM(binary, "", args...)
		if err != nil {
			t.Fatalf("pm %s failed: %v\n%s", transportCommandName(args), err, output)
		}
		sanitizedOutputs = append(sanitizedOutputs, output)
		return output
	}

	mustRun("init", "--root", root, "--json")
	for _, credential := range []string{"github-source", "github-destination"} {
		mustRun(
			"credentials", "add", credential,
			"--connector", "github",
			"--config", "owner=acme",
			"--config", "repo=widgets",
			"--config", "public_access=true",
			"--config", "base_url="+server.URL,
			"--root", root,
			"--json",
		)
	}
	connectionOutput := mustRun(
		"connections", "create", "transport-demo",
		"--source", "github:github-source",
		"--destination", "github:github-destination",
		"--stream", "issues",
		"--sync-mode", "full_append",
		"--table", "issues",
		"--source-config", "transport_source_issue_number=4081001",
		"--destination-config", "transport_target_issue_number=4081002",
		"--destination-config", "transport_label=pm-transport-demo-4081",
		"--root", root,
		"--json",
	)
	connectionID := transportConnectionIDFromOutput(t, connectionOutput)

	forwardPlanOutput, err := runTransportPM(binary, "",
		"etl", "transport", "github-issue-label", "plan",
		"--connection", "transport-demo",
		"--root", root,
		"--json",
	)
	if err != nil {
		// This is deliberately the first unavailable carrier on the RED commit.
		t.Fatalf("closed GitHub transport plan command failed: %v\n%s", err, forwardPlanOutput)
	}
	sanitizedOutputs = append(sanitizedOutputs, forwardPlanOutput)
	forwardPlanID := assertTransportPlanOutput(t, forwardPlanOutput, "issue_label_transport", "add_issue_labels", connectionID, "")

	forwardPreviewJSON := mustRun(
		"etl", "transport", "github-issue-label", "preview", forwardPlanID,
		"--root", root,
		"--json",
	)
	assertTransportPreviewOutput(t, forwardPreviewJSON, "issue_label_transport", "add_issue_labels", false)
	forwardPreviewHuman, err := runTransportPM(binary, "",
		"etl", "transport", "github-issue-label", "preview", forwardPlanID,
		"--root", root,
	)
	if err != nil {
		t.Fatalf("human forward preview failed: %v", err)
	}
	forwardToken := assertTransportPreviewOutput(t, forwardPreviewHuman, "issue_label_transport", "add_issue_labels", true)

	forwardRunArgs := []string{
		"etl", "run",
		"--connection", "transport-demo",
		"--stream", "issues",
		"--batch-size", "1",
		"--approval-plan", forwardPlanID,
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root,
		"--json",
	}
	assertTransportTokenNotInInvocation(t, forwardToken, forwardRunArgs)
	forwardRunOutput, err := runTransportPM(binary, forwardToken+"\n", forwardRunArgs...)
	if err != nil {
		t.Fatalf("approved forward ETL run failed: %v\n%s", err, redactTransportFailureOutput(forwardRunOutput, forwardToken))
	}
	sanitizedOutputs = append(sanitizedOutputs, forwardRunOutput)
	forwardRunID := assertTransportRunOutput(t, forwardRunOutput, "transport-demo")
	assertTransportWarehouseArtifacts(t, root, connectionID)
	assertTransportCheckpoint(t, root, forwardRunID, "transport-demo")

	cleanupPlanOutput := mustRun(
		"etl", "transport", "github-issue-label", "cleanup", "plan",
		"--connection", "transport-demo",
		"--forward-plan", forwardPlanID,
		"--root", root,
		"--json",
	)
	cleanupPlanID := assertTransportPlanOutput(t, cleanupPlanOutput, "issue_label_transport_cleanup", "remove_issue_label", connectionID, forwardPlanID)
	cleanupPreviewJSON := mustRun(
		"etl", "transport", "github-issue-label", "preview", cleanupPlanID,
		"--root", root,
		"--json",
	)
	assertTransportPreviewOutput(t, cleanupPreviewJSON, "issue_label_transport_cleanup", "remove_issue_label", false)
	cleanupPreviewHuman, err := runTransportPM(binary, "",
		"etl", "transport", "github-issue-label", "preview", cleanupPlanID,
		"--root", root,
	)
	if err != nil {
		t.Fatalf("human cleanup preview failed: %v", err)
	}
	cleanupToken := assertTransportPreviewOutput(t, cleanupPreviewHuman, "issue_label_transport_cleanup", "remove_issue_label", true)

	cleanupRunArgs := []string{
		"etl", "transport", "github-issue-label", "cleanup", "run", cleanupPlanID,
		"--connection", "transport-demo",
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root,
		"--json",
	}
	assertTransportTokenNotInInvocation(t, cleanupToken, cleanupRunArgs)
	cleanupRunOutput, err := runTransportPM(binary, cleanupToken+"\n", cleanupRunArgs...)
	if err != nil {
		t.Fatalf("approved cleanup run failed: %v", err)
	}
	sanitizedOutputs = append(sanitizedOutputs, cleanupRunOutput)
	assertTransportCleanupRunOutput(t, cleanupRunOutput, cleanupPlanID, forwardPlanID, connectionID)

	// A new cleanup plan and grant are required to demonstrate GitHub's typed
	// missing_ok_status. Replaying the first grant must never reach DELETE.
	missingPlanOutput := mustRun(
		"etl", "transport", "github-issue-label", "cleanup", "plan",
		"--connection", "transport-demo",
		"--forward-plan", forwardPlanID,
		"--root", root,
		"--json",
	)
	missingPlanID := assertTransportPlanOutput(t, missingPlanOutput, "issue_label_transport_cleanup", "remove_issue_label", connectionID, forwardPlanID)
	missingPreviewHuman, err := runTransportPM(binary, "",
		"etl", "transport", "github-issue-label", "preview", missingPlanID,
		"--root", root,
	)
	if err != nil {
		t.Fatalf("human missing-label cleanup preview failed: %v", err)
	}
	missingToken := assertTransportPreviewOutput(t, missingPreviewHuman, "issue_label_transport_cleanup", "remove_issue_label", true)
	missingRunArgs := []string{
		"etl", "transport", "github-issue-label", "cleanup", "run", missingPlanID,
		"--connection", "transport-demo",
		"--approval-token-stdin",
		"--confirm", "destructive",
		"--root", root,
		"--json",
	}
	assertTransportTokenNotInInvocation(t, missingToken, missingRunArgs)
	missingRunOutput, err := runTransportPM(binary, missingToken+"\n", missingRunArgs...)
	if err != nil {
		t.Fatalf("missing-label cleanup run failed: %v", err)
	}
	sanitizedOutputs = append(sanitizedOutputs, missingRunOutput)
	assertTransportCleanupRunOutput(t, missingRunOutput, missingPlanID, forwardPlanID, connectionID)

	// The just-consumed grant is a replay attempt, not a second 404 proof.
	replayOutput, replayErr := runTransportPM(binary, missingToken+"\n", missingRunArgs...)
	if replayErr == nil {
		t.Fatal("replayed cleanup approval succeeded")
	}
	sanitizedOutputs = append(sanitizedOutputs, replayOutput)

	assertTransportTokensAreEphemeral(t, root, sanitizedOutputs, forwardToken, cleanupToken, missingToken)
	events, labelPresent := server.snapshot()
	if labelPresent {
		t.Fatalf("faithful GitHub server still has the transport label after cleanup; events=%v", events)
	}
	assertFaithfulIssueLabelTransportOrder(t, events)
	emitTransportLifecycleEvidence(t, sha, size, events)
}

// emitTransportLifecycleEvidence keeps the exact-binary proof consumable as a
// single sanitized JSON test-log record. It intentionally excludes approval
// tokens, credential configuration, request bodies, and temporary paths.
func emitTransportLifecycleEvidence(t *testing.T, binarySHA256 string, binarySize int64, events []string) {
	t.Helper()
	evidence := struct {
		Kind                        string   `json:"kind"`
		BinarySHA256                string   `json:"binary_sha256"`
		BinarySizeBytes             int64    `json:"binary_size_bytes"`
		ArtifactKinds               []string `json:"artifact_kinds"`
		SourceRecords               int      `json:"source_records"`
		ReopenedRecords             int      `json:"reopened_records"`
		IndependentReadBack         bool     `json:"independent_read_back"`
		CheckpointAfterAcknowledged bool     `json:"checkpoint_after_acknowledged"`
		CleanupStatuses             []int    `json:"cleanup_statuses"`
		ReplayRejected              bool     `json:"replay_rejected"`
		ZeroResidue                 bool     `json:"zero_residue"`
		ProviderEvents              []string `json:"provider_events"`
	}{
		Kind:                        "IssueLabelWarehouseTransportEvidence",
		BinarySHA256:                binarySHA256,
		BinarySizeBytes:             binarySize,
		ArtifactKinds:               []string{"wal_jsonl", "duckdb_parquet", "manifest"},
		SourceRecords:               1,
		ReopenedRecords:             1,
		IndependentReadBack:         true,
		CheckpointAfterAcknowledged: true,
		CleanupStatuses:             []int{http.StatusNoContent, http.StatusNotFound},
		ReplayRejected:              true,
		ZeroResidue:                 true,
		ProviderEvents:              append([]string(nil), events...),
	}
	encoded, err := json.Marshal(evidence)
	if err != nil {
		t.Fatalf("encode sanitized transport evidence: %v", err)
	}
	t.Logf("transport_evidence=%s", encoded)
}

var pmTestBinaryFixture struct {
	sync.Once
	path string
	dir  string
	err  error
}

func buildTransportPM(t *testing.T) string {
	t.Helper()
	pmTestBinaryFixture.Do(func() {
		dir, err := os.MkdirTemp("", "polymetrics-cli-pm-")
		if err != nil {
			pmTestBinaryFixture.err = fmt.Errorf("create temporary pm fixture directory: %w", err)
			return
		}
		pmTestBinaryFixture.dir = dir

		binary := filepath.Join(dir, "pm")
		command := exec.Command("go", "build", "-o", binary, "./cmd/pm")
		command.Dir = transportRepositoryRoot(t)
		if output, err := command.CombinedOutput(); err != nil {
			pmTestBinaryFixture.err = fmt.Errorf("build fresh pm fixture: %w\n%s", err, output)
			return
		}
		pmTestBinaryFixture.path = binary
	})
	if pmTestBinaryFixture.err != nil {
		t.Fatal(pmTestBinaryFixture.err)
	}
	return pmTestBinaryFixture.path
}

func removePMTestBinaryFixture() error {
	if pmTestBinaryFixture.dir == "" {
		return nil
	}
	return os.RemoveAll(pmTestBinaryFixture.dir)
}

func runTransportPM(binary, stdin string, args ...string) (string, error) {
	command := exec.Command(binary, args...)
	if stdin != "" {
		command.Stdin = strings.NewReader(stdin)
	}
	output, err := command.CombinedOutput()
	return string(output), err
}

func transportCommandName(args []string) string {
	if len(args) == 0 {
		return "pm"
	}
	return "pm " + strings.Join(args, " ")
}

func transportRepositoryRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository go.mod for fresh pm build")
		}
		dir = parent
	}
}

func transportBinaryIdentity(t *testing.T, path string) (string, int64) {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:]), info.Size()
}

type transportPlanEnvelope struct {
	APIVersion       string `json:"api_version"`
	Kind             string `json:"kind"`
	ApprovalRequired bool   `json:"approval_required"`
	Plan             struct {
		ID            string `json:"id"`
		Status        string `json:"status"`
		Mode          string `json:"mode"`
		ConnectionID  string `json:"connection_id"`
		Action        string `json:"action"`
		RecordCount   int    `json:"record_count"`
		ForwardPlanID string `json:"forward_plan_id"`
		Confirmation  struct {
			Kind string `json:"kind"`
		} `json:"confirmation"`
	} `json:"plan"`
}

func assertTransportPlanOutput(t *testing.T, output, wantMode, wantAction, wantConnectionID, wantForwardPlanID string) string {
	t.Helper()
	var envelope transportPlanEnvelope
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode transport plan JSON: %v", err)
	}
	if envelope.APIVersion != apiVersion || envelope.Kind != "ETLTransportPlan" || !envelope.ApprovalRequired {
		t.Fatalf("transport plan envelope is not the closed approval-required plan contract")
	}
	if envelope.Plan.ID == "" || envelope.Plan.Status != "planned" || envelope.Plan.Mode != wantMode || envelope.Plan.ConnectionID != wantConnectionID || envelope.Plan.Action != wantAction || envelope.Plan.RecordCount != 1 || envelope.Plan.ForwardPlanID != wantForwardPlanID || envelope.Plan.Confirmation.Kind != "destructive" {
		t.Fatalf("transport plan is not the expected closed one-record destructive plan")
	}
	return envelope.Plan.ID
}

type transportPreviewEnvelope struct {
	APIVersion          string `json:"api_version"`
	Kind                string `json:"kind"`
	ApprovalRequired    bool   `json:"approval_required"`
	ApprovalTokenIssued bool   `json:"approval_token_issued"`
	ApprovalToken       string `json:"approval_token"`
	ApprovalTokenHash   string `json:"approval_token_hash"`
	ApprovalGrant       any    `json:"approval_grant"`
	Plan                struct {
		Mode   string `json:"mode"`
		Action string `json:"action"`
	} `json:"plan"`
	WritePreview struct {
		RecordsStaged int    `json:"records_staged"`
		Action        string `json:"action"`
		Digest        string `json:"digest"`
	} `json:"write_preview"`
}

func assertTransportPreviewOutput(t *testing.T, output, wantMode, wantAction string, wantToken bool) string {
	t.Helper()
	if wantToken {
		const prefix = "Approval token: "
		if strings.Count(output, prefix) != 1 || !strings.Contains(output, "Confirmation required: --confirm destructive") {
			t.Fatal("human transport preview did not render exactly one destructive approval token")
		}
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, prefix) {
				token := strings.TrimSpace(strings.TrimPrefix(line, prefix))
				if token == "" {
					t.Fatal("human transport preview emitted an empty approval token")
				}
				return token
			}
		}
		t.Fatal("human transport preview omitted its approval token")
	}

	var envelope transportPreviewEnvelope
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode transport preview JSON: %v", err)
	}
	if envelope.APIVersion != apiVersion || envelope.Kind != "ETLTransportPreview" || !envelope.ApprovalRequired || !envelope.ApprovalTokenIssued || envelope.ApprovalToken != "" || envelope.ApprovalTokenHash != "" || envelope.ApprovalGrant != nil {
		t.Fatal("transport preview JSON did not keep approval material ephemeral")
	}
	if envelope.Plan.Mode != wantMode || envelope.Plan.Action != wantAction || envelope.WritePreview.RecordsStaged != 1 || envelope.WritePreview.Action != wantAction || envelope.WritePreview.Digest == "" {
		t.Fatal("transport preview JSON is not the expected closed write preview")
	}
	return ""
}

func transportConnectionIDFromOutput(t *testing.T, output string) string {
	t.Helper()
	var envelope struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		Connection struct {
			ID string `json:"id"`
		} `json:"connection"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode connection JSON: %v", err)
	}
	if envelope.APIVersion != apiVersion || envelope.Kind != "Connection" || envelope.Connection.ID == "" {
		t.Fatal("connection creation did not return an immutable connection ID")
	}
	return envelope.Connection.ID
}

func assertTransportRunOutput(t *testing.T, output, wantConnection string) string {
	t.Helper()
	var envelope struct {
		APIVersion      string `json:"api_version"`
		Kind            string `json:"kind"`
		RuntimeRecorded bool   `json:"runtime_recorded"`
		Run             struct {
			ID            string `json:"id"`
			Connection    string `json:"connection"`
			Stream        string `json:"stream"`
			Status        string `json:"status"`
			RecordsRead   int    `json:"records_read"`
			RecordsLoaded int    `json:"records_loaded"`
			RecordsFailed int    `json:"records_failed"`
			BatchCount    int    `json:"batch_count"`
		} `json:"run"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode forward ETL JSON: %v", err)
	}
	if envelope.APIVersion != apiVersion || envelope.Kind != "ETLRun" || envelope.RuntimeRecorded || envelope.Run.ID == "" || envelope.Run.Connection != wantConnection || envelope.Run.Stream != "issues" || envelope.Run.Status != "completed" || envelope.Run.RecordsRead != 1 || envelope.Run.RecordsLoaded != 1 || envelope.Run.RecordsFailed != 0 || envelope.Run.BatchCount != 1 {
		t.Fatal("forward ETL output is not the one-record completed closed transport run")
	}
	return envelope.Run.ID
}

func assertTransportCleanupRunOutput(t *testing.T, output, wantPlanID, wantForwardPlanID, wantConnectionID string) {
	t.Helper()
	var envelope struct {
		APIVersion    string `json:"api_version"`
		Kind          string `json:"kind"`
		Status        string `json:"status"`
		PlanID        string `json:"plan_id"`
		ForwardPlanID string `json:"forward_plan_id"`
		ConnectionID  string `json:"connection_id"`
		Action        string `json:"action"`
		Result        struct {
			RecordsWritten   int `json:"records_written"`
			RecordsFailed    int `json:"records_failed"`
			RecordsUnchanged int `json:"records_unchanged"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode cleanup JSON: %v", err)
	}
	if envelope.APIVersion != apiVersion || envelope.Kind != "ETLTransportCleanupRun" || envelope.Status != "completed" || envelope.PlanID != wantPlanID || envelope.ForwardPlanID != wantForwardPlanID || envelope.ConnectionID != wantConnectionID || envelope.Action != "remove_issue_label" || envelope.Result.RecordsWritten+envelope.Result.RecordsUnchanged != 1 || envelope.Result.RecordsFailed != 0 {
		t.Fatal("cleanup output is not the completed typed one-record inverse")
	}
}

func assertTransportTokenNotInInvocation(t *testing.T, token string, args []string) {
	t.Helper()
	if token == "" || strings.Contains(strings.Join(args, "\x00"), token) || strings.Contains(strings.Join(os.Environ(), "\x00"), token) {
		t.Fatal("approval token was present in an execution argv or environment")
	}
}

func assertTransportWarehouseArtifacts(t *testing.T, root, wantOwner string) {
	t.Helper()
	warehouseRoot := filepath.Join(root, ".polymetrics", "warehouse")
	var manifests, wals, parquets []string
	if err := filepath.WalkDir(warehouseRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel := filepath.ToSlash(path)
		switch {
		case strings.Contains(rel, "/transport/") && strings.HasSuffix(path, ".json"):
			manifests = append(manifests, path)
		case strings.Contains(rel, "/wal/") && strings.HasSuffix(path, ".jsonl") && strings.Contains(filepath.Base(path), "transport-"):
			wals = append(wals, path)
		case strings.Contains(rel, "/tables/") && strings.HasSuffix(path, ".parquet") && strings.Contains(filepath.Base(path), "transport-"):
			parquets = append(parquets, path)
		}
		return nil
	}); err != nil {
		t.Fatalf("walk transport warehouse artifacts: %v", err)
	}
	if len(manifests) != 1 || len(wals) != 1 || len(parquets) != 1 {
		t.Fatalf("transport artifacts manifests=%d wals=%d parquets=%d, want one durable artifact of each kind", len(manifests), len(wals), len(parquets))
	}
	for _, path := range append(append([]string{}, wals...), parquets...) {
		info, err := os.Stat(path)
		if err != nil || info.Size() == 0 {
			t.Fatalf("transport artifact %q is not a non-empty durable file", filepath.Base(path))
		}
	}
	contents, err := os.ReadFile(manifests[0])
	if err != nil {
		t.Fatalf("read transport manifest: %v", err)
	}
	var manifest struct {
		ID            string `json:"id"`
		Owner         string `json:"owner"`
		Generation    int64  `json:"generation"`
		Records       int    `json:"records"`
		WALSHA256     string `json:"wal_sha256"`
		ParquetSHA256 string `json:"parquet_sha256"`
		ContentSHA256 string `json:"content_sha256"`
	}
	if err := json.Unmarshal(contents, &manifest); err != nil {
		t.Fatalf("decode transport manifest: %v", err)
	}
	if manifest.ID == "" || manifest.Owner != wantOwner || manifest.Generation <= 0 || manifest.Records != 1 || manifest.WALSHA256 == "" || manifest.ParquetSHA256 == "" || manifest.ContentSHA256 == "" {
		t.Fatal("transport manifest is not a complete one-record connection-owned durable receipt")
	}
}

func assertTransportCheckpoint(t *testing.T, root, runID, wantConnection string) {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(root, ".polymetrics", "state", "state.json"))
	if err != nil {
		t.Fatalf("read persisted transport state: %v", err)
	}
	var state struct {
		Runs []struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"runs"`
		StreamStates map[string]struct {
			Connection string          `json:"connection"`
			Stream     string          `json:"stream"`
			Checkpoint json.RawMessage `json:"checkpoint"`
		} `json:"stream_states"`
	}
	if err := json.Unmarshal(contents, &state); err != nil {
		t.Fatalf("decode persisted transport state: %v", err)
	}
	completed := false
	for _, run := range state.Runs {
		if run.ID == runID && run.Status == "completed" {
			completed = true
			break
		}
	}
	if !completed {
		t.Fatal("forward ETL run was not durably completed")
	}
	for _, streamState := range state.StreamStates {
		if streamState.Connection != wantConnection || streamState.Stream != "issues" {
			continue
		}
		var checkpoint map[string]json.RawMessage
		if len(streamState.Checkpoint) == 0 || json.Unmarshal(streamState.Checkpoint, &checkpoint) != nil || len(checkpoint["committed_at"]) == 0 {
			t.Fatal("transport stream state did not persist its acknowledged checkpoint")
		}
		return
	}
	t.Fatal("transport stream state was not persisted for the completed connection")
}

func assertTransportTokensAreEphemeral(t *testing.T, root string, outputs []string, tokens ...string) {
	t.Helper()
	for index, output := range outputs {
		for _, token := range tokens {
			if token != "" && strings.Contains(output, token) {
				t.Fatalf("sanitized command output %d leaked an approval token", index)
			}
		}
	}
	projectDir := filepath.Join(root, ".polymetrics")
	if err := filepath.WalkDir(projectDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, token := range tokens {
			if token != "" && strings.Contains(string(contents), token) {
				return fmt.Errorf("approval token persisted")
			}
		}
		return nil
	}); err != nil {
		t.Fatal("approval token leaked into project artifacts")
	}
}

func redactTransportFailureOutput(output string, tokens ...string) string {
	for _, token := range tokens {
		if token != "" {
			output = strings.ReplaceAll(output, token, "[REDACTED]")
		}
	}
	return output
}

type faithfulIssueLabelTransportServer struct {
	*httptest.Server
	mu           sync.Mutex
	events       []string
	labelPresent bool
}

func newFaithfulIssueLabelTransportServer(t *testing.T) *faithfulIssueLabelTransportServer {
	t.Helper()
	server := &faithfulIssueLabelTransportServer{}
	server.Server = httptest.NewServer(http.HandlerFunc(server.serveHTTP))
	t.Cleanup(server.Close)
	return server
}

func (s *faithfulIssueLabelTransportServer) serveHTTP(w http.ResponseWriter, request *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if request.Header.Get("Accept") != "application/vnd.github+json" || request.Header.Get("X-GitHub-Api-Version") != "2026-03-10" || request.Header.Get("User-Agent") != "polymetrics-go-cli" || request.Header.Get("Authorization") != "" {
		http.Error(w, "request did not use the declared public GitHub transport contract", http.StatusBadRequest)
		return
	}

	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/widgets/issues":
		if request.URL.Query().Get("state") != "all" {
			http.Error(w, "issues read omitted declared state=all", http.StatusBadRequest)
			return
		}
		if request.URL.Query().Get("per_page") != "100" {
			http.Error(w, "issues read did not use the bounded declared page size", http.StatusBadRequest)
			return
		}
		if s.labelPresent {
			s.events = append(s.events, "GET:read-back:100")
		} else {
			s.events = append(s.events, "GET:source:100")
		}
		labels := []map[string]any{}
		if s.labelPresent {
			labels = append(labels, map[string]any{"name": "pm-transport-demo-4081"})
		}
		records := []map[string]any{faithfulIssue(4081001, nil), faithfulIssue(4081002, labels)}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(records)
	case request.Method == http.MethodPost && request.URL.Path == "/repos/acme/widgets/issues/4081002/labels":
		var body struct {
			Labels []string `json:"labels"`
		}
		if request.Header.Get("Content-Type") != "application/json" || json.NewDecoder(request.Body).Decode(&body) != nil || len(body.Labels) != 1 || body.Labels[0] != "pm-transport-demo-4081" || s.labelPresent {
			http.Error(w, "invalid typed add-issue-label payload", http.StatusBadRequest)
			return
		}
		s.labelPresent = true
		s.events = append(s.events, "POST")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{"name": "pm-transport-demo-4081"}})
	case request.Method == http.MethodDelete && request.URL.Path == "/repos/acme/widgets/issues/4081002/labels/pm-transport-demo-4081":
		if s.labelPresent {
			s.labelPresent = false
			s.events = append(s.events, "DELETE:204")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		s.events = append(s.events, "DELETE:404")
		w.WriteHeader(http.StatusNotFound)
	default:
		http.Error(w, fmt.Sprintf("unexpected faithful GitHub request %s %s", request.Method, request.URL.Path), http.StatusNotFound)
	}
}

func faithfulIssue(number int, labels []map[string]any) map[string]any {
	if labels == nil {
		labels = []map[string]any{}
	}
	return map[string]any{
		"id":         number,
		"node_id":    fmt.Sprintf("I_%d", number),
		"number":     number,
		"title":      "faithful transport fixture",
		"state":      "open",
		"labels":     labels,
		"created_at": "2026-08-13T00:00:00Z",
		"updated_at": "2026-08-13T00:00:00Z",
		"user": map[string]any{
			"login": "fixture",
			"id":    1,
		},
	}
}

func (s *faithfulIssueLabelTransportServer) snapshot() (events []string, labelPresent bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.events...), s.labelPresent
}

func assertFaithfulIssueLabelTransportOrder(t *testing.T, events []string) {
	t.Helper()
	want := []string{"GET:source:100", "POST", "GET:read-back:100", "DELETE:204", "DELETE:404"}
	if len(events) != len(want) {
		t.Fatalf("faithful GitHub lifecycle events = %v, want %v", events, want)
	}
	for index := range want {
		if events[index] != want[index] {
			t.Fatalf("faithful GitHub lifecycle events = %v, want %v", events, want)
		}
	}
}
