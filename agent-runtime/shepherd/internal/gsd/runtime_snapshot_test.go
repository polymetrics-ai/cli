package gsd

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestPrepareInstalledOfficialHostRuntime(t *testing.T) {
	loader := os.Getenv("GSD_OFFICIAL_LOADER")
	if loader == "" {
		t.Skip("GSD_OFFICIAL_LOADER is not configured")
	}
	gsdHome := t.TempDir()
	t.Cleanup(func() { _ = makeRuntimeTreeWritable(gsdHome) })
	command, err := PreparePinnedHostRuntime(context.Background(), []string{"node", loader}, gsdHome, "1.11.0", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	guard, err := NewPinnedHostRuntimeGuard(command)
	if err != nil {
		t.Fatal(err)
	}
	if err := guard.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePinnedPromptToolContracts(command, gsdHome, "1.11.0"); err != nil {
		t.Fatal(err)
	}
}

func TestPreparePinnedHostRuntimeBuildsPrivatePatchedSnapshot(t *testing.T) {
	t.Parallel()
	command, sourceRoot := writeHostRuntimeSnapshotFixture(t)
	options := snapshotQualificationForFixture(t, sourceRoot)
	gsdHome := t.TempDir()
	t.Cleanup(func() { _ = makeRuntimeTreeWritable(gsdHome) })
	prepared, err := preparePinnedHostRuntimeWithOptions(context.Background(), command, gsdHome, "1.11.0", []string{t.TempDir()}, options)
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(prepared[0]) || prepared[1] == command[1] {
		t.Fatalf("prepared command=%v", prepared)
	}
	within, err := pathInside(gsdHome, prepared[1])
	if err != nil || !within {
		t.Fatalf("snapshot loader=%q is not under controlled GSD home", prepared[1])
	}
	digest, _, _, err := runtimeTreeDigest(context.Background(), filepath.Dir(filepath.Dir(prepared[1])), options)
	if err != nil || digest != options.qualification.SnapshotTreeSHA256 {
		t.Fatalf("snapshot digest=%q err=%v", digest, err)
	}
	composerRaw, err := os.ReadFile(filepath.Join(filepath.Dir(prepared[1]), "resources", "extensions", "gsd", "unit-context-composer.js"))
	if err != nil || !strings.Contains(string(composerRaw), planningGuidancePatched) {
		t.Fatalf("prompt patch missing: %v", err)
	}
	headlessRaw, err := os.ReadFile(filepath.Join(filepath.Dir(prepared[1]), "headless.js"))
	if err != nil || !strings.Contains(string(headlessRaw), "inFlightToolCallIds.add(toolCallId)") {
		t.Fatalf("headless patch missing: %v", err)
	}
	info, err := os.Stat(prepared[1])
	if err != nil || info.Mode().Perm()&0o222 != 0 {
		t.Fatalf("snapshot loader mode=%v err=%v", info.Mode(), err)
	}
	second, err := preparePinnedHostRuntimeWithOptions(context.Background(), command, gsdHome, "1.11.0", nil, options)
	if err != nil || strings.Join(second, "\x00") != strings.Join(prepared, "\x00") {
		t.Fatalf("idempotent snapshot=%v err=%v", second, err)
	}
	workDir := t.TempDir()
	settingsPath := filepath.Join(gsdHome, "agent", "settings.json")
	if err := os.WriteFile(settingsPath, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	preferences := "---\nmodels:\n  research: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n  planning: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n  discuss: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }\n  execution_simple: { provider: openai-codex, model: gpt-5.5, thinking: high }\n  completion: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n  validation: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n  subagent: { provider: openai-codex, model: gpt-5.5, thinking: high }\n  uat: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }\n---\n"
	if err := os.WriteFile(filepath.Join(gsdHome, "PREFERENCES.md"), []byte(preferences), 0o600); err != nil {
		t.Fatal(err)
	}
	guard := &HostRuntimeGuard{nodePath: prepared[0], root: filepath.Dir(filepath.Dir(prepared[1])), options: options,
		gsdHome: gsdHome, promptRegistry: mustDecodeRegistryFixture(t), modelWorkDir: workDir,
		coordinatorModel: "openai-codex/gpt-5.6-sol", implementationModel: "openai-codex/gpt-5.5", expectedThinking: "high"}
	if err := guard.Validate(context.Background()); err != nil {
		t.Fatalf("valid per-launch settings rejected: %v", err)
	}
	if err := os.WriteFile(settingsPath, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"off"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := guard.Validate(context.Background()); err == nil {
		t.Fatal("per-launch runtime settings drift was accepted")
	}
	if err := os.WriteFile(settingsPath, []byte(`{"defaultProvider":"openai-codex","defaultModel":"gpt-5.6-sol","defaultThinkingLevel":"high"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(prepared[1], 0o644); err != nil {
		t.Fatal(err)
	}
	if err := guard.Validate(context.Background()); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("writable descendant error=%v, want runtime contract mismatch", err)
	}
	if err := os.Chmod(prepared[1], 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(guard.root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := guard.Validate(context.Background()); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("writable root error=%v, want runtime contract mismatch", err)
	}
}

func TestPreparePinnedHostRuntimeRejectsWorkerControlledSourceAndNodeDrift(t *testing.T) {
	t.Parallel()
	command, sourceRoot := writeHostRuntimeSnapshotFixture(t)
	options := snapshotQualificationForFixture(t, sourceRoot)
	if _, err := preparePinnedHostRuntimeWithOptions(context.Background(), command, t.TempDir(), "1.11.0", []string{filepath.Dir(sourceRoot)}, options); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("worker-controlled source error=%v", err)
	}
	gsdHome := t.TempDir()
	redirect := t.TempDir()
	if err := os.Symlink(redirect, filepath.Join(gsdHome, "agent")); err != nil {
		t.Fatal(err)
	}
	if _, err := preparePinnedHostRuntimeWithOptions(context.Background(), command, gsdHome, "1.11.0", nil, options); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("ancestor symlink error=%v", err)
	}
	options.qualification.NodeSHA256 = strings.Repeat("0", 64)
	if _, err := preparePinnedHostRuntimeWithOptions(context.Background(), command, t.TempDir(), "1.11.0", nil, options); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("node drift error=%v", err)
	}
}

func TestBoundedRuntimeSourceRejectsSymlink(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.WriteFile(target, []byte("trusted"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if _, _, err := boundedRuntimeFileSHA256(link, 1024); !errors.Is(err, ErrRuntimeContractMismatch) {
		t.Fatalf("symlink hash error=%v", err)
	}
}

func writeHostRuntimeSnapshotFixture(t *testing.T) ([]string, string) {
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
	headless := strings.Join([]string{
		headlessInteractiveSet,
		headlessIdleOriginal,
		headlessStartOriginal,
		headlessEndOriginal,
	}, "\n") + "\n"
	composer := "const CONTEXT_MODE_GUIDANCE_BY_LANE = {\n    planning: " + strconv.Quote(planningGuidanceOriginal) + ",\n};\n"
	files := map[string]string{
		filepath.Join(root, "package.json"):                `{"name":"@opengsd/gsd-pi","version":"1.11.0","type":"module"}`,
		filepath.Join(dist, "loader.js"):                   "export {};\n",
		filepath.Join(dist, "headless.js"):                 headless,
		filepath.Join(gsdRoot, "unit-context-composer.js"): composer,
		filepath.Join(gsdRoot, "unit-registry.js"):         officialRegistryModuleFixture(),
		filepath.Join(sharedRoot, "browser-contract.js"):   "export const BROWSER_CONTRACT_TOOL_NAMES = [];\n",
	}
	for path, raw := range files {
		if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	node, err = filepath.Abs(node)
	if err != nil {
		t.Fatal(err)
	}
	return []string{node, filepath.Join(dist, "loader.js")}, root
}

func snapshotQualificationForFixture(t *testing.T, sourceRoot string) hostRuntimeSnapshotOptions {
	t.Helper()
	composerPath := filepath.Join(sourceRoot, "dist", "resources", "extensions", "gsd", "unit-context-composer.js")
	composerRaw, err := os.ReadFile(composerPath)
	if err != nil {
		t.Fatal(err)
	}
	patchedComposer := strings.Replace(string(composerRaw), planningGuidanceOriginal, planningGuidancePatched, 1)
	options := hostRuntimeSnapshotOptions{originalComposerHash: sha256String(string(composerRaw)), patchedComposerHash: sha256String(patchedComposer), maxEntries: 1000, maxTreeBytes: 16 * 1024 * 1024, maxFileBytes: 2 * 1024 * 1024}
	sourceDigest, _, _, err := runtimeTreeDigest(context.Background(), sourceRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	patchedRoot := t.TempDir()
	if err := os.Chmod(patchedRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = makeRuntimeTreeWritable(patchedRoot) })
	if err := copyRuntimeTree(context.Background(), sourceRoot, patchedRoot, options); err != nil {
		t.Fatal(err)
	}
	command := []string{"node", filepath.Join(patchedRoot, "dist", "loader.js")}
	if err := ApplyPinnedHeadlessToolPatch(command, "1.11.0"); err != nil {
		t.Fatal(err)
	}
	if err := patchPromptContractRootWithHashes(filepath.Join(patchedRoot, "dist", "resources", "extensions", "gsd"), options.originalComposerHash, options.patchedComposerHash); err != nil {
		t.Fatal(err)
	}
	patchedDigest, _, _, err := runtimeTreeDigest(context.Background(), patchedRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	if err := makeRuntimeTreeReadOnly(patchedRoot); err != nil {
		t.Fatal(err)
	}
	snapshotDigest, _, _, err := runtimeTreeDigest(context.Background(), patchedRoot, options)
	if err != nil {
		t.Fatal(err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	node, err = filepath.Abs(node)
	if err != nil {
		t.Fatal(err)
	}
	nodeDigest, _, err := boundedRuntimeFileSHA256(node, 256*1024*1024)
	if err != nil {
		t.Fatal(err)
	}
	options.qualification = hostRuntimeQualification{NodeSHA256: nodeDigest, SourceTreeSHA256: sourceDigest, PatchedTreeSHA256: patchedDigest, CopiedPatchedTreeSHA256: patchedDigest, SnapshotTreeSHA256: snapshotDigest}
	return options
}
