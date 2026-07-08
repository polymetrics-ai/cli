package certify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type SurfaceResult struct {
	Result          string         `json:"result"`
	Endpoints       int            `json:"endpoints"`
	Covered         int            `json:"covered"`
	Blocked         int            `json:"blocked"`
	CoveredBy       map[string]int `json:"covered_by,omitempty"`
	BlockedByModel  map[string]int `json:"blocked_by_model,omitempty"`
	BlockedByStatus map[string]int `json:"blocked_by_status,omitempty"`
	Reason          string         `json:"reason,omitempty"`
}

type apiSurfaceFile struct {
	Endpoints []apiSurfaceEndpoint `json:"endpoints"`
}

type apiSurfaceEndpoint struct {
	CoveredBy map[string]any       `json:"covered_by"`
	Operation *apiSurfaceOperation `json:"operation"`
}

type apiSurfaceOperation struct {
	Model  string `json:"model"`
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func stageSurfaceInventory(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		skipStage(rc, rep, "surface_inventory", "skipped: --full not set (surface inventory is full-certificate only)")
		return nil
	}

	recordStage(rc, rep, "surface_inventory", 0, func() (bool, CLIStageInfo, string) {
		result, err := surfaceInventoryFor(rc.opts.Connector)
		if errors.Is(err, fs.ErrNotExist) {
			reason := fmt.Sprintf("skipped: connector %q has no embedded api_surface inventory", rc.opts.Connector)
			rep.Capabilities.Surface = &SurfaceResult{Result: "skipped", Reason: reason}
			return false, CLIStageInfo{}, reason
		}
		if err != nil {
			rep.Capabilities.Surface = &SurfaceResult{Result: "fail", Reason: err.Error()}
			return false, CLIStageInfo{}, err.Error()
		}
		rep.Capabilities.Surface = &result
		if result.Result != "pass" {
			return false, CLIStageInfo{}, result.Reason
		}
		return true, CLIStageInfo{}, ""
	})
	return nil
}

func surfaceInventoryFor(connector string) (SurfaceResult, error) {
	path, err := findAPISurfacePath(connector)
	if err != nil {
		return SurfaceResult{}, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return SurfaceResult{}, fmt.Errorf("read %s api surface: %w", connector, err)
	}
	var file apiSurfaceFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return SurfaceResult{}, fmt.Errorf("parse %s api surface: %w", connector, err)
	}
	result := SurfaceResult{
		Result:          "pass",
		Endpoints:       len(file.Endpoints),
		CoveredBy:       map[string]int{},
		BlockedByModel:  map[string]int{},
		BlockedByStatus: map[string]int{},
	}
	for i, endpoint := range file.Endpoints {
		covered := len(endpoint.CoveredBy) > 0
		blocked := endpoint.Operation != nil && endpoint.Operation.Model != "" && endpoint.Operation.Status != "" && endpoint.Operation.Reason != ""
		switch {
		case covered:
			result.Covered++
			for key, value := range endpoint.CoveredBy {
				result.CoveredBy[key] += coveredCount(value)
			}
		case blocked:
			result.Blocked++
			result.BlockedByModel[endpoint.Operation.Model]++
			result.BlockedByStatus[endpoint.Operation.Status]++
		default:
			result.Result = "fail"
			result.Reason = fmt.Sprintf("api_surface endpoint %d is neither covered nor blocked with typed reason", i)
			return result, nil
		}
	}
	return result, nil
}

func findAPISurfacePath(connector string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}
	for {
		candidate := filepath.Join(wd, "internal", "connectors", "defs", connector, "api_surface.json")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("stat %s api surface: %w", connector, err)
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", fmt.Errorf("read %s api surface: %w", connector, fs.ErrNotExist)
		}
		wd = parent
	}
}

func coveredCount(value any) int {
	switch typed := value.(type) {
	case []any:
		return len(typed)
	case []string:
		return len(typed)
	case string:
		if typed == "" {
			return 0
		}
		return 1
	default:
		if value == nil {
			return 0
		}
		return 1
	}
}
