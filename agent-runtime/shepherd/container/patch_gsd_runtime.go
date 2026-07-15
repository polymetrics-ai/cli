package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type replacement struct {
	before string
	after  string
}

func main() {
	if len(os.Args) != 2 {
		fatal(errors.New("package root is required"))
	}
	root := os.Args[1]
	if err := patchExact(filepath.Join(root, "dist", "headless.js"),
		"d5d007f367a4008204ac43199a218d0ec12fac5ad63f0a4f85e594e0aa376fa2",
		"07de6ec472df1b610f6aaea8d453dbd451b73a8b3da34da8502289e1c5b6384f",
		[]replacement{
			{before: "    const interactiveToolCallIds = new Set();", after: "    const interactiveToolCallIds = new Set();\n    const inFlightToolCallIds = new Set();"},
			{before: "shouldArmHeadlessIdleTimeout(toolCallCount, interactiveToolCallIds.size, isQuickCmd)", after: "shouldArmHeadlessIdleTimeout(toolCallCount, Math.max(interactiveToolCallIds.size, inFlightToolCallIds.size), isQuickCmd)"},
			{before: `            if (toolCallId && isInteractiveHeadlessTool(String(eventObj.toolName ?? ''))) {
                interactiveToolCallIds.add(toolCallId);
            }`, after: `            if (toolCallId) {
                inFlightToolCallIds.add(toolCallId);
                if (isInteractiveHeadlessTool(String(eventObj.toolName ?? ''))) {
                    interactiveToolCallIds.add(toolCallId);
                }
            }`},
			{before: `            if (toolCallId) {
                interactiveToolCallIds.delete(toolCallId);
            }`, after: `            if (toolCallId) {
                inFlightToolCallIds.delete(toolCallId);
                interactiveToolCallIds.delete(toolCallId);
            }`},
		}); err != nil {
		fatal(err)
	}
	if err := patchExact(filepath.Join(root, "dist", "resources", "extensions", "gsd", "unit-context-composer.js"),
		"e286aedcbf4a3c22cbeae66e59851023ca1f3f7be38eb38f15556d4de481f2c7",
		"a43406c4b532f3a817dffa63dc6381329f098b2e8110b71f8d9b7e66a81c3b9e",
		[]replacement{{
			before: "Use `gsd_resume` for planning continuity, `gsd_exec` for noisy checks, and `gsd_exec_search` before rerunning diagnostics.",
			after:  "Use only the phase-scoped planning tools exposed for this unit (`gsd_milestone_status`, `gsd_plan_milestone`, `gsd_plan_slice`, `gsd_plan_task`, `gsd_requirement_update`, and `gsd_decision_save`). Do not call `gsd_resume`, `gsd_exec`, or `gsd_exec_search` from a planning unit.",
		}}); err != nil {
		fatal(err)
	}
}

func patchExact(path, originalHash, patchedHash string, replacements []replacement) error {
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() > 2*1024*1024 {
		return fmt.Errorf("unsafe runtime patch source %s", filepath.Base(path))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	observed := digest(raw)
	if observed == patchedHash {
		return nil
	}
	if observed != originalHash {
		return fmt.Errorf("runtime patch source drift: %s", filepath.Base(path))
	}
	content := string(raw)
	for _, change := range replacements {
		if strings.Count(content, change.before) != 1 {
			return fmt.Errorf("runtime patch shape drift: %s", filepath.Base(path))
		}
		content = strings.Replace(content, change.before, change.after, 1)
	}
	updated := []byte(content)
	if digest(updated) != patchedHash {
		return fmt.Errorf("runtime patched digest drift: %s", filepath.Base(path))
	}
	return os.WriteFile(path, updated, info.Mode().Perm())
}

func digest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
