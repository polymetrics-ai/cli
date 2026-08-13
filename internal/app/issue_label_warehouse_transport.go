package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/synctransport"
)

const (
	issueLabelSourceExecutorID           = "issue_label_source"
	issueLabelDestinationExecutorID      = "issue_label_destination"
	issueLabelEvidenceSuite              = "closed_transport_demo"
	issueLabelSourceEvidenceRun          = "accepted_issue_source"
	issueLabelDestinationEvidenceRun     = "accepted_issue_label_destination"
	issueLabelAddAction                  = "add_issue_labels"
	issueLabelRemoveAction               = "remove_issue_label"
	issueLabelTransportSourceIssueConfig = "transport_source_issue_number"
	issueLabelTransportTargetIssueConfig = "transport_target_issue_number"
	issueLabelTransportLabelConfig       = "transport_label"
)

var (
	issueLabelSourceReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     issueLabelSourceExecutorID,
	}
	issueLabelDestinationReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     issueLabelDestinationExecutorID,
	}
	issueLabelSourceEvidence = connectors.ConformanceEvidenceReference{
		Suite: issueLabelEvidenceSuite,
		RunID: issueLabelSourceEvidenceRun,
	}
	issueLabelDestinationEvidence = connectors.ConformanceEvidenceReference{
		Suite: issueLabelEvidenceSuite,
		RunID: issueLabelDestinationEvidenceRun,
	}
)

// issueLabelTransportConnector embeds the concrete declarative connector,
// preserving every optional engine capability while adding only this walking
// slice's exact transport descriptor. No generic API transport is exposed.
type issueLabelTransportConnector struct{ *engine.Connector }

func (c *issueLabelTransportConnector) Definition() connectors.Definition {
	definition := c.Connector.Definition()
	definition.SyncTransport = issueLabelTransportDescriptor()
	return definition
}

func issueLabelTransportDescriptor() *connectors.SyncTransportDescriptor {
	return &connectors.SyncTransportDescriptor{
		Source: &connectors.SourceTransportDescriptor{
			Executor:        issueLabelSourceReference,
			EligibleStreams: []string{"issues"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: connectors.DeliveryGuarantees{
				Idempotency: connectors.DeliveryIdempotencyKeyed,
				Ordering:    connectors.DeliveryOrderingSource,
				Deletes:     connectors.DeliveryDeletesTombstone,
			},
			Conformance: issueLabelSourceEvidence,
		},
		Destination: &connectors.DestinationTransportDescriptor{
			Executor:        issueLabelDestinationReference,
			EligibleActions: []string{issueLabelAddAction},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: connectors.DeliveryGuarantees{
				Idempotency: connectors.DeliveryIdempotencyKeyed,
				Ordering:    connectors.DeliveryOrderingSource,
				Deletes:     connectors.DeliveryDeletesUnavailable,
			},
			Conformance:     issueLabelDestinationEvidence,
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{
				Mode:     synccontract.ModeFullAppend,
				Strategy: connectors.ApplyStrategyAppend,
				Action:   issueLabelAddAction,
			}},
		},
	}
}

// acceptedIssueLabelEvidenceVerifier is intentionally read-only. Its selected
// connector name is definition-derived at composition time, so shared runtime
// code does not carry a provider policy or self-certify a descriptor.
type acceptedIssueLabelEvidenceVerifier struct{ connectorName string }

func (v acceptedIssueLabelEvidenceVerifier) VerifyTransportConformance(request synctransport.ConformanceVerification) error {
	switch {
	case request.Role == connectors.TransportRoleSource && request.ConnectorName == v.connectorName && request.Executor == issueLabelSourceReference && request.Evidence == issueLabelSourceEvidence:
		return nil
	case request.Role == connectors.TransportRoleDestination && request.ConnectorName == v.connectorName && request.Executor == issueLabelDestinationReference && request.Evidence == issueLabelDestinationEvidence:
		return nil
	default:
		return fmt.Errorf("closed issue-label transport evidence is not accepted for %s %s", request.Role, request.Executor.ID)
	}
}

// newIssueLabelWarehouseMediatedTransport is the sole production composition
// root for the walking slice. It selects the one declarative definition that
// owns the exact issues/add-label/remove-label contract, installs an
// owner-scoped stage and read-only evidence verifier, then registers its fixed
// source and destination roles.
func newIssueLabelWarehouseMediatedTransport(a *App) (*synctransport.Registry, synctransport.WarehouseStage, error) {
	if a == nil || a.registry == nil {
		return nil, nil, fmt.Errorf("closed issue-label transport requires an app registry")
	}
	connector, err := issueLabelTransportEngine(a.registry)
	if err != nil {
		return nil, nil, err
	}
	wrapped := &issueLabelTransportConnector{Connector: connector}
	a.registry.Register(wrapped)

	registry := synctransport.NewRegistry(acceptedIssueLabelEvidenceVerifier{connectorName: connector.Name()})
	if err := registry.RegisterSource(&issueLabelSourceExecutor{connector: connector}); err != nil {
		return nil, nil, err
	}
	if err := registry.RegisterDestination(&issueLabelDestinationExecutor{app: a, connector: connector}); err != nil {
		return nil, nil, err
	}
	return registry, newConnectionWarehouseStage(a), nil
}

// issueLabelTransportEngine resolves the exact existing declarative bundle by
// its closed typed contract. It must remain unique; ambiguity fails closed
// instead of turning the walking slice into a generic connector transport.
func issueLabelTransportEngine(registry *connectors.Registry) (*engine.Connector, error) {
	if registry == nil {
		return nil, fmt.Errorf("closed issue-label transport registry is unavailable")
	}
	var selected *engine.Connector
	for _, metadata := range registry.List() {
		registered, ok := registry.Get(metadata.Name)
		if !ok {
			continue
		}
		var candidate *engine.Connector
		switch connector := registered.(type) {
		case *engine.Connector:
			candidate = connector
		case *issueLabelTransportConnector:
			candidate = connector.Connector
		default:
			continue
		}
		if !definitionSupportsIssueLabelTransport(candidate.Definition()) {
			continue
		}
		if selected != nil && selected.Name() != candidate.Name() {
			return nil, fmt.Errorf("closed issue-label transport contract is ambiguous across declarative connectors")
		}
		selected = candidate
	}
	if selected == nil {
		return nil, fmt.Errorf("closed issue-label transport requires one declarative connector with the exact issue-label actions")
	}
	return selected, nil
}

func definitionSupportsIssueLabelTransport(definition connectors.Definition) bool {
	hasIssues := false
	for _, stream := range definition.Streams {
		if stream.Name == "issues" {
			hasIssues = true
			break
		}
	}
	if !hasIssues {
		return false
	}
	hasAdd, hasRemove := false, false
	for _, action := range definition.WriteActions {
		switch action.Name {
		case issueLabelAddAction:
			hasAdd = true
		case issueLabelRemoveAction:
			hasRemove = true
		}
	}
	return hasAdd && hasRemove
}

type issueLabelSourceExecutor struct{ connector *engine.Connector }

func (*issueLabelSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return issueLabelSourceReference
}

func (e *issueLabelSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if e == nil || e.connector == nil {
		return fmt.Errorf("closed issue-label transport source is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != e.connector.Name() || request.Stream != "issues" || request.Mode != synccontract.ModeFullAppend {
		return fmt.Errorf("closed issue-label transport source accepts only its issues/full_append contract")
	}
	if request.BatchSize <= 0 {
		return fmt.Errorf("closed issue-label transport source requires a positive batch size")
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return fmt.Errorf("closed issue-label transport source requires a complete resume identity")
	}
	if err := issueLabelTransportRepositoryConfig(request.Runtime.Config); err != nil {
		return err
	}
	sourceIssue, err := issueLabelTransportIssueNumber(request.Runtime.Config, issueLabelTransportSourceIssueConfig)
	if err != nil {
		return err
	}

	records := make([]connectors.Record, 0, 1)
	err = e.connector.Read(ctx, connectors.ReadRequest{
		Stream: "issues",
		Config: request.Runtime,
		Limit:  request.BatchSize,
	}, func(record connectors.Record) error {
		number, err := issueNumberFromRecord(record)
		if err != nil {
			return err
		}
		if number != sourceIssue {
			return nil
		}
		if len(records) != 0 {
			return fmt.Errorf("closed issue-label source found duplicate configured issue %d", sourceIssue)
		}
		cloned, err := cloneTransportRecord(record)
		if err != nil {
			return err
		}
		records = append(records, cloned)
		return nil
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

type issueLabelDestinationExecutor struct {
	app       *App
	connector *engine.Connector
}

func (*issueLabelDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return issueLabelDestinationReference
}

func (e *issueLabelDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if e == nil || e.connector == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != e.connector.Name() || request.Stream != "issues" || request.Mode != synccontract.ModeFullAppend || request.ApplyStrategy.Action != issueLabelAddAction || request.ApplyStrategy.Strategy != connectors.ApplyStrategyAppend {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label destination accepts only its issues/add-label contract")
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

func (e *issueLabelDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if e == nil || e.app == nil || e.connector == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Plan.ApplyStrategy.Action != issueLabelAddAction || request.Plan.ApplyStrategy.Strategy != connectors.ApplyStrategyAppend {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("closed issue-label destination received an undeclared apply strategy")
	}
	if request.Workset.ID == "" || len(request.Workset.Records) != 1 {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("closed issue-label destination requires exactly one reopened issue record")
	}
	if _, err := e.app.ApplyIssueLabelTransport(ctx, request.ConnectionID, request.Approval, request.Runtime, request.Receipt, request.Workset); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	return synccontract.NewDurableDownstreamAcknowledgement(e.connector.Name(), time.Now().UTC())
}

func (e *issueLabelDestinationExecutor) ReadBackDestination(ctx context.Context, request synctransport.DestinationReadBackRequest) error {
	if e == nil || e.connector == nil {
		return fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Plan.ApplyStrategy.Action != issueLabelAddAction || request.Workset.ID == "" {
		return fmt.Errorf("closed issue-label destination read-back received an undeclared receipt")
	}
	targetIssue, err := issueLabelTransportIssueNumber(request.Runtime.Config, issueLabelTransportTargetIssueConfig)
	if err != nil {
		return err
	}
	label, err := issueLabelTransportLabel(request.Runtime.Config)
	if err != nil {
		return err
	}
	found := false
	err = e.connector.Read(ctx, connectors.ReadRequest{
		Stream: "issues",
		Config: request.Runtime,
		Limit:  100,
	}, func(record connectors.Record) error {
		number, err := issueNumberFromRecord(record)
		if err != nil {
			return err
		}
		if number == targetIssue {
			found = issueHasLabel(record, label)
		}
		return nil
	})
	if err := connectors.IgnoreReadLimit(err); err != nil {
		return fmt.Errorf("independently read back issue label: %w", err)
	}
	if !found {
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
	value, ok := record["number"]
	if !ok {
		return 0, fmt.Errorf("issue record has no number")
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
	return 0, fmt.Errorf("issue record number is not a positive integer")
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

func cloneTransportRecord(record connectors.Record) (connectors.Record, error) {
	encoded, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("encode issue record: %w", err)
	}
	var clone connectors.Record
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return nil, fmt.Errorf("decode issue record: %w", err)
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
