package gsd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestLoadInstalledOfficialUnitRegistry(t *testing.T) {
	loader := os.Getenv("GSD_OFFICIAL_LOADER")
	if loader == "" {
		t.Skip("GSD_OFFICIAL_LOADER is not configured")
	}
	registry, err := LoadPinnedUnitRegistry(context.Background(), []string{"node", loader}, t.TempDir(), "1.11.0")
	if err != nil {
		t.Fatal(err)
	}
	metadata, ok := registry.Lookup("run-uat")
	if !ok || len(metadata.RequiredWorkflowTools) != 5 || len(metadata.AllowedGSDTools) != 6 {
		t.Fatalf("installed run-uat metadata=%+v ok=%v", metadata, ok)
	}
}

func TestLoadPinnedUnitRegistryNormalizesSpreadArraysAndCompleteContracts(t *testing.T) {
	t.Parallel()
	command, gsdHome, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
	registry, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.11.0", options)
	if err != nil {
		t.Fatal(err)
	}
	metadata, ok := registry.Lookup("run-uat")
	if !ok {
		t.Fatal("run-uat metadata is missing")
	}
	wantAllowed := []string{"gsd_uat_exec", "gsd_uat_result_save", "gsd_resume", "gsd_milestone_status", "gsd_journal_query", "subagent"}
	if strings.Join(metadata.AllowedGSDTools, ",") != strings.Join(wantAllowed, ",") {
		t.Fatalf("run-uat allowed tools=%v want %v", metadata.AllowedGSDTools, wantAllowed)
	}
	if strings.Join(metadata.RequiredWorkflowTools, ",") != strings.Join(wantAllowed[:5], ",") {
		t.Fatalf("run-uat required tools=%v", metadata.RequiredWorkflowTools)
	}
	if got := metadata.ForbiddenGSDTools["gsd_exec"]; got != "Use gsd_uat_exec so acceptance evidence is typed as UAT-owned." {
		t.Fatalf("run-uat forbidden reason=%q", got)
	}
	if metadata.Kind != "primary" || metadata.ScopeClass != "standard" || strings.Join(metadata.PhaseChain, ",") != "uat,completion" {
		t.Fatalf("run-uat metadata=%+v", metadata)
	}
}

func TestVerifiedRegistryDataURLsAreImmutableAfterSourceReplacement(t *testing.T) {
	t.Parallel()
	command, _, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
	runtimePaths, err := resolvePinnedRegistryRuntime(command, "1.11.0", options)
	if err != nil {
		t.Fatal(err)
	}
	moduleURL, dependencyURL, err := verifiedRegistryDataURLs(runtimePaths.registryPath, runtimePaths.dependencyPath, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(runtimePaths.registryPath, []byte(`throw new Error("mutable source executed")`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(runtimePaths.nodePath, "--input-type=module", "--eval", normalizedRegistryExporter,
		moduleURL, "1.11.0", options.expectedRegistrySHA256, dependencyURL, options.expectedDependencySHA256,
		"verified-bytes", options.expectedLoaderSHA256, options.expectedHeadlessSHA256, options.expectedComposerSHA256)
	cmd.Env = registryExporterEnvironment()
	raw, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := decodeNormalizedUnitRegistryWithSource(raw, "1.11.0", options.expectedRegistrySHA256, options.expectedDependencySHA256, options.expectedLoaderSHA256, options.expectedHeadlessSHA256, options.expectedComposerSHA256); err != nil {
		t.Fatal(err)
	}
}

func TestObservedWorkflowToolsRequireTrustedNamespaceAndOfficialContract(t *testing.T) {
	t.Parallel()
	registry := mustDecodeRegistryFixture(t)
	for _, tool := range []string{"mcp__gsd-workflow__gsd_plan_task", "mcp__context7__resolve-library-id"} {
		if err := registry.ValidateObservedTool("plan-milestone", tool); err != nil {
			t.Fatalf("allowed tool %q rejected: %v", tool, err)
		}
	}
	for _, tool := range []string{"mcp__untrusted__gsd_plan_task", "gsd_resume"} {
		if err := registry.ValidateObservedTool("plan-milestone", tool); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("tool %q error=%v, want runtime contract mismatch", tool, err)
		}
	}
	if err := registry.ValidateObservedTool("new-milestone", "gsd_exec"); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("uncontracted workflow tool error=%v, want runtime contract mismatch", err)
	}
	if err := registry.ValidateObservedTool("new-milestone", "mcp__context7__resolve-library-id"); err != nil {
		t.Fatalf("non-workflow tool under uncontracted command rejected: %v", err)
	}
}

func TestCanonicalInvocationPreservesDiscussAndNextUnitIdentity(t *testing.T) {
	t.Parallel()
	registry := mustDecodeRegistryFixture(t)
	for command, observed := range map[string]string{"discuss": "discuss-milestone", "next": "plan-milestone", "execute-task": "execute-task"} {
		unitType, canonical, err := registry.CanonicalUnitForInvocation(command, observed)
		if err != nil || !canonical || unitType != observed {
			t.Fatalf("%s/%s resolved to %s canonical=%t err=%v", command, observed, unitType, canonical, err)
		}
	}
	if !registry.IsCanonicalCommand("discuss") {
		t.Fatal("discuss alias did not retain canonical identity")
	}
}

func TestNormalizedRegistryRoutesOnlyFromOfficialPhaseMetadata(t *testing.T) {
	t.Parallel()
	registry := mustDecodeRegistryFixture(t)
	for _, unitType := range []string{"execute-task", "execute-task-simple", "reactive-execute"} {
		role, err := registry.ModelRoleForUnit(unitType)
		if err != nil || role != ModelRoleImplementation {
			t.Fatalf("%s role=%q err=%v, want implementation", unitType, role, err)
		}
	}
	for _, unitType := range []string{"research-slice", "plan-slice", "complete-slice", "validate-milestone", "run-uat", "complete-milestone"} {
		role, err := registry.ModelRoleForUnit(unitType)
		if err != nil || role != ModelRoleCoordinator {
			t.Fatalf("%s role=%q err=%v, want coordinator", unitType, role, err)
		}
	}
	if _, err := registry.ModelRoleForUnit("quick-task"); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("null-metadata sidecar role error=%v, want runtime contract mismatch", err)
	}
}

func TestSidecarPolicyIsSeparateVersionedAndNarrow(t *testing.T) {
	t.Parallel()
	registry := mustDecodeRegistryFixture(t)
	for _, unitType := range []string{"quick-task", "triage-captures"} {
		if _, ok := registry.Lookup(unitType); ok {
			t.Fatalf("%s was represented as official routable metadata", unitType)
		}
		entry, ok := PinnedSidecarPolicy().Lookup(unitType)
		if !ok || entry.Supported || entry.Reason == "" {
			t.Fatalf("%s sidecar policy=%+v ok=%v", unitType, entry, ok)
		}
	}
	if PinnedSidecarPolicy().Version != "shepherd.gsd-sidecars/v1" {
		t.Fatalf("sidecar policy version=%q", PinnedSidecarPolicy().Version)
	}
}

func TestDecodeNormalizedRegistryRejectsMalformedPartialDuplicateAndUnexpectedJSON(t *testing.T) {
	t.Parallel()
	valid := normalizedRegistryFixture(t)
	validRaw := marshalNormalizedFixture(t, valid)
	tests := map[string][]byte{
		"malformed":        []byte(`{"schemaVersion":`),
		"partial":          []byte(`{"schemaVersion":"shepherd.gsd-unit-registry/v1"}`),
		"duplicate field":  []byte(strings.Replace(string(validRaw), `"schemaVersion":`, `"schemaVersion":"shepherd.gsd-unit-registry/v1","schemaVersion":`, 1)),
		"unexpected field": []byte(strings.Replace(string(validRaw), `"units":`, `"unexpected":true,"units":`, 1)),
		"trailing json":    append(append([]byte{}, validRaw...), []byte(` {}`)...),
	}
	for name, raw := range tests {
		name, raw := name, raw
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := decodeNormalizedUnitRegistry(raw); !errors.Is(err, ErrRuntimeContractMismatch) {
				t.Fatalf("error=%v, want runtime contract mismatch", err)
			}
		})
	}
}

func TestDecodeNormalizedRegistryRejectsNullUnknownMissingAndDuplicateMetadata(t *testing.T) {
	t.Parallel()
	tests := map[string]func(*normalizedRegistryDocument){
		"null official phase": func(document *normalizedRegistryDocument) {
			document.Units[0].PhaseChain = nil
		},
		"unknown phase": func(document *normalizedRegistryDocument) {
			document.Units[0].PhaseChain = []string{"future_phase"}
		},
		"unknown unit": func(document *normalizedRegistryDocument) {
			document.Units[0].UnitType = "future-unit"
		},
		"missing unit": func(document *normalizedRegistryDocument) {
			document.Units = document.Units[1:]
		},
		"duplicate unit": func(document *normalizedRegistryDocument) {
			document.Units[1].UnitType = document.Units[0].UnitType
		},
		"duplicate allowed tool": func(document *normalizedRegistryDocument) {
			document.Units[0].AllowedGSDTools = []string{"gsd_summary_save", "gsd_summary_save"}
		},
		"unexpected sidecar": func(document *normalizedRegistryDocument) {
			document.ExcludedSidecars = append(document.ExcludedSidecars, "future-sidecar")
		},
		"missing forbidden reason": func(document *normalizedRegistryDocument) {
			document.Units[0].ForbiddenGSDTools = []normalizedForbiddenTool{{Tool: "gsd_exec"}}
		},
	}
	for name, mutate := range tests {
		name, mutate := name, mutate
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			document := normalizedRegistryFixture(t)
			mutate(&document)
			if _, err := decodeNormalizedUnitRegistry(marshalNormalizedFixture(t, document)); !errors.Is(err, ErrRuntimeContractMismatch) {
				t.Fatalf("error=%v, want runtime contract mismatch", err)
			}
		})
	}
}

func TestLoadPinnedUnitRegistryRejectsWrongVersionSourceDriftSymlinksAndOversize(t *testing.T) {
	t.Parallel()
	t.Run("wrong version", func(t *testing.T) {
		t.Parallel()
		command, gsdHome, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
		if _, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.12.0", options); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	})
	t.Run("source drift", func(t *testing.T) {
		t.Parallel()
		command, gsdHome, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
		path := filepath.Join(filepath.Dir(command[1]), "resources", "extensions", "gsd", "unit-registry.js")
		if err := os.WriteFile(path, []byte(officialRegistryModuleFixture()+"\n// drift\n"), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.11.0", options); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	})
	t.Run("registry symlink", func(t *testing.T) {
		t.Parallel()
		command, gsdHome, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
		path := filepath.Join(filepath.Dir(command[1]), "resources", "extensions", "gsd", "unit-registry.js")
		target := path + ".target"
		if err := os.Rename(path, target); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, path); err != nil {
			t.Fatal(err)
		}
		if _, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.11.0", options); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	})
	t.Run("missing registry", func(t *testing.T) {
		t.Parallel()
		command, gsdHome, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
		path := filepath.Join(filepath.Dir(command[1]), "resources", "extensions", "gsd", "unit-registry.js")
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
		if _, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.11.0", options); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	})
	t.Run("unclean escaped loader path", func(t *testing.T) {
		t.Parallel()
		command, gsdHome, options := writeRegistryRuntimeFixture(t, officialRegistryModuleFixture())
		command[1] = filepath.Dir(command[1]) + string(filepath.Separator) + ".." + string(filepath.Separator) + "dist" + string(filepath.Separator) + "loader.js"
		if _, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.11.0", options); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	})
	t.Run("oversized normalized output", func(t *testing.T) {
		t.Parallel()
		document := normalizedRegistryFixture(t)
		raw := marshalNormalizedFixture(t, document)
		if _, err := decodeNormalizedUnitRegistry(append(raw, make([]byte, maxNormalizedRegistryBytes)...)); !errors.Is(err, ErrRuntimeContractMismatch) {
			t.Fatalf("error=%v, want runtime contract mismatch", err)
		}
	})
}

func TestLoadPinnedUnitRegistryHonorsCancellationAndTimeout(t *testing.T) {
	t.Parallel()
	module := officialRegistryModuleFixture() + `
import { spawn } from "node:child_process";
spawn(process.execPath, ["--eval", "setInterval(() => {}, 1000)"], { stdio: ["ignore", "inherit", "inherit"] });
await new Promise(() => {});
`
	command, gsdHome, options := writeRegistryRuntimeFixture(t, module)
	options.timeout = 25 * time.Millisecond
	started := time.Now()
	_, err := loadPinnedUnitRegistryWithOptions(context.Background(), command, gsdHome, "1.11.0", options)
	if !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("error=%v, want runtime contract mismatch", err)
	}
	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("registry exporter timeout took %s", elapsed)
	}
}

func mustDecodeRegistryFixture(t *testing.T) UnitRegistry {
	t.Helper()
	registry, err := decodeNormalizedUnitRegistry(marshalNormalizedFixture(t, normalizedRegistryFixture(t)))
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func normalizedRegistryFixture(t *testing.T) normalizedRegistryDocument {
	t.Helper()
	units := make([]normalizedUnitMetadata, 0, len(expectedOfficialUnitTypes))
	for _, unitType := range expectedOfficialUnitTypes {
		phaseChain := []string{"planning"}
		scopeClass := "standard"
		kind := "primary"
		switch unitType {
		case "execute-task", "reactive-execute":
			phaseChain, scopeClass = []string{"execution"}, "execute-task"
		case "execute-task-simple":
			phaseChain, scopeClass, kind = []string{"execution_simple", "execution"}, "execute-task", "variant"
		case "discuss-slice":
			phaseChain, kind = []string{"discuss", "planning"}, "variant"
		case "discuss-milestone", "workflow-preferences", "discuss-project", "discuss-requirements", "research-decision":
			phaseChain = []string{"discuss", "planning"}
		case "research-milestone", "research-slice", "research-project":
			phaseChain = []string{"research"}
		case "validate-milestone", "reassess-roadmap", "gate-evaluate", "rewrite-docs":
			phaseChain = []string{"validation", "planning"}
		case "complete-milestone", "complete-slice":
			phaseChain = []string{"completion"}
		case "run-uat":
			phaseChain = []string{"uat", "completion"}
		}
		unit := normalizedUnitMetadata{
			UnitType: unitType, Kind: kind, ScopeClass: scopeClass, PhaseChain: phaseChain,
			PromptTemplates: []string{}, AllowedGSDTools: []string{"gsd_summary_save"}, RequiredWorkflowTools: []string{}, ForbiddenGSDTools: []normalizedForbiddenTool{},
		}
		if unitType == "plan-milestone" {
			unit.AllowedGSDTools = []string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_plan_slice", "gsd_plan_task", "gsd_decision_save", "gsd_requirement_update"}
			unit.RequiredWorkflowTools = []string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_plan_slice", "gsd_plan_task"}
		}
		if unitType == "run-uat" {
			unit.AllowedGSDTools = []string{"gsd_uat_exec", "gsd_uat_result_save", "gsd_resume", "gsd_milestone_status", "gsd_journal_query", "subagent"}
			unit.RequiredWorkflowTools = append([]string{}, unit.AllowedGSDTools[:5]...)
			unit.ForbiddenGSDTools = []normalizedForbiddenTool{{Tool: "gsd_exec", Reason: "Use gsd_uat_exec so acceptance evidence is typed as UAT-owned."}}
		}
		units = append(units, unit)
	}
	return normalizedRegistryDocument{
		SchemaVersion: normalizedRegistrySchemaVersion,
		Source:        normalizedRegistrySource{PackageName: "@opengsd/gsd-pi", PackageVersion: "1.11.0", ModulePath: officialRegistryModulePath, ModuleSHA256: officialRegistrySHA256, DependencyPath: officialRegistryDependencyPath, DependencySHA256: officialRegistryDependencySHA256, LoaderPath: "dist/loader.js", LoaderSHA256: officialLoaderSHA256, HeadlessSHA256: officialPatchedHeadlessSHA256, ComposerSHA256: officialPatchedComposerSHA256},
		Units:         units, ExcludedSidecars: []string{"quick-task", "triage-captures"},
	}
}

func marshalNormalizedFixture(t *testing.T, document normalizedRegistryDocument) []byte {
	t.Helper()
	raw, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func writeRegistryRuntimeFixture(t *testing.T, module string) ([]string, string, registryLoadOptions) {
	t.Helper()
	root := t.TempDir()
	dist := filepath.Join(root, "dist")
	gsdRoot := filepath.Join(dist, "resources", "extensions", "gsd")
	sharedRoot := filepath.Join(dist, "resources", "extensions", "shared")
	if err := os.MkdirAll(gsdRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(sharedRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	loader := filepath.Join(dist, "loader.js")
	loaderSource := "export {};\n"
	browser := "export const BROWSER_CONTRACT_TOOL_NAMES = [];\n"
	for path, raw := range map[string]string{
		filepath.Join(root, "package.json"): `{"name":"@opengsd/gsd-pi","version":"1.11.0","type":"module"}`,
		loader:                              loaderSource,
		filepath.Join(gsdRoot, "unit-registry.js"):       module,
		filepath.Join(sharedRoot, "browser-contract.js"): browser,
	} {
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	nodePath, err = filepath.Abs(nodePath)
	if err != nil {
		t.Fatal(err)
	}
	nodeDigest, _, err := boundedRuntimeFileSHA256(nodePath, 256*1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	return []string{"node", loader}, t.TempDir(), registryLoadOptions{
		expectedRegistrySHA256:   sha256String(module),
		expectedDependencySHA256: sha256String(browser),
		expectedLoaderSHA256:     sha256String(loaderSource),
		expectedHeadlessSHA256:   sha256String("fixture-headless"),
		expectedComposerSHA256:   sha256String("fixture-composer"),
		expectedNodeSHA256:       nodeDigest,
		timeout:                  5 * time.Second,
	}
}

func sha256String(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func officialRegistryModuleFixture() string {
	document := normalizedRegistryDocument{}
	_ = document
	unitTypes := append([]string{}, expectedOfficialUnitTypes...)
	sort.Strings(unitTypes)
	var builder strings.Builder
	builder.WriteString("import { BROWSER_CONTRACT_TOOL_NAMES } from \"../shared/browser-contract.js\";\n")
	builder.WriteString("export const RUN_UAT_WORKFLOW_TOOL_NAMES = [\"gsd_uat_exec\",\"gsd_uat_result_save\",\"gsd_resume\",\"gsd_milestone_status\",\"gsd_journal_query\"];\n")
	builder.WriteString("export const UNIT_REGISTRY = {\n")
	for _, unitType := range unitTypes {
		fixture := normalizedRegistryFixtureForModule(unitType)
		builder.WriteString("  ")
		key, _ := json.Marshal(unitType)
		builder.Write(key)
		builder.WriteString(": { kind: ")
		kind, _ := json.Marshal(fixture.Kind)
		builder.Write(kind)
		builder.WriteString(", scopeClass: ")
		scope, _ := json.Marshal(fixture.ScopeClass)
		builder.Write(scope)
		builder.WriteString(", phaseChain: ")
		phases, _ := json.Marshal(fixture.PhaseChain)
		builder.Write(phases)
		builder.WriteString(", toolContract: { allowedGsdTools: ")
		if unitType == "run-uat" {
			builder.WriteString(`[...RUN_UAT_WORKFLOW_TOOL_NAMES,"subagent"]`)
		} else {
			allowed, _ := json.Marshal(fixture.AllowedGSDTools)
			builder.Write(allowed)
		}
		builder.WriteString(", requiredWorkflowTools: ")
		if unitType == "run-uat" {
			builder.WriteString(`[...RUN_UAT_WORKFLOW_TOOL_NAMES]`)
		} else {
			required, _ := json.Marshal(fixture.RequiredWorkflowTools)
			builder.Write(required)
		}
		if len(fixture.ForbiddenGSDTools) > 0 {
			builder.WriteString(`, forbiddenGsdTools: { "gsd_exec": "Use gsd_uat_exec so acceptance evidence is typed as UAT-owned." }`)
		}
		builder.WriteString(" } },\n")
	}
	builder.WriteString("  \"quick-task\": { kind: \"primary\", scopeClass: \"standard\", phaseChain: null, toolContract: null },\n")
	builder.WriteString("  \"triage-captures\": { kind: \"primary\", scopeClass: \"standard\", phaseChain: null, toolContract: null },\n")
	builder.WriteString("};\n")
	return builder.String()
}

func normalizedRegistryFixtureForModule(unitType string) normalizedUnitMetadata {
	phaseChain := []string{"planning"}
	scopeClass := "standard"
	kind := "primary"
	switch unitType {
	case "execute-task", "reactive-execute":
		phaseChain, scopeClass = []string{"execution"}, "execute-task"
	case "execute-task-simple":
		phaseChain, scopeClass, kind = []string{"execution_simple", "execution"}, "execute-task", "variant"
	case "discuss-slice":
		phaseChain, kind = []string{"discuss", "planning"}, "variant"
	case "discuss-milestone", "workflow-preferences", "discuss-project", "discuss-requirements", "research-decision":
		phaseChain = []string{"discuss", "planning"}
	case "research-milestone", "research-slice", "research-project":
		phaseChain = []string{"research"}
	case "validate-milestone", "reassess-roadmap", "gate-evaluate", "rewrite-docs":
		phaseChain = []string{"validation", "planning"}
	case "complete-milestone", "complete-slice":
		phaseChain = []string{"completion"}
	case "run-uat":
		phaseChain = []string{"uat", "completion"}
	}
	unit := normalizedUnitMetadata{UnitType: unitType, Kind: kind, ScopeClass: scopeClass, PhaseChain: phaseChain, AllowedGSDTools: []string{"gsd_summary_save"}, RequiredWorkflowTools: []string{}, ForbiddenGSDTools: []normalizedForbiddenTool{}}
	if unitType == "plan-milestone" {
		unit.AllowedGSDTools = []string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_plan_slice", "gsd_plan_task", "gsd_decision_save", "gsd_requirement_update"}
		unit.RequiredWorkflowTools = []string{"gsd_milestone_status", "gsd_plan_milestone", "gsd_plan_slice", "gsd_plan_task"}
	}
	if unitType == "run-uat" {
		unit.AllowedGSDTools = []string{"gsd_uat_exec", "gsd_uat_result_save", "gsd_resume", "gsd_milestone_status", "gsd_journal_query", "subagent"}
		unit.RequiredWorkflowTools = append([]string{}, unit.AllowedGSDTools[:5]...)
		unit.ForbiddenGSDTools = []normalizedForbiddenTool{{Tool: "gsd_exec", Reason: "Use gsd_uat_exec so acceptance evidence is typed as UAT-owned."}}
	}
	return unit
}
