package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

// declarativeSingleAttemptDestinationExecutor is an opt-in, declaration-owned
// destination for a bounded workset whose provider action has no source-backed
// retry or read-back guarantee. It never receives a caller-authored action,
// route, body, or retry policy.
type declarativeSingleAttemptDestinationExecutor struct{}

type declarativeSingleAttemptDestinationContract struct {
	connector  connectors.DeclarativeTypedDestination
	descriptor connectors.DestinationTransportDescriptor
	actions    map[string]connectors.WriteActionInfo
}

func (*declarativeSingleAttemptDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return declarativeSingleAttemptDestinationReference
}

func (*declarativeSingleAttemptDestinationExecutor) DefinitionOwnedApprovalDestination() {}

func (e *declarativeSingleAttemptDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if e == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("declarative single-attempt destination is unavailable")
	}
	if err := validateDeclarativeTypedDestinationApproval(request.Approval); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	contract, err := declarativeSingleAttemptDestinationContractFor(request.Connector)
	if err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if _, err := contract.plan(request.Source, request.Stream, request.Mode, request.ApplyStrategy); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	actionDefinitionSHA256, err := contract.actionDefinitionDigest(request.ApplyStrategy.Action)
	if err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if err := validateDeclarativeTypedDestinationApprovalDefinition(request.Approval, actionDefinitionSHA256); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	physicalActions := declarativeTypedDestinationPhysicalActions(declarativeTypedDestinationBinding{
		Action: request.ApplyStrategy.Action, ActionDefinitionSHA256: actionDefinitionSHA256,
	})
	if err := validateDeclarativeTypedDestinationApprovedPhysicalActions(request.Approval, physicalActions); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy, ActionDefinitionSHA256: actionDefinitionSHA256, PhysicalActions: physicalActions}, nil
}

func (e *declarativeSingleAttemptDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if e == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative single-attempt destination is unavailable")
	}
	if err := validateDeclarativeTypedDestinationApproval(request.Approval); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	contract, err := declarativeSingleAttemptDestinationContractFor(request.Destination)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	binding, err := contract.plan(request.Source, request.Stream, request.Mode, request.Plan.ApplyStrategy)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	actionDefinitionSHA256, err := contract.actionDefinitionDigest(request.Plan.ApplyStrategy.Action)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if request.Plan.ActionDefinitionSHA256 != actionDefinitionSHA256 {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative single-attempt action %q definition changed; replan and reapprove", request.Plan.ApplyStrategy.Action)
	}
	if err := validateDeclarativeTypedDestinationApprovalDefinition(request.Approval, actionDefinitionSHA256); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	physicalActions := declarativeTypedDestinationPhysicalActions(declarativeTypedDestinationBinding{
		Action: request.Plan.ApplyStrategy.Action, ActionDefinitionSHA256: actionDefinitionSHA256,
	})
	if err := validateDeclarativeTypedDestinationPhysicalActions(request.Plan.PhysicalActions, physicalActions); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative single-attempt plan %w", err)
	}
	if err := validateDeclarativeTypedDestinationApprovedPhysicalActions(request.Approval, physicalActions); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := validateDeclarativeSingleAttemptWorkset(request, binding); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	records, err := declarativeTypedDestinationRecords(request.Workset.Records, binding)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	result, writeErr := declarativeSingleAttemptWrite(ctx, contract, request, request.Plan.ApplyStrategy.Action, records)
	results := []declarativeTypedDestinationActionResult{{Action: request.Plan.ApplyStrategy.Action, Result: result}}
	output, outputErr := declarativeTypedDestinationApplyOutput(results, request.Runtime.Secrets)
	if outputErr != nil {
		return synccontract.DownstreamAcknowledgement{}, declarativeTypedDestinationApplyFailure(outputErr, results, request.Runtime.Secrets)
	}
	if writeErr != nil {
		return synccontract.DownstreamAcknowledgement{}, synctransport.NewDestinationApplyOutputError(writeErr, output)
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(contract.connector.Name(), time.Now().UTC())
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, declarativeTypedDestinationApplyFailure(err, results, request.Runtime.Secrets)
	}
	acknowledgement, err = acknowledgement.WithOutput(output)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, declarativeTypedDestinationApplyFailure(fmt.Errorf("attach declarative single-attempt action %q output: %w", request.Plan.ApplyStrategy.Action, err), results, request.Runtime.Secrets)
	}
	return acknowledgement, nil
}

// ReadBackDestination deliberately performs no provider request. A
// single-attempt declaration is admitted only when its response capture is the
// stated acknowledgement boundary; retry-safe destinations use the separate
// declarative_typed_destination adapter and its provider read-back contract.
func (e *declarativeSingleAttemptDestinationExecutor) ReadBackDestination(_ context.Context, request synctransport.DestinationReadBackRequest) error {
	if e == nil {
		return fmt.Errorf("declarative single-attempt destination is unavailable")
	}
	contract, err := declarativeSingleAttemptDestinationContractFor(request.Destination)
	if err != nil {
		return err
	}
	binding, err := contract.plan(request.Source, request.Stream, request.Mode, request.Plan.ApplyStrategy)
	if err != nil {
		return err
	}
	actionDefinitionSHA256, err := contract.actionDefinitionDigest(request.Plan.ApplyStrategy.Action)
	if err != nil {
		return err
	}
	if request.Plan.ActionDefinitionSHA256 != actionDefinitionSHA256 {
		return fmt.Errorf("declarative single-attempt action %q definition changed; replan and reapprove", request.Plan.ApplyStrategy.Action)
	}
	if request.Workset.ID == "" || len(request.Workset.Records) == 0 || len(request.Workset.Tombstones) != 0 || len(request.Workset.Records) > binding.Batch.MaxRecords {
		return fmt.Errorf("declarative single-attempt destination requires its exact reopened record workset")
	}
	if request.Acknowledgement.Sink != contract.connector.Name() || request.Acknowledgement.AcknowledgedAt.IsZero() {
		return fmt.Errorf("declarative single-attempt destination requires its captured acknowledgement")
	}
	if len(request.Acknowledgement.Output) == 0 {
		return fmt.Errorf("declarative single-attempt destination acknowledgement has no captured provider response")
	}
	return nil
}

func declarativeSingleAttemptDestinationContractFor(connector connectors.Connector) (declarativeSingleAttemptDestinationContract, error) {
	candidate, ok := connector.(connectors.DeclarativeTypedDestination)
	if !ok || declarativeTypedDestinationIsNil(candidate) {
		return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires a registered typed write capability")
	}
	definition, defined := connectors.DefinitionOf(candidate)
	if !defined || definition.SyncTransport == nil || definition.SyncTransport.Destination == nil {
		return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires a destination declaration")
	}
	descriptor := *definition.SyncTransport.Destination
	if descriptor.Executor != declarativeSingleAttemptDestinationReference {
		return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires its exact executor")
	}
	if err := descriptor.Validate(); err != nil {
		return declarativeSingleAttemptDestinationContract{}, err
	}
	if descriptor.Acknowledgement != connectors.TransportAcknowledgementDurableWarehouse {
		return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires durable_warehouse acknowledgement")
	}
	if descriptor.Delivery.Idempotency != connectors.DeliveryIdempotencySingleAttempt || descriptor.Delivery.Deletes != connectors.DeliveryDeletesUnavailable {
		return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires single_attempt delivery and declared unavailable tombstones")
	}
	if len(descriptor.SourceBindings) == 0 || descriptor.ReadBack != nil || descriptor.TombstoneReadBack != nil {
		return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires exact source bindings without provider read-back")
	}
	actions := make(map[string]connectors.WriteActionInfo, len(definition.WriteActions))
	for _, action := range definition.WriteActions {
		if _, duplicate := actions[action.Name]; duplicate {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination duplicates write action %q", action.Name)
		}
		actions[action.Name] = action
	}
	recordActions := make(map[string]struct{}, len(descriptor.SourceBindings))
	for _, binding := range descriptor.SourceBindings {
		if binding.Action == "" || binding.Batch == nil || binding.TombstoneMapping != nil {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires one action-owned record binding")
		}
		if err := binding.Batch.Validate(); err != nil {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt action %q batch disposition: %w", binding.Action, err)
		}
		if binding.RecordMapping.Kind != connectors.SourceRecordMappingKindInputFields {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination requires input_fields source mapping")
		}
		recordActions[binding.Action] = struct{}{}
	}
	for _, strategy := range descriptor.ApplyStrategies {
		if strategy.Mode != synccontract.ModeFullAppend || strategy.Strategy != connectors.ApplyStrategyAppend || strategy.TombstoneAction != "" || strategy.ReadBack != nil || strategy.TombstoneReadBack != nil {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt destination supports only full_append without tombstone or read-back actions")
		}
		if _, bound := recordActions[strategy.Action]; !bound {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt action %q has no action-owned source binding", strategy.Action)
		}
		action, found := actions[strategy.Action]
		if !found || action.Name == "" || strings.TrimSpace(action.Method) == "" || strings.TrimSpace(action.Path) == "" || action.TransportBinding != nil {
			return declarativeSingleAttemptDestinationContract{}, fmt.Errorf("declarative single-attempt strategy %q names unavailable typed action %q", strategy.Mode, strategy.Action)
		}
	}
	return declarativeSingleAttemptDestinationContract{connector: candidate, descriptor: descriptor, actions: actions}, nil
}

func (c declarativeSingleAttemptDestinationContract) plan(source connectors.Connector, stream string, mode synccontract.Mode, strategy connectors.DestinationApplyStrategy) (connectors.DestinationSourceBinding, error) {
	if mode != synccontract.ModeFullAppend {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative single-attempt destination does not implement %s", mode)
	}
	expected, err := c.descriptor.ApplyStrategyForAction(mode, strategy.Action)
	if err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	if expected != strategy || strategy.TombstoneAction != "" {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative single-attempt action %q is not the declared strategy for mode %q", strategy.Action, mode)
	}
	if _, found := c.actions[strategy.Action]; !found {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative single-attempt action %q is unavailable", strategy.Action)
	}
	sourceDescriptor, declared := connectors.SourceTransportDescriptorOf(source)
	if !declared {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative single-attempt source has no transport declaration")
	}
	binding, admitted := c.descriptor.SourceBindingForAction(sourceDescriptor.Executor, stream, strategy.Action)
	if !admitted || binding.Action != strategy.Action || binding.Batch == nil || binding.RecordMapping.Kind != connectors.SourceRecordMappingKindInputFields {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative single-attempt action %q has no exact input_fields source binding", strategy.Action)
	}
	inputs := make([]string, len(binding.RecordMapping.Inputs))
	for index, input := range binding.RecordMapping.Inputs {
		inputs[index] = input.Input
	}
	if err := c.connector.PreflightWriteRecordFieldMapping(strategy.Action, inputs); err != nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative single-attempt action %q source inputs are not an exact complete record schema mapping: %w", strategy.Action, err)
	}
	return binding, nil
}

func (c declarativeSingleAttemptDestinationContract) actionDefinitionDigest(action string) (string, error) {
	digest, err := c.connector.DeclarativeTypedDestinationActionDigest(action)
	if err != nil {
		return "", fmt.Errorf("hash declarative single-attempt action %q: %w", action, err)
	}
	if strings.TrimSpace(digest) == "" {
		return "", fmt.Errorf("declarative single-attempt action %q has no definition digest", action)
	}
	return digest, nil
}

func validateDeclarativeSingleAttemptWorkset(request synctransport.DestinationApplyRequest, binding connectors.DestinationSourceBinding) error {
	if err := request.Receipt.Validate(); err != nil {
		return fmt.Errorf("declarative single-attempt receipt: %w", err)
	}
	if request.ConnectionID == "" || request.Receipt.Owner != request.ConnectionID || request.Receipt.ID != request.Workset.ID {
		return fmt.Errorf("declarative single-attempt receipt does not bind the reopened workset")
	}
	if len(request.Workset.Records) == 0 || len(request.Workset.Tombstones) != 0 || request.Receipt.Records != len(request.Workset.Records) || request.Receipt.Tombstones != 0 {
		return fmt.Errorf("declarative single-attempt destination requires a non-empty record-only reopened workset")
	}
	if len(request.Workset.Records) > binding.Batch.MaxRecords {
		return fmt.Errorf("declarative single-attempt action %q workset has %d records, exceeding declaration batch maximum %d", binding.Action, len(request.Workset.Records), binding.Batch.MaxRecords)
	}
	return nil
}

func declarativeSingleAttemptWrite(ctx context.Context, contract declarativeSingleAttemptDestinationContract, request synctransport.DestinationApplyRequest, action string, records []connectors.Record) (connectors.WriteResult, error) {
	writeEvidence := request.Approval.Evidence
	var err error
	if request.Approval.IssueWriteEvidence != nil {
		writeEvidence, err = request.Approval.IssueWriteEvidence(ctx)
		if err != nil {
			return connectors.WriteResult{}, fmt.Errorf("authorize declarative single-attempt action %q: %w", action, err)
		}
	}
	if writeEvidence == nil {
		return connectors.WriteResult{}, fmt.Errorf("declarative single-attempt action %q has no write evidence", action)
	}
	writeRequest := connectors.WriteRequest{
		Stream: request.Stream, Table: "sync_transport", Action: action, Config: request.Runtime, Approval: writeEvidence, DisableRetries: true,
	}
	if err := contract.connector.ValidateWrite(ctx, writeRequest, records); err != nil {
		return connectors.WriteResult{}, fmt.Errorf("validate declarative single-attempt action %q: %w", action, err)
	}
	result, err := contract.connector.Write(transportpolicy.MarkDestructive(ctx), writeRequest, records)
	if err != nil {
		return result, fmt.Errorf("apply declarative single-attempt action %q: %w", action, err)
	}
	completed := result.RecordsWritten
	if declared, found := contract.actions[action]; found && declared.Kind == "delete" {
		completed += result.RecordsUnchanged
	}
	if completed != len(records) || result.RecordsFailed != 0 {
		return result, fmt.Errorf("declarative single-attempt action %q wrote=%d unchanged=%d failed=%d, want %d durable outcomes", action, result.RecordsWritten, result.RecordsUnchanged, result.RecordsFailed, len(records))
	}
	return result, nil
}
