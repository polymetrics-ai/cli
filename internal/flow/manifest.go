package flow

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

// StepKind is the kind of a flow step.
type StepKind string

const (
	KindSync   StepKind = "sync"
	KindQuery  StepKind = "query"
	KindRLM    StepKind = "rlm"
	KindAction StepKind = "action"
)

// ActionConfig holds the configuration for a step of kind "action".
type ActionConfig struct {
	SourceTable string `json:"source_table"`
	// SourceConnection scopes source_table to one connection's warehouse
	// materialization. _unattributed names a root-owned table. An omitted
	// selector remains intentionally ambiguous when several owners exist.
	SourceConnection      string            `json:"source_connection,omitempty"`
	DestinationConnector  string            `json:"destination_connector"`
	DestinationCredential string            `json:"destination_credential"`
	DestinationConfig     map[string]string `json:"destination_config,omitempty"`
	// DestinationTable is the stable action target bound into authorization.
	DestinationTable string            `json:"destination_table,omitempty"`
	Action           string            `json:"action"` // upsert | create | delete; defaults to "upsert"
	Mappings         map[string]string `json:"mappings"`
	// AuthorizationReference is the opaque durable scope identity created by a
	// prior plan → preview → approval → execute lifecycle.
	AuthorizationReference string `json:"authorization_reference,omitempty"`
	// ReadBackStream is independently read after acknowledgement and before a
	// successful action can be checkpointed.
	ReadBackStream string `json:"read_back_stream,omitempty"`
	MaxRetries     int    `json:"max_retries,omitempty"` // default 3
	BatchSize      int    `json:"batch_size,omitempty"`  // default 100
}

// FlowStep describes a single step in a flow manifest.
type FlowStep struct {
	ID   string   `json:"id"`
	Kind StepKind `json:"kind"`
	// Job is a positive reference to an existing App job. Sync steps resolve it
	// to a connection; action steps resolve it to an already-approved reverse
	// plan. Executable action scope is derived from that job on every run.
	Job string `json:"job,omitempty"`
	// Connection identifies the sync connection for sync steps and scopes the
	// source warehouse views for query steps. _unattributed selects root-owned
	// tables for a query. It remains optional for query steps so an omitted
	// selector is refused as ambiguous rather than guessed.
	Connection string        `json:"connection,omitempty"`
	Streams    []string      `json:"streams,omitempty"`
	SQL        string        `json:"sql,omitempty"`
	Spec       string        `json:"spec,omitempty"`
	Mode       string        `json:"mode,omitempty"`
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

// ManifestDigest returns a stable content digest for one fully resolved flow.
// Prepared action identities bind it so edits to upstream queries, job scope,
// dependencies, or action configuration cannot reuse old prepared evidence.
func ManifestDigest(manifest FlowManifest) (string, error) {
	payload, err := json.Marshal(manifest)
	if err != nil {
		return "", fmt.Errorf("encode flow manifest digest: %w", err)
	}
	digest := sha256.Sum256(payload)
	return fmt.Sprintf("fmd_%x", digest[:]), nil
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

// JobReferenceReason is a stable reason a flow job reference was refused.
type JobReferenceReason string

const (
	JobReferenceMissing      JobReferenceReason = "missing"
	JobReferenceMalformed    JobReferenceReason = "malformed"
	JobReferenceUnapproved   JobReferenceReason = "unapproved"
	JobReferenceUnrecognized JobReferenceReason = "unrecognized"
)

// JobReferenceError names the exact flow step and reference that could not be
// positively resolved. Err retains a more specific authorization refusal.
type JobReferenceError struct {
	Flow      string
	StepID    string
	Kind      StepKind
	Reference string
	Reason    JobReferenceReason
	Err       error
}

func (e *JobReferenceError) Error() string {
	if e == nil {
		return "flow job reference refused"
	}
	message := fmt.Sprintf("flow %q step %q %s job %q is %s", e.Flow, e.StepID, e.Kind, e.Reference, e.Reason)
	if e.Err != nil {
		message += ": " + e.Err.Error()
	}
	return message
}

func (e *JobReferenceError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func IsJobReferenceError(err error, reason JobReferenceReason) bool {
	var referenceErr *JobReferenceError
	return errors.As(err, &referenceErr) && referenceErr.Reason == reason
}

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
			if s.Job == "" && s.Connection == "" {
				add("step %q (sync) must have a job or connection reference", s.ID)
			}
			if len(s.Streams) == 0 {
				add("step %q (sync) must have at least one stream", s.ID)
			}
		case KindQuery:
			if s.SQL == "" {
				add("step %q (query) must have non-empty sql", s.ID)
			}
		case KindRLM:
			mode := s.Mode
			if mode == "" {
				mode = "deterministic"
			}
			if s.Spec == "" {
				add("step %q (rlm) must have spec", s.ID)
			}
			if len(s.Out) == 0 {
				add("step %q (rlm) must have at least one out table", s.ID)
			}
			switch mode {
			case "deterministic", "model", "agent":
				if len(s.In) == 0 {
					add("step %q (rlm) must have at least one input table in %s mode", s.ID, mode)
				}
			case "fixture":
			default:
				add("step %q (rlm) has unknown mode %q", s.ID, mode)
			}
		case KindAction:
			cfg := s.ActionCfg
			if cfg == nil {
				add("step %q (action) must have action_cfg", s.ID)
				break
			}
			if s.Job == "" {
				if cfg.SourceTable == "" {
					add("step %q (action) must have source_table or an approved job reference", s.ID)
				}
				if cfg.DestinationConnector == "" {
					add("step %q (action) must have destination_connector or an approved job reference", s.ID)
				}
				if len(cfg.Mappings) == 0 {
					add("step %q (action) must have at least one mapping or an approved job reference", s.ID)
				}
			}
			// Default action to "upsert" — not an error if empty.
			if cfg.Action == "" && s.Job == "" {
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
