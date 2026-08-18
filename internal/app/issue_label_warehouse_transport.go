package app

import (
	"bytes"
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
	declarativeStreamSourceExecutorID          = "declarative_stream_source"
	issueLabelDestinationExecutorID            = "issue_label_destination"
	declarativeStreamSourceEvidenceSuite       = "declarative_stream_transport"
	declarativeStreamSourceEvidenceRun         = "all_executable_streams_v1"
	issueLabelDestinationEvidenceSuite         = "closed_transport_demo"
	issueLabelDestinationEvidenceRun           = "accepted_issue_label_destination"
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
	issueLabelDestinationReference = connectors.TransportExecutorReference{
		Family: connectors.TransportExecutorFamilyDeclarativeAPI,
		ID:     issueLabelDestinationExecutorID,
	}
	declarativeStreamSourceEvidence = connectors.ConformanceEvidenceReference{
		Suite: declarativeStreamSourceEvidenceSuite,
		RunID: declarativeStreamSourceEvidenceRun,
	}
	issueLabelDestinationEvidence = connectors.ConformanceEvidenceReference{
		Suite: issueLabelDestinationEvidenceSuite,
		RunID: issueLabelDestinationEvidenceRun,
	}
)

// issueLabelTransportDefinitionFactories names only the two adapter hooks.
// Their selection and evidence admission are performed generically from
// sync_transport.json by synctransport.RegisterDeclaredTransports.
func issueLabelTransportDefinitionFactories(a *App) []synctransport.DefinitionFactory {
	return []synctransport.DefinitionFactory{
		{
			Reference:      declarativeStreamSourceReference,
			SourceEvidence: declarativeStreamSourceEvidence,
			BuildSource: func(connector connectors.Connector) (synctransport.SourceExecutor, error) {
				engineConnector, descriptor, err := declarativeStreamTransportConnector(connector)
				if err != nil {
					return nil, err
				}
				return &declarativeStreamSourceExecutor{connector: engineConnector, descriptor: descriptor}, nil
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
	if descriptor.Executor != declarativeStreamSourceReference || descriptor.Conformance != declarativeStreamSourceEvidence {
		return nil, connectors.SourceTransportDescriptor{}, fmt.Errorf("declarative stream transport requires its exact executor and evidence")
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
	connector  *engine.Connector
	descriptor connectors.SourceTransportDescriptor
}

func (*declarativeStreamSourceExecutor) TransportExecutorReference() connectors.TransportExecutorReference {
	return declarativeStreamSourceReference
}

// AllowEmptySourceResult admits an executable provider collection containing
// zero records without fabricating an opaque navigation checkpoint.
func (*declarativeStreamSourceExecutor) AllowEmptySourceResult() {}

func (e *declarativeStreamSourceExecutor) ReadTransport(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	if e == nil || e.connector == nil {
		return fmt.Errorf("declarative stream transport source is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != e.connector.Name() || !transportContainsName(e.descriptor.EligibleStreams, request.Stream) {
		return fmt.Errorf("declarative stream transport source received an undeclared connector or stream")
	}
	if !transportContainsMode(e.descriptor.Modes, request.Mode) {
		return fmt.Errorf("declarative stream transport source does not support sync mode %q", request.Mode)
	}
	if request.BatchSize <= 0 || request.BatchSize > issueCollectionTransportMaxRecords {
		return fmt.Errorf("declarative stream transport batch size must be between 1 and %d", issueCollectionTransportMaxRecords)
	}
	if err := request.Resume.Source.Validate(); err != nil || len(request.Resume.SourceGeneration) == 0 {
		return fmt.Errorf("declarative stream transport source requires a complete resume identity")
	}
	if request.Checkpoint != nil {
		if err := request.Checkpoint.ValidateResume(request.Resume); err != nil {
			return err
		}
	}
	configuredIssue := strings.TrimSpace(request.Runtime.Config[issueLabelTransportSourceIssueConfig])
	if configuredIssue != "" && request.Stream != "issues" {
		return fmt.Errorf("%s is valid only for the issues stream", issueLabelTransportSourceIssueConfig)
	}
	if configuredIssue == "" {
		return e.readDeclarativeCollection(ctx, request, emit)
	}
	return e.readConfiguredIssue(ctx, request, emit)
}

func (e *declarativeStreamSourceExecutor) readConfiguredIssue(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
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
	err = e.connector.Read(ctx, connectors.ReadRequest{
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
func (e *declarativeStreamSourceExecutor) readDeclarativeCollection(ctx context.Context, request synctransport.SourceRequest, emit func(synctransport.SourcePage) error) error {
	maxPages, err := declarativeTransportMaxPages(request.Runtime.Config)
	if err != nil {
		return err
	}
	records := make([]connectors.Record, 0, request.BatchSize)
	pageOrdinal := 0
	// Current-state and history dedupe both derive their source version from the
	// bounded provider page. Replaying that page is the safe way to observe a
	// changed record at the same primary key: their declared warehouse apply
	// strategies collapse an identical replay and retain a distinct version.
	// Suppressing it by an old page hash would instead turn an ordinary provider
	// update into an invalid checkpoint before the destination can compare it.
	waitingForResume := request.Checkpoint != nil && !declarativeCollectionReplaysForMode(request.Mode)
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
		page := synctransport.SourcePage{Records: append([]connectors.Record(nil), records...), CandidateCheckpoint: candidate}
		records = records[:0]
		return emit(page)
	}
	err = e.connector.Read(ctx, connectors.ReadRequest{
		Stream:           request.Stream,
		Config:           request.Runtime,
		MaxPages:         maxPages,
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
	if err != nil {
		return fmt.Errorf("read declarative stream collection: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := emitBatch(); err != nil {
		return err
	}
	if waitingForResume {
		return synccontract.RequireRebootstrap(synccontract.RecoveryOutcomeInvalidCheckpoint, "declarative stream resume page is no longer present")
	}
	return nil
}

func declarativeCollectionReplaysForMode(mode synccontract.Mode) bool {
	return mode == synccontract.ModeIncrementalDedupe || mode == synccontract.ModeIncrementalDedupeHistory
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
	if e == nil || e.connector == nil {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Connector == nil || request.Connector.Name() != e.connector.Name() {
		return synctransport.DestinationPlan{}, fmt.Errorf("closed issue-label destination received an undeclared strategy")
	}
	if !e.contract.matchesApplyStrategy(request.ApplyStrategy) {
		return synctransport.DestinationPlan{}, &IssueLabelTransportUnsupportedActionError{Action: request.ApplyStrategy.Action}
	}
	if _, err := e.sourceBindingFor(request.Source, request.Stream); err != nil {
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

func (e *issueLabelDestinationExecutor) sourceBindingFor(source connectors.Connector, stream string) (connectors.DestinationSourceBinding, error) {
	if e == nil || e.connector == nil || source == nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination received an undeclared source")
	}
	sourceDescriptor, ok := connectors.SourceTransportDescriptorOf(source)
	if !ok {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination received a source without a transport declaration")
	}
	destination := e.connector.Definition().SyncTransport
	if destination == nil || destination.Destination == nil {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination lost its transport declaration")
	}
	binding, admitted := destination.Destination.SourceBindingFor(sourceDescriptor.Executor, stream)
	if !admitted {
		return connectors.DestinationSourceBinding{}, fmt.Errorf("closed issue-label destination does not admit source executor %q for stream %q", sourceDescriptor.Executor.ID, stream)
	}
	return binding, nil
}

func (e *issueLabelDestinationExecutor) ApplyDestination(ctx context.Context, request synctransport.DestinationApplyRequest) (synccontract.DownstreamAcknowledgement, error) {
	if e == nil || e.app == nil || e.connector == nil {
		return synccontract.DownstreamAcknowledgement{}, fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if !e.contract.matchesApplyStrategy(request.Plan.ApplyStrategy) {
		return synccontract.DownstreamAcknowledgement{}, &IssueLabelTransportUnsupportedActionError{Action: request.Plan.ApplyStrategy.Action}
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
	if e == nil || e.app == nil || e.connector == nil {
		return fmt.Errorf("closed issue-label transport destination is unavailable")
	}
	if request.Workset.ID == "" {
		return fmt.Errorf("closed issue-label destination read-back received an undeclared receipt")
	}
	if !e.contract.matchesApplyStrategy(request.Plan.ApplyStrategy) {
		return &IssueLabelTransportUnsupportedActionError{Action: request.Plan.ApplyStrategy.Action}
	}
	action, err := e.contract.actionForSyncMode(request.Plan.ApplyStrategy.Mode)
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
