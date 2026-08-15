package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/database"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	reversePlanModePostgresManagedTarget = "postgres_managed_target_transport"
	postgresManagedTargetOperation       = "postgres_managed_target"
	postgresManagedTargetBindingDomain   = "postgres_managed_target_transport/v1"
)

var (
	ErrPostgresManagedTargetApprovalRequired = errors.New("PostgreSQL managed target approval is required")
	ErrPostgresManagedTargetApprovalStale    = errors.New("PostgreSQL managed target approval is stale")
)

type postgresManagedTargetApprovalBinding struct {
	Domain              string            `json:"domain"`
	ConnectionID        string            `json:"connection_id"`
	Stream              string            `json:"stream"`
	StreamID            string            `json:"stream_id"`
	Mode                synccontract.Mode `json:"mode"`
	SourceConnector     string            `json:"source_connector"`
	SourceSchema        string            `json:"source_schema_fingerprint"`
	SourceCredential    string            `json:"source_credential_revision"`
	SourceConfiguration string            `json:"source_configuration_digest"`
	Destination         string            `json:"destination"`
	PrimaryKey          []string          `json:"primary_key,omitempty"`
	CredentialRevision  string            `json:"credential_revision"`
	ConfigurationDigest string            `json:"configuration_digest"`
	ApprovalScope       string            `json:"approval_scope"`
}

func (a *App) PlanPostgresManagedTargetTransport(ctx context.Context, connectionName, streamName string) (ReversePlan, error) {
	conn, stream, mode, runtime, binding, err := a.preparePostgresManagedTargetApproval(ctx, connectionName, streamName)
	if err != nil {
		return ReversePlan{}, err
	}
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return ReversePlan{}, err
	}
	planHash, err := hashJSON(struct {
		Domain  string `json:"domain"`
		Binding string `json:"binding"`
	}{postgresManagedTargetBindingDomain, bindingSHA256})
	if err != nil {
		return ReversePlan{}, err
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, err
	}
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	action := postgresManagedTargetAction(mode.ContractMode)
	seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
		PlanID: id, PlanHash: planHash, Mode: reversePlanModePostgresManagedTarget,
		Connector: conn.Destination.Connector, Operation: action,
		CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
		Batchable: true, Scope: runtime.WriteApprovalScope, Confirmation: confirmation,
	})
	if err != nil {
		return ReversePlan{}, err
	}
	plan := ReversePlan{
		ID: id, Name: "PostgreSQL managed target transport", Status: "planned",
		Mode: reversePlanModePostgresManagedTarget, SourceConnection: conn.Name,
		DestinationConnector: conn.Destination.Connector, DestinationCredential: conn.Destination.Credential,
		DestinationConfig: cloneStringMap(conn.Destination.Config), Action: action,
		Mappings: map[string]string{}, ConfirmationChallenge: string(confirmation.Kind), ConfirmationPolicy: confirmation,
		RecordCount: 0, PlanHash: planHash, PlanSeal: &seal,
		TransportConnectionID: conn.ID, TransportStream: streamName, TransportBindingSHA256: bindingSHA256,
		CreatedAt: seal.IssuedAt, ExpiresAt: seal.ExpiresAt,
	}
	_ = stream
	a.state.ReversePlans = append(a.state.ReversePlans, plan)
	if err := a.save(); err != nil {
		return ReversePlan{}, err
	}
	return plan, nil
}

func (a *App) PreviewPostgresManagedTargetTransport(ctx context.Context, planID string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(planID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if plan.Mode != reversePlanModePostgresManagedTarget {
		return ReversePlan{}, connectors.WritePreview{}, errors.New("reverse plan is not a PostgreSQL managed target transport plan")
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	_, _, _, runtime, binding, err := a.preparePostgresManagedTargetApproval(ctx, plan.SourceConnection, plan.TransportStream)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	preview, err := a.validatePostgresManagedTargetPlan(plan, runtime, binding)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	issued, err := a.persistDestructivePreview(plan, preview)
	return issued, preview, err
}

func (a *App) authorizePostgresManagedTargetTransport(ctx context.Context, conn Connection, streamName string, approval synctransport.DestinationApproval, runtime connectors.RuntimeConfig) (synctransport.DestinationApproval, error) {
	if strings.TrimSpace(approval.PlanID) == "" || strings.TrimSpace(approval.ApprovalToken) == "" {
		return synctransport.DestinationApproval{}, fmt.Errorf("%w: transport requires a previewed approval plan and stdin token", ErrPostgresManagedTargetApprovalRequired)
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if plan.Mode == reversePlanModePostgresManagedTarget && plan.Status == "executed" {
		return synctransport.DestinationApproval{}, &AuthorizationTokenReplayError{Reference: plan.ID}
	}
	if plan.Mode != reversePlanModePostgresManagedTarget || plan.TransportConnectionID != conn.ID || plan.TransportStream != streamName {
		return synctransport.DestinationApproval{}, errors.New("PostgreSQL managed target approval does not match this connection stream")
	}
	_, _, _, resolvedRuntime, binding, err := a.preparePostgresManagedTargetApproval(ctx, conn.Name, streamName)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if !runtimeApprovalIdentityEqual(runtime, resolvedRuntime) {
		return synctransport.DestinationApproval{}, errors.New("PostgreSQL managed target approval runtime changed")
	}
	preview, err := a.validatePostgresManagedTargetPlan(plan, resolvedRuntime, binding)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	evidence, _, err := a.consumePlanApproval(plan, RunReverseETLRequest{
		PlanID: approval.PlanID, ApprovalToken: approval.ApprovalToken, Confirmation: approval.Confirmation,
	}, preview, nil)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if evidence == nil {
		return synctransport.DestinationApproval{}, errors.New("PostgreSQL managed target approval produced no authenticated evidence")
	}
	approval.ApprovalToken = ""
	approval.Evidence = evidence
	approval.Target = preview.ApprovalTarget
	approval.PreviewDigest = preview.Digest
	return approval, nil
}

func (a *App) markPostgresManagedTargetPlanExecuted(planID string) error {
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID != planID {
				continue
			}
			if current.ReversePlans[i].Mode != reversePlanModePostgresManagedTarget || current.ReversePlans[i].Status != reversePlanStatusApprovalConsumptionUncertain {
				return current, fmt.Errorf("PostgreSQL managed target plan %q is not awaiting transport completion", planID)
			}
			current.ReversePlans[i].Status = "executed"
			current.ReversePlans[i].ApprovalUncertainAt = time.Time{}
			return current, nil
		}
		return current, fmt.Errorf("PostgreSQL managed target plan %q not found", planID)
	})
	if err != nil {
		return err
	}
	a.state = updated
	return nil
}

func (a *App) validatePostgresManagedTargetPlan(plan ReversePlan, runtime connectors.RuntimeConfig, binding postgresManagedTargetApprovalBinding) (connectors.WritePreview, error) {
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	if plan.TransportBindingSHA256 != bindingSHA256 || plan.DestinationConnector != binding.Destination || plan.Action != postgresManagedTargetAction(binding.Mode) {
		return connectors.WritePreview{}, fmt.Errorf("%w: plan no longer matches the connection", ErrPostgresManagedTargetApprovalStale)
	}
	planHash, err := hashJSON(struct {
		Domain  string `json:"domain"`
		Binding string `json:"binding"`
	}{postgresManagedTargetBindingDomain, bindingSHA256})
	if err != nil || plan.PlanHash != planHash {
		return connectors.WritePreview{}, fmt.Errorf("%w: plan hash changed", ErrPostgresManagedTargetApprovalStale)
	}
	if err := a.verifyPlanSealForRuntime(plan, runtime); err != nil {
		return connectors.WritePreview{}, err
	}
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	target := connectors.WriteApprovalTarget{
		Connector: binding.Destination, Operation: plan.Action, Method: "POSTGRESQL",
		MutationClass: string(binding.Mode), TargetDigest: bindingSHA256,
		CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest,
		Batchable: true, Scope: runtime.WriteApprovalScope, Confirmation: confirmation,
	}
	digest, err := hashJSON(struct {
		Domain string                         `json:"domain"`
		Target connectors.WriteApprovalTarget `json:"target"`
	}{postgresManagedTargetBindingDomain, target})
	if err != nil {
		return connectors.WritePreview{}, err
	}
	return connectors.WritePreview{RecordsStaged: 0, Action: plan.Action, Digest: digest, ApprovalTarget: target}, nil
}

func (a *App) preparePostgresManagedTargetApproval(ctx context.Context, connectionName, streamName string) (Connection, StreamConfig, SyncMode, connectors.RuntimeConfig, postgresManagedTargetApprovalBinding, error) {
	conn, ok := a.findConnection(connectionName)
	if !ok {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, fmt.Errorf("connection %q not found", connectionName)
	}
	stream, ok := conn.Streams[streamName]
	if !ok || strings.TrimSpace(stream.StreamID) == "" {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, fmt.Errorf("stream %q is not configured with structural identity", streamName)
	}
	mode, err := ParseStreamSyncMode(stream)
	if err != nil || !mode.IsContractMode() {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, errors.New("PostgreSQL managed target requires a contract sync mode")
	}
	source, sourceRuntime, err := a.resolveEndpoint(ctx, conn.Source)
	if err != nil {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, err
	}
	destination, runtime, err := a.resolveEndpoint(ctx, conn.Destination)
	if err != nil {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, err
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{
		Source: source, Destination: destination, Stream: streamName, Mode: mode.ContractMode,
	})
	if err != nil {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, err
	}
	if _, ok := resolved.Destination.(synctransport.ManagedTargetApprovalDestination); !ok {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, errors.New("closed managed target approval requires its declared destination executor")
	}
	sourceSchema, err := postgresManagedTargetSourceSchema(ctx, source, sourceRuntime, streamName)
	if err != nil {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, err
	}
	binding := postgresManagedTargetApprovalBinding{
		Domain: postgresManagedTargetBindingDomain, ConnectionID: conn.ID, Stream: streamName, StreamID: stream.StreamID,
		Mode: mode.ContractMode, SourceConnector: conn.Source.Connector, SourceSchema: sourceSchema,
		SourceCredential: sourceRuntime.CredentialRevision, SourceConfiguration: sourceRuntime.ConfigurationDigest,
		Destination: conn.Destination.Connector,
		PrimaryKey:  append([]string(nil), stream.PrimaryKey...), CredentialRevision: runtime.CredentialRevision,
		ConfigurationDigest: runtime.ConfigurationDigest, ApprovalScope: runtime.WriteApprovalScope,
	}
	return conn, stream, mode, runtime, binding, nil
}

func postgresManagedTargetSourceSchema(ctx context.Context, source connectors.Connector, runtime connectors.RuntimeConfig, streamName string) (string, error) {
	catalog, err := database.CatalogForManagedTargetSource(ctx, source, runtime, streamName)
	if err != nil {
		return "", err
	}
	return catalog.Fingerprint().String(), nil
}

func postgresManagedTargetAction(mode synccontract.Mode) string { return "managed_" + string(mode) }

func runtimeApprovalIdentityEqual(left, right connectors.RuntimeConfig) bool {
	return left.CredentialRevision == right.CredentialRevision && left.ConfigurationDigest == right.ConfigurationDigest && left.WriteApprovalScope == right.WriteApprovalScope
}
