package gsd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePromptToolContractRejectsAdvertisedForbiddenTool(t *testing.T) {
	t.Parallel()

	err := ValidatePromptToolContract("plan-milestone",
		[]string{"gsd_milestone_status", "gsd_plan_milestone"},
		[]string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_resume"},
	)
	if !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want runtime contract mismatch", err)
	}
}

func TestPinnedPromptToolPatchRepairsPackageAndActiveRuntime(t *testing.T) {
	t.Parallel()

	oldComposer := `const CONTEXT_MODE_LANE_LABELS = {
    planning: "planning",
};
const CONTEXT_MODE_GUIDANCE_BY_LANE = {
    planning: "Use ` + "`gsd_resume`" + ` for planning continuity, ` + "`gsd_exec`" + ` for noisy checks, and ` + "`gsd_exec_search`" + ` before rerunning diagnostics.",
};
`
	registrySource := officialRegistryModuleFixture()
	registry := mustDecodeRegistryFixture(t)
	patchedComposer := strings.Replace(oldComposer, planningGuidanceOriginal, planningGuidancePatched, 1)
	for _, directory := range []string{t.TempDir(), t.TempDir()} {
		writeRuntimeContractFixture(t, directory, oldComposer, registrySource)
		if err := patchPromptContractRootWithHashes(directory, sha256String(oldComposer), sha256String(patchedComposer)); err != nil {
			t.Fatal(err)
		}
		if err := validatePromptContractRootWithHashes(directory, registry, sha256String(registrySource), sha256String(patchedComposer)); err != nil {
			t.Fatalf("patched contract rejected: %v", err)
		}
		raw, err := os.ReadFile(filepath.Join(directory, "unit-context-composer.js"))
		if err != nil {
			t.Fatal(err)
		}
		content := string(raw)
		if !strings.Contains(content, "Use only the phase-scoped planning tools") ||
			!strings.Contains(content, "Do not call `gsd_resume`") {
			t.Fatalf("runtime was not repaired: %s", content)
		}
		if err := patchPromptContractRootWithHashes(directory, sha256String(oldComposer), sha256String(patchedComposer)); err != nil {
			t.Fatalf("idempotent repair failed: %v", err)
		}
	}
}

func TestValidatePinnedPromptToolContractsRejectsUnqualifiedShape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	registrySource := officialRegistryModuleFixture()
	writeRuntimeContractFixture(t, root, "const unrelated = true;", registrySource)
	err := validatePromptContractRootWithHashes(root, mustDecodeRegistryFixture(t), sha256String(registrySource), sha256String("const unrelated = true;"))
	if !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want runtime contract mismatch", err)
	}
}

func TestValidatePromptContractRejectsMissingDeclaredPromptTemplate(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	registrySource := officialRegistryModuleFixture()
	writeRuntimeContractFixture(t, root, `const CONTEXT_MODE_GUIDANCE_BY_LANE = {
    planning: "Use only the phase-scoped planning tools exposed for this unit (`+"`gsd_milestone_status`"+`, `+"`gsd_plan_milestone`"+`, `+"`gsd_plan_slice`"+`, `+"`gsd_plan_task`"+`, `+"`gsd_requirement_update`"+`, and `+"`gsd_decision_save`"+`). Do not call `+"`gsd_resume`"+`.",
};`, registrySource)
	registry := mustDecodeRegistryFixture(t)
	metadata := registry.Units["plan-milestone"]
	metadata.PromptTemplates = []string{"plan-milestone"}
	registry.Units["plan-milestone"] = metadata
	composerRaw, readErr := os.ReadFile(filepath.Join(root, "unit-context-composer.js"))
	if readErr != nil {
		t.Fatal(readErr)
	}
	if err := validatePromptContractRootWithHashes(root, registry, sha256String(registrySource), sha256String(string(composerRaw))); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want missing prompt template contract mismatch", err)
	}
}

func writeRuntimeContractFixture(t *testing.T, directory, composer, registry string) {
	t.Helper()
	if err := os.MkdirAll(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "unit-context-composer.js"), []byte(composer), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "unit-registry.js"), []byte(registry), 0o600); err != nil {
		t.Fatal(err)
	}
}
