package vault_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoSessionExportSymbol enforces "No session export, ever" (F15 §5.3):
// there must be no exported function anywhere in internal/vault whose name
// suggests it hands raw session/credential material back out in bulk
// (Export, Dump, Backup). "Backup" means "re-authenticate", never "copy the
// vault's plaintext contents somewhere else" — this is a mechanical guard,
// not a comment promise, so a future edit that adds such a function fails
// the build instead of silently shipping.
func TestNoSessionExportSymbol(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)^func\s+(\([^)]*\)\s+)?(Export|Dump|Backup)[A-Za-z0-9_]*\s*\(`)

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(raw), "\n") {
			trimmed := strings.TrimSpace(line)
			if forbidden.MatchString(trimmed) {
				t.Errorf("%s: forbidden session-export-shaped exported function: %q", path, trimmed)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/vault: %v", err)
	}
}
