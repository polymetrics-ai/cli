package safety_test

import (
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

func TestRedactErrorTextRemovesHTTPURLQueryAndBodySecrets(t *testing.T) {
	input := `http 401 for https://api.example.test/v1/items?api_key=secret-token&cursor=abc: {"error":"secret-token denied","access_token":"abc"}`
	got := safety.RedactErrorText(input)
	for _, leaked := range []string{"secret-token", "api_key=", "access_token", "cursor=abc"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("RedactErrorText leaked %q in %q", leaked, got)
		}
	}
	if !strings.Contains(got, "https://api.example.test/v1/items") {
		t.Fatalf("RedactErrorText removed useful URL context: %q", got)
	}
}
