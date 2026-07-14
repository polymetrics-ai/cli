package gsd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSubagentProgressAndOrphanReconciliation(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	work := filepath.Join(t.TempDir(), "issue-worktree")
	if err := os.MkdirAll(work, 0o700); err != nil {
		t.Fatal(err)
	}
	directory := filepath.Join(home, "agent", "subagent-runs")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "run-1.json")
	run := map[string]any{
		"schemaVersion": 1, "runId": "run-1", "status": "running", "cwd": filepath.Join(work, ".gsd-worktrees", "M001"),
		"startedAt": "2026-07-14T12:50:48.729Z", "updatedAt": "2026-07-14T12:54:38.942Z",
		"children": []any{map[string]any{
			"index": 0, "trackingName": "swift-lantern", "status": "running",
			"usage": map[string]any{"turns": 26, "input": 86832, "output": 6027},
		}},
	}
	raw, err := json.Marshal(run)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	progress, err := ReadSubagentProgress(home, work)
	if err != nil {
		t.Fatal(err)
	}
	if progress.RunningChildren != 1 || progress.Turns != 26 || progress.Status != "running" {
		t.Fatalf("progress=%+v", progress)
	}

	reconciled, err := ReconcileOrphanedSubagents(home, work, time.Date(2026, 7, 14, 13, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if reconciled.InterruptedRuns != 1 || reconciled.InterruptedChildren != 1 {
		t.Fatalf("reconciled=%+v", reconciled)
	}
	progress, err = ReadSubagentProgress(home, work)
	if err != nil {
		t.Fatal(err)
	}
	if progress.RunningChildren != 0 || progress.Status != "interrupted" {
		t.Fatalf("post-reconcile progress=%+v", progress)
	}
}

func TestOrphanReconciliationDoesNotTouchAnotherIssue(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	work := filepath.Join(t.TempDir(), "issue-a")
	other := filepath.Join(t.TempDir(), "issue-b")
	for _, path := range []string{work, other, filepath.Join(home, "agent", "subagent-runs")} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(home, "agent", "subagent-runs", "other.json")
	raw := []byte(`{"schemaVersion":1,"runId":"other","status":"running","cwd":"` + other + `","children":[]}`)
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := ReconcileOrphanedSubagents(home, work, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	if result.InterruptedRuns != 0 {
		t.Fatalf("other issue was reconciled: %+v", result)
	}
}
