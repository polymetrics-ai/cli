package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestPMBinaryExecutesGitHubWarehouseTransportDemo is the executable proof for
// #4081. It deliberately builds and invokes a fresh pm binary instead of
// reaching the App directly, so an approval token cannot cross a process
// boundary or be supplied as an argument.
func TestPMBinaryExecutesGitHubWarehouseTransportDemo(t *testing.T) {
	server := newFaithfulGitHubTransportServer(t)
	binary := buildTransportDemoPM(t)
	root := filepath.Join(t.TempDir(), "project")

	if output, err := runTransportDemoPM(binary, "init", "--root", root, "--json"); err != nil {
		t.Fatalf("pm init = %v\n%s", err, output)
	}
	output, err := runTransportDemoPM(binary,
		"demo", "github-warehouse-transport",
		"--github-base-url", server.URL,
		"--owner", "acme",
		"--repo", "widgets",
		"--source-issue", "4081001",
		"--target-issue", "4081002",
		"--root", root,
		"--json",
	)
	if err != nil {
		t.Fatalf("pm demo github-warehouse-transport = %v\n%s", err, output)
	}

	var result struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		Binary     struct {
			SHA256    string `json:"sha256"`
			SizeBytes int64  `json:"size_bytes"`
			Commit    string `json:"commit"`
		} `json:"binary"`
		Evidence struct {
			ConnectionID                 string   `json:"connection_id"`
			ForwardPlanID                string   `json:"forward_plan_id"`
			RunID                        string   `json:"run_id"`
			RecordsRead                  int      `json:"records_read"`
			RecordsStaged                int      `json:"records_staged"`
			RecordsApplied               int      `json:"records_applied"`
			IndependentReopenRecords     int      `json:"independent_reopen_records"`
			DestinationReadBackVerified  bool     `json:"destination_read_back_verified"`
			ReceiptBeforeCheckpointCAS   bool     `json:"receipt_before_checkpoint_cas"`
			CheckpointCommittedAt        string   `json:"checkpoint_committed_at"`
			CleanupPlanIDs               []string `json:"cleanup_plan_ids"`
			RepeatCleanupMissingAccepted bool     `json:"repeat_cleanup_missing_accepted"`
			ZeroResidueVerified          bool     `json:"zero_residue_verified"`
			Receipt                      struct {
				ID             string `json:"id"`
				Owner          string `json:"owner"`
				Generation     int64  `json:"generation"`
				ManifestSHA256 string `json:"manifest_sha256"`
				ContentSHA256  string `json:"content_sha256"`
				ParquetSHA256  string `json:"parquet_sha256"`
				Records        int    `json:"records"`
			} `json:"receipt"`
		} `json:"evidence"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("decode demo evidence: %v\n%s", err, output)
	}
	if result.APIVersion != apiVersion || result.Kind != "GitHubWarehouseTransportDemo" {
		t.Fatalf("demo envelope = %+v, want %s GitHubWarehouseTransportDemo", result, apiVersion)
	}
	sha, size := transportDemoBinaryIdentity(t, binary)
	if result.Binary.SHA256 != sha || result.Binary.SizeBytes != size || strings.TrimSpace(result.Binary.Commit) == "" {
		t.Fatalf("binary evidence = %+v, want sha=%q size=%d and a non-empty commit identity", result.Binary, sha, size)
	}
	if result.Evidence.ConnectionID == "" || result.Evidence.ForwardPlanID == "" || result.Evidence.RunID == "" {
		t.Fatalf("demo identity evidence = %+v, want connection, plan, and run IDs", result.Evidence)
	}
	if result.Evidence.RecordsRead != 1 || result.Evidence.RecordsStaged != 1 || result.Evidence.RecordsApplied != 1 || result.Evidence.IndependentReopenRecords != 1 {
		t.Fatalf("demo record evidence = %+v, want exactly one source/staged/applied/reopened record", result.Evidence)
	}
	if !result.Evidence.DestinationReadBackVerified || !result.Evidence.ReceiptBeforeCheckpointCAS || result.Evidence.CheckpointCommittedAt == "" {
		t.Fatalf("demo acknowledgement evidence = %+v, want read-back and receipt-before-CAS proof", result.Evidence)
	}
	if len(result.Evidence.CleanupPlanIDs) != 2 || !result.Evidence.RepeatCleanupMissingAccepted || !result.Evidence.ZeroResidueVerified {
		t.Fatalf("demo cleanup evidence = %+v, want two independently approved cleanups and zero residue", result.Evidence)
	}
	receipt := result.Evidence.Receipt
	if receipt.ID == "" || receipt.Owner != result.Evidence.ConnectionID || receipt.Generation <= 0 || receipt.ManifestSHA256 == "" || receipt.ContentSHA256 == "" || receipt.ParquetSHA256 == "" || receipt.Records != 1 {
		t.Fatalf("demo durable receipt = %+v, want a complete one-record connection-owned parquet receipt", receipt)
	}

	events, labelPresent, label := server.snapshot()
	if label == "" || strings.Contains(output, label) {
		t.Fatalf("demo output leaked the generated provider label: %q", output)
	}
	if strings.Contains(output, "approval_token") || strings.Contains(output, "acme") || strings.Contains(output, "widgets") {
		t.Fatalf("demo output contains unsanitized provider or approval data: %q", output)
	}
	if labelPresent {
		t.Fatalf("faithful GitHub server still has the demo label after cleanup; events=%v", events)
	}
	assertFaithfulGitHubTransportOrder(t, events)
}

func buildTransportDemoPM(t *testing.T) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "pm")
	command := exec.Command("go", "build", "-o", binary, "./cmd/pm")
	command.Dir = transportDemoRepositoryRoot(t)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("build fresh pm: %v\n%s", err, output)
	}
	return binary
}

func runTransportDemoPM(binary string, args ...string) (string, error) {
	command := exec.Command(binary, args...)
	output, err := command.CombinedOutput()
	return string(output), err
}

func transportDemoRepositoryRoot(t *testing.T) string {
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

func transportDemoBinaryIdentity(t *testing.T, path string) (string, int64) {
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

type faithfulGitHubTransportServer struct {
	*httptest.Server
	mu           sync.Mutex
	events       []string
	label        string
	labelPresent bool
}

func newFaithfulGitHubTransportServer(t *testing.T) *faithfulGitHubTransportServer {
	t.Helper()
	server := &faithfulGitHubTransportServer{}
	server.Server = httptest.NewServer(http.HandlerFunc(server.serveHTTP))
	t.Cleanup(server.Close)
	return server
}

func (s *faithfulGitHubTransportServer) serveHTTP(w http.ResponseWriter, request *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/repos/acme/widgets/issues":
		s.events = append(s.events, "GET")
		labels := []map[string]any{}
		if s.labelPresent {
			labels = append(labels, map[string]any{"name": s.label})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{
			faithfulGitHubIssue(4081001, nil),
			faithfulGitHubIssue(4081002, labels),
		})
	case request.Method == http.MethodPost && request.URL.Path == "/repos/acme/widgets/issues/4081002/labels":
		var body struct {
			Labels []string `json:"labels"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil || len(body.Labels) != 1 || !strings.HasPrefix(body.Labels[0], "pm-transport-demo-") {
			http.Error(w, "invalid typed label payload", http.StatusBadRequest)
			return
		}
		if s.label != "" && s.label != body.Labels[0] {
			http.Error(w, "unexpected second label", http.StatusConflict)
			return
		}
		s.label = body.Labels[0]
		s.labelPresent = true
		s.events = append(s.events, "POST")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{"name": s.label}})
	case request.Method == http.MethodDelete && s.label != "" && request.URL.Path == "/repos/acme/widgets/issues/4081002/labels/"+s.label:
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

func faithfulGitHubIssue(number int, labels []map[string]any) map[string]any {
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

func (s *faithfulGitHubTransportServer) snapshot() (events []string, labelPresent bool, label string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.events...), s.labelPresent, s.label
}

func assertFaithfulGitHubTransportOrder(t *testing.T, events []string) {
	t.Helper()
	post := -1
	firstDelete := -1
	secondDelete := -1
	for index, event := range events {
		switch event {
		case "POST":
			if post >= 0 {
				t.Fatalf("faithful GitHub received multiple forward mutations: %v", events)
			}
			post = index
		case "DELETE:204":
			firstDelete = index
		case "DELETE:404":
			secondDelete = index
		}
	}
	if len(events) == 0 || events[0] != "GET" || post < 1 || firstDelete <= post || secondDelete <= firstDelete {
		t.Fatalf("faithful GitHub lifecycle order = %v, want read -> one POST -> DELETE 204 -> DELETE 404", events)
	}
	readBack := false
	for _, event := range events[post+1 : firstDelete] {
		if event == "GET" {
			readBack = true
			break
		}
	}
	if !readBack {
		t.Fatalf("faithful GitHub saw no independent read-back between destination POST and cleanup: %v", events)
	}
}
