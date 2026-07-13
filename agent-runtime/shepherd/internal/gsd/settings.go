package gsd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func ValidateRuntimeSettings(gsdHome, expectedModel, expectedThinking string) error {
	path := filepath.Join(gsdHome, "agent", "settings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read controlled GSD settings: %w", err)
	}
	var settings struct {
		DefaultProvider      string `json:"defaultProvider"`
		DefaultModel         string `json:"defaultModel"`
		DefaultThinkingLevel string `json:"defaultThinkingLevel"`
	}
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
	return nil
}
