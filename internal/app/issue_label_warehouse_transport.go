package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	declarativeStreamSourceExecutorID          = "declarative_stream_source"
	declarativeTypedDestinationExecutorID      = "declarative_typed_destination"
	issueLabelDestinationExecutorID            = "issue_label_destination"
	issueLabelTransportSourceIssueConfig       = "transport_source_issue_number"
	issueLabelTransportTargetIssueConfig       = "transport_target_issue_number"
	issueLabelTransportLabelConfig             = "transport_label"
	issueLabelTransportSetReplaceConsentConfig = "transport_allow_set_replace"
	issueLabelTransportKeyedConsentConfig      = "transport_allow_keyed"
	issueLabelTransportMaxReadPages            = 1
	issueCollectionTransportMaxRecords         = 1000
	declarativeTransportMaxPagesConfig         = "max_pages"
)

var (
	declarativeStreamSourceReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     declarativeStreamSourceExecutorID,
	}
	declarativeTypedDestinationReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     declarativeTypedDestinationExecutorID,
	}
	issueLabelDestinationReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     issueLabelDestinationExecutorID,
	}
)

// declarativeDestinationAdapters is a closed set of destination adapters.
// Definitions choose an exact reference; the collection below never chooses a
// connector by name. The issue-label executor remains its own typed adapter
// because its contract includes provider-state read-back, while the generic
// typed destination accepts ordinary schema-backed write actions.
var declarativeDestinationAdapters = []struct {
	reference connectors.TransportExecutorReference
	build     func(*App, connectors.Connector) (synctransport.DestinationExecutor, error)
}{
	{
		reference: issueLabelDestinationReference,
		build: func(app *App, connector connectors.Connector) (synctransport.DestinationExecutor, error) {
			if _, _, err := issueLabelTransportConnectorContract(connector); err != nil {
				return nil, err
			}
			return &issueLabelDestinationExecutor{app: app}, nil
		},
	},
	{
		reference: declarativeTypedDestinationReference,
		build: func(_ *App, connector connectors.Connector) (synctransport.DestinationExecutor, error) {
			if _, err := declarativeTypedDestinationContractFor(connector); err != nil {
				return nil, err
			}
			return &declarativeTypedDestinationExecutor{}, nil
		},
	},
}

// definitionTransportDefinitionFactories supplies the reusable declarative
// source and closed typed-action destination adapters from definitions already
// admitted to the registry. Evidence comes only from the role declaration that
// selected the adapter; the loop is intentionally over descriptors and exact
// executor references, never connector names.
func definitionTransportDefinitionFactories(a *App, registry *connectors.Registry) ([]synctransport.DefinitionFactory, error) {
	if a == nil {
		return nil, fmt.Errorf("definition transport factories require an app")
	}
	if registry == nil {
		return nil, fmt.Errorf("definition transport factories require a connector registry")
	}
	var sourceEvidences []connectors.ConformanceEvidenceReference
	destinationEvidences := make(map[connectors.TransportExecutorReference][]connectors.ConformanceEvidenceReference, len(declarativeDestinationAdapters))
	for _, metadata := range registry.List() {
		connector, ok := registry.Get(metadata.Name)
		if !ok {
			return nil, fmt.Errorf("declared connector %q disappeared from registry", metadata.Name)
		}
		descriptor, declared := connectors.SyncTransportDescriptorOf(connector)
		if !declared {
			continue
		}
		if descriptor.Source != nil && descriptor.Source.Executor == declarativeStreamSourceReference {
			sourceEvidences = appendDefinitionTransportEvidence(sourceEvidences, descriptor.Source.Conformance)
		}
		if descriptor.Destination != nil {
			for _, adapter := range declarativeDestinationAdapters {
				if descriptor.Destination.Executor == adapter.reference {
					destinationEvidences[adapter.reference] = appendDefinitionTransportEvidence(destinationEvidences[adapter.reference], descriptor.Destination.Conformance)
				}
			}
		}
	}
	factories := make([]synctransport.DefinitionFactory, 0, 1+len(declarativeDestinationAdapters))
	if len(sourceEvidences) != 0 {
		factories = append(factories, synctransport.DefinitionFactory{
			Reference:               declarativeStreamSourceReference,
			SourceEvidence:          sourceEvidences[0],
			AcceptedSourceEvidences: append([]connectors.ConformanceEvidenceReference(nil), sourceEvidences[1:]...),
			BuildSource: func(connector connectors.Connector) (synctransport.SourceExecutor, error) {
				if _, _, err := declarativeStreamTransportConnector(connector); err != nil {
					return nil, err
				}
				return &declarativeStreamSourceExecutor{}, nil
			},
		})
	}
	for _, adapter := range declarativeDestinationAdapters {
		evidences := destinationEvidences[adapter.reference]
		if len(evidences) == 0 {
			continue
		}
		build := adapter.build
		factories = append(factories, synctransport.DefinitionFactory{
			Reference:                    adapter.reference,
			DestinationEvidence:          evidences[0],
			AcceptedDestinationEvidences: append([]connectors.ConformanceEvidenceReference(nil), evidences[1:]...),
			BuildDestination: func(connector connectors.Connector) (synctransport.DestinationExecutor, error) {
				return build(a, connector)
			},
		})
	}
	return factories, nil
}

func appendDefinitionTransportEvidence(values []connectors.ConformanceEvidenceReference, evidence connectors.ConformanceEvidenceReference) []connectors.ConformanceEvidenceReference {
	for _, value := range values {
		if value == evidence {
			return values
		}
	}
	return append(values, evidence)
}

// declarativeTypedDestinationExecutor applies one named writes.json action
// selected by the destination declaration. It receives neither a method, URL,
// body template, nor an action name from the caller: those stay owned by the
// connector bundle and its source binding.
type declarativeTypedDestinationExecutor struct{}

type declarativeTypedDestinationContract struct {
	connector          connectors.DeclarativeTypedDestination
	readBack           connectors.DeclarativeTypedDestinationReadBack
	descriptor         connectors.DestinationTransportDescriptor
	actions            map[string]connectors.WriteActionInfo
	idempotencyHeaders map[string]string
}

func (*declarativeTypedDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return declarativeTypedDestinationReference
}

func (*declarativeTypedDestinationExecutor) DefinitionOwnedApprovalDestination() {}

func (e *declarativeTypedDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if e == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("declarative typed destination is unavailable")
	}
	if err := validateDeclarativeTypedDestinationApproval(request.Approval); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	contract, err := declarativeTypedDestinationContractFor(request.Connector)
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
	if err := validateDeclarativeTypedDestinationIdempotencyProof(request.Approval, contract.descriptor.Executor, actionDefinitionSHA256, contract.idempotencyHeaders[request.ApplyStrategy.Action]); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	tombstoneActionDefinitionSHA256 := ""
	tombstoneIdempotencyKeyHeader := ""
	if request.ApplyStrategy.TombstoneAction != "" {
		if _, err := contract.tombstoneBinding(request.Source, request.Stream, request.Mode, request.ApplyStrategy); err != nil {
			return synctransport.DestinationPlan{}, err
		}
		tombstoneActionDefinitionSHA256, err = contract.actionDefinitionDigest(request.ApplyStrategy.TombstoneAction)
		if err != nil {
			return synctransport.DestinationPlan{}, err
		}
		if err := validateDeclarativeTypedDestinationTombstoneApprovalDefinition(request.Approval, tombstoneActionDefinitionSHA256); err != nil {
			return synctransport.DestinationPlan{}, err
		}
		if err := validateDeclarativeTypedDestinationTombstoneIdempotencyProof(request.Approval, contract.descriptor.Executor, tombstoneActionDefinitionSHA256, contract.idempotencyHeaders[request.ApplyStrategy.TombstoneAction]); err != nil {
			return synctransport.DestinationPlan{}, err
		}
		tombstoneIdempotencyKeyHeader = contract.idempotencyHeaders[request.ApplyStrategy.TombstoneAction]
	}
	physicalActions := declarativeTypedDestinationPhysicalActions(declarativeTypedDestinationBinding{
		Action: request.ApplyStrategy.Action, ActionDefinitionSHA256: actionDefinitionSHA256, IdempotencyKeyHeader: contract.idempotencyHeaders[request.ApplyStrategy.Action],
		TombstoneAction: request.ApplyStrategy.TombstoneAction, TombstoneActionDefinitionSHA256: tombstoneActionDefinitionSHA256, TombstoneIdempotencyKeyHeader: tombstoneIdempotencyKeyHeader,
	})
	if err := validateDeclarativeTypedDestinationApprovedPhysicalActions(request.Approval, physicalActions); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy, ActionDefinitionSHA256: actionDefinitionSHA256, TombstoneActionDefinitionSHA256: tombstoneActionDefinitionSHA256, PhysicalActions: physicalActions}, nil
}

func (e *declarativeTypedDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if e == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative typed destination is unavailable")
	}
	if err := validateDeclarativeTypedDestinationApproval(request.Approval); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	contract, err := declarativeTypedDestinationContractFor(request.Destination)
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
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative typed destination action %q definition changed; replan and reapprove", request.Plan.ApplyStrategy.Action)
	}
	if err := validateDeclarativeTypedDestinationApprovalDefinition(request.Approval, actionDefinitionSHA256); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := validateDeclarativeTypedDestinationIdempotencyProof(request.Approval, contract.descriptor.Executor, actionDefinitionSHA256, contract.idempotencyHeaders[request.Plan.ApplyStrategy.Action]); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	var tombstoneBinding *connectors.DestinationSourceBinding
	var tombstoneActionDefinitionSHA256, tombstoneIdempotencyKeyHeader string
	if request.Plan.ApplyStrategy.TombstoneAction != "" {
		resolvedTombstoneBinding, bindingErr := contract.tombstoneBinding(request.Source, request.Stream, request.Mode, request.Plan.ApplyStrategy)
		if bindingErr != nil {
			return synccontract.DownstreamAcknowledgement{}, bindingErr
		}
		tombstoneBinding = &resolvedTombstoneBinding
		tombstoneActionDefinitionSHA256, err = contract.actionDefinitionDigest(request.Plan.ApplyStrategy.TombstoneAction)
		if err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
		if request.Plan.TombstoneActionDefinitionSHA256 != tombstoneActionDefinitionSHA256 {
			return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative typed destination tombstone action %q definition changed; replan and reapprove", request.Plan.ApplyStrategy.TombstoneAction)
		}
		if err := validateDeclarativeTypedDestinationTombstoneApprovalDefinition(request.Approval, tombstoneActionDefinitionSHA256); err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
		if err := validateDeclarativeTypedDestinationTombstoneIdempotencyProof(request.Approval, contract.descriptor.Executor, tombstoneActionDefinitionSHA256, contract.idempotencyHeaders[request.Plan.ApplyStrategy.TombstoneAction]); err != nil {
			return synccontract.DownstreamAcknowledgement{}, err
		}
		tombstoneIdempotencyKeyHeader = contract.idempotencyHeaders[request.Plan.ApplyStrategy.TombstoneAction]
	}
	physicalActions := declarativeTypedDestinationPhysicalActions(declarativeTypedDestinationBinding{
		Action: request.Plan.ApplyStrategy.Action, ActionDefinitionSHA256: actionDefinitionSHA256, IdempotencyKeyHeader: contract.idempotencyHeaders[request.Plan.ApplyStrategy.Action],
		TombstoneAction: request.Plan.ApplyStrategy.TombstoneAction, TombstoneActionDefinitionSHA256: tombstoneActionDefinitionSHA256, TombstoneIdempotencyKeyHeader: tombstoneIdempotencyKeyHeader,
	})
	if err := validateDeclarativeTypedDestinationPhysicalActions(request.Plan.PhysicalActions, physicalActions); err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("declarative typed destination plan %w", err)
	}
	if err := validateDeclarativeTypedDestinationApprovedPhysicalActions(request.Approval, physicalActions); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := validateDeclarativeTypedDestinationWorkset(request, binding, tombstoneBinding); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	results := make([]declarativeTypedDestinationActionResult, 0, 2)
	var recordReceipt json.RawMessage
	if len(request.Workset.Records) != 0 {
		records, mapErr := declarativeTypedDestinationRecords(request.Workset.Records, binding)
		if mapErr != nil {
			return synccontract.DownstreamAcknowledgement{}, mapErr
		}
		result, writeErr := declarativeTypedDestinationWrite(ctx, contract, request, request.Plan.ApplyStrategy.Action, records)
		results = append(results, declarativeTypedDestinationActionResult{Action: request.Plan.ApplyStrategy.Action, Result: result})
		output, outputErr := declarativeTypedDestinationApplyOutput(results, request.Runtime.Secrets)
		if outputErr != nil {
			return synccontract.DownstreamAcknowledgement{}, outputErr
		}
		if writeErr != nil {
			return synccontract.DownstreamAcknowledgement{}, synctransport.NewDestinationApplyOutputError(writeErr, output)
		}
		policy, policyErr := contract.readBackPolicy(request.Plan.ApplyStrategy)
		if policyErr != nil {
			return synccontract.DownstreamAcknowledgement{}, policyErr
		}
		recordReceipt, err = declarativeTypedDestinationReadBackReceipt(result, policy, actionDefinitionSHA256)
		if err != nil {
			return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("extract declarative typed destination action %q read-back receipt: %w", request.Plan.ApplyStrategy.Action, err)
		}
	}
	var tombstoneReceipt json.RawMessage
	if len(request.Workset.Tombstones) != 0 {
		tombstones, mapErr := declarativeTypedDestinationTombstoneRecords(request.Workset.Tombstones, *tombstoneBinding)
		if mapErr != nil {
			return synccontract.DownstreamAcknowledgement{}, mapErr
		}
		result, writeErr := declarativeTypedDestinationWrite(ctx, contract, request, request.Plan.ApplyStrategy.TombstoneAction, tombstones)
		results = append(results, declarativeTypedDestinationActionResult{Action: request.Plan.ApplyStrategy.TombstoneAction, Result: result})
		output, outputErr := declarativeTypedDestinationApplyOutput(results, request.Runtime.Secrets)
		if outputErr != nil {
			return synccontract.DownstreamAcknowledgement{}, outputErr
		}
		if writeErr != nil {
			return synccontract.DownstreamAcknowledgement{}, synctransport.NewDestinationApplyOutputError(writeErr, output)
		}
		policy, policyErr := contract.tombstoneReadBackPolicy(request.Plan.ApplyStrategy)
		if policyErr != nil {
			return synccontract.DownstreamAcknowledgement{}, policyErr
		}
		tombstoneReceipt, err = declarativeTypedDestinationReadBackReceiptForLocator(result, policy.ReceiptLocator, policy.MaxRecords, tombstoneActionDefinitionSHA256)
		if err != nil {
			return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("extract declarative typed destination tombstone action %q read-back receipt: %w", request.Plan.ApplyStrategy.TombstoneAction, err)
		}
	}
	output, outputErr := declarativeTypedDestinationApplyOutput(results, request.Runtime.Secrets)
	if outputErr != nil {
		return synccontract.DownstreamAcknowledgement{}, outputErr
	}
	privateReceipt, receiptErr := declarativeTypedDestinationCompositeReceipt(recordReceipt, tombstoneReceipt)
	if receiptErr != nil {
		return synccontract.DownstreamAcknowledgement{}, receiptErr
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(contract.connector.Name(), time.Now().UTC())
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	acknowledgement, err = acknowledgement.WithOutput(output)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("attach declarative typed destination action %q output: %w", request.Plan.ApplyStrategy.Action, err)
	}
	acknowledgement, err = acknowledgement.WithPrivateReceipt(privateReceipt)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("attach declarative typed destination action %q private read-back receipt: %w", request.Plan.ApplyStrategy.Action, err)
	}
	return acknowledgement, nil
}

type declarativeTypedDestinationActionResult struct {
	Action string                 `json:"action"`
	Result connectors.WriteResult `json:"result"`
}

type declarativeTypedDestinationCompositeReadBackReceipt struct {
	Version    int             `json:"version"`
	Records    json.RawMessage `json:"records,omitempty"`
	Tombstones json.RawMessage `json:"tombstones,omitempty"`
}

func declarativeTypedDestinationWrite(ctx context.Context, contract declarativeTypedDestinationContract, request synctransport.DestinationApplyRequest, action string, records []connectors.Record) (connectors.WriteResult, error) {
	writeEvidence := request.Approval.Evidence
	var err error
	if request.Approval.IssueWriteEvidence != nil {
		writeEvidence, err = request.Approval.IssueWriteEvidence(ctx)
		if err != nil {
			return connectors.WriteResult{}, fmt.Errorf("authorize declarative typed destination action %q: %w", action, err)
		}
	}
	if writeEvidence == nil {
		return connectors.WriteResult{}, fmt.Errorf("declarative typed destination action %q has no write evidence", action)
	}
	writeRequest := connectors.WriteRequest{Stream: request.Stream, Table: "sync_transport", Action: action, Config: request.Runtime, Approval: writeEvidence}
	if err := contract.connector.ValidateWrite(ctx, writeRequest, records); err != nil {
		return connectors.WriteResult{}, fmt.Errorf("validate declarative typed destination action %q: %w", action, err)
	}
	result, err := contract.connector.Write(transportpolicy.MarkDestructive(ctx), writeRequest, records)
	if err != nil {
		return result, fmt.Errorf("apply declarative typed destination action %q: %w", action, err)
	}
	if result.RecordsWritten != len(records) || result.RecordsFailed != 0 {
		return result, fmt.Errorf("declarative typed destination action %q wrote=%d failed=%d, want %d durable writes", action, result.RecordsWritten, result.RecordsFailed, len(records))
	}
	return result, nil
}

func declarativeTypedDestinationApplyOutput(results []declarativeTypedDestinationActionResult, secrets map[string]string) (json.RawMessage, error) {
	if len(results) == 1 && results[0].Action != "" {
		return json.Marshal(connectors.SanitizeWriteResultForOutput(results[0].Result, secrets))
	}
	for index := range results {
		results[index].Result = connectors.SanitizeWriteResultForOutput(results[index].Result, secrets)
	}
	return json.Marshal(struct {
		Actions []declarativeTypedDestinationActionResult `json:"actions"`
	}{Actions: results})
}

func declarativeTypedDestinationCompositeReceipt(records, tombstones json.RawMessage) (json.RawMessage, error) {
	if len(records) != 0 && len(tombstones) == 0 {
		return append(json.RawMessage(nil), records...), nil
	}
	if len(tombstones) == 0 {
		return nil, fmt.Errorf("declarative typed destination has no read-back receipt")
	}
	receipt, err := json.Marshal(declarativeTypedDestinationCompositeReadBackReceipt{Version: 1, Records: records, Tombstones: tombstones})
	if err != nil {
		return nil, fmt.Errorf("encode declarative typed destination composite read-back receipt: %w", err)
	}
	if len(receipt) > 16<<10 {
		return nil, fmt.Errorf("declarative typed destination composite read-back receipt exceeds its byte bound")
	}
	return receipt, nil
}

func declarativeTypedDestinationReadBackReceipt(result connectors.WriteResult, policy connectors.DestinationReadBackPolicy, actionDefinitionSHA256 string) (json.RawMessage, error) {
	return declarativeTypedDestinationReadBackReceiptForLocator(result, policy.ReceiptLocator, policy.MaxRecords, actionDefinitionSHA256)
}

func declarativeTypedDestinationReadBackReceiptForLocator(result connectors.WriteResult, locator connectors.DestinationReceiptLocator, maxRecords int, actionDefinitionSHA256 string) (json.RawMessage, error) {
	if err := locator.Validate(); err != nil {
		return nil, err
	}
	// A typed destination action is one declaration-owned physical request per
	// record. A compound protocol needs a dedicated closed adapter instead of
	// letting this generic path guess a receipt occurrence.
	if locator.ResponseIndex != 0 {
		return nil, fmt.Errorf("declarative typed destination receipt locator response_index %d is unavailable for a single-action write", locator.ResponseIndex)
	}
	if result.RecordsWritten < 1 || len(result.ProviderResponses) != result.RecordsWritten {
		return nil, fmt.Errorf("declarative typed destination write has incomplete provider receipts")
	}
	responses := make(map[int]connectors.WriteProviderResponse, len(result.ProviderResponses))
	for _, response := range result.ProviderResponses {
		if _, duplicate := responses[response.RecordIndex]; duplicate {
			return nil, fmt.Errorf("declarative typed destination write duplicates provider receipt index %d", response.RecordIndex)
		}
		responses[response.RecordIndex] = response
	}
	locators := make([]string, 0, result.RecordsWritten)
	for recordIndex := 0; recordIndex < result.RecordsWritten; recordIndex++ {
		response, found := responses[recordIndex]
		if !found {
			return nil, fmt.Errorf("declarative typed destination write is missing provider receipt index %d", recordIndex)
		}
		body, ok := response.Body.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("declarative typed destination provider receipt %d has no JSON object body", recordIndex)
		}
		value, found := body[locator.BodyField]
		if !found {
			return nil, fmt.Errorf("declarative typed destination provider receipt %d is missing locator field %q", recordIndex, locator.BodyField)
		}
		locatorValue, err := declarativeTypedDestinationReceiptLocatorValue(value, locator.MaxValueBytes)
		if err != nil {
			return nil, fmt.Errorf("declarative typed destination provider receipt %d locator: %w", recordIndex, err)
		}
		locators = append(locators, locatorValue)
	}
	return connectors.NewDeclarativeTypedDestinationReadBackReceipt(actionDefinitionSHA256, locator, locators, maxRecords)
}

func declarativeTypedDestinationReceiptLocatorValue(value any, maxBytes int) (string, error) {
	var locator string
	switch typed := value.(type) {
	case string:
		locator = typed
	case json.Number:
		locator = typed.String()
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, bool:
		locator = fmt.Sprint(typed)
	default:
		return "", fmt.Errorf("must be a scalar string, number, integer, or boolean")
	}
	if locator == "" || len(locator) > maxBytes {
		return "", fmt.Errorf("is outside its byte bound")
	}
	return locator, nil
}

// validateDeclarativeTypedDestinationApproval keeps the generic adapter on
// the existing plan -> preview -> approval path even when the selected action
// has no per-action destructive confirmation. The plan-owned target and
// digest are established before source I/O by PlanDestination; the
// orchestrator rechecks AuthorizeNextUnit before every apply unit.
func validateDeclarativeTypedDestinationApproval(approval synctransport.DestinationApproval) error {
	if approval.Evidence == nil {
		return fmt.Errorf("declarative typed destination requires approved write evidence")
	}
	if approval.AuthorizeNextUnit == nil {
		return fmt.Errorf("declarative typed destination requires a per-unit approval authorization")
	}
	if err := approval.Evidence.Validate(approval.Target, approval.PreviewDigest, time.Now().UTC()); err != nil {
		return fmt.Errorf("declarative typed destination approval: %w", err)
	}
	return nil
}

func validateDeclarativeTypedDestinationApprovedPhysicalActions(approval synctransport.DestinationApproval, expected []synctransport.DestinationPhysicalAction) error {
	if approval.Target.Scope == connectors.WriteApprovalScopeFixture {
		return nil
	}
	if err := validateDeclarativeTypedDestinationPhysicalActions(approval.PhysicalActions, expected); err != nil {
		return fmt.Errorf("declarative typed destination approval %w; replan and reapprove", err)
	}
	return nil
}

func validateDeclarativeTypedDestinationApprovalDefinition(approval synctransport.DestinationApproval, actionDefinitionSHA256 string) error {
	if approval.Target.Scope == connectors.WriteApprovalScopeFixture {
		return nil
	}
	if !constantTimeStringEqual(approval.ActionDefinitionSHA256, actionDefinitionSHA256) {
		return fmt.Errorf("declarative typed destination approval does not bind action definition; replan and reapprove")
	}
	return nil
}

func validateDeclarativeTypedDestinationTombstoneApprovalDefinition(approval synctransport.DestinationApproval, actionDefinitionSHA256 string) error {
	if approval.Target.Scope == connectors.WriteApprovalScopeFixture {
		return nil
	}
	if !constantTimeStringEqual(approval.TombstoneActionDefinitionSHA256, actionDefinitionSHA256) {
		return fmt.Errorf("declarative typed destination approval does not bind tombstone action definition; replan and reapprove")
	}
	return nil
}

// validateDeclarativeTypedDestinationIdempotencyProof keeps declaration
// conformance separate from action admission. The descriptor can honestly say
// it expects keyed delivery, but only the approved, definition-bound action
// proof can show which provider header carries that stable key. Fixture scope
// remains a hermetic test seam; production plans are always sealed below.
func validateDeclarativeTypedDestinationIdempotencyProof(approval synctransport.DestinationApproval, executor connectors.TransportExecutorReference, actionDefinitionSHA256, header string) error {
	if approval.Target.Scope == connectors.WriteApprovalScopeFixture {
		return nil
	}
	if err := executor.Validate(); err != nil {
		return fmt.Errorf("declarative typed destination idempotency executor: %w", err)
	}
	if strings.TrimSpace(header) == "" {
		return fmt.Errorf("declarative typed destination action has no independent idempotency proof")
	}
	proof := approval.IdempotencyProof
	if proof.Executor != executor || !constantTimeStringEqual(proof.ActionDefinitionSHA256, actionDefinitionSHA256) || !strings.EqualFold(strings.TrimSpace(proof.EffectiveHeader), header) {
		return fmt.Errorf("declarative typed destination approval does not bind the exact executor, action definition, and idempotency header")
	}
	return nil
}

func validateDeclarativeTypedDestinationTombstoneIdempotencyProof(approval synctransport.DestinationApproval, executor connectors.TransportExecutorReference, actionDefinitionSHA256, header string) error {
	if approval.Target.Scope == connectors.WriteApprovalScopeFixture {
		return nil
	}
	if err := executor.Validate(); err != nil {
		return fmt.Errorf("declarative typed destination tombstone idempotency executor: %w", err)
	}
	if strings.TrimSpace(header) == "" {
		return fmt.Errorf("declarative typed destination tombstone action has no independent idempotency proof")
	}
	proof := approval.TombstoneIdempotencyProof
	if proof.Executor != executor || !constantTimeStringEqual(proof.ActionDefinitionSHA256, actionDefinitionSHA256) || !strings.EqualFold(strings.TrimSpace(proof.EffectiveHeader), header) {
		return fmt.Errorf("declarative typed destination approval does not bind the exact tombstone executor, action definition, and idempotency header")
	}
	return nil
}

func (e *declarativeTypedDestinationExecutor) ReadBackDestination(ctx context.Context, request synctransport.DestinationReadBackRequest) error {
	if e == nil {
		return fmt.Errorf("declarative typed destination is unavailable")
	}
	contract, err := declarativeTypedDestinationContractFor(request.Destination)
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
		return fmt.Errorf("declarative typed destination action %q definition changed; replan and reapprove", request.Plan.ApplyStrategy.Action)
	}
	if request.Workset.ID == "" {
		return fmt.Errorf("declarative typed destination read-back requires a reopened workset")
	}
	if request.Acknowledgement.Sink != contract.connector.Name() || request.Acknowledgement.AcknowledgedAt.IsZero() {
		return fmt.Errorf("declarative typed destination read-back requires its durable acknowledgement")
	}
	privateReceipt, found := request.Acknowledgement.PrivateReceipt()
	if !found {
		return fmt.Errorf("declarative typed destination read-back requires its private provider receipt")
	}
	var tombstoneBinding *connectors.DestinationSourceBinding
	var tombstoneActionDefinitionSHA256 string
	if len(request.Workset.Tombstones) != 0 {
		resolvedTombstoneBinding, bindingErr := contract.tombstoneBinding(request.Source, request.Stream, request.Mode, request.Plan.ApplyStrategy)
		if bindingErr != nil {
			return bindingErr
		}
		tombstoneBinding = &resolvedTombstoneBinding
		tombstoneActionDefinitionSHA256, err = contract.actionDefinitionDigest(request.Plan.ApplyStrategy.TombstoneAction)
		if err != nil {
			return err
		}
		if request.Plan.TombstoneActionDefinitionSHA256 != tombstoneActionDefinitionSHA256 {
			return fmt.Errorf("declarative typed destination tombstone action %q definition changed; replan and reapprove", request.Plan.ApplyStrategy.TombstoneAction)
		}
	}
	if len(request.Workset.Records) != 0 {
		expected, mapErr := declarativeTypedDestinationRecords(request.Workset.Records, binding)
		if mapErr != nil {
			return mapErr
		}
		policy, policyErr := contract.readBackPolicy(request.Plan.ApplyStrategy)
		if policyErr != nil {
			return policyErr
		}
		if len(expected) > policy.MaxRecords {
			return fmt.Errorf("declarative typed destination read-back expected records exceed declared max_records %d", policy.MaxRecords)
		}
		recordReceipt, receiptErr := declarativeTypedDestinationReceiptPart(privateReceipt, false)
		if receiptErr != nil {
			return receiptErr
		}
		if err := declarativeTypedDestinationReadBack(ctx, contract.readBack, request.Runtime, policy.Operation, policy.MaxRecords, policy.MaxAttempts, policy.TimeoutMilliseconds, policy.RetryDelayMilliseconds, policy.ReceiptLocator, actionDefinitionSHA256, recordReceipt, func(provider []connectors.Record) error {
			return matchDeclarativeTypedDestinationProviderState(expected, provider, policy)
		}); err != nil {
			return err
		}
	}
	if len(request.Workset.Tombstones) != 0 {
		expected, mapErr := declarativeTypedDestinationTombstoneRecords(request.Workset.Tombstones, *tombstoneBinding)
		if mapErr != nil {
			return mapErr
		}
		policy, policyErr := contract.tombstoneReadBackPolicy(request.Plan.ApplyStrategy)
		if policyErr != nil {
			return policyErr
		}
		if len(expected) > policy.MaxRecords {
			return fmt.Errorf("declarative typed destination tombstone read-back expected records exceed declared max_records %d", policy.MaxRecords)
		}
		tombstoneReceipt, receiptErr := declarativeTypedDestinationReceiptPart(privateReceipt, true)
		if receiptErr != nil {
			return receiptErr
		}
		if err := declarativeTypedDestinationReadBack(ctx, contract.readBack, request.Runtime, policy.Operation, policy.MaxRecords, policy.MaxAttempts, policy.TimeoutMilliseconds, policy.RetryDelayMilliseconds, policy.ReceiptLocator, tombstoneActionDefinitionSHA256, tombstoneReceipt, func(provider []connectors.Record) error {
			return matchDeclarativeTypedDestinationProviderAbsence(expected, provider, *policy)
		}); err != nil {
			return err
		}
	}
	return nil
}

func declarativeTypedDestinationReceiptPart(privateReceipt json.RawMessage, tombstones bool) (json.RawMessage, error) {
	if !tombstones {
		var composite declarativeTypedDestinationCompositeReadBackReceipt
		if err := json.Unmarshal(privateReceipt, &composite); err == nil && composite.Version == 1 && len(composite.Tombstones) != 0 {
			if len(composite.Records) == 0 {
				return nil, fmt.Errorf("declarative typed destination composite receipt is missing ordinary record evidence")
			}
			return append(json.RawMessage(nil), composite.Records...), nil
		}
		return append(json.RawMessage(nil), privateReceipt...), nil
	}
	var composite declarativeTypedDestinationCompositeReadBackReceipt
	decoder := json.NewDecoder(bytes.NewReader(privateReceipt))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&composite); err != nil || composite.Version != 1 || len(composite.Tombstones) == 0 {
		return nil, fmt.Errorf("declarative typed destination tombstone read-back requires its composite private receipt")
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("declarative typed destination composite receipt has trailing values")
	}
	return append(json.RawMessage(nil), composite.Tombstones...), nil
}

func declarativeTypedDestinationReadBack(ctx context.Context, reader connectors.DeclarativeTypedDestinationReadBack, runtime connectors.RuntimeConfig, operation string, maxRecords, maxAttempts, timeoutMilliseconds, retryDelayMilliseconds int, locator connectors.DestinationReceiptLocator, actionDefinitionSHA256 string, receipt json.RawMessage, match func([]connectors.Record) error) error {
	readCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMilliseconds)*time.Millisecond)
	defer cancel()
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		provider, readErr := reader.ReadBackDeclarativeDestination(readCtx, connectors.DeclarativeTypedDestinationReadBackRequest{Operation: operation, Runtime: runtime, MaxRecords: maxRecords, Receipt: receipt, ReceiptLocator: locator, ActionDefinitionSHA256: actionDefinitionSHA256})
		if readErr == nil {
			readErr = match(provider)
		}
		if readErr == nil {
			return nil
		}
		lastErr = readErr
		if attempt == maxAttempts {
			break
		}
		timer := time.NewTimer(time.Duration(retryDelayMilliseconds) * time.Millisecond)
		select {
		case <-readCtx.Done():
			timer.Stop()
			return fmt.Errorf("declarative typed destination provider read-back: %w", readCtx.Err())
		case <-timer.C:
		}
	}
	return fmt.Errorf("declarative typed destination provider read-back did not confirm expected state: %w", lastErr)
}

func matchDeclarativeTypedDestinationProviderState(expected, provider []connectors.Record, policy connectors.DestinationReadBackPolicy) error {
	providerByIdentity := make(map[string]connectors.Record, len(provider))
	for _, record := range provider {
		identity, err := declarativeDestinationReadBackIdentity(record, policy.Identity, true)
		if err != nil {
			return err
		}
		if _, duplicate := providerByIdentity[identity]; duplicate {
			return fmt.Errorf("provider read-back returned duplicate destination identity")
		}
		providerByIdentity[identity] = record
	}
	seenExpected := make(map[string]struct{}, len(expected))
	for _, record := range expected {
		identity, err := declarativeDestinationReadBackIdentity(record, policy.Identity, false)
		if err != nil {
			return err
		}
		if _, duplicate := seenExpected[identity]; duplicate {
			return fmt.Errorf("destination workset contains duplicate read-back identity")
		}
		seenExpected[identity] = struct{}{}
		actual, found := providerByIdentity[identity]
		if !found {
			return fmt.Errorf("provider read-back is missing an expected destination identity")
		}
		for _, field := range policy.Expected {
			want, wantFound := record[field.ExpectedField]
			got, gotFound := actual[field.ProviderField]
			if !wantFound || !gotFound {
				return fmt.Errorf("provider read-back field %q does not match expected destination state", field.ProviderField)
			}
			equal, err := declarativeReadBackValuesEqual(want, got)
			if err != nil {
				return fmt.Errorf("provider read-back field %q comparison: %w", field.ProviderField, err)
			}
			if !equal {
				return fmt.Errorf("provider read-back field %q does not match expected destination state", field.ProviderField)
			}
		}
	}
	return nil
}

func matchDeclarativeTypedDestinationProviderAbsence(expected, provider []connectors.Record, policy connectors.DestinationTombstoneReadBackPolicy) error {
	providerByIdentity := make(map[string]struct{}, len(provider))
	for _, record := range provider {
		identity, err := declarativeDestinationReadBackIdentity(record, policy.Identity, true)
		if err != nil {
			return err
		}
		if _, duplicate := providerByIdentity[identity]; duplicate {
			return fmt.Errorf("provider tombstone read-back returned duplicate destination identity")
		}
		providerByIdentity[identity] = struct{}{}
	}
	seenExpected := make(map[string]struct{}, len(expected))
	for _, record := range expected {
		identity, err := declarativeDestinationReadBackIdentity(record, policy.Identity, false)
		if err != nil {
			return err
		}
		if _, duplicate := seenExpected[identity]; duplicate {
			return fmt.Errorf("destination tombstone workset contains duplicate read-back identity")
		}
		seenExpected[identity] = struct{}{}
		if _, found := providerByIdentity[identity]; found {
			return fmt.Errorf("provider tombstone read-back still contains an expected deleted destination identity")
		}
	}
	return nil
}

func declarativeDestinationReadBackIdentity(record connectors.Record, fields []connectors.DestinationReadBackField, provider bool) (string, error) {
	values := make([]any, 0, len(fields))
	for _, field := range fields {
		name := field.ExpectedField
		if provider {
			name = field.ProviderField
		}
		value, found := record[name]
		if !found || value == nil {
			return "", fmt.Errorf("destination read-back identity field %q is missing", name)
		}
		values = append(values, value)
	}
	canonical, err := canonicalDeclarativeReadBackValue(values)
	if err != nil {
		return "", fmt.Errorf("canonicalize destination read-back identity: %w", err)
	}
	return hashString(string(canonical)), nil
}

func declarativeTypedDestinationContractFor(connector connectors.Connector) (declarativeTypedDestinationContract, error) {
	candidate, ok := connector.(connectors.DeclarativeTypedDestination)
	if !ok || declarativeTypedDestinationIsNil(candidate) {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires a registered typed write capability")
	}
	definition, defined := connectors.DefinitionOf(candidate)
	if !defined {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires a definition")
	}
	if definition.SyncTransport == nil || definition.SyncTransport.Destination == nil {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires a destination declaration")
	}
	descriptor := *definition.SyncTransport.Destination
	if descriptor.Executor != declarativeTypedDestinationReference {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires its exact executor")
	}
	if err := descriptor.Validate(); err != nil {
		return declarativeTypedDestinationContract{}, err
	}
	if descriptor.Acknowledgement != connectors.TransportAcknowledgementDurableWarehouse {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires durable_warehouse acknowledgement")
	}
	if descriptor.Delivery.Idempotency != connectors.DeliveryIdempotencyKeyed || (descriptor.Delivery.Deletes != connectors.DeliveryDeletesUnavailable && descriptor.Delivery.Deletes != connectors.DeliveryDeletesTombstone) {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires keyed idempotency and declared delete semantics")
	}
	if len(descriptor.SourceBindings) == 0 {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires explicit source bindings")
	}
	recordActions := make(map[string]struct{}, len(descriptor.SourceBindings))
	tombstoneActions := make(map[string]struct{}, len(descriptor.SourceBindings))
	for _, binding := range descriptor.SourceBindings {
		if binding.Action == "" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires an action-owned source binding")
		}
		if binding.Batch == nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q requires a declaration batch", binding.Action)
		}
		if err := binding.Batch.Validate(); err != nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q batch disposition: %w", binding.Action, err)
		}
		if binding.TombstoneMapping != nil {
			tombstoneActions[binding.Action] = struct{}{}
			continue
		}
		if binding.RecordMapping.Kind != connectors.SourceRecordMappingKindInputFields {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires input_fields source mapping")
		}
		recordActions[binding.Action] = struct{}{}
	}
	if descriptor.ReadBack != nil || descriptor.TombstoneReadBack != nil {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination requires action-owned read-back declarations")
	}
	if descriptor.Delivery.Deletes != connectors.DeliveryDeletesTombstone && len(tombstoneActions) != 0 {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination with unavailable deletes cannot declare tombstone bindings or read-back")
	}
	readBack, ok := candidate.(connectors.DeclarativeTypedDestinationReadBack)
	if !ok || reflect.ValueOf(readBack).Kind() == reflect.Pointer && reflect.ValueOf(readBack).IsNil() {
		return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination does not implement its declared provider read-back")
	}
	actions := make(map[string]connectors.WriteActionInfo, len(definition.WriteActions))
	idempotencyHeaders := make(map[string]string, len(definition.WriteActions))
	for _, action := range definition.WriteActions {
		if _, duplicate := actions[action.Name]; duplicate {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination duplicates write action %q", action.Name)
		}
		actions[action.Name] = action
	}
	for _, strategy := range descriptor.ApplyStrategies {
		if _, bound := recordActions[strategy.Action]; !bound {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q has no action-owned source binding", strategy.Action)
		}
		action, found := actions[strategy.Action]
		if !found || action.Name == "" || strings.TrimSpace(action.Method) == "" || strings.TrimSpace(action.Path) == "" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination strategy %q names unavailable typed action %q", strategy.Mode, strategy.Action)
		}
		if action.TransportBinding != nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q selects a different closed adapter", action.Name)
		}
		if action.Kind == "delete" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination ordinary apply action %q must not be a delete", action.Name)
		}
		header, err := candidate.DeclarativeTypedDestinationIdempotencyHeader(action.Name)
		if err != nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q has no independent idempotency proof: %w", action.Name, err)
		}
		if strings.TrimSpace(header) == "" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q has no independent idempotency proof", action.Name)
		}
		idempotencyHeaders[action.Name] = header
		if strategy.ReadBack == nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q requires action-owned read-back", action.Name)
		}
		if strategy.ReadBack.ReceiptLocator.ResponseIndex != 0 {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q receipt locator response_index %d requires a dedicated compound destination adapter", action.Name, strategy.ReadBack.ReceiptLocator.ResponseIndex)
		}
		if strategy.ReadBack.Conformance != descriptor.Conformance {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q read-back conformance is not bound to the admitted destination evidence", action.Name)
		}
		if descriptor.Delivery.Deletes == connectors.DeliveryDeletesUnavailable {
			if strategy.TombstoneAction != "" || strategy.TombstoneReadBack != nil {
				return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination unavailable deletes cannot select tombstone action %q", strategy.TombstoneAction)
			}
			continue
		}
		if strategy.TombstoneAction == "" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination action %q has no declaration-owned tombstone action", strategy.Action)
		}
		if _, bound := tombstoneActions[strategy.TombstoneAction]; !bound {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q has no exact tombstone source binding", strategy.TombstoneAction)
		}
		deleteAction, found := actions[strategy.TombstoneAction]
		if !found || deleteAction.Name == "" || strings.TrimSpace(deleteAction.Method) == "" || strings.TrimSpace(deleteAction.Path) == "" || deleteAction.Kind != "delete" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q must name an available delete action", strategy.TombstoneAction)
		}
		if deleteAction.TransportBinding != nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q selects a different closed adapter", deleteAction.Name)
		}
		deleteHeader, err := candidate.DeclarativeTypedDestinationIdempotencyHeader(deleteAction.Name)
		if err != nil || strings.TrimSpace(deleteHeader) == "" {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q has no independent idempotency proof", deleteAction.Name)
		}
		idempotencyHeaders[deleteAction.Name] = deleteHeader
		if strategy.TombstoneReadBack == nil {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q requires action-owned read-back", deleteAction.Name)
		}
		if strategy.TombstoneReadBack.ReceiptLocator.ResponseIndex != 0 {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q receipt locator response_index %d requires a dedicated compound destination adapter", deleteAction.Name, strategy.TombstoneReadBack.ReceiptLocator.ResponseIndex)
		}
		if strategy.TombstoneReadBack.Conformance != descriptor.Conformance {
			return declarativeTypedDestinationContract{}, fmt.Errorf("declarative typed destination tombstone action %q read-back conformance is not bound to the admitted destination evidence", deleteAction.Name)
		}
	}
	return declarativeTypedDestinationContract{connector: candidate, readBack: readBack, descriptor: descriptor, actions: actions, idempotencyHeaders: idempotencyHeaders}, nil
}

func declarativeTypedDestinationIsNil(candidate connectors.DeclarativeTypedDestination) bool {
	if candidate == nil {
		return true
	}
	value := reflect.ValueOf(candidate)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func (c declarativeTypedDestinationContract) plan(source connectors.Connector, stream string, mode synccontract.Mode, strategy connectors.DestinationApplyStrategy) (connectors.DestinationSourceBinding, error) {
	if mode == synccontract.ModeFullOverwrite {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination does not implement full_overwrite")
	}
	expected, err := c.descriptor.ApplyStrategyForAction(mode, strategy.Action)
	if err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	if !reflect.DeepEqual(expected, strategy) {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination action %q is not the declared strategy for mode %q", strategy.Action, mode)
	}
	if _, found := c.actions[strategy.Action]; !found {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination action %q is unavailable", strategy.Action)
	}
	sourceDescriptor, declared := connectors.SourceTransportDescriptorOf(source)
	if !declared {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination source has no transport declaration")
	}
	binding, admitted := c.descriptor.SourceBindingForAction(sourceDescriptor.Executor, stream, strategy.Action)
	if !admitted {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination does not admit source executor %q for stream %q", sourceDescriptor.Executor.ID, stream)
	}
	if binding.Action != strategy.Action {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination action %q has no exact source binding", strategy.Action)
	}
	if binding.RecordMapping.Kind != connectors.SourceRecordMappingKindInputFields {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination requires input_fields source mapping")
	}
	inputs := make([]string, len(binding.RecordMapping.Inputs))
	for index, input := range binding.RecordMapping.Inputs {
		inputs[index] = input.Input
	}
	if err := c.connector.PreflightWriteRecordFieldMapping(strategy.Action, inputs); err != nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination action %q source inputs are not an exact complete record schema mapping: %w", strategy.Action, err)
	}
	policy, err := c.readBackPolicy(strategy)
	if err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	if err := validateDeclarativeTypedDestinationReadBackFields(strategy.Action, binding.RecordMapping.Inputs, policy.Identity, policy.Expected); err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	return binding, nil
}

// tombstoneBinding resolves the delete action sealed beside an ordinary apply
// strategy. It never falls back to the ordinary record mapping, so a source
// tombstone can reach only the declaration-owned delete action and image.
func (c declarativeTypedDestinationContract) tombstoneBinding(source connectors.Connector, stream string, mode synccontract.Mode, strategy connectors.DestinationApplyStrategy) (connectors.DestinationSourceBinding, error) {
	if c.descriptor.Delivery.Deletes != connectors.DeliveryDeletesTombstone || strategy.TombstoneAction == "" {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination action %q does not declare tombstone delivery", strategy.Action)
	}
	if _, err := c.plan(source, stream, mode, strategy); err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	sourceDescriptor, declared := connectors.SourceTransportDescriptorOf(source)
	if !declared {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination source has no transport declaration")
	}
	binding, admitted := c.descriptor.SourceBindingForAction(sourceDescriptor.Executor, stream, strategy.TombstoneAction)
	if !admitted || binding.Action != strategy.TombstoneAction || binding.TombstoneMapping == nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination tombstone action %q has no exact tombstone source binding", strategy.TombstoneAction)
	}
	inputs := make([]string, len(binding.TombstoneMapping.Inputs))
	for index, input := range binding.TombstoneMapping.Inputs {
		inputs[index] = input.Input
	}
	if err := c.connector.PreflightWriteRecordFieldMapping(strategy.TombstoneAction, inputs); err != nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("declarative typed destination tombstone action %q inputs are not an exact complete record schema mapping: %w", strategy.TombstoneAction, err)
	}
	policy, err := c.tombstoneReadBackPolicy(strategy)
	if err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	if err := validateDeclarativeTypedDestinationReadBackFields(strategy.TombstoneAction, binding.TombstoneMapping.Inputs, policy.Identity, nil); err != nil {
		return connectors.DestinationSourceBinding{}, err
	}
	return binding, nil
}

func (c declarativeTypedDestinationContract) readBackPolicy(strategy connectors.DestinationApplyStrategy) (connectors.DestinationReadBackPolicy, error) {
	if strategy.ReadBack == nil {
		return connectors.DestinationReadBackPolicy{}, fmt.Errorf("declarative typed destination action %q requires action-owned read-back", strategy.Action)
	}
	return *strategy.ReadBack, nil
}

func (c declarativeTypedDestinationContract) tombstoneReadBackPolicy(strategy connectors.DestinationApplyStrategy) (*connectors.DestinationTombstoneReadBackPolicy, error) {
	if strategy.TombstoneAction == "" {
		return nil, nil
	}
	if strategy.TombstoneReadBack == nil {
		return nil, fmt.Errorf("declarative typed destination tombstone action %q requires action-owned read-back", strategy.TombstoneAction)
	}
	policy := *strategy.TombstoneReadBack
	return &policy, nil
}

func validateDeclarativeTypedDestinationReadBackFields(action string, inputs []connectors.SourceRecordInputBinding, identity, expected []connectors.DestinationReadBackField) error {
	available := make(map[string]struct{}, len(inputs))
	for _, input := range inputs {
		available[input.Input] = struct{}{}
	}
	for _, field := range append(append([]connectors.DestinationReadBackField(nil), identity...), expected...) {
		if _, found := available[field.ExpectedField]; !found {
			return fmt.Errorf("declarative typed destination action %q action-owned read-back field %q is not supplied by its exact record mapping", action, field.ExpectedField)
		}
	}
	return nil
}

func (c declarativeTypedDestinationContract) actionDefinitionDigest(action string) (string, error) {
	digest, err := c.connector.DeclarativeTypedDestinationActionDigest(action)
	if err != nil {
		return "", fmt.Errorf("hash declarative typed destination action %q: %w", action, err)
	}
	if strings.TrimSpace(digest) == "" {
		return "", fmt.Errorf("declarative typed destination action %q has no definition digest", action)
	}
	return digest, nil
}

func (c declarativeTypedDestinationContract) idempotencyHeader(action string) (string, error) {
	header, found := c.idempotencyHeaders[action]
	if !found || strings.TrimSpace(header) == "" {
		return "", fmt.Errorf("declarative typed destination action %q has no independent idempotency proof", action)
	}
	return header, nil
}

// validateDeclarativeTypedDestinationSelection binds a registry-selected
// action to its action schema before a connection is persisted or a run is
// allowed to read a source. The registry proves connector/source/mode/action
// eligibility; this adds the exact writes.json record-property proof for the
// selected action without accepting a caller-provided request shape.
func validateDeclarativeTypedDestinationSelection(source, destination connectors.Connector, stream string, mode synccontract.Mode, strategy connectors.DestinationApplyStrategy) error {
	contract, err := declarativeTypedDestinationContractFor(destination)
	if err != nil {
		return err
	}
	_, err = contract.plan(source, stream, mode, strategy)
	return err
}

// declarativeTypedDestinationEffectiveBatchSize clamps the source page to the
// tightest declaration-owned action unit. The caller may ask to fetch fewer
// rows, but cannot expand an acknowledged workset beyond either its ordinary
// action or its paired tombstone action.
func declarativeTypedDestinationEffectiveBatchSize(source, destination connectors.Connector, stream string, mode synccontract.Mode, strategy connectors.DestinationApplyStrategy, requested int) (int, error) {
	if requested < 1 {
		return 0, fmt.Errorf("declarative typed destination requested batch size must be positive")
	}
	contract, err := declarativeTypedDestinationContractFor(destination)
	if err != nil {
		return 0, err
	}
	binding, err := contract.plan(source, stream, mode, strategy)
	if err != nil {
		return 0, err
	}
	effective := requested
	if binding.Batch.MaxRecords < effective {
		effective = binding.Batch.MaxRecords
	}
	if strategy.TombstoneAction != "" {
		tombstoneBinding, err := contract.tombstoneBinding(source, stream, mode, strategy)
		if err != nil {
			return 0, err
		}
		if tombstoneBinding.Batch.MaxRecords < effective {
			effective = tombstoneBinding.Batch.MaxRecords
		}
	}
	return effective, nil
}

func validateDeclarativeTypedDestinationWorkset(request synctransport.DestinationApplyRequest, binding connectors.DestinationSourceBinding, tombstoneBinding *connectors.DestinationSourceBinding) error {
	if err := request.Receipt.Validate(); err != nil {
		return fmt.Errorf("declarative typed destination receipt: %w", err)
	}
	if request.ConnectionID == "" || request.Receipt.Owner != request.ConnectionID || request.Receipt.ID != request.Workset.ID {
		return fmt.Errorf("declarative typed destination receipt does not bind the reopened workset")
	}
	if request.Receipt.Records != len(request.Workset.Records) || request.Receipt.Tombstones != len(request.Workset.Tombstones) {
		return fmt.Errorf("declarative typed destination receipt counts do not bind the reopened workset")
	}
	if len(request.Workset.Records) == 0 && len(request.Workset.Tombstones) == 0 {
		return fmt.Errorf("declarative typed destination requires a non-empty reopened workset")
	}
	if len(request.Workset.Records) > binding.Batch.MaxRecords {
		return fmt.Errorf("declarative typed destination action %q workset has %d records, exceeding declaration batch maximum %d", binding.Action, len(request.Workset.Records), binding.Batch.MaxRecords)
	}
	if len(request.Workset.Tombstones) != 0 {
		if tombstoneBinding == nil {
			return fmt.Errorf("declarative typed destination does not declare tombstone deletes")
		}
		if len(request.Workset.Tombstones) > tombstoneBinding.Batch.MaxRecords {
			return fmt.Errorf("declarative typed destination tombstone action %q workset has %d tombstones, exceeding declaration batch maximum %d", tombstoneBinding.Action, len(request.Workset.Tombstones), tombstoneBinding.Batch.MaxRecords)
		}
	}
	return nil
}

func declarativeTypedDestinationRecords(source []connectors.Record, binding connectors.DestinationSourceBinding) ([]connectors.Record, error) {
	records := make([]connectors.Record, 0, len(source))
	for index, row := range source {
		if row == nil {
			return nil, fmt.Errorf("declarative typed destination source row %d is absent", index)
		}
		record := make(connectors.Record, len(binding.RecordMapping.Inputs))
		for _, input := range binding.RecordMapping.Inputs {
			value, found := row[input.Field]
			if !found {
				return nil, fmt.Errorf("declarative typed destination source row %d has no value for action input %q", index, input.Input)
			}
			record[input.Input] = value
		}
		records = append(records, record)
	}
	return records, nil
}

func declarativeTypedDestinationTombstoneRecords(tombstones []synccontract.Tombstone, binding connectors.DestinationSourceBinding) ([]connectors.Record, error) {
	if binding.TombstoneMapping == nil {
		return nil, fmt.Errorf("declarative typed destination tombstone action %q has no tombstone mapping", binding.Action)
	}
	records := make([]connectors.Record, 0, len(tombstones))
	for index, tombstone := range tombstones {
		if err := tombstone.Validate(); err != nil {
			return nil, fmt.Errorf("declarative typed destination tombstone %d: %w", index, err)
		}
		if tombstone.Operation != synccontract.OperationDelete {
			return nil, fmt.Errorf("declarative typed destination tombstone %d has unsupported operation %q", index, tombstone.Operation)
		}
		image := tombstone.Key
		if binding.TombstoneMapping.Image == connectors.TombstoneRecordMappingImageBefore {
			if tombstone.DeleteImage != synccontract.DeleteImageBefore {
				return nil, fmt.Errorf("declarative typed destination tombstone %d has no declared before image", index)
			}
			image = tombstone.Before
		}
		decoder := json.NewDecoder(bytes.NewReader(image))
		decoder.UseNumber()
		values := make(map[string]any)
		if err := decoder.Decode(&values); err != nil {
			return nil, fmt.Errorf("declarative typed destination tombstone %d declared image must be an object: %w", index, err)
		}
		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("declarative typed destination tombstone %d declared image has trailing values", index)
		}
		record := make(connectors.Record, len(binding.TombstoneMapping.Inputs))
		for _, input := range binding.TombstoneMapping.Inputs {
			value, found := values[input.Field]
			if !found {
				return nil, fmt.Errorf("declarative typed destination tombstone %d has no value for delete input %q", index, input.Input)
			}
			record[input.Input] = value
		}
		records = append(records, record)
	}
	return records, nil
}

func declarativeStreamTransportConnector(connector connectors.Connector) (*engine.Connector, connectors.SourceTransportDescriptor, error) {
	candidate, ok := connector.(*engine.Connector)
	if !ok || candidate == nil {
		return nil, connectors.SourceTransportDescriptor{}, fmt.Errorf("declarative stream transport requires an engine connector")
	}
	definition := candidate.Definition()
	if definition.SyncTransport == nil || definition.SyncTransport.Source == nil {
		return nil, connectors.SourceTransportDescriptor{}, fmt.Errorf("declarative stream transport requires a source declaration")
	}
	descriptor := *definition.SyncTransport.Source
	if descriptor.Executor != declarativeStreamSourceReference {
		return nil, connectors.SourceTransportDescriptor{}, fmt.Errorf("declarative stream transport requires its exact executor")
	}
	if err := validateDeclarativeStreamEligibility(definition.Streams, descriptor.EligibleStreams); err != nil {
		return nil, connectors.SourceTransportDescriptor{}, err
	}
	return candidate, descriptor, nil
}

func validateDeclarativeStreamEligibility(streams []connectors.StreamSummary, eligible []string) error {
	if len(streams) == 0 || len(eligible) != len(streams) {
		return fmt.Errorf("declarative stream transport eligibility must match every executable stream")
	}
	want := make(map[string]struct{}, len(streams))
	for _, stream := range streams {
		if strings.TrimSpace(stream.Name) == "" {
			return fmt.Errorf("declarative stream transport contains an unnamed stream")
		}
		want[stream.Name] = struct{}{}
	}
	seen := make(map[string]struct{}, len(eligible))
	for _, stream := range eligible {
		if stream == "*" {
			return fmt.Errorf("declarative stream transport requires a concrete positive allowlist")
		}
		if _, ok := want[stream]; !ok {
			return fmt.Errorf("declarative stream transport eligibility names unknown stream %q", stream)
		}
		if _, duplicate := seen[stream]; duplicate {
			return fmt.Errorf("declarative stream transport eligibility repeats stream %q", stream)
		}
		seen[stream] = struct{}{}
	}
	return nil
}

func issueLabelTransportConnectorContract(connector connectors.Connector) (*engine.Connector, issueLabelTransportContract, error) {
	candidate, ok := connector.(*engine.Connector)
	if !ok || candidate == nil {
		return nil, issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport requires a declarative connector")
	}
	contract, err := issueLabelTransportContractForDefinition(candidate.Definition())
	if err != nil {
		return nil, issueLabelTransportContract{}, err
	}
	return candidate, contract, nil
}

// issueLabelTransportEngine resolves the exact existing declarative bundle by
// its closed typed contract. It must remain unique; ambiguity fails closed
// instead of turning the walking slice into a generic connector transport.
func issueLabelTransportEngine(registry *connectors.Registry) (*engine.Connector, issueLabelTransportContract, error) {
	if registry == nil {
		return nil, issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport registry is unavailable")
	}
	var selected *engine.Connector
	var selectedContract issueLabelTransportContract
	for _, metadata := range registry.List() {
		registered, ok := registry.Get(metadata.Name)
		if !ok {
			continue
		}
		candidate, contract, err := issueLabelTransportConnectorContract(registered)
		if err != nil {
			definition, ok := connectors.DefinitionOf(registered)
			if !ok || !definitionDeclaresIssueLabelTransport(definition) {
				continue
			}
			return nil, issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition %q is invalid: %w", registered.Name(), err)
		}
		if selected != nil && selected.Name() != candidate.Name() {
			return nil, issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport contract is ambiguous across declarative connectors")
		}
		selected = candidate
		selectedContract = contract
	}
	if selected == nil {
		return nil, issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport requires one declarative connector with the exact issue-label capability")
	}
	return selected, selectedContract, nil
}

func definitionDeclaresIssueLabelTransport(definition connectors.Definition) bool {
	for _, action := range definition.WriteActions {
		if action.TransportBinding != nil && action.TransportBinding.Capability == connectors.TransportCapabilityIssueLabel {
			return true
		}
	}
	return false
}

type issueLabelTransportAction struct {
	name    string
	binding connectors.TransportActionBinding
}

// IssueLabelTransportRowMappingError is a pre-write refusal for a source row
// that cannot satisfy the destination definition's closed transport
// inputs. It carries only the input name and a structural reason, never the
// row value.
type IssueLabelTransportRowMappingError struct {
	Input  string
	Reason string
}

func (e *IssueLabelTransportRowMappingError) Error() string {
	if e == nil {
		return "issue-label transport row cannot map to the destination action"
	}
	if e.Reason == "" {
		return fmt.Sprintf("issue-label transport row cannot map input %q", e.Input)
	}
	return fmt.Sprintf("issue-label transport row cannot map input %q: %s", e.Input, e.Reason)
}

// IssueLabelTransportUnsupportedActionError identifies an action outside the
// two definition-owned label actions. It is deliberately returned
// before a workset is applied, so an untrusted transport plan cannot turn a
// malformed action into provider I/O.
type IssueLabelTransportUnsupportedActionError struct {
	Action string
}

func (e *IssueLabelTransportUnsupportedActionError) Error() string {
	if e == nil || strings.TrimSpace(e.Action) == "" {
		return "closed issue-label destination received an unsupported action"
	}
	return fmt.Sprintf("closed issue-label destination received unsupported action %q", e.Action)
}

// IssueLabelTransportDeletesUnavailableError refuses a receipt carrying a
// delete the destination cannot represent. A source declaring
// deletes:not_available must surface it rather than silently treating a
// malformed receipt as an ordinary label write.
type IssueLabelTransportDeletesUnavailableError struct {
	Tombstones int
}

func (e *IssueLabelTransportDeletesUnavailableError) Error() string {
	if e == nil || e.Tombstones <= 0 {
		return "closed issue-label destination does not support deletes"
	}
	return fmt.Sprintf("closed issue-label destination does not support %d tombstone delete(s)", e.Tombstones)
}

type issueLabelTransportContract struct {
	stream  string
	apply   issueLabelTransportAction
	replace issueLabelTransportAction
	cleanup issueLabelTransportAction
}

func (c issueLabelTransportContract) modes() []synccontract.Mode {
	seen := make(map[synccontract.Mode]bool)
	for _, action := range []issueLabelTransportAction{c.apply, c.replace} {
		for _, mode := range action.binding.Modes {
			seen[mode] = true
		}
	}
	ordered := make([]synccontract.Mode, 0, len(seen))
	for _, mode := range synccontract.AllModes() {
		if seen[mode] {
			ordered = append(ordered, mode)
		}
	}
	return ordered
}

func (c issueLabelTransportContract) destinationActionNames() []string {
	return []string{c.apply.name, c.replace.name}
}

func (c issueLabelTransportContract) actionForSyncMode(mode synccontract.Mode) (issueLabelTransportAction, error) {
	for _, action := range []issueLabelTransportAction{c.apply, c.replace} {
		for _, declared := range action.binding.Modes {
			if declared == mode {
				return action, nil
			}
		}
	}
	return issueLabelTransportAction{}, &synccontract.ModeNotExecutableError{
		Mode:   mode,
		Reason: "closed issue-label destination has no definition-owned action for this mode",
	}
}

func (c issueLabelTransportContract) applyStrategies() ([]connectors.DestinationApplyStrategy, error) {
	modes := c.modes()
	strategies := make([]connectors.DestinationApplyStrategy, 0, len(modes))
	for _, mode := range modes {
		action, err := c.actionForSyncMode(mode)
		if err != nil {
			return nil, err
		}
		strategy, err := issueLabelTransportApplyStrategy(mode)
		if err != nil {
			return nil, err
		}
		strategies = append(strategies, connectors.DestinationApplyStrategy{Mode: mode, Strategy: strategy, Action: action.name})
	}
	return strategies, nil
}

func (c issueLabelTransportContract) matchesApplyStrategy(strategy connectors.DestinationApplyStrategy) bool {
	action, err := c.actionForSyncMode(strategy.Mode)
	if err != nil || action.name != strategy.Action {
		return false
	}
	want, err := issueLabelTransportApplyStrategy(strategy.Mode)
	return err == nil && strategy.Strategy == want
}

func issueLabelTransportApplyStrategy(mode synccontract.Mode) (connectors.ApplyStrategy, error) {
	switch mode {
	case synccontract.ModeFullAppend:
		return connectors.ApplyStrategyAppend, nil
	case synccontract.ModeFullOverwrite:
		return connectors.ApplyStrategyReplace, nil
	case synccontract.ModeIncrementalUpsert:
		return connectors.ApplyStrategyMerge, nil
	default:
		return "", fmt.Errorf("closed issue-label transport has no destination strategy for sync mode %q", mode)
	}
}

func issueLabelTransportContractForDefinition(definition connectors.Definition) (issueLabelTransportContract, error) {
	hasIssues := false
	for _, stream := range definition.Streams {
		if stream.Name == "issues" {
			hasIssues = true
			break
		}
	}
	if !hasIssues {
		return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition has no issues stream")
	}
	contract := issueLabelTransportContract{stream: "issues"}
	for _, action := range definition.WriteActions {
		if action.TransportBinding == nil || action.TransportBinding.Capability != connectors.TransportCapabilityIssueLabel {
			continue
		}
		bound, err := issueLabelTransportActionFromDefinition(action)
		if err != nil {
			return issueLabelTransportContract{}, err
		}
		switch bound.binding.Role {
		case connectors.TransportActionRoleApply:
			if contract.apply.name != "" {
				return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition declares more than one apply action")
			}
			contract.apply = bound
		case connectors.TransportActionRoleReplace:
			if contract.replace.name != "" {
				return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition declares more than one replace action")
			}
			contract.replace = bound
		case connectors.TransportActionRoleCleanup:
			if contract.cleanup.name != "" {
				return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition declares more than one cleanup action")
			}
			contract.cleanup = bound
		default:
			return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport action %q declares an unknown role %q", action.Name, bound.binding.Role)
		}
	}
	if contract.apply.name == "" || contract.replace.name == "" || contract.cleanup.name == "" {
		return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition requires one apply, replace, and cleanup action")
	}
	if !issueLabelTransportActionHasExactModes(contract.apply, synccontract.ModeFullAppend) ||
		!issueLabelTransportActionHasExactModes(contract.replace, synccontract.ModeFullOverwrite, synccontract.ModeIncrementalUpsert) ||
		!issueLabelTransportActionHasExactModes(contract.cleanup, synccontract.ModeFullAppend) {
		return issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport definition declares unsupported action modes")
	}
	return contract, nil
}

func issueLabelTransportActionHasExactModes(action issueLabelTransportAction, expected ...synccontract.Mode) bool {
	if len(action.binding.Modes) != len(expected) {
		return false
	}
	seen := make(map[synccontract.Mode]bool, len(expected))
	for _, mode := range action.binding.Modes {
		if err := mode.Validate(); err != nil || seen[mode] {
			return false
		}
		seen[mode] = true
	}
	for _, mode := range expected {
		if !seen[mode] {
			return false
		}
	}
	return true
}

func issueLabelTransportActionFromDefinition(action connectors.WriteActionInfo) (issueLabelTransportAction, error) {
	binding := action.TransportBinding
	if binding == nil || binding.Capability != connectors.TransportCapabilityIssueLabel || strings.TrimSpace(action.Name) == "" {
		return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action is not definition-owned")
	}
	if strings.TrimSpace(action.Method) == "" || strings.TrimSpace(action.Path) == "" {
		return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q has no declared endpoint", action.Name)
	}
	if binding.Role != connectors.TransportActionRoleApply && binding.Role != connectors.TransportActionRoleReplace && binding.Role != connectors.TransportActionRoleCleanup {
		return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q declares an unknown role %q", action.Name, binding.Role)
	}
	if len(binding.Modes) == 0 {
		return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q declares no sync modes", action.Name)
	}
	if len(binding.Inputs) != 2 {
		return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q must bind exactly two typed inputs", action.Name)
	}
	seenInputs := make(map[string]bool, len(binding.Inputs))
	seenFields := make(map[string]bool, len(binding.Inputs))
	for _, input := range binding.Inputs {
		if strings.TrimSpace(input.Field) == "" || seenInputs[input.Input] || seenFields[input.Field] {
			return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q has an invalid input binding", action.Name)
		}
		seenInputs[input.Input] = true
		seenFields[input.Field] = true
		switch input.Input {
		case connectors.TransportInputTargetIssue:
			if input.Shape != connectors.TransportInputShapeScalar {
				return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q must bind target_issue as a scalar", action.Name)
			}
		case connectors.TransportInputLabel:
			if input.Shape != connectors.TransportInputShapeScalar && input.Shape != connectors.TransportInputShapeList {
				return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q has an unsupported label shape", action.Name)
			}
		default:
			return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q has an unknown input %q", action.Name, input.Input)
		}
	}
	if !seenInputs[connectors.TransportInputTargetIssue] || !seenInputs[connectors.TransportInputLabel] {
		return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport action %q must bind target_issue and label", action.Name)
	}
	return issueLabelTransportAction{name: action.Name, binding: *binding.Clone()}, nil
}

func (a issueLabelTransportAction) record(issueNumber int, label string) (connectors.Record, error) {
	if issueNumber <= 0 || strings.TrimSpace(label) == "" {
		return nil, fmt.Errorf("closed issue-label transport requires a positive issue number and non-empty label")
	}
	record := make(connectors.Record, len(a.binding.Inputs))
	for _, input := range a.binding.Inputs {
		switch input.Input {
		case connectors.TransportInputTargetIssue:
			record[input.Field] = issueNumber
		case connectors.TransportInputLabel:
			if input.Shape == connectors.TransportInputShapeList {
				record[input.Field] = []string{label}
			} else {
				record[input.Field] = label
			}
		default:
			return nil, fmt.Errorf("closed issue-label transport action %q has an unknown input %q", a.name, input.Input)
		}
	}
	return record, nil
}

// recordFromSourceRecord maps only the declaration-owned source input fields
// into the action record. The selected binding is not a generic mapping
// surface: it must provide exactly the action's declared inputs.
func (a issueLabelTransportAction) recordFromSourceRecord(source connectors.Record, mappings []connectors.SourceRecordInputBinding) (connectors.Record, error) {
	if source == nil {
		return nil, &IssueLabelTransportRowMappingError{Reason: "row is absent"}
	}
	fields := make(map[string]string, len(mappings))
	for _, mapping := range mappings {
		if _, duplicate := fields[mapping.Input]; duplicate {
			return nil, &IssueLabelTransportRowMappingError{Input: mapping.Input, Reason: "source binding repeats an input"}
		}
		fields[mapping.Input] = mapping.Field
	}
	if len(fields) != len(a.binding.Inputs) {
		return nil, &IssueLabelTransportRowMappingError{Reason: "source binding does not provide exactly the declared inputs"}
	}
	var targetIssue int
	var label string
	for _, input := range a.binding.Inputs {
		field, declared := fields[input.Input]
		if !declared {
			return nil, &IssueLabelTransportRowMappingError{Input: input.Input, Reason: "source binding does not declare an input field"}
		}
		value, ok := source[field]
		if !ok || value == nil {
			return nil, &IssueLabelTransportRowMappingError{Input: input.Input, Reason: "value is null or absent"}
		}
		switch input.Input {
		case connectors.TransportInputTargetIssue:
			parsed, err := issueLabelTransportPositiveInteger(value)
			if err != nil {
				return nil, &IssueLabelTransportRowMappingError{Input: input.Input, Reason: "must be a positive integer"}
			}
			targetIssue = parsed
		case connectors.TransportInputLabel:
			parsed, ok := value.(string)
			if !ok || strings.TrimSpace(parsed) == "" {
				return nil, &IssueLabelTransportRowMappingError{Input: input.Input, Reason: "must be a non-empty string"}
			}
			label = strings.TrimSpace(parsed)
		default:
			return nil, &IssueLabelTransportRowMappingError{Input: input.Input, Reason: "is not declared for issue-label transport"}
		}
	}
	if targetIssue == 0 || label == "" {
		return nil, &IssueLabelTransportRowMappingError{Reason: "definition did not provide both target_issue and label"}
	}
	return a.record(targetIssue, label)
}

func issueLabelTransportPositiveInteger(value any) (int, error) {
	switch number := value.(type) {
	case int:
		if number > 0 {
			return number, nil
		}
	case int64:
		if number > 0 && number <= int64(^uint(0)>>1) {
			return int(number), nil
		}
	case float64:
		if number > 0 && number == float64(int(number)) {
			return int(number), nil
		}
	case json.Number:
		parsed, err := number.Int64()
		if err == nil && parsed > 0 && parsed <= int64(^uint(0)>>1) {
			return int(parsed), nil
		}
	case string:
		parsed, err := strconv.Atoi(number)
		if err == nil && parsed > 0 {
			return parsed, nil
		}
	}
	return 0, fmt.Errorf("not a positive integer")
}

func (a issueLabelTransportAction) targetAndLabel(record connectors.Record) (int, string, error) {
	var targetIssue int
	var label string
	for _, input := range a.binding.Inputs {
		value, ok := record[input.Field]
		if !ok || value == nil {
			return 0, "", fmt.Errorf("closed issue-label action record has no %s field", input.Input)
		}
		switch input.Input {
		case connectors.TransportInputTargetIssue:
			parsed, err := issueLabelTransportPositiveInteger(value)
			if err != nil {
				return 0, "", err
			}
			targetIssue = parsed
		case connectors.TransportInputLabel:
			switch values := value.(type) {
			case []string:
				if len(values) != 1 || strings.TrimSpace(values[0]) == "" {
					return 0, "", fmt.Errorf("closed issue-label action record has an invalid labels array")
				}
				label = values[0]
			case string:
				if strings.TrimSpace(values) == "" {
					return 0, "", fmt.Errorf("closed issue-label action record has an empty label")
				}
				label = values
			default:
				return 0, "", fmt.Errorf("closed issue-label action record has an invalid label")
			}
		}
	}
	if targetIssue == 0 || label == "" {
		return 0, "", fmt.Errorf("closed issue-label action record is incomplete")
	}
	return targetIssue, label, nil
}

type declarativeStreamSourceExecutor struct {
}

func (*declarativeStreamSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return declarativeStreamSourceReference
}

// AllowEmptySourceResult admits an executable provider collection containing
// zero records without fabricating an opaque navigation checkpoint.
func (*declarativeStreamSourceExecutor) AllowEmptySourceResult() {}

func (e *declarativeStreamSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	outcome, err := e.ReadTransportWithOutcome(ctx, request, emit)
	if err != nil {
		return err
	}
	if outcome.Exhausted {
		return nil
	}
	return &synctransport.SourceBudgetStoppedError{Continuation: *outcome.Continuation.Clone()}
}

func (e *declarativeStreamSourceExecutor) ReadTransportWithOutcome(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) (synctransport.SourceReadOutcome, error) {
	if e == nil || request.Connector == nil {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("declarative stream transport source is unavailable")
	}
	descriptor, ok := connectors.SourceTransportDescriptorOf(request.Connector)
	if !ok || descriptor.Executor != declarativeStreamSourceReference || !transportContainsName(descriptor.EligibleStreams, request.Stream) {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("declarative stream transport source received an undeclared connector or stream")
	}
	if !transportContainsMode(descriptor.Modes, request.Mode) {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("declarative stream transport source does not support sync mode %q", request.Mode)
	}
	if request.BatchSize <= 0 || request.BatchSize > issueCollectionTransportMaxRecords {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("declarative stream transport batch size must be between 1 and %d", issueCollectionTransportMaxRecords)
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("declarative stream transport source requires a complete resume identity")
	}
	if request.Checkpoint != nil {
		if err := request.Checkpoint.ValidateResume(request.Resume); err != nil {
			return synctransport.SourceReadOutcome{}, err
		}
	}
	configuredIssue := strings.TrimSpace(request.Runtime.Config[issueLabelTransportSourceIssueConfig])
	if configuredIssue != "" && request.Stream != "issues" {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("%s is valid only for the issues stream", issueLabelTransportSourceIssueConfig)
	}
	if configuredIssue == "" {
		return e.readDeclarativeCollection(ctx, request.Connector, request, emit)
	}
	if err := e.readConfiguredIssue(ctx, request.Connector, request, emit); err != nil {
		return synctransport.SourceReadOutcome{}, err
	}
	return synctransport.SourceReadOutcome{Exhausted: true}, nil
}

func (e *declarativeStreamSourceExecutor) readConfiguredIssue(ctx context.Context, connector connectors.Connector, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if err := issueLabelTransportRepositoryConfig(request.Runtime.Config); err != nil {
		return err
	}
	if request.BatchSize != 1 {
		return fmt.Errorf("closed issue-label transport requires batch size 1 when %s is configured", issueLabelTransportSourceIssueConfig)
	}
	sourceIssue, err := issueLabelTransportIssueNumber(request.Runtime.Config, issueLabelTransportSourceIssueConfig)
	if err != nil {
		return err
	}

	records := make([]connectors.Record, 0, 1)
	err = connector.Read(ctx, connectors.ReadRequest{
		Stream:           "issues",
		Config:           request.Runtime,
		Limit:            request.BatchSize,
		MaxPages:         issueLabelTransportMaxReadPages,
		PageDeadline:     request.UnitDeadline,
		ObservePageFetch: request.RecordExtraction,
	}, func(record connectors.Record) error {
		number, err := issueNumberFromRecord(record)
		if err != nil {
			return err
		}
		if number != sourceIssue {
			return nil
		}
		cloned, err := cloneTransportRecord(record)
		if err != nil {
			return err
		}
		records = append(records, cloned)
		return connectors.ErrReadLimitReached
	})
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return fmt.Errorf("read configured issue: %w", err)
	}
	if len(records) != 1 {
		return fmt.Errorf("closed issue-label source did not find configured issue %d in its bounded page", sourceIssue)
	}
	candidate, err := issueTransportCheckpoint(request.Resume, records)
	if err != nil {
		return err
	}
	return emit(synctransport.SourcePage{Records: records, CandidateCheckpoint: candidate})
}

// readDeclarativeCollection emits bounded transport pages while the engine
// retains ownership of provider pagination. A persisted candidate is matched
// and suppressed on resume, so acknowledged pages are not re-delivered even
// though the provider sequence must be traversed again to recover its position.
func (e *declarativeStreamSourceExecutor) readDeclarativeCollection(ctx context.Context, connector connectors.Connector, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) (synctransport.SourceReadOutcome, error) {
	maxPages, err := declarativeTransportMaxPages(request.Runtime.Config)
	if err != nil {
		return synctransport.SourceReadOutcome{}, err
	}
	engineConnector, _, err := declarativeStreamTransportConnector(connector)
	if err != nil {
		return synctransport.SourceReadOutcome{}, err
	}
	records := make([]connectors.Record, 0, request.BatchSize)
	pageOrdinal := 0
	// Current-state and history dedupe both derive their source version from the
	// bounded provider page. Replaying that page is the safe way to observe a
	// changed record at the same primary key: their declared warehouse apply
	// strategies collapse an identical replay and retain a distinct version.
	// Suppressing it by an old page hash would instead turn an ordinary provider
	// update into an invalid checkpoint before the destination can compare it.
	waitingForResume := request.Checkpoint != nil && request.Checkpoint.Continuation == nil && !declarativeCollectionReplaysForMode(request.Mode)
	deferCheckpoint := maxPages > 0 && declarativeCollectionIncrementalMode(request.Mode)
	emitBatch := func() error {
		if len(records) == 0 {
			return nil
		}
		pageOrdinal++
		candidate, err := declarativeTransportCheckpoint(request.Resume, request.Stream, pageOrdinal, records)
		if err != nil {
			return err
		}
		if waitingForResume {
			if checkpointPositionEqual(candidate.Position, request.Checkpoint.Position) {
				waitingForResume = false
			}
			records = records[:0]
			return nil
		}
		page := synctransport.SourcePage{Records: append([]connectors.Record(nil), records...), CandidateCheckpoint: candidate, DeferCheckpoint: deferCheckpoint}
		records = records[:0]
		return emit(page)
	}
	var continuation *connectors.ReadContinuation
	if request.Checkpoint != nil && request.Checkpoint.Continuation != nil {
		continuation = &connectors.ReadContinuation{Kind: request.Checkpoint.Continuation.Kind, Token: append([]byte(nil), request.Checkpoint.Continuation.Token...)}
	}
	err = engineConnector.ReadWithOutcome(ctx, connectors.ReadRequest{
		Stream:           request.Stream,
		Config:           request.Runtime,
		MaxPages:         maxPages,
		Continuation:     continuation,
		PageDeadline:     request.UnitDeadline,
		ObservePageFetch: request.RecordExtraction,
	}, func(record connectors.Record) error {
		cloned, err := cloneTransportRecord(record)
		if err != nil {
			return err
		}
		records = append(records, cloned)
		if len(records) == request.BatchSize {
			return emitBatch()
		}
		return nil
	})
	var budgetStop *connectors.ReadBudgetStoppedError
	if err != nil && !errors.As(err, &budgetStop) {
		return synctransport.SourceReadOutcome{}, fmt.Errorf("read declarative stream collection: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return synctransport.SourceReadOutcome{}, err
	}
	if err := emitBatch(); err != nil {
		return synctransport.SourceReadOutcome{}, err
	}
	if waitingForResume {
		return synctransport.SourceReadOutcome{}, synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "declarative stream resume page is no longer present")
	}
	if budgetStop != nil {
		continuation := &synccontract.SourceContinuation{Kind: budgetStop.Continuation.Kind, Token: append(synccontract.OpaqueToken(nil), budgetStop.Continuation.Token...)}
		return synctransport.SourceReadOutcome{Continuation: continuation}, nil
	}
	return synctransport.SourceReadOutcome{Exhausted: true}, nil
}

func declarativeCollectionReplaysForMode(mode synccontract.Mode) bool {
	return mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory
}

func declarativeCollectionIncrementalMode(mode synccontract.Mode) bool {
	switch mode {
	case synccontract.ModeIncrementalAppend, synccontract.ModeIncrementalUpsert, synccontract.ModeIncrementalDedupe, synccontract.ModeIncrementalDedupeHistory:
		return true
	default:
		return false
	}
}

func declarativeTransportMaxPages(config map[string]string) (int, error) {
	raw := strings.TrimSpace(config[declarativeTransportMaxPagesConfig])
	if raw == "" {
		return 1, nil
	}
	switch strings.ToLower(raw) {
	case "0", "all", "unlimited":
		return 0, nil
	}
	pages, err := strconv.Atoi(raw)
	if err != nil || pages <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, 0, all, or unlimited", declarativeTransportMaxPagesConfig)
	}
	return pages, nil
}

func transportContainsName(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func transportContainsMode(values []synccontract.Mode, want synccontract.Mode) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func checkpointPositionEqual(left, right synccontract.CheckpointPosition) bool {
	return bytes.Equal(left.Primary, right.Primary) && bytes.Equal(left.TieBreaker, right.TieBreaker)
}

type issueLabelDestinationExecutor struct {
	app       *App
	connector *engine.Connector
	contract  issueLabelTransportContract
}

func (*issueLabelDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return issueLabelDestinationReference
}

func (e *issueLabelDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	connector, contract, err := e.destinationFor(request.Connector)
	if err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if e == nil || connector == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if !contract.matchesApplyStrategy(request.ApplyStrategy) {
		return synctransport.DestinationPlan{}, &IssueLabelTransportUnsupportedActionError{Action: request.ApplyStrategy.Action}
	}
	if _, err := e.sourceBindingFor(connector, request.Source, request.Stream); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if err := issueLabelTransportRepositoryConfig(request.Runtime.Config); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if _, err := issueLabelTransportIssueNumber(request.Runtime.Config, issueLabelTransportTargetIssueConfig); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if _, err := issueLabelTransportLabel(request.Runtime.Config); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}

func (e *issueLabelDestinationExecutor) destinationFor(candidate connectors.Connector) (*engine.Connector, issueLabelTransportContract, error) {
	if e == nil {
		return nil, issueLabelTransportContract{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if candidate == nil {
		candidate = e.connector
	}
	return issueLabelTransportConnectorContract(candidate)
}

func (e *issueLabelDestinationExecutor) sourceBindingFor(destination *engine.Connector, source connectors.Connector, stream string) (connectors.DestinationSourceBinding, error) {
	if e == nil || destination == nil || source == nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination received an undeclared source")
	}
	sourceDescriptor, ok := connectors.SourceTransportDescriptorOf(source)
	if !ok {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination received a source without a transport declaration")
	}
	descriptor := destination.Definition().SyncTransport
	if descriptor == nil || descriptor.Destination == nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination lost its transport declaration")
	}
	binding, admitted := descriptor.Destination.SourceBindingFor(sourceDescriptor.Executor, stream)
	if !admitted {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination does not admit source executor %q for stream %q", sourceDescriptor.Executor.ID, stream)
	}
	return binding, nil
}

func (e *issueLabelDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	connector, contract, err := e.destinationFor(request.Destination)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if e == nil || e.app == nil || connector == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if !contract.matchesApplyStrategy(request.Plan.ApplyStrategy) {
		return synccontract.DownstreamAcknowledgement{}, &IssueLabelTransportUnsupportedActionError{Action: request.Plan.ApplyStrategy.Action}
	}
	if request.Workset.ID == "" || len(request.Workset.Records) != 1 {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("closed issue-label destination requires exactly one reopened issue record")
	}
	if _, err := e.app.ApplyIssueLabelTransport(ctx, request.ConnectionID, request.Approval, request.Runtime, request.Receipt, request.Workset); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	return synccontract.NewDurableDownstreamAcknowledgement(connector.Name(), time.Now().UTC())
}

func (e *issueLabelDestinationExecutor) ReadBackDestination(ctx context.Context, request synctransport.DestinationReadBackRequest) error {
	connector, contract, err := e.destinationFor(request.Destination)
	if err != nil {
		return err
	}
	if e == nil || e.app == nil || connector == nil {
		return fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Workset.ID == "" {
		return fmt.Errorf("closed issue-label destination read-back received an undeclared receipt")
	}
	if !contract.matchesApplyStrategy(request.Plan.ApplyStrategy) {
		return &IssueLabelTransportUnsupportedActionError{Action: request.Plan.ApplyStrategy.Action}
	}
	action, err := contract.actionForSyncMode(request.Plan.ApplyStrategy.Mode)
	if err != nil {
		return err
	}
	var targetIssue int
	var label string
	if request.Binding.ConnectionID == "" {
		// Focused legacy executor tests construct a read-back request directly.
		// Production always supplies the persisted binding and exercises the
		// reopened-source mapping below.
		targetIssue, err = issueLabelTransportIssueNumber(request.Runtime.Config, issueLabelTransportTargetIssueConfig)
		if err == nil {
			label, err = issueLabelTransportLabel(request.Runtime.Config)
		}
		if err != nil {
			return err
		}
	} else {
		conn, err := e.app.issueLabelTransportConnection(request.Binding.ConnectionID)
		if err != nil {
			return err
		}
		mappedRecord, err := e.app.issueLabelTransportMappedSourceRecord(conn, action, request.Workset.Records[0])
		if err != nil {
			return err
		}
		targetIssue, label, err = action.targetAndLabel(mappedRecord)
		if err != nil {
			return err
		}
	}
	found := false
	exact := request.Plan.ApplyStrategy.Mode == synccontract.ModeFullOverwrite || request.Plan.ApplyStrategy.Mode == synccontract.ModeIncrementalUpsert
	err = connector.Read(ctx, connectors.ReadRequest{
		Stream:   "issues",
		Config:   request.Runtime,
		Limit:    100,
		MaxPages: issueLabelTransportMaxReadPages,
	}, func(record connectors.Record) error {
		number, err := issueNumberFromRecord(record)
		if err != nil {
			return err
		}
		if number == targetIssue {
			found = issueHasLabel(record, label)
			if exact {
				found = issueHasExactlyLabel(record, label)
			}
			return connectors.ErrReadLimitReached
		}
		return nil
	})
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return fmt.Errorf("independently read back issue label: %w", err)
	}
	if !found {
		if exact {
			return fmt.Errorf("closed issue-label destination read-back did not find exact label set %q on issue %d", label, targetIssue)
		}
		return fmt.Errorf("closed issue-label destination read-back did not find label %q on issue %d", label, targetIssue)
	}
	return nil
}

func issueLabelTransportRepositoryConfig(config map[string]string) error {
	if strings.TrimSpace(config["owner"]) == "" || strings.TrimSpace(config["repo"]) == "" {
		return fmt.Errorf("closed issue-label transport requires owner and repo configuration")
	}
	return nil
}

func issueLabelTransportIssueNumber(config map[string]string, key string) (int, error) {
	raw := strings.TrimSpace(config[key])
	if raw == "" {
		return 0, fmt.Errorf("closed issue-label transport requires %s configuration", key)
	}
	number, err := strconv.Atoi(raw)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("closed issue-label transport %s must be a positive issue number", key)
	}
	return number, nil
}

func issueLabelTransportLabel(config map[string]string) (string, error) {
	label := strings.TrimSpace(config[issueLabelTransportLabelConfig])
	if label == "" {
		return "", fmt.Errorf("closed issue-label transport requires %s configuration", issueLabelTransportLabelConfig)
	}
	return label, nil
}

func issueNumberFromRecord(record connectors.Record) (int, error) {
	return issueNumberFromRecordField(record, "number")
}

func issueNumberFromRecordField(record connectors.Record, field string) (int, error) {
	field = strings.TrimSpace(field)
	if field == "" {
		return 0, fmt.Errorf("issue record field is required")
	}
	value, ok := record[field]
	if !ok {
		return 0, fmt.Errorf("issue record has no %s", field)
	}
	switch number := value.(type) {
	case int:
		if number > 0 {
			return number, nil
		}
	case int64:
		if number > 0 && number <= int64(^uint(0)>>1) {
			return int(number), nil
		}
	case float64:
		if number > 0 && number == float64(int(number)) {
			return int(number), nil
		}
	case json.Number:
		parsed, err := number.Int64()
		if err == nil && parsed > 0 && parsed <= int64(^uint(0)>>1) {
			return int(parsed), nil
		}
	case string:
		parsed, err := strconv.Atoi(number)
		if err == nil && parsed > 0 {
			return parsed, nil
		}
	}
	return 0, fmt.Errorf("issue record %s is not a positive integer", field)
}

func issueHasLabel(record connectors.Record, want string) bool {
	labels, ok := record["labels"]
	if !ok {
		return false
	}
	switch values := labels.(type) {
	case []string:
		for _, value := range values {
			if value == want {
				return true
			}
		}
	case []any:
		for _, value := range values {
			switch label := value.(type) {
			case string:
				if label == want {
					return true
				}
			case map[string]any:
				if name, _ := label["name"].(string); name == want {
					return true
				}
			}
		}
	}
	return false
}

func issueHasExactlyLabel(record connectors.Record, want string) bool {
	labels, ok := record["labels"]
	if !ok {
		return false
	}
	count := 0
	found := false
	for _, value := range issueLabelNames(labels) {
		count++
		if value == want {
			found = true
		}
	}
	return found && count == 1
}

func issueLabelNames(labels any) []string {
	var names []string
	switch values := labels.(type) {
	case []string:
		return append([]string(nil), values...)
	case []any:
		for _, value := range values {
			switch label := value.(type) {
			case string:
				names = append(names, label)
			case map[string]any:
				if name, _ := label["name"].(string); name != "" {
					names = append(names, name)
				}
			}
		}
	}
	return names
}

func cloneTransportRecord(record connectors.Record) (connectors.Record, error) {
	clone, err := synctransport.CloneRecord(record)
	if err != nil {
		return nil, fmt.Errorf("clone declarative transport record: %w", err)
	}
	return clone, nil
}

func issueTransportCheckpoint(resume synccontract.ResumeExpectation, records []connectors.Record) (synccontract.CheckpointEnvelope, error) {
	identity, err := hashJSON(records)
	if err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	positionObserved := true
	token := synccontract.OpaqueToken([]byte(identity))
	now := time.Now().UTC()
	checkpoint := synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           resume.Source,
		Mechanism:        "declarative_issues_engine_read",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "issues_page", Token: append(synccontract.OpaqueToken(nil), token...)},
		Position:         synccontract.CheckpointPosition{Primary: append(synccontract.OpaqueToken(nil), token...), TieBreaker: append(synccontract.OpaqueToken(nil), token...)},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: append(synccontract.OpaqueToken(nil), resume.SourceGeneration...),
		SchemaVersion:    "issues-v1",
		ProtocolVersion:  "engine-read-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "issue_page", Value: append(synccontract.OpaqueToken(nil), token...)},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "issue_page", Start: append(synccontract.OpaqueToken(nil), token...), End: append(synccontract.OpaqueToken(nil), token...)},
		ObservedAt:       now,
	}
	if err := checkpoint.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	return checkpoint, nil
}

func declarativeTransportCheckpoint(resume synccontract.ResumeExpectation, stream string, ordinal int, records []connectors.Record) (synccontract.CheckpointEnvelope, error) {
	if ordinal <= 0 {
		return synccontract.CheckpointEnvelope{}, fmt.Errorf("declarative transport checkpoint ordinal must be positive")
	}
	identity, err := hashJSON(struct {
		Stream  string              `json:"stream"`
		Ordinal int                 `json:"ordinal"`
		Records []connectors.Record `json:"records"`
	}{Stream: stream, Ordinal: ordinal, Records: records})
	if err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	positionObserved := true
	ordinalToken := synccontract.OpaqueToken([]byte(fmt.Sprintf("%020d", ordinal)))
	identityToken := synccontract.OpaqueToken([]byte(identity))
	now := time.Now().UTC()
	checkpoint := synccontract.CheckpointEnvelope{
		StateVersion:     synccontract.StateVersion,
		Source:           resume.Source,
		Mechanism:        "declarative_stream_engine_read",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "declarative_page", Token: append(synccontract.OpaqueToken(nil), identityToken...)},
		Position:         synccontract.CheckpointPosition{Primary: append(synccontract.OpaqueToken(nil), ordinalToken...), TieBreaker: append(synccontract.OpaqueToken(nil), identityToken...)},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: append(synccontract.OpaqueToken(nil), resume.SourceGeneration...),
		SchemaVersion:    "declarative-stream-v1",
		ProtocolVersion:  "engine-read-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "declarative_page", Value: append(synccontract.OpaqueToken(nil), identityToken...)},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "declarative_page", Start: append(synccontract.OpaqueToken(nil), identityToken...), End: append(synccontract.OpaqueToken(nil), identityToken...)},
		ObservedAt:       now,
	}
	if err := checkpoint.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	return checkpoint, nil
}
