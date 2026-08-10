package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/warehouse"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

// validFlowManifestJSON is a minimal valid two-step flow manifest for CLI tests.
const validFlowManifestJSON = `{
	"version": 1,
	"name": "test-flow",
	"description": "CLI test flow",
	"steps": [
		{
			"id": "sync-step",
			"kind": "sync",
			"connection": "conn-1",
			"streams": ["users"],
			"in": [],
			"out": ["users"]
		},
		{
			"id": "query-step",
			"kind": "query",
			"sql": "SELECT * FROM users",
			"in": ["users"],
			"out": ["scored"]
		}
	]
}`

const cyclicFlowManifestJSON = `{
	"version": 1,
	"name": "cyclic-flow",
	"steps": [
		{
			"id": "A",
			"kind": "query",
			"sql": "SELECT 1",
			"in": ["tb"],
			"out": ["ta"]
		},
		{
			"id": "B",
			"kind": "query",
			"sql": "SELECT 2",
			"in": ["ta"],
			"out": ["tb"]
		}
	]
}`

// writeManifestFile writes content to a temp file and returns the path.
func writeManifestFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
	return path
}

// TestFlowList checks that `pm flow list` with an empty flows dir returns {"flows":[]}.
func TestFlowList(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := runFlow(testCtx(t), config.Config{}, nil, []string{"list", "--flows-dir", dir}, &out, true)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	flows, ok := result["flows"]
	require.True(t, ok, "response must have 'flows' key")
	flowsSlice, ok := flows.([]any)
	require.True(t, ok)
	assert.Empty(t, flowsSlice)
}

// TestFlowPlanValid checks that `pm flow plan --file valid.json` exits 0 with status=ok.
func TestFlowPlanValid(t *testing.T) {
	path := writeManifestFile(t, validFlowManifestJSON)
	var out bytes.Buffer
	err := runFlow(testCtx(t), config.Config{}, nil, []string{"plan", "--file", path}, &out, true)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	assert.Equal(t, "ok", result["status"])
}

// TestFlowPlanCyclic checks that `pm flow plan --file cyclic.json` returns a non-nil error.
func TestFlowPlanCyclic(t *testing.T) {
	path := writeManifestFile(t, cyclicFlowManifestJSON)
	var out bytes.Buffer
	err := runFlow(testCtx(t), config.Config{}, nil, []string{"plan", "--file", path}, &out, true)
	require.Error(t, err, "cyclic manifest should produce an error")
	assert.True(t, strings.Contains(err.Error(), "cyclic") || strings.Contains(err.Error(), "flow:"),
		"error should mention cycle: %v", err)
}

// TestFlowStatusMissing checks that `pm flow status <missing>` returns an error.
func TestFlowStatusMissing(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := runFlow(testCtx(t), config.Config{}, nil, []string{"status", "nonexistent", "--flows-dir", dir}, &out, true)
	require.Error(t, err)
}

// TestFlowPreviewValid checks that `pm flow preview --file valid.json` returns dry_run status.
func TestFlowPreviewValid(t *testing.T) {
	path := writeManifestFile(t, validFlowManifestJSON)
	var out bytes.Buffer
	err := runFlow(testCtx(t), config.Config{}, nil, []string{"preview", "--file", path}, &out, true)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	assert.Equal(t, "dry_run", result["status"])
}

func TestFlowRunByNameResolvesProjectFlowManifest(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)
	flowsDir := filepath.Join(root, ".polymetrics", "flows")
	require.NoError(t, os.MkdirAll(flowsDir, 0o755))

	spec := `{
		"name": "lead-score",
		"features": [
			{"name": "email", "weight": 1.0, "score_if_set": 1.0}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(flowsDir, "lead-score.json"), []byte(spec), 0o644))

	manifest := `{
		"version": 1,
		"name": "named-flow",
		"steps": [
			{
				"id": "score",
				"kind": "rlm",
				"spec": "lead-score.json",
				"mode": "fixture",
				"in": [],
				"out": ["named_scores"]
			}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(flowsDir, "named-flow.json"), []byte(manifest), 0o644))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "flow", "run", "named-flow"}, &stdout, &stderr)
	require.Equalf(t, 0, code, "stderr = %s stdout = %s", stderr.String(), stdout.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	assert.Equal(t, "ok", result["status"])
	_, err := os.Stat(filepath.Join(root, ".polymetrics", "warehouse", "named_scores.ndjson"))
	require.NoError(t, err)
}

func TestFlowRunRLMFixtureMaterializesOutTable(t *testing.T) {
	root := t.TempDir()
	initProject(t, root)

	flowDir := t.TempDir()
	spec := `{
		"name": "lead-score",
		"features": [
			{"name": "email", "weight": 0.5, "score_if_set": 1.0},
			{"name": "company", "weight": 0.5, "score_if_set": 1.0}
		]
	}`
	require.NoError(t, os.WriteFile(filepath.Join(flowDir, "lead-score.json"), []byte(spec), 0o644))

	manifest := `{
		"version": 1,
		"name": "fixture-leads",
		"steps": [
			{
				"id": "score",
				"kind": "rlm",
				"spec": "lead-score.json",
				"mode": "fixture",
				"in": [],
				"out": ["lead_scores"]
			}
		]
	}`
	manifestPath := filepath.Join(flowDir, "flow.json")
	require.NoError(t, os.WriteFile(manifestPath, []byte(manifest), 0o644))

	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", root, "--json", "flow", "run", "--file", manifestPath}, &stdout, &stderr)
	require.Equalf(t, 0, code, "stderr = %s stdout = %s", stderr.String(), stdout.String())

	var result map[string]any
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
	assert.Equal(t, "ok", result["status"])

	outPath := filepath.Join(root, ".polymetrics", "warehouse", "lead_scores.ndjson")
	data, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.NotEmpty(t, strings.TrimSpace(string(data)))
}

// TestFlowSourceConnectionSelectorsReadOnlyOwningRows is the red-first
// regression for #3897. It deliberately uses two structurally owned Parquet
// tables with the same name, then drives query and action source reads through
// the flow adapter. The action runner is a local capture stub: no provider
// call or mutation is possible in this test.
func TestFlowSourceConnectionSelectorsReadOnlyOwningRows(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	t.Run("query selector reads only acme rows", func(t *testing.T) {
		adapter := &recordingFlowAppAdapter{app: a}
		manifest, err := flow.ParseManifest([]byte(`{
			"version": 1,
			"name": "scoped-query",
			"steps": [{
				"id": "query-records",
				"kind": "query",
				"connection": "acme",
				"sql": "SELECT * FROM records",
				"in": [],
				"out": []
			}]
		}`))
		require.NoError(t, err)
		encoded, err := json.Marshal(manifest)
		require.NoError(t, err)
		roundTripped, err := flow.ParseManifest(encoded)
		require.NoError(t, err)
		assert.Equal(t, "acme", roundTripped.Steps[0].Connection)

		engine := &flow.Engine{
			Manifest: manifest,
			App:      adapter,
			LockDir:  t.TempDir(),
		}
		result, err := engine.Run(ctx, flow.RunOptions{})
		require.NoError(t, err)
		assert.Equal(t, "ok", result.Status)
		require.Len(t, adapter.lastRows, 1)
		assert.Equal(t, "acme-1", adapter.lastRows[0]["id"])
		assert.Equal(t, "acme", adapter.lastSQLConnection)
	})

	t.Run("action source selector reads only globex rows", func(t *testing.T) {
		adapter := &recordingFlowAppAdapter{app: a}
		runner := &captureFlowActionRunner{}
		manifest, err := flow.ParseManifest([]byte(`{
			"version": 1,
			"name": "scoped-action",
			"steps": [{
				"id": "act-on-records",
				"kind": "action",
				"action_cfg": {
					"source_table": "records",
					"source_connection": "globex",
					"destination_connector": "future-connector",
					"destination_credential": "future-credential",
					"action": "create",
					"mappings": {"id": "external_id"}
				},
				"in": [],
				"out": []
			}]
		}`))
		require.NoError(t, err)
		encoded, err := json.Marshal(manifest)
		require.NoError(t, err)
		roundTripped, err := flow.ParseManifest(encoded)
		require.NoError(t, err)
		require.NotNil(t, roundTripped.Steps[0].ActionCfg)
		assert.Equal(t, "globex", roundTripped.Steps[0].ActionCfg.SourceConnection)

		engine := &flow.Engine{
			Manifest:     manifest,
			App:          adapter,
			ActionRunner: runner,
			LockDir:      t.TempDir(),
		}
		result, err := engine.Run(ctx, flow.RunOptions{ApprovalToken: "test-only"})
		require.NoError(t, err)
		assert.Equal(t, "ok", result.Status)
		require.Len(t, runner.records, 1)
		assert.Equal(t, "globex-1", runner.records[0]["id"])
		assert.Equal(t, "globex", adapter.lastTableConnection)
		require.NotNil(t, runner.step.ActionCfg)
		assert.Equal(t, "globex", runner.step.ActionCfg.SourceConnection)
	})
}

func TestFlowSourceConnectionSelectorRefusesOmissionAndAcceptsUnattributed(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	t.Run("query omission is a typed ambiguity with a manifest remedy", func(t *testing.T) {
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    "ambiguous-query",
				Steps: []flow.FlowStep{{
					ID:   "query-records",
					Kind: flow.KindQuery,
					SQL:  "SELECT * FROM records",
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:     adapter,
			LockDir: t.TempDir(),
		}

		_, err := engine.Run(ctx, flow.RunOptions{})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.NotContains(t, err.Error(), "--")
		assert.Contains(t, err.Error(), "`connection`")
		assert.Contains(t, ambiguous.Error(), "acme")
		assert.Contains(t, ambiguous.Error(), "globex")
	})

	t.Run("action omission is typed before the local action runner", func(t *testing.T) {
		adapter := &recordingFlowAppAdapter{app: a}
		runner := &captureFlowActionRunner{}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    "ambiguous-action",
				Steps: []flow.FlowStep{{
					ID:   "act-on-records",
					Kind: flow.KindAction,
					ActionCfg: &flow.ActionConfig{
						SourceTable:           "records",
						DestinationConnector:  "future-connector",
						DestinationCredential: "future-credential",
						Action:                "create",
						Mappings:              map[string]string{"id": "external_id"},
					},
					In:  []string{},
					Out: []string{},
				}},
			},
			App:          adapter,
			ActionRunner: runner,
			LockDir:      t.TempDir(),
		}

		_, err := engine.Run(ctx, flow.RunOptions{ApprovalToken: "test-only"})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.NotContains(t, err.Error(), "--")
		assert.Contains(t, err.Error(), "`action_cfg.source_connection`")
		assert.Zero(t, runner.calls, "the source ambiguity must stop before any action dispatch")
	})

	t.Run("unattributed selectors see root-owned rows only", func(t *testing.T) {
		rootTable := filepath.Join(a.ProjectDir(), "warehouse", "records"+warehouse.TableFileExt)
		require.NoError(t, warehouse.WriteTable(ctx, rootTable, []warehouse.Row{{"id": "root-1"}}))

		queryAdapter := &recordingFlowAppAdapter{app: a}
		queryEngine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    "root-query",
				Steps: []flow.FlowStep{{
					ID:         "query-root-records",
					Kind:       flow.KindQuery,
					Connection: warehouse.UnattributedConnection,
					SQL:        "SELECT * FROM records",
					In:         []string{},
					Out:        []string{},
				}},
			},
			App:     queryAdapter,
			LockDir: t.TempDir(),
		}
		_, err := queryEngine.Run(ctx, flow.RunOptions{})
		require.NoError(t, err)
		require.Len(t, queryAdapter.lastRows, 1)
		assert.Equal(t, "root-1", queryAdapter.lastRows[0]["id"])

		actionAdapter := &recordingFlowAppAdapter{app: a}
		runner := &captureFlowActionRunner{}
		actionEngine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    "root-action",
				Steps: []flow.FlowStep{{
					ID:   "act-on-root-records",
					Kind: flow.KindAction,
					ActionCfg: &flow.ActionConfig{
						SourceTable:           "records",
						SourceConnection:      warehouse.UnattributedConnection,
						DestinationConnector:  "future-connector",
						DestinationCredential: "future-credential",
						Action:                "create",
						Mappings:              map[string]string{"id": "external_id"},
					},
					In:  []string{},
					Out: []string{},
				}},
			},
			App:          actionAdapter,
			ActionRunner: runner,
			LockDir:      t.TempDir(),
		}
		_, err = actionEngine.Run(ctx, flow.RunOptions{ApprovalToken: "test-only"})
		require.NoError(t, err)
		require.Len(t, runner.records, 1)
		assert.Equal(t, "root-1", runner.records[0]["id"])
		assert.Equal(t, warehouse.UnattributedConnection, actionAdapter.lastTableConnection)
	})
}

type recordingFlowAppAdapter struct {
	app                 *app.App
	lastRows            []map[string]any
	lastSQLConnection   string
	lastTableConnection string
}

func (a *recordingFlowAppAdapter) ETLRun(_ context.Context, _ string, _ []string) (flow.ETLResult, error) {
	return flow.ETLResult{}, nil
}

func (a *recordingFlowAppAdapter) QuerySQL(ctx context.Context, sql, connection string, limit int) ([]map[string]any, error) {
	records, err := a.app.QuerySQL(ctx, app.QuerySQLRequest{SQL: sql, Connection: connection, Limit: limit})
	if err != nil {
		return nil, err
	}
	a.lastSQLConnection = connection
	a.lastRows = recordsToMapRows(records)
	return a.lastRows, nil
}

func (a *recordingFlowAppAdapter) QueryTable(ctx context.Context, table, connection string, limit int) ([]map[string]any, error) {
	records, err := a.app.QueryTable(ctx, app.QueryTableRequest{Table: table, Connection: connection, Limit: limit})
	if err != nil {
		return nil, err
	}
	a.lastTableConnection = connection
	a.lastRows = recordsToMapRows(records)
	return a.lastRows, nil
}

func recordsToMapRows(records []connectors.Record) []map[string]any {
	out := make([]map[string]any, len(records))
	for i, record := range records {
		out[i] = map[string]any(record)
	}
	return out
}

func (a *recordingFlowAppAdapter) RLMRun(_ context.Context, _ flow.RLMRunRequest) (flow.RLMResult, error) {
	return flow.RLMResult{}, nil
}

type captureFlowActionRunner struct {
	records []map[string]any
	step    flow.FlowStep
	calls   int
}

func (r *captureFlowActionRunner) ExecuteStep(_ context.Context, step flow.FlowStep, records []map[string]any, _ string, _ string) (flow.ActionResult, error) {
	r.calls++
	r.step = step
	r.records = append([]map[string]any(nil), records...)
	return flow.ActionResult{RecordsAttempted: len(records), RecordsSucceeded: len(records)}, nil
}

func newFlowScopedWarehouseApp(t *testing.T, ctx context.Context) *app.App {
	t.Helper()
	root := t.TempDir()
	initProject(t, root)
	a, err := app.Open(root)
	require.NoError(t, err)

	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	_, err = a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": warehouseDir},
	})
	require.NoError(t, err)

	for _, name := range []string{"acme", "globex"} {
		_, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
			Name: name,
			Source: app.EndpointConfig{
				Connector:  "warehouse",
				Credential: "warehouse-local",
			},
			Destination: app.EndpointConfig{
				Connector:  "warehouse",
				Credential: "warehouse-local",
			},
			Streams: map[string]app.StreamConfig{
				"records": {
					SyncMode:         "incremental_append_deduped",
					CursorField:      "updated_at",
					PrimaryKey:       []string{"id"},
					DestinationTable: "records",
				},
			},
		})
		require.NoError(t, err)
	}

	for _, connection := range a.ListConnections() {
		location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", connection.ID, connection.Name)
		require.NoError(t, err)
		require.NoError(t, location.EnsureOwnership())
		path, err := location.TablePath("records")
		require.NoError(t, err)
		row := warehouse.Row{
			"id":         connection.Name + "-1",
			"updated_at": "2026-08-11T00:00:00Z",
		}
		require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{row}))
	}
	return a
}
