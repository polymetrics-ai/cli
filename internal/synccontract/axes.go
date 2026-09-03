package synccontract

import "fmt"

// SourceProgression is the closed way a source advances through its records.
type SourceProgression string

const (
	SourceProgressionSnapshot      SourceProgression = "snapshot"
	SourceProgressionCursor        SourceProgression = "cursor"
	SourceProgressionChangeCapture SourceProgression = "change_capture"
)

// DestinationApply is the closed way a destination applies source records.
type DestinationApply string

const (
	DestinationApplyReplace       DestinationApply = "replace"
	DestinationApplyAppend        DestinationApply = "append"
	DestinationApplyDedupeHistory DestinationApply = "dedupe_history"
	DestinationApplyChangeApply   DestinationApply = "change_apply"
)

// ObjectClass distinguishes physical records from append-only events.
type ObjectClass string

const (
	ObjectClassObject ObjectClass = "object"
	ObjectClassEvent  ObjectClass = "event"
)

// BindingKind identifies the source-backed execution binding without naming an executor.
type BindingKind string

const (
	BindingKindStream BindingKind = "stream"
	BindingKindAction BindingKind = "action"
)

// KeyShape records the source key guarantee required by an apply mode.
type KeyShape string

const (
	KeyShapeNone      KeyShape = "none"
	KeyShapeStable    KeyShape = "stable"
	KeyShapeComposite KeyShape = "composite"
)

// DeletePolicy records the source deletion semantics carried by a plan.
type DeletePolicy string

const (
	DeletePolicyNone         DeletePolicy = "none"
	DeletePolicyTombstone    DeletePolicy = "tombstone"
	DeletePolicyHistoryClose DeletePolicy = "history_close"
)

// Budget is an immutable upper bound supplied by declarations or callers. A
// later resolver may only reduce these values.
type Budget struct {
	MaxRecords int64 `json:"max_records,omitempty"`
	MaxBytes   int64 `json:"max_bytes,omitempty"`
	MaxBatches int64 `json:"max_batches,omitempty"`
}

func (b Budget) Validate() error {
	if b.MaxRecords < 0 || b.MaxBytes < 0 || b.MaxBatches < 0 {
		return fmt.Errorf("sync budget values must be non-negative")
	}
	return nil
}

// ExecutionAxes contains the independent, closed facts consumed by the later
// resolver. It carries no executor, transport, credential, or I/O handle.
type ExecutionAxes struct {
	Progression SourceProgression `json:"progression"`
	Apply       DestinationApply  `json:"apply"`
	Object      ObjectClass       `json:"object"`
	Binding     BindingKind       `json:"binding"`
	Key         KeyShape          `json:"key"`
	Delete      DeletePolicy      `json:"delete"`
	Budget      Budget            `json:"budget"`
}

func (a ExecutionAxes) Validate() error {
	if !containsSourceProgression(a.Progression) {
		return &UnknownAxisValueError{Axis: "progression", Value: string(a.Progression)}
	}
	if !containsDestinationApply(a.Apply) {
		return &UnknownAxisValueError{Axis: "apply", Value: string(a.Apply)}
	}
	if !containsObjectClass(a.Object) {
		return &UnknownAxisValueError{Axis: "object", Value: string(a.Object)}
	}
	if !containsBindingKind(a.Binding) {
		return &UnknownAxisValueError{Axis: "binding", Value: string(a.Binding)}
	}
	if !containsKeyShape(a.Key) {
		return &UnknownAxisValueError{Axis: "key", Value: string(a.Key)}
	}
	if !containsDeletePolicy(a.Delete) {
		return &UnknownAxisValueError{Axis: "delete", Value: string(a.Delete)}
	}
	return a.Budget.Validate()
}

// UnknownAxisValueError is stable machine-readable pre-I/O validation output.
type UnknownAxisValueError struct {
	Axis  string
	Value string
}

func (e *UnknownAxisValueError) Error() string {
	if e == nil {
		return "unknown sync execution axis value"
	}
	return fmt.Sprintf("unknown sync execution %s %q", e.Axis, e.Value)
}

func containsSourceProgression(value SourceProgression) bool {
	return value == SourceProgressionSnapshot || value == SourceProgressionCursor || value == SourceProgressionChangeCapture
}

func containsDestinationApply(value DestinationApply) bool {
	return value == DestinationApplyReplace || value == DestinationApplyAppend || value == DestinationApplyDedupeHistory || value == DestinationApplyChangeApply
}

func containsObjectClass(value ObjectClass) bool {
	return value == ObjectClassObject || value == ObjectClassEvent
}

func containsBindingKind(value BindingKind) bool {
	return value == BindingKindStream || value == BindingKindAction
}

func containsKeyShape(value KeyShape) bool {
	return value == KeyShapeNone || value == KeyShapeStable || value == KeyShapeComposite
}

func containsDeletePolicy(value DeletePolicy) bool {
	return value == DeletePolicyNone || value == DeletePolicyTombstone || value == DeletePolicyHistoryClose
}
