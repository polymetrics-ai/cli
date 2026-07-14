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

func TestApplyPinnedPromptToolPatchRepairsPackageAndActiveRuntime(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dist := filepath.Join(root, "dist")
	resources := filepath.Join(dist, "resources", "extensions", "gsd")
	if err := os.MkdirAll(resources, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"@opengsd/gsd-pi","version":"1.11.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loader := filepath.Join(dist, "loader.js")
	if err := os.WriteFile(loader, []byte("loader"), 0o600); err != nil {
		t.Fatal(err)
	}

	oldComposer := `const CONTEXT_MODE_GUIDANCE_BY_LANE = {
    planning: "Use ` + "`gsd_resume`" + ` to restore prior planning context, ` + "`gsd_exec`" + ` for noisy discovery, and ` + "`gsd_exec_search`" + ` before repeating scans.",
};
`
	registry := `export const UNIT_REGISTRY = {
    "plan-milestone": {
        kind: "primary",
        scopeClass: "standard",
        phaseChain: ["planning"],
        toolContract: {
            allowedGsdTools: [
                "gsd_milestone_status",
                "gsd_plan_milestone",
                "gsd_plan_slice",
                "gsd_plan_task",
                "gsd_decision_save",
                "gsd_requirement_update",
            ],
        },
    },
    "execute-task": { kind: "primary", scopeClass: "execute-task", phaseChain: ["execution"] },
    "validate-milestone": { kind: "primary", scopeClass: "section-close", phaseChain: ["validation", "planning"] },
    "complete-milestone": { kind: "primary", scopeClass: "section-close", phaseChain: ["completion", "validation"] },
};
`
	writeRuntimeContractFixture(t, resources, oldComposer, registry)

	homeResources := filepath.Join(root, "home", "agent", "extensions", "gsd")
	if err := os.MkdirAll(homeResources, 0o700); err != nil {
		t.Fatal(err)
	}
	writeRuntimeContractFixture(t, homeResources, oldComposer, registry)

	command := []string{"node", loader}
	gsdHome := filepath.Join(root, "home")
	if err := ApplyPinnedPromptToolPatch(command, gsdHome, "1.11.0"); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePinnedPromptToolContracts(command, gsdHome, "1.11.0"); err != nil {
		t.Fatalf("patched contract rejected: %v", err)
	}
	for _, directory := range []string{resources, homeResources} {
		raw, err := os.ReadFile(filepath.Join(directory, "unit-context-composer.js"))
		if err != nil {
			t.Fatal(err)
		}
		content := string(raw)
		if !strings.Contains(content, "Use only the phase-scoped planning tools") ||
			!strings.Contains(content, "Do not call `gsd_resume`") {
			t.Fatalf("runtime was not repaired: %s", content)
		}
	}

	if err := ApplyPinnedPromptToolPatch(command, gsdHome, "1.11.0"); err != nil {
		t.Fatalf("idempotent repair failed: %v", err)
	}
}

func TestValidatePinnedPromptToolContractsRejectsUnqualifiedShape(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dist := filepath.Join(root, "dist")
	resources := filepath.Join(dist, "resources", "extensions", "gsd")
	if err := os.MkdirAll(resources, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "package.json"), []byte(`{"name":"@opengsd/gsd-pi","version":"1.11.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loader := filepath.Join(dist, "loader.js")
	if err := os.WriteFile(loader, []byte("loader"), 0o600); err != nil {
		t.Fatal(err)
	}
	writeRuntimeContractFixture(t, resources, "const unrelated = true;", "export const UNIT_REGISTRY = {};")

	err := ValidatePinnedPromptToolContracts([]string{"node", loader}, filepath.Join(root, "home"), "1.11.0")
	if !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want runtime contract mismatch", err)
	}
}

func writeRuntimeContractFixture(t *testing.T, directory, composer, registry string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, "unit-context-composer.js"), []byte(composer), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "unit-registry.js"), []byte(registry), 0o600); err != nil {
		t.Fatal(err)
	}
}
