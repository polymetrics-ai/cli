package browserauth_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPackageNeverParsesArgv is half of the guard against the plan's
// explicit lesson from bcharleson/linkedincli's own defect ("--li-at
// <cookie> / --jsessionid <cookie> put a live session token in shell
// history and in ps output"). This half proves the structural claim: no
// file under internal/browserauth (this package and its loopback/device/
// driver/store subpackages) imports "flag" or reads os.Args, so this
// library has zero capability to consume a command-line argument at all —
// a credential can only enter it as a Flow.Login() return value or as a Go
// struct field the CALLER sets in code, never as parsed CLI text. The other
// half of the guard, which asserts the actual flag-consumption boundary
// (internal/cli, where pm's --key value flags are read) never reads a
// secret-shaped flag name, lives in
// internal/cli/no_argv_credentials_internal_test.go — this file alone
// cannot prove that half because this package contains no flag parsing to
// inspect.
func TestPackageNeverParsesArgv(t *testing.T) {
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		src := string(raw)
		if strings.Contains(src, `"flag"`) {
			t.Errorf("%s: imports the \"flag\" package — this library must never parse CLI arguments itself", path)
		}
		if strings.Contains(src, "os.Args") {
			t.Errorf("%s: reads os.Args — this library must never parse CLI arguments itself", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/browserauth: %v", err)
	}
}
