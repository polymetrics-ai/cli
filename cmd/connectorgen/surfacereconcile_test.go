package main

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const reconcileFixtureMethod = "POST"
const reconcileFixturePath = "/v2/meetings/integration/status"

func TestRunSurfaceReconcileRequiresRuntimePreflightForCoverage(t *testing.T) {
	root, surfacePath := writeSurfaceReconcileFixture(t, "implemented", "direct_read")
	before, err := os.ReadFile(surfacePath)
	if err != nil {
		t.Fatalf("read fixture before check: %v", err)
	}

	var stdout, stderr bytes.Buffer
	if code := run([]string{"surface-reconcile", root, "--check"}, &stdout, &stderr); code != 1 {
		t.Fatalf("surface-reconcile --check exit = %d, want 1; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	if !strings.Contains(stdout.String(), "gong: would reconcile covered=1") {
		t.Fatalf("surface-reconcile --check stdout = %q, want runtime-covered result", stdout.String())
	}
	afterCheck, err := os.ReadFile(surfacePath)
	if err != nil {
		t.Fatalf("read fixture after check: %v", err)
	}
	if !bytes.Equal(before, afterCheck) {
		t.Fatal("surface-reconcile --check wrote api_surface.json")
	}

	stdout.Reset()
	stderr.Reset()
	if code := run([]string{"surface-reconcile", root}, &stdout, &stderr); code != 0 {
		t.Fatalf("surface-reconcile write exit = %d, want 0; stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	endpoint := readSurfaceReconcileEndpoint(t, surfacePath)
	if endpoint.Operation != nil {
		t.Fatalf("reconciled endpoint still has operation = %+v", endpoint.Operation)
	}
	if endpoint.CoveredBy == nil || endpoint.CoveredBy.DirectRead != "meetings integration-status" {
		t.Fatalf("reconciled covered_by = %+v, want runtime-reachable command", endpoint.CoveredBy)
	}
}

func TestRunSurfaceReconcileHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"surface-reconcile", "--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("surface-reconcile --help exit = %d, want 0; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Derive direct-read api_surface coverage") {
		t.Fatalf("surface-reconcile --help stdout = %q, want command description", stdout.String())
	}
}

func TestSurfaceReconcileKeepsUnreachableRowsBlockedAndRefusesUnknownModel(t *testing.T) {
	t.Run("declared but planned command", func(t *testing.T) {
		root, surfacePath := writeSurfaceReconcileFixture(t, "planned", "direct_read")
		stats, err := reconcileBundle(filepath.Join(root, "gong"), false, "")
		if err != nil {
			t.Fatalf("reconcileBundle: %v", err)
		}
		if stats.Covered != 0 || stats.Blocked != 1 || stats.Refused != 0 {
			t.Fatalf("stats = %+v, want one blocked reason update", stats)
		}
		endpoint := readSurfaceReconcileEndpoint(t, surfacePath)
		if endpoint.Operation == nil || !strings.Contains(endpoint.Operation.Reason, "declared planned, not implemented") {
			t.Fatalf("blocked endpoint = %+v, want current planned-command reason", endpoint)
		}
		if endpoint.CoveredBy != nil {
			t.Fatalf("planned command produced coverage: %+v", endpoint.CoveredBy)
		}
	})

	t.Run("implemented command rejected by runtime preflight", func(t *testing.T) {
		root, surfacePath := writeSurfaceReconcileFixture(t, "implemented", "direct_read", "gong.not_declared")
		stats, err := reconcileBundle(filepath.Join(root, "gong"), false, "")
		if err != nil {
			t.Fatalf("reconcileBundle: %v", err)
		}
		if stats.Covered != 0 || stats.Blocked != 1 || stats.Refused != 0 {
			t.Fatalf("stats = %+v, want one preflight-derived blocked reason", stats)
		}
		endpoint := readSurfaceReconcileEndpoint(t, surfacePath)
		if endpoint.Operation == nil || !strings.Contains(endpoint.Operation.Reason, "fails runtime preflight") || endpoint.CoveredBy != nil {
			t.Fatalf("preflight-rejected endpoint = %+v, want blocked operation reason", endpoint)
		}
	})

	t.Run("unsupported operation model", func(t *testing.T) {
		root, surfacePath := writeSurfaceReconcileFixture(t, "implemented", "binary_read")
		before, err := os.ReadFile(surfacePath)
		if err != nil {
			t.Fatalf("read fixture before refusal: %v", err)
		}
		stats, err := reconcileBundle(filepath.Join(root, "gong"), false, "")
		if err != nil {
			t.Fatalf("reconcileBundle: %v", err)
		}
		if stats.Covered != 0 || stats.Blocked != 0 || stats.Refused != 1 {
			t.Fatalf("stats = %+v, want one refusal", stats)
		}
		after, err := os.ReadFile(surfacePath)
		if err != nil {
			t.Fatalf("read fixture after refusal: %v", err)
		}
		if !bytes.Equal(before, after) {
			t.Fatal("unsupported model was rewritten instead of refused")
		}
	})
}

type reconcileFixtureEndpoint struct {
	Method    string `json:"method"`
	Path      string `json:"path"`
	Operation *struct {
		Reason string `json:"reason"`
	} `json:"operation"`
	CoveredBy *struct {
		DirectRead string `json:"direct_read"`
	} `json:"covered_by"`
}

func writeSurfaceReconcileFixture(t *testing.T, availability, model string, commandOperation ...string) (string, string) {
	t.Helper()
	root := t.TempDir()
	sourceRoot, err := repoRoot()
	if err != nil {
		t.Fatalf("repoRoot: %v", err)
	}
	source := filepath.Join(sourceRoot, "internal", "connectors", "defs", "gong")
	target := filepath.Join(root, "gong")
	if err := copySurfaceReconcileTree(source, target); err != nil {
		t.Fatalf("copy Gong fixture: %v", err)
	}

	surfacePath := filepath.Join(target, "api_surface.json")
	var surface map[string]any
	raw, err := os.ReadFile(surfacePath)
	if err != nil {
		t.Fatalf("read fixture api surface: %v", err)
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("decode fixture api surface: %v", err)
	}
	endpoints, ok := surface["endpoints"].([]any)
	if !ok {
		t.Fatal("fixture api surface endpoints are not an array")
	}
	found := false
	for _, rawEndpoint := range endpoints {
		endpoint, ok := rawEndpoint.(map[string]any)
		if !ok {
			t.Fatal("fixture api surface endpoint is not an object")
		}
		if endpoint["method"] != reconcileFixtureMethod || endpoint["path"] != reconcileFixturePath {
			continue
		}
		delete(endpoint, "covered_by")
		endpoint["operation"] = map[string]any{
			"model":              model,
			"status":             "blocked",
			"risk":               "low",
			"blocked_by_default": true,
			"reason":             "Blocked by missing shared foundation #2985; stale fixture prose.",
		}
		found = true
	}
	if !found {
		t.Fatalf("Gong fixture lacks %s %s", reconcileFixtureMethod, reconcileFixturePath)
	}
	writeSurfaceReconcileJSON(t, surfacePath, surface)

	cliPath := filepath.Join(target, "cli_surface.json")
	var cli map[string]any
	raw, err = os.ReadFile(cliPath)
	if err != nil {
		t.Fatalf("read fixture cli surface: %v", err)
	}
	if err := json.Unmarshal(raw, &cli); err != nil {
		t.Fatalf("decode fixture cli surface: %v", err)
	}
	commands, ok := cli["commands"].([]any)
	if !ok {
		t.Fatal("fixture cli surface commands are not an array")
	}
	for _, rawCommand := range commands {
		command, ok := rawCommand.(map[string]any)
		if !ok {
			t.Fatal("fixture cli surface command is not an object")
		}
		if command["path"] == "meetings integration-status" {
			command["availability"] = availability
			if len(commandOperation) > 0 {
				command["operation"] = commandOperation[0]
			}
			writeSurfaceReconcileJSON(t, cliPath, cli)
			return root, surfacePath
		}
	}
	t.Fatalf("Gong fixture lacks meetings integration-status command")
	return "", ""
}

func copySurfaceReconcileTree(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		destination := filepath.Join(target, rel)
		if entry.IsDir() {
			return os.MkdirAll(destination, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(destination, data, 0o644)
	})
}

func writeSurfaceReconcileJSON(t *testing.T, path string, value any) {
	t.Helper()
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	if err := os.WriteFile(path, append(raw, '\n'), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func readSurfaceReconcileEndpoint(t *testing.T, path string) reconcileFixtureEndpoint {
	t.Helper()
	var surface struct {
		Endpoints []reconcileFixtureEndpoint `json:"endpoints"`
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, &surface); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	for _, endpoint := range surface.Endpoints {
		if endpoint.Method == reconcileFixtureMethod && endpoint.Path == reconcileFixturePath {
			return endpoint
		}
	}
	t.Fatalf("endpoint %s %s not found", reconcileFixtureMethod, reconcileFixturePath)
	return reconcileFixtureEndpoint{}
}
