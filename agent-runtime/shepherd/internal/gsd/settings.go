package gsd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ValidateRuntimeSettings(gsdHome, workDir, expectedModel, expectedThinking string) error {
	path := filepath.Join(gsdHome, "agent", "settings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read controlled GSD settings: %w", err)
	}
	type runtimeSettings struct {
		DefaultProvider      string `json:"defaultProvider"`
		DefaultModel         string `json:"defaultModel"`
		DefaultThinkingLevel string `json:"defaultThinkingLevel"`
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
	projectRaw, err := os.ReadFile(projectPath)
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
	raw, err := os.ReadFile(packagePath)
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
