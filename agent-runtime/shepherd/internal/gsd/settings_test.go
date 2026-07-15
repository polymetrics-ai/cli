package gsd

import (
	"encoding/json"
	"errors"
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

func TestValidateRuntimeSettingsRejectsSymlinkedAndOversizedPolicy(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	work := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, "agent"), 0o700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, "settings-target.json")
	if err := os.WriteFile(target, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(home, "agent", "settings.json")); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntimeSettings(home, work, "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("symlinked runtime policy was accepted")
	}
	if err := os.Remove(filepath.Join(home, "agent", "settings.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "agent", "settings.json"), make([]byte, 1024*1024+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRuntimeSettings(home, work, "openai-codex/gpt-5.6-sol", "high"); err == nil {
		t.Fatal("oversized runtime policy was accepted")
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

func TestNormalizeRuntimeSettingsRestoresOnlyGovernedImplementationDrift(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	agent := filepath.Join(home, "agent")
	if err := os.Mkdir(agent, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(agent, "settings.json")
	raw := `{"defaultProvider":"openai-codex","defaultModel":"gpt-5.5","defaultThinkingLevel":"high","quietStartup":true}`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := NormalizeRuntimeSettings(home, "openai-codex/gpt-5.6-sol", "openai-codex/gpt-5.5", "high"); err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(updated, &settings); err != nil {
		t.Fatal(err)
	}
	if settings["defaultModel"] != "gpt-5.6-sol" || settings["quietStartup"] != true {
		t.Fatalf("settings not restored with unrelated fields preserved: %s", updated)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("settings mode=%o, want 600", info.Mode().Perm())
	}
}

func TestNormalizeRuntimeSettingsRejectsUnknownIdentity(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	agent := filepath.Join(home, "agent")
	if err := os.Mkdir(agent, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(agent, "settings.json")
	if err := os.WriteFile(path, []byte(`{"defaultProvider":"other","defaultModel":"unknown","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := NormalizeRuntimeSettings(home, "openai-codex/gpt-5.6-sol", "openai-codex/gpt-5.5", "high"); err == nil {
		t.Fatal("unknown runtime identity was rewritten instead of rejected")
	}
}

func TestValidateModelPreferencesRequiresGovernedEffectivePhaseRouting(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	work := t.TempDir()
	preferences := `---
version: 1
models:
  research: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  planning:
    provider: openai-codex
    model: gpt-5.6-sol
    thinking: high
  discuss: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }
  execution_simple: { provider: openai-codex, model: gpt-5.5, thinking: high }
  completion: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  validation: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  subagent: { provider: openai-codex, model: gpt-5.5, thinking: high }
  uat: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
skill_discovery: suggest
---
`
	path := filepath.Join(home, "PREFERENCES.md")
	if err := os.WriteFile(path, []byte(preferences), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateModelPreferences(home, work, mustDecodeRegistryFixture(t), "openai-codex/gpt-5.6-sol", "openai-codex/gpt-5.5", "high"); err != nil {
		t.Fatalf("valid phase routing rejected: %v", err)
	}

	drifted := strings.Replace(preferences, "model: gpt-5.5, thinking: high", "model: gpt-5.6-sol, thinking: high", 1)
	if err := os.WriteFile(path, []byte(drifted), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateModelPreferences(home, work, mustDecodeRegistryFixture(t), "openai-codex/gpt-5.6-sol", "openai-codex/gpt-5.5", "high"); err == nil {
		t.Fatal("execution model drift accepted")
	}
}

func TestValidateModelPreferencesRejectsMissingMalformedAndDuplicatePolicy(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		raw  string
	}{
		{name: "missing file"},
		{name: "missing frontmatter close", raw: "---\nversion: 1\nmodels: {}\n"},
		{name: "duplicate phase", raw: "---\nmodels:\n  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }\n  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }\n---\n"},
		{name: "duplicate models after field", raw: "---\nmodels:\n  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }\nversion: 1\nmodels:\n  planning: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n---\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			work := t.TempDir()
			if test.raw != "" {
				if err := os.WriteFile(filepath.Join(home, "PREFERENCES.md"), []byte(test.raw), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			if err := ValidateModelPreferences(home, work, mustDecodeRegistryFixture(t), "openai-codex/gpt-5.6-sol", "openai-codex/gpt-5.5", "high"); err == nil {
				t.Fatal("invalid phase policy accepted")
			}
		})
	}
}

func TestValidateModelPreferencesRejectsProjectPhaseOverride(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	work := t.TempDir()
	global := `---
models:
  research: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  planning: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  discuss: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }
  execution_simple: { provider: openai-codex, model: gpt-5.5, thinking: high }
  completion: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  validation: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  subagent: { provider: openai-codex, model: gpt-5.5, thinking: high }
  uat: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
---
`
	if err := os.WriteFile(filepath.Join(home, "PREFERENCES.md"), []byte(global), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(work, ".gsd"), 0o700); err != nil {
		t.Fatal(err)
	}
	project := `---
models:
  execution: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
---
`
	if err := os.WriteFile(filepath.Join(work, ".gsd", "PREFERENCES.md"), []byte(project), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := ValidateModelPreferences(home, work, mustDecodeRegistryFixture(t), "openai-codex/gpt-5.6-sol", "openai-codex/gpt-5.5", "high"); err == nil {
		t.Fatal("conflicting project phase override accepted")
	}
}

func TestApplyPinnedHeadlessToolPatchReportsContractMismatch(t *testing.T) {
	t.Parallel()
	if err := ApplyPinnedHeadlessToolPatch([]string{"node", filepath.Join(t.TempDir(), "dist", "loader.js")}, "1.12.0"); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want runtime contract mismatch", err)
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
