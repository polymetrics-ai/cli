// Package syncplan owns immutable resolver inputs and outputs. It does not
// select executors or construct runtime resources.
package syncplan

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/synccontract"
)

const ContractVersion uint = 2

type ResultKind string

const (
	ResultKindExecutable    ResultKind = "executable"
	ResultKindIncompatible  ResultKind = "incompatible"
	ResultKindFoundationGap ResultKind = "foundation_gap"
)

// BindingRef names one immutable stream or action binding without carrying a
// runtime handle, route, credential, or executable code.
type BindingRef struct {
	Kind synccontract.BindingKind `json:"kind"`
	ID   string                   `json:"id"`
}

func (b BindingRef) Validate() error {
	if !validBindingKind(b.Kind) || strings.TrimSpace(b.ID) == "" {
		return &ValidationError{Axis: "binding", Code: "invalid_binding"}
	}
	return nil
}

// ExecutorRole identifies the one declared executor selected for each side of
// a saved synchronization plan.
type ExecutorRole string

const (
	ExecutorRoleSource      ExecutorRole = "source"
	ExecutorRoleDestination ExecutorRole = "destination"
)

// ExecutorRef is one selected source or destination executor identity and its
// exact immutable digest.
type ExecutorRef struct {
	Role   ExecutorRole `json:"role"`
	ID     string       `json:"id"`
	Digest string       `json:"digest"`
}

func (e ExecutorRef) Validate() error {
	if (e.Role != ExecutorRoleSource && e.Role != ExecutorRoleDestination) || strings.TrimSpace(e.ID) == "" || !validDigest(e.Digest) {
		return &ValidationError{Axis: "executor", Code: "invalid_executor"}
	}
	return nil
}

// FoundationRef is the immutable foundation fact supplied by the Foundation
// Atlas projection. Unavailable foundations resolve to C4 without I/O.
type FoundationRef struct {
	ID        string `json:"id"`
	Digest    string `json:"digest"`
	Available bool   `json:"available"`
	Reference string `json:"reference"`
}

func (f FoundationRef) Validate() error {
	if strings.TrimSpace(f.ID) == "" || !validDigest(f.Digest) || strings.TrimSpace(f.Reference) == "" {
		return &ValidationError{Axis: "foundation", Code: "invalid_foundation"}
	}
	return nil
}

// Plan is the sealed, serializable resolver input and executable output shape.
// It binds source/target artifacts and every construction-relevant immutable
// digest before any runtime boundary exists.
type Plan struct {
	ContractVersion  uint                       `json:"contract_version"`
	Source           BindingRef                 `json:"source"`
	Target           BindingRef                 `json:"target"`
	Mode             synccontract.Mode          `json:"mode"`
	Axes             synccontract.ExecutionAxes `json:"axes"`
	GenerationDigest string                     `json:"generation_digest"`
	ArtifactDigest   string                     `json:"artifact_digest"`
	Executors        []ExecutorRef              `json:"executors"`
	Foundation       FoundationRef              `json:"foundation"`
	EvidenceDigest   string                     `json:"evidence_digest"`
}

func (p Plan) Validate() error {
	if p.ContractVersion != ContractVersion {
		return &ValidationError{Axis: "contract_version", Code: "unsupported_contract_version"}
	}
	if err := p.Source.Validate(); err != nil {
		return err
	}
	if err := p.Target.Validate(); err != nil {
		return err
	}
	if err := p.Mode.Validate(); err != nil {
		return &ValidationError{Axis: "mode", Code: "invalid_mode"}
	}
	if err := p.Axes.Validate(); err != nil {
		return err
	}
	if !validDigest(p.GenerationDigest) || !validDigest(p.ArtifactDigest) || !validDigest(p.EvidenceDigest) {
		return &ValidationError{Axis: "identity", Code: "invalid_digest"}
	}
	if err := p.Foundation.Validate(); err != nil {
		return err
	}
	for index, executor := range p.Executors {
		if err := executor.Validate(); err != nil {
			return err
		}
		if index > 0 && executorKey(p.Executors[index-1]) >= executorKey(executor) {
			return &ValidationError{Axis: "executor", Code: "executors_not_strictly_sorted"}
		}
	}
	return nil
}

// ValidateExecutable adds the pre-I/O selection and durability invariants that
// distinguish an executable result from C3/C4 input facts.
func (p Plan) ValidateExecutable() error {
	if err := p.Validate(); err != nil {
		return err
	}
	if !p.Foundation.Available {
		return &ValidationError{Axis: "foundation", Code: "foundation_unavailable"}
	}
	if !hasExactlyOneExecutorPerRole(p.Executors) {
		return &ValidationError{Axis: "executor", Code: "ambiguous_executor_selection"}
	}
	if p.Source.Kind != synccontract.BindingKindStream || (p.Target.Kind != synccontract.BindingKindStream && p.Target.Kind != synccontract.BindingKindAction) {
		return &ValidationError{Axis: "binding", Code: "incompatible_binding"}
	}
	return synccontract.ValidateModeAxes(p.Mode, p.Axes)
}

// ValidationError identifies the real closed axis responsible for a C3 result.
type ValidationError struct {
	Axis string
	Code string
}

func (e *ValidationError) Error() string {
	if e == nil {
		return "invalid sync plan"
	}
	return fmt.Sprintf("invalid sync plan %s: %s", e.Axis, e.Code)
}

// SyncAxis identifies the closed axis for resolver classification.
func (e *ValidationError) SyncAxis() string {
	if e == nil {
		return ""
	}
	return e.Axis
}

type Incompatibility struct {
	Axis string `json:"axis"`
	Code string `json:"code"`
}

func (i Incompatibility) Validate() error {
	if !validAxis(i.Axis) || strings.TrimSpace(i.Code) == "" {
		return fmt.Errorf("invalid sync incompatibility")
	}
	return nil
}

type FoundationGap struct {
	FoundationID string `json:"foundation_id"`
	Reference    string `json:"reference"`
}

func (g FoundationGap) Validate() error {
	if strings.TrimSpace(g.FoundationID) == "" || strings.TrimSpace(g.Reference) == "" {
		return fmt.Errorf("sync foundation gap requires foundation_id and reference")
	}
	return nil
}

// Result is one and only one pre-I/O resolution outcome.
type Result struct {
	Kind            ResultKind       `json:"kind"`
	Plan            *Plan            `json:"plan,omitempty"`
	Incompatibility *Incompatibility `json:"incompatibility,omitempty"`
	FoundationGap   *FoundationGap   `json:"foundation_gap,omitempty"`
}

func (r Result) Validate() error {
	switch r.Kind {
	case ResultKindExecutable:
		if r.Plan == nil || r.Incompatibility != nil || r.FoundationGap != nil {
			return fmt.Errorf("executable sync result requires only a plan")
		}
		return r.Plan.ValidateExecutable()
	case ResultKindIncompatible:
		if r.Plan != nil || r.Incompatibility == nil || r.FoundationGap != nil {
			return fmt.Errorf("incompatible sync result requires only an incompatibility")
		}
		return r.Incompatibility.Validate()
	case ResultKindFoundationGap:
		if r.Plan != nil || r.Incompatibility != nil || r.FoundationGap == nil {
			return fmt.Errorf("foundation-gap sync result requires only a foundation gap")
		}
		return r.FoundationGap.Validate()
	default:
		return fmt.Errorf("unknown sync result kind %q", r.Kind)
	}
}

func validAxis(axis string) bool {
	switch axis {
	case "contract_version", "mode", "binding", "executor", "foundation", "identity", "progression", "apply", "object", "key", "delete", "retry", "idempotency", "receipt", "acknowledgement", "checkpoint", "budget":
		return true
	default:
		return false
	}
}

func validBindingKind(kind synccontract.BindingKind) bool {
	return kind == synccontract.BindingKindStream || kind == synccontract.BindingKindAction
}

func validDigest(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	for _, rune := range value[len("sha256:"):] {
		if !(rune >= '0' && rune <= '9') && !(rune >= 'a' && rune <= 'f') {
			return false
		}
	}
	return true
}

func executorKey(executor ExecutorRef) string {
	return string(executor.Role) + "\x00" + executor.ID + "\x00" + executor.Digest
}

func hasExactlyOneExecutorPerRole(executors []ExecutorRef) bool {
	if len(executors) != 2 {
		return false
	}
	return executors[0].Role == ExecutorRoleDestination && executors[1].Role == ExecutorRoleSource
}
