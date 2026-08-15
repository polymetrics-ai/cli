package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
)

// ExecuteAuthorizedFlowAction prepares one payload-bound execution and then
// dispatches it through the standing authorization. The split methods let
// callers prove the prepared identity without turning it into authority.
func (a *App) ExecuteAuthorizedFlowAction(ctx context.Context, req FlowActionExecutionRequest) (FlowActionExecutionResult, error) {
	prepared, err := a.PrepareAuthorizedFlowAction(ctx, req)
	if err != nil {
		return FlowActionExecutionResult{}, err
	}
	return a.ExecutePreparedFlowAction(ctx, prepared)
}

// PrepareAuthorizedFlowAction validates the standing authorization and exact
// payload, then derives safe payload-bound evidence. It makes no write and
// creates no approval or per-firing grant.
func (a *App) PrepareAuthorizedFlowAction(ctx context.Context, req FlowActionExecutionRequest) (PreparedFlowAction, error) {
	if strings.TrimSpace(req.RunID) == "" {
		return PreparedFlowAction{}, errors.New("flow action run id is required")
	}
	if strings.TrimSpace(req.ManifestDigest) == "" {
		return PreparedFlowAction{}, errors.New("flow action manifest digest is required")
	}
	writer, scope, runtime, err := a.validateAuthorizedFlowAction(ctx, req)
	if err != nil {
		return PreparedFlowAction{}, err
	}

	req = cloneFlowActionExecutionRequest(req)
	mapped := mapReverseRecords(req.Records, req.Mappings)
	writeRequest := connectors.WriteRequest{
		Stream: "records", Table: req.DestinationTable, Action: req.Action, Config: runtime,
	}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, writeRequest, mapped); err != nil {
			return PreparedFlowAction{}, fmt.Errorf("validate flow action destination: %w", err)
		}
	}
	preview := connectors.WritePreview{}
	if scope.ConfirmationPolicy.Kind != "" {
		preview, err = validateAuthorizedDestructivePreview(ctx, writer, writeRequest, mapped)
		if err != nil {
			return PreparedFlowAction{}, err
		}
	}
	scopeIdentity, err := AuthorizationScopeIdentity(scope)
	if err != nil {
		return PreparedFlowAction{}, err
	}
	identity, err := a.preparedFlowActionIdentity(req, mapped, runtime, scopeIdentity, preview)
	if err != nil {
		return PreparedFlowAction{}, err
	}
	return PreparedFlowAction{
		Identity: identity, FiringID: req.RunID,
		request: req, mappedRecords: deepCloneRecords(mapped), preview: preview,
		sealedIdentity: identity, scopeIdentity: scopeIdentity,
	}, nil
}

// ExecutePreparedFlowAction revalidates the live authorization and preview,
// acquires a non-authoritative replay lease immediately before provider write,
// then persists a receipt only after acknowledgement and independent read-back.
func (a *App) ExecutePreparedFlowAction(ctx context.Context, prepared PreparedFlowAction) (FlowActionExecutionResult, error) {
	result := FlowActionExecutionResult{
		RecordsAttempted: len(prepared.mappedRecords), PreparedExecutionIdentity: prepared.Identity, FiringID: prepared.FiringID,
	}
	if ctx == nil {
		return result, errors.New("flow action context is required")
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if a == nil {
		return result, errors.New("flow action approval authority is required")
	}
	if !projectApprovalStringEqual(prepared.Identity, prepared.sealedIdentity) ||
		!projectApprovalStringEqual(prepared.FiringID, prepared.request.RunID) {
		return result, &PreparedExecutionRefusedError{Identity: prepared.Identity, Reason: "prepared_identity_changed"}
	}
	// Serialize the same prepared identity before live state refresh. The marker
	// is replay evidence, not authority: every refusal before possible dispatch
	// releases it, while a possibly-sent write leaves it parked fail-closed.
	lease, err := a.acquirePreparedExecutionLease(prepared.Identity)
	if err != nil {
		return result, err
	}
	dispatchMayHaveOccurred := false
	defer func() {
		if !dispatchMayHaveOccurred {
			_ = lease.Release()
		}
	}()

	writer, scope, runtime, err := a.validateAuthorizedFlowAction(ctx, prepared.request)
	if err != nil {
		return result, err
	}
	scopeIdentity, err := AuthorizationScopeIdentity(scope)
	if err != nil {
		return result, err
	}
	if !projectApprovalStringEqual(scopeIdentity, prepared.scopeIdentity) {
		return result, &PreparedExecutionRefusedError{Identity: prepared.Identity, Reason: "authorization_scope_changed"}
	}

	mapped := deepCloneRecords(prepared.mappedRecords)
	writeRequest := connectors.WriteRequest{
		Stream: "records", Table: prepared.request.DestinationTable, Action: prepared.request.Action, Config: runtime,
	}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, writeRequest, mapped); err != nil {
			return result, fmt.Errorf("validate prepared flow action destination: %w", err)
		}
	}
	if scope.ConfirmationPolicy.Kind != "" {
		preview, err := validateAuthorizedDestructivePreview(ctx, writer, writeRequest, mapped)
		if err != nil {
			return result, err
		}
		if !projectApprovalStringEqual(preview.Digest, prepared.preview.Digest) || !sameProjectApprovalTarget(preview.ApprovalTarget, prepared.preview.ApprovalTarget) {
			return result, &PreparedExecutionRefusedError{Identity: prepared.Identity, Reason: "preview_changed"}
		}
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if scope.ConfirmationPolicy.Kind != "" {
		evidence, err := durableAuthorizationEvidence(scope)
		if err != nil {
			return result, err
		}
		writeRequest.Approval = evidence
	}

	dispatchMayHaveOccurred = true
	writeResult, err := writer.Write(ctx, writeRequest, mapped)
	result.RecordsSucceeded = writeResult.RecordsWritten
	result.RecordsFailed = writeResult.RecordsFailed
	if err != nil {
		return result, err
	}
	if writeResult.RecordsFailed != 0 || writeResult.RecordsWritten != len(mapped) {
		return result, fmt.Errorf("flow action connector acknowledgement is incomplete: wrote %d of %d records", writeResult.RecordsWritten, len(mapped))
	}
	acknowledgedAt := time.Now().UTC()

	readBackRecords, err := readFlowActionTarget(ctx, writer, prepared.request.ReadBackStream, runtime, writeResult.RecordsWritten)
	if err != nil {
		return result, err
	}
	if len(readBackRecords) < writeResult.RecordsWritten {
		return result, fmt.Errorf("flow action read-back returned %d records after %d acknowledged writes", len(readBackRecords), writeResult.RecordsWritten)
	}

	receipt, err := a.recordFlowActionReceipt(FlowActionReceipt{
		RunID: prepared.FiringID, FiringID: prepared.FiringID, PreparedExecutionIdentity: prepared.Identity,
		FlowName: prepared.request.FlowName, StepID: prepared.request.StepID,
		AuthorizationReference: prepared.request.AuthorizationReference,
		DestinationConnector:   writer.Name(), Action: prepared.request.Action,
		AcknowledgedAt: acknowledgedAt, ReadBackAt: time.Now().UTC(),
	})
	if err != nil {
		return result, err
	}
	result.ReceiptID = receipt.ID
	return result, nil
}

func (a *App) preparedFlowActionIdentity(req FlowActionExecutionRequest, mapped []connectors.Record, runtime connectors.RuntimeConfig, scopeIdentity string, preview connectors.WritePreview) (string, error) {
	binding := struct {
		AuthorizationReference string                         `json:"authorization_reference"`
		ScopeIdentity          string                         `json:"scope_identity"`
		FiringID               string                         `json:"firing_id"`
		ManifestDigest         string                         `json:"manifest_digest"`
		FlowName               string                         `json:"flow_name"`
		StepID                 string                         `json:"step_id"`
		SourceTable            string                         `json:"source_table"`
		SourceConnection       string                         `json:"source_connection"`
		DestinationTable       string                         `json:"destination_table"`
		DestinationConnector   string                         `json:"destination_connector"`
		Action                 string                         `json:"action"`
		CredentialRevision     string                         `json:"credential_revision"`
		ConfigurationDigest    string                         `json:"configuration_digest"`
		Mappings               map[string]string              `json:"mappings"`
		Records                []connectors.Record            `json:"records"`
		PreviewDigest          string                         `json:"preview_digest"`
		ApprovalTarget         connectors.WriteApprovalTarget `json:"approval_target"`
	}{
		AuthorizationReference: req.AuthorizationReference, ScopeIdentity: scopeIdentity,
		FiringID: req.RunID, ManifestDigest: req.ManifestDigest, FlowName: req.FlowName, StepID: req.StepID,
		SourceTable: req.SourceTable, SourceConnection: req.SourceConnection,
		DestinationTable: req.DestinationTable, DestinationConnector: req.DestinationConnector, Action: req.Action,
		CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
		Mappings: cloneStringMap(req.Mappings), Records: deepCloneRecords(mapped),
		PreviewDigest: preview.Digest, ApprovalTarget: preview.ApprovalTarget,
	}
	payload, err := json.Marshal(binding)
	if err != nil {
		return "", fmt.Errorf("encode prepared flow action identity: %w", err)
	}
	return "pex_" + hashString("prepared-flow-action-v1\x00"+string(payload)), nil
}

func cloneFlowActionExecutionRequest(req FlowActionExecutionRequest) FlowActionExecutionRequest {
	clone := req
	clone.DestinationConfig = cloneStringMap(req.DestinationConfig)
	clone.Mappings = cloneStringMap(req.Mappings)
	clone.Records = deepCloneRecords(req.Records)
	return clone
}

func deepCloneRecords(records []connectors.Record) []connectors.Record {
	cloned := make([]connectors.Record, len(records))
	for i, record := range records {
		cloned[i] = deepCloneRecord(record)
	}
	return cloned
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
	if receipt.RunID == "" || receipt.FiringID == "" || receipt.PreparedExecutionIdentity == "" || receipt.FlowName == "" || receipt.StepID == "" || receipt.AuthorizationReference == "" || receipt.DestinationConnector == "" || receipt.Action == "" || receipt.AcknowledgedAt.IsZero() || receipt.ReadBackAt.IsZero() {
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
