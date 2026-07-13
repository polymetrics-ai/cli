package gsd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateRuntimeSettingsFailsClosedOnThinkingMismatch(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	agent := filepath.Join(home, "agent")
	if err := os.Mkdir(agent, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(agent, "settings.json")
	if err := os.WriteFile(path, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"off"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntimeSettings(home, "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("expected thinking mismatch to fail")
	}
	if err := os.WriteFile(path, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntimeSettings(home, "openai-codex/gpt-5.6-sol", "high"); err != nil {
		t.Fatalf("valid settings rejected: %v", err)
	}
}
