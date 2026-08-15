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
	issueLabelSourceExecutorID                 = "issue_label_source"
	issueLabelDestinationExecutorID            = "issue_label_destination"
	issueLabelEvidenceSuite                    = "closed_transport_demo"
	issueLabelSourceEvidenceRun                = "accepted_issue_source"
	issueLabelDestinationEvidenceRun           = "accepted_issue_label_destination"
	issueLabelTransportSourceIssueConfig       = "transport_source_issue_number"
	issueLabelTransportTargetIssueConfig       = "transport_target_issue_number"
	issueLabelTransportLabelConfig             = "transport_label"
	issueLabelTransportSetReplaceConsentConfig = "transport_allow_set_replace"
	issueLabelTransportKeyedConsentConfig      = "transport_allow_keyed"
	issueLabelTransportMaxReadPages            = 1
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

// issueLabelTransportDefinitionFactories names only the two GitHub adapter
// hooks. Their selection and evidence admission are performed generically from
// sync_transport.json by synctransport.RegisterDeclaredTransports.
func issueLabelTransportDefinitionFactories(a *App) []synctransport.DefinitionFactory {
	return []synctransport.DefinitionFactory{
		{
			Reference:      issueLabelSourceReference,
			SourceEvidence: issueLabelSourceEvidence,
			BuildSource: func(connector connectors.Connector) (synctransport.SourceExecutor, error) {
				engineConnector, contract, err := issueLabelTransportConnectorContract(connector)
				if err != nil {
					return nil, err
				}
				return &issueLabelSourceExecutor{connector: engineConnector, contract: contract}, nil
			},
		},
		{
			Reference:           issueLabelDestinationReference,
			DestinationEvidence: issueLabelDestinationEvidence,
			BuildDestination: func(connector connectors.Connector) (synctransport.DestinationExecutor, error) {
				if a == nil {
					return nil, fmt.Errorf("closed issue-label transport requires an app")
				}
				engineConnector, contract, err := issueLabelTransportConnectorContract(connector)
				if err != nil {
					return nil, err
				}
				return &issueLabelDestinationExecutor{app: a, connector: engineConnector, contract: contract}, nil
			},
		},
	}
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
	return issueLabelTransportAction{}, fmt.Errorf("closed issue-label transport has no action for sync mode %q", mode)
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

type issueLabelSourceExecutor struct {
	connector *engine.Connector
	contract  issueLabelTransportContract
}

func (*issueLabelSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return issueLabelSourceReference
}

func (e *issueLabelSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if e == nil || e.connector == nil {
		return fmt.Errorf("closed issue-label transport source is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != e.connector.Name() || request.Stream != e.contract.stream {
		return fmt.Errorf("closed issue-label transport source received an undeclared connector or stream")
	}
	if _, err := e.contract.actionForSyncMode(request.Mode); err != nil {
		return fmt.Errorf("closed issue-label transport source does not support sync mode %q", request.Mode)
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
		Stream:   "issues",
		Config:   request.Runtime,
		Limit:    request.BatchSize,
		MaxPages: issueLabelTransportMaxReadPages,
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

type issueLabelDestinationExecutor struct {
	app       *App
	connector *engine.Connector
	contract  issueLabelTransportContract
}

func (*issueLabelDestinationExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return issueLabelDestinationReference
}

func (e *issueLabelDestinationExecutor) PlanDestination(_ context.Context, request synctransport.DestinationPlanRequest) (synctransport.DestinationPlan, error) {
	if e == nil || e.connector == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != e.connector.Name() || request.Stream != e.contract.stream || !e.contract.matchesApplyStrategy(request.ApplyStrategy) {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label destination received an undeclared strategy")
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
	if !e.contract.matchesApplyStrategy(request.Plan.ApplyStrategy) {
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
	if !e.contract.matchesApplyStrategy(request.Plan.ApplyStrategy) || request.Workset.ID == "" {
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
	exact := request.Plan.ApplyStrategy.Mode == synccontract.ModeFullOverwrite || request.Plan.ApplyStrategy.Mode == synccontract.ModeIncrementalUpsert
	err = e.connector.Read(ctx, connectors.ReadRequest{
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
