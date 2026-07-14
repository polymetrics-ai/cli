package gsd

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestParseUnitRegistryRoutesOfficialPhaseChains(t *testing.T) {
	t.Parallel()
	registry, err := ParseUnitRegistry(`export const UNIT_REGISTRY = {
    "execute-task-simple": { kind: "variant", scopeClass: "execute-task", phaseChain: ["execution_simple", "execution"] },
    "execute-task": { kind: "primary", scopeClass: "execute-task", phaseChain: ["execution"] },
    "plan-milestone": { kind: "primary", scopeClass: "standard", phaseChain: ["planning"] },
    "validate-milestone": { kind: "primary", scopeClass: "section-close", phaseChain: ["validation", "planning"] },
    "complete-milestone": { kind: "primary", scopeClass: "section-close", phaseChain: ["completion", "validation"] },
};`)
	if err != nil {
		t.Fatal(err)
	}
	for _, unitType := range []string{"execute-task", "execute-task-simple"} {
		role, err := registry.ModelRoleForUnit(unitType)
		if err != nil || role != ModelRoleImplementation {
			t.Fatalf("%s role=%q err=%v, want implementation", unitType, role, err)
		}
	}
	for _, unitType := range []string{"plan-milestone", "validate-milestone", "complete-milestone"} {
		role, err := registry.ModelRoleForUnit(unitType)
		if err != nil || role != ModelRoleCoordinator {
			t.Fatalf("%s role=%q err=%v, want coordinator", unitType, role, err)
		}
	}
}

func TestParseUnitRegistryRejectsMissingMalformedPartialMetadata(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		`export const UNIT_REGISTRY = {};`,
		`export const UNIT_REGISTRY = { "execute-task": { kind: "primary", scopeClass: "execute-task" } };`,
		`export const UNIT_REGISTRY = { "execute-task": { kind: "primary", phaseChain: ["execution"] } };`,
	} {
		if _, err := ParseUnitRegistry(raw); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	}
}

func TestLoadUnitRegistryRejectsSymlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	target := filepath.Join(root, "elsewhere.js")
	if err := os.WriteFile(target, []byte("export const UNIT_REGISTRY = {};"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "unit-registry.js")); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadUnitRegistryFromRoot(root); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want runtime contract mismatch", err)
	}
}
