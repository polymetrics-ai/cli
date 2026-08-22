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
	reversePlanModeDeclarativeTypedDestinationTransport = "declarative_typed_destination_transport"
	declarativeTypedDestinationBindingDomain            = "declarative_typed_destination_transport/v1"
)

// declarativeTypedDestinationBinding is the sealed, payload-free identity of
// one persisted connection stream and its definition-selected typed action.
// It intentionally carries no URL, request body, arbitrary action, or source
// record. Those stay in the bundle and the reopened warehouse workset.
type declarativeTypedDestinationBinding struct {
	Domain                          string                                  `json:"domain"`
	ConnectionID                    string                                  `json:"connection_id"`
	Stream                          string                                  `json:"stream"`
	StreamID                        string                                  `json:"stream_id"`
	Mode                            synccontract.Mode                       `json:"mode"`
	Source                          string                                  `json:"source"`
	SourceExecutor                  connectors.TransportExecutorReference   `json:"source_executor"`
	SourceEvidence                  connectors.ConformanceEvidenceReference `json:"source_evidence"`
	Destination                     string                                  `json:"destination"`
	DestinationExecutor             connectors.TransportExecutorReference   `json:"destination_executor"`
	DestinationEvidence             connectors.ConformanceEvidenceReference `json:"destination_evidence"`
	Action                          string                                  `json:"action"`
	ActionDefinitionSHA256          string                                  `json:"action_definition_sha256"`
	IdempotencyKeyHeader            string                                  `json:"idempotency_key_header"`
	Strategy                        connectors.ApplyStrategy                `json:"strategy"`
	SourceMapping                   connectors.SourceRecordMapping          `json:"source_mapping"`
	Batch                           connectors.DestinationBatch             `json:"batch"`
	TombstoneAction                 string                                  `json:"tombstone_action,omitempty"`
	TombstoneActionDefinitionSHA256 string                                  `json:"tombstone_action_definition_sha256,omitempty"`
	TombstoneIdempotencyKeyHeader   string                                  `json:"tombstone_idempotency_key_header,omitempty"`
	TombstoneSourceMapping          *connectors.TombstoneRecordMapping      `json:"tombstone_source_mapping,omitempty"`
	TombstoneBatch                  *connectors.DestinationBatch            `json:"tombstone_batch,omitempty"`
	ReadBackSHA256                  string                                  `json:"read_back_sha256"`
	CredentialRevision              string                                  `json:"credential_revision"`
	ConfigurationDigest             string                                  `json:"configuration_digest"`
	ApprovalScope                   string                                  `json:"approval_scope"`
}

type preparedDeclarativeTypedDestinationTransport struct {
	connection  Connection
	stream      StreamConfig
	mode        SyncMode
	source      connectors.Connector
	sourceRun   connectors.RuntimeConfig
	destination connectors.Connector
	credential  CredentialMeta
	runtime     connectors.RuntimeConfig
	resolved    synctransport.ResolvedTransport
	contract    declarativeTypedDestinationContract
	binding     declarativeTypedDestinationBinding
	target      connectors.WriteApprovalTarget
}

// PlanDeclarativeTypedDestinationTransport creates the plan for exactly one
// persisted declarative typed destination action. The caller chooses only a
// connection and stream; the definition and stream identity choose the action.
func (a *App) PlanDeclarativeTypedDestinationTransport(ctx context.Context, connectionName, streamName string) (ReversePlan, error) {
	prepared, err := a.prepareDeclarativeTypedDestinationTransport(ctx, connectionName, streamName)
	if err != nil {
		return ReversePlan{}, err
	}
	bindingSHA256, err := hashJSON(prepared.binding)
	if err != nil {
		return ReversePlan{}, fmt.Errorf("hash declarative typed destination binding: %w", err)
	}
	planHash, err := hashJSON(struct {
		Domain  string `json:"domain"`
		Binding string `json:"binding"`
	}{declarativeTypedDestinationBindingDomain, bindingSHA256})
	if err != nil {
		return ReversePlan{}, err
	}
	id, err := prefixedID("rplan")
	if err != nil {
		return ReversePlan{}, err
	}
	confirmation := connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
	seal, err := a.approval.IssueWritePlanSeal(connectors.WritePlanSealRequest{
		PlanID: id, PlanHash: planHash, Mode: reversePlanModeDeclarativeTypedDestinationTransport,
		Connector: prepared.destination.Name(), Operation: prepared.resolved.ApplyStrategy.Action,
		CredentialRevision: prepared.runtime.CredentialRevision, ConfigurationDigest: prepared.runtime.ConfigurationDigest,
		Batchable: a.actionIsBatchable(prepared.destination.Name(), prepared.resolved.ApplyStrategy.Action), Scope: prepared.runtime.WriteApprovalScope,
		Confirmation: confirmation,
	})
	if err != nil {
		return ReversePlan{}, err
	}
	plan := ReversePlan{
		ID: id, Name: "declarative typed destination transport", Status: "planned",
		Mode: reversePlanModeDeclarativeTypedDestinationTransport, SourceConnection: prepared.connection.Name,
		DestinationConnector: prepared.destination.Name(), DestinationCredential: prepared.connection.Destination.Credential,
		DestinationConfig: cloneStringMap(prepared.connection.Destination.Config), Action: prepared.resolved.ApplyStrategy.Action,
		Mappings: map[string]string{}, ConfirmationChallenge: string(confirmation.Kind), ConfirmationPolicy: confirmation,
		RecordCount: 0, PlanHash: planHash, PlanSeal: &seal, TransportConnectionID: prepared.connection.ID,
		TransportStream: streamName, TransportBindingSHA256: bindingSHA256, TransportActionDefinitionSHA256: prepared.binding.ActionDefinitionSHA256, CreatedAt: seal.IssuedAt, ExpiresAt: seal.ExpiresAt,
	}
	a.state.ReversePlans = append(a.state.ReversePlans, plan)
	if err := a.save(); err != nil {
		return ReversePlan{}, err
	}
	return plan, nil
}

// PreviewDeclarativeTypedDestinationTransport produces a payload-free,
// definition-bound preview. The first provider request is the post-approval
// typed action over a reopened warehouse workset, so preview cannot become a
// second generic provider-write surface.
func (a *App) PreviewDeclarativeTypedDestinationTransport(ctx context.Context, planID string) (ReversePlan, connectors.WritePreview, error) {
	plan, err := a.GetReversePlan(planID)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if plan.Mode != reversePlanModeDeclarativeTypedDestinationTransport {
		return ReversePlan{}, connectors.WritePreview{}, fmt.Errorf("reverse plan %q is not a declarative typed destination transport plan", plan.ID)
	}
	if err := a.previewabilityError(plan, time.Now().UTC()); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	prepared, err := a.prepareDeclarativeTypedDestinationTransport(ctx, plan.SourceConnection, plan.TransportStream)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.validateDeclarativeTypedDestinationPlan(plan, prepared); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	preview := connectors.WritePreview{RecordsStaged: 0, Action: plan.Action, Digest: plan.TransportBindingSHA256, ApprovalTarget: prepared.target}
	issued, err := a.persistDestructivePreview(plan, preview)
	if err != nil {
		return ReversePlan{}, connectors.WritePreview{}, err
	}
	return issued, preview, nil
}

func (a *App) authorizeDeclarativeTypedDestinationTransport(ctx context.Context, conn Connection, streamName string, approval synctransport.DestinationApproval, runtime connectors.RuntimeConfig) (synctransport.DestinationApproval, error) {
	if strings.TrimSpace(approval.PlanID) == "" {
		return synctransport.DestinationApproval{}, fmt.Errorf("declarative typed destination transport requires a previewed approval plan")
	}
	plan, err := a.GetReversePlan(approval.PlanID)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	prepared, err := a.prepareDeclarativeTypedDestinationTransport(ctx, conn.Name, streamName)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if !declarativeTypedDestinationRuntimeEqual(runtime, prepared.runtime) {
		return synctransport.DestinationApproval{}, fmt.Errorf("declarative typed destination runtime does not match the connection-owned approval identity")
	}
	if err := a.validateDeclarativeTypedDestinationPlan(plan, prepared); err != nil {
		return synctransport.DestinationApproval{}, err
	}
	if err := a.verifyPlanSealForRuntime(plan, prepared.runtime); err != nil {
		return synctransport.DestinationApproval{}, err
	}
	preview := connectors.WritePreview{RecordsStaged: 0, Action: plan.Action, Digest: plan.TransportBindingSHA256, ApprovalTarget: prepared.target}
	scope := a.declarativeTypedDestinationAuthorizationScope(prepared, plan)
	authorizationReference := plan.AuthorizationReference
	if authorizationReference != "" {
		if strings.TrimSpace(approval.ApprovalToken) != "" {
			return synctransport.DestinationApproval{}, &AuthorizationTokenReplayError{Reference: authorizationReference}
		}
		if _, err := a.requireAuthorization(authorizationReference, scope, time.Now().UTC()); err != nil {
			return synctransport.DestinationApproval{}, err
		}
	} else {
		if strings.TrimSpace(approval.ApprovalToken) == "" {
			return synctransport.DestinationApproval{}, fmt.Errorf("declarative typed destination transport requires an stdin approval token")
		}
		authorization, err := newAuthorizationRecord(scope, time.Now().UTC())
		if err != nil {
			return synctransport.DestinationApproval{}, err
		}
		if _, _, err := a.consumePlanApproval(plan, RunReverseETLRequest{
			PlanID: plan.ID, ApprovalToken: approval.ApprovalToken, Confirmation: approval.Confirmation,
		}, preview, &authorization); err != nil {
			return synctransport.DestinationApproval{}, err
		}
		authorizationReference = authorization.Reference
	}
	evidence, err := durableAuthorizationEvidence(scope)
	if err != nil {
		return synctransport.DestinationApproval{}, err
	}
	return synctransport.DestinationApproval{
		PlanID: plan.ID, Evidence: evidence, Target: prepared.target, PreviewDigest: preview.Digest, ActionDefinitionSHA256: prepared.binding.ActionDefinitionSHA256,
		TombstoneActionDefinitionSHA256: prepared.binding.TombstoneActionDefinitionSHA256,
		IdempotencyProof: synctransport.DestinationIdempotencyProof{
			Executor:               prepared.binding.DestinationExecutor,
			ActionDefinitionSHA256: prepared.binding.ActionDefinitionSHA256,
			EffectiveHeader:        prepared.binding.IdempotencyKeyHeader,
		},
		TombstoneIdempotencyProof: synctransport.DestinationIdempotencyProof{
			Executor:               prepared.binding.DestinationExecutor,
			ActionDefinitionSHA256: prepared.binding.TombstoneActionDefinitionSHA256,
			EffectiveHeader:        prepared.binding.TombstoneIdempotencyKeyHeader,
		},
		AuthorizeNextUnit: func(unitCtx context.Context) error {
			if unitCtx == nil {
				return fmt.Errorf("declarative typed destination authorization context is required")
			}
			if err := unitCtx.Err(); err != nil {
				return err
			}
			_, err := a.requireAuthorization(authorizationReference, scope, time.Now().UTC())
			return err
		},
		IssueWriteEvidence: func(unitCtx context.Context) (*connectors.WriteApprovalEvidence, error) {
			if unitCtx == nil {
				return nil, fmt.Errorf("declarative typed destination write authorization context is required")
			}
			if err := unitCtx.Err(); err != nil {
				return nil, err
			}
			if _, err := a.requireAuthorization(authorizationReference, scope, time.Now().UTC()); err != nil {
				return nil, err
			}
			return durableAuthorizationEvidence(scope)
		},
	}, nil
}

// declarativeTypedDestinationAuthorizationScope is the revocable, payload-free
// authority behind one persisted stream/action declaration. The bundle owns
// the endpoint, verb, body, mapping and action; this scope records only their
// stable binding evidence, so no runtime caller can broaden the provider
// operation while reusing an approval.
func (a *App) declarativeTypedDestinationAuthorizationScope(prepared preparedDeclarativeTypedDestinationTransport, plan ReversePlan) AuthorizationScope {
	fieldMappings := map[string]string{
		"transport_binding_sha256": plan.TransportBindingSHA256,
		"action_definition_sha256": prepared.binding.ActionDefinitionSHA256,
		"idempotency_key_header":   prepared.binding.IdempotencyKeyHeader,
		"source_connector":         prepared.binding.Source,
		"source_executor":          string(prepared.binding.SourceExecutor.Family) + ":" + prepared.binding.SourceExecutor.ID,
		"destination_executor":     string(prepared.binding.DestinationExecutor.Family) + ":" + prepared.binding.DestinationExecutor.ID,
		"apply_strategy":           string(prepared.binding.Strategy),
	}
	for _, input := range prepared.binding.SourceMapping.Inputs {
		fieldMappings["input/"+input.Input] = input.Field
	}
	allowedWriteActions := []string{prepared.resolved.ApplyStrategy.Action}
	enabledOperations := []string{string(prepared.binding.Mode), prepared.resolved.ApplyStrategy.Action, string(prepared.binding.Strategy)}
	if prepared.binding.TombstoneSourceMapping != nil {
		fieldMappings["tombstone_action_definition_sha256"] = prepared.binding.TombstoneActionDefinitionSHA256
		fieldMappings["tombstone_idempotency_key_header"] = prepared.binding.TombstoneIdempotencyKeyHeader
		for _, input := range prepared.binding.TombstoneSourceMapping.Inputs {
			fieldMappings["tombstone_input/"+input.Input] = input.Field
		}
		allowedWriteActions = append(allowedWriteActions, prepared.binding.TombstoneAction)
		enabledOperations = append(enabledOperations, prepared.binding.TombstoneAction)
	}
	return canonicalAuthorizationScope(AuthorizationScope{
		SourceConnection:              prepared.connection.ID,
		DestinationConnection:         prepared.credential.ID,
		DestinationCredentialRevision: prepared.runtime.CredentialRevision,
		StreamTables: []AuthorizationStreamTable{{
			Stream: prepared.binding.Stream, SourceTable: prepared.binding.Stream, DestinationTable: "sync_transport",
		}},
		FieldMappings:                  fieldMappings,
		WriteAction:                    prepared.resolved.ApplyStrategy.Action,
		AllowedWriteActions:            allowedWriteActions,
		DestinationConfigurationDigest: prepared.runtime.ConfigurationDigest,
		EnabledOperations:              enabledOperations,
		ConfirmationPolicy:             plan.ConfirmationPolicy,
		ExpiresAt:                      plan.ExpiresAt,
	})
}

func (a *App) prepareDeclarativeTypedDestinationTransport(ctx context.Context, connectionName, streamName string) (preparedDeclarativeTypedDestinationTransport, error) {
	conn, ok := a.findConnection(connectionName)
	if !ok {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("connection %q not found", connectionName)
	}
	stream, ok := conn.Streams[streamName]
	if !ok {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("stream %q not configured on connection %q", streamName, connectionName)
	}
	mode, err := ParseStreamSyncMode(stream)
	if err != nil || mode.ContractMode == "" {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("declarative typed destination requires a contract sync mode")
	}
	source, _, sourceRuntime, err := a.resolveEndpointWithCredential(ctx, conn.Source)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	destination, credential, runtime, err := a.resolveEndpointWithCredential(ctx, conn.Destination)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	if a.transports == nil {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("declarative typed destination transport registry is unavailable")
	}
	resolved, err := a.transports.Preflight(synctransport.PreflightRequest{Source: source, Destination: destination, Stream: streamName, Mode: mode.ContractMode, DestinationAction: stream.DestinationAction})
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	if _, ok := resolved.Destination.(synctransport.DefinitionOwnedApprovalDestination); !ok || resolved.Destination.TransportExecutorReference() != declarativeTypedDestinationReference {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("connection %q does not select the declarative typed destination adapter", conn.Name)
	}
	contract, err := declarativeTypedDestinationContractFor(destination)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	binding, err := contract.plan(source, streamName, mode.ContractMode, resolved.ApplyStrategy)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	action, found := contract.actions[resolved.ApplyStrategy.Action]
	if !found {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("declarative typed destination action %q is unavailable", resolved.ApplyStrategy.Action)
	}
	actionDefinitionSHA256, err := contract.actionDefinitionDigest(action.Name)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	idempotencyKeyHeader, err := contract.idempotencyHeader(action.Name)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	var tombstoneBinding *connectors.DestinationSourceBinding
	var tombstoneActionDefinitionSHA256, tombstoneIdempotencyKeyHeader string
	if resolved.ApplyStrategy.TombstoneAction != "" {
		resolvedTombstoneBinding, bindingErr := contract.tombstoneBinding(source, streamName, mode.ContractMode, resolved.ApplyStrategy)
		if bindingErr != nil {
			return preparedDeclarativeTypedDestinationTransport{}, bindingErr
		}
		tombstoneBinding = &resolvedTombstoneBinding
		tombstoneActionDefinitionSHA256, err = contract.actionDefinitionDigest(resolved.ApplyStrategy.TombstoneAction)
		if err != nil {
			return preparedDeclarativeTypedDestinationTransport{}, err
		}
		tombstoneIdempotencyKeyHeader, err = contract.idempotencyHeader(resolved.ApplyStrategy.TombstoneAction)
		if err != nil {
			return preparedDeclarativeTypedDestinationTransport{}, err
		}
	}
	readBackSHA256, err := hashJSON(struct {
		ReadBack          *connectors.DestinationReadBackPolicy          `json:"read_back"`
		TombstoneReadBack *connectors.DestinationTombstoneReadBackPolicy `json:"tombstone_read_back,omitempty"`
	}{ReadBack: contract.descriptor.ReadBack, TombstoneReadBack: contract.descriptor.TombstoneReadBack})
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, fmt.Errorf("hash declarative typed destination read-back declaration: %w", err)
	}
	declaration := declarativeTypedDestinationBinding{
		Domain: declarativeTypedDestinationBindingDomain, ConnectionID: conn.ID, Stream: streamName, StreamID: stream.StreamID, Mode: mode.ContractMode,
		Source: source.Name(), SourceExecutor: resolved.SourceDescriptor.Executor, SourceEvidence: resolved.SourceDescriptor.Conformance,
		Destination: destination.Name(), DestinationExecutor: resolved.DestinationDescriptor.Executor, DestinationEvidence: resolved.DestinationDescriptor.Conformance,
		Action: resolved.ApplyStrategy.Action, ActionDefinitionSHA256: actionDefinitionSHA256, IdempotencyKeyHeader: idempotencyKeyHeader, Strategy: resolved.ApplyStrategy.Strategy, SourceMapping: binding.RecordMapping.Clone(), Batch: *binding.Batch, ReadBackSHA256: readBackSHA256,
		CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest, ApprovalScope: runtime.WriteApprovalScope,
	}
	if tombstoneBinding != nil {
		mapping := tombstoneBinding.TombstoneMapping.Clone()
		batch := *tombstoneBinding.Batch
		declaration.TombstoneAction = resolved.ApplyStrategy.TombstoneAction
		declaration.TombstoneActionDefinitionSHA256 = tombstoneActionDefinitionSHA256
		declaration.TombstoneIdempotencyKeyHeader = tombstoneIdempotencyKeyHeader
		declaration.TombstoneSourceMapping = &mapping
		declaration.TombstoneBatch = &batch
	}
	bindingSHA256, err := hashJSON(declaration)
	if err != nil {
		return preparedDeclarativeTypedDestinationTransport{}, err
	}
	target := connectors.WriteApprovalTarget{Connector: destination.Name(), Operation: action.Name, Method: action.Method, MutationClass: "transport", TargetDigest: bindingSHA256, CredentialRevision: runtime.CredentialRevision, ConfigurationDigest: runtime.ConfigurationDigest, Batchable: a.actionIsBatchable(destination.Name(), action.Name), Scope: runtime.WriteApprovalScope, Confirmation: connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}}
	return preparedDeclarativeTypedDestinationTransport{connection: conn, stream: stream, mode: mode, source: source, sourceRun: sourceRuntime, destination: destination, credential: credential, runtime: runtime, resolved: resolved, contract: contract, binding: declaration, target: target}, nil
}

func (a *App) validateDeclarativeTypedDestinationPlan(plan ReversePlan, prepared preparedDeclarativeTypedDestinationTransport) error {
	bindingSHA256, err := hashJSON(prepared.binding)
	if err != nil {
		return err
	}
	planHash, err := hashJSON(struct {
		Domain  string `json:"domain"`
		Binding string `json:"binding"`
	}{declarativeTypedDestinationBindingDomain, bindingSHA256})
	if err != nil {
		return err
	}
	if plan.Mode != reversePlanModeDeclarativeTypedDestinationTransport || plan.TransportConnectionID != prepared.connection.ID || plan.TransportStream != prepared.binding.Stream || plan.SourceConnection != prepared.connection.Name || plan.DestinationConnector != prepared.destination.Name() || plan.DestinationCredential != prepared.connection.Destination.Credential || plan.Action != prepared.resolved.ApplyStrategy.Action || plan.RecordCount != 0 || len(plan.Mappings) != 0 || validateDeclarativeTypedDestinationPlanBinding(plan, bindingSHA256) != nil || !constantTimeStringEqual(plan.TransportActionDefinitionSHA256, prepared.binding.ActionDefinitionSHA256) || !constantTimeStringEqual(plan.PlanHash, planHash) {
		return fmt.Errorf("declarative typed destination approval plan does not bind the exact persisted connection action")
	}
	if plan.ConfirmationPolicy.Kind != connectors.ConfirmationKindDestructive || a.confirmationPolicyForPlan(plan).Kind != connectors.ConfirmationKindDestructive {
		return fmt.Errorf("declarative typed destination approval plan does not require destructive confirmation")
	}
	return nil
}

func validateDeclarativeTypedDestinationPlanBinding(plan ReversePlan, bindingSHA256 string) error {
	if !constantTimeStringEqual(plan.TransportBindingSHA256, bindingSHA256) {
		return fmt.Errorf("declarative typed destination approval plan does not bind the exact persisted connection action")
	}
	return nil
}

func declarativeTypedDestinationRuntimeEqual(left, right connectors.RuntimeConfig) bool {
	return left.CredentialRevision == right.CredentialRevision && left.ConfigurationDigest == right.ConfigurationDigest && left.WriteApprovalScope == right.WriteApprovalScope
}

func (a *App) markDeclarativeTypedDestinationPlanExecuted(planID string) error {
	updated, err := a.updateState(func(current state) (state, error) {
		for i := range current.ReversePlans {
			if current.ReversePlans[i].ID != planID {
				continue
			}
			if current.ReversePlans[i].Status == "executed" {
				return current, nil
			}
			if current.ReversePlans[i].Status != reversePlanStatusApprovalConsumptionUncertain {
				return current, fmt.Errorf("declarative typed destination plan %q is not awaiting write completion", planID)
			}
			current.ReversePlans[i].Status = "executed"
			current.ReversePlans[i].ApprovalUncertainAt = time.Time{}
			return current, nil
		}
		return current, fmt.Errorf("declarative typed destination plan %q not found", planID)
	})
	if err != nil {
		return err
	}
	a.state = updated
	return nil
}
