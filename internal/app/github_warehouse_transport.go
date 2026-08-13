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
	githubIssuesSourceExecutorID          = "github_issues_source"
	githubIssueLabelDestinationExecutorID = "github_issue_label_destination"
	githubIssuesSourceEvidenceSuite       = "github_transport_demo"
	githubIssuesSourceEvidenceRun         = "accepted_github_issues_source"
	githubIssueLabelEvidenceRun           = "accepted_github_issue_label_destination"
	githubIssueAddLabelAction             = "add_issue_labels"
	githubIssueRemoveLabelAction          = "remove_issue_label"
	githubTransportSourceIssueConfig      = "transport_source_issue_number"
	githubTransportTargetIssueConfig      = "transport_target_issue_number"
	githubTransportLabelConfig            = "transport_label"
)

var (
	githubIssuesSourceReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     githubIssuesSourceExecutorID,
	}
	githubIssueLabelDestinationReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     githubIssueLabelDestinationExecutorID,
	}
	githubIssuesSourceEvidence = connectors.ConformanceEvidenceReference{
		Suite: githubIssuesSourceEvidenceSuite,
		RunID: githubIssuesSourceEvidenceRun,
	}
	githubIssueLabelDestinationEvidence = connectors.ConformanceEvidenceReference{
		Suite: githubIssuesSourceEvidenceSuite,
		RunID: githubIssueLabelEvidenceRun,
	}
)

// githubTransportConnector embeds the concrete declarative engine connector,
// preserving every optional engine capability while adding only this walking
// slice's exact transport descriptor. No generic API transport is exposed.
type githubTransportConnector struct{ *engine.Connector }

func (c *githubTransportConnector) Definition() connectors.Definition {
	definition := c.Connector.Definition()
	definition.SyncTransport = githubWarehouseTransportDescriptor()
	return definition
}

func githubWarehouseTransportDescriptor() *connectors.SyncTransportDescriptor {
	return &connectors.SyncTransportDescriptor{
		Source: &connectors.SourceTransportDescriptor{
			Executor:        githubIssuesSourceReference,
			EligibleStreams: []string{"issues"},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: connectors.DeliveryGuarantees{
				Idempotency: connectors.DeliveryIdempotencyKeyed,
				Ordering:    connectors.DeliveryOrderingSource,
				Deletes:     connectors.DeliveryDeletesTombstone,
			},
			Conformance: githubIssuesSourceEvidence,
		},
		Destination: &connectors.DestinationTransportDescriptor{
			Executor:        githubIssueLabelDestinationReference,
			EligibleActions: []string{githubIssueAddLabelAction},
			Modes:           []synccontract.Mode{synccontract.ModeFullAppend},
			Delivery: connectors.DeliveryGuarantees{
				Idempotency: connectors.DeliveryIdempotencyKeyed,
				Ordering:    connectors.DeliveryOrderingSource,
				Deletes:     connectors.DeliveryDeletesUnavailable,
			},
			Conformance:     githubIssueLabelDestinationEvidence,
			Acknowledgement: connectors.TransportAcknowledgementDurableWarehouse,
			ApplyStrategies: []connectors.DestinationApplyStrategy{{
				Mode:     synccontract.ModeFullAppend,
				Strategy: connectors.ApplyStrategyAppend,
				Action:   githubIssueAddLabelAction,
			}},
		},
	}
}

// githubAcceptedEvidenceVerifier is intentionally read-only and accepts only
// the two fixed evidence references this local demonstrable composition owns.
// A descriptor cannot admit itself by merely naming either reference.
type githubAcceptedEvidenceVerifier struct{}

func (githubAcceptedEvidenceVerifier) VerifyTransportConformance(request synctransport.ConformanceVerification) error {
	switch {
	case request.Role == connectors.TransportRoleSource && request.ConnectorName == "github" && request.Executor == githubIssuesSourceReference && request.Evidence == githubIssuesSourceEvidence:
		return nil
	case request.Role == connectors.TransportRoleDestination && request.ConnectorName == "github" && request.Executor == githubIssueLabelDestinationReference && request.Evidence == githubIssueLabelDestinationEvidence:
		return nil
	default:
		return fmt.Errorf("GitHub transport evidence is not accepted for %s %s", request.Role, request.Executor.ID)
	}
}

// newGitHubWarehouseMediatedTransport is the sole production composition
// root for the #4081 walking slice. It installs an owner-scoped stage, a
// read-only accepted-evidence verifier, and exact GitHub source/destination
// registrations; callers get no default or inferred transport pairing.
func newGitHubWarehouseMediatedTransport(a *App) (*synctransport.Registry, synctransport.WarehouseStage, error) {
	if a == nil || a.registry == nil {
		return nil, nil, fmt.Errorf("GitHub warehouse-mediated transport requires an app registry")
	}
	registered, ok := a.registry.Get("github")
	if !ok {
		return nil, nil, fmt.Errorf("GitHub warehouse-mediated transport requires the GitHub connector")
	}
	github, ok := registered.(*engine.Connector)
	if !ok {
		return nil, nil, fmt.Errorf("GitHub warehouse-mediated transport requires the declarative GitHub engine connector")
	}
	wrapped := &githubTransportConnector{Connector: github}
	a.registry.Register(wrapped)

	registry := synctransport.NewRegistry(githubAcceptedEvidenceVerifier{})
	if err := registry.RegisterSource(&githubIssuesSourceExecutor{connector: github}); err != nil {
		return nil, nil, err
	}
	if err := registry.RegisterDestination(&githubIssueLabelDestinationExecutor{connector: github}); err != nil {
		return nil, nil, err
	}
	return registry, newConnectionWarehouseStage(a), nil
}

type githubIssuesSourceExecutor struct{ connector *engine.Connector }

func (*githubIssuesSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return githubIssuesSourceReference
}

func (e *githubIssuesSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if e == nil || e.connector == nil {
		return fmt.Errorf("GitHub issues transport source is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != "github" || request.Stream != "issues" || request.Mode != synccontract.ModeFullAppend {
		return fmt.Errorf("GitHub issues transport source accepts only the exact github/issues/full_append contract")
	}
	if request.BatchSize <= 0 {
		return fmt.Errorf("GitHub issues transport source requires a positive batch size")
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return fmt.Errorf("GitHub issues transport source requires a complete resume identity")
	}
	if err := githubTransportRepositoryConfig(request.Runtime.Config); err != nil {
		return err
	}
	sourceIssue, err := githubTransportIssueNumber(request.Runtime.Config, githubTransportSourceIssueConfig)
	if err != nil {
		return err
	}

	records := make([]connectors.Record, 0, 1)
	err = e.connector.Read(ctx, connectors.ReadRequest{
		Stream: "issues",
		Config: request.Runtime,
		Limit:  request.BatchSize,
	}, func(record connectors.Record) error {
		number, err := githubIssueNumberFromRecord(record)
		if err != nil {
			return err
		}
		if number != sourceIssue {
			return nil
		}
		if len(records) != 0 {
			return fmt.Errorf("GitHub issues transport source found duplicate configured issue %d", sourceIssue)
		}
		cloned, err := cloneGitHubTransportRecord(record)
		if err != nil {
			return err
		}
		records = append(records, cloned)
		return nil
	})
	if err != nil {
		return fmt.Errorf("read configured GitHub issue: %w", err)
	}
	if len(records) != 1 {
		return fmt.Errorf("GitHub issues transport source did not find configured issue %d in its bounded page", sourceIssue)
	}
	candidate, err := githubIssuesCheckpoint(request.Resume, records)
	if err != nil {
		return err
	}
	return emit(synctransport.SourcePage{Records: records, CandidateCheckpoint: candidate})
}

type githubIssueLabelDestinationExecutor struct{ connector *engine.Connector }

func (*githubIssueLabelDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return githubIssueLabelDestinationReference
}

func (e *githubIssueLabelDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if e == nil || e.connector == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("GitHub issue-label transport destination is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != "github" || request.Stream != "issues" || request.Mode != synccontract.ModeFullAppend || request.ApplyStrategy.Action != githubIssueAddLabelAction || request.ApplyStrategy.Strategy != connectors.ApplyStrategyAppend {
		return synctransport.DestinationPlan{}, fmt.Errorf("GitHub issue-label destination accepts only the exact github/issues/add_issue_labels contract")
	}
	if err := githubTransportRepositoryConfig(request.Runtime.Config); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if _, err := githubTransportIssueNumber(request.Runtime.Config, githubTransportTargetIssueConfig); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	if _, err := githubTransportLabel(request.Runtime.Config); err != nil {
		return synctransport.DestinationPlan{}, err
	}
	return synctransport.DestinationPlan{ApplyStrategy: request.ApplyStrategy}, nil
}

func (e *githubIssueLabelDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if e == nil || e.connector == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("GitHub issue-label transport destination is unavailable")
	}
	if request.Plan.ApplyStrategy.Action != githubIssueAddLabelAction || request.Plan.ApplyStrategy.Strategy != connectors.ApplyStrategyAppend {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("GitHub issue-label destination received an undeclared apply strategy")
	}
	if request.Workset.ID == "" || len(request.Workset.Records) != 1 {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("GitHub issue-label destination requires exactly one reopened issue record")
	}
	targetIssue, err := githubTransportIssueNumber(request.Runtime.Config, githubTransportTargetIssueConfig)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	label, err := githubTransportLabel(request.Runtime.Config)
	if err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	if err := e.writeIssueLabel(ctx, request.Runtime, githubIssueAddLabelAction, targetIssue, label); err != nil {
		return synccontract.DownstreamAcknowledgement{}, err
	}
	return synccontract.NewDurableDownstreamAcknowledgement("github", time.Now().UTC())
}

func (e *githubIssueLabelDestinationExecutor) ReadBackDestination(ctx context.Context, request synctransport.DestinationReadBackRequest) error {
	if e == nil || e.connector == nil {
		return fmt.Errorf("GitHub issue-label transport destination is unavailable")
	}
	if request.Plan.ApplyStrategy.Action != githubIssueAddLabelAction || request.Workset.ID == "" {
		return fmt.Errorf("GitHub issue-label destination read-back received an undeclared receipt")
	}
	targetIssue, err := githubTransportIssueNumber(request.Runtime.Config, githubTransportTargetIssueConfig)
	if err != nil {
		return err
	}
	label, err := githubTransportLabel(request.Runtime.Config)
	if err != nil {
		return err
	}
	found := false
	err = e.connector.Read(ctx, connectors.ReadRequest{
		Stream: "issues",
		Config: request.Runtime,
		Limit:  1,
	}, func(record connectors.Record) error {
		number, err := githubIssueNumberFromRecord(record)
		if err != nil {
			return err
		}
		if number == targetIssue {
			found = githubIssueHasLabel(record, label)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("independently read back GitHub issue label: %w", err)
	}
	if !found {
		return fmt.Errorf("GitHub issue-label destination read-back did not find label %q on issue %d", label, targetIssue)
	}
	return nil
}

// RemoveIssueLabel is the typed inverse used by the walking demo cleanup.
// It accepts runtime only for this call and never stores secret/config state.
func (e *githubIssueLabelDestinationExecutor) RemoveIssueLabel(ctx context.Context, runtime connectors.RuntimeConfig, issueNumber int, label string) error {
	return e.writeIssueLabel(ctx, runtime, githubIssueRemoveLabelAction, issueNumber, label)
}

func (e *githubIssueLabelDestinationExecutor) writeIssueLabel(ctx context.Context, runtime connectors.RuntimeConfig, action string, issueNumber int, label string) error {
	if err := githubTransportRepositoryConfig(runtime.Config); err != nil {
		return err
	}
	if issueNumber <= 0 || strings.TrimSpace(label) == "" {
		return fmt.Errorf("GitHub issue-label action requires a positive issue number and non-empty label")
	}
	record := connectors.Record{"issue_number": issueNumber}
	switch action {
	case githubIssueAddLabelAction:
		record["labels"] = []string{label}
	case githubIssueRemoveLabelAction:
		record["name"] = label
	default:
		return fmt.Errorf("GitHub issue-label action %q is not declared", action)
	}
	result, err := e.connector.Write(ctx, connectors.WriteRequest{
		Stream: "issues",
		Table:  "github_transport_issue_label",
		Action: action,
		Config: runtime,
	}, []connectors.Record{record})
	if err != nil {
		return fmt.Errorf("execute typed GitHub %s: %w", action, err)
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		return fmt.Errorf("typed GitHub %s result written=%d failed=%d, want one durable write", action, result.RecordsWritten, result.RecordsFailed)
	}
	return nil
}

func githubTransportRepositoryConfig(config map[string]string) error {
	if strings.TrimSpace(config["owner"]) == "" || strings.TrimSpace(config["repo"]) == "" {
		return fmt.Errorf("GitHub transport requires owner and repo configuration")
	}
	return nil
}

func githubTransportIssueNumber(config map[string]string, key string) (int, error) {
	raw := strings.TrimSpace(config[key])
	if raw == "" {
		return 0, fmt.Errorf("GitHub transport requires %s configuration", key)
	}
	number, err := strconv.Atoi(raw)
	if err != nil || number <= 0 {
		return 0, fmt.Errorf("GitHub transport %s must be a positive issue number", key)
	}
	return number, nil
}

func githubTransportLabel(config map[string]string) (string, error) {
	label := strings.TrimSpace(config[githubTransportLabelConfig])
	if label == "" {
		return "", fmt.Errorf("GitHub transport requires %s configuration", githubTransportLabelConfig)
	}
	return label, nil
}

func githubIssueNumberFromRecord(record connectors.Record) (int, error) {
	value, ok := record["number"]
	if !ok {
		return 0, fmt.Errorf("GitHub issue record has no number")
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
	return 0, fmt.Errorf("GitHub issue record number is not a positive integer")
}

func githubIssueHasLabel(record connectors.Record, want string) bool {
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

func cloneGitHubTransportRecord(record connectors.Record) (connectors.Record, error) {
	encoded, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("encode GitHub issue record: %w", err)
	}
	var clone connectors.Record
	if err := json.Unmarshal(encoded, &clone); err != nil {
		return nil, fmt.Errorf("decode GitHub issue record: %w", err)
	}
	return clone, nil
}

func githubIssuesCheckpoint(resume synccontract.ResumeExpectation, records []connectors.Record) (synccontract.CheckpointEnvelope, error) {
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
		Mechanism:        "github_issues_engine_read",
		SnapshotBarrier:  &synccontract.SnapshotBarrier{Kind: "github_issues_page", Token: append(synccontract.OpaqueToken(nil), token...)},
		Position:         synccontract.CheckpointPosition{Primary: append(synccontract.OpaqueToken(nil), token...), TieBreaker: append(synccontract.OpaqueToken(nil), token...)},
		PositionObserved: &positionObserved,
		Partitions:       []synccontract.PartitionState{},
		SourceGeneration: append(synccontract.OpaqueToken(nil), resume.SourceGeneration...),
		SchemaVersion:    "github-issues-v1",
		ProtocolVersion:  "engine-read-v1",
		Dedupe:           synccontract.DedupeIdentity{Kind: "github_issue_page", Value: append(synccontract.OpaqueToken(nil), token...)},
		DedupeWindow:     synccontract.DedupeWindow{Kind: "github_issue_page", Start: append(synccontract.OpaqueToken(nil), token...), End: append(synccontract.OpaqueToken(nil), token...)},
		ObservedAt:       now,
	}
	if err := checkpoint.Validate(); err != nil {
		return synccontract.CheckpointEnvelope{}, err
	}
	return checkpoint, nil
}
