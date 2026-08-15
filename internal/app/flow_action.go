package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

// ExecuteAuthorizedFlowAction is the connector-backed execution boundary for
// one action step. It validates the durable authorization before any provider
// call, then performs typed validation, connector acknowledgement, independent
// read-back, and finally durable receipt persistence in that order.
func (a *App) ExecuteAuthorizedFlowAction(ctx context.Context, req FlowActionExecutionRequest) (FlowActionExecutionResult, error) {
	if strings.TrimSpace(req.RunID) == "" {
		return FlowActionExecutionResult{}, errors.New("flow action run id is required")
	}
	writer, scope, runtime, err := a.validateAuthorizedFlowAction(ctx, req)
	if err != nil {
		return FlowActionExecutionResult{}, err
	}

	mapped := mapReverseRecords(req.Records, req.Mappings)
	writeRequest := connectors.WriteRequest{
		Stream: "records", Table: req.DestinationTable, Action: req.Action, Config: runtime,
	}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, writeRequest, mapped); err != nil {
			return FlowActionExecutionResult{}, fmt.Errorf("validate flow action destination: %w", err)
		}
	}
	if scope.ConfirmationPolicy.Kind != "" {
		if _, err := validateAuthorizedDestructivePreview(ctx, writer, writeRequest, mapped); err != nil {
			return FlowActionExecutionResult{}, err
		}
		evidence, err := durableAuthorizationEvidence(scope)
		if err != nil {
			return FlowActionExecutionResult{}, err
		}
		writeRequest.Approval = evidence
	}

	writeResult, err := writer.Write(ctx, writeRequest, mapped)
	result := FlowActionExecutionResult{
		RecordsAttempted: len(mapped),
		RecordsSucceeded: writeResult.RecordsWritten,
		RecordsFailed:    writeResult.RecordsFailed,
	}
	if err != nil {
		return result, err
	}
	if writeResult.RecordsFailed != 0 || writeResult.RecordsWritten != len(mapped) {
		return result, fmt.Errorf("flow action connector acknowledgement is incomplete: wrote %d of %d records", writeResult.RecordsWritten, len(mapped))
	}
	acknowledgedAt := time.Now().UTC()

	readBackRecords, err := readFlowActionTarget(ctx, writer, req.ReadBackStream, runtime, writeResult.RecordsWritten)
	if err != nil {
		return result, err
	}
	if len(readBackRecords) < writeResult.RecordsWritten {
		return result, fmt.Errorf("flow action read-back returned %d records after %d acknowledged writes", len(readBackRecords), writeResult.RecordsWritten)
	}

	receipt, err := a.recordFlowActionReceipt(FlowActionReceipt{
		RunID:                  req.RunID,
		FlowName:               req.FlowName,
		StepID:                 req.StepID,
		AuthorizationReference: req.AuthorizationReference,
		DestinationConnector:   writer.Name(),
		Action:                 req.Action,
		AcknowledgedAt:         acknowledgedAt,
		ReadBackAt:             time.Now().UTC(),
	})
	if err != nil {
		return result, err
	}
	result.ReceiptID = receipt.ID
	return result, nil
}

// ValidateAuthorizedFlowAction verifies every non-payload precondition for a
// flow action. It makes no provider request and is used by Engine preflight to
// reject a changed scope before any flow step can dispatch.
func (a *App) ValidateAuthorizedFlowAction(ctx context.Context, req FlowActionExecutionRequest) error {
	_, _, _, err := a.validateAuthorizedFlowAction(ctx, req)
	return err
}

func (a *App) validateAuthorizedFlowAction(ctx context.Context, req FlowActionExecutionRequest) (connectors.Connector, AuthorizationScope, connectors.RuntimeConfig, error) {
	if ctx == nil {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, errors.New("flow action context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, err
	}
	if err := validateFlowActionAuthorizationRequest(req); err != nil {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, err
	}

	writer, credential, runtime, err := a.resolveEndpointWithCredential(ctx, EndpointConfig{
		Connector: req.DestinationConnector, Credential: req.DestinationCredential, Config: req.DestinationConfig,
	})
	if err != nil {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, fmt.Errorf("resolve flow action destination: %w", err)
	}
	if !writer.Metadata().Capabilities.Write {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, fmt.Errorf("connector %q does not support flow action writes", writer.Name())
	}

	expiresAt, err := a.flowAuthorizationExpiry(req.AuthorizationReference)
	if err != nil {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, err
	}
	scope := a.authorizationScopeForFlowAction(req, credential, runtime, expiresAt)
	if _, err := a.requireAuthorization(req.AuthorizationReference, scope, time.Now().UTC()); err != nil {
		return nil, AuthorizationScope{}, connectors.RuntimeConfig{}, err
	}
	return writer, scope, runtime, nil
}

func validateFlowActionAuthorizationRequest(req FlowActionExecutionRequest) error {
	for field, value := range map[string]string{
		"flow action flow name":               req.FlowName,
		"flow action step id":                 req.StepID,
		"flow action source table":            req.SourceTable,
		"flow action source connection":       req.SourceConnection,
		"flow action destination table":       req.DestinationTable,
		"flow action destination connector":   req.DestinationConnector,
		"flow action destination credential":  req.DestinationCredential,
		"flow action":                         req.Action,
		"flow action authorization reference": req.AuthorizationReference,
		"flow action read-back stream":        req.ReadBackStream,
	} {
		if strings.TrimSpace(value) == "" {
			return errors.New(field + " is required")
		}
	}
	if len(req.Mappings) == 0 {
		return errors.New("flow action mappings are required")
	}
	return nil
}

func (a *App) authorizationScopeForFlowAction(req FlowActionExecutionRequest, credential CredentialMeta, runtime connectors.RuntimeConfig, expiresAt time.Time) AuthorizationScope {
	confirmation := a.confirmationPolicyForAction(req.DestinationConnector, req.Action)
	return canonicalAuthorizationScope(AuthorizationScope{
		SourceConnection:              req.SourceConnection,
		DestinationConnection:         credential.ID,
		DestinationCredentialRevision: runtime.CredentialRevision,
		StreamTables: []AuthorizationStreamTable{{
			Stream: "records", SourceTable: req.SourceTable, DestinationTable: req.DestinationTable,
		}},
		FieldMappings:                  cloneStringMap(req.Mappings),
		WriteAction:                    req.Action,
		DestinationConfigurationDigest: runtime.ConfigurationDigest,
		EnabledOperations:              []string{req.Action},
		ConfirmationPolicy:             confirmation,
		// The record's expiry is itself part of the durable scope. It is loaded
		// from that opaque record, never accepted from the manifest or CLI.
		ExpiresAt: expiresAt,
	})
}

func (a *App) flowAuthorizationExpiry(reference string) (time.Time, error) {
	loaded, err := a.store.LoadReadOnly()
	if err != nil {
		return time.Time{}, err
	}
	if err := a.normalizeLoadedState(loaded, false); err != nil {
		return time.Time{}, err
	}
	for _, authorization := range a.state.Authorizations {
		if authorization.Reference == reference {
			return authorization.Scope.ExpiresAt, nil
		}
	}
	return time.Time{}, fmtAuthorizationNotFound(reference)
}

func readFlowActionTarget(ctx context.Context, connector connectors.Connector, stream string, runtime connectors.RuntimeConfig, want int) ([]connectors.Record, error) {
	rows := make([]connectors.Record, 0, want)
	err := connector.Read(ctx, connectors.ReadRequest{Stream: stream, Config: runtime, Limit: want}, connectors.LimitEmitter(want, func(record connectors.Record) error {
		rows = append(rows, cloneRecord(record))
		return nil
	}))
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return nil, err
	}
	return rows, nil
}

func (a *App) recordFlowActionReceipt(receipt FlowActionReceipt) (FlowActionReceipt, error) {
	if receipt.RunID == "" || receipt.FlowName == "" || receipt.StepID == "" || receipt.AuthorizationReference == "" || receipt.DestinationConnector == "" || receipt.Action == "" || receipt.AcknowledgedAt.IsZero() || receipt.ReadBackAt.IsZero() {
		return FlowActionReceipt{}, errors.New("flow action receipt is incomplete")
	}
	id, err := prefixedID("fact")
	if err != nil {
		return FlowActionReceipt{}, err
	}
	receipt.ID = id
	updated, err := a.updateState(func(current state) (state, error) {
		current.FlowActionReceipts = append(current.FlowActionReceipts, receipt)
		return current, nil
	})
	if err != nil {
		return FlowActionReceipt{}, err
	}
	a.state = updated
	return receipt, nil
}

// ListFlowActionReceipts returns safe immutable copies of the durable action
// receipts. It contains no records, credentials, configuration, or approval
// token material.
func (a *App) ListFlowActionReceipts() []FlowActionReceipt {
	out := append([]FlowActionReceipt(nil), a.state.FlowActionReceipts...)
	return out
}
