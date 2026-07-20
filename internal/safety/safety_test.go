package safety_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/safety"
)

func TestSanitizeTerminalStripsControlAndDangerousUnicode(t *testing.T) {
	input := "ok\x1b[31mred\x1b[0m\u202Eflip\u200Bzero\rreturn"
	got := safety.SanitizeTerminal(input)
	for _, bad := range []string{"\x1b", "\u202E", "\u200B", "\r"} {
		if strings.Contains(got, bad) {
			t.Fatalf("SanitizeTerminal retained %q in %q", bad, got)
		}
	}
	if !strings.Contains(got, "ok") || !strings.Contains(got, "red") {
		t.Fatalf("SanitizeTerminal removed normal text: %q", got)
	}
}

func TestValidateIdentifierRejectsAgentUnsafeValues(t *testing.T) {
	tests := []string{
		"",
		"../secret",
		"table name",
		"bad\x1b[31m",
		"\u202Espoof",
		"semi;colon",
	}
	for _, in := range tests {
		t.Run(in, func(t *testing.T) {
			if err := safety.ValidateIdentifier(in, "test"); err == nil {
				t.Fatalf("ValidateIdentifier(%q) succeeded, want error", in)
			}
		})
	}
}

func TestValidateIdentifierAcceptsExpectedNames(t *testing.T) {
	for _, in := range []string{"github", "sample_customers", "run-01", "table.name"} {
		if err := safety.ValidateIdentifier(in, "test"); err != nil {
			t.Fatalf("ValidateIdentifier(%q) error = %v", in, err)
		}
	}
}

func TestValidateLocalWritePathResolvesSymlinksThroughNearestExistingAncestor(t *testing.T) {
	root := t.TempDir()
	external := t.TempDir()
	redirect := filepath.Join(root, "redirect")
	if err := os.Symlink(external, redirect); err != nil {
		t.Skipf("symlinks unavailable on this platform: %v", err)
	}
	path := filepath.Join(redirect, "not-created", "warehouse")

	if err := safety.ValidateLocalWritePath(root, path, "warehouse path", false); err == nil {
		t.Fatal("symlink-resolved external path succeeded without explicit policy")
	}
	if err := safety.ValidateLocalWritePath(root, path, "warehouse path", true); err != nil {
		t.Fatalf("explicit external-path policy rejected symlink-resolved path: %v", err)
	}
	if _, err := os.Stat(filepath.Join(external, "not-created")); !os.IsNotExist(err) {
		t.Fatalf("path validation created an external filesystem effect: %v", err)
	}
}

func TestRedactErrorTextRemovesHTTPURLQueryAndBodySecrets(t *testing.T) {
	marker := "pm-test-safety-url-marker-404"
	input := `http 401 for https://api.example.test/v1/items?api_key=` + marker + `&cursor=abc: {"error":"` + marker + ` denied","access_token":"abc"}`
	got := safety.RedactErrorText(input)
	for _, leaked := range []string{marker, "api_key=", "access_token", "cursor=abc"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("RedactErrorText leaked sensitive URL/body component")
		}
	}
	if !strings.Contains(got, "https://api.example.test/v1/items") {
		t.Fatalf("RedactErrorText removed useful URL context")
	}
}

func TestRedactErrorTextCanonicalizesURLsCaseInsensitively(t *testing.T) {
	marker := "pm-test-safety-case-marker-404"
	input := "failed HTTPS://User:" + marker + "@API.Example.Test/v1/items?token=" + marker + "#frag and postgres://user:" + marker + "@db.example.test/app?sslmode=require"
	got := safety.RedactErrorText(input)
	for _, leaked := range []string{marker, "User:", "token=", "#frag", "sslmode=require"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("RedactErrorText leaked URL credential/query/fragment component")
		}
	}
	for _, want := range []string{"https://api.example.test/v1/items", "postgres://db.example.test/app"} {
		if !strings.Contains(got, want) {
			t.Fatalf("RedactErrorText missing sanitized URL context")
		}
	}
}

func TestRedactErrorTextFailsClosedForMalformedCredentialURLs(t *testing.T) {
	marker := "pm-test-safety-malformed-marker-404"
	inputs := []string{
		"dial HTTPS://user:" + marker + "@%zz.example.test/path?token=" + marker,
		"dial https://user:" + marker + "@/path?token=" + marker,
		"dial postgres://user:" + marker + "@%zz/db?password=" + marker,
	}
	for _, input := range inputs {
		got := safety.RedactErrorText(input)
		if strings.Contains(got, marker) || strings.Contains(got, "token=") || strings.Contains(got, "password=") || strings.Contains(got, "user:") {
			t.Fatalf("RedactErrorText returned raw malformed credential/query URL")
		}
		if !strings.Contains(got, "[redacted-url]") {
			t.Fatalf("RedactErrorText did not use fail-closed URL marker")
		}
	}
}

func TestRedactErrorTextDropsPercentEncodedQueryValues(t *testing.T) {
	marker := "pm test safety percent marker 404"
	input := "failed https://api.example.test/items?cursor=abc&token=pm+test+safety+percent+marker+404#frag"
	got := safety.RedactErrorText(input)
	for _, leaked := range []string{marker, "pm+test+safety+percent+marker+404", "cursor=abc", "token=", "#frag"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("RedactErrorText leaked percent-encoded query component")
		}
	}
	if !strings.Contains(got, "https://api.example.test/items") {
		t.Fatalf("RedactErrorText missing URL path context")
	}
}
