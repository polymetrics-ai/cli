package engine

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// DestructiveTarget is the provider-neutral description consumed by the
// shared pre-dispatch gate. Both writes.json actions and future operations.json
// executors normalize into this shape.
type DestructiveTarget struct {
	Connector     string
	Operation     string
	Method        string
	MutationClass string
	Destructive   bool
	Confirmation  connectors.ConfirmationKind
}

// RequiresApproval reports whether any normalized provider metadata makes the
// operation destructive.
func (t DestructiveTarget) RequiresApproval() bool {
	method := strings.ToUpper(strings.TrimSpace(t.Method))
	mutation := strings.ToLower(strings.TrimSpace(t.MutationClass))
	return method == http.MethodDelete || mutation == "delete" || mutation == "destructive" || t.Destructive || strings.TrimSpace(string(t.Confirmation)) != ""
}

// DestructiveTargetForWrite normalizes a writes.json action for the shared
// approval gate.
func DestructiveTargetForWrite(connector string, action WriteAction) DestructiveTarget {
	confirmation := connectors.ConfirmationKind("")
	if action.Confirmation != nil {
		confirmation = connectors.ConfirmationKind(strings.TrimSpace(string(action.Confirmation.Kind)))
	}
	if confirmation == "" {
		confirmation = connectors.ConfirmationKind(strings.TrimSpace(action.Confirm))
	}
	return DestructiveTarget{
		Connector:     connector,
		Operation:     action.Name,
		Method:        action.Method,
		MutationClass: action.Kind,
		Confirmation:  confirmation,
	}
}

func confirmationKindForWriteAction(action WriteAction) string {
	if DestructiveTargetForWrite("", action).RequiresApproval() {
		return string(connectors.ConfirmationKindDestructive)
	}
	return ""
}

// DestructiveTargetForOperation is the executor seam for operations.json.
// The future rest_write executor can call the same gate without introducing
// executor-specific confirmation logic.
func DestructiveTargetForOperation(connector string, operation OperationSpec) DestructiveTarget {
	confirmation := connectors.ConfirmationKind("")
	if operation.Confirmation != nil {
		confirmation = connectors.ConfirmationKind(strings.TrimSpace(string(operation.Confirmation.Kind)))
	}
	method := ""
	if operation.REST != nil {
		method = operation.REST.Method
	}
	return DestructiveTarget{
		Connector:     connector,
		Operation:     operation.ID,
		Method:        method,
		MutationClass: operation.MutationClass,
		Destructive:   operation.Destructive,
		Confirmation:  confirmation,
	}
}

// GateDestructiveExecution validates typed approval evidence before invoking
// execute. Safe targets pass through, so an executor has one dispatch seam.
func GateDestructiveExecution(ctx context.Context, target DestructiveTarget, evidence *connectors.WriteApprovalEvidence, previewDigest string, execute func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if target.RequiresApproval() {
		if evidence == nil {
			return fmt.Errorf("engine: destructive operation %q requires approval evidence", target.Operation)
		}
		if strings.TrimSpace(evidence.PlanID) == "" || strings.TrimSpace(evidence.PlanHash) == "" {
			return fmt.Errorf("engine: destructive operation %q requires plan-bound approval evidence", target.Operation)
		}
		if strings.TrimSpace(previewDigest) == "" || strings.TrimSpace(evidence.PreviewDigest) == "" || subtle.ConstantTimeCompare([]byte(previewDigest), []byte(evidence.PreviewDigest)) != 1 {
			return fmt.Errorf("engine: destructive operation %q approval does not match its preview", target.Operation)
		}
		if evidence.ApprovedAt.IsZero() {
			return fmt.Errorf("engine: destructive operation %q requires explicit approval", target.Operation)
		}
		if evidence.Confirmation.Kind != connectors.ConfirmationKindDestructive {
			return fmt.Errorf("engine: destructive operation %q requires typed destructive confirmation", target.Operation)
		}
	}
	return execute(ctx)
}
