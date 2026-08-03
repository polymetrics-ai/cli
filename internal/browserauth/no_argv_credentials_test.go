package browserauth_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoCredentialFlagsInPackageSource guards the plan's explicit lesson
// from bcharleson/linkedincli's own defect ("--li-at <cookie> /
// --jsessionid <cookie> put a live session token in shell history and in ps
// output"): browserauth and its subpackages must never define a flag/field
// shaped like a raw credential value meant to arrive as a CLI argument.
// Credentials enter this package only as the return value of a Flow's
// Login() (browser-driven capture) or as fields the caller sets in Go code
// (loopback.Config.ClientSecret, say, which a CLI layer must source from an
// env var or stdin per AGENTS.md, never from bare argv text) — never as a
// struct field literally named/tagged for a command-line flag that accepts
// a session token or password.
func TestNoCredentialFlagsInPackageSource(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)flag\.\w+\(\s*"(password|li_at|li-at|jsessionid|session-token|session_token|cookie)"`)

	err := filepath.WalkDir(".", func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, line := range strings.Split(string(raw), "\n") {
			code, _, _ := strings.Cut(line, "//")
			if loc := forbidden.FindString(code); loc != "" {
				t.Errorf("%s: forbidden argv-credential flag definition: %q", path, loc)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/browserauth: %v", err)
	}
}
