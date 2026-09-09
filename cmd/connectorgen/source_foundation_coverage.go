package main

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"
)

// Coverage describes accounting, never capability. The unassessed complement
// is calculated from independent retained source identities, not supplied JSON.
type sourceFoundationCoverage struct {
	UniverseCount    int                    `json:"universe_count"`
	UniverseSHA256   string                 `json:"universe_sha256"`
	Assessed         []sourceFoundationCell `json:"assessed"`
	AssessedSHA256   string                 `json:"assessed_sha256"`
	Unassessed       []sourceFoundationCell `json:"unassessed"`
	UnassessedSHA256 string                 `json:"unassessed_sha256"`
}

func sourceFoundationCellLess(a, b sourceFoundationCell) bool {
	if a.Key != b.Key {
		return sourceKeyLess(a.Key, b.Key)
	}
	return slices.Index(sourceLaneNames(), a.Lane) < slices.Index(sourceLaneNames(), b.Lane)
}

func sourceFoundationCellsHash(cells []sourceFoundationCell) (string, error) {
	raw, err := json.Marshal(cells)
	if err != nil {
		return "", fmt.Errorf("foundation cell identity encoding: %w", err)
	}
	canonical, err := canonicalSourceJSON(raw)
	if err != nil {
		return "", fmt.Errorf("foundation cell identity canonicalization: %w", err)
	}
	return sourceBytesHash(canonical), nil
}

func buildSourceFoundationCoverage(observed sourceFoundationAssessmentObservations) (sourceFoundationCoverage, error) {
	result := sourceFoundationCoverage{Assessed: []sourceFoundationCell{}, Unassessed: []sourceFoundationCell{}}
	assessed := map[sourceFoundationCell]bool{}
	for _, row := range observed.authored {
		cell := sourceFoundationCell{Key: row.Key, Lane: row.Lane}
		if assessed[cell] {
			return result, fmt.Errorf("foundation coverage duplicate assessed cell")
		}
		assessed[cell] = true
	}
	universe := append([]sourceFoundationCell{}, observed.universe.cells...)
	sort.Slice(universe, func(i, j int) bool { return sourceFoundationCellLess(universe[i], universe[j]) })
	seen := map[sourceFoundationCell]bool{}
	for _, cell := range universe {
		if seen[cell] {
			return result, fmt.Errorf("foundation coverage duplicate universe cell")
		}
		seen[cell] = true
		if assessed[cell] {
			result.Assessed = append(result.Assessed, cell)
		} else {
			result.Unassessed = append(result.Unassessed, cell)
		}
	}
	if len(result.Assessed) != len(assessed) {
		return result, fmt.Errorf("foundation coverage assessment outside universe")
	}
	result.UniverseCount = len(universe)
	var err error
	result.UniverseSHA256, err = sourceFoundationCellsHash(universe)
	if err != nil {
		return result, err
	}
	result.AssessedSHA256, err = sourceFoundationCellsHash(result.Assessed)
	if err != nil {
		return result, err
	}
	result.UnassessedSHA256, err = sourceFoundationCellsHash(result.Unassessed)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Validate emitted membership against the original retained manifest and
// independently admitted assessments. Agreement between a forged array and
// its own recomputed digest cannot satisfy this check.
func validateSourceFoundationCoverage(candidate sourceFoundationCoverage, observed sourceFoundationAssessmentObservations) error {
	expected := map[sourceFoundationCell]bool{}
	for _, row := range observed.universe.manifest.SourceOperations {
		for _, lane := range sourceLaneNames() {
			expected[sourceFoundationCell{Key: row.Source.Key, Lane: lane}] = false
		}
	}
	for _, row := range observed.authored {
		cell := sourceFoundationCell{Key: row.Key, Lane: row.Lane}
		if _, exists := expected[cell]; !exists {
			return fmt.Errorf("foundation coverage authored identity outside retained sources")
		}
		expected[cell] = true
	}
	if candidate.UniverseCount != len(expected) {
		return fmt.Errorf("foundation coverage universe count mismatch")
	}
	seen := map[sourceFoundationCell]bool{}
	for _, group := range []struct {
		cells    []sourceFoundationCell
		assessed bool
		digest   string
	}{{candidate.Assessed, true, candidate.AssessedSHA256}, {candidate.Unassessed, false, candidate.UnassessedSHA256}} {
		if group.cells == nil {
			return fmt.Errorf("foundation coverage null membership")
		}
		for i, cell := range group.cells {
			wanted, exists := expected[cell]
			if !exists || seen[cell] || wanted != group.assessed || i > 0 && !sourceFoundationCellLess(group.cells[i-1], cell) {
				return fmt.Errorf("foundation coverage wrong or duplicate membership")
			}
			seen[cell] = true
		}
		digest, err := sourceFoundationCellsHash(group.cells)
		if err != nil || digest != group.digest {
			return fmt.Errorf("foundation coverage membership digest mismatch")
		}
	}
	if len(seen) != len(expected) {
		return fmt.Errorf("foundation coverage omitted source cell")
	}
	all := make([]sourceFoundationCell, 0, len(expected))
	for cell := range expected {
		all = append(all, cell)
	}
	sort.Slice(all, func(i, j int) bool { return sourceFoundationCellLess(all[i], all[j]) })
	digest, err := sourceFoundationCellsHash(all)
	if err != nil || digest != candidate.UniverseSHA256 {
		return fmt.Errorf("foundation coverage universe digest mismatch")
	}
	return nil
}
