package recovery

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
)

const (
	SchemaVersion = 1
	PolicyVersion = 1

	RequiredModel    = "openai-codex/gpt-5.6-sol"
	RequiredThinking = "high"
)

type FailureClass string

const (
	FailureRuntimeContractMismatch FailureClass = "runtime_contract_mismatch"
	FailureModelMismatch           FailureClass = "model_mismatch"
	FailureThinkingMismatch        FailureClass = "thinking_mismatch"
	FailureArtifactMissing         FailureClass = "artifact_missing"
	FailureArtifactInvalid         FailureClass = "artifact_invalid"
	FailureStaleHead               FailureClass = "stale_head"
	FailureDirtyTree               FailureClass = "dirty_tree"
	FailureWriteScopeBreach        FailureClass = "write_scope_breach"
	FailureDeadWorker              FailureClass = "dead_worker"
	FailureSilentTool              FailureClass = "silent_tool"
	FailureInterrupted             FailureClass = "interrupted"
	FailureGitHubPublishUncertain  FailureClass = "github_publish_uncertain"
	FailureOutboxUncertain         FailureClass = "outbox_uncertain"
	FailureValidationFailed        FailureClass = "validation_failed"
	FailureRatificationFailed      FailureClass = "ratification_failed"
	FailureRetryExhausted          FailureClass = "retry_exhausted"
	FailureHumanRequired           FailureClass = "human_required"
	FailureUnknown                 FailureClass = "unknown"
)

var failureClasses = []FailureClass{
	FailureRuntimeContractMismatch,
	FailureModelMismatch,
	FailureThinkingMismatch,
	FailureArtifactMissing,
	FailureArtifactInvalid,
	FailureStaleHead,
	FailureDirtyTree,
	FailureWriteScopeBreach,
	FailureDeadWorker,
	FailureSilentTool,
	FailureInterrupted,
	FailureGitHubPublishUncertain,
	FailureOutboxUncertain,
	FailureValidationFailed,
	FailureRatificationFailed,
	FailureRetryExhausted,
	FailureHumanRequired,
	FailureUnknown,
}

func AllFailureClasses() []FailureClass {
	return append([]FailureClass(nil), failureClasses...)
}

func ParseFailureClass(value string) (FailureClass, error) {
	class := FailureClass(value)
	for _, allowed := range failureClasses {
		if class == allowed {
			return class, nil
		}
	}
	return "", fmt.Errorf("unknown recovery failure class %q", value)
}

type Action string

const (
	ActionRetrySameUnit     Action = "retry_same_unit"
	ActionRetryAfterBackoff Action = "retry_after_backoff"
	ActionRunRecoveryPlan   Action = "run_recovery_plan"
	ActionAwaitDecision     Action = "await_decision"
	ActionBlock             Action = "block"
	ActionFinalHumanGate    Action = "final_human_gate"
)

var actions = []Action{
	ActionRetrySameUnit,
	ActionRetryAfterBackoff,
	ActionRunRecoveryPlan,
	ActionAwaitDecision,
	ActionBlock,
	ActionFinalHumanGate,
}

func ParseAction(value string) (Action, error) {
	action := Action(value)
	for _, allowed := range actions {
		if action == allowed {
			return action, nil
		}
	}
	return "", fmt.Errorf("unknown recovery action %q", value)
}

type Primitive string

const (
	PrimitiveInspectRetainedAttempt    Primitive = "inspect_retained_attempt"
	PrimitiveReconcileAttemptResources Primitive = "reconcile_attempt_resources"
	PrimitiveVerifyExpectedArtifacts   Primitive = "verify_expected_artifacts"
	PrimitiveRetryFreshAttempt         Primitive = "retry_fresh_attempt"
)

var primitives = map[Primitive]struct{}{
	PrimitiveInspectRetainedAttempt:    {},
	PrimitiveReconcileAttemptResources: {},
	PrimitiveVerifyExpectedArtifacts:   {},
	PrimitiveRetryFreshAttempt:         {},
}

type PlanStep struct {
	Primitive Primitive `json:"primitive"`
}

type Failure struct {
	Class      FailureClass
	Reversible bool
}

type classifiedError struct {
	failure Failure
	err     error
}

type decisionError struct {
	failure Failure
	action  Action
	err     error
}

func (e *classifiedError) Error() string { return e.err.Error() }
func (e *classifiedError) Unwrap() error { return e.err }
func (e *decisionError) Error() string   { return e.err.Error() }
func (e *decisionError) Unwrap() error   { return e.err }

func MarkFailure(class FailureClass, reversible bool, err error) error {
	if err == nil {
		err = errors.New("recovery: classified failure")
	}
	if _, parseErr := ParseFailureClass(string(class)); parseErr != nil {
		return &classifiedError{failure: Failure{Class: FailureUnknown}, err: err}
	}
	return &classifiedError{failure: Failure{Class: class, Reversible: reversible}, err: err}
}

func MarkDecision(failure Failure, action Action, err error) error {
	if err == nil {
		err = errors.New("recovery: selected controller action")
	}
	return &decisionError{failure: failure, action: action, err: err}
}

func SelectedAction(err error) (Action, bool) {
	decision, override, ok := selectedDecision(err)
	if !ok || override.Class != "" {
		return "", false
	}
	return decision.action, true
}

func selectedDecision(err error) (*decisionError, Failure, bool) {
	var selected *decisionError
	if !errors.As(err, &selected) {
		return nil, Failure{}, false
	}
	outside := make([]Failure, 0, 2)
	unknownOutside := false
	var walk func(error)
	walk = func(current error) {
		if current == nil || current == selected {
			return
		}
		if _, ok := current.(*decisionError); ok {
			unknownOutside = true
			return
		}
		if typed, ok := current.(*classifiedError); ok {
			outside = append(outside, typed.failure)
			return
		}
		if current == gsd.ErrRuntimeContractMismatch {
			outside = append(outside, Failure{Class: FailureRuntimeContractMismatch})
			return
		}
		if joined, ok := current.(interface{ Unwrap() []error }); ok {
			for _, child := range joined.Unwrap() {
				walk(child)
			}
			return
		}
		if wrapped, ok := current.(interface{ Unwrap() error }); ok {
			walk(wrapped.Unwrap())
			return
		}
		unknownOutside = true
	}
	walk(err)
	if unknownOutside {
		return selected, Failure{Class: FailureUnknown}, false
	}
	if len(outside) > 0 {
		joined := make([]error, 0, len(outside))
		for _, failure := range outside {
			joined = append(joined, MarkFailure(failure.Class, failure.Reversible, errors.New("joined recovery failure")))
		}
		return selected, dominantClassifiedFailure(errors.Join(joined...)), false
	}
	return selected, Failure{}, true
}

func ShouldRetry(err error) bool {
	action, ok := SelectedAction(err)
	return ok && (action == ActionRetrySameUnit || action == ActionRetryAfterBackoff || action == ActionRunRecoveryPlan)
}

func Classify(result gsd.Result) Failure {
	if result.Err == nil {
		return Failure{Class: FailureUnknown}
	}
	if decision, override, ok := selectedDecision(result.Err); decision != nil {
		if ok {
			return decision.failure
		}
		if override.Class != "" {
			return override
		}
	}
	if typed := dominantClassifiedFailure(result.Err); typed.Class != "" {
		return typed
	}
	if result.Terminal == gsd.TerminalBlocked || result.Terminal == gsd.TerminalRejected {
		return Failure{Class: FailureUnknown}
	}
	return Failure{Class: FailureUnknown}
}

func dominantClassifiedFailure(err error) Failure {
	failures := make([]Failure, 0, 2)
	unknownSibling := false
	var walk func(error)
	walk = func(current error) {
		if current == nil {
			return
		}
		if typed, ok := current.(*classifiedError); ok {
			failures = append(failures, typed.failure)
			return
		}
		if _, ok := current.(*exec.ExitError); ok {
			return
		}
		switch current {
		case gsd.ErrRuntimeContractMismatch:
			failures = append(failures, Failure{Class: FailureRuntimeContractMismatch})
			return
		case gsd.ErrSilentTool:
			failures = append(failures, Failure{Class: FailureSilentTool, Reversible: true})
			return
		case gsd.ErrDeadWorker:
			failures = append(failures, Failure{Class: FailureDeadWorker, Reversible: true})
			return
		case context.Canceled, context.DeadlineExceeded:
			failures = append(failures, Failure{Class: FailureInterrupted, Reversible: true})
			return
		}
		if joined, ok := current.(interface{ Unwrap() []error }); ok {
			for _, child := range joined.Unwrap() {
				walk(child)
			}
			return
		}
		if wrapped, ok := current.(interface{ Unwrap() error }); ok {
			walk(wrapped.Unwrap())
			return
		}
		unknownSibling = true
	}
	walk(err)
	if unknownSibling {
		return Failure{Class: FailureUnknown}
	}
	if len(failures) == 0 {
		return Failure{}
	}
	unsafeOrder := []FailureClass{FailureWriteScopeBreach, FailureDirtyTree, FailureModelMismatch,
		FailureThinkingMismatch, FailureRuntimeContractMismatch, FailureStaleHead, FailureRatificationFailed,
		FailureGitHubPublishUncertain, FailureOutboxUncertain, FailureRetryExhausted, FailureHumanRequired, FailureUnknown}
	for _, unsafe := range unsafeOrder {
		for _, failure := range failures {
			if failure.Class == unsafe {
				return Failure{Class: unsafe}
			}
		}
	}
	selected := failures[0]
	for _, failure := range failures[1:] {
		if failure.Class != selected.Class {
			return Failure{Class: FailureUnknown}
		}
		selected.Reversible = selected.Reversible && failure.Reversible
	}
	return selected
}

type Policy struct {
	PlannerEligible bool
	DirectAction    Action
	ExhaustedAction Action
	allowed         map[Action]struct{}
}

func PolicyFor(failure Failure) (Policy, error) {
	if _, err := ParseFailureClass(string(failure.Class)); err != nil {
		return Policy{}, err
	}
	plannerAllowed := map[Action]struct{}{
		ActionRetrySameUnit: {}, ActionRetryAfterBackoff: {}, ActionRunRecoveryPlan: {},
		ActionAwaitDecision: {}, ActionBlock: {},
	}
	silentAllowed := map[Action]struct{}{
		ActionRetryAfterBackoff: {}, ActionRunRecoveryPlan: {}, ActionAwaitDecision: {}, ActionBlock: {},
	}
	planner := func(allowed map[Action]struct{}) Policy {
		return Policy{PlannerEligible: true, ExhaustedAction: ActionAwaitDecision, allowed: allowed}
	}
	switch failure.Class {
	case FailureDeadWorker, FailureArtifactMissing, FailureArtifactInvalid, FailureInterrupted:
		if !failure.Reversible {
			return directPolicy(ActionBlock), nil
		}
		return planner(plannerAllowed), nil
	case FailureSilentTool:
		if !failure.Reversible {
			return directPolicy(ActionBlock), nil
		}
		return planner(silentAllowed), nil
	case FailureValidationFailed:
		if failure.Reversible {
			return planner(plannerAllowed), nil
		}
		return directPolicy(ActionBlock), nil
	case FailureGitHubPublishUncertain, FailureOutboxUncertain:
		return directPolicy(ActionBlock), nil
	case FailureRetryExhausted, FailureHumanRequired, FailureUnknown:
		return directPolicy(ActionAwaitDecision), nil
	case FailureRuntimeContractMismatch, FailureModelMismatch, FailureThinkingMismatch, FailureStaleHead,
		FailureDirtyTree, FailureWriteScopeBreach, FailureRatificationFailed:
		return directPolicy(ActionBlock), nil
	default:
		return Policy{}, fmt.Errorf("no recovery policy for class %q", failure.Class)
	}
}

func directPolicy(action Action) Policy {
	return Policy{DirectAction: action, ExhaustedAction: action, allowed: map[Action]struct{}{action: {}}}
}

func (p Policy) Allows(action Action) bool {
	_, ok := p.allowed[action]
	return ok
}

func (p Policy) ValidateAction(action Action) error {
	if _, err := ParseAction(string(action)); err != nil {
		return err
	}
	if !p.Allows(action) {
		return fmt.Errorf("recovery action %q is forbidden for this failure class", action)
	}
	return nil
}

func ValidatePlan(action Action, steps []PlanStep) error {
	if len(steps) > 4 {
		return errors.New("recovery plan exceeds its step bound")
	}
	seen := make(map[Primitive]struct{}, len(steps))
	for _, step := range steps {
		if _, ok := primitives[step.Primitive]; !ok {
			return fmt.Errorf("unknown recovery plan primitive %q", step.Primitive)
		}
		if _, duplicate := seen[step.Primitive]; duplicate {
			return fmt.Errorf("duplicate recovery plan primitive %q", step.Primitive)
		}
		seen[step.Primitive] = struct{}{}
	}
	switch action {
	case ActionRetrySameUnit, ActionRetryAfterBackoff:
		if len(steps) > 1 || (len(steps) == 1 && steps[0].Primitive != PrimitiveRetryFreshAttempt) {
			return errors.New("retry action permits only the fresh-attempt primitive")
		}
	case ActionRunRecoveryPlan:
		if len(steps) == 0 || steps[len(steps)-1].Primitive != PrimitiveRetryFreshAttempt {
			return errors.New("typed recovery plan must end with a fresh attempt")
		}
	case ActionAwaitDecision, ActionBlock, ActionFinalHumanGate:
		if len(steps) != 0 {
			return errors.New("terminal recovery action cannot carry plan steps")
		}
	default:
		return fmt.Errorf("unknown recovery action %q", action)
	}
	return nil
}

func Backoff(base, maximum time.Duration, attempt int64) (time.Duration, error) {
	if base < 0 || maximum < base || maximum <= 0 || attempt <= 0 {
		return 0, errors.New("invalid recovery backoff policy")
	}
	backoff := base
	for index := int64(1); index < attempt && backoff < maximum; index++ {
		if backoff > maximum/2 {
			backoff = maximum
			break
		}
		backoff *= 2
	}
	if backoff > maximum {
		backoff = maximum
	}
	return backoff, nil
}
