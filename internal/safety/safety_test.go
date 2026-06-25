package safety_test

import (
	"strings"
	"testing"

	"polymetrics/internal/safety"
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
