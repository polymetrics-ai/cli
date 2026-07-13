package gsd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	headlessInteractiveSet = "    const interactiveToolCallIds = new Set();"
	headlessAllToolsSet    = headlessInteractiveSet + "\n    const inFlightToolCallIds = new Set();"
	headlessIdleOriginal   = "shouldArmHeadlessIdleTimeout(toolCallCount, interactiveToolCallIds.size, isQuickCmd)"
	headlessIdlePatched    = "shouldArmHeadlessIdleTimeout(toolCallCount, Math.max(interactiveToolCallIds.size, inFlightToolCallIds.size), isQuickCmd)"
	headlessStartOriginal  = `            if (toolCallId && isInteractiveHeadlessTool(String(eventObj.toolName ?? ''))) {
                interactiveToolCallIds.add(toolCallId);
            }`
	headlessStartPatched = `            if (toolCallId) {
                inFlightToolCallIds.add(toolCallId);
                if (isInteractiveHeadlessTool(String(eventObj.toolName ?? ''))) {
                    interactiveToolCallIds.add(toolCallId);
                }
            }`
	headlessEndOriginal = `            if (toolCallId) {
                interactiveToolCallIds.delete(toolCallId);
            }`
	headlessEndPatched = `            if (toolCallId) {
                inFlightToolCallIds.delete(toolCallId);
                interactiveToolCallIds.delete(toolCallId);
            }`
)

// ApplyPinnedHeadlessToolPatch fixes the official 1.11.0 headless idle timer,
// which otherwise declares completion while a non-interactive tool is still
// running. The patch is exact, local to the pinned package, and idempotent.
func ApplyPinnedHeadlessToolPatch(command []string, expectedVersion string) error {
	if expectedVersion != "1.11.0" {
		return errors.New("headless in-flight tool patch is qualified only for GSD 1.11.0")
	}
	if err := ValidatePinnedCommand(command, expectedVersion); err != nil {
		return err
	}
	loader, err := filepath.EvalSymlinks(command[1])
	if err != nil {
		return err
	}
	path := filepath.Join(filepath.Dir(loader), "headless.js")
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect pinned headless runtime: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return errors.New("pinned headless runtime must be a regular file")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read pinned headless runtime: %w", err)
	}
	content := string(raw)
	if strings.Contains(content, "inFlightToolCallIds.add(toolCallId)") {
		for _, marker := range []string{headlessAllToolsSet, headlessIdlePatched, headlessStartPatched, headlessEndPatched} {
			if strings.Count(content, marker) != 1 {
				return errors.New("installed headless compatibility patch has unexpected shape")
			}
		}
		return nil
	}
	replacements := [][2]string{
		{headlessInteractiveSet, headlessAllToolsSet},
		{headlessIdleOriginal, headlessIdlePatched},
		{headlessStartOriginal, headlessStartPatched},
		{headlessEndOriginal, headlessEndPatched},
	}
	for _, replacement := range replacements {
		if strings.Count(content, replacement[0]) != 1 {
			return errors.New("pinned headless runtime does not match the qualified 1.11.0 patch shape")
		}
		content = strings.Replace(content, replacement[0], replacement[1], 1)
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".headless-patch-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(info.Mode().Perm()); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.WriteString(content); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("install pinned headless compatibility patch: %w", err)
	}
	return nil
}
