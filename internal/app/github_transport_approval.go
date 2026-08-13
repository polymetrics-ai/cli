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
	reversePlanModeGitHubIssueLabelTransport        = "github_issue_label_transport"
	reversePlanModeGitHubIssueLabelTransportCleanup = "github_issue_label_transport_cleanup"
	githubIssueLabelTransportBindingDomain          = "github_issue_label_transport/v1"
	githubIssueLabelTransportTable                  = "github_transport_issue_label"
)

// githubIssueLabelTransportBinding is deliberately closed over the only
// deterministic destination this walking slice supports. It is not a generic
// action, URL, or record capability: the connection supplies every value and
// App rederives it before a grant can be consumed.
type githubIssueLabelTransportBinding struct {
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

type githubIssueLabelPreparedWrite struct {
	writer  connectors.Connector
	runtime connectors.RuntimeConfig
	record  connectors.Record
	preview connectors.WritePreview
}

// PlanGitHubIssueLabelTransport creates the forward half of the explicit PM
// plan -> preview -> approval -> execute lifecycle before a transport run.
// It prepares only the fixed destination configured on the named GitHub
// connection; it does not read, stage, reopen, or mutate a provider resource.
func (a *App) PlanGitHubIssueLabelTransport(ctx context.Context, connectionID string) (ReversePlan, error) {
	conn, err := a.githubIssueLabelTransportConnection(connectionID)
	if err != nil {
		return ReversePlan{}, err
	}
	return a.planGitHubIssueLabelTransport(ctx, conn, reversePlanModeGitHubIssueLabelTransport, githubIssueAddLabelAction, conn.Destination, "")
}

// PlanGitHubIssueLabelTransportCleanup derives the only permitted inverse from
// a completed forward plan. Callers cannot select another action, issue, label,
// URL, or reuse the forward plan's approval token.
func (a *App) PlanGitHubIssueLabelTransportCleanup(ctx context.Context, connectionID, forwardPlanID string) (ReversePlan, error) {
	conn, err := a.githubIssueLabelTransportConnection(connectionID)
	if err != nil {
		return ReversePlan{}, err
	}
	forward, endpoint, err := a.authenticatedGitHubIssueLabelTransportForwardPlan(ctx, conn, forwardPlanID)
	if err != nil {
		return ReversePlan{}, err
	}
	return a.planGitHubIssueLabelTransport(ctx, conn, reversePlanModeGitHubIssueLabelTransportCleanup, githubIssueRemoveLabelAction, endpoint, forward.ID)
}

func (a *App) planGitHubIssueLabelTransport(ctx context.Context, conn Connection, mode, action string, endpoint EndpointConfig, forwardPlanID string) (ReversePlan, error) {
	if a == nil || a.approval == nil {
		return ReversePlan{}, fmt.Errorf("GitHub transport approval requires an app approval authority")
	}
	prepared, err := a.prepareGitHubIssueLabelTransportWrite(ctx, endpoint, action)
	if err != nil {
		return ReversePlan{}, err
	}
	binding, err := a.githubIssueLabelTransportBinding(conn, mode, action, prepared.runtime, prepared.preview, prepared.runtime.Config, forwardPlanID)
	if err != nil {
		return ReversePlan{}, err
	}
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return ReversePlan{}, fmt.Errorf("hash GitHub transport destination binding: %w", err)
	}
	planHash, err := githubIssueLabelTransportPlanHash(mode, bindingSHA256)
	if err != nil {
		return ReversePlan{}, err
	}
	confirmation := a.confirmationPolicyForAction("github", action)
	if confirmation.Kind != connectors.ConfirmationKindDestructive {
		return ReversePlan{}, fmt.Errorf("GitHub transport action %q must require destructive confirmation", action)
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, err
	}
	seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
		PlanID:              id,
		PlanHash:            planHash,
		Mode:                mode,
		Connector:           "github",
		Operation:           action,
		CredentialRevision:  prepared.runtime.CredentialRevision,
		ConfigurationDigest: prepared.runtime.ConfigurationDigest,
		Batchable:           a.actionIsBatchable("github", action),
		Scope:               prepared.runtime.WriteApprovalScope,
		Confirmation:        confirmation,
	})
	if err != nil {
		return ReversePlan{}, err
	}
	plan := ReversePlan{
		ID:                     id,
		Name:                   githubIssueLabelTransportPlanName(mode),
		Status:                 "planned",
		Mode:                   mode,
		SourceConnection:       conn.Name,
		DestinationConnector:   "github",
		DestinationCredential:  endpoint.Credential,
		DestinationConfig:      cloneStringMap(endpoint.Config),
		Action:                 action,
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

// PreviewGitHubIssueLabelTransport re-prepares the exact fixed record and
// persists the existing project-scoped destructive grant. It is an explicit
// pre-run operation; Transport never calls it and never mints a token.
func (a *App) PreviewGitHubIssueLabelTransport(ctx context.Context, planID string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(planID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if !isGitHubIssueLabelTransportMode(plan.Mode) {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("reverse plan %q is not a GitHub transport label plan", plan.ID)
	}
	if err := approvalConsumptionUncertainError(plan, nil); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	conn, err := a.githubIssueLabelTransportConnection(plan.TransportConnectionID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.validateGitHubIssueLabelTransportPlan(plan, conn, plan.Mode, plan.Action, plan.TransportForwardPlanID); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	endpoint := EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: cloneStringMap(plan.DestinationConfig)}
	if plan.Mode == reversePlanModeGitHubIssueLabelTransportCleanup {
		_, authenticatedEndpoint, err := a.authenticatedGitHubIssueLabelTransportForwardPlan(ctx, conn, plan.TransportForwardPlanID)
		if err != nil {
			return ReversePlan{}, connectors.WritePreview{}, err
		}
		if !githubIssueLabelTransportEndpointEqual(endpoint, authenticatedEndpoint) {
			return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("GitHub transport cleanup plan no longer matches its authenticated forward destination")
		}
		endpoint = authenticatedEndpoint
	}
	prepared, err := a.prepareGitHubIssueLabelTransportWrite(ctx, endpoint, plan.Action)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	binding, err := a.githubIssueLabelTransportBinding(conn, plan.Mode, plan.Action, prepared.runtime, prepared.preview, prepared.runtime.Config, plan.TransportForwardPlanID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := validateGitHubIssueLabelTransportBinding(plan, binding); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	issued, err := a.persistDestructivePreview(plan, prepared.preview)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	return issued, prepared.preview, nil
}

// ApplyGitHubIssueLabelTransport validates the durable receipt and reopened
// singleton against the pre-run binding before consuming the grant and handing
// opaque approval evidence to the declarative engine immediately before I/O.
func (a *App) ApplyGitHubIssueLabelTransport(ctx context.Context, connectionID string, approval synctransport.DestinationApproval, runtime connectors.RuntimeConfig, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset) (connectors.WriteResult, error) {
	if err := validateGitHubIssueLabelTransportApproval(approval); err != nil {
		return connectors.WriteResult{}, err
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	conn, err := a.githubIssueLabelTransportConnection(connectionID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := a.validateGitHubIssueLabelTransportPlan(plan, conn, reversePlanModeGitHubIssueLabelTransport, githubIssueAddLabelAction, ""); err != nil {
		return connectors.WriteResult{}, err
	}
	if err := validateGitHubIssueLabelTransportWorkset(conn, receipt, workset); err != nil {
		return connectors.WriteResult{}, err
	}
	prepared, err := a.prepareGitHubIssueLabelTransportWrite(ctx, conn.Destination, githubIssueAddLabelAction)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if !githubIssueLabelTransportRuntimeEqual(runtime, prepared.runtime) {
		return connectors.WriteResult{}, fmt.Errorf("GitHub transport destination runtime does not match the connection-owned approval identity")
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return connectors.WriteResult{}, err
	}
	binding, err := a.githubIssueLabelTransportBinding(conn, plan.Mode, plan.Action, prepared.runtime, prepared.preview, prepared.runtime.Config, "")
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := validateGitHubIssueLabelTransportBinding(plan, binding); err != nil {
		return connectors.WriteResult{}, err
	}
	evidence, _, err := a.consumePlanApproval(plan, RunReverseETLRequest{
		PlanID:        approval.PlanID,
		ApprovalToken: approval.ApprovalToken,
		Confirmation:  approval.Confirmation,
	}, prepared.preview)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if evidence == nil {
		return connectors.WriteResult{}, fmt.Errorf("GitHub transport approval did not yield destructive write evidence")
	}
	result, err := prepared.writer.Write(ctx, connectors.WriteRequest{
		Stream:   "issues",
		Table:    githubIssueLabelTransportTable,
		Action:   githubIssueAddLabelAction,
		Config:   prepared.runtime,
		Approval: evidence,
	}, []connectors.Record{prepared.record})
	if err != nil {
		return connectors.WriteResult{}, fmt.Errorf("execute approved GitHub add-issue-label transport: %w", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		return connectors.WriteResult{}, fmt.Errorf("approved GitHub add-issue-label result written=%d failed=%d, want one durable write", result.RecordsWritten, result.RecordsFailed)
	}
	if err := a.markGitHubIssueLabelTransportPlanExecuted(plan.ID); err != nil {
		return connectors.WriteResult{}, err
	}
	return result, nil
}

// ApplyGitHubIssueLabelTransportCleanup executes only the inverse derived from
// its closed forward plan. It has a distinct preview, token, grant, and
// evidence; the forward approval is not accepted here.
func (a *App) ApplyGitHubIssueLabelTransportCleanup(ctx context.Context, connectionID string, approval synctransport.DestinationApproval) (connectors.WriteResult, error) {
	if err := validateGitHubIssueLabelTransportApproval(approval); err != nil {
		return connectors.WriteResult{}, err
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	conn, err := a.githubIssueLabelTransportConnection(connectionID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := a.validateGitHubIssueLabelTransportPlan(plan, conn, reversePlanModeGitHubIssueLabelTransportCleanup, githubIssueRemoveLabelAction, plan.TransportForwardPlanID); err != nil {
		return connectors.WriteResult{}, err
	}
	forward, endpoint, err := a.authenticatedGitHubIssueLabelTransportForwardPlan(ctx, conn, plan.TransportForwardPlanID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if !githubIssueLabelTransportEndpointEqual(EndpointConfig{Connector: plan.DestinationConnector, Credential: plan.DestinationCredential, Config: cloneStringMap(plan.DestinationConfig)}, endpoint) {
		return connectors.WriteResult{}, fmt.Errorf("GitHub transport cleanup plan no longer matches its authenticated forward destination")
	}
	prepared, err := a.prepareGitHubIssueLabelTransportWrite(ctx, endpoint, githubIssueRemoveLabelAction)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return connectors.WriteResult{}, err
	}
	binding, err := a.githubIssueLabelTransportBinding(conn, plan.Mode, plan.Action, prepared.runtime, prepared.preview, prepared.runtime.Config, forward.ID)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if err := validateGitHubIssueLabelTransportBinding(plan, binding); err != nil {
		return connectors.WriteResult{}, err
	}
	evidence, _, err := a.consumePlanApproval(plan, RunReverseETLRequest{
		PlanID:        approval.PlanID,
		ApprovalToken: approval.ApprovalToken,
		Confirmation:  approval.Confirmation,
	}, prepared.preview)
	if err != nil {
		return connectors.WriteResult{}, err
	}
	if evidence == nil {
		return connectors.WriteResult{}, fmt.Errorf("GitHub transport cleanup approval did not yield destructive write evidence")
	}
	result, err := prepared.writer.Write(ctx, connectors.WriteRequest{
		Stream:   "issues",
		Table:    githubIssueLabelTransportTable,
		Action:   githubIssueRemoveLabelAction,
		Config:   prepared.runtime,
		Approval: evidence,
	}, []connectors.Record{prepared.record})
	if err != nil {
		return connectors.WriteResult{}, fmt.Errorf("execute approved GitHub remove-issue-label transport cleanup: %w", err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		return connectors.WriteResult{}, fmt.Errorf("approved GitHub remove-issue-label result written=%d failed=%d, want one durable write", result.RecordsWritten, result.RecordsFailed)
	}
	if err := a.markGitHubIssueLabelTransportPlanExecuted(plan.ID); err != nil {
		return connectors.WriteResult{}, err
	}
	return result, nil
}

func (a *App) prepareGitHubIssueLabelTransportWrite(ctx context.Context, endpoint EndpointConfig, action string) (githubIssueLabelPreparedWrite, error) {
	if endpoint.Connector != "github" {
		return githubIssueLabelPreparedWrite{}, fmt.Errorf("GitHub transport approval requires the GitHub destination connector")
	}
	writer, runtime, err := a.resolveEndpoint(ctx, endpoint)
	if err != nil {
		return githubIssueLabelPreparedWrite{}, fmt.Errorf("resolve GitHub transport destination: %w", err)
	}
	if err := githubTransportRepositoryConfig(runtime.Config); err != nil {
		return githubIssueLabelPreparedWrite{}, err
	}
	targetIssue, err := githubTransportIssueNumber(runtime.Config, githubTransportTargetIssueConfig)
	if err != nil {
		return githubIssueLabelPreparedWrite{}, err
	}
	label, err := githubTransportLabel(runtime.Config)
	if err != nil {
		return githubIssueLabelPreparedWrite{}, err
	}
	if writer.Name() != "github" || !writer.Metadata().Capabilities.Write {
		return githubIssueLabelPreparedWrite{}, fmt.Errorf("GitHub transport destination does not expose typed writes")
	}
	record, err := githubIssueLabelTransportRecord(action, targetIssue, label)
	if err != nil {
		return githubIssueLabelPreparedWrite{}, err
	}
	request := connectors.WriteRequest{Stream: "issues", Table: githubIssueLabelTransportTable, Action: action, Config: runtime}
	if validator, ok := writer.(connectors.WriteValidator); ok {
		if err := validator.ValidateWrite(ctx, request, []connectors.Record{record}); err != nil {
			return githubIssueLabelPreparedWrite{}, fmt.Errorf("validate GitHub transport write: %w", err)
		}
	}
	dryRunner, ok := writer.(connectors.DryRunWriter)
	if !ok {
		return githubIssueLabelPreparedWrite{}, fmt.Errorf("GitHub transport destination does not support typed write previews")
	}
	preview, err := dryRunner.DryRunWrite(ctx, request, []connectors.Record{record})
	if err != nil {
		return githubIssueLabelPreparedWrite{}, fmt.Errorf("prepare GitHub transport write: %w", err)
	}
	if strings.TrimSpace(preview.Digest) == "" || preview.ApprovalTarget.Connector != "github" || preview.ApprovalTarget.Operation != action || preview.ApprovalTarget.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		return githubIssueLabelPreparedWrite{}, fmt.Errorf("GitHub transport write preview does not bind a destructive %s target", action)
	}
	return githubIssueLabelPreparedWrite{writer: writer, runtime: runtime, record: record, preview: preview}, nil
}

func (a *App) githubIssueLabelTransportBinding(conn Connection, mode, action string, runtime connectors.RuntimeConfig, preview connectors.WritePreview, config map[string]string, forwardPlanID string) (githubIssueLabelTransportBinding, error) {
	targetIssue, err := githubTransportIssueNumber(config, githubTransportTargetIssueConfig)
	if err != nil {
		return githubIssueLabelTransportBinding{}, err
	}
	label, err := githubTransportLabel(config)
	if err != nil {
		return githubIssueLabelTransportBinding{}, err
	}
	binding := githubIssueLabelTransportBinding{
		Domain:              githubIssueLabelTransportBindingDomain,
		ConnectionID:        conn.ID,
		Stream:              "issues",
		Mode:                synccontract.ModeFullAppend,
		Destination:         "github",
		Action:              action,
		TargetIssue:         targetIssue,
		Labels:              []string{label},
		CredentialRevision:  runtime.CredentialRevision,
		ConfigurationDigest: runtime.ConfigurationDigest,
		ApprovalScope:       runtime.WriteApprovalScope,
		PreviewDigest:       preview.Digest,
		ForwardPlanID:       forwardPlanID,
	}
	if mode == reversePlanModeGitHubIssueLabelTransport {
		sourceIssue, err := githubTransportIssueNumber(conn.Source.Config, githubTransportSourceIssueConfig)
		if err != nil {
			return githubIssueLabelTransportBinding{}, err
		}
		binding.ExpectedSourceIssue = sourceIssue
	}
	return binding, nil
}

func githubIssueLabelTransportPlanHash(mode, bindingSHA256 string) (string, error) {
	if !isGitHubIssueLabelTransportMode(mode) || strings.TrimSpace(bindingSHA256) == "" {
		return "", fmt.Errorf("GitHub transport plan requires a known mode and destination binding")
	}
	return hashJSON(struct {
		Domain              string `json:"domain"`
		Mode                string `json:"mode"`
		TransportBindingSHA string `json:"transport_binding_sha256"`
	}{
		Domain:              githubIssueLabelTransportBindingDomain,
		Mode:                mode,
		TransportBindingSHA: bindingSHA256,
	})
}

func validateGitHubIssueLabelTransportBinding(plan ReversePlan, binding githubIssueLabelTransportBinding) error {
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return fmt.Errorf("hash GitHub transport destination binding: %w", err)
	}
	if !constantTimeStringEqual(plan.TransportBindingSHA256, bindingSHA256) {
		return fmt.Errorf("GitHub transport approval binding does not match the reopened connection workset or destination identity")
	}
	planHash, err := githubIssueLabelTransportPlanHash(plan.Mode, bindingSHA256)
	if err != nil {
		return err
	}
	if !constantTimeStringEqual(plan.PlanHash, planHash) {
		return fmt.Errorf("GitHub transport approval plan hash does not match its destination binding")
	}
	return nil
}

func (a *App) githubIssueLabelTransportConnection(connectionID string) (Connection, error) {
	if a == nil {
		return Connection{}, fmt.Errorf("GitHub transport approval requires an app")
	}
	if strings.TrimSpace(connectionID) == "" {
		return Connection{}, fmt.Errorf("GitHub transport approval requires a connection ID")
	}
	conn, ok := a.findConnectionByID(connectionID)
	if !ok {
		return Connection{}, fmt.Errorf("GitHub transport connection %q was not found", connectionID)
	}
	if conn.Source.Connector != "github" || conn.Destination.Connector != "github" {
		return Connection{}, fmt.Errorf("GitHub transport connection %q must use GitHub as both source and destination", conn.ID)
	}
	stream, ok := conn.Streams["issues"]
	if !ok {
		return Connection{}, fmt.Errorf("GitHub transport connection %q has no issues stream", conn.ID)
	}
	mode, err := ParseStreamSyncMode(stream)
	if err != nil {
		return Connection{}, err
	}
	if mode.ContractMode != synccontract.ModeFullAppend {
		return Connection{}, fmt.Errorf("GitHub transport connection %q must use issues/full_append", conn.ID)
	}
	if _, err := githubTransportIssueNumber(conn.Source.Config, githubTransportSourceIssueConfig); err != nil {
		return Connection{}, err
	}
	if _, err := githubTransportIssueNumber(conn.Destination.Config, githubTransportTargetIssueConfig); err != nil {
		return Connection{}, err
	}
	if _, err := githubTransportLabel(conn.Destination.Config); err != nil {
		return Connection{}, err
	}
	return conn, nil
}

func (a *App) validateGitHubIssueLabelTransportPlan(plan ReversePlan, conn Connection, mode, action, forwardPlanID string) error {
	if plan.Mode != mode || plan.TransportConnectionID != conn.ID || plan.SourceConnection != conn.Name || plan.DestinationConnector != "github" || plan.Action != action || plan.RecordCount != 1 || plan.ConnectorCommandOperation != "" || len(plan.Mappings) != 0 || plan.TransportForwardPlanID != forwardPlanID || strings.TrimSpace(plan.TransportBindingSHA256) == "" {
		return fmt.Errorf("GitHub transport approval plan does not bind the exact connection-owned %s mutation", action)
	}
	if plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive || a.confirmationPolicyForPlan(plan).Kind != connectors.ConfirmationKindDestructive {
		return fmt.Errorf("GitHub transport approval plan does not require destructive confirmation")
	}
	return nil
}

func (a *App) authenticatedGitHubIssueLabelTransportForwardPlan(ctx context.Context, conn Connection, id string) (ReversePlan, EndpointConfig, error) {
	forward, err := a.GetReversePlan(id)
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	if forward.Mode != reversePlanModeGitHubIssueLabelTransport || forward.TransportConnectionID != conn.ID || forward.Action != githubIssueAddLabelAction || strings.TrimSpace(forward.TransportBindingSHA256) == "" || forward.Status != "executed" {
		return ReversePlan{}, EndpointConfig{}, fmt.Errorf("GitHub transport cleanup requires an executed forward issue-label plan for connection %q", conn.ID)
	}
	if err := a.validateGitHubIssueLabelTransportPlan(forward, conn, reversePlanModeGitHubIssueLabelTransport, githubIssueAddLabelAction, ""); err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	endpoint := EndpointConfig{
		Connector:  forward.DestinationConnector,
		Credential: forward.DestinationCredential,
		Config:     cloneStringMap(forward.DestinationConfig),
	}
	prepared, err := a.prepareGitHubIssueLabelTransportWrite(ctx, endpoint, githubIssueAddLabelAction)
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	if err := a.verifyPlanSealForRuntime(forward, prepared.runtime); err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	binding, err := a.githubIssueLabelTransportBinding(conn, forward.Mode, forward.Action, prepared.runtime, prepared.preview, prepared.runtime.Config, "")
	if err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	if err := validateGitHubIssueLabelTransportBinding(forward, binding); err != nil {
		return ReversePlan{}, EndpointConfig{}, err
	}
	return forward, endpoint, nil
}

func validateGitHubIssueLabelTransportWorkset(conn Connection, receipt synctransport.WarehouseReceipt, workset synctransport.WarehouseWorkset) error {
	if err := receipt.Validate(); err != nil {
		return fmt.Errorf("GitHub transport receipt: %w", err)
	}
	if receipt.Owner != conn.ID || receipt.ID != workset.ID || receipt.Stream != "issues" || receipt.Mode != synccontract.ModeFullAppend || receipt.Records != len(workset.Records) || receipt.Tombstones != len(workset.Tombstones) {
		return fmt.Errorf("GitHub transport receipt does not bind the reopened workset")
	}
	if len(workset.Records) != 1 {
		return fmt.Errorf("GitHub transport requires exactly one reopened source issue")
	}
	expectedSource, err := githubTransportIssueNumber(conn.Source.Config, githubTransportSourceIssueConfig)
	if err != nil {
		return err
	}
	actualSource, err := githubIssueNumberFromRecord(workset.Records[0])
	if err != nil {
		return fmt.Errorf("GitHub transport reopened source issue: %w", err)
	}
	if actualSource != expectedSource {
		return fmt.Errorf("GitHub transport reopened source issue %d does not match configured source issue %d", actualSource, expectedSource)
	}
	return nil
}

func validateGitHubIssueLabelTransportApproval(approval synctransport.DestinationApproval) error {
	if strings.TrimSpace(approval.PlanID) == "" || strings.TrimSpace(approval.ApprovalToken) == "" || approval.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		return fmt.Errorf("GitHub transport requires a pre-run plan-preview-approved destructive grant")
	}
	return nil
}

func githubIssueLabelTransportRuntimeEqual(got, want connectors.RuntimeConfig) bool {
	return constantTimeStringEqual(got.CredentialRevision, want.CredentialRevision) &&
		constantTimeStringEqual(got.ConfigurationDigest, want.ConfigurationDigest) &&
		got.WriteApprovalScope == want.WriteApprovalScope
}

func githubIssueLabelTransportRecord(action string, issueNumber int, label string) (connectors.Record, error) {
	if issueNumber <= 0 || strings.TrimSpace(label) == "" {
		return nil, fmt.Errorf("GitHub issue-label transport requires a positive issue number and non-empty label")
	}
	switch action {
	case githubIssueAddLabelAction:
		return connectors.Record{"issue_number": issueNumber, "labels": []string{label}}, nil
	case githubIssueRemoveLabelAction:
		return connectors.Record{"issue_number": issueNumber, "name": label}, nil
	default:
		return nil, fmt.Errorf("GitHub issue-label transport action %q is not declared", action)
	}
}

func githubIssueLabelTransportPlanName(mode string) string {
	if mode == reversePlanModeGitHubIssueLabelTransportCleanup {
		return "github_issue_label_transport_cleanup"
	}
	return "github_issue_label_transport"
}

func isGitHubIssueLabelTransportMode(mode string) bool {
	return mode == reversePlanModeGitHubIssueLabelTransport || mode == reversePlanModeGitHubIssueLabelTransportCleanup
}

func githubIssueLabelTransportConfigEqual(left, right map[string]string) bool {
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

func githubIssueLabelTransportEndpointEqual(left, right EndpointConfig) bool {
	return left.Connector == right.Connector &&
		left.Credential == right.Credential &&
		githubIssueLabelTransportConfigEqual(left.Config, right.Config)
}

func (a *App) markGitHubIssueLabelTransportPlanExecuted(planID string) error {
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID != planID {
				continue
			}
			if current.ReversePlans[i].Status != reversePlanStatusApprovalConsumptionUncertain {
				return current, fmt.Errorf("GitHub transport plan %q is not awaiting write completion", planID)
			}
			current.ReversePlans[i].Status = "executed"
			current.ReversePlans[i].ApprovalUncertainAt = time.Time{}
			return current, nil
		}
		return current, fmt.Errorf("GitHub transport plan %q not found", planID)
	})
	if err != nil {
		return err
	}
	a.state = updated
	return nil
}
