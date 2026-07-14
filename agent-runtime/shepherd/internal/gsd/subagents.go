package gsd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxSubagentRunBytes = 1024 * 1024
const maxSubagentRunFiles = 1000

type SubagentProgress struct {
	Status          string
	RunningChildren int
	Turns           int64
	UpdatedAt       time.Time
}

type SubagentReconcileResult struct {
	InterruptedRuns     int
	InterruptedChildren int
}

type subagentRun struct {
	RunID     string          `json:"runId"`
	Status    string          `json:"status"`
	CWD       string          `json:"cwd"`
	UpdatedAt string          `json:"updatedAt"`
	Children  []subagentChild `json:"children"`
}

type subagentChild struct {
	Status string `json:"status"`
	Usage  struct {
		Turns int64 `json:"turns"`
	} `json:"usage"`
}

func ReadSubagentProgress(gsdHome, workDir string) (SubagentProgress, error) {
	files, err := scopedSubagentRunFiles(gsdHome, workDir)
	if err != nil {
		return SubagentProgress{}, err
	}
	progress := SubagentProgress{Status: "none"}
	for _, file := range files {
		run, _, _, err := readSubagentRun(file)
		if err != nil {
			return SubagentProgress{}, err
		}
		updated, _ := time.Parse(time.RFC3339Nano, run.UpdatedAt)
		if updated.After(progress.UpdatedAt) {
			progress.Status, progress.UpdatedAt = run.Status, updated
		}
		if run.Status != "running" {
			continue
		}
		for _, child := range run.Children {
			if child.Status == "running" {
				progress.RunningChildren++
				if child.Usage.Turns > 0 {
					progress.Turns += child.Usage.Turns
				}
			}
		}
	}
	if progress.RunningChildren > 0 {
		progress.Status = "running"
	}
	return progress, nil
}

// ReconcileOrphanedSubagents is called only while Shepherd owns the delivery
// lease and no GSD process is live. A persisted running child under that issue
// therefore has no owner and cannot be considered resumable.
func ReconcileOrphanedSubagents(gsdHome, workDir string, now time.Time) (SubagentReconcileResult, error) {
	files, err := scopedSubagentRunFiles(gsdHome, workDir)
	if err != nil {
		return SubagentReconcileResult{}, err
	}
	var result SubagentReconcileResult
	for _, file := range files {
		run, document, mode, err := readSubagentRun(file)
		if err != nil {
			return SubagentReconcileResult{}, err
		}
		if run.Status != "running" {
			continue
		}
		document["status"] = "interrupted"
		document["updatedAt"] = now.UTC().Format(time.RFC3339Nano)
		children, ok := document["children"].([]any)
		if !ok {
			return SubagentReconcileResult{}, errors.New("subagent run children are malformed")
		}
		for _, value := range children {
			child, ok := value.(map[string]any)
			if !ok {
				return SubagentReconcileResult{}, errors.New("subagent child is malformed")
			}
			if child["status"] == "running" {
				child["status"] = "interrupted"
				child["stopReason"] = "shepherd_reconcile_no_live_owner"
				result.InterruptedChildren++
			}
		}
		raw, err := json.MarshalIndent(document, "", "  ")
		if err != nil {
			return SubagentReconcileResult{}, err
		}
		if err := atomicReplaceRuntimeFile(file, append(raw, '\n'), mode); err != nil {
			return SubagentReconcileResult{}, err
		}
		result.InterruptedRuns++
	}
	return result, nil
}

func scopedSubagentRunFiles(gsdHome, workDir string) ([]string, error) {
	if !filepath.IsAbs(gsdHome) || !filepath.IsAbs(workDir) {
		return nil, errors.New("absolute GSD home and work directory are required")
	}
	directory := filepath.Join(gsdHome, "agent", "subagent-runs")
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(entries) > maxSubagentRunFiles {
		return nil, errors.New("subagent run directory exceeds the bounded file count")
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		path := filepath.Join(directory, entry.Name())
		run, _, _, err := readSubagentRun(path)
		if err != nil {
			return nil, err
		}
		within, err := pathWithin(workDir, run.CWD)
		if err != nil {
			return nil, err
		}
		if within {
			files = append(files, path)
		}
	}
	return files, nil
}

func readSubagentRun(path string) (subagentRun, map[string]any, os.FileMode, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return subagentRun{}, nil, 0, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() > maxSubagentRunBytes {
		return subagentRun{}, nil, 0, errors.New("subagent run must be a bounded regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return subagentRun{}, nil, 0, err
	}
	var run subagentRun
	var document map[string]any
	if err := json.Unmarshal(raw, &run); err != nil || json.Unmarshal(raw, &document) != nil ||
		run.RunID == "" || run.Status == "" || !filepath.IsAbs(run.CWD) {
		return subagentRun{}, nil, 0, fmt.Errorf("invalid subagent run %s", filepath.Base(path))
	}
	return run, document, info.Mode().Perm(), nil
}
