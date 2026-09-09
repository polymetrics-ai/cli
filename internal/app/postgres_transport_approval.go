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
	reversePlanModePostgresManagedTarget              = "postgres_managed_target_transport"
	postgresManagedTargetOperation                    = "postgres_managed_target"
	postgresManagedTargetBindingDomain                = "postgres_managed_target_transport/v1"
	postgresManagedTargetAuthorizationDefaultLifetime = 24 * time.Hour
	postgresManagedTargetAuthorizationMinimumLifetime = 24 * time.Hour
	postgresManagedTargetAuthorizationMaximumLifetime = 48 * time.Hour
)

var (
	ErrPostgresManagedTargetApprovalRequired      = errors.New("PostgreSQL managed target approval is required")
	ErrPostgresManagedTargetApprovalStale         = errors.New("PostgreSQL managed target approval is stale")
	ErrPostgresManagedTargetAuthorizationLifetime = errors.New("PostgreSQL managed target authorization lifetime must be between 24h and 48h")
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
	// TransformPlanHash binds the normalized closed mapping into the plan,
	// preview digest, approval target and durable authorization scope. It is
	// deliberately a hash only; no expression or source SQL reaches approval
	// persistence.
	TransformPlanHash       string `json:"transform_plan_hash,omitempty"`
	TargetCopyWorkers       int    `json:"target_copy_workers,omitempty"`
	TargetCopyWorkerMaximum int    `json:"target_copy_worker_maximum,omitempty"`
	CredentialRevision      string `json:"credential_revision"`
	ConfigurationDigest     string `json:"configuration_digest"`
	ApprovalScope           string `json:"approval_scope"`
}

func (a *App) PlanPostgresManagedTargetTransport(ctx context.Context, connectionName, streamName string) (ReversePlan, error) {
	return a.PlanPostgresManagedTargetTransportWithAuthorizationLifetime(ctx, connectionName, streamName, 0)
}

// PlanPostgresManagedTargetTransportWithAuthorizationLifetime creates the
// one-time-token plan for the closed PostgreSQL route. The lifetime becomes
// part of its sealed, shape-bound standing authorization only after that token
// is consumed; it never extends the token or its preview grant.
func (a *App) PlanPostgresManagedTargetTransportWithAuthorizationLifetime(ctx context.Context, connectionName, streamName string, authorizationLifetime time.Duration) (ReversePlan, error) {
	authorizationLifetime, err := normalizePostgresManagedTargetAuthorizationLifetime(authorizationLifetime)
	if err != nil {
		return ReversePlan{}, err
	}
	conn, stream, mode, runtime, binding, err := a.preparePostgresManagedTargetApproval(ctx, connectionName, streamName)
	if err != nil {
		return ReversePlan{}, err
	}
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return ReversePlan{}, err
	}
	planHash, err := postgresManagedTargetPlanHash(bindingSHA256, authorizationLifetime)
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
		AuthorizationLifetime:   authorizationLifetime,
		TargetCopyWorkers:       binding.TargetCopyWorkers,
		TargetCopyWorkerMaximum: binding.TargetCopyWorkerMaximum,
		CreatedAt:               seal.IssuedAt, ExpiresAt: seal.ExpiresAt,
	}
	_ = stream
	a.state.ReversePlans = append(a.state.ReversePlans, plan)
	if err := a.save(); err != nil {
		return ReversePlan{}, err
	}
	return plan, nil
}

func normalizePostgresManagedTargetAuthorizationLifetime(lifetime time.Duration) (time.Duration, error) {
	if lifetime == 0 {
		return postgresManagedTargetAuthorizationDefaultLifetime, nil
	}
	if lifetime < postgresManagedTargetAuthorizationMinimumLifetime || lifetime > postgresManagedTargetAuthorizationMaximumLifetime {
		return 0, ErrPostgresManagedTargetAuthorizationLifetime
	}
	return lifetime, nil
}

func postgresManagedTargetPlanHash(bindingSHA256 string, authorizationLifetime time.Duration) (string, error) {
	return hashJSON(struct {
		Domain                     string        `json:"domain"`
		Binding                    string        `json:"binding"`
		AuthorizationLifetimeNanos time.Duration `json:"authorization_lifetime_ns"`
	}{postgresManagedTargetBindingDomain, bindingSHA256, authorizationLifetime})
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
	if strings.TrimSpace(approval.PlanID) == "" {
		return synctransport.DestinationApproval{}, fmt.Errorf("%w: transport requires a previewed approval plan", ErrPostgresManagedTargetApprovalRequired)
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if plan.Mode == reversePlanModePostgresManagedTarget && plan.Status == "executed" && plan.AuthorizationReference == "" && strings.TrimSpace(approval.ApprovalToken) != "" {
		return synctransport.DestinationApproval{}, &AuthorizationTokenReplayError{Reference: plan.ID}
	}
	if plan.Mode != reversePlanModePostgresManagedTarget || plan.TransportConnectionID != conn.ID || plan.TransportStream != streamName {
		return synctransport.DestinationApproval{}, errors.New("PostgreSQL managed target approval does not match this connection stream")
	}
	resolvedConn, stream, mode, resolvedRuntime, binding, err := a.preparePostgresManagedTargetApproval(ctx, conn.Name, streamName)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if !runtimeApprovalIdentityEqual(runtime, resolvedRuntime) {
		return synctransport.DestinationApproval{}, errors.New("PostgreSQL managed target approval runtime changed")
	}
	// The one-time preview seal is deliberately short lived. Once that seal
	// mints the standing authorization, later units validate the durable record
	// and its shape instead; otherwise an expired preview seal would silently
	// reintroduce a run-scoped ceiling below the authorization lifetime.
	preview, err := a.validatePostgresManagedTargetPlanWithSeal(plan, resolvedRuntime, binding, plan.AuthorizationReference == "")
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}

	var (
		scope                  AuthorizationScope
		authorizationReference string
	)
	if plan.AuthorizationReference != "" {
		if strings.TrimSpace(approval.ApprovalToken) != "" {
			return synctransport.DestinationApproval{}, &AuthorizationTokenReplayError{Reference: plan.AuthorizationReference}
		}
		record, err := a.authorizationRecord(plan.AuthorizationReference)
		if err != nil {
			return synctransport.DestinationApproval{}, err
		}
		scope, err = a.postgresManagedTargetAuthorizationScope(resolvedConn, streamName, stream, mode, plan, resolvedRuntime, binding, record.Scope.ExpiresAt)
		if err != nil {
			return synctransport.DestinationApproval{}, err
		}
		if _, err := a.requireAuthorization(record.Reference, scope, time.Now().UTC()); err != nil {
			return synctransport.DestinationApproval{}, err
		}
		authorizationReference = record.Reference
	} else {
		if strings.TrimSpace(approval.ApprovalToken) == "" {
			return synctransport.DestinationApproval{}, fmt.Errorf("%w: transport requires a previewed approval plan and stdin token", ErrPostgresManagedTargetApprovalRequired)
		}
		lifetime, err := normalizePostgresManagedTargetAuthorizationLifetime(plan.AuthorizationLifetime)
		if err != nil {
			return synctransport.DestinationApproval{}, err
		}
		scope, err = a.postgresManagedTargetAuthorizationScope(resolvedConn, streamName, stream, mode, plan, resolvedRuntime, binding, time.Now().UTC().Add(lifetime))
		if err != nil {
			return synctransport.DestinationApproval{}, err
		}
		authorization, err := newAuthorizationRecord(scope, time.Now().UTC())
		if err != nil {
			return synctransport.DestinationApproval{}, err
		}
		if _, _, err := a.consumePlanApproval(plan, RunReverseETLRequest{
			PlanID: approval.PlanID, ApprovalToken: approval.ApprovalToken, Confirmation: approval.Confirmation,
		}, preview, &authorization); err != nil {
			return synctransport.DestinationApproval{}, err
		}
		authorizationReference = authorization.Reference
	}
	evidence, err := durableAuthorizationEvidence(scope)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	approval.ApprovalToken = ""
	approval.Evidence = evidence
	approval.Target = preview.ApprovalTarget
	approval.PreviewDigest = preview.Digest
	approval.AuthorizeNextUnit = func(unitCtx context.Context) error {
		if unitCtx == nil {
			return errors.New("PostgreSQL managed target authorization context is required")
		}
		if err := unitCtx.Err(); err != nil {
			return err
		}
		_, err := a.requireAuthorization(authorizationReference, scope, time.Now().UTC())
		return err
	}
	return approval, nil
}

func (a *App) postgresManagedTargetAuthorizationScope(conn Connection, streamName string, stream StreamConfig, mode SyncMode, plan ReversePlan, runtime connectors.RuntimeConfig, binding postgresManagedTargetApprovalBinding, expiresAt time.Time) (AuthorizationScope, error) {
	credential, ok := a.findCredential(conn.Destination.Credential)
	if !ok {
		return AuthorizationScope{}, fmt.Errorf("PostgreSQL managed target authorization credential %q was not found", conn.Destination.Credential)
	}
	destinationTable := stream.DestinationTable
	if destinationTable == "" {
		destinationTable = streamName
	}
	return canonicalAuthorizationScope(AuthorizationScope{
		SourceConnection:              conn.ID,
		DestinationConnection:         credential.ID,
		DestinationCredentialRevision: runtime.CredentialRevision,
		StreamTables: []AuthorizationStreamTable{{
			Stream: streamName, SourceTable: binding.SourceSchema, DestinationTable: destinationTable,
		}},
		FieldMappings: map[string]string{
			"source_connector":            binding.SourceConnector,
			"source_credential_revision":  binding.SourceCredential,
			"source_configuration_digest": binding.SourceConfiguration,
			"sync_mode":                   string(mode.ContractMode),
			"transport_binding_sha256":    plan.TransportBindingSHA256,
			"transform_plan_hash":         binding.TransformPlanHash,
		},
		WriteAction:                    plan.Action,
		DestinationConfigurationDigest: runtime.ConfigurationDigest,
		EnabledOperations:              []string{plan.Action},
		ConfirmationPolicy:             connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
		ExpiresAt:                      expiresAt.UTC(),
	}), nil
}

func (a *App) markPostgresManagedTargetPlanExecuted(planID string) error {
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID != planID {
				continue
			}
			if current.ReversePlans[i].Mode != reversePlanModePostgresManagedTarget {
				return current, fmt.Errorf("PostgreSQL managed target plan %q is not awaiting transport completion", planID)
			}
			if current.ReversePlans[i].Status == "executed" && current.ReversePlans[i].AuthorizationReference != "" {
				return current, nil
			}
			if current.ReversePlans[i].Status != reversePlanStatusApprovalConsumptionUncertain {
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
	return a.validatePostgresManagedTargetPlanWithSeal(plan, runtime, binding, true)
}

// validatePostgresManagedTargetPlanWithSeal validates the sealed initial
// approval shape. A pre-token plan must have a current seal. A plan that has
// already minted a durable authorization is instead bound by the authorization
// record's exact scope identity, including this plan binding, and may outlive
// the intentionally short preview seal.
func (a *App) validatePostgresManagedTargetPlanWithSeal(plan ReversePlan, runtime connectors.RuntimeConfig, binding postgresManagedTargetApprovalBinding, requireCurrentSeal bool) (connectors.WritePreview, error) {
	bindingSHA256, err := hashJSON(binding)
	if err != nil {
		return connectors.WritePreview{}, err
	}
	if plan.TransportBindingSHA256 != bindingSHA256 || plan.DestinationConnector != binding.Destination || plan.Action != postgresManagedTargetAction(binding.Mode) {
		return connectors.WritePreview{}, fmt.Errorf("%w: plan no longer matches the connection", ErrPostgresManagedTargetApprovalStale)
	}
	authorizationLifetime, err := normalizePostgresManagedTargetAuthorizationLifetime(plan.AuthorizationLifetime)
	if err != nil {
		return connectors.WritePreview{}, fmt.Errorf("%w: authorization lifetime", ErrPostgresManagedTargetApprovalStale)
	}
	planHash, err := postgresManagedTargetPlanHash(bindingSHA256, authorizationLifetime)
	if err != nil || plan.PlanHash != planHash {
		return connectors.WritePreview{}, fmt.Errorf("%w: plan hash changed", ErrPostgresManagedTargetApprovalStale)
	}
	if requireCurrentSeal {
		if plan.PlanSeal == nil {
			return connectors.WritePreview{}, fmt.Errorf("%w: plan seal is missing", ErrPostgresManagedTargetApprovalStale)
		}
		if err := a.verifyPlanSealForRuntime(plan, runtime); err != nil {
			return connectors.WritePreview{}, err
		}
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
	if err := a.ensureTransportRegistry(); err != nil {
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
	copyWorkerMaximum := targetCopyWorkerMaximum(destination)
	if conn.TargetCopyWorkers > 0 && copyWorkerMaximum == 0 {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, &TargetCopyWorkersUnsupportedError{Destination: destination.Name()}
	}
	if conn.TargetCopyWorkers > copyWorkerMaximum {
		return Connection{}, StreamConfig{}, SyncMode{}, connectors.RuntimeConfig{}, postgresManagedTargetApprovalBinding{}, &TargetCopyWorkersRangeError{Requested: conn.TargetCopyWorkers, Maximum: copyWorkerMaximum}
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
		TransformPlanHash: stream.TransformPlanHash, ConfigurationDigest: runtime.ConfigurationDigest, ApprovalScope: runtime.WriteApprovalScope,
		TargetCopyWorkers: conn.TargetCopyWorkers, TargetCopyWorkerMaximum: copyWorkerMaximum,
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
