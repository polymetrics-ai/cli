package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	reversePlanModeIssueLabelTransport        = "issue_label_transport"
	reversePlanModeIssueLabelTransportCleanup = "issue_label_transport_cleanup"
	issueLabelTransportBindingDomain          = "issue_label_transport/v1"
	issueLabelTransportTable                  = "transport_issue_label"
)

// issueLabelTransportBinding is deliberately closed over the only
// deterministic destination this walking slice supports. It is not a generic
// action, URL, or record capability: the connection supplies every value and
// App rederives it before a grant can be consumed.
type issueLabelTransportBinding struct {
	Domain              string            `json:"domain"`
	ConnectionID        string            `json:"connection_id"`
	Stream              string            `json:"stream"`
	Mode                synccontract.Mode `json:"mode"`
	ExpectedSourceIssue int               `json:"expected_source_issue,omitempty"`
	Destination         string            `json:"destination"`
	Action              string            `json:"action"`
	TargetIssue         int               `json:"target_issue"`
	Labels              []string          `json:"labels"`
	CredentialRevision  string            `json:"credential_revision"`
	ConfigurationDigest string            `json:"configuration_digest"`
	ApprovalScope       string            `json:"approval_scope"`
	PreviewDigest       string            `json:"preview_digest"`
	ForwardPlanID       string            `json:"forward_plan_id,omitempty"`
}

type issueLabelPreparedWrite struct {
	writer     connectors.Connector
	credential CredentialMeta
	runtime    connectors.RuntimeConfig
	record     connectors.Record
	preview    connectors.WritePreview
}

// PlanIssueLabelTransport creates the forward half of the explicit PM
// plan -> preview -> approval -> execute lifecycle before a transport run.
// It prepares only the fixed destination configured on the named connection;
// connection; it does not read, stage, reopen, or mutate a provider resource.
func (a *App) PlanIssueLabelTransport(ctx context.Context, connectionID string) (ReversePlan, error) {
	conn, err := a.issueLabelTransportConnection(connectionID)
	if err != nil {
		return ReversePlan{}, err
	}
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return ReversePlan{}, err
	}
	syncMode, err := issueLabelTransportMode(conn)
	if err != nil {
		return ReversePlan{}, err
	}
	action, err := contract.actionForSyncMode(syncMode)
	if err != nil {
		return ReversePlan{}, err
	}
	return a.planIssueLabelTransport(ctx, conn, syncMode, reversePlanModeIssueLabelTransport, action, conn.Destination, "")
}

// PlanIssueLabelTransportCleanup derives the only permitted inverse from
// a completed forward plan. Callers cannot select another action, issue, label,
// URL, or reuse the forward plan's approval token.
func (a *App) PlanIssueLabelTransportCleanup(ctx context.Context, connectionID, forwardPlanID string) (ReversePlan, error) {
	conn, err := a.issueLabelTransportConnection(connectionID)
	if err != nil {
		return ReversePlan{}, err
	}
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return ReversePlan{}, err
	}
	syncMode, err := issueLabelTransportMode(conn)
	if err != nil {
		return ReversePlan{}, err
	}
	if syncMode != synccontract.ModeFullAppend {
		return ReversePlan{}, fmt.Errorf("closed issue-label transport cleanup is available only for issues/full_append")
	}
	forward, endpoint, err := a.authenticatedIssueLabelTransportForwardPlan(ctx, conn, forwardPlanID)
	if err != nil {
		return ReversePlan{}, err
	}
	return a.planIssueLabelTransport(ctx, conn, syncMode, reversePlanModeIssueLabelTransportCleanup, contract.cleanup, endpoint, forward.ID)
}

func (a *App) planIssueLabelTransport(ctx context.Context, conn Connection, syncMode synccontract.Mode, mode string, action issueLabelTransportAction, endpoint EndpointConfig, forwardPlanID string) (ReversePlan, error) {
	if a == nil || a.approval == nil {
		return ReversePlan{}, fmt.Errorf("closed issue-label transport approval requires an app approval authority")
	}
	prepared, err := a.prepareIssueLabelTransportWrite(ctx, conn, endpoint, action)
	if err != nil {
		return ReversePlan{}, err
	}
	binding, err := a.issueLabelTransportBinding(conn, syncMode, mode, action, prepared.runtime, prepared.preview, prepared.runtime.Config, forwardPlanID)
	if err != nil {
		return ReversePlan{}, err
	}
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return ReversePlan{}, fmt.Errorf("hash closed issue-label transport destination binding: %w", err)
	}
	planHash, err := issueLabelTransportPlanHash(mode, bindingSHA256)
	if err != nil {
		return ReversePlan{}, err
	}
	confirmation := a.confirmationPolicyForAction(conn.Destination.Connector, action.name)
	if confirmation.Kind != connectors.ConfirmationKindDestructive {
		return ReversePlan{}, fmt.Errorf("closed issue-label transport action %q must require destructive confirmation", action.name)
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, err
	}
	seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
		PlanID:              id,
		PlanHash:            planHash,
		Mode:                mode,
		Connector:           conn.Destination.Connector,
		Operation:           action.name,
		CredentialRevision:  prepared.runtime.CredentialRevision,
		ConfigurationDigest: prepared.runtime.ConfigurationDigest,
		Batchable:           a.actionIsBatchable(conn.Destination.Connector, action.name),
		Scope:               prepared.runtime.WriteApprovalScope,
		Confirmation:        confirmation,
	})
	if err != nil {
		return ReversePlan{}, err
	}
	plan := ReversePlan{
		ID:                     id,
		Name:                   issueLabelTransportPlanName(mode),
		Status:                 "planned",
		Mode:                   mode,
		SourceConnection:       conn.Name,
		DestinationConnector:   conn.Destination.Connector,
		DestinationCredential:  endpoint.Credential,
		DestinationConfig:      cloneStringMap(endpoint.Config),
		Action:                 action.name,
		Mappings:               map[string]string{},
		ConfirmationChallenge:  string(confirmation.Kind),
		ConfirmationPolicy:     confirmation,
		RecordCount:            1,
		Sample:                 []connectors.Record{cloneRecord(prepared.record)},
		PlanHash:               planHash,
		PlanSeal:               &seal,
		TransportConnectionID:  conn.ID,
		TransportBindingSHA256: bindingSHA256,
		TransportForwardPlanID: forwardPlanID,
		CreatedAt:              seal.IssuedAt,
		ExpiresAt:              seal.ExpiresAt,
	}
	stored := plan
	stored.ApprovalToken = ""
	a.state.ReversePlans = append(a.state.ReversePlans, stored)
	if err := a.save(); err != nil {
		return ReversePlan{}, err
	}
	return plan, nil
}

// PreviewIssueLabelTransport re-prepares the exact fixed record and
// persists the existing project-scoped destructive grant. It is an explicit
// pre-run operation; Transport never calls it and never mints a token.
func (a *App) PreviewIssueLabelTransport(ctx context.Context, planID string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(planID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if !isIssueLabelTransportMode(plan.Mode) {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("reverse plan %q is not a closed issue-label transport plan", plan.ID)
	}
	if err := approvalConsumptionUncertainError(plan, nil); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	conn, err := a.issueLabelTransportConnection(plan.TransportConnectionID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	syncMode, err := issueLabelTransportMode(conn)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	var action issueLabelTransportAction
	if plan.Mode == reversePlanModeIssueLabelTransportCleanup {
		if syncMode != synccontract.ModeFullAppend {
			return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("closed issue-label transport cleanup is available only for issues/full_append")
		}
		action = contract.cleanup
	} else {
		action, err = contract.actionForSyncMode(syncMode)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
	}
	if err := a.validateIssueLabelTransportPlan(plan, conn, plan.Mode, action, plan.TransportForwardPlanID); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	endpoint := EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: cloneStringMap(plan.DestinationConfig)}
	if plan.Mode == reversePlanModeIssueLabelTransportCleanup {
		_, authenticatedEndpoint, err := a.authenticatedIssueLabelTransportForwardPlan(ctx, conn, plan.TransportForwardPlanID)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
		if !issueLabelTransportEndpointEqual(endpoint, authenticatedEndpoint) {
			return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("closed issue-label cleanup plan no longer matches its authenticated forward destination")
		}
		endpoint = authenticatedEndpoint
	}
	prepared, err := a.prepareIssueLabelTransportWrite(ctx, conn, endpoint, action)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	binding, err := a.issueLabelTransportBinding(conn, syncMode, plan.Mode, action, prepared.runtime, prepared.preview, prepared.runtime.Config, plan.TransportForwardPlanID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := validateIssueLabelTransportBinding(plan, binding); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	issued, err := a.persistDestructivePreview(plan, prepared.preview)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	return issued, prepared.preview, nil
}

// ApplyIssueLabelTransport validates the durable receipt and reopened
// singleton against the pre-run binding before consuming the grant and handing
// opaque approval evidence to the declarative engine immediately before I/O.
func (a *App) ApplyIssueLabelTransport(ctx context.Context, connectionID string, approval synctransport.DestinationApproval, runtime connectors.RuntimeConfig, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset) (connectors.WriteResult, error) {
	if strings.TrimSpace(approval.PlanID) == "" {
		return connectors.WriteResult{}, fmt.Errorf("closed issue-label transport requires a pre-run plan")
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	conn, err := a.issueLabelTransportConnection(connectionID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	syncMode, err := issueLabelTransportMode(conn)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	action, err := contract.actionForSyncMode(syncMode)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := a.validateIssueLabelTransportPlan(plan, conn, reversePlanModeIssueLabelTransport, action, ""); err != nil {
		return connectors.WriteResult{}, err
	}
	if err := validateIssueLabelTransportWorkset(conn, syncMode, receipt, workset); err != nil {
		return connectors.WriteResult{}, err
	}
	prepared, err := a.prepareIssueLabelTransportWrite(ctx, conn, conn.Destination, action)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if !issueLabelTransportRuntimeEqual(runtime, prepared.runtime) {
		return connectors.WriteResult{}, fmt.Errorf("closed issue-label transport destination runtime does not match the connection-owned approval identity")
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return connectors.WriteResult{}, err
	}
	binding, err := a.issueLabelTransportBinding(conn, syncMode, plan.Mode, action, prepared.runtime, prepared.preview, prepared.runtime.Config, "")
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := validateIssueLabelTransportBinding(plan, binding); err != nil {
		return connectors.WriteResult{}, err
	}
	scope := a.issueLabelTransportAuthorizationScope(conn, syncMode, action, plan, prepared)
	var evidence *connectors.WriteApprovalEvidence
	if plan.AuthorizationReference != "" {
		if strings.TrimSpace(approval.ApprovalToken) != "" {
			return connectors.WriteResult{}, &AuthorizationTokenReplayError{Reference: plan.AuthorizationReference}
		}
		if _, err := a.requireAuthorization(plan.AuthorizationReference, scope, time.Now().UTC()); err != nil {
			return connectors.WriteResult{}, err
		}
		evidence, err = durableAuthorizationEvidence(scope)
	} else {
		if err := validateIssueLabelTransportApproval(approval); err != nil {
			return connectors.WriteResult{}, err
		}
		authorization, err := newAuthorizationRecord(scope, time.Now().UTC())
		if err != nil {
			return connectors.WriteResult{}, err
		}
		evidence, _, err = a.consumePlanApproval(plan, RunReverseETLRequest{
			PlanID:        approval.PlanID,
			ApprovalToken: approval.ApprovalToken,
			Confirmation:  approval.Confirmation,
		}, prepared.preview, &authorization)
	}
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if evidence == nil {
		return connectors.WriteResult{}, fmt.Errorf("closed issue-label transport approval did not yield destructive write evidence")
	}
	result, err := prepared.writer.Write(ctx, connectors.WriteRequest{
		Stream:   "issues",
		Table:    issueLabelTransportTable,
		Action:   action.name,
		Config:   prepared.runtime,
		Approval: evidence,
	}, []connectors.Record{prepared.record})
	if err != nil {
		return connectors.WriteResult{}, fmt.Errorf("execute approved issue-label transport: %w", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		return connectors.WriteResult{}, fmt.Errorf("approved issue-label result written=%d failed=%d, want one durable write", result.RecordsWritten, result.RecordsFailed)
	}
	// The original plan moves from the consumed approval state to executed only
	// once. A durable authorization intentionally permits later identical-scope
	// runs without re-consuming the one-time token or mutating that plan state.
	if plan.AuthorizationReference == "" {
		if err := a.markIssueLabelTransportPlanExecuted(plan.ID); err != nil {
			return connectors.WriteResult{}, err
		}
	}
	return result, nil
}

// ApplyIssueLabelTransportCleanup executes only the inverse derived from
// its closed forward plan. It has a distinct preview, token, grant, and
// evidence; the forward approval is not accepted here.
func (a *App) ApplyIssueLabelTransportCleanup(ctx context.Context, connectionID string, approval synctransport.DestinationApproval) (connectors.WriteResult, error) {
	if err := validateIssueLabelTransportApproval(approval); err != nil {
		return connectors.WriteResult{}, err
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	conn, err := a.issueLabelTransportConnection(connectionID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	syncMode, err := issueLabelTransportMode(conn)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if syncMode != synccontract.ModeFullAppend {
		return connectors.WriteResult{}, fmt.Errorf("closed issue-label transport cleanup is available only for issues/full_append")
	}
	if err := a.validateIssueLabelTransportPlan(plan, conn, reversePlanModeIssueLabelTransportCleanup, contract.cleanup, plan.TransportForwardPlanID); err != nil {
		return connectors.WriteResult{}, err
	}
	forward, endpoint, err := a.authenticatedIssueLabelTransportForwardPlan(ctx, conn, plan.TransportForwardPlanID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if !issueLabelTransportEndpointEqual(EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: cloneStringMap(plan.DestinationConfig)}, endpoint) {
		return connectors.WriteResult{}, fmt.Errorf("closed issue-label cleanup plan no longer matches its authenticated forward destination")
	}
	prepared, err := a.prepareIssueLabelTransportWrite(ctx, conn, endpoint, contract.cleanup)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return connectors.WriteResult{}, err
	}
	binding, err := a.issueLabelTransportBinding(conn, syncMode, plan.Mode, contract.cleanup, prepared.runtime, prepared.preview, prepared.runtime.Config, forward.ID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := validateIssueLabelTransportBinding(plan, binding); err != nil {
		return connectors.WriteResult{}, err
	}
	evidence, _, err := a.consumePlanApproval(plan, RunReverseETLRequest{
		PlanID:        approval.PlanID,
		ApprovalToken: approval.ApprovalToken,
		Confirmation:  approval.Confirmation,
	}, prepared.preview, nil)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if evidence == nil {
		return connectors.WriteResult{}, fmt.Errorf("closed issue-label cleanup approval did not yield destructive write evidence")
	}
	result, err := prepared.writer.Write(ctx, connectors.WriteRequest{
		Stream:   "issues",
		Table:    issueLabelTransportTable,
		Action:   contract.cleanup.name,
		Config:   prepared.runtime,
		Approval: evidence,
	}, []connectors.Record{prepared.record})
	if err != nil {
		return connectors.WriteResult{}, fmt.Errorf("execute approved issue-label transport cleanup: %w", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		return connectors.WriteResult{}, fmt.Errorf("approved issue-label cleanup result written=%d failed=%d, want one durable write", result.RecordsWritten, result.RecordsFailed)
	}
	if err := a.markIssueLabelTransportPlanExecuted(plan.ID); err != nil {
		return connectors.WriteResult{}, err
	}
	return result, nil
}

func (a *App) prepareIssueLabelTransportWrite(ctx context.Context, conn Connection, endpoint EndpointConfig, action issueLabelTransportAction) (issueLabelPreparedWrite, error) {
	if endpoint.Connector != conn.Destination.Connector {
		return issueLabelPreparedWrite{}, fmt.Errorf("closed issue-label transport approval requires the connection-owned destination connector")
	}
	writer, credential, runtime, err := a.resolveEndpointWithCredential(ctx, endpoint)
	if err != nil {
		return issueLabelPreparedWrite{}, fmt.Errorf("resolve closed issue-label transport destination: %w", err)
	}
	if err := issueLabelTransportRepositoryConfig(runtime.Config); err != nil {
		return issueLabelPreparedWrite{}, err
	}
	targetIssue, err := issueLabelTransportIssueNumber(runtime.Config, issueLabelTransportTargetIssueConfig)
	if err != nil {
		return issueLabelPreparedWrite{}, err
	}
	label, err := issueLabelTransportLabel(runtime.Config)
	if err != nil {
		return issueLabelPreparedWrite{}, err
	}
	if writer.Name() != conn.Destination.Connector || !writer.Metadata().Capabilities.Write {
		return issueLabelPreparedWrite{}, fmt.Errorf("closed issue-label transport destination does not expose typed writes")
	}
	record, err := action.record(targetIssue, label)
	if err != nil {
		return issueLabelPreparedWrite{}, err
	}
	request := connectors.WriteRequest{Stream: "issues", Table: issueLabelTransportTable, Action: action.name, Config: runtime}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, request, []connectors.Record{record}); err != nil {
			return issueLabelPreparedWrite{}, fmt.Errorf("validate closed issue-label transport write: %w", err)
		}
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return issueLabelPreparedWrite{}, fmt.Errorf("closed issue-label transport destination does not support typed write previews")
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, []connectors.Record{record})
	if err != nil {
		return issueLabelPreparedWrite{}, fmt.Errorf("prepare closed issue-label transport write: %w", err)
	}
	if strings.TrimSpace(preview.Digest) == "" || preview.ApprovalTarget.Connector != conn.Destination.Connector || preview.ApprovalTarget.Operation != action.name || preview.ApprovalTarget.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		return issueLabelPreparedWrite{}, fmt.Errorf("closed issue-label transport write preview does not bind a destructive %s target", action.name)
	}
	return issueLabelPreparedWrite{writer: writer, credential: credential, runtime: runtime, record: record, preview: preview}, nil
}

func (a *App) issueLabelTransportBinding(conn Connection, syncMode synccontract.Mode, mode string, action issueLabelTransportAction, runtime connectors.RuntimeConfig, preview connectors.WritePreview, config map[string]string, forwardPlanID string) (issueLabelTransportBinding, error) {
	targetIssue, err := issueLabelTransportIssueNumber(config, issueLabelTransportTargetIssueConfig)
	if err != nil {
		return issueLabelTransportBinding{}, err
	}
	label, err := issueLabelTransportLabel(config)
	if err != nil {
		return issueLabelTransportBinding{}, err
	}
	binding := issueLabelTransportBinding{
		Domain:              issueLabelTransportBindingDomain,
		ConnectionID:        conn.ID,
		Stream:              "issues",
		Mode:                syncMode,
		Destination:         conn.Destination.Connector,
		Action:              action.name,
		TargetIssue:         targetIssue,
		Labels:              []string{label},
		CredentialRevision:  runtime.CredentialRevision,
		ConfigurationDigest: runtime.ConfigurationDigest,
		ApprovalScope:       runtime.WriteApprovalScope,
		PreviewDigest:       preview.Digest,
		ForwardPlanID:       forwardPlanID,
	}
	if mode == reversePlanModeIssueLabelTransport {
		sourceIssue, err := issueLabelTransportIssueNumber(conn.Source.Config, issueLabelTransportSourceIssueConfig)
		if err != nil {
			return issueLabelTransportBinding{}, err
		}
		binding.ExpectedSourceIssue = sourceIssue
	}
	return binding, nil
}

func issueLabelTransportPlanHash(mode, bindingSHA256 string) (string, error) {
	if !isIssueLabelTransportMode(mode) || strings.TrimSpace(bindingSHA256) == "" {
		return "", fmt.Errorf("closed issue-label transport plan requires a known mode and destination binding")
	}
	return hashJSON(struct {
		Domain              string `json:"domain"`
		Mode                string `json:"mode"`
		TransportBindingSHA string `json:"transport_binding_sha256"`
	}{
		Domain:              issueLabelTransportBindingDomain,
		Mode:                mode,
		TransportBindingSHA: bindingSHA256,
	})
}

func validateIssueLabelTransportBinding(plan ReversePlan, binding issueLabelTransportBinding) error {
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return fmt.Errorf("hash closed issue-label transport destination binding: %w", err)
	}
	if !constantTimeStringEqual(plan.TransportBindingSHA256, bindingSHA256) {
		return fmt.Errorf("closed issue-label transport approval binding does not match the reopened connection workset or destination identity")
	}
	planHash, err := issueLabelTransportPlanHash(plan.Mode, bindingSHA256)
	if err != nil {
		return err
	}
	if !constantTimeStringEqual(plan.PlanHash, planHash) {
		return fmt.Errorf("closed issue-label transport approval plan hash does not match its destination binding")
	}
	return nil
}

func (a *App) issueLabelTransportConnection(connectionID string) (Connection, error) {
	if a == nil {
		return Connection{}, fmt.Errorf("closed issue-label transport approval requires an app")
	}
	if strings.TrimSpace(connectionID) == "" {
		return Connection{}, fmt.Errorf("closed issue-label transport approval requires a connection ID")
	}
	conn, ok := a.findConnectionByID(connectionID)
	if !ok {
		return Connection{}, fmt.Errorf("closed issue-label transport connection %q was not found", connectionID)
	}
	if conn.Source.Connector == "" || conn.Source.Connector != conn.Destination.Connector {
		return Connection{}, fmt.Errorf("closed issue-label transport connection %q must use one definition-owned connector as source and destination", conn.ID)
	}
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return Connection{}, fmt.Errorf("closed issue-label transport connection %q does not select the exact definition-owned issue-label contract", conn.ID)
	}
	mode, err := issueLabelTransportMode(conn)
	if err != nil {
		return Connection{}, err
	}
	if _, err := contract.actionForSyncMode(mode); err != nil {
		return Connection{}, err
	}
	if consent, required := issueLabelTransportConsentConfigForMode(mode); required && !strings.EqualFold(strings.TrimSpace(conn.Destination.Config[consent]), "true") {
		return Connection{}, fmt.Errorf("closed issue-label transport connection %q requires explicit destination config %s=true for sync mode %q", conn.ID, consent, mode)
	}
	if _, err := issueLabelTransportIssueNumber(conn.Source.Config, issueLabelTransportSourceIssueConfig); err != nil {
		return Connection{}, err
	}
	if _, err := issueLabelTransportIssueNumber(conn.Destination.Config, issueLabelTransportTargetIssueConfig); err != nil {
		return Connection{}, err
	}
	if _, err := issueLabelTransportLabel(conn.Destination.Config); err != nil {
		return Connection{}, err
	}
	return conn, nil
}

func issueLabelTransportMode(conn Connection) (synccontract.Mode, error) {
	stream, ok := conn.Streams["issues"]
	if !ok {
		return "", fmt.Errorf("closed issue-label transport connection %q has no issues stream", conn.ID)
	}
	mode, err := ParseStreamSyncMode(stream)
	if err != nil {
		return "", err
	}
	if mode.ContractMode == "" {
		return "", fmt.Errorf("closed issue-label transport connection %q has no contract sync mode", conn.ID)
	}
	return mode.ContractMode, nil
}

func issueLabelTransportConsentConfigForMode(mode synccontract.Mode) (string, bool) {
	switch mode {
	case synccontract.ModeFullOverwrite:
		return issueLabelTransportSetReplaceConsentConfig, true
	case synccontract.ModeIncrementalUpsert:
		return issueLabelTransportKeyedConsentConfig, true
	default:
		return "", false
	}
}

func (a *App) issueLabelTransportContract(conn Connection) (issueLabelTransportContract, error) {
	if a == nil || a.registry == nil {
		return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport registry is unavailable")
	}
	registered, ok := a.registry.Get(conn.Source.Connector)
	if !ok {
		return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport connector %q is unavailable", conn.Source.Connector)
	}
	definition, ok := connectors.DefinitionOf(registered)
	if !ok {
		return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport connector %q has no definition", conn.Source.Connector)
	}
	return issueLabelTransportContractForDefinition(definition)
}

func (a *App) validateIssueLabelTransportPlan(plan ReversePlan, conn Connection, mode string, action issueLabelTransportAction, forwardPlanID string) error {
	if plan.Mode != mode || plan.TransportConnectionID != conn.ID || plan.SourceConnection != conn.Name || plan.DestinationConnector != conn.Destination.Connector || plan.Action != action.name || plan.RecordCount != 1 || plan.ConnectorCommandOperation != "" || len(plan.Mappings) != 0 || plan.TransportForwardPlanID != forwardPlanID || strings.TrimSpace(plan.TransportBindingSHA256) == "" {
		return fmt.Errorf("closed issue-label transport approval plan does not bind the exact connection-owned %s mutation", action.name)
	}
	if plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive || a.confirmationPolicyForPlan(plan).Kind != connectors.ConfirmationKindDestructive {
		return fmt.Errorf("closed issue-label transport approval plan does not require destructive confirmation")
	}
	return nil
}

func (a *App) authenticatedIssueLabelTransportForwardPlan(ctx context.Context, conn Connection, id string) (ReversePlan, EndpointConfig, error) {
	contract, err := a.issueLabelTransportContract(conn)
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	forward, err := a.GetReversePlan(id)
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	if forward.Mode != reversePlanModeIssueLabelTransport || forward.TransportConnectionID != conn.ID || forward.Action != contract.apply.name || strings.TrimSpace(forward.TransportBindingSHA256) == "" || forward.Status != "executed" {
		return ReversePlan{}, EndpointConfig{}, fmt.Errorf("closed issue-label cleanup requires an executed forward plan for connection %q", conn.ID)
	}
	if err := a.validateIssueLabelTransportPlan(forward, conn, reversePlanModeIssueLabelTransport, contract.apply, ""); err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	endpoint := EndpointConfig{
		Connector:  forward.DestinationConnector,
		Credential: forward.DestinationCredential,
		Config:     cloneStringMap(forward.DestinationConfig),
	}
	prepared, err := a.prepareIssueLabelTransportWrite(ctx, conn, endpoint, contract.apply)
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	if err := a.verifyPlanSealForRuntime(forward, prepared.runtime); err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	binding, err := a.issueLabelTransportBinding(conn, synccontract.ModeFullAppend, forward.Mode, contract.apply, prepared.runtime, prepared.preview, prepared.runtime.Config, "")
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	if err := validateIssueLabelTransportBinding(forward, binding); err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	return forward, endpoint, nil
}

func validateIssueLabelTransportWorkset(conn Connection, mode synccontract.Mode, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset) error {
	if err := receipt.Validate(); err != nil {
		return fmt.Errorf("closed issue-label transport receipt: %w", err)
	}
	if receipt.Owner != conn.ID || receipt.ID != workset.ID || receipt.Stream != "issues" || receipt.Mode != mode || receipt.Records != len(workset.Records) || receipt.Tombstones != len(workset.Tombstones) {
		return fmt.Errorf("closed issue-label transport receipt does not bind the reopened workset")
	}
	if len(workset.Records) != 1 {
		return fmt.Errorf("closed issue-label transport requires exactly one reopened source issue")
	}
	expectedSource, err := issueLabelTransportIssueNumber(conn.Source.Config, issueLabelTransportSourceIssueConfig)
	if err != nil {
		return err
	}
	actualSource, err := issueNumberFromRecord(workset.Records[0])
	if err != nil {
		return fmt.Errorf("closed issue-label transport reopened source issue: %w", err)
	}
	if actualSource != expectedSource {
		return fmt.Errorf("closed issue-label transport reopened source issue %d does not match configured source issue %d", actualSource, expectedSource)
	}
	return nil
}

func validateIssueLabelTransportApproval(approval synctransport.DestinationApproval) error {
	if strings.TrimSpace(approval.PlanID) == "" || strings.TrimSpace(approval.ApprovalToken) == "" || approval.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		return fmt.Errorf("closed issue-label transport requires a pre-run plan-preview-approved destructive grant")
	}
	return nil
}

func (a *App) issueLabelTransportAuthorizationScope(conn Connection, mode synccontract.Mode, action issueLabelTransportAction, plan ReversePlan, prepared issueLabelPreparedWrite) AuthorizationScope {
	fieldMappings := make(map[string]string, len(action.binding.Inputs))
	for _, input := range action.binding.Inputs {
		fieldMappings[input.Input] = input.Field
	}
	return canonicalAuthorizationScope(AuthorizationScope{
		SourceConnection:              conn.ID,
		DestinationConnection:         prepared.credential.ID,
		DestinationCredentialRevision: prepared.runtime.CredentialRevision,
		StreamTables: []AuthorizationStreamTable{{
			Stream: "issues", SourceTable: "issues", DestinationTable: issueLabelTransportTable,
		}},
		FieldMappings: fieldMappings,
		// The writer gate compares WriteAction directly to the declaration-owned
		// operation name. The mode remains independently bound in
		// EnabledOperations, so set-replace and keyed grants cannot be exchanged.
		WriteAction:                    action.name,
		DestinationConfigurationDigest: prepared.runtime.ConfigurationDigest,
		EnabledOperations:              []string{string(mode), action.name},
		ConfirmationPolicy:             plan.ConfirmationPolicy,
		ExpiresAt:                      plan.ExpiresAt,
	})
}

func issueLabelTransportRuntimeEqual(got, want connectors.RuntimeConfig) bool {
	return constantTimeStringEqual(got.CredentialRevision, want.CredentialRevision) &&
		constantTimeStringEqual(got.ConfigurationDigest, want.ConfigurationDigest) &&
		got.WriteApprovalScope == want.WriteApprovalScope
}

func issueLabelTransportPlanName(mode string) string {
	if mode == reversePlanModeIssueLabelTransportCleanup {
		return reversePlanModeIssueLabelTransportCleanup
	}
	return reversePlanModeIssueLabelTransport
}

func isIssueLabelTransportMode(mode string) bool {
	return mode == reversePlanModeIssueLabelTransport || mode == reversePlanModeIssueLabelTransportCleanup
}

func issueLabelTransportConfigEqual(left, right map[string]string) bool {
	if len(left) != len(right) {
		return false
	}
	for key, value := range left {
		if right[key] != value {
			return false
		}
	}
	return true
}

func issueLabelTransportEndpointEqual(left, right EndpointConfig) bool {
	return left.Connector == right.Connector &&
		left.Credential == right.Credential &&
		issueLabelTransportConfigEqual(left.Config, right.Config)
}

func (a *App) markIssueLabelTransportPlanExecuted(planID string) error {
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID != planID {
				continue
			}
			if current.ReversePlans[i].Status != reversePlanStatusApprovalConsumptionUncertain {
				return current, fmt.Errorf("closed issue-label transport plan %q is not awaiting write completion", planID)
			}
			current.ReversePlans[i].Status = "executed"
			current.ReversePlans[i].ApprovalUncertainAt = time.Time{}
			return current, nil
		}
		return current, fmt.Errorf("closed issue-label transport plan %q not found", planID)
	})
	if err != nil {
		return err
	}
	a.state = updated
	return nil
}
