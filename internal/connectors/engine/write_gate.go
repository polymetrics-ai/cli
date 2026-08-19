package engine

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/transportpolicy"
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
	// Secret operations are required by the bundle validator to declare
	// typed_confirmation. Make that declaration effective at the shared
	// dispatch gate as well: otherwise a generated secret GraphQL mutation
	// could carry a correct redaction policy yet reach the provider without the
	// promised preview/approval/confirmation evidence.
	if confirmation == "" && operation.SensitivePolicy != nil && strings.EqualFold(strings.TrimSpace(operation.SensitivePolicy.ApprovalMode), "typed_confirmation") {
		confirmation = connectors.ConfirmationKindDestructive
	}
	method := ""
	if operation.REST != nil {
		method = operation.REST.Method
	} else if operation.Kind == "graphql_mutation" && operation.GraphQL != nil {
		// A fixed GraphQL mutation is transported as POST, even though its
		// safety class comes from the declared mutation metadata rather than
		// HTTP DELETE. Recording the actual transport makes preview grants bind
		// the request the executor will issue.
		method = http.MethodPost
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
func GateDestructiveExecution(ctx context.Context, target DestructiveTarget, evidence *connectors.WriteApprovalEvidence, previewDigest string, approvalTarget connectors.WriteApprovalTarget, execute func(context.Context) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if target.RequiresApproval() {
		if evidence == nil {
			return fmt.Errorf("engine: destructive operation %q requires approval evidence", target.Operation)
		}
		if strings.TrimSpace(previewDigest) == "" {
			return fmt.Errorf("engine: destructive operation %q requires a prepared preview", target.Operation)
		}
		if err := evidence.Authorize(approvalTarget, previewDigest, time.Now().UTC()); err != nil {
			return fmt.Errorf("engine: destructive operation %q: %w", target.Operation, err)
		}
		ctx = transportpolicy.MarkDestructive(ctx)
	}
	return execute(ctx)
}
