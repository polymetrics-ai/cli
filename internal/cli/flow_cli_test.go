package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/safety"
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

// TestFlowPlanSanitizesUnsafeStepIDsInHumanOutput keeps human flow output terminal-safe without changing JSON.
func TestFlowPlanSanitizesUnsafeStepIDsInHumanOutput(t *testing.T) {
	unsafeID := "score\x1b]0;owned\x07\u202Edone"
	manifest := map[string]any{
		"version": 1,
		"name":    "unsafe-flow",
		"steps": []map[string]any{
			{
				"id":   unsafeID,
				"kind": "query",
				"sql":  "SELECT * FROM customers",
				"in":   []string{},
				"out":  []string{"scored_customers"},
			},
		},
	}
	data, err := json.Marshal(manifest)
	require.NoError(t, err)
	path := writeManifestFile(t, string(data))

	var human bytes.Buffer
	err = runFlow(testCtx(t), config.Config{}, nil, []string{"plan", "--file", path}, &human, false)
	require.NoError(t, err)
	assertNoUnsafeTerminalControls(t, human.String())
	assert.NotContains(t, human.String(), unsafeID)

	var machine bytes.Buffer
	err = runFlow(testCtx(t), config.Config{}, nil, []string{"plan", "--file", path}, &machine, true)
	require.NoError(t, err)
	var out struct {
		Order []string `json:"order"`
	}
	require.NoError(t, json.Unmarshal(machine.Bytes(), &out))
	require.Equal(t, []string{unsafeID}, out.Order)
}

func TestFlowListSanitizesUnsafeFilenamesInHumanOutput(t *testing.T) {
	dir := t.TempDir()
	unsafeName := "nightly\x1b]2;owned\x07\u202Eflow"
	err := os.WriteFile(filepath.Join(dir, unsafeName+".json"), []byte(`{"version":1}`), 0o644)
	require.NoError(t, err)

	var human bytes.Buffer
	err = runFlow(testCtx(t), config.Config{}, nil, []string{"list", "--flows-dir", dir}, &human, false)
	require.NoError(t, err)
	assertNoUnsafeTerminalControls(t, human.String())
	assert.NotContains(t, human.String(), unsafeName)

	var machine bytes.Buffer
	err = runFlow(testCtx(t), config.Config{}, nil, []string{"list", "--flows-dir", dir}, &machine, true)
	require.NoError(t, err)
	var out struct {
		Flows []string `json:"flows"`
	}
	require.NoError(t, json.Unmarshal(machine.Bytes(), &out))
	require.Equal(t, []string{unsafeName}, out.Flows)
}

func assertNoUnsafeTerminalControls(t *testing.T, text string) {
	t.Helper()
	for _, r := range text {
		if r == '\n' || r == '\t' {
			continue
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) || safety.IsDangerousUnicode(r) {
			t.Fatalf("human output contains unsafe terminal rune %U in %q", r, text)
		}
	}
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
