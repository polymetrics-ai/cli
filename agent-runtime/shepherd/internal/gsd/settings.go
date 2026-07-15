package gsd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type runtimeSettings struct {
	DefaultProvider      string `json:"defaultProvider"`
	DefaultModel         string `json:"defaultModel"`
	DefaultThinkingLevel string `json:"defaultThinkingLevel"`
}

func ValidateRuntimeSettings(gsdHome, workDir, expectedModel, expectedThinking string) error {
	path := filepath.Join(gsdHome, "agent", "settings.json")
	raw, err := readGovernedPolicyFile(path)
	if err != nil {
		return fmt.Errorf("read controlled GSD settings: %w", err)
	}
	var settings runtimeSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("decode controlled GSD settings: %w", err)
	}
	if settings.DefaultProvider == "" || settings.DefaultModel == "" {
		return errors.New("controlled GSD settings do not pin provider and model")
	}
	observed := settings.DefaultProvider + "/" + settings.DefaultModel
	if observed != expectedModel || settings.DefaultThinkingLevel != expectedThinking {
		return fmt.Errorf("controlled GSD runtime is %s/%s, expected %s/%s", observed, settings.DefaultThinkingLevel, expectedModel, expectedThinking)
	}
	projectPath := filepath.Join(workDir, ".pi", "settings.json")
	projectRaw, err := readGovernedPolicyFile(projectPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("read project Pi settings: %w", err)
	}
	if err == nil {
		var project runtimeSettings
		if err := json.Unmarshal(projectRaw, &project); err != nil {
			return fmt.Errorf("decode project Pi settings: %w", err)
		}
		if project.DefaultProvider != "" && project.DefaultProvider != settings.DefaultProvider {
			return errors.New("project settings override the governed provider")
		}
		if project.DefaultModel != "" && project.DefaultModel != settings.DefaultModel {
			return errors.New("project settings override the governed model")
		}
		if project.DefaultThinkingLevel != "" && project.DefaultThinkingLevel != expectedThinking {
			return errors.New("project settings override the governed thinking level")
		}
	}
	return nil
}

// NormalizeRuntimeSettings restores the stable coordinator identity after an
// official GSD direct execution launch persists the governed implementation
// model. Any identity outside those two configured models fails closed.
func NormalizeRuntimeSettings(gsdHome, coordinatorModel, implementationModel, expectedThinking string) error {
	path := filepath.Join(gsdHome, "agent", "settings.json")
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect controlled GSD settings: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("controlled GSD settings must be a regular file")
	}
	raw, err := readGovernedPolicyFile(path)
	if err != nil {
		return fmt.Errorf("read controlled GSD settings: %w", err)
	}
	var settings runtimeSettings
	if err := json.Unmarshal(raw, &settings); err != nil {
		return fmt.Errorf("decode controlled GSD settings: %w", err)
	}
	if settings.DefaultThinkingLevel != expectedThinking {
		return fmt.Errorf("controlled GSD thinking is %s, expected %s", settings.DefaultThinkingLevel, expectedThinking)
	}
	observed := settings.DefaultProvider + "/" + settings.DefaultModel
	if observed == coordinatorModel {
		if err := os.Chmod(path, 0o600); err != nil {
			return fmt.Errorf("secure controlled GSD settings: %w", err)
		}
		return nil
	}
	if observed != implementationModel {
		return fmt.Errorf("controlled GSD runtime is %s, expected governed coordinator %s or implementation %s", observed, coordinatorModel, implementationModel)
	}
	provider, model, ok := strings.Cut(coordinatorModel, "/")
	if !ok || provider == "" || model == "" {
		return errors.New("coordinator model must be provider-qualified")
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		return fmt.Errorf("decode controlled GSD settings document: %w", err)
	}
	document["defaultProvider"], _ = json.Marshal(provider)
	document["defaultModel"], _ = json.Marshal(model)
	document["defaultThinkingLevel"], _ = json.Marshal(expectedThinking)
	updated, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return fmt.Errorf("encode controlled GSD settings: %w", err)
	}
	return atomicWriteSettings(path, append(updated, '\n'))
}

func atomicWriteSettings(path string, raw []byte) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".settings-*")
	if err != nil {
		return fmt.Errorf("create temporary controlled settings: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure temporary controlled settings: %w", err)
	}
	if _, err := temporary.Write(raw); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary controlled settings: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("sync temporary controlled settings: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary controlled settings: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace controlled settings: %w", err)
	}
	dir, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open controlled settings directory: %w", err)
	}
	syncErr := dir.Sync()
	closeErr := dir.Close()
	if syncErr != nil {
		return fmt.Errorf("sync controlled settings directory: %w", syncErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close controlled settings directory: %w", closeErr)
	}
	return nil
}

type phaseModelPreference struct {
	Provider string
	Model    string
	Thinking string
}

// ValidateModelPreferences requires the official GSD Pi effective policy to
// route each governed phase explicitly. The controller-owned global policy is
// merged with project overrides using GSD's documented per-key precedence.
func ValidateModelPreferences(gsdHome, workDir string, registry UnitRegistry, coordinatorModel, implementationModel, expectedThinking string) error {
	path := filepath.Join(gsdHome, "PREFERENCES.md")
	raw, err := readGovernedPolicyFile(path)
	if err != nil {
		return fmt.Errorf("read controlled GSD preferences: %w", err)
	}
	models, err := parsePhaseModelPreferences(raw, true)
	if err != nil {
		return fmt.Errorf("decode controlled GSD model preferences: %w", err)
	}
	projectPath := filepath.Join(workDir, ".gsd", "PREFERENCES.md")
	projectRaw, projectErr := readGovernedPolicyFile(projectPath)
	if projectErr != nil && !errors.Is(projectErr, os.ErrNotExist) {
		return fmt.Errorf("read project GSD preferences: %w", projectErr)
	}
	if projectErr == nil {
		projectModels, parseErr := parsePhaseModelPreferences(projectRaw, false)
		if parseErr != nil {
			return fmt.Errorf("decode project GSD model preferences: %w", parseErr)
		}
		for phase, preference := range projectModels {
			models[phase] = preference
		}
	}
	phaseRoles, err := registry.GovernedPhases()
	if err != nil {
		return err
	}
	// GSD resolves delegated implementation outside UNIT_REGISTRY through its
	// explicit subagent phase. Keep that narrow controller policy separate from
	// official unit metadata.
	phaseRoles["subagent"] = ModelRoleImplementation
	for phase, role := range phaseRoles {
		qualifiedModel := coordinatorModel
		if role == ModelRoleImplementation {
			qualifiedModel = implementationModel
		}
		provider, model, ok := strings.Cut(qualifiedModel, "/")
		if !ok || provider == "" || model == "" {
			return fmt.Errorf("governed model for phase %s must be provider-qualified", phase)
		}
		observed, ok := models[phase]
		if !ok {
			return fmt.Errorf("effective GSD preferences do not pin the %s phase", phase)
		}
		if observed.Provider != provider || observed.Model != model || observed.Thinking != expectedThinking {
			return fmt.Errorf("effective GSD %s phase is %s/%s/%s, expected %s/%s/%s", phase,
				observed.Provider, observed.Model, observed.Thinking, provider, model, expectedThinking)
		}
	}
	return nil
}

func parsePhaseModelPreferences(raw []byte, requireModels bool) (map[string]phaseModelPreference, error) {
	lines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	if len(lines) < 3 || strings.TrimSpace(lines[0]) != "---" {
		return nil, errors.New("preferences must start with YAML frontmatter")
	}
	frontmatterEnd := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			frontmatterEnd = i
			break
		}
	}
	if frontmatterEnd < 0 {
		return nil, errors.New("preferences frontmatter is not closed")
	}

	fields := make(map[string]map[string]string)
	inModels := false
	modelsSeen := false
	currentPhase := ""
	for lineNumber := 1; lineNumber < frontmatterEnd; lineNumber++ {
		line := strings.TrimRight(lines[lineNumber], " \t")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.ContainsRune(line, '\t') {
			return nil, fmt.Errorf("line %d contains a tab", lineNumber+1)
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 {
			currentPhase = ""
			key, value, found := strings.Cut(trimmed, ":")
			if !found {
				if inModels {
					return nil, fmt.Errorf("line %d is not a YAML field", lineNumber+1)
				}
				continue
			}
			if key == "models" {
				if modelsSeen {
					return nil, errors.New("duplicate models block")
				}
				modelsSeen = true
				modelValue := strings.TrimSpace(value)
				if modelValue == "{}" {
					inModels = true
					continue
				}
				if modelValue != "" {
					return nil, errors.New("models must be an expanded mapping")
				}
				inModels = true
				continue
			}
			if inModels {
				inModels = false
			}
			continue
		}
		if !inModels {
			continue
		}
		if indent == 2 {
			phase, value, found := strings.Cut(trimmed, ":")
			phase = strings.TrimSpace(phase)
			if !found || phase == "" {
				return nil, fmt.Errorf("line %d has an invalid model phase", lineNumber+1)
			}
			if _, duplicate := fields[phase]; duplicate {
				return nil, fmt.Errorf("duplicate model phase %q", phase)
			}
			fields[phase] = make(map[string]string)
			currentPhase = phase
			value = strings.TrimSpace(value)
			if value != "" {
				parsed, parseErr := parseInlineModelPreference(value)
				if parseErr != nil {
					return nil, fmt.Errorf("phase %s: %w", phase, parseErr)
				}
				fields[phase] = parsed
				currentPhase = ""
			}
			continue
		}
		if indent != 4 || currentPhase == "" {
			return nil, fmt.Errorf("line %d has unsupported model preference indentation", lineNumber+1)
		}
		key, value, found := strings.Cut(trimmed, ":")
		if !found {
			return nil, fmt.Errorf("line %d has an invalid model field", lineNumber+1)
		}
		if err := addModelField(fields[currentPhase], strings.TrimSpace(key), strings.Trim(strings.TrimSpace(value), "\"'")); err != nil {
			return nil, fmt.Errorf("phase %s: %w", currentPhase, err)
		}
	}
	if !modelsSeen && requireModels {
		return nil, errors.New("preferences do not contain a models block")
	}

	models := make(map[string]phaseModelPreference, len(fields))
	for phase, values := range fields {
		models[phase] = phaseModelPreference{Provider: values["provider"], Model: values["model"], Thinking: values["thinking"]}
	}
	return models, nil
}

func parseInlineModelPreference(value string) (map[string]string, error) {
	if !strings.HasPrefix(value, "{") || !strings.HasSuffix(value, "}") {
		return nil, errors.New("model phase must use an object with provider, model, and thinking")
	}
	fields := make(map[string]string)
	for _, part := range strings.Split(strings.TrimSpace(value[1:len(value)-1]), ",") {
		key, fieldValue, found := strings.Cut(strings.TrimSpace(part), ":")
		if !found {
			return nil, errors.New("inline model field is invalid")
		}
		if err := addModelField(fields, strings.TrimSpace(key), strings.Trim(strings.TrimSpace(fieldValue), "\"'")); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

func addModelField(fields map[string]string, key, value string) error {
	if key != "provider" && key != "model" && key != "thinking" {
		return fmt.Errorf("unsupported model field %q", key)
	}
	if value == "" {
		return fmt.Errorf("model field %q is empty", key)
	}
	if _, duplicate := fields[key]; duplicate {
		return fmt.Errorf("duplicate model field %q", key)
	}
	fields[key] = value
	return nil
}

func ValidatePinnedContainer(ctx context.Context, config ContainerConfig, expectedVersion string) error {
	cmd := exec.CommandContext(ctx, config.Engine, "run", "--rm", "--pull=never", "--network=none", config.Image, "--version")
	configureProcessTree(cmd)
	var stdout, stderr boundedBuffer
	stdout.limit, stderr.limit = 64*1024, 64*1024
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run pinned GSD container: %w", err)
	}
	if stdout.Truncated() || stderr.Truncated() {
		return errors.New("pinned GSD container version output is oversized")
	}
	return validateContainerVersionOutput(stdout.String(), expectedVersion)
}

func readGovernedPolicyFile(path string) ([]byte, error) {
	const maxPolicyBytes int64 = 1024 * 1024
	before, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if before.Mode()&os.ModeSymlink != 0 || !before.Mode().IsRegular() || !runtimePathOwnedByCurrentUser(before) || before.Size() < 0 || before.Size() > maxPolicyBytes {
		return nil, errors.New("governed policy must be a bounded owned regular non-symlink file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(before, opened) {
		return nil, errors.New("governed policy changed before reading")
	}
	raw, err := io.ReadAll(io.LimitReader(file, maxPolicyBytes+1))
	if err != nil || int64(len(raw)) != opened.Size() || int64(len(raw)) > maxPolicyBytes {
		return nil, errors.New("governed policy is unreadable or oversized")
	}
	after, statErr := file.Stat()
	pathAfter, pathErr := os.Lstat(path)
	if statErr != nil || pathErr != nil || !os.SameFile(opened, after) || !os.SameFile(after, pathAfter) || opened.Size() != after.Size() || !opened.ModTime().Equal(after.ModTime()) {
		return nil, errors.New("governed policy changed while reading")
	}
	return raw, nil
}

func validateContainerVersionOutput(output, expectedVersion string) error {
	if strings.TrimSpace(output) != expectedVersion {
		return fmt.Errorf("container does not report exact GSD version %s", expectedVersion)
	}
	return nil
}

func ValidatePinnedCommand(command []string, expectedVersion string) error {
	if len(command) != 2 || filepath.Base(command[0]) != "node" || !filepath.IsAbs(command[1]) {
		return errors.New("GSD command must be node plus an absolute pinned loader path")
	}
	loader, err := filepath.EvalSymlinks(command[1])
	if err != nil {
		return fmt.Errorf("resolve GSD loader: %w", err)
	}
	if filepath.Base(loader) != "loader.js" || filepath.Base(filepath.Dir(loader)) != "dist" {
		return errors.New("GSD command does not target the packaged loader")
	}
	packagePath := filepath.Join(filepath.Dir(filepath.Dir(loader)), "package.json")
	raw, err := readGovernedPolicyFile(packagePath)
	if err != nil {
		return fmt.Errorf("read GSD package metadata: %w", err)
	}
	var metadata struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return fmt.Errorf("decode GSD package metadata: %w", err)
	}
	if metadata.Name != "@opengsd/gsd-pi" || metadata.Version != expectedVersion {
		return fmt.Errorf("GSD package is %s@%s, expected @opengsd/gsd-pi@%s", metadata.Name, metadata.Version, expectedVersion)
	}
	return nil
}
