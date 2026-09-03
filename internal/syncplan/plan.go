// Package syncplan owns immutable resolver inputs and outputs. It does not
// select executors or construct runtime resources.
package syncplan

import (
	"fmt"
	"strings"

	"polymetrics.ai/internal/synccontract"
)

const ContractVersion uint = 1

type ResultKind string

const (
	ResultKindExecutable    ResultKind = "executable"
	ResultKindIncompatible  ResultKind = "incompatible"
	ResultKindFoundationGap ResultKind = "foundation_gap"
)

// Plan is the sealed, serializable resolver output shape. CP03 supplies its
// digest and construction; CP02 freezes its facts and discriminants.
type Plan struct {
	ContractVersion uint                       `json:"contract_version"`
	SourceBinding   string                     `json:"source_binding"`
	TargetBinding   string                     `json:"target_binding"`
	Mode            synccontract.Mode          `json:"mode"`
	Axes            synccontract.ExecutionAxes `json:"axes"`
}

func (p Plan) Validate() error {
	if p.ContractVersion != ContractVersion {
		return fmt.Errorf("unsupported sync plan contract version %d", p.ContractVersion)
	}
	if strings.TrimSpace(p.SourceBinding) == "" || strings.TrimSpace(p.TargetBinding) == "" {
		return fmt.Errorf("sync plan source_binding and target_binding are required")
	}
	if err := p.Mode.Validate(); err != nil {
		return err
	}
	return p.Axes.Validate()
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
		return r.Plan.Validate()
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
	case "executor", "progression", "apply", "object", "binding", "key", "delete", "budget":
		return true
	default:
		return false
	}
}
