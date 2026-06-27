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
	err := runFlow(testCtx(t), nil, []string{"list", "--flows-dir", dir}, &out, true)
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
	err := runFlow(testCtx(t), nil, []string{"plan", "--file", path}, &out, true)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	assert.Equal(t, "ok", result["status"])
}

// TestFlowPlanCyclic checks that `pm flow plan --file cyclic.json` returns a non-nil error.
func TestFlowPlanCyclic(t *testing.T) {
	path := writeManifestFile(t, cyclicFlowManifestJSON)
	var out bytes.Buffer
	err := runFlow(testCtx(t), nil, []string{"plan", "--file", path}, &out, true)
	require.Error(t, err, "cyclic manifest should produce an error")
	assert.True(t, strings.Contains(err.Error(), "cyclic") || strings.Contains(err.Error(), "flow:"),
		"error should mention cycle: %v", err)
}

// TestFlowStatusMissing checks that `pm flow status <missing>` returns an error.
func TestFlowStatusMissing(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	err := runFlow(testCtx(t), nil, []string{"status", "nonexistent", "--flows-dir", dir}, &out, true)
	require.Error(t, err)
}

// TestFlowPreviewValid checks that `pm flow preview --file valid.json` returns dry_run status.
func TestFlowPreviewValid(t *testing.T) {
	path := writeManifestFile(t, validFlowManifestJSON)
	var out bytes.Buffer
	err := runFlow(testCtx(t), nil, []string{"preview", "--file", path}, &out, true)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	assert.Equal(t, "dry_run", result["status"])
}
