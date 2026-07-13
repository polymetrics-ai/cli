package gsd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateContainerVersionOutputRequiresExactPinnedVersion(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		output  string
		wantErr bool
	}{
		{name: "real GSD output", output: "1.11.0\n"},
		{name: "wrong version", output: "1.12.0\n", wantErr: true},
		{name: "substring spoof", output: "untrusted 1.11.0 output\n", wantErr: true},
		{name: "legacy decoration", output: "GSD v1.11.0\n", wantErr: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			err := validateContainerVersionOutput(test.output, "1.11.0")
			if (err != nil) != test.wantErr {
				t.Fatalf("validateContainerVersionOutput() error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

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
	if err := ValidateRuntimeSettings(home, t.TempDir(), "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("expected thinking mismatch to fail")
	}
	if err := os.WriteFile(path, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntimeSettings(home, t.TempDir(), "openai-codex/gpt-5.6-sol", "high"); err != nil {
		t.Fatalf("valid settings rejected: %v", err)
	}
}

func TestValidateRuntimeSettingsRejectsProjectOverride(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	agent := filepath.Join(home, "agent")
	if err := os.Mkdir(agent, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(agent, "settings.json"), []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	work := t.TempDir()
	if err := os.Mkdir(filepath.Join(work, ".pi"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(work, ".pi", "settings.json"), []byte(`{"defaultThinkingLevel":"off"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntimeSettings(home, work, "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("expected project override to fail admission")
	}
}

func TestApplyPinnedHeadlessToolPatchIsExactAndIdempotent(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dist := filepath.Join(root, "dist")
	if err := os.Mkdir(dist, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"@opengsd/gsd-pi","version":"1.11.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loader := filepath.Join(dist, "loader.js")
	if err := os.WriteFile(loader, []byte("loader"), 0o600); err != nil {
		t.Fatal(err)
	}
	original := `    const interactiveToolCallIds = new Set();
            shouldArmHeadlessIdleTimeout(toolCallCount, interactiveToolCallIds.size, isQuickCmd)) {
            if (toolCallId && isInteractiveHeadlessTool(String(eventObj.toolName ?? ''))) {
                interactiveToolCallIds.add(toolCallId);
            }
            if (toolCallId) {
                interactiveToolCallIds.delete(toolCallId);
            }
`
	path := filepath.Join(dist, "headless.js")
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	command := []string{"node", loader}
	if err := ApplyPinnedHeadlessToolPatch(command, "1.11.0"); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "inFlightToolCallIds.add(toolCallId)") {
		t.Fatalf("patch missing: %s", first)
	}
	if err := ApplyPinnedHeadlessToolPatch(command, "1.11.0"); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Fatal("idempotent patch changed bytes")
	}
}
