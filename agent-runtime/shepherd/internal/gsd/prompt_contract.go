package gsd

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var ErrRuntimeContractMismatch = errors.New("runtime_contract_mismatch")

const (
	planningGuidanceOriginal      = "Use `gsd_resume` for planning continuity, `gsd_exec` for noisy checks, and `gsd_exec_search` before rerunning diagnostics."
	planningGuidancePatched       = "Use only the phase-scoped planning tools exposed for this unit (`gsd_milestone_status`, `gsd_plan_milestone`, `gsd_plan_slice`, `gsd_plan_task`, `gsd_requirement_update`, and `gsd_decision_save`). Do not call `gsd_resume`, `gsd_exec`, or `gsd_exec_search` from a planning unit."
	officialComposerSHA256        = "e286aedcbf4a3c22cbeae66e59851023ca1f3f7be38eb38f15556d4de481f2c7"
	officialPatchedComposerSHA256 = "a43406c4b532f3a817dffa63dc6381329f098b2e8110b71f8d9b7e66a81c3b9e"
)

var gsdToolPattern = regexp.MustCompile("`(gsd_[a-z0-9_]+)`")

// ValidatePromptToolContract enforces the runtime invariant that every GSD
// lifecycle tool positively advertised to a unit is present in its hard tool
// contract. Negative guidance is removed by the caller before validation.
func ValidatePromptToolContract(unit string, allowed, advertised []string) error {
	if strings.TrimSpace(unit) == "" || len(allowed) == 0 || len(advertised) == 0 {
		return fmt.Errorf("%w: %s has an empty prompt or tool contract", ErrRuntimeContractMismatch, unit)
	}
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, tool := range allowed {
		tool = strings.TrimSpace(tool)
		if strings.HasPrefix(tool, "gsd_") {
			allowedSet[tool] = struct{}{}
		}
	}
	var forbidden []string
	for _, tool := range advertised {
		tool = strings.TrimSpace(tool)
		if !strings.HasPrefix(tool, "gsd_") {
			continue
		}
		if _, ok := allowedSet[tool]; !ok {
			forbidden = append(forbidden, tool)
		}
	}
	if len(forbidden) == 0 {
		return nil
	}
	sort.Strings(forbidden)
	return fmt.Errorf("%w: %s prompt advertises forbidden tools: %s", ErrRuntimeContractMismatch, unit, strings.Join(forbidden, ", "))
}

// ApplyPinnedPromptToolPatch repairs the qualified GSD Pi 1.11.0 planning
// guidance in both the installed package and the controlled active resource
// cache. It refuses unknown or partial source shapes instead of guessing.
func ApplyPinnedPromptToolPatch(command []string, gsdHome, expectedVersion string) error {
	if expectedVersion != "1.11.0" {
		return fmt.Errorf("%w: prompt compatibility is qualified only for GSD 1.11.0", ErrRuntimeContractMismatch)
	}
	if _, err := LoadPinnedUnitRegistry(context.Background(), command, gsdHome, expectedVersion); err != nil {
		return err
	}
	roots, err := promptContractRoots(command, gsdHome)
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := patchPromptContractRoot(root); err != nil {
			return err
		}
	}
	if len(roots) == 2 {
		if err := makeRuntimeTreeReadOnly(roots[1]); err != nil {
			return fmt.Errorf("secure active prompt runtime: %w", err)
		}
	}
	return ValidatePinnedPromptToolContracts(command, gsdHome, expectedVersion)
}

// ValidatePinnedPromptToolContracts validates every active copy Shepherd can
// observe. The controlled cache is authoritative when present, while the
// packaged copy is also checked so a cache refresh cannot reintroduce drift.
func ValidatePinnedPromptToolContracts(command []string, gsdHome, expectedVersion string) error {
	registry, err := LoadPinnedUnitRegistry(context.Background(), command, gsdHome, expectedVersion)
	if err != nil {
		return err
	}
	return validatePinnedPromptToolContractsWithRegistry(command, gsdHome, registry)
}

func validatePinnedPromptToolContractsWithRegistry(command []string, gsdHome string, registry UnitRegistry) error {
	roots, err := promptContractRoots(command, gsdHome)
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := validatePromptContractRoot(root, registry); err != nil {
			return err
		}
	}
	if len(roots) == 2 {
		options := hostRuntimeSnapshotOptions{maxEntries: 2000, maxTreeBytes: 64 * 1024 * 1024, maxFileBytes: 2 * 1024 * 1024}
		packageDigest, packageEntries, packageBytes, err := runtimeTreeDigest(context.Background(), roots[0], options)
		if err != nil {
			return err
		}
		activeDigest, activeEntries, activeBytes, err := runtimeTreeDigest(context.Background(), roots[1], options)
		if err != nil {
			return err
		}
		if packageDigest != activeDigest || packageEntries != activeEntries || packageBytes != activeBytes {
			return fmt.Errorf("%w: active prompt runtime differs from the private package snapshot", ErrRuntimeContractMismatch)
		}
	}
	return nil
}

func promptContractRoots(command []string, gsdHome string) ([]string, error) {
	loader, err := filepath.EvalSymlinks(command[1])
	if err != nil {
		return nil, fmt.Errorf("resolve GSD loader: %w", err)
	}
	packageRoot := filepath.Join(filepath.Dir(loader), "resources", "extensions", "gsd")
	roots := []string{packageRoot}
	extensionsRoot := filepath.Join(gsdHome, "agent", "extensions")
	activeRoot := filepath.Join(extensionsRoot, "gsd")
	activeInfo, activeErr := os.Lstat(activeRoot)
	if errors.Is(activeErr, os.ErrNotExist) {
		return roots, nil
	}
	if activeErr != nil || activeInfo.Mode()&os.ModeSymlink != 0 || !activeInfo.IsDir() || !runtimePathOwnedByCurrentUser(activeInfo) {
		return nil, fmt.Errorf("%w: controlled GSD prompt root must be an owned non-symlink directory", ErrRuntimeContractMismatch)
	}
	for _, directory := range []string{filepath.Join(gsdHome, "agent"), extensionsRoot} {
		info, err := os.Lstat(directory)
		if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !runtimePathOwnedByCurrentUser(info) {
			return nil, fmt.Errorf("%w: controlled GSD prompt ancestors must be owned non-symlink directories", ErrRuntimeContractMismatch)
		}
	}
	composer := filepath.Join(activeRoot, "unit-context-composer.js")
	registry := filepath.Join(activeRoot, "unit-registry.js")
	composerInfo, composerErr := os.Lstat(composer)
	registryInfo, registryErr := os.Lstat(registry)
	switch {
	case errors.Is(composerErr, os.ErrNotExist) && errors.Is(registryErr, os.ErrNotExist):
		return nil, fmt.Errorf("%w: controlled GSD prompt root exists without the required runtime", ErrRuntimeContractMismatch)
	case composerErr != nil || registryErr != nil:
		return nil, fmt.Errorf("%w: controlled GSD prompt cache is partial", ErrRuntimeContractMismatch)
	case !composerInfo.Mode().IsRegular() || composerInfo.Mode()&os.ModeSymlink != 0 ||
		!registryInfo.Mode().IsRegular() || registryInfo.Mode()&os.ModeSymlink != 0:
		return nil, fmt.Errorf("%w: controlled GSD prompt cache must contain regular files", ErrRuntimeContractMismatch)
	default:
		return append(roots, activeRoot), nil
	}
}

func patchPromptContractRoot(root string) error {
	return patchPromptContractRootWithHashes(root, officialComposerSHA256, officialPatchedComposerSHA256)
}

func patchPromptContractRootWithHashes(root, originalHash, patchedHash string) error {
	path := filepath.Join(root, "unit-context-composer.js")
	info, raw, err := readRegularRuntimeFile(path)
	if err != nil {
		return err
	}
	content := string(raw)
	observedHash := sha256.Sum256(raw)
	observed := hex.EncodeToString(observedHash[:])
	if observed != originalHash && observed != patchedHash {
		return fmt.Errorf("%w: planning composer source drifted", ErrRuntimeContractMismatch)
	}
	switch {
	case strings.Count(content, planningGuidancePatched) == 1:
		return nil
	case strings.Count(content, planningGuidanceOriginal) == 1:
		content = strings.Replace(content, planningGuidanceOriginal, planningGuidancePatched, 1)
	default:
		return fmt.Errorf("%w: planning guidance has an unqualified source shape", ErrRuntimeContractMismatch)
	}
	updated := []byte(content)
	updatedHash := sha256.Sum256(updated)
	if hex.EncodeToString(updatedHash[:]) != patchedHash {
		return fmt.Errorf("%w: patched planning composer differs from the qualified source", ErrRuntimeContractMismatch)
	}
	return atomicReplaceRuntimeFile(path, updated, info.Mode().Perm())
}

func validatePromptContractRoot(root string, registry UnitRegistry) error {
	return validatePromptContractRootWithHashes(root, registry, officialRegistrySHA256, officialPatchedComposerSHA256)
}

func validatePromptContractRootWithHashes(root string, registry UnitRegistry, expectedRegistryHash, expectedComposerHash string) error {
	_, composerRaw, err := readRegularRuntimeFile(filepath.Join(root, "unit-context-composer.js"))
	if err != nil {
		return err
	}
	composerHash := sha256.Sum256(composerRaw)
	if hex.EncodeToString(composerHash[:]) != expectedComposerHash {
		return fmt.Errorf("%w: active prompt composer differs from the qualified patched runtime", ErrRuntimeContractMismatch)
	}
	_, registryRaw, err := readRegularRuntimeFile(filepath.Join(root, "unit-registry.js"))
	if err != nil {
		return err
	}
	registryHash := sha256.Sum256(registryRaw)
	if hex.EncodeToString(registryHash[:]) != expectedRegistryHash {
		return fmt.Errorf("%w: prompt registry source differs from the qualified runtime", ErrRuntimeContractMismatch)
	}
	for _, metadata := range registry.Units {
		for _, template := range metadata.PromptTemplates {
			info, promptRaw, err := readRegularRuntimeFile(filepath.Join(root, "prompts", template+".md"))
			if err != nil {
				return err
			}
			if info.Size() > 1024*1024 {
				return fmt.Errorf("%w: prompt template %s is oversized", ErrRuntimeContractMismatch, template)
			}
			if err := validatePromptTemplateToolMentions(metadata, template, string(promptRaw)); err != nil {
				return err
			}
		}
	}
	guidance, err := extractPlanningGuidance(string(composerRaw))
	if err != nil {
		return err
	}
	planMetadata, ok := registry.Lookup("plan-milestone")
	if !ok || len(planMetadata.AllowedGSDTools) == 0 {
		return fmt.Errorf("%w: plan-milestone allowed tools are missing", ErrRuntimeContractMismatch)
	}
	allowed := planMetadata.AllowedGSDTools
	positive := guidance
	if index := strings.Index(strings.ToLower(positive), "do not call"); index >= 0 {
		positive = positive[:index]
	}
	advertisedMatches := gsdToolPattern.FindAllStringSubmatch(positive, -1)
	advertised := make([]string, 0, len(advertisedMatches))
	for _, match := range advertisedMatches {
		advertised = append(advertised, match[1])
	}
	return ValidatePromptToolContract("plan-milestone", allowed, advertised)
}

func validatePromptTemplateToolMentions(metadata UnitMetadata, template, content string) error {
	for _, line := range strings.Split(content, "\n") {
		matches := gsdToolPattern.FindAllStringSubmatch(line, -1)
		if len(matches) == 0 {
			continue
		}
		lower := strings.ToLower(line)
		negative := strings.Contains(lower, "do not") || strings.Contains(lower, "don't") || strings.Contains(lower, "must not") ||
			strings.Contains(lower, "never ") || strings.Contains(lower, " unavailable") || strings.Contains(lower, "instead of") || strings.Contains(lower, "forbidden")
		if negative {
			continue
		}
		advertised := make([]string, 0, len(matches))
		for _, match := range matches {
			advertised = append(advertised, match[1])
		}
		if err := ValidatePromptToolContract(metadata.UnitType+"/"+template, metadata.AllowedGSDTools, advertised); err != nil {
			return err
		}
	}
	return nil
}

func extractPlanningGuidance(content string) (string, error) {
	marker := `planning: "`
	start := strings.LastIndex(content, marker)
	if start < 0 {
		return "", fmt.Errorf("%w: planning guidance is missing", ErrRuntimeContractMismatch)
	}
	start += len(marker)
	end := strings.Index(content[start:], `",`)
	if end < 0 {
		return "", fmt.Errorf("%w: planning guidance is malformed", ErrRuntimeContractMismatch)
	}
	return content[start : start+end], nil
}

func readRegularRuntimeFile(path string) (os.FileInfo, []byte, error) {
	raw, info, err := readBoundedRuntimeFile(path, maxRuntimeSourceBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: read bounded non-symlink %s: %v", ErrRuntimeContractMismatch, filepath.Base(path), err)
	}
	return info, raw, nil
}

func atomicReplaceRuntimeFile(path string, raw []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".runtime-contract-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(mode); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(raw); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	dir, err := os.Open(directory)
	if err != nil {
		return err
	}
	return errors.Join(dir.Sync(), dir.Close())
}
