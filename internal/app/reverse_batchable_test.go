package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors/engine"
)

const humanProxyConnector = "humanproxy-demo"

type recordedWrite struct {
	Method string
	Path   string
	Body   string
}

// setupHumanProxyApp registers the humanproxy-demo fixture bundle — which
// declares one batchable:false action (cast_vote) and one batchable action
// (sync_profile) — into a real App registry, pointed at a real HTTP server.
//
// The fixture is loaded through engine.Load/engine.New and served by the same
// registry, manifest synthesis, and write path a shipped connector uses. It
// lives in testdata rather than internal/connectors/defs because bundles are
// embedded at compile time and this lane must not edit a connector bundle.
func setupHumanProxyApp(t *testing.T, ctx context.Context, baseURL string) (*app.App, string) {
	t.Helper()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() error = %v", err)
	}
	a, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	bundle, err := engine.Load(os.DirFS("testdata/bundles"), humanProxyConnector)
	if err != nil {
		t.Fatalf("engine.Load(%s) error = %v", humanProxyConnector, err)
	}
	a.Registry().Register(engine.New(bundle, nil))

	if _, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "humanproxy-local",
		Connector: humanProxyConnector,
		Config:    map[string]string{"base_url": baseURL},
	}); err != nil {
		t.Fatalf("AddCredential(%s) error = %v", humanProxyConnector, err)
	}
	return a, root
}

// rewriteStoredPlanAction edits the persisted plan's action in place, standing
// in for a plan created before the declaration existed or a hand-edited state
// file. The guard must not trust it.
func rewriteStoredPlanAction(t *testing.T, root, planID, action string) {
	t.Helper()
	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	raw, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatalf("read state: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("unmarshal state: %v", err)
	}
	plans, ok := doc["reverse_plans"].([]any)
	if !ok || len(plans) == 0 {
		t.Fatalf("state has no reverse_plans: %s", raw)
	}
	rewritten := false
	for _, entry := range plans {
		plan, ok := entry.(map[string]any)
		if !ok || plan["id"] != planID {
			continue
		}
		plan["action"] = action
		rewritten = true
	}
	if !rewritten {
		t.Fatalf("plan %q not found in state", planID)
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal state: %v", err)
	}
	if err := os.WriteFile(statePath, out, 0o600); err != nil {
		t.Fatalf("write state: %v", err)
	}
}

func recordingServer(t *testing.T) (*httptest.Server, func() []recordedWrite) {
	t.Helper()
	var mu sync.Mutex
	var writes []recordedWrite
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, r.ContentLength)
		if r.ContentLength > 0 {
			if _, err := r.Body.Read(body); err != nil && err.Error() != "EOF" {
				t.Errorf("read request body: %v", err)
			}
		}
		mu.Lock()
		writes = append(writes, recordedWrite{Method: r.Method, Path: r.URL.Path, Body: string(body)})
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	return server, func() []recordedWrite {
		mu.Lock()
		defer mu.Unlock()
		return append([]recordedWrite(nil), writes...)
	}
}

// R7/R8/R9: the refusal half. A non-batchable action must not reach a bulk
// reverse-ETL plan, must say why in terms the operator can act on, and must be
// detectable without string matching.
func TestPlanReverseETLRefusesNonBatchableAction(t *testing.T) {
	ctx := context.Background()
	server, recorded := recordingServer(t)
	defer server.Close()

	a, root := setupHumanProxyApp(t, ctx, server.URL)
	seedWarehouseTableRows(t, root, "saved_posts",
		`{"id":"t3_abc","dir":1}`,
		`{"id":"t3_def","dir":1}`,
	)

	_, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bulk_vote",
		SourceTable:           "saved_posts",
		DestinationConnector:  humanProxyConnector,
		DestinationCredential: "humanproxy-local",
		Action:                "cast_vote",
		Mappings:              map[string]string{"id": "id", "dir": "dir"},
	})
	if err == nil {
		t.Fatal("PlanReverseETL() planned a non-batchable action in bulk")
	}

	var nonBatchable *app.NonBatchableActionError
	if !errors.As(err, &nonBatchable) {
		t.Fatalf("PlanReverseETL() error = %T (%v), want *app.NonBatchableActionError", err, err)
	}
	if nonBatchable.Action != "cast_vote" || nonBatchable.Connector != humanProxyConnector {
		t.Fatalf("error fields = %+v, want action cast_vote on %s", nonBatchable, humanProxyConnector)
	}
	if nonBatchable.SourceTable != "saved_posts" {
		t.Fatalf("error SourceTable = %q, want saved_posts", nonBatchable.SourceTable)
	}

	message := err.Error()
	for _, want := range []string{"cast_vote", humanProxyConnector, "saved_posts", "non-batchable", "pm humanproxy-demo vote"} {
		if !strings.Contains(message, want) {
			t.Errorf("refusal %q does not mention %q; the message must name the action and the command to run instead", message, want)
		}
	}

	if plans := a.ListReversePlans(); len(plans) != 0 {
		t.Fatalf("ListReversePlans() = %d plans, want 0; a refused action must not leave an approvable plan behind", len(plans))
	}
	if calls := recorded(); len(calls) != 0 {
		t.Fatalf("destination received %d requests during planning, want 0", len(calls))
	}
}

// R10: defence in depth. A plan that predates the declaration — or one written
// straight into state.json — must still be refused at execute time, because the
// guard reads the live manifest rather than the stored plan.
func TestRunReverseETLRefusesStoredNonBatchablePlan(t *testing.T) {
	ctx := context.Background()
	server, recorded := recordingServer(t)
	defer server.Close()

	a, root := setupHumanProxyApp(t, ctx, server.URL)
	seedWarehouseTableRows(t, root, "profiles", `{"id":"u_1","display_name":"Ada"}`)

	// Plan a batchable action, then rewrite the stored plan's action to the
	// non-batchable one, standing in for a plan created before the declaration
	// existed or a hand-edited state file.
	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bulk_profile",
		SourceTable:           "profiles",
		DestinationConnector:  humanProxyConnector,
		DestinationCredential: "humanproxy-local",
		Action:                "sync_profile",
		Mappings:              map[string]string{"id": "id", "display_name": "display_name"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(sync_profile) error = %v", err)
	}
	rewriteStoredPlanAction(t, root, plan.ID, "cast_vote")

	reopened, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	bundle, err := engine.Load(os.DirFS("testdata/bundles"), humanProxyConnector)
	if err != nil {
		t.Fatalf("engine.Load error = %v", err)
	}
	reopened.Registry().Register(engine.New(bundle, nil))

	_, err = reopened.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err == nil {
		t.Fatal("RunReverseETL() executed a stored non-batchable bulk plan")
	}
	var nonBatchable *app.NonBatchableActionError
	if !errors.As(err, &nonBatchable) {
		t.Fatalf("RunReverseETL() error = %T (%v), want *app.NonBatchableActionError", err, err)
	}
	if calls := recorded(); len(calls) != 0 {
		t.Fatalf("destination received %d requests, want 0", len(calls))
	}
}

// R11: the half that makes the primitive worth having. The same batchable:false
// action must still run end to end when a human invokes it as its own command,
// and the request must actually arrive at the destination.
func TestNonBatchableActionStillExecutesAsConnectorCommand(t *testing.T) {
	ctx := context.Background()
	server, recorded := recordingServer(t)
	defer server.Close()

	a, _ := setupHumanProxyApp(t, ctx, server.URL)

	plan, preview, err := a.PlanConnectorCommand(ctx, app.PlanConnectorCommandRequest{
		Connector:  humanProxyConnector,
		Credential: "humanproxy-local",
		Path:       []string{"vote"},
		Flags:      map[string][]string{"id": {"t3_abc"}, "dir": {"1"}},
		Preview:    true,
	})
	if err != nil {
		t.Fatalf("PlanConnectorCommand(vote) error = %v", err)
	}
	if preview == nil {
		t.Fatal("PlanConnectorCommand(vote) preview = nil, want a typed preview")
	}
	if plan.RecordCount != 1 {
		t.Fatalf("plan RecordCount = %d, want 1", plan.RecordCount)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err != nil {
		t.Fatalf("RunReverseETL(vote) error = %v", err)
	}
	if run.Status != "completed" || run.RecordsSucceeded != 1 {
		t.Fatalf("run = %+v, want one completed write", run)
	}

	calls := recorded()
	if len(calls) != 1 {
		t.Fatalf("destination received %d requests, want exactly 1", len(calls))
	}
	if calls[0].Method != http.MethodPost || calls[0].Path != "/api/vote" {
		t.Fatalf("request = %s %s, want POST /api/vote", calls[0].Method, calls[0].Path)
	}
	if !strings.Contains(calls[0].Body, "id=t3_abc") {
		t.Fatalf("request body = %q, want the typed record fields", calls[0].Body)
	}
}

// R12: the declaration must be surgical. An action that does not declare it
// keeps planning and executing in bulk exactly as before.
func TestBatchableActionStillPlansAndExecutesInBulk(t *testing.T) {
	ctx := context.Background()
	server, recorded := recordingServer(t)
	defer server.Close()

	a, root := setupHumanProxyApp(t, ctx, server.URL)
	seedWarehouseTableRows(t, root, "profiles",
		`{"id":"u_1","display_name":"Ada"}`,
		`{"id":"u_2","display_name":"Grace"}`,
	)

	plan, err := a.PlanReverseETL(ctx, app.PlanReverseETLRequest{
		Name:                  "bulk_profile",
		SourceTable:           "profiles",
		DestinationConnector:  humanProxyConnector,
		DestinationCredential: "humanproxy-local",
		Action:                "sync_profile",
		Mappings:              map[string]string{"id": "id", "display_name": "display_name"},
	})
	if err != nil {
		t.Fatalf("PlanReverseETL(sync_profile) error = %v", err)
	}
	if plan.RecordCount != 2 {
		t.Fatalf("plan RecordCount = %d, want 2", plan.RecordCount)
	}

	run, err := a.RunReverseETL(ctx, app.RunReverseETLRequest{
		PlanID:        plan.ID,
		ApprovalToken: plan.ApprovalToken,
	})
	if err != nil {
		t.Fatalf("RunReverseETL(sync_profile) error = %v", err)
	}
	if run.Status != "completed" || run.RecordsSucceeded != 2 {
		t.Fatalf("run = %+v, want two completed writes", run)
	}
	if calls := recorded(); len(calls) != 2 {
		t.Fatalf("destination received %d requests, want 2", len(calls))
	}
}
