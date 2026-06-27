package flow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// T-01 — Manifest parse + validate
// ---------------------------------------------------------------------------

func validTwoStepManifestJSON() []byte {
	return []byte(`{
		"version": 1,
		"name": "likely-customers",
		"description": "Sync contacts then score them",
		"steps": [
			{
				"id": "sync-hubspot",
				"kind": "sync",
				"connection": "hubspot-prod",
				"streams": ["contacts", "companies"],
				"in": [],
				"out": ["contacts", "companies"]
			},
			{
				"id": "score-contacts",
				"kind": "query",
				"sql": "SELECT * FROM contacts WHERE email IS NOT NULL",
				"in": ["contacts"],
				"out": ["scored_contacts"]
			}
		]
	}`)
}

func TestManifestParse(t *testing.T) {
	t.Run("valid two-step manifest round-trips", func(t *testing.T) {
		m, err := ParseManifest(validTwoStepManifestJSON())
		require.NoError(t, err)
		assert.Equal(t, 1, m.Version)
		assert.Equal(t, "likely-customers", m.Name)
		assert.Len(t, m.Steps, 2)
		assert.Equal(t, "sync-hubspot", m.Steps[0].ID)
		assert.Equal(t, KindSync, m.Steps[0].Kind)
		assert.Equal(t, "score-contacts", m.Steps[1].ID)
		assert.Equal(t, KindQuery, m.Steps[1].Kind)
	})

	t.Run("malformed JSON returns error", func(t *testing.T) {
		_, err := ParseManifest([]byte(`{not valid json`))
		require.Error(t, err)
	})
}

func TestManifestValidate(t *testing.T) {
	base := func() FlowManifest {
		return FlowManifest{
			Version:     1,
			Name:        "likely-customers",
			Description: "test",
			Steps: []FlowStep{
				{
					ID:         "sync-hubspot",
					Kind:       KindSync,
					Connection: "hubspot-prod",
					Streams:    []string{"contacts", "companies"},
					In:         []string{},
					Out:        []string{"contacts", "companies"},
				},
				{
					ID:   "score-contacts",
					Kind: KindQuery,
					SQL:  "SELECT * FROM contacts WHERE email IS NOT NULL",
					In:   []string{"contacts"},
					Out:  []string{"scored_contacts"},
				},
			},
		}
	}

	tests := []struct {
		name    string
		mutate  func(*FlowManifest)
		wantErr bool
	}{
		{
			name:    "valid manifest produces no errors",
			mutate:  func(m *FlowManifest) {},
			wantErr: false,
		},
		{
			name:    "empty name is invalid",
			mutate:  func(m *FlowManifest) { m.Name = "" },
			wantErr: true,
		},
		{
			name:    "version 2 is invalid",
			mutate:  func(m *FlowManifest) { m.Version = 2 },
			wantErr: true,
		},
		{
			name: "unknown step kind is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[0].Kind = "unknown"
			},
			wantErr: true,
		},
		{
			name: "sync step missing connection is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[0].Connection = ""
			},
			wantErr: true,
		},
		{
			name: "sync step empty streams is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[0].Streams = []string{}
			},
			wantErr: true,
		},
		{
			name: "query step missing sql is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[1].SQL = ""
			},
			wantErr: true,
		},
		{
			name: "duplicate step IDs is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[1].ID = m.Steps[0].ID
			},
			wantErr: true,
		},
		{
			name: "in references table not in any out is invalid",
			mutate: func(m *FlowManifest) {
				m.Steps[1].In = []string{"nonexistent_table"}
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := base()
			tc.mutate(&m)
			errs := ValidateManifest(m)
			if tc.wantErr {
				require.NotEmpty(t, errs, "expected validation errors")
				for _, e := range errs {
					assert.True(t, errors.Is(e, ErrManifestInvalid),
						"expected ErrManifestInvalid, got %v", e)
				}
			} else {
				assert.Empty(t, errs, "expected no validation errors")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// T-02 — DAG build + topological sort + cycle detection
// ---------------------------------------------------------------------------

func makeManifestFromSteps(steps []FlowStep) FlowManifest {
	return FlowManifest{Version: 1, Name: "test-dag", Steps: steps}
}

func TestDAGLinearChain(t *testing.T) {
	// A produces "t1", B consumes "t1" and produces "t2", C consumes "t2"
	steps := []FlowStep{
		{ID: "A", Kind: KindSync, Connection: "c", Streams: []string{"s"}, Out: []string{"t1"}, In: []string{}},
		{ID: "B", Kind: KindQuery, SQL: "SELECT 1", In: []string{"t1"}, Out: []string{"t2"}},
		{ID: "C", Kind: KindQuery, SQL: "SELECT 1", In: []string{"t2"}, Out: []string{}},
	}
	order, err := BuildDAG(makeManifestFromSteps(steps))
	require.NoError(t, err)
	require.Equal(t, []string{"A", "B", "C"}, order)
}

func TestDAGIndependentSteps(t *testing.T) {
	steps := []FlowStep{
		{ID: "A", Kind: KindSync, Connection: "c", Streams: []string{"s"}, Out: []string{"t1"}, In: []string{}},
		{ID: "B", Kind: KindSync, Connection: "c", Streams: []string{"s"}, Out: []string{"t2"}, In: []string{}},
	}
	order, err := BuildDAG(makeManifestFromSteps(steps))
	require.NoError(t, err)
	assert.Len(t, order, 2)
	assert.Contains(t, order, "A")
	assert.Contains(t, order, "B")
}

func TestDAGTwoCycle(t *testing.T) {
	// A→B (B consumes A's out), B→A (A consumes B's out) — impossible but tests cycle detection
	steps := []FlowStep{
		{ID: "A", Kind: KindQuery, SQL: "SELECT 1", In: []string{"tb"}, Out: []string{"ta"}},
		{ID: "B", Kind: KindQuery, SQL: "SELECT 1", In: []string{"ta"}, Out: []string{"tb"}},
	}
	_, err := BuildDAG(makeManifestFromSteps(steps))
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCyclicDependency),
		"expected ErrCyclicDependency, got %v", err)
}

func TestDAGThreeNodeCycle(t *testing.T) {
	steps := []FlowStep{
		{ID: "A", Kind: KindQuery, SQL: "SELECT 1", In: []string{"tc"}, Out: []string{"ta"}},
		{ID: "B", Kind: KindQuery, SQL: "SELECT 1", In: []string{"ta"}, Out: []string{"tb"}},
		{ID: "C", Kind: KindQuery, SQL: "SELECT 1", In: []string{"tb"}, Out: []string{"tc"}},
	}
	_, err := BuildDAG(makeManifestFromSteps(steps))
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrCyclicDependency))
}

func TestDAGDiamond(t *testing.T) {
	// A→B, A→C, B→D, C→D
	steps := []FlowStep{
		{ID: "A", Kind: KindSync, Connection: "c", Streams: []string{"s"}, Out: []string{"ta"}, In: []string{}},
		{ID: "B", Kind: KindQuery, SQL: "SELECT 1", In: []string{"ta"}, Out: []string{"tb"}},
		{ID: "C", Kind: KindQuery, SQL: "SELECT 1", In: []string{"ta"}, Out: []string{"tc"}},
		{ID: "D", Kind: KindQuery, SQL: "SELECT 1", In: []string{"tb", "tc"}, Out: []string{}},
	}
	order, err := BuildDAG(makeManifestFromSteps(steps))
	require.NoError(t, err)
	require.Len(t, order, 4)
	assert.Equal(t, "A", order[0], "A must be first")
	assert.Equal(t, "D", order[3], "D must be last")
}

// ---------------------------------------------------------------------------
// T-03 — Checkpoint store
// ---------------------------------------------------------------------------

func newFileCheckpoint(t *testing.T) *FileCheckpointStore {
	t.Helper()
	return &FileCheckpointStore{Dir: t.TempDir()}
}

func TestCheckpointSetGet(t *testing.T) {
	cs := newFileCheckpoint(t)
	require.NoError(t, cs.Set("myflow", "step-a", "success"))
	got, err := cs.Get("myflow", "step-a")
	require.NoError(t, err)
	assert.Equal(t, "success", got)
}

func TestCheckpointGetUnknown(t *testing.T) {
	cs := newFileCheckpoint(t)
	got, err := cs.Get("myflow", "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestCheckpointClearFlow(t *testing.T) {
	cs := newFileCheckpoint(t)
	require.NoError(t, cs.Set("myflow", "step-a", "success"))
	require.NoError(t, cs.Set("myflow", "step-b", "success"))
	require.NoError(t, cs.Clear("myflow"))

	got, err := cs.Get("myflow", "step-a")
	require.NoError(t, err)
	assert.Equal(t, "", got)
}

func TestCheckpointClearDoesNotAffectOtherFlow(t *testing.T) {
	cs := newFileCheckpoint(t)
	require.NoError(t, cs.Set("flow-one", "step-a", "success"))
	require.NoError(t, cs.Set("flow-two", "step-a", "success"))
	require.NoError(t, cs.Clear("flow-one"))

	got, err := cs.Get("flow-two", "step-a")
	require.NoError(t, err)
	assert.Equal(t, "success", got, "flow-two should be untouched")
}

func TestCheckpointConcurrentSets(t *testing.T) {
	cs := newFileCheckpoint(t)
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			stepID := fmt.Sprintf("step-%d", i)
			_ = cs.Set("myflow", stepID, "success")
		}(i)
	}
	wg.Wait()
	// Just assert no panics + all steps present
	for i := 0; i < 10; i++ {
		got, err := cs.Get("myflow", fmt.Sprintf("step-%d", i))
		require.NoError(t, err)
		assert.Equal(t, "success", got)
	}
}

// ---------------------------------------------------------------------------
// T-04 — Engine lease contention
// ---------------------------------------------------------------------------

// stubApp is a test double for AppAdapter.
type stubApp struct {
	mu      sync.Mutex
	calls   []string
	results map[string]error
}

func (s *stubApp) ETLRun(_ context.Context, connectionID string, _ []string) (ETLResult, error) {
	s.mu.Lock()
	s.calls = append(s.calls, connectionID)
	s.mu.Unlock()
	if s.results != nil {
		if err, ok := s.results[connectionID]; ok {
			return ETLResult{}, err
		}
	}
	return ETLResult{RecordsRead: 10, RecordsWritten: 10}, nil
}

func (s *stubApp) QuerySQL(_ context.Context, sql string, _ int) ([]map[string]any, error) {
	s.mu.Lock()
	s.calls = append(s.calls, sql)
	s.mu.Unlock()
	return nil, nil
}

// stubLedger records all appended LedgerRecords in memory.
type stubLedger struct {
	mu      sync.Mutex
	records []LedgerRecord
}

func (l *stubLedger) Append(_ context.Context, r LedgerRecord) error {
	l.mu.Lock()
	l.records = append(l.records, r)
	l.mu.Unlock()
	return nil
}

func (l *stubLedger) all() []LedgerRecord {
	l.mu.Lock()
	defer l.mu.Unlock()
	cp := make([]LedgerRecord, len(l.records))
	copy(cp, l.records)
	return cp
}

func newEngineForTest(t *testing.T, m FlowManifest, app AppAdapter, ledger LedgerAdapter, cs CheckpointStore) *Engine {
	t.Helper()
	return &Engine{
		Manifest:   m,
		App:        app,
		Ledger:     ledger,
		Checkpoint: cs,
		LockDir:    t.TempDir(),
	}
}

func TestEngineLockHeldReturnsErrLeaseHeld(t *testing.T) {
	dir := t.TempDir()
	m, _ := ParseManifest(validTwoStepManifestJSON())
	// Pre-create the lock file so Engine sees it as held
	lockPath := filepath.Join(dir, "flow-likely-customers.lock")
	require.NoError(t, os.WriteFile(lockPath, []byte("99999\n"), 0o600))

	app := &stubApp{}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := &Engine{
		Manifest:   m,
		App:        app,
		Ledger:     led,
		Checkpoint: cs,
		LockDir:    dir,
	}
	_, err := e.Run(context.Background(), RunOptions{})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrLeaseHeld), "expected ErrLeaseHeld, got %v", err)
}

func TestEngineLockReleasedAfterRun(t *testing.T) {
	m, _ := ParseManifest(validTwoStepManifestJSON())
	app := &stubApp{}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, m, app, led, cs)

	_, err := e.Run(context.Background(), RunOptions{})
	// Implementation is "not implemented" — just assert lock file is gone after call
	// When implemented, this should return no error
	lockPath := filepath.Join(e.LockDir, "flow-likely-customers.lock")
	_, statErr := os.Stat(lockPath)
	assert.True(t, os.IsNotExist(statErr), "lock file should be removed after Run; err=%v runErr=%v", statErr, err)
}

// ---------------------------------------------------------------------------
// T-05 — Engine dependency ordering
// ---------------------------------------------------------------------------

// stepCallTracker is a stubApp that records step execution order by step ID.
type stepCallTracker struct {
	mu    sync.Mutex
	order []string
	fail  map[string]error // stepID -> error to return
}

func (s *stepCallTracker) ETLRun(_ context.Context, connectionID string, _ []string) (ETLResult, error) {
	s.mu.Lock()
	s.order = append(s.order, connectionID)
	err := s.fail[connectionID]
	s.mu.Unlock()
	return ETLResult{RecordsRead: 1, RecordsWritten: 1}, err
}

func (s *stepCallTracker) QuerySQL(_ context.Context, sql string, _ int) ([]map[string]any, error) {
	s.mu.Lock()
	s.order = append(s.order, sql)
	err := s.fail[sql]
	s.mu.Unlock()
	return nil, err
}

func twoStepManifest() FlowManifest {
	return FlowManifest{
		Version: 1,
		Name:    "likely-customers",
		Steps: []FlowStep{
			{
				ID:         "sync-hubspot",
				Kind:       KindSync,
				Connection: "hubspot-prod",
				Streams:    []string{"contacts", "companies"},
				In:         []string{},
				Out:        []string{"contacts", "companies"},
			},
			{
				ID:   "score-contacts",
				Kind: KindQuery,
				SQL:  "SELECT * FROM contacts WHERE email IS NOT NULL",
				In:   []string{"contacts"},
				Out:  []string{"scored_contacts"},
			},
		},
	}
}

func TestEngineDependencyOrderABeforeB(t *testing.T) {
	tracker := &stepCallTracker{}
	// Use connection ID and sql as traceable keys
	// sync-hubspot uses connection "hubspot-prod", score-contacts uses SQL text
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, twoStepManifest(), tracker, led, cs)

	result, err := e.Run(context.Background(), RunOptions{})
	require.NoError(t, err)
	require.Equal(t, "ok", result.Status)
	// Both steps should have been called; sync-hubspot (via connection) before score-contacts (via sql)
	require.Len(t, tracker.order, 2)
	assert.Equal(t, "hubspot-prod", tracker.order[0])
}

func TestEngineStepAFailsStepBNotCalled(t *testing.T) {
	tracker := &stepCallTracker{
		fail: map[string]error{"hubspot-prod": errors.New("sync failed")},
	}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, twoStepManifest(), tracker, led, cs)

	result, err := e.Run(context.Background(), RunOptions{})
	require.Error(t, err)
	assert.Equal(t, "failed", result.Status)
	// score-contacts (the SQL step) must NOT have been called
	for _, call := range tracker.order {
		assert.NotEqual(t, "SELECT * FROM contacts WHERE email IS NOT NULL", call,
			"score-contacts should not have been called after sync failure")
	}
}

func TestEngineThreeStepsMiddleFailsThirdSkipped(t *testing.T) {
	// Build: A sync→ (out: t1), B query consuming t1 (out: t2) — fails, C query consuming t2
	m := FlowManifest{
		Version: 1,
		Name:    "three-step",
		Steps: []FlowStep{
			{ID: "A", Kind: KindSync, Connection: "conn-a", Streams: []string{"s"}, Out: []string{"t1"}, In: []string{}},
			{ID: "B", Kind: KindQuery, SQL: "SELECT-B", In: []string{"t1"}, Out: []string{"t2"}},
			{ID: "C", Kind: KindQuery, SQL: "SELECT-C", In: []string{"t2"}, Out: []string{}},
		},
	}
	tracker := &stepCallTracker{
		fail: map[string]error{"SELECT-B": errors.New("query failed")},
	}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, m, tracker, led, cs)

	result, _ := e.Run(context.Background(), RunOptions{})
	assert.Equal(t, "failed", result.Status)
	for _, call := range tracker.order {
		assert.NotEqual(t, "SELECT-C", call, "C must not be called when B failed")
	}
}

// ---------------------------------------------------------------------------
// T-06 — Engine checkpoint/resume
// ---------------------------------------------------------------------------

func TestEngineSkipsPreseededStep(t *testing.T) {
	m := twoStepManifest()
	// Pre-seed step sync-hubspot as success
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	require.NoError(t, cs.Set(m.Name, "sync-hubspot", "success"))

	tracker := &stepCallTracker{}
	led := &stubLedger{}
	e := newEngineForTest(t, m, tracker, led, cs)

	result, err := e.Run(context.Background(), RunOptions{})
	require.NoError(t, err)

	// sync-hubspot should not be dispatched
	for _, call := range tracker.order {
		assert.NotEqual(t, "hubspot-prod", call, "sync-hubspot should have been skipped")
	}

	// Find the StepResult for sync-hubspot and assert status=skipped
	var syncResult *StepResult
	for i := range result.Steps {
		if result.Steps[i].ID == "sync-hubspot" {
			syncResult = &result.Steps[i]
			break
		}
	}
	require.NotNil(t, syncResult, "sync-hubspot should appear in Steps")
	assert.Equal(t, "skipped", syncResult.Status)
}

func TestEngineForceReclears(t *testing.T) {
	m := twoStepManifest()
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	require.NoError(t, cs.Set(m.Name, "sync-hubspot", "success"))

	tracker := &stepCallTracker{}
	led := &stubLedger{}
	e := newEngineForTest(t, m, tracker, led, cs)

	_, err := e.Run(context.Background(), RunOptions{Force: true})
	require.NoError(t, err)

	// With Force, sync-hubspot must be called even though it was pre-seeded
	found := false
	for _, call := range tracker.order {
		if call == "hubspot-prod" {
			found = true
			break
		}
	}
	assert.True(t, found, "sync-hubspot should have been called with Force=true")
}

func TestEngineCheckpointsPersistedAfterRun(t *testing.T) {
	m := twoStepManifest()
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	tracker := &stepCallTracker{}
	led := &stubLedger{}
	e := newEngineForTest(t, m, tracker, led, cs)

	_, err := e.Run(context.Background(), RunOptions{})
	require.NoError(t, err)

	a, err := cs.Get(m.Name, "sync-hubspot")
	require.NoError(t, err)
	assert.Equal(t, "success", a)

	b, err := cs.Get(m.Name, "score-contacts")
	require.NoError(t, err)
	assert.Equal(t, "success", b)
}

// ---------------------------------------------------------------------------
// T-07 — Engine ledger writes
// ---------------------------------------------------------------------------

func TestEngineLedgerEntriesSuccessfulRun(t *testing.T) {
	m := twoStepManifest()
	tracker := &stepCallTracker{}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, m, tracker, led, cs)

	result, err := e.Run(context.Background(), RunOptions{})
	require.NoError(t, err)
	require.Equal(t, "ok", result.Status)

	records := led.all()
	// Expect: 2 entries per step (running + success) = 4, plus 1 flow-level entry = 5
	assert.GreaterOrEqual(t, len(records), 5,
		"expected at least 5 ledger entries (2 per step × 2 + 1 flow), got %d: %v", len(records), records)

	// Count running and success entries per step
	type key struct{ op, status string }
	counts := map[key]int{}
	for _, r := range records {
		counts[key{r.Operation, r.Status}]++
	}
	assert.Equal(t, 1, counts[key{"likely-customers/sync-hubspot", "running"}])
	assert.Equal(t, 1, counts[key{"likely-customers/sync-hubspot", "success"}])
	assert.Equal(t, 1, counts[key{"likely-customers/score-contacts", "running"}])
	assert.Equal(t, 1, counts[key{"likely-customers/score-contacts", "success"}])
}

func TestEngineLedgerEntriesFailedRun(t *testing.T) {
	m := twoStepManifest()
	tracker := &stepCallTracker{
		fail: map[string]error{"hubspot-prod": errors.New("sync failed")},
	}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, m, tracker, led, cs)

	result, err := e.Run(context.Background(), RunOptions{})
	require.Error(t, err)
	assert.Equal(t, "failed", result.Status)

	type key struct{ op, status string }
	counts := map[key]int{}
	for _, r := range led.all() {
		counts[key{r.Operation, r.Status}]++
	}
	assert.Equal(t, 1, counts[key{"likely-customers/sync-hubspot", "running"}])
	assert.Equal(t, 1, counts[key{"likely-customers/sync-hubspot", "failed"}])
}

// ---------------------------------------------------------------------------
// T-09 — Integration smoke: full sync→query chain
// ---------------------------------------------------------------------------

func TestEngineIntegrationSyncQueryChain(t *testing.T) {
	m := twoStepManifest()
	tracker := &stepCallTracker{}
	led := &stubLedger{}
	cs := &FileCheckpointStore{Dir: t.TempDir()}
	e := newEngineForTest(t, m, tracker, led, cs)

	result, err := e.Run(context.Background(), RunOptions{})
	require.NoError(t, err)

	// Execution order
	require.Len(t, tracker.order, 2, "both steps should execute")
	assert.Equal(t, "hubspot-prod", tracker.order[0], "sync step first")

	// RunResult
	assert.Equal(t, "ok", result.Status)
	assert.Equal(t, "likely-customers", result.FlowName)
	assert.Len(t, result.Steps, 2)

	// Checkpoints persisted
	a, _ := cs.Get(m.Name, "sync-hubspot")
	assert.Equal(t, "success", a)
	b, _ := cs.Get(m.Name, "score-contacts")
	assert.Equal(t, "success", b)

	// Ledger has correct entry count
	assert.GreaterOrEqual(t, len(led.all()), 5)
}
