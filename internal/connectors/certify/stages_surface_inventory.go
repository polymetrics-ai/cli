package certify

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"polymetrics.ai/internal/connectors/engine"
)

// SurfaceProvenanceResult is the provider-artifact evidence attached to an
// api_surface inventory. It is independent of executable endpoint coverage.
type SurfaceProvenanceResult struct {
	Status         string `json:"status"`
	LedgerVersion  int    `json:"ledger_version"`
	ArtifactCount  int    `json:"artifact_count"`
	EndpointCount  int    `json:"endpoint_count"`
	CitedEndpoints int    `json:"cited_endpoints"`
	Reason         string `json:"reason,omitempty"`
}

type SurfaceResult struct {
	Result          string                   `json:"result"`
	Endpoints       int                      `json:"endpoints"`
	Covered         int                      `json:"covered"`
	Blocked         int                      `json:"blocked"`
	CoveredBy       map[string]int           `json:"covered_by,omitempty"`
	BlockedByModel  map[string]int           `json:"blocked_by_model,omitempty"`
	BlockedByStatus map[string]int           `json:"blocked_by_status,omitempty"`
	Provenance      *SurfaceProvenanceResult `json:"provenance,omitempty"`
	Reason          string                   `json:"reason,omitempty"`
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
	result, err := surfaceInventoryFromRaw(raw)
	if err != nil {
		return SurfaceResult{}, fmt.Errorf("parse %s api surface: %w", connector, err)
	}
	return result, nil
}

func surfaceInventoryFromRaw(raw []byte) (SurfaceResult, error) {
	surface, err := engine.ParseAPISurface(raw)
	if err != nil {
		return SurfaceResult{}, err
	}
	provenance := engine.ValidateSurfaceProvenance(&surface)
	result := SurfaceResult{
		Result:          "pass",
		Endpoints:       len(surface.Endpoints),
		CoveredBy:       map[string]int{},
		BlockedByModel:  map[string]int{},
		BlockedByStatus: map[string]int{},
		Provenance: &SurfaceProvenanceResult{
			Status:         provenance.Status,
			LedgerVersion:  provenance.LedgerVersion,
			ArtifactCount:  provenance.ArtifactCount,
			EndpointCount:  provenance.EndpointCount,
			CitedEndpoints: provenance.CitedEndpoints,
		},
	}
	if provenance.Status == engine.SurfaceProvenanceInvalid {
		result.Result = "fail"
		result.Reason = provenance.Issues[0].Error()
		result.Provenance.Reason = result.Reason
		return result, nil
	}

	for i, endpoint := range surface.Endpoints {
		covered := hasSurfaceCoverage(endpoint.CoveredBy)
		blocked := endpoint.Operation != nil && endpoint.Operation.Model != "" && endpoint.Operation.Status != "" && endpoint.Operation.Reason != ""
		switch {
		case covered:
			result.Covered++
			addSurfaceCoverageCounts(result.CoveredBy, endpoint.CoveredBy)
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

func hasSurfaceCoverage(coverage *engine.SurfaceCoverage) bool {
	return coverage != nil && (coverage.Stream != "" || len(coverage.WriteTargets()) > 0 || coverage.DirectRead != "" || len(coverage.DirectReads) > 0 || len(coverage.OperationTargets()) > 0)
}

func addSurfaceCoverageCounts(counts map[string]int, coverage *engine.SurfaceCoverage) {
	if coverage.Stream != "" {
		counts["stream"]++
	}
	if writes := coverage.WriteTargets(); len(writes) > 0 {
		counts["write"] += len(writes)
	}
	if coverage.DirectRead != "" {
		counts["direct_read"]++
	}
	if len(coverage.DirectReads) > 0 {
		counts["direct_reads"] += len(coverage.DirectReads)
	}
	if operations := coverage.OperationTargets(); len(operations) > 0 {
		counts["operation"] += len(operations)
	}
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
