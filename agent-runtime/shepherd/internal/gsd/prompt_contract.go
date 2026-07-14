package gsd

import (
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
	planningGuidanceOriginal = "Use `gsd_resume` to restore prior planning context, `gsd_exec` for noisy discovery, and `gsd_exec_search` before repeating scans."
	planningGuidancePatched  = "Use only the phase-scoped planning tools exposed for this unit (`gsd_milestone_status`, `gsd_plan_milestone`, `gsd_plan_slice`, `gsd_plan_task`, `gsd_requirement_update`, and `gsd_decision_save`). Do not call `gsd_resume`, `gsd_exec`, or `gsd_exec_search` from a planning unit."
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
	if err := ValidatePinnedCommand(command, expectedVersion); err != nil {
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
	return ValidatePinnedPromptToolContracts(command, gsdHome, expectedVersion)
}

// ValidatePinnedPromptToolContracts validates every active copy Shepherd can
// observe. The controlled cache is authoritative when present, while the
// packaged copy is also checked so a cache refresh cannot reintroduce drift.
func ValidatePinnedPromptToolContracts(command []string, gsdHome, expectedVersion string) error {
	if err := ValidatePinnedCommand(command, expectedVersion); err != nil {
		return err
	}
	roots, err := promptContractRoots(command, gsdHome)
	if err != nil {
		return err
	}
	for _, root := range roots {
		if err := validatePromptContractRoot(root); err != nil {
			return err
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
	activeRoot := filepath.Join(gsdHome, "agent", "extensions", "gsd")
	composer := filepath.Join(activeRoot, "unit-context-composer.js")
	registry := filepath.Join(activeRoot, "unit-registry.js")
	composerInfo, composerErr := os.Lstat(composer)
	registryInfo, registryErr := os.Lstat(registry)
	switch {
	case errors.Is(composerErr, os.ErrNotExist) && errors.Is(registryErr, os.ErrNotExist):
		return roots, nil
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
	path := filepath.Join(root, "unit-context-composer.js")
	info, raw, err := readRegularRuntimeFile(path)
	if err != nil {
		return err
	}
	content := string(raw)
	switch {
	case strings.Count(content, planningGuidancePatched) == 1:
		return nil
	case strings.Count(content, planningGuidanceOriginal) == 1:
		content = strings.Replace(content, planningGuidanceOriginal, planningGuidancePatched, 1)
	default:
		return fmt.Errorf("%w: planning guidance has an unqualified source shape", ErrRuntimeContractMismatch)
	}
	return atomicReplaceRuntimeFile(path, []byte(content), info.Mode().Perm())
}

func validatePromptContractRoot(root string) error {
	_, composerRaw, err := readRegularRuntimeFile(filepath.Join(root, "unit-context-composer.js"))
	if err != nil {
		return err
	}
	_, registryRaw, err := readRegularRuntimeFile(filepath.Join(root, "unit-registry.js"))
	if err != nil {
		return err
	}
	guidance, err := extractPlanningGuidance(string(composerRaw))
	if err != nil {
		return err
	}
	registry, err := ParseUnitRegistry(string(registryRaw))
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

func extractPlanningGuidance(content string) (string, error) {
	marker := `planning: "`
	start := strings.Index(content, marker)
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

func extractPlanMilestoneAllowedTools(content string) ([]string, error) {
	block, err := extractJSObjectBlock(content, `"plan-milestone"`)
	if err != nil {
		return nil, err
	}
	marker := "allowedGsdTools: ["
	start := strings.Index(block, marker)
	if start < 0 {
		return nil, fmt.Errorf("%w: plan-milestone allowed tools are missing", ErrRuntimeContractMismatch)
	}
	start += len(marker)
	end := strings.Index(block[start:], "]")
	if end < 0 {
		return nil, fmt.Errorf("%w: plan-milestone allowed tools are malformed", ErrRuntimeContractMismatch)
	}
	quoted := regexp.MustCompile(`"(gsd_[a-z0-9_]+)"`).FindAllStringSubmatch(block[start:start+end], -1)
	tools := make([]string, 0, len(quoted))
	for _, match := range quoted {
		tools = append(tools, match[1])
	}
	if len(tools) == 0 {
		return nil, fmt.Errorf("%w: plan-milestone allowed tools are empty", ErrRuntimeContractMismatch)
	}
	return tools, nil
}

func extractJSObjectBlock(content, quotedKey string) (string, error) {
	start := strings.Index(content, quotedKey)
	if start < 0 {
		return "", fmt.Errorf("%w: %s registry entry is missing", ErrRuntimeContractMismatch, quotedKey)
	}
	open := strings.Index(content[start:], "{")
	if open < 0 {
		return "", fmt.Errorf("%w: %s registry entry is malformed", ErrRuntimeContractMismatch, quotedKey)
	}
	open += start
	depth := 0
	for index := open; index < len(content); index++ {
		switch content[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return content[open : index+1], nil
			}
		}
	}
	return "", fmt.Errorf("%w: %s registry entry is unterminated", ErrRuntimeContractMismatch, quotedKey)
}

func readRegularRuntimeFile(path string) (os.FileInfo, []byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: inspect %s: %v", ErrRuntimeContractMismatch, filepath.Base(path), err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, nil, fmt.Errorf("%w: %s must be a regular file", ErrRuntimeContractMismatch, filepath.Base(path))
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: read %s: %v", ErrRuntimeContractMismatch, filepath.Base(path), err)
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
