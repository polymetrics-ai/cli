package gsd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ModelRole string

const (
	ModelRoleCoordinator    ModelRole = "coordinator"
	ModelRoleImplementation ModelRole = "implementation"
)

type UnitMetadata struct {
	UnitType              string
	Kind                  string
	ScopeClass            string
	PhaseChain            []string
	AllowedGSDTools       []string
	RequiredWorkflowTools []string
}

type UnitRegistry struct {
	Units map[string]UnitMetadata
}

func BuiltinUnitRegistry() UnitRegistry {
	units := map[string]UnitMetadata{
		"research-milestone":   {UnitType: "research-milestone", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"research"}},
		"plan-milestone":       {UnitType: "plan-milestone", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"discuss-milestone":    {UnitType: "discuss-milestone", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"discuss", "planning"}},
		"discuss-slice":        {UnitType: "discuss-slice", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"discuss", "planning"}},
		"validate-milestone":   {UnitType: "validate-milestone", Kind: "primary", ScopeClass: "section-close", PhaseChain: []string{"validation", "planning"}},
		"complete-milestone":   {UnitType: "complete-milestone", Kind: "primary", ScopeClass: "section-close", PhaseChain: []string{"completion", "validation"}},
		"research-slice":       {UnitType: "research-slice", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"research", "planning"}},
		"plan-slice":           {UnitType: "plan-slice", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"refine-slice":         {UnitType: "refine-slice", Kind: "variant", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"replan-slice":         {UnitType: "replan-slice", Kind: "variant", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"complete-slice":       {UnitType: "complete-slice", Kind: "primary", ScopeClass: "section-close", PhaseChain: []string{"completion", "validation"}},
		"reassess-roadmap":     {UnitType: "reassess-roadmap", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"execute-task":         {UnitType: "execute-task", Kind: "primary", ScopeClass: "execute-task", PhaseChain: []string{"execution"}},
		"execute-task-simple":  {UnitType: "execute-task-simple", Kind: "variant", ScopeClass: "execute-task", PhaseChain: []string{"execution_simple", "execution"}},
		"reactive-execute":     {UnitType: "reactive-execute", Kind: "variant", ScopeClass: "execute-task", PhaseChain: []string{"execution"}},
		"run-uat":              {UnitType: "run-uat", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"uat", "completion"}},
		"gate-evaluate":        {UnitType: "gate-evaluate", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"validation", "planning"}},
		"rewrite-docs":         {UnitType: "rewrite-docs", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"completion", "planning"}},
		"triage-captures":      {UnitType: "triage-captures", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"quick-task":           {UnitType: "quick-task", Kind: "primary", ScopeClass: "execute-task", PhaseChain: []string{"execution_simple", "execution"}},
		"workflow-preferences": {UnitType: "workflow-preferences", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"planning"}},
		"discuss-project":      {UnitType: "discuss-project", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"discuss", "planning"}},
		"discuss-requirements": {UnitType: "discuss-requirements", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"discuss", "planning"}},
		"research-decision":    {UnitType: "research-decision", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"research", "planning"}},
		"research-project":     {UnitType: "research-project", Kind: "primary", ScopeClass: "standard", PhaseChain: []string{"research", "planning"}},
	}
	return UnitRegistry{Units: units}
}

func LoadPinnedUnitRegistry(command []string, gsdHome, expectedVersion string) (UnitRegistry, error) {
	if expectedVersion != "1.11.0" {
		return UnitRegistry{}, fmt.Errorf("%w: unit registry compatibility is qualified only for GSD 1.11.0", ErrRuntimeContractMismatch)
	}
	if err := ValidatePinnedCommand(command, expectedVersion); err != nil {
		return UnitRegistry{}, err
	}
	roots, err := promptContractRoots(command, gsdHome)
	if err != nil {
		return UnitRegistry{}, err
	}
	var registry UnitRegistry
	for _, root := range roots {
		loaded, err := LoadUnitRegistryFromRoot(root)
		if err != nil {
			return UnitRegistry{}, err
		}
		if len(registry.Units) == 0 {
			registry = loaded
			continue
		}
		if err := registry.compatibleWith(loaded); err != nil {
			return UnitRegistry{}, err
		}
	}
	return registry, nil
}

func LoadUnitRegistryFromRoot(root string) (UnitRegistry, error) {
	if strings.TrimSpace(root) == "" || !filepath.IsAbs(root) {
		return UnitRegistry{}, fmt.Errorf("%w: absolute unit registry root is required", ErrRuntimeContractMismatch)
	}
	canonicalRoot, err := filepath.EvalSymlinks(filepath.Clean(root))
	if err != nil {
		return UnitRegistry{}, fmt.Errorf("%w: resolve unit registry root: %v", ErrRuntimeContractMismatch, err)
	}
	path := filepath.Join(canonicalRoot, "unit-registry.js")
	info, err := os.Lstat(path)
	if err != nil {
		return UnitRegistry{}, fmt.Errorf("%w: inspect unit registry: %v", ErrRuntimeContractMismatch, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return UnitRegistry{}, fmt.Errorf("%w: unit registry must be a regular file", ErrRuntimeContractMismatch)
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return UnitRegistry{}, fmt.Errorf("%w: resolve unit registry: %v", ErrRuntimeContractMismatch, err)
	}
	if !strings.HasPrefix(resolved, canonicalRoot+string(os.PathSeparator)) {
		return UnitRegistry{}, fmt.Errorf("%w: unit registry escapes its runtime root", ErrRuntimeContractMismatch)
	}
	_, raw, err := readRegularRuntimeFile(path)
	if err != nil {
		return UnitRegistry{}, err
	}
	return ParseUnitRegistry(string(raw))
}

func ParseUnitRegistry(content string) (UnitRegistry, error) {
	if !strings.Contains(content, "UNIT_REGISTRY") {
		return UnitRegistry{}, fmt.Errorf("%w: UNIT_REGISTRY export is missing", ErrRuntimeContractMismatch)
	}
	entryPattern := regexp.MustCompile(`"([a-z0-9-]+)"\s*:\s*\{`)
	matches := entryPattern.FindAllStringSubmatchIndex(content, -1)
	units := make(map[string]UnitMetadata, len(matches))
	for _, match := range matches {
		unitType := content[match[2]:match[3]]
		block, err := extractJSObjectBlock(content[match[0]:], `"`+unitType+`"`)
		if err != nil {
			return UnitRegistry{}, err
		}
		metadata, err := parseUnitMetadata(unitType, block)
		if err != nil {
			return UnitRegistry{}, err
		}
		units[unitType] = metadata
	}
	if len(units) == 0 {
		return UnitRegistry{}, fmt.Errorf("%w: unit registry is empty", ErrRuntimeContractMismatch)
	}
	registry := UnitRegistry{Units: units}
	for _, required := range []string{"plan-milestone", "execute-task", "validate-milestone", "complete-milestone"} {
		if _, ok := registry.Units[required]; !ok {
			return UnitRegistry{}, fmt.Errorf("%w: required unit %s is missing", ErrRuntimeContractMismatch, required)
		}
	}
	return registry, nil
}

func parseUnitMetadata(unitType, block string) (UnitMetadata, error) {
	phaseChain := quotedArray(block, "phaseChain")
	if len(phaseChain) == 0 {
		if fallback, ok := BuiltinUnitRegistry().Lookup(unitType); ok && len(fallback.PhaseChain) > 0 {
			phaseChain = fallback.PhaseChain
		} else {
			return UnitMetadata{}, fmt.Errorf("%w: %s phaseChain is missing", ErrRuntimeContractMismatch, unitType)
		}
	}
	for _, phase := range phaseChain {
		if !safeUnitSlug(phase) {
			return UnitMetadata{}, fmt.Errorf("%w: %s phaseChain has an unsafe phase", ErrRuntimeContractMismatch, unitType)
		}
	}
	metadata := UnitMetadata{
		UnitType:              unitType,
		Kind:                  quotedProperty(block, "kind"),
		ScopeClass:            quotedProperty(block, "scopeClass"),
		PhaseChain:            phaseChain,
		AllowedGSDTools:       quotedArray(block, "allowedGsdTools"),
		RequiredWorkflowTools: quotedArray(block, "requiredWorkflowTools"),
	}
	if !safeUnitSlug(metadata.UnitType) || strings.TrimSpace(metadata.Kind) == "" || strings.TrimSpace(metadata.ScopeClass) == "" {
		return UnitMetadata{}, fmt.Errorf("%w: %s metadata is incomplete", ErrRuntimeContractMismatch, unitType)
	}
	return metadata, nil
}

func quotedProperty(block, name string) string {
	pattern := regexp.MustCompile(name + `\s*:\s*"([^"]+)"`)
	match := pattern.FindStringSubmatch(block)
	if len(match) != 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func quotedArray(block, name string) []string {
	pattern := regexp.MustCompile(name + `\s*:\s*\[([^\]]*)\]`)
	match := pattern.FindStringSubmatch(block)
	if len(match) != 2 {
		return nil
	}
	quoted := regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(match[1], -1)
	values := make([]string, 0, len(quoted))
	for _, item := range quoted {
		value := strings.TrimSpace(item[1])
		if value != "" {
			values = append(values, value)
		}
	}
	return values
}

func (r UnitRegistry) Lookup(unitType string) (UnitMetadata, bool) {
	metadata, ok := r.Units[unitType]
	return metadata, ok
}

func (r UnitRegistry) CommandForUnit(unitType string) (string, error) {
	if _, ok := r.Lookup(unitType); !ok {
		return "", fmt.Errorf("unsupported canonical unit type %q", unitType)
	}
	if unitType == "discuss-milestone" {
		return "discuss", nil
	}
	return unitType, nil
}

func (r UnitRegistry) ModelRoleForUnit(unitType string) (ModelRole, error) {
	metadata, ok := r.Lookup(unitType)
	if !ok {
		return "", fmt.Errorf("unsupported canonical unit type %q", unitType)
	}
	for _, phase := range metadata.PhaseChain {
		switch phase {
		case "execution", "execution_simple", "subagent":
			return ModelRoleImplementation, nil
		case "research", "planning", "discuss", "completion", "validation", "uat":
			return ModelRoleCoordinator, nil
		}
	}
	return "", fmt.Errorf("%w: %s has no governed model phase", ErrRuntimeContractMismatch, unitType)
}

func (r UnitRegistry) CanonicalCommands() map[string]struct{} {
	commands := make(map[string]struct{}, len(r.Units))
	for unitType := range r.Units {
		command, err := r.CommandForUnit(unitType)
		if err == nil && command != "discuss" {
			commands[command] = struct{}{}
		}
	}
	return commands
}

func (r UnitRegistry) compatibleWith(other UnitRegistry) error {
	if len(r.Units) != len(other.Units) {
		return fmt.Errorf("%w: active unit registry count differs from package", ErrRuntimeContractMismatch)
	}
	for unitType, metadata := range r.Units {
		otherMetadata, ok := other.Units[unitType]
		if !ok {
			return fmt.Errorf("%w: active unit registry is missing %s", ErrRuntimeContractMismatch, unitType)
		}
		if strings.Join(metadata.PhaseChain, ",") != strings.Join(otherMetadata.PhaseChain, ",") || metadata.Kind != otherMetadata.Kind || metadata.ScopeClass != otherMetadata.ScopeClass {
			return fmt.Errorf("%w: active unit registry metadata differs for %s", ErrRuntimeContractMismatch, unitType)
		}
	}
	return nil
}

func safeUnitSlug(value string) bool {
	if value == "" {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func (r UnitRegistry) UnitTypes() []string {
	unitTypes := make([]string, 0, len(r.Units))
	for unitType := range r.Units {
		unitTypes = append(unitTypes, unitType)
	}
	sort.Strings(unitTypes)
	return unitTypes
}
