package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestNoSecretShapedFlagIsEverRead is the other half of the guard against
// bcharleson/linkedincli's own defect ("--li-at <cookie> / --jsessionid
// <cookie> put a live session token in shell history and in ps output").
// internal/browserauth/no_argv_credentials_test.go proves that library has
// no argv-parsing capability at all; this half proves the boundary at the
// ONE place in this codebase where a --key value flag's value is actually
// read: parsedFlags.first (parse.go). pm's flag parser (parseFlags) is
// fully generic — it accepts any --key value pair with no allow-list — so
// the only enforceable guarantee is that no call site ever reads a flag
// name shaped like a raw credential value. The forbidden list below is the
// exact session material the dual-mechanism connector plan names per
// provider (report §3.4: reddit_session/modhash, auth_token/ct0,
// li_at/JSESSIONID) plus the generic OAuth/session-secret vocabulary, so a
// future `pm auth login <x>-web --li-at <cookie>`-shaped flag (the literal
// bug being guarded against) fails this test before it ever ships.
//
// Legitimate flags that read a NAME or REFERENCE, never the secret value
// itself, are unaffected: --credential <name>, --credentials-file <path>,
// --approve/--confirm <challenge-or-token-derived-from-plan-state>, and
// this PR's own --accept-risk <sha256-of-public-warning-text> all stay
// clear of every forbidden name below.
func TestNoSecretShapedFlagIsEverRead(t *testing.T) {
	forbidden := regexp.MustCompile(`(?i)\.first\(\s*"(` + strings.Join([]string{
		"password",
		"li_at", "li-at",
		"jsessionid",
		"ct0",
		"auth_token", "auth-token",
		"reddit_session",
		"modhash",
		"session_token", "session-token",
		"access_token", "access-token",
		"refresh_token", "refresh-token",
		"client_secret", "client-secret",
		"cookie",
		"csrf_token", "csrf-token",
	}, "|") + `)"\s*\)`)

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
		for _, line := range strings.Split(string(raw), "\n") {
			code, _, _ := strings.Cut(line, "//")
			if loc := forbidden.FindString(code); loc != "" {
				t.Errorf("%s: reads a secret-shaped flag value %q — credentials must arrive via env var or stdin, never a CLI argument (AGENTS.md; the bcharleson/linkedincli --li-at defect)", path, loc)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/cli: %v", err)
	}
}
