package flow

import (
	"encoding/json"
	"fmt"
	"regexp"
)

// StepKind is the kind of a flow step.
type StepKind string

const (
	KindSync   StepKind = "sync"
	KindQuery  StepKind = "query"
	KindAction StepKind = "action"
)

// ActionConfig holds the configuration for a step of kind "action".
type ActionConfig struct {
	SourceTable           string            `json:"source_table"`
	DestinationConnector  string            `json:"destination_connector"`
	DestinationCredential string            `json:"destination_credential"`
	DestinationConfig     map[string]string `json:"destination_config,omitempty"`
	Action                string            `json:"action"` // upsert | create | delete; defaults to "upsert"
	Mappings              map[string]string `json:"mappings"`
	MaxRetries            int               `json:"max_retries,omitempty"` // default 3
	BatchSize             int               `json:"batch_size,omitempty"`  // default 100
}

// FlowStep describes a single step in a flow manifest.
type FlowStep struct {
	ID         string        `json:"id"`
	Kind       StepKind      `json:"kind"`
	Connection string        `json:"connection,omitempty"`
	Streams    []string      `json:"streams,omitempty"`
	SQL        string        `json:"sql,omitempty"`
	ActionCfg  *ActionConfig `json:"action_cfg,omitempty"`
	In         []string      `json:"in"`
	Out        []string      `json:"out"`
}

// FlowManifest is the top-level structure of a flow definition.
type FlowManifest struct {
	Version     int        `json:"version"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Steps       []FlowStep `json:"steps"`
}

// ParseManifest decodes a JSON-encoded flow manifest.
// Returns an error if the JSON is malformed.
func ParseManifest(data []byte) (FlowManifest, error) {
	var m FlowManifest
	if err := json.Unmarshal(data, &m); err != nil {
		return FlowManifest{}, err
	}
	return m, nil
}

var validNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// ValidateManifest checks the manifest for rule violations.
// Returns a slice of errors (wrapping ErrManifestInvalid) for every violation found.
func ValidateManifest(m FlowManifest) []error {
	var errs []error
	add := func(format string, args ...any) {
		errs = append(errs, fmt.Errorf("%w: "+format, append([]any{ErrManifestInvalid}, args...)...))
	}

	if m.Version != 1 {
		add("version must be 1, got %d", m.Version)
	}
	if m.Name == "" || !validNameRe.MatchString(m.Name) {
		add("name must be non-empty and contain only alphanumeric, dash, or underscore characters")
	}

	// Build set of all produced tables for `in` validation.
	allOut := map[string]bool{}
	for _, s := range m.Steps {
		for _, t := range s.Out {
			allOut[t] = true
		}
	}

	seenIDs := map[string]bool{}
	for _, s := range m.Steps {
		if s.ID == "" {
			add("step has empty id")
		} else if seenIDs[s.ID] {
			add("duplicate step id %q", s.ID)
		}
		seenIDs[s.ID] = true

		switch s.Kind {
		case KindSync:
			if s.Connection == "" {
				add("step %q (sync) must have a connection", s.ID)
			}
			if len(s.Streams) == 0 {
				add("step %q (sync) must have at least one stream", s.ID)
			}
		case KindQuery:
			if s.SQL == "" {
				add("step %q (query) must have non-empty sql", s.ID)
			}
		case KindAction:
			cfg := s.ActionCfg
			if cfg == nil {
				add("step %q (action) must have action_cfg", s.ID)
				break
			}
			if cfg.SourceTable == "" {
				add("step %q (action) must have source_table", s.ID)
			}
			if cfg.DestinationConnector == "" {
				add("step %q (action) must have destination_connector", s.ID)
			}
			if len(cfg.Mappings) == 0 {
				add("step %q (action) must have at least one mapping", s.ID)
			}
			// Default action to "upsert" — not an error if empty.
			if cfg.Action == "" {
				cfg.Action = "upsert"
			}
		default:
			add("step %q has unknown kind %q", s.ID, s.Kind)
		}

		for _, t := range s.In {
			if !allOut[t] {
				add("step %q references input table %q which is not produced by any step", s.ID, t)
			}
		}
	}

	return errs
}
