package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func TestFlowOmittedConnectionRejectsGeneratedOwnerAliases(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	var globex app.Connection
	for _, connection := range a.ListConnections() {
		if connection.Name == "globex" {
			globex = connection
			break
		}
	}
	require.NotEmpty(t, globex.ID)
	alias := "records__" + globex.ID

	assertRejected := func(t *testing.T, name, sql string) {
		t.Helper()
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    name,
				Steps: []flow.FlowStep{{
					ID:   "query-records",
					Kind: flow.KindQuery,
					SQL:  sql,
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:        adapter,
			Checkpoint: checkpoints,
			LockDir:    t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.Equal(t, "records", ambiguous.Table)
		assert.Contains(t, err.Error(), "set `connection` in flow query step")
		assert.NotContains(t, err.Error(), "--")
		assert.Equal(t, "failed", result.Status)
		require.Len(t, result.Steps, 1)
		assert.Equal(t, "failed", result.Steps[0].Status)
		assert.Empty(t, adapter.lastRows)
		checkpoint, checkpointErr := checkpoints.Get(name, "query-records")
		require.NoError(t, checkpointErr)
		assert.Empty(t, checkpoint)
	}

	t.Run("unquoted generated alias", func(t *testing.T) {
		assertRejected(t, "unquoted-generated-owner-alias", "SELECT id FROM "+alias)
	})

	t.Run("quoted generated alias", func(t *testing.T) {
		assertRejected(t, "quoted-generated-owner-alias", "SELECT id FROM \""+alias+"\"")
	})

	t.Run("generic query keeps the same alias", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := Run([]string{
			"--root", filepath.Dir(a.ProjectDir()),
			"--json",
			"query", "run",
			"--sql", "SELECT id FROM " + alias,
		}, &stdout, &stderr)
		require.Equalf(t, 0, code, "stderr = %s stdout = %s", stderr.String(), stdout.String())

		var result struct {
			Kind  string           `json:"kind"`
			Count int              `json:"count"`
			Rows  []map[string]any `json:"rows"`
		}
		require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
		assert.Equal(t, "QueryResult", result.Kind)
		require.Len(t, result.Rows, 1)
		assert.Equal(t, 1, result.Count)
		assert.Equal(t, "globex-1", result.Rows[0]["id"])
	})

	t.Run("select one remains valid", func(t *testing.T) {
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    "unscoped-select-one",
				Steps: []flow.FlowStep{{
					ID:   "query-one",
					Kind: flow.KindQuery,
					SQL:  "SELECT 1 AS n",
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:     adapter,
			LockDir: t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{})
		require.NoError(t, err)
		assert.Equal(t, "ok", result.Status)
		require.Len(t, adapter.lastRows, 1)
		assert.Equal(t, "1", fmt.Sprint(adapter.lastRows[0]["n"]))
	})
}

func TestFlowGeneratedOwnerAliasCollision(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	var acme, globex app.Connection
	for _, connection := range a.ListConnections() {
		switch connection.Name {
		case "acme":
			acme = connection
		case "globex":
			globex = connection
		}
	}
	require.NotEmpty(t, acme.ID)
	require.NotEmpty(t, globex.ID)
	alias := "records__" + globex.ID

	warehouseDir := filepath.Join(a.ProjectDir(), "warehouse")
	location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", acme.ID, acme.Name)
	require.NoError(t, err)
	require.NoError(t, location.EnsureOwnership())
	path, err := location.TablePath(alias)
	require.NoError(t, err)
	require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{{
		"id":         "real-collision-table",
		"updated_at": "2026-08-11T00:00:00Z",
	}}))

	assertRejected := func(t *testing.T, name, sql string) {
		t.Helper()
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    name,
				Steps: []flow.FlowStep{{
					ID:   "query-records",
					Kind: flow.KindQuery,
					SQL:  sql,
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:        adapter,
			Checkpoint: checkpoints,
			LockDir:    t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.Equal(t, "records", ambiguous.Table)
		assert.Contains(t, err.Error(), "set `connection` in flow query step")
		assert.Equal(t, "failed", result.Status)
		require.Len(t, result.Steps, 1)
		assert.Equal(t, "failed", result.Steps[0].Status)
		assert.Empty(t, adapter.lastRows)
		checkpoint, checkpointErr := checkpoints.Get(name, "query-records")
		require.NoError(t, checkpointErr)
		assert.Empty(t, checkpoint)
	}

	for _, test := range []struct {
		name string
		sql  string
	}{
		{name: "unquoted", sql: "SELECT id FROM " + alias},
		{name: "quoted", sql: "SELECT id FROM \"" + alias + "\""},
	} {
		t.Run("omitted flow "+test.name, func(t *testing.T) {
			assertRejected(t, "collision-"+test.name, test.sql)
		})

		t.Run("generic real table "+test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{
				"--root", filepath.Dir(a.ProjectDir()),
				"--json",
				"query", "run",
				"--sql", test.sql,
			}, &stdout, &stderr)
			require.Equalf(t, 0, code, "stderr = %s stdout = %s", stderr.String(), stdout.String())

			var result struct {
				Kind  string           `json:"kind"`
				Count int              `json:"count"`
				Rows  []map[string]any `json:"rows"`
			}
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
			assert.Equal(t, "QueryResult", result.Kind)
			require.Len(t, result.Rows, 1)
			assert.Equal(t, 1, result.Count)
			assert.Equal(t, "real-collision-table", result.Rows[0]["id"])
		})
	}
}

func TestFlowGeneratedOwnerAliasCaseVariantCollision(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	var acme, globex app.Connection
	for _, connection := range a.ListConnections() {
		switch connection.Name {
		case "acme":
			acme = connection
		case "globex":
			globex = connection
		}
	}
	require.NotEmpty(t, acme.ID)
	require.NotEmpty(t, globex.ID)
	alias := "records__" + globex.ID
	caseVariantAlias := strings.ToUpper(alias)

	warehouseDir := filepath.Join(a.ProjectDir(), "warehouse")
	location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", acme.ID, acme.Name)
	require.NoError(t, err)
	require.NoError(t, location.EnsureOwnership())
	path, err := location.TablePath(caseVariantAlias)
	require.NoError(t, err)
	require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{{
		"id":         "real-case-collision-table",
		"updated_at": "2026-08-11T00:00:00Z",
	}}))

	assertRejected := func(t *testing.T, name, sql string) {
		t.Helper()
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    name,
				Steps: []flow.FlowStep{{
					ID:   "query-records",
					Kind: flow.KindQuery,
					SQL:  sql,
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:        adapter,
			Checkpoint: checkpoints,
			LockDir:    t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.Equal(t, "records", ambiguous.Table)
		assert.Contains(t, err.Error(), "set `connection` in flow query step")
		assert.Equal(t, "failed", result.Status)
		require.Len(t, result.Steps, 1)
		assert.Equal(t, "failed", result.Steps[0].Status)
		assert.Empty(t, adapter.lastRows)
		checkpoint, checkpointErr := checkpoints.Get(name, "query-records")
		require.NoError(t, checkpointErr)
		assert.Empty(t, checkpoint)
	}

	for _, test := range []struct {
		name string
		sql  string
	}{
		{name: "unquoted lowercase", sql: "SELECT id FROM " + alias},
		{name: "quoted lowercase", sql: "SELECT id FROM \"" + alias + "\""},
		{name: "unquoted uppercase", sql: "SELECT id FROM " + caseVariantAlias},
		{name: "quoted uppercase", sql: "SELECT id FROM \"" + caseVariantAlias + "\""},
	} {
		t.Run("omitted flow "+test.name, func(t *testing.T) {
			assertRejected(t, "case-collision-"+test.name, test.sql)
		})

		t.Run("generic real table "+test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{
				"--root", filepath.Dir(a.ProjectDir()),
				"--json",
				"query", "run",
				"--sql", test.sql,
			}, &stdout, &stderr)
			require.Equalf(t, 0, code, "stderr = %s stdout = %s", stderr.String(), stdout.String())

			var result struct {
				Kind  string           `json:"kind"`
				Count int              `json:"count"`
				Rows  []map[string]any `json:"rows"`
			}
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
			assert.Equal(t, "QueryResult", result.Kind)
			require.Len(t, result.Rows, 1)
			assert.Equal(t, 1, result.Count)
			assert.Equal(t, "real-case-collision-table", result.Rows[0]["id"])
		})
	}
}

func TestFlowCaseVariantBareTableAmbiguity(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	caseVariantTable := strings.ToUpper("records")
	caseConnection, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
		Name: "case-owner",
		Source: app.EndpointConfig{
			Connector:  "warehouse",
			Credential: "warehouse-local",
		},
		Destination: app.EndpointConfig{
			Connector:  "warehouse",
			Credential: "warehouse-local",
		},
		Streams: map[string]app.StreamConfig{
			"case-records": {
				SyncMode:         "incremental_append_deduped",
				CursorField:      "updated_at",
				PrimaryKey:       []string{"id"},
				DestinationTable: caseVariantTable,
			},
		},
	})
	require.NoError(t, err)

	warehouseDir := filepath.Join(a.ProjectDir(), "warehouse")
	location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", caseConnection.ID, caseConnection.Name)
	require.NoError(t, err)
	require.NoError(t, location.EnsureOwnership())
	path, err := location.TablePath(caseVariantTable)
	require.NoError(t, err)
	require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{{
		"id":         "real-bare-case-table",
		"updated_at": "2026-08-11T00:00:00Z",
	}}))

	assertRejected := func(t *testing.T, name, sql string) {
		t.Helper()
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    name,
				Steps: []flow.FlowStep{{
					ID:   "query-records",
					Kind: flow.KindQuery,
					SQL:  sql,
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:        adapter,
			Checkpoint: checkpoints,
			LockDir:    t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.Equal(t, "records", ambiguous.Table)
		assert.Contains(t, err.Error(), "set `connection` in flow query step")
		assert.Equal(t, "failed", result.Status)
		require.Len(t, result.Steps, 1)
		assert.Equal(t, "failed", result.Steps[0].Status)
		assert.Empty(t, adapter.lastRows)
		checkpoint, checkpointErr := checkpoints.Get(name, "query-records")
		require.NoError(t, checkpointErr)
		assert.Empty(t, checkpoint)
	}

	for _, test := range []struct {
		name string
		sql  string
	}{
		{name: "unquoted lowercase", sql: "SELECT id FROM records"},
		{name: "quoted lowercase", sql: "SELECT id FROM \"records\""},
		{name: "unquoted uppercase", sql: "SELECT id FROM " + caseVariantTable},
		{name: "quoted uppercase", sql: "SELECT id FROM \"" + caseVariantTable + "\""},
	} {
		t.Run("omitted flow "+test.name, func(t *testing.T) {
			assertRejected(t, "bare-case-"+test.name, test.sql)
		})

		t.Run("generic real table "+test.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run([]string{
				"--root", filepath.Dir(a.ProjectDir()),
				"--json",
				"query", "run",
				"--sql", test.sql,
			}, &stdout, &stderr)
			require.Equalf(t, 0, code, "stderr = %s stdout = %s", stderr.String(), stdout.String())

			var result struct {
				Kind  string           `json:"kind"`
				Count int              `json:"count"`
				Rows  []map[string]any `json:"rows"`
			}
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &result))
			assert.Equal(t, "QueryResult", result.Kind)
			require.Len(t, result.Rows, 1)
			assert.Equal(t, 1, result.Count)
			assert.Equal(t, "real-bare-case-table", result.Rows[0]["id"])
		})
	}
}

func TestFlowCaseEquivalentUniqueTablesPreserveGenericSQLAndTypedAmbiguity(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowCaseEquivalentUniqueWarehouseApp(t, ctx)

	var acme app.Connection
	for _, connection := range a.ListConnections() {
		if connection.Name == "acme" {
			acme = connection
			break
		}
	}
	require.NotEmpty(t, acme.ID)
	collisionTable := "RECORDS__" + acme.ID
	warehouseDir := filepath.Join(a.ProjectDir(), "warehouse")
	location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", acme.ID, acme.Name)
	require.NoError(t, err)
	path, err := location.TablePath(collisionTable)
	require.NoError(t, err)
	require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{{
		"id":         "physical-owner-alias-collision",
		"updated_at": "2026-08-12T00:00:00Z",
	}}))

	for _, test := range []struct {
		name       string
		connection string
		table      string
		wantID     string
	}{
		{name: "acme lower-case table", connection: "acme", table: "records", wantID: "acme-case-unique"},
		{name: "globex upper-case table", connection: "globex", table: "RECORDS", wantID: "globex-case-unique"},
	} {
		t.Run("selected "+test.name, func(t *testing.T) {
			records, err := a.QuerySQL(ctx, app.QuerySQLRequest{
				SQL:        "SELECT id FROM " + test.table,
				Connection: test.connection,
			})
			require.NoError(t, err)
			require.Len(t, records, 1)
			assert.Equal(t, test.wantID, fmt.Sprint(records[0]["id"]))
		})
	}

	t.Run("unrelated generic SQL remains executable", func(t *testing.T) {
		records, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "SELECT 1 AS n"})
		require.NoError(t, err)
		require.Len(t, records, 1)
		assert.Equal(t, "1", fmt.Sprint(records[0]["n"]))
	})

	t.Run("real case-equivalent owner alias remains queryable", func(t *testing.T) {
		records, err := a.QuerySQL(ctx, app.QuerySQLRequest{
			SQL: "SELECT id FROM \"" + collisionTable + "\"",
		})
		require.NoError(t, err)
		require.Len(t, records, 1)
		assert.Equal(t, "physical-owner-alias-collision", fmt.Sprint(records[0]["id"]))
	})

	assertRejected := func(t *testing.T, name, sql string) {
		t.Helper()
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    name,
				Steps: []flow.FlowStep{{
					ID:   "query-records",
					Kind: flow.KindQuery,
					SQL:  sql,
					In:   []string{},
					Out:  []string{},
				}},
			},
			App:        adapter,
			Checkpoint: checkpoints,
			LockDir:    t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{})
		require.Error(t, err)
		var ambiguous *warehouse.AmbiguousTableError
		require.Truef(t, errors.As(err, &ambiguous), "error = %T %v", err, err)
		assert.Equal(t, "records", ambiguous.Table)
		assert.Contains(t, err.Error(), "set `connection` in flow query step")
		assert.Equal(t, "failed", result.Status)
		require.Len(t, result.Steps, 1)
		assert.Equal(t, "failed", result.Steps[0].Status)
		assert.Empty(t, adapter.lastRows)
		checkpoint, checkpointErr := checkpoints.Get(name, "query-records")
		require.NoError(t, checkpointErr)
		assert.Empty(t, checkpoint)
	}

	for _, test := range []struct {
		name string
		sql  string
	}{
		{name: "unquoted lower-case", sql: "SELECT id FROM records"},
		{name: "quoted upper-case", sql: "SELECT id FROM \"RECORDS\""},
	} {
		t.Run("omitted flow "+test.name, func(t *testing.T) {
			assertRejected(t, "case-equivalent-unique-"+test.name, test.sql)
		})
	}
}

func TestFlowRefusesCaseEquivalentFaultedOwner(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowCaseEquivalentUniqueWarehouseApp(t, ctx)

	var globex app.Connection
	for _, connection := range a.ListConnections() {
		if connection.Name == "globex" {
			globex = connection
			break
		}
	}
	require.NotEmpty(t, globex.ID)
	warehouseDir := filepath.Join(a.ProjectDir(), "warehouse")
	location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", globex.ID, globex.Name)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(location.OwnerPath(), []byte("{"), 0o600))

	direct, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "records", Limit: 1})
	require.NoError(t, err)
	require.Len(t, direct, 1)
	assert.Equal(t, "acme-case-unique", fmt.Sprint(direct[0]["id"]))

	generic, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "SELECT id FROM records"})
	require.NoError(t, err)
	require.Len(t, generic, 1)
	assert.Equal(t, "acme-case-unique", fmt.Sprint(generic[0]["id"]))

	selected, err := a.QuerySQL(ctx, app.QuerySQLRequest{
		SQL:        "SELECT id FROM records",
		Connection: "acme",
		Origin:     app.QuerySQLOriginFlow,
	})
	require.NoError(t, err)
	require.Len(t, selected, 1)
	assert.Equal(t, "acme-case-unique", fmt.Sprint(selected[0]["id"]))

	for _, tc := range []struct {
		name string
		sql  string
	}{
		{name: "lowercase", sql: "SELECT id FROM records"},
		{name: "quoted uppercase", sql: `SELECT id FROM "RECORDS"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
			adapter := &recordingFlowAppAdapter{app: a}
			engine := &flow.Engine{
				Manifest: flow.FlowManifest{
					Version: 1,
					Name:    "faulted-case-equivalent-" + tc.name,
					Steps: []flow.FlowStep{{
						ID:   "query-records",
						Kind: flow.KindQuery,
						SQL:  tc.sql,
						In:   []string{},
						Out:  []string{},
					}},
				},
				App:        adapter,
				Checkpoint: checkpoints,
				LockDir:    t.TempDir(),
			}

			result, runErr := engine.Run(ctx, flow.RunOptions{})
			require.Error(t, runErr)
			var faulted *warehouse.FaultError
			require.Truef(t, errors.As(runErr, &faulted), "error = %T %v", runErr, runErr)
			assert.True(t, faulted.Undecided)
			assert.Equal(t, "failed", result.Status)
			require.Len(t, result.Steps, 1)
			assert.Equal(t, "failed", result.Steps[0].Status)
			assert.Empty(t, adapter.lastRows)
			checkpoint, checkpointErr := checkpoints.Get(engine.Manifest.Name, "query-records")
			require.NoError(t, checkpointErr)
			assert.Empty(t, checkpoint)
		})
	}
}

// TestFlowLegacySameOwnerCaseEquivalentInventoryStopsAtTheTypedBoundary
// covers the old state that #4069 cannot migrate. The owner selection cannot
// resolve two destinations declared by that same owner, so every SQL spelling
// must fail truthfully while a query unrelated to warehouse inventory keeps
// running. Running the same flow again models the scheduler payload, which is
// simply `pm flow run` re-entering this boundary.
func TestFlowLegacySameOwnerCaseEquivalentInventoryStopsAtTheTypedBoundary(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowLegacySameOwnerCaseEquivalentWarehouseApp(t, ctx)

	rows, err := a.QuerySQL(ctx, app.QuerySQLRequest{SQL: "SELECT 1 AS n"})
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, "1", fmt.Sprint(rows[0]["n"]))

	assertRejected := func(t *testing.T, name, connection, sql string, attempts int) {
		t.Helper()
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		adapter := &recordingFlowAppAdapter{app: a}
		engine := &flow.Engine{
			Manifest: flow.FlowManifest{
				Version: 1,
				Name:    name,
				Steps: []flow.FlowStep{{
					ID:         "query-records",
					Kind:       flow.KindQuery,
					Connection: connection,
					SQL:        sql,
					In:         []string{},
					Out:        []string{},
				}},
			},
			App:        adapter,
			Checkpoint: checkpoints,
			LockDir:    t.TempDir(),
		}
		for attempt := 1; attempt <= attempts; attempt++ {
			result, runErr := engine.Run(ctx, flow.RunOptions{})
			require.Error(t, runErr)
			var collision *warehouse.SameOwnerCaseEquivalentTableError
			require.Truef(t, errors.As(runErr, &collision), "attempt %d error = %T %v", attempt, runErr, runErr)
			var ambiguous *warehouse.AmbiguousTableError
			assert.Falsef(t, errors.As(runErr, &ambiguous), "attempt %d used cross-owner ambiguity: %T %v", attempt, runErr, runErr)
			assert.Contains(t, collision.Error(), "acme")
			assert.Contains(t, collision.Error(), "records")
			assert.Contains(t, collision.Error(), "RECORDS")
			assert.NotContains(t, collision.Error(), "set `connection`", "a selector cannot resolve one-owner destinations")
			assert.NotContains(t, collision.Error(), "Catalog Error")
			assert.Contains(t, collision.Error(), "exact resolver-visible table spelling")
			assert.Contains(t, collision.Error(), "replacement connections")
			assert.Equal(t, "failed", result.Status)
			require.Len(t, result.Steps, 1)
			assert.Equal(t, "failed", result.Steps[0].Status)
			assert.Empty(t, adapter.lastRows)
			checkpoint, checkpointErr := checkpoints.Get(name, "query-records")
			require.NoError(t, checkpointErr)
			assert.Empty(t, checkpoint, "attempt %d must not create a successful checkpoint", attempt)
		}
	}

	for _, tc := range []struct {
		name       string
		connection string
		sql        string
	}{
		{name: "omitted bare", sql: "SELECT id FROM records"},
		{name: "omitted quoted", sql: `SELECT id FROM "RECORDS"`},
		{name: "selected bare", connection: "acme", sql: "SELECT id FROM records"},
		{name: "selected quoted", connection: "acme", sql: `SELECT id FROM "RECORDS"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			attempts := 1
			if tc.name == "omitted bare" {
				attempts = 2 // scheduler re-entry calls the same flow boundary again.
			}
			assertRejected(t, "same-owner-"+strings.ReplaceAll(tc.name, " ", "-"), tc.connection, tc.sql, attempts)
		})
	}
}

func TestFlowActionSourceReadsAllSelectedConnectionRows(t *testing.T) {
	ctx := testCtx(t)
	a := newFlowScopedWarehouseApp(t, ctx)

	var acme app.Connection
	for _, connection := range a.ListConnections() {
		if connection.Name == "acme" {
			acme = connection
			break
		}
	}
	require.NotEmpty(t, acme.ID)
	warehouseDir := filepath.Join(a.ProjectDir(), "warehouse")
	location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", acme.ID, acme.Name)
	require.NoError(t, err)
	path, err := location.TablePath("records")
	require.NoError(t, err)

	wantIDs := make([]string, 101)
	rows := make([]warehouse.Row, len(wantIDs))
	for i := range rows {
		wantIDs[i] = fmt.Sprintf("acme-%03d", i)
		rows[i] = warehouse.Row{
			"id":         wantIDs[i],
			"updated_at": "2026-08-11T00:00:00Z",
		}
	}
	require.NoError(t, warehouse.WriteTable(ctx, path, rows))

	publicRows, err := a.QueryTable(ctx, app.QueryTableRequest{Table: "records", Connection: "acme", Limit: 0})
	require.NoError(t, err)
	require.Len(t, publicRows, 100)

	manifest := func(name string) flow.FlowManifest {
		return flow.FlowManifest{
			Version: 1,
			Name:    name,
			Steps: []flow.FlowStep{{
				ID:   "act-on-records",
				Kind: flow.KindAction,
				ActionCfg: &flow.ActionConfig{
					SourceTable:           "records",
					SourceConnection:      "acme",
					DestinationConnector:  "future-connector",
					DestinationCredential: "future-credential",
					Action:                "create",
					Mappings:              map[string]string{"id": "external_id"},
				},
				In:  []string{},
				Out: []string{},
			}},
		}
	}
	assertSelectedRows := func(t *testing.T, records []map[string]any) {
		t.Helper()
		require.Len(t, records, len(wantIDs))
		seen := make(map[string]struct{}, len(records))
		for _, record := range records {
			id, ok := record["id"].(string)
			require.Truef(t, ok, "record id = %#v", record["id"])
			seen[id] = struct{}{}
		}
		for _, wantID := range wantIDs {
			_, ok := seen[wantID]
			assert.Truef(t, ok, "action records do not include %q", wantID)
		}
		_, includesGlobex := seen["globex-1"]
		assert.False(t, includesGlobex)
	}

	t.Run("success dispatches every selected row", func(t *testing.T) {
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		runner := &captureFlowActionRunner{}
		engine := &flow.Engine{
			Manifest:     manifest("all-selected-action-success"),
			App:          &appFlowAdapter{app: a},
			ActionRunner: runner,
			Checkpoint:   checkpoints,
			LockDir:      t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{ApprovalToken: "test-only"})
		require.NoError(t, err)
		assert.Equal(t, "ok", result.Status)
		assertSelectedRows(t, runner.records)
		status, err := checkpoints.Get("all-selected-action-success", "act-on-records")
		require.NoError(t, err)
		assert.Equal(t, "success", status)
	})

	t.Run("failed dispatch leaves no success checkpoint", func(t *testing.T) {
		checkpoints := &flow.FileCheckpointStore{Dir: t.TempDir()}
		runner := &captureFlowActionRunner{err: errors.New("local action dispatch failed")}
		engine := &flow.Engine{
			Manifest:     manifest("all-selected-action-failure"),
			App:          &appFlowAdapter{app: a},
			ActionRunner: runner,
			Checkpoint:   checkpoints,
			LockDir:      t.TempDir(),
		}

		result, err := engine.Run(ctx, flow.RunOptions{ApprovalToken: "test-only"})
		require.Error(t, err)
		assert.Equal(t, "failed", result.Status)
		assertSelectedRows(t, runner.records)
		status, err := checkpoints.Get("all-selected-action-failure", "act-on-records")
		require.NoError(t, err)
		assert.Empty(t, status)
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
	records, err := a.app.QuerySQL(ctx, app.QuerySQLRequest{
		SQL:        sql,
		Connection: connection,
		Limit:      limit,
		Origin:     app.QuerySQLOriginFlow,
	})
	if err != nil {
		return nil, err
	}
	a.lastSQLConnection = connection
	a.lastRows = recordsToMapRows(records)
	return a.lastRows, nil
}

func (a *recordingFlowAppAdapter) ReadActionSource(ctx context.Context, table, connection string) ([]map[string]any, error) {
	records, err := a.app.ReadActionSource(ctx, app.ActionSourceReadRequest{Table: table, Connection: connection})
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
	err     error
}

func (r *captureFlowActionRunner) ExecuteStep(_ context.Context, step flow.FlowStep, records []map[string]any, _ string, _ string) (flow.ActionResult, error) {
	r.calls++
	r.step = step
	r.records = append([]map[string]any(nil), records...)
	if r.err != nil {
		return flow.ActionResult{RecordsAttempted: len(records)}, r.err
	}
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

func newFlowCaseEquivalentUniqueWarehouseApp(t *testing.T, ctx context.Context) *app.App {
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

	for _, spec := range []struct {
		name  string
		table string
	}{
		{name: "acme", table: "records"},
		{name: "globex", table: "RECORDS"},
	} {
		_, err := a.CreateConnection(ctx, app.CreateConnectionRequest{
			Name: spec.name,
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
					DestinationTable: spec.table,
				},
			},
		})
		require.NoError(t, err)
	}

	for _, connection := range a.ListConnections() {
		table, rowID := "records", "acme-case-unique"
		if connection.Name == "globex" {
			table, rowID = "RECORDS", "globex-case-unique"
		}
		location, err := warehouse.LocationFor(warehouseDir, "flow_scope", "warehouse", connection.ID, connection.Name)
		require.NoError(t, err)
		require.NoError(t, location.EnsureOwnership())
		path, err := location.TablePath(table)
		require.NoError(t, err)
		require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{{
			"id":         rowID,
			"updated_at": "2026-08-12T00:00:00Z",
		}}))
	}
	return a
}

// newFlowLegacySameOwnerCaseEquivalentWarehouseApp writes the exact legacy
// state a pre-admission client could have persisted. It intentionally mutates
// only a test project state file after creating a valid one-stream connection:
// the production creation path must reject the second spelling once this
// correction lands.
func newFlowLegacySameOwnerCaseEquivalentWarehouseApp(t *testing.T, ctx context.Context) *app.App {
	t.Helper()
	root := t.TempDir()
	initProject(t, root)
	created, err := app.Open(root)
	require.NoError(t, err)

	warehouseDir := filepath.Join(root, ".polymetrics", "warehouse")
	_, err = created.AddCredential(ctx, app.AddCredentialRequest{
		Name:      "warehouse-local",
		Connector: "warehouse",
		Config:    map[string]string{"path": warehouseDir},
	})
	require.NoError(t, err)
	connection, err := created.CreateConnection(ctx, app.CreateConnectionRequest{
		Name: "acme",
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

	statePath := filepath.Join(root, ".polymetrics", "state", "state.json")
	raw, err := os.ReadFile(statePath)
	require.NoError(t, err)
	var state map[string]any
	require.NoError(t, json.Unmarshal(raw, &state))
	workspaceID, ok := state["workspace_id"].(string)
	require.True(t, ok)
	connections, ok := state["connections"].([]any)
	require.True(t, ok)
	require.Len(t, connections, 1)
	connectionState, ok := connections[0].(map[string]any)
	require.True(t, ok)
	streams, ok := connectionState["streams"].(map[string]any)
	require.True(t, ok)
	records, ok := streams["records"].(map[string]any)
	require.True(t, ok)
	caseRecords := make(map[string]any, len(records))
	for key, value := range records {
		caseRecords[key] = value
	}
	caseRecords["destination_table"] = "RECORDS"
	caseRecords["stream_id"] = "stream_legacy_case_records"
	streams["case-records"] = caseRecords
	updated, err := json.Marshal(state)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(statePath, updated, 0o600))

	location, err := warehouse.LocationFor(warehouseDir, workspaceID, "warehouse", connection.ID, connection.Name)
	require.NoError(t, err)
	require.NoError(t, location.EnsureOwnership())
	path, err := location.TablePath("records")
	require.NoError(t, err)
	require.NoError(t, warehouse.WriteTable(ctx, path, []warehouse.Row{{
		"id":         "legacy-physical-record",
		"updated_at": "2026-08-12T00:00:00Z",
	}}))

	reopened, err := app.Open(root)
	require.NoError(t, err)
	return reopened
}
