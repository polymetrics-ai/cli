package flow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// T-10 — Action manifest validation
// ---------------------------------------------------------------------------

func baseActionManifest() FlowManifest {
	return FlowManifest{
		Version: 1,
		Name:    "myflow",
		Steps: []FlowStep{
			{
				ID:   "send-emails",
				Kind: KindAction,
				ActionCfg: &ActionConfig{
					SourceTable:           "lead_outreach",
					DestinationConnector:  "sendgrid",
					DestinationCredential: "sg-prod",
					Action:                "create",
					Mappings:              map[string]string{"to": "email"},
				},
				In:  []string{},
				Out: []string{},
			},
		},
	}
}

func TestActionManifestValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*FlowManifest)
		wantErr bool
		errKind error
	}{
		{
			name:    "valid action step passes",
			mutate:  func(m *FlowManifest) {},
			wantErr: false,
		},
		{
			name: "missing source_table is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[0].ActionCfg.SourceTable = ""
			},
			wantErr: true,
			errKind: ErrManifestInvalid,
		},
		{
			name: "missing destination_connector is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[0].ActionCfg.DestinationConnector = ""
			},
			wantErr: true,
			errKind: ErrManifestInvalid,
		},
		{
			name: "missing mappings is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[0].ActionCfg.Mappings = nil
			},
			wantErr: true,
			errKind: ErrManifestInvalid,
		},
		{
			name: "missing action field defaults to upsert — no error",
			mutate: func(m *FlowManifest) {
				m.Steps[0].ActionCfg.Action = ""
			},
			wantErr: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := baseActionManifest()
			tc.mutate(&m)
			errs := ValidateManifest(m)
			if tc.wantErr {
				require.NotEmpty(t, errs)
				if tc.errKind != nil {
					assert.True(t, errors.Is(errs[0], tc.errKind))
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// helpers: fake destination httptest.Server
// ---------------------------------------------------------------------------

// fakeDestination tracks received records and can be configured to fail.
type fakeDestination struct {
	mu              sync.Mutex
	received        []map[string]any // all records received
	pmIDs           []string
	callCount       int32           // atomic
	fail429For      int             // first N calls return 429
	failPermanently map[string]bool // pm_ids that always 400
	returnExtID     bool
	srv             *httptest.Server
}

func newFakeDestination(t *testing.T) *fakeDestination {
	t.Helper()
	fd := &fakeDestination{
		failPermanently: map[string]bool{},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&fd.callCount, 1)
		if int(n) <= fd.fail429For {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		var records []map[string]any
		if err := json.NewDecoder(r.Body).Decode(&records); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		var accepted []string
		var failed []string
		for _, rec := range records {
			pmID, _ := rec["_pm_id"].(string)
			if fd.failPermanently[pmID] {
				failed = append(failed, pmID)
				continue
			}
			fd.mu.Lock()
			fd.received = append(fd.received, rec)
			fd.pmIDs = append(fd.pmIDs, pmID)
			fd.mu.Unlock()
			accepted = append(accepted, pmID)
		}
		resp := map[string]any{"accepted": accepted, "failed": failed}
		if fd.returnExtID && len(accepted) > 0 {
			resp["external_ids"] = map[string]string{accepted[0]: "ext-123"}
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})
	fd.srv = httptest.NewServer(mux)
	t.Cleanup(fd.srv.Close)
	return fd
}

func (fd *fakeDestination) ReceivedPMIDs() []string {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	cp := make([]string, len(fd.pmIDs))
	copy(cp, fd.pmIDs)
	return cp
}

func (fd *fakeDestination) ReceivedCount() int {
	fd.mu.Lock()
	defer fd.mu.Unlock()
	return len(fd.received)
}

// buildRunner creates an ActionRunner wired to the fake destination.
// stateDir is a temp dir for identity map and schema snapshots.
func buildRunner(t *testing.T, fd *fakeDestination, stateDir string, ledger LedgerAdapter) *HTTPActionRunner {
	t.Helper()
	return &HTTPActionRunner{
		FlowName:   "myflow",
		StepID:     "send-emails",
		StateDir:   stateDir,
		DLQDir:     filepath.Join(stateDir, "dlq"),
		Ledger:     ledger,
		MaxRetries: 3,
		BatchSize:  100,
		DestURL:    fd.srv.URL + "/write",
		HTTPClient: fd.srv.Client(),
		Sleep:      func(_ context.Context, _ time.Duration) error { return nil }, // instant for tests
	}
}

// sampleRecords returns n records with deterministic content.
func sampleRecords(n int) []map[string]any {
	records := make([]map[string]any, n)
	for i := range records {
		records[i] = map[string]any{
			"email":  fmt.Sprintf("user%d@example.com", i),
			"domain": "example.com",
			"name":   fmt.Sprintf("User %d", i),
		}
	}
	return records
}

// ---------------------------------------------------------------------------
// T-11 — Idempotent writes
// ---------------------------------------------------------------------------

func TestActionIdempotentWritesNoDuplicateOnReRun(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	led := &stubLedger{}
	runner := buildRunner(t, fd, stateDir, led)

	records := sampleRecords(5)

	// First run: all 5 should reach the server.
	res1, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	assert.Equal(t, 5, res1.RecordsSucceeded)
	assert.Equal(t, 5, fd.ReceivedCount())

	// Second run with same records: 0 should reach the server.
	res2, err := runner.Execute(context.Background(), records, "run-2")
	require.NoError(t, err)
	assert.Equal(t, 0, res2.RecordsSucceeded, "re-run should not duplicate records")
	assert.Equal(t, 5, fd.ReceivedCount(), "server must not receive duplicates")
}

func TestActionIdempotentPartialRerun(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	first3 := sampleRecords(3)
	all5 := sampleRecords(5)

	// Send first 3.
	_, err := runner.Execute(context.Background(), first3, "run-1")
	require.NoError(t, err)

	// Send all 5 — only 2 new should go through.
	res, err := runner.Execute(context.Background(), all5, "run-2")
	require.NoError(t, err)
	assert.Equal(t, 2, res.RecordsSucceeded, "only the 2 new records should be sent")
	assert.Equal(t, 5, fd.ReceivedCount())
}

// ---------------------------------------------------------------------------
// T-12 — Identity mapping
// ---------------------------------------------------------------------------

func TestActionIdentityMappingPersisted(t *testing.T) {
	fd := newFakeDestination(t)
	fd.returnExtID = true
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	records := sampleRecords(1)
	_, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)

	// Check identity map file exists and contains pm_id → ext_id.
	mapPath := filepath.Join(stateDir, "identity_map.json")
	data, err := os.ReadFile(mapPath)
	require.NoError(t, err)
	var idMap map[string]string
	require.NoError(t, json.Unmarshal(data, &idMap))
	assert.NotEmpty(t, idMap, "identity map must have at least one entry")
}

func TestActionIdentityMappingSkipsOnRerun(t *testing.T) {
	fd := newFakeDestination(t)
	fd.returnExtID = true
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	records := sampleRecords(1)
	_, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	firstCount := fd.ReceivedCount()

	// Re-run: identity map should skip.
	_, err = runner.Execute(context.Background(), records, "run-2")
	require.NoError(t, err)
	assert.Equal(t, firstCount, fd.ReceivedCount(), "identity-mapped record must not be re-sent")
}

// ---------------------------------------------------------------------------
// T-13 — Dedupe
// ---------------------------------------------------------------------------

func TestActionDedupeByEmail(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	// 3 records, but only 2 unique emails.
	records := []map[string]any{
		{"email": "alice@example.com", "name": "Alice v1"},
		{"email": "bob@example.com", "name": "Bob"},
		{"email": "alice@example.com", "name": "Alice v2"}, // duplicate email
	}
	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	assert.Equal(t, 2, res.RecordsSucceeded, "duplicates must be removed before write")
	assert.Equal(t, 2, fd.ReceivedCount())
}

func TestActionDedupeNoFalsePositives(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	records := sampleRecords(3) // all unique emails
	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	assert.Equal(t, 3, res.RecordsSucceeded)
}

// ---------------------------------------------------------------------------
// T-14 — Rate-limit handling (429 → backoff → success)
// ---------------------------------------------------------------------------

func TestAction429BackoffSuccess(t *testing.T) {
	fd := newFakeDestination(t)
	fd.fail429For = 2 // first 2 calls return 429
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	records := sampleRecords(1)
	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err, "should succeed after retries")
	assert.Equal(t, 1, res.RecordsSucceeded)
	assert.GreaterOrEqual(t, int(atomic.LoadInt32(&fd.callCount)), 3,
		"server must have been called at least 3 times (2 retries + 1 success)")
}

func TestAction429ExhaustedGoesToDLQ(t *testing.T) {
	fd := newFakeDestination(t)
	fd.fail429For = 999 // always 429
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})
	runner.MaxRetries = 2

	records := sampleRecords(1)
	res, err := runner.Execute(context.Background(), records, "run-1")
	// Should not return top-level error, but record goes to DLQ.
	require.NoError(t, err)
	assert.Equal(t, 0, res.RecordsSucceeded)
	assert.Equal(t, 1, res.RecordsFailed)
	assert.NotEmpty(t, res.DLQPath)
}

// ---------------------------------------------------------------------------
// T-15 — Dead-letter queue
// ---------------------------------------------------------------------------

func TestActionDLQOnPermanentFailure(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	records := sampleRecords(2)
	// Compute the pm_id of record 0 to configure permanent failure.
	pmID0 := deterministicRecordID("myflow", "send-emails", records[0])
	fd.failPermanently[pmID0] = true

	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err, "run should not abort; DLQ should be written")
	assert.Equal(t, 1, res.RecordsSucceeded)
	assert.Equal(t, 1, res.RecordsFailed)
	assert.NotEmpty(t, res.DLQPath)

	// DLQ file must exist and contain the failed pm_id.
	data, err := os.ReadFile(res.DLQPath)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), pmID0),
		"DLQ file must contain the failed pm_id")
}

// ---------------------------------------------------------------------------
// T-16 — Schema-drift detection
// ---------------------------------------------------------------------------

// schemaSnapshot stores a simple field→type map for test purposes.
func writeSchemaSnapshot(t *testing.T, stateDir, flowName, stepID string, fields map[string]string) {
	t.Helper()
	snap := schemaSnapshotForTest{
		Flow:   flowName,
		Step:   stepID,
		Fields: fields,
	}
	data, err := json.Marshal(snap)
	require.NoError(t, err)
	path := filepath.Join(stateDir, fmt.Sprintf("schema_snap_%s_%s.json", flowName, stepID))
	require.NoError(t, os.WriteFile(path, data, 0o600))
}

// schemaSnapshotForTest mirrors the internal SchemaSnapshot struct layout.
type schemaSnapshotForTest struct {
	Flow   string            `json:"flow"`
	Step   string            `json:"step"`
	Fields map[string]string `json:"fields"`
}

func TestActionSchemaDriftTypeChangeHalts(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	// Store snapshot saying "email" is string.
	writeSchemaSnapshot(t, stateDir, "myflow", "send-emails", map[string]string{
		"email": "string",
		"name":  "string",
	})

	// Live schema says "email" is integer (breaking).
	runner.LiveSchema = map[string]string{
		"email": "integer",
		"name":  "string",
	}

	records := sampleRecords(3)
	_, err := runner.Execute(context.Background(), records, "run-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSchemaDrift))
	assert.Equal(t, 0, fd.ReceivedCount(), "no records must be sent on schema drift")
}

func TestActionSchemaDriftFieldRemovedHalts(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	writeSchemaSnapshot(t, stateDir, "myflow", "send-emails", map[string]string{
		"email": "string",
		"name":  "string",
	})
	// "name" removed.
	runner.LiveSchema = map[string]string{"email": "string"}

	_, err := runner.Execute(context.Background(), sampleRecords(1), "run-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrSchemaDrift))
	assert.Equal(t, 0, fd.ReceivedCount())
}

func TestActionSchemaDriftAdditiveFieldNonBreaking(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	writeSchemaSnapshot(t, stateDir, "myflow", "send-emails", map[string]string{
		"email": "string",
	})
	// "name" is new — additive, non-breaking.
	runner.LiveSchema = map[string]string{
		"email": "string",
		"name":  "string",
	}

	records := sampleRecords(2)
	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	assert.Equal(t, 2, res.RecordsSucceeded)
}

func TestActionSchemaDriftNoDrift(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	runner := buildRunner(t, fd, stateDir, &stubLedger{})

	writeSchemaSnapshot(t, stateDir, "myflow", "send-emails", map[string]string{
		"email": "string",
		"name":  "string",
	})
	runner.LiveSchema = map[string]string{
		"email": "string",
		"name":  "string",
	}

	records := sampleRecords(2)
	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	assert.Equal(t, 2, res.RecordsSucceeded)
}

// ---------------------------------------------------------------------------
// T-17 — Receipts/audit
// ---------------------------------------------------------------------------

func TestActionReceiptWrittenToLedger(t *testing.T) {
	fd := newFakeDestination(t)
	stateDir := t.TempDir()
	led := &stubLedger{}
	runner := buildRunner(t, fd, stateDir, led)

	records := sampleRecords(3)
	res, err := runner.Execute(context.Background(), records, "run-1")
	require.NoError(t, err)
	assert.Equal(t, 3, res.RecordsSucceeded)

	entries := led.all()
	var hasReceipt bool
	for _, e := range entries {
		if e.Mode == "action" && e.Status == "receipt" {
			hasReceipt = true
			break
		}
	}
	assert.True(t, hasReceipt, "ledger must contain at least one action receipt entry")
}

// ---------------------------------------------------------------------------
// T-18 — Engine approval gate
// ---------------------------------------------------------------------------

// stubActionRunner is a test double for the ActionRunner used by the engine.
type stubActionRunner struct {
	mu       sync.Mutex
	executed []string // run IDs passed to Execute
}

func (s *stubActionRunner) ExecuteStep(ctx context.Context, step FlowStep, records []map[string]any, token, runID string) (ActionResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.executed = append(s.executed, runID)
	return ActionResult{RecordsSucceeded: len(records)}, nil
}

func actionManifestWithSyncAndAction() FlowManifest {
	return FlowManifest{
		Version: 1,
		Name:    "myflow",
		Steps: []FlowStep{
			{
				ID:         "sync-step",
				Kind:       KindSync,
				Connection: "conn-1",
				Streams:    []string{"users"},
				In:         []string{},
				Out:        []string{"users"},
			},
			{
				ID:   "send-emails",
				Kind: KindAction,
				ActionCfg: &ActionConfig{
					SourceTable:           "users",
					DestinationConnector:  "sendgrid",
					DestinationCredential: "sg-prod",
					Action:                "create",
					Mappings:              map[string]string{"to": "email"},
				},
				In:  []string{"users"},
				Out: []string{},
			},
		},
	}
}

func TestEngineActionRequiresApprovalToken(t *testing.T) {
	m := actionManifestWithSyncAndAction()
	appStub := &stubApp{}
	actionRunner := &stubActionRunner{}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}

	e := &Engine{
		Manifest:     m,
		App:          appStub,
		Ledger:       led,
		Checkpoint:   cs,
		LockDir:      t.TempDir(),
		ActionRunner: actionRunner,
	}

	_, err := e.Run(context.Background(), RunOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApprovalRequired),
		"expected ErrApprovalRequired, got %v", err)
	// ActionRunner must not have been called.
	assert.Empty(t, actionRunner.executed)
}

func TestEngineActionWithTokenExecutes(t *testing.T) {
	m := actionManifestWithSyncAndAction()
	appStub := &stubApp{}
	actionRunner := &stubActionRunner{}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}

	e := &Engine{
		Manifest:     m,
		App:          appStub,
		Ledger:       led,
		Checkpoint:   cs,
		LockDir:      t.TempDir(),
		ActionRunner: actionRunner,
	}

	_, err := e.Run(context.Background(), RunOptions{ApprovalToken: "valid-token"})
	require.NoError(t, err)
	assert.NotEmpty(t, actionRunner.executed, "ActionRunner must have been called with a token")
}

func TestEngineActionOnlyFlowNoTokenFails(t *testing.T) {
	m := FlowManifest{
		Version: 1,
		Name:    "action-only",
		Steps: []FlowStep{
			{
				ID:   "send",
				Kind: KindAction,
				ActionCfg: &ActionConfig{
					SourceTable:           "users",
					DestinationConnector:  "sendgrid",
					DestinationCredential: "sg-prod",
					Action:                "create",
					Mappings:              map[string]string{"to": "email"},
				},
				In:  []string{},
				Out: []string{},
			},
		},
	}
	actionRunner := &stubActionRunner{}
	e := &Engine{
		Manifest:     m,
		App:          &stubApp{},
		Ledger:       &stubLedger{},
		Checkpoint:   &FileCheckpointStore{Dir: t.TempDir()},
		LockDir:      t.TempDir(),
		ActionRunner: actionRunner,
	}

	_, err := e.Run(context.Background(), RunOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrApprovalRequired))
	assert.Empty(t, actionRunner.executed)
}

// ---------------------------------------------------------------------------
// Helpers used by T-16 that reference unexported types — forward declarations
// (these will resolve once action.go is written).
// ---------------------------------------------------------------------------

// deterministicRecordID is called in tests; it must be exported or the test must
// call the package-level function. We call the package-level unexported one here
// since the test is in the same package.
var _ = deterministicRecordID // ensure symbol is referenced
