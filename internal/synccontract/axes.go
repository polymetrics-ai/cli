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
	DestinationApplyUpsert        DestinationApply = "upsert"
	DestinationApplyDedupe        DestinationApply = "dedupe"
	DestinationApplyDedupeHistory DestinationApply = "dedupe_history"
	DestinationApplyChangeApply   DestinationApply = "change_apply"
)

// ObjectClass distinguishes physical records from append-only events.
type ObjectClass string

const (
	ObjectClassObject ObjectClass = "object"
	ObjectClassEvent  ObjectClass = "event"
)

// BindingKind identifies a source-backed stream or action binding.
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

// RetryPolicy closes automatic execution retry behavior.
type RetryPolicy string

const (
	RetryPolicySingleAttempt RetryPolicy = "single_attempt"
	RetryPolicyRetrySafe     RetryPolicy = "retry_safe"
)

// IdempotencyPolicy records the durable retry identity promised by a plan.
type IdempotencyPolicy string

const (
	IdempotencyPolicyNone  IdempotencyPolicy = "none"
	IdempotencyPolicyKeyed IdempotencyPolicy = "keyed"
)

// ReceiptPolicy records whether an execution produces a durable receipt.
type ReceiptPolicy string

const (
	ReceiptPolicyNone    ReceiptPolicy = "none"
	ReceiptPolicyDurable ReceiptPolicy = "durable"
)

// AcknowledgementPolicy records the required downstream acknowledgement.
type AcknowledgementPolicy string

const (
	AcknowledgementNone             AcknowledgementPolicy = "none"
	AcknowledgementDurableWarehouse AcknowledgementPolicy = "durable_warehouse"
)

// CheckpointPolicy records when a source position may advance.
type CheckpointPolicy string

const (
	CheckpointNone                 CheckpointPolicy = "none"
	CheckpointAfterAcknowledgement CheckpointPolicy = "after_acknowledgement"
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
		return &AxisError{Axis: "budget", Code: "negative_budget"}
	}
	return nil
}

// ExecutionAxes contains the independent, closed facts consumed by the pure
// resolver. It carries no executor, transport, credential, or I/O handle.
type ExecutionAxes struct {
	Progression     SourceProgression     `json:"progression"`
	Apply           DestinationApply      `json:"apply"`
	Object          ObjectClass           `json:"object"`
	Key             KeyShape              `json:"key"`
	Delete          DeletePolicy          `json:"delete"`
	Retry           RetryPolicy           `json:"retry"`
	Idempotency     IdempotencyPolicy     `json:"idempotency"`
	Receipt         ReceiptPolicy         `json:"receipt"`
	Acknowledgement AcknowledgementPolicy `json:"acknowledgement"`
	Checkpoint      CheckpointPolicy      `json:"checkpoint"`
	Budget          Budget                `json:"budget"`
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
	if !containsKeyShape(a.Key) {
		return &UnknownAxisValueError{Axis: "key", Value: string(a.Key)}
	}
	if !containsDeletePolicy(a.Delete) {
		return &UnknownAxisValueError{Axis: "delete", Value: string(a.Delete)}
	}
	if !containsRetryPolicy(a.Retry) {
		return &UnknownAxisValueError{Axis: "retry", Value: string(a.Retry)}
	}
	if !containsIdempotencyPolicy(a.Idempotency) {
		return &UnknownAxisValueError{Axis: "idempotency", Value: string(a.Idempotency)}
	}
	if !containsReceiptPolicy(a.Receipt) {
		return &UnknownAxisValueError{Axis: "receipt", Value: string(a.Receipt)}
	}
	if !containsAcknowledgementPolicy(a.Acknowledgement) {
		return &UnknownAxisValueError{Axis: "acknowledgement", Value: string(a.Acknowledgement)}
	}
	if !containsCheckpointPolicy(a.Checkpoint) {
		return &UnknownAxisValueError{Axis: "checkpoint", Value: string(a.Checkpoint)}
	}
	return a.Budget.Validate()
}

// ModeAxes returns the one canonical durability contract for mode. Binding
// references remain plan-level because a stream may target either a warehouse
// stream or a declared action without changing the mode's durability rules.
func ModeAxes(mode Mode) (ExecutionAxes, bool) {
	switch mode {
	case ModeFullOverwrite:
		return ExecutionAxes{Progression: SourceProgressionSnapshot, Apply: DestinationApplyReplace, Object: ObjectClassObject, Key: KeyShapeNone, Delete: DeletePolicyNone, Retry: RetryPolicySingleAttempt, Idempotency: IdempotencyPolicyNone, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	case ModeFullAppend:
		return ExecutionAxes{Progression: SourceProgressionSnapshot, Apply: DestinationApplyAppend, Object: ObjectClassObject, Key: KeyShapeNone, Delete: DeletePolicyNone, Retry: RetryPolicySingleAttempt, Idempotency: IdempotencyPolicyNone, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	case ModeIncrementalAppend:
		return ExecutionAxes{Progression: SourceProgressionCursor, Apply: DestinationApplyAppend, Object: ObjectClassEvent, Key: KeyShapeNone, Delete: DeletePolicyNone, Retry: RetryPolicySingleAttempt, Idempotency: IdempotencyPolicyNone, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	case ModeIncrementalUpsert:
		return ExecutionAxes{Progression: SourceProgressionCursor, Apply: DestinationApplyUpsert, Object: ObjectClassObject, Key: KeyShapeStable, Delete: DeletePolicyNone, Retry: RetryPolicyRetrySafe, Idempotency: IdempotencyPolicyKeyed, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	case ModeIncrementalDedupe:
		return ExecutionAxes{Progression: SourceProgressionCursor, Apply: DestinationApplyDedupe, Object: ObjectClassObject, Key: KeyShapeStable, Delete: DeletePolicyNone, Retry: RetryPolicyRetrySafe, Idempotency: IdempotencyPolicyKeyed, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	case ModeIncrementalDedupeHistory:
		return ExecutionAxes{Progression: SourceProgressionCursor, Apply: DestinationApplyDedupeHistory, Object: ObjectClassObject, Key: KeyShapeComposite, Delete: DeletePolicyHistoryClose, Retry: RetryPolicyRetrySafe, Idempotency: IdempotencyPolicyKeyed, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	case ModeChangeCapture:
		return ExecutionAxes{Progression: SourceProgressionChangeCapture, Apply: DestinationApplyChangeApply, Object: ObjectClassEvent, Key: KeyShapeStable, Delete: DeletePolicyTombstone, Retry: RetryPolicyRetrySafe, Idempotency: IdempotencyPolicyKeyed, Receipt: ReceiptPolicyDurable, Acknowledgement: AcknowledgementDurableWarehouse, Checkpoint: CheckpointAfterAcknowledgement}, true
	default:
		return ExecutionAxes{}, false
	}
}

// ValidateModeAxes rejects any cross-axis combination that does not implement
// the exact named mode before an executor or I/O boundary exists.
func ValidateModeAxes(mode Mode, axes ExecutionAxes) error {
	if err := mode.Validate(); err != nil {
		return &AxisError{Axis: "mode", Code: "invalid_mode"}
	}
	if err := axes.Validate(); err != nil {
		return err
	}
	want, _ := ModeAxes(mode)
	for _, comparison := range []struct {
		axis string
		want any
		got  any
	}{
		{"progression", want.Progression, axes.Progression},
		{"apply", want.Apply, axes.Apply},
		{"object", want.Object, axes.Object},
		{"key", want.Key, axes.Key},
		{"delete", want.Delete, axes.Delete},
		{"retry", want.Retry, axes.Retry},
		{"idempotency", want.Idempotency, axes.Idempotency},
		{"receipt", want.Receipt, axes.Receipt},
		{"acknowledgement", want.Acknowledgement, axes.Acknowledgement},
		{"checkpoint", want.Checkpoint, axes.Checkpoint},
	} {
		if comparison.want != comparison.got {
			return &AxisError{Axis: comparison.axis, Code: "mode_axis_mismatch"}
		}
	}
	return nil
}

// AxisError is stable machine-readable pre-I/O validation output for a known
// axis whose values do not satisfy a mode or durability invariant.
type AxisError struct {
	Axis string
	Code string
}

func (e *AxisError) Error() string {
	if e == nil {
		return "invalid sync execution axis"
	}
	return fmt.Sprintf("invalid sync execution %s: %s", e.Axis, e.Code)
}

// SyncAxis identifies the closed axis for resolver classification.
func (e *AxisError) SyncAxis() string {
	if e == nil {
		return ""
	}
	return e.Axis
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

// SyncAxis identifies the closed axis for resolver classification.
func (e *UnknownAxisValueError) SyncAxis() string {
	if e == nil {
		return ""
	}
	return e.Axis
}

func containsSourceProgression(value SourceProgression) bool {
	return value == SourceProgressionSnapshot || value == SourceProgressionCursor || value == SourceProgressionChangeCapture
}

func containsDestinationApply(value DestinationApply) bool {
	return value == DestinationApplyReplace || value == DestinationApplyAppend || value == DestinationApplyUpsert || value == DestinationApplyDedupe || value == DestinationApplyDedupeHistory || value == DestinationApplyChangeApply
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

func containsRetryPolicy(value RetryPolicy) bool {
	return value == RetryPolicySingleAttempt || value == RetryPolicyRetrySafe
}

func containsIdempotencyPolicy(value IdempotencyPolicy) bool {
	return value == IdempotencyPolicyNone || value == IdempotencyPolicyKeyed
}

func containsReceiptPolicy(value ReceiptPolicy) bool {
	return value == ReceiptPolicyNone || value == ReceiptPolicyDurable
}

func containsAcknowledgementPolicy(value AcknowledgementPolicy) bool {
	return value == AcknowledgementNone || value == AcknowledgementDurableWarehouse
}

func containsCheckpointPolicy(value CheckpointPolicy) bool {
	return value == CheckpointNone || value == CheckpointAfterAcknowledgement
}
