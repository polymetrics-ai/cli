package rlm

import (
	"encoding/json"
	"fmt"
)

// Spec defines the scoring specification for an RLM run.
// Spec files are JSON (stdlib encoding/json — no YAML dependency).
type Spec struct {
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Features    []Feature `json:"features"`
}

// Feature describes one scoring dimension within a Spec.
type Feature struct {
	Name       string   `json:"name"`                   // source field name in warehouse record
	Weight     float64  `json:"weight"`                 // relative weight; all weights summed = raw total
	ScoreIfSet float64  `json:"score_if_set,omitempty"` // score to assign if field is non-empty/non-zero
	ScoreIfGT  *float64 `json:"score_if_gt,omitempty"`  // score if numeric field > threshold
	Threshold  *float64 `json:"threshold,omitempty"`    // used with ScoreIfGT
	Default    float64  `json:"default,omitempty"`      // score when condition is false
}

// ParseSpec parses a JSON spec file and validates it.
// Returns an error if the spec is invalid.
func ParseSpec(data []byte) (*Spec, error) {
	var s Spec
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("rlm: parse spec: %w", err)
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// Validate checks the Spec for structural correctness.
func (s *Spec) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("rlm: spec missing required field 'name'")
	}
	if len(s.Features) == 0 {
		return fmt.Errorf("rlm: spec must have at least one feature")
	}
	for i, f := range s.Features {
		if f.Weight < 0 {
			return fmt.Errorf("rlm: feature[%d] %q has negative weight %f", i, f.Name, f.Weight)
		}
		if f.ScoreIfGT != nil && f.Threshold == nil {
			return fmt.Errorf("rlm: feature[%d] %q has score_if_gt but missing threshold", i, f.Name)
		}
	}
	return nil
}
