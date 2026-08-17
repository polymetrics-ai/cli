package agentcontract

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"polymetrics.ai/internal/certificationcatalog"
)

const (
	certificationGateSchemaVersion = 1
	certificationDecisionProceed   = CertificationGateDecision("PROCEED")
	certificationDecisionRetry     = CertificationGateDecision("RETRY")
	certificationDecisionHalt      = CertificationGateDecision("HALT")
	certificationCheckCommand      = "go run ./cmd/connectorgen certification-matrix --check"
	certificationAllCommand        = "go run ./cmd/connectorgen certification-matrix --all"

	fullParityCredentialScope              = "full_parity"
	observedOperationsCredentialScope      = "observed_operations"
	fullParityCredentialScopeProof         = "full_parity_stage"
	observedOperationsCredentialScopeProof = "protocol_exchanges"
	fullParityCredentialNote               = "Certification used a full-parity credential; a narrower credential exposes a subset of this certified surface."
	observedOperationsCredentialNote       = "Only the credential use documented by this record's protocol exchanges was verified; no broader credential scope is claimed."
	localParquetWarehouse                  = "local_parquet_warehouse"
)

var (
	safeCertificationIdentifier = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)
	sha256Digest                = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

// CertificationGateDecision is deliberately closed. A caller may continue a protected
// transition only after PROCEED; RETRY and HALT are both blocking verdicts.
type CertificationGateDecision string

const (
	CertificationGateProceed = certificationDecisionProceed
	CertificationGateRetry   = certificationDecisionRetry
	CertificationGateHalt    = certificationDecisionHalt
)

// CertificationGateRequest is the adapter-neutral input a Shepherd supplies for a protected
// transition. Inputs are repeated rather than inferred so no adapter can silently select a
// different artifact root.
type CertificationGateRequest struct {
	SchemaVersion int                     `json:"schema_version"`
	Connector     string                  `json:"connector"`
	Transition    string                  `json:"transition"`
	Inputs        CertificationGateInputs `json:"inputs"`
}

// CertificationGateVerdict is the machine-readable, stable result supplied to Shepherd. Every
// blocking fact carries a coordinate; callers must retain it when they request a retry or halt.
type CertificationGateVerdict struct {
	SchemaVersion int                        `json:"schema_version"`
	Connector     string                     `json:"connector"`
	Transition    string                     `json:"transition"`
	Decision      CertificationGateDecision  `json:"decision"`
	Failures      []CertificationGateFailure `json:"failures"`
}

type CertificationGateFailure struct {
	ID         string `json:"id"`
	Class      string `json:"class"`
	CellID     string `json:"cell_id,omitempty"`
	EvidenceID string `json:"evidence_id,omitempty"`
	Message    string `json:"message"`
}

// CertificationGateBlockedError makes the policy boundary explicit for state-machine callers.
// It carries the verdict itself so a caller cannot lose exact RETRY/HALT evidence while handling
// the error.
type CertificationGateBlockedError struct {
	Verdict CertificationGateVerdict
}

func (err *CertificationGateBlockedError) Error() string {
	if len(err.Verdict.Failures) == 0 {
		return fmt.Sprintf("connector certification Shepherd blocked %s for %s", err.Verdict.Transition, err.Verdict.Connector)
	}
	return fmt.Sprintf("connector certification Shepherd %s blocked %s for %s: %s", err.Verdict.Decision, err.Verdict.Transition, err.Verdict.Connector, err.Verdict.Failures[0].ID)
}

// DecodeCertificationGateRequest accepts only the versioned, complete request shape. It does not
// assign defaults: missing adapter-local fields are evaluated as a fail-closed HALT.
func DecodeCertificationGateRequest(raw []byte) (CertificationGateRequest, error) {
	var request CertificationGateRequest
	if err := decodeCertificationJSON(raw, &request); err != nil {
		return CertificationGateRequest{}, fmt.Errorf("decode certification Shepherd request: %w", err)
	}
	return request, nil
}

// EvaluateCertificationGateJSON converts a malformed adapter request into the same HALT-shaped
// output callers use for malformed generated inputs.
func EvaluateCertificationGateJSON(root string, contract *Contract, raw []byte) (CertificationGateVerdict, error) {
	request, err := DecodeCertificationGateRequest(raw)
	if err != nil {
		return certificationGateHalt(CertificationGateRequest{}, "request/decode", "request", err.Error()), nil
	}
	return EvaluateCertificationGate(root, contract, request)
}

// EvaluateCertificationGate is pure and read-only. It reads the versioned generated artifacts
// declared by the canonical contract; it never invokes a provider, loads credentials, creates
// evidence, writes state, or runs a command.
func EvaluateCertificationGate(root string, contract *Contract, request CertificationGateRequest) (CertificationGateVerdict, error) {
	if contract == nil {
		return certificationGateHalt(request, "contract/missing", "contract", "canonical delivery contract is required"), nil
	}
	if err := contract.Validate(); err != nil {
		return certificationGateHalt(request, "contract/invalid", "contract", err.Error()), nil
	}
	gate := contract.CertificationGate
	if request.SchemaVersion != gate.InputSchemaVersion {
		return certificationGateHalt(request, "request/schema_version", "request", fmt.Sprintf("schema_version %d is unsupported", request.SchemaVersion)), nil
	}
	if !safeCertificationID(request.Connector) {
		return certificationGateHalt(request, "request/connector", "request", "connector must be a safe non-empty identifier"), nil
	}
	if !slices.Contains(gate.EnforcedTransitions, request.Transition) {
		return certificationGateHalt(request, "request/transition", "request", fmt.Sprintf("transition %q is not certification-gated", request.Transition)), nil
	}
	if request.Inputs != gate.Inputs {
		return certificationGateHalt(request, "request/inputs", "request", "adapter inputs must exactly match the canonical certification gate inputs"), nil
	}
	if err := request.Inputs.Validate(); err != nil {
		return certificationGateHalt(request, "request/inputs", "request", err.Error()), nil
	}

	status, err := loadStatusArtifact(root, gate)
	if err != nil {
		return certificationGateHalt(request, certificationInputFailureID("status", err), "input", err.Error()), nil
	}
	var capabilities certificationCapabilityMatrix
	var flow certificationFlowMatrix
	if gate.Inputs.CertificationShards != "" {
		capabilities, flow, err = loadCertificationShardMatrices(root, gate, status)
		if err != nil {
			return certificationGateHalt(request, certificationInputFailureID("certification_shards", err), "input", err.Error()), nil
		}
	} else {
		capabilities, err = loadCapabilityMatrix(root, gate)
		if err != nil {
			return certificationGateHalt(request, certificationInputFailureID("capability_matrix", err), "input", err.Error()), nil
		}
		flow, err = loadFlowMatrix(root, gate)
		if err != nil {
			return certificationGateHalt(request, certificationInputFailureID("flow_matrix", err), "input", err.Error()), nil
		}
	}
	evidence, err := loadAcceptedCertificationEvidence(root, gate)
	if err != nil {
		return certificationGateHalt(request, certificationInputFailureID("evidence", err), "input", err.Error()), nil
	}

	if err := validateCapabilityMatrix(capabilities, gate); err != nil {
		return certificationValidationHalt(request, "capability_matrix", err), nil
	}
	if err := validateStatusArtifact(status, gate); err != nil {
		return certificationValidationHalt(request, "status", err), nil
	}
	if err := validateFlowMatrix(flow, gate, capabilities, status); err != nil {
		return certificationValidationHalt(request, "flow_matrix", err), nil
	}

	artifacts := certificationArtifacts{
		capabilities: capabilities,
		flow:         flow,
		status:       status,
		evidence:     evidence,
	}
	if halt, ok := artifacts.validateDerivedReports(request, gate); ok {
		return halt, nil
	}
	if halt, ok := artifacts.validateConnectorPresence(request.Connector); ok {
		halt.Connector = request.Connector
		halt.Transition = request.Transition
		return halt, nil
	}

	retries := make([]CertificationGateFailure, 0)
	if halt, ok := artifacts.evaluateCapability(request.Connector, gate, &retries); ok {
		halt.Connector = request.Connector
		halt.Transition = request.Transition
		return halt, nil
	}
	if halt, ok := artifacts.evaluateWorkflows(request.Connector, gate, &retries); ok {
		halt.Connector = request.Connector
		halt.Transition = request.Transition
		return halt, nil
	}
	if halt, ok := artifacts.evaluateSyncModes(request.Connector, gate, &retries); ok {
		halt.Connector = request.Connector
		halt.Transition = request.Transition
		return halt, nil
	}
	if halt, ok := artifacts.evaluateFlowPairs(request.Connector, gate, &retries); ok {
		halt.Connector = request.Connector
		halt.Transition = request.Transition
		return halt, nil
	}
	if halt, ok := artifacts.evaluateStatus(request.Connector, &retries); ok {
		halt.Connector = request.Connector
		halt.Transition = request.Transition
		return halt, nil
	}

	sortCertificationFailures(retries)
	decision := CertificationGateProceed
	if len(retries) != 0 {
		decision = CertificationGateRetry
	}
	return CertificationGateVerdict{
		SchemaVersion: gate.VerdictSchemaVersion,
		Connector:     request.Connector,
		Transition:    request.Transition,
		Decision:      decision,
		Failures:      retries,
	}, nil
}

// EnforceCertificationGate is the transition boundary used by a Shepherd or parent workflow.
// A non-PROCEED verdict is returned and blocks the caller before it can integrate, accept, or
// mark a connector/parent ready.
func EnforceCertificationGate(root string, contract *Contract, request CertificationGateRequest) (CertificationGateVerdict, error) {
	verdict, err := EvaluateCertificationGate(root, contract, request)
	if err != nil {
		return verdict, err
	}
	if verdict.Decision != CertificationGateProceed {
		return verdict, &CertificationGateBlockedError{Verdict: verdict}
	}
	return verdict, nil
}

func certificationGateHalt(request CertificationGateRequest, id, class, message string) CertificationGateVerdict {
	return certificationGateHaltAt(request, id, class, "", "", message)
}

func certificationGateHaltAt(request CertificationGateRequest, id, class, cellID, evidenceID, message string) CertificationGateVerdict {
	return CertificationGateVerdict{
		SchemaVersion: certificationGateSchemaVersion,
		Connector:     request.Connector,
		Transition:    request.Transition,
		Decision:      CertificationGateHalt,
		Failures: []CertificationGateFailure{{
			ID:         id,
			Class:      class,
			CellID:     cellID,
			EvidenceID: evidenceID,
			Message:    message,
		}},
	}
}

func certificationInputFailureID(input string, err error) string {
	if errors.Is(err, fs.ErrNotExist) {
		return "input/" + input + "/missing"
	}
	return "input/" + input + "/decode"
}

func certificationValidationHalt(request CertificationGateRequest, input string, err error) CertificationGateVerdict {
	var pointerError *certificationInvalidPointerError
	if errors.As(err, &pointerError) {
		return certificationGateHaltAt(request, "evidence/invalid_pointer", "evidence", pointerError.cellID, pointerError.evidenceID, "invalid_pointer")
	}
	var cellError *certificationInvalidCellError
	if errors.As(err, &cellError) {
		return certificationGateHaltAt(request, cellError.id, "input", cellError.cellID, "", cellError.message)
	}
	return certificationGateHalt(request, input+"/invalid", "input", err.Error())
}

func sortCertificationFailures(failures []CertificationGateFailure) {
	sort.Slice(failures, func(left, right int) bool {
		if failures[left].ID != failures[right].ID {
			return failures[left].ID < failures[right].ID
		}
		if failures[left].EvidenceID != failures[right].EvidenceID {
			return failures[left].EvidenceID < failures[right].EvidenceID
		}
		return failures[left].Message < failures[right].Message
	})
}

func decodeCertificationJSON(raw []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

type certificationArtifacts struct {
	capabilities certificationCapabilityMatrix
	flow         certificationFlowMatrix
	status       certificationStatusArtifact
	evidence     map[string]certificationAcceptedEvidence
}

type certificationEvidenceCell struct {
	cellID  string
	binding certificationEvidenceBinding
	facts   certificationFacts
}

type certificationEvidenceBindingError struct {
	id         string
	cellID     string
	evidenceID string
	message    string
}

func (err *certificationEvidenceBindingError) Error() string {
	return err.message
}

type certificationDerivedReports struct {
	capabilityComplete map[string]bool
	workflowComplete   map[string]bool
	syncModeComplete   map[string]bool
	flowComplete       map[string]bool
	statuses           map[string]certificationConnectorStatus
	capabilityBaseline certificationCapabilityBaseline
	flowBaseline       certificationFlowBaseline
}

func (artifacts certificationArtifacts) validateDerivedReports(request CertificationGateRequest, gate ConnectorCertificationGate) (CertificationGateVerdict, bool) {
	if err := artifacts.validateEvidenceBindings(gate); err != nil {
		return certificationEvidenceHalt(request, err), true
	}
	reports := artifacts.deriveReports()
	for _, connector := range certificationCapabilityConnectorNames(artifacts.capabilities.Connectors) {
		capability, _ := findCapabilityConnector(artifacts.capabilities.Connectors, connector)
		if capability.CapabilityComplete != reports.capabilityComplete[connector] {
			return certificationGateHaltAt(request, "capability/"+connector+"/capability_complete", "input", "capability/"+connector, "", "capability_complete disagrees with matched evidence"), true
		}
		workflow, _ := findWorkflowSet(artifacts.flow.Workflows, connector)
		if workflow.Complete != reports.workflowComplete[connector] {
			return certificationGateHaltAt(request, "workflow/"+connector+"/complete", "input", "workflow/"+connector, "", "workflow complete disagrees with matched evidence"), true
		}
		syncModes, _ := findSyncSet(artifacts.flow.SyncModeCells, connector)
		if syncModes.Complete != reports.syncModeComplete[connector] {
			return certificationGateHaltAt(request, "sync_mode/"+connector+"/complete", "input", "sync_mode/"+connector, "", "sync-mode complete disagrees with matched evidence"), true
		}
		expected := reports.statuses[connector]
		flowStatus, _ := findStatus(artifacts.flow.ConnectorStatuses, connector)
		if flowStatus != expected {
			return certificationGateHaltAt(request, "flow_status/"+connector+"/derived_mismatch", "input", "flow_status/"+connector, "", "flow connector status disagrees with matched evidence"), true
		}
		status, _ := findStatus(artifacts.status.Connectors, connector)
		if status != expected {
			return certificationGateHaltAt(request, "status/"+connector+"/derived_mismatch", "input", "status/"+connector, "", "status artifact disagrees with matched evidence"), true
		}
	}
	if !reflect.DeepEqual(artifacts.capabilities.Baseline, reports.capabilityBaseline) {
		return certificationGateHalt(request, "capability/baseline/derived_mismatch", "input", "capability baseline disagrees with matched evidence"), true
	}
	if !reflect.DeepEqual(artifacts.flow.Baseline, reports.flowBaseline) {
		return certificationGateHalt(request, "flow/baseline/derived_mismatch", "input", "flow baseline disagrees with matched evidence"), true
	}
	return CertificationGateVerdict{}, false
}

func certificationCapabilityConnectorNames(connectors []certificationCapabilityConnector) []string {
	names := make([]string, 0, len(connectors))
	for _, connector := range connectors {
		names = append(names, connector.Name)
	}
	sort.Strings(names)
	return names
}

func certificationFlowConnectorNames(roles []certificationConnectorRoles) []string {
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		names = append(names, role.Connector)
	}
	sort.Strings(names)
	return names
}

func (artifacts certificationArtifacts) validateEvidenceBindings(gate ConnectorCertificationGate) error {
	cells, err := artifacts.evidenceCells()
	if err != nil {
		return err
	}
	for _, cell := range cells {
		if err := artifacts.validateCellEvidence(cell, gate); err != nil {
			return err
		}
	}
	return nil
}

func (artifacts certificationArtifacts) evidenceCells() ([]certificationEvidenceCell, error) {
	cells := make([]certificationEvidenceCell, 0)
	for _, connector := range artifacts.capabilities.Connectors {
		for _, cell := range connector.Cells {
			cells = append(cells, certificationEvidenceCell{
				cellID:  "capability/" + connector.Name + "/" + cell.FunctionKind,
				binding: certificationEvidenceBinding{Scope: "capability", Connector: connector.Name, FunctionKind: cell.FunctionKind},
				facts:   cell.certificationFacts,
			})
		}
	}
	for _, set := range artifacts.flow.Workflows {
		for _, cell := range set.Cells {
			cells = append(cells, certificationEvidenceCell{
				cellID:  "workflow/" + set.Connector + "/" + cell.WorkflowKind,
				binding: certificationEvidenceBinding{Scope: "workflow", Connector: set.Connector, WorkflowKind: cell.WorkflowKind},
				facts:   cell.certificationFacts,
			})
		}
	}
	for _, set := range artifacts.flow.SyncModeCells {
		for _, cell := range set.Cells {
			cells = append(cells, certificationEvidenceCell{
				cellID:  "sync_mode/" + set.Connector + "/" + cell.SyncMode + "/" + cell.Primitive,
				binding: certificationEvidenceBinding{Scope: "sync_mode", Connector: set.Connector, SyncMode: cell.SyncMode, Primitive: cell.Primitive},
				facts:   cell.certificationFacts,
			})
		}
	}
	for _, set := range artifacts.flow.PairSets {
		for _, source := range set.SourceConnectors {
			for _, destination := range set.DestinationConnectors {
				cells = append(cells, certificationEvidenceCell{
					cellID:  "flow/" + set.FlowKind + "/" + source + "/" + destination,
					binding: certificationEvidenceBinding{Scope: "flow", Source: source, Destination: destination, FlowKind: set.FlowKind},
					facts:   set.Cell.certificationFacts,
				})
			}
		}
	}
	for _, override := range artifacts.flow.PairOverrides {
		cells = append(cells, certificationEvidenceCell{
			cellID:  "flow/" + override.FlowKind + "/" + override.Source + "/" + override.Destination,
			binding: certificationEvidenceBinding{Scope: "flow", Source: override.Source, Destination: override.Destination, FlowKind: override.FlowKind},
			facts:   override.Cell.certificationFacts,
		})
	}
	sort.SliceStable(cells, func(left, right int) bool { return cells[left].cellID < cells[right].cellID })
	return cells, nil
}

func (artifacts certificationArtifacts) validateCellEvidence(cell certificationEvidenceCell, gate ConnectorCertificationGate) error {
	if len(cell.facts.LiveEvidence) == 0 {
		return nil
	}
	seen := make(map[string]bool, len(cell.facts.LiveEvidence))
	for _, pointer := range cell.facts.LiveEvidence {
		evidenceID := ""
		if safeEvidenceRecordPath(pointer.Record, gate.Inputs.EvidenceDirectory) {
			evidenceID = pointer.Record
		}
		if evidenceID == "" || seen[pointer.Record] {
			return &certificationInvalidPointerError{cellID: cell.cellID, evidenceID: evidenceID}
		}
		seen[pointer.Record] = true
		if err := validateEvidencePointer(pointer, gate); err != nil {
			return &certificationInvalidPointerError{cellID: cell.cellID, evidenceID: evidenceID}
		}
		record, ok := artifacts.evidence[evidenceID]
		if !ok {
			return &certificationEvidenceBindingError{id: "evidence/" + evidenceID + "/missing", cellID: cell.cellID, evidenceID: evidenceID, message: "referenced live evidence record is missing"}
		}
		if !evidencePointerMatchesRecord(pointer, record) {
			return &certificationEvidenceBindingError{id: "evidence/" + evidenceID + "/mismatch", cellID: cell.cellID, evidenceID: evidenceID, message: "matrix pointer and accepted evidence record differ"}
		}
		if err := cell.binding.matches(record); err != nil {
			return &certificationEvidenceBindingError{id: "evidence/" + evidenceID + "/binding", cellID: cell.cellID, evidenceID: evidenceID, message: err.Error()}
		}
	}
	return nil
}

func certificationEvidenceHalt(request CertificationGateRequest, err error) CertificationGateVerdict {
	var pointerError *certificationInvalidPointerError
	if errors.As(err, &pointerError) {
		return certificationGateHaltAt(request, "evidence/invalid_pointer", "evidence", pointerError.cellID, pointerError.evidenceID, "invalid_pointer")
	}
	var bindingError *certificationEvidenceBindingError
	if errors.As(err, &bindingError) {
		return certificationGateHaltAt(request, bindingError.id, "evidence", bindingError.cellID, bindingError.evidenceID, bindingError.message)
	}
	return certificationGateHalt(request, "evidence/invalid", "evidence", "invalid evidence binding")
}

func (artifacts certificationArtifacts) deriveReports() certificationDerivedReports {
	reports := certificationDerivedReports{
		capabilityComplete: make(map[string]bool, len(artifacts.capabilities.Connectors)),
		workflowComplete:   make(map[string]bool, len(artifacts.flow.Workflows)),
		syncModeComplete:   make(map[string]bool, len(artifacts.flow.SyncModeCells)),
		flowComplete:       make(map[string]bool, len(artifacts.flow.ConnectorRoles)),
		statuses:           make(map[string]certificationConnectorStatus, len(artifacts.capabilities.Connectors)),
	}
	for _, connector := range artifacts.capabilities.Connectors {
		reports.capabilityComplete[connector.Name] = certificationCapabilityCellsComplete(connector.Cells)
	}
	for _, set := range artifacts.flow.Workflows {
		reports.workflowComplete[set.Connector] = certificationWorkflowCellsComplete(set.Cells)
	}
	for _, set := range artifacts.flow.SyncModeCells {
		reports.syncModeComplete[set.Connector] = certificationSyncModeCellsComplete(set.Cells)
	}
	for _, connector := range artifacts.flow.ConnectorRoles {
		reports.flowComplete[connector.Connector] = certificationConnectorFlowComplete(connector.Connector, artifacts.flow)
	}
	for _, connector := range artifacts.capabilities.Connectors {
		certified := reports.capabilityComplete[connector.Name] && reports.workflowComplete[connector.Name] && reports.syncModeComplete[connector.Name] && reports.flowComplete[connector.Name]
		reports.statuses[connector.Name] = certificationDerivedStatus(connector.Name, certified)
	}
	reports.capabilityBaseline = deriveCertificationCapabilityBaseline(artifacts.capabilities, reports.capabilityComplete)
	reports.flowBaseline = deriveCertificationFlowBaseline(artifacts.flow, reports.statuses)
	return reports
}

func certificationCapabilityCellsComplete(cells []certificationCapabilityCell) bool {
	applicable := 0
	for _, cell := range cells {
		if !cell.Applicable {
			continue
		}
		applicable++
		if !certificationFactsComplete(cell.certificationFacts) {
			return false
		}
	}
	return applicable != 0
}

func certificationDerivedStatus(connector string, certified bool) certificationConnectorStatus {
	status := certificationConnectorStatus{Connector: connector, Certified: certified}
	if certified {
		status.Label = "CERTIFIED"
		return status
	}
	status.Label = "COMMUNITY BUILD, UNCERTIFIED"
	status.Warning = "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED."
	return status
}

func deriveCertificationCapabilityBaseline(matrix certificationCapabilityMatrix, complete map[string]bool) certificationCapabilityBaseline {
	baseline := certificationCapabilityBaseline{
		Connectors: len(matrix.Connectors),
		PerKind:    make([]certificationKindBaseline, 0, len(matrix.FunctionKinds)),
	}
	for _, connector := range matrix.Connectors {
		if complete[connector.Name] {
			baseline.CapabilityComplete++
		}
	}
	for _, kind := range matrix.FunctionKinds {
		totals := certificationKindBaseline{FunctionKind: kind.ID, Connectors: len(matrix.Connectors)}
		for _, connector := range matrix.Connectors {
			for _, cell := range connector.Cells {
				if cell.FunctionKind != kind.ID || !cell.Applicable {
					continue
				}
				totals.Applicable++
				if cell.Declared {
					totals.Declared++
				}
				if cell.Implemented {
					totals.Implemented++
				}
				if cell.FixtureTested {
					totals.FixtureTested++
				}
				if cell.LiveTested {
					totals.LiveTested++
				}
				if certificationFactsComplete(cell.certificationFacts) {
					totals.Complete++
				}
			}
		}
		baseline.PerKind = append(baseline.PerKind, totals)
	}
	return baseline
}

func deriveCertificationFlowBaseline(matrix certificationFlowMatrix, statuses map[string]certificationConnectorStatus) certificationFlowBaseline {
	baseline := certificationFlowBaseline{
		Connectors: len(matrix.ConnectorRoles),
		Workflows:  deriveCertificationWorkflowBaseline(matrix.Workflows, matrix.WorkflowKinds),
		SyncModes:  deriveCertificationSyncModeBaseline(matrix.SyncModeCells, matrix.SyncModeKinds, matrix.SyncPrimitives),
		PerKind:    make([]certificationFlowBaselineKind, 0, len(matrix.FlowKinds)),
	}
	connectors := certificationFlowConnectorNames(matrix.ConnectorRoles)
	for _, kind := range matrix.FlowKinds {
		totals := certificationFlowBaselineKind{FlowKind: kind.ID}
		for _, source := range connectors {
			for _, destination := range connectors {
				pair, ok := matrix.resolvedFlowPair(kind.ID, source, destination)
				if !ok {
					continue
				}
				addCertificationFlowCellTotals(&totals, pair.Cell)
			}
		}
		baseline.PerKind = append(baseline.PerKind, totals)
	}
	for _, status := range statuses {
		if status.Certified {
			baseline.Certified++
		}
	}
	return baseline
}

func deriveCertificationWorkflowBaseline(sets []certificationWorkflowSet, kinds []certificationWorkflowKind) []certificationWorkflowBaseline {
	baseline := make([]certificationWorkflowBaseline, 0, len(kinds))
	for _, kind := range kinds {
		totals := certificationWorkflowBaseline{WorkflowKind: kind.ID, Connectors: len(sets)}
		for _, set := range sets {
			for _, cell := range set.Cells {
				if cell.WorkflowKind != kind.ID || !cell.Applicable {
					continue
				}
				totals.Applicable++
				if cell.Declared {
					totals.Declared++
				}
				if cell.Implemented {
					totals.Implemented++
				}
				if cell.FixtureTested {
					totals.FixtureTested++
				}
				if cell.LiveTested {
					totals.LiveTested++
				}
				if certificationFactsComplete(cell.certificationFacts) {
					totals.Complete++
				}
			}
		}
		baseline = append(baseline, totals)
	}
	return baseline
}

func deriveCertificationSyncModeBaseline(sets []certificationSyncModeSet, modes []certificationSyncModeKind, primitives []certificationSyncPrimitive) []certificationSyncModeBaseline {
	baseline := make([]certificationSyncModeBaseline, 0, len(modes)*len(primitives))
	for _, mode := range modes {
		for _, primitive := range primitives {
			totals := certificationSyncModeBaseline{SyncMode: mode.ID, Primitive: primitive.ID, Connectors: len(sets)}
			for _, set := range sets {
				for _, cell := range set.Cells {
					if cell.SyncMode != mode.ID || cell.Primitive != primitive.ID || !cell.Applicable {
						continue
					}
					totals.Applicable++
					if cell.Declared {
						totals.Declared++
					}
					if cell.Implemented {
						totals.Implemented++
					}
					if cell.FixtureTested {
						totals.FixtureTested++
					}
					if cell.LiveTested {
						totals.LiveTested++
					}
					if certificationFactsComplete(cell.certificationFacts) {
						totals.Complete++
					}
				}
			}
			baseline = append(baseline, totals)
		}
	}
	return baseline
}

func addCertificationFlowCellTotals(totals *certificationFlowBaselineKind, cell certificationFlowCell) {
	totals.Pairs++
	if !cell.Applicable {
		return
	}
	totals.Applicable++
	if cell.Declared {
		totals.Declared++
	}
	if cell.Implemented {
		totals.Implemented++
	}
	if cell.FixtureTested {
		totals.FixtureTested++
	}
	if cell.LiveTested {
		totals.LiveTested++
	}
	if certificationFactsComplete(cell.certificationFacts) {
		totals.Complete++
	}
}

func certificationConnectorFlowComplete(connector string, matrix certificationFlowMatrix) bool {
	foundApplicable := false
	for _, kind := range matrix.FlowKinds {
		for _, roles := range matrix.ConnectorRoles {
			for _, pairIdentity := range [][2]string{{connector, roles.Connector}, {roles.Connector, connector}} {
				pair, ok := matrix.resolvedFlowPair(kind.ID, pairIdentity[0], pairIdentity[1])
				if !ok || !pair.Cell.Applicable {
					continue
				}
				foundApplicable = true
				if !certificationFactsComplete(pair.Cell.certificationFacts) {
					return false
				}
			}
		}
	}
	return foundApplicable
}

func (artifacts certificationArtifacts) validateConnectorPresence(connector string) (CertificationGateVerdict, bool) {
	if _, ok := findCapabilityConnector(artifacts.capabilities.Connectors, connector); !ok {
		return certificationGateHalt(CertificationGateRequest{}, "capability/"+connector+"/missing", "input", "connector is absent from capability matrix"), true
	}
	if _, ok := findWorkflowSet(artifacts.flow.Workflows, connector); !ok {
		return certificationGateHalt(CertificationGateRequest{}, "workflow/"+connector+"/missing", "input", "connector is absent from workflow matrix"), true
	}
	if _, ok := findSyncSet(artifacts.flow.SyncModeCells, connector); !ok {
		return certificationGateHalt(CertificationGateRequest{}, "sync_mode/"+connector+"/missing", "input", "connector is absent from sync-mode matrix"), true
	}
	if _, ok := findConnectorRoles(artifacts.flow.ConnectorRoles, connector); !ok {
		return certificationGateHalt(CertificationGateRequest{}, "flow/"+connector+"/roles_missing", "input", "connector is absent from flow roles"), true
	}
	if _, ok := findStatus(artifacts.status.Connectors, connector); !ok {
		return certificationGateHalt(CertificationGateRequest{}, "status/"+connector+"/missing", "input", "connector is absent from status artifact"), true
	}
	if _, ok := findStatus(artifacts.flow.ConnectorStatuses, connector); !ok {
		return certificationGateHalt(CertificationGateRequest{}, "flow_status/"+connector+"/missing", "input", "connector is absent from flow status"), true
	}
	return CertificationGateVerdict{}, false
}

func (artifacts certificationArtifacts) evaluateCapability(connector string, gate ConnectorCertificationGate, retries *[]CertificationGateFailure) (CertificationGateVerdict, bool) {
	entry, _ := findCapabilityConnector(artifacts.capabilities.Connectors, connector)
	known := make(map[string]bool, len(artifacts.capabilities.FunctionKinds))
	for _, kind := range artifacts.capabilities.FunctionKinds {
		known[kind.ID] = true
	}
	seen := make(map[string]bool, len(entry.Cells))
	for _, cell := range entry.Cells {
		cellID := "capability/" + connector + "/" + cell.FunctionKind
		if !known[cell.FunctionKind] {
			return certificationGateHalt(CertificationGateRequest{}, cellID+"/unknown", "input", "cell names an unknown capability kind"), true
		}
		if seen[cell.FunctionKind] {
			return certificationGateHalt(CertificationGateRequest{}, cellID+"/duplicate", "input", "capability cell is duplicated"), true
		}
		seen[cell.FunctionKind] = true
		if halt, ok := artifacts.evaluateCell(cellID, certificationEvidenceBinding{Scope: "capability", Connector: connector, FunctionKind: cell.FunctionKind}, cell.certificationFacts, gate, retries); ok {
			return halt, true
		}
	}
	for _, kind := range artifacts.capabilities.FunctionKinds {
		if !seen[kind.ID] {
			return certificationGateHalt(CertificationGateRequest{}, "capability/"+connector+"/"+kind.ID+"/missing", "input", "capability matrix omits a declared kind"), true
		}
	}
	return CertificationGateVerdict{}, false
}

func (artifacts certificationArtifacts) evaluateWorkflows(connector string, gate ConnectorCertificationGate, retries *[]CertificationGateFailure) (CertificationGateVerdict, bool) {
	entry, _ := findWorkflowSet(artifacts.flow.Workflows, connector)
	known := make(map[string]bool, len(artifacts.flow.WorkflowKinds))
	for _, kind := range artifacts.flow.WorkflowKinds {
		known[kind.ID] = true
	}
	seen := make(map[string]bool, len(entry.Cells))
	for _, cell := range entry.Cells {
		cellID := "workflow/" + connector + "/" + cell.WorkflowKind
		if !known[cell.WorkflowKind] {
			return certificationGateHalt(CertificationGateRequest{}, cellID+"/unknown", "input", "cell names an unknown workflow kind"), true
		}
		if seen[cell.WorkflowKind] {
			return certificationGateHalt(CertificationGateRequest{}, cellID+"/duplicate", "input", "workflow cell is duplicated"), true
		}
		seen[cell.WorkflowKind] = true
		if halt, ok := artifacts.evaluateCell(cellID, certificationEvidenceBinding{Scope: "workflow", Connector: connector, WorkflowKind: cell.WorkflowKind}, cell.certificationFacts, gate, retries); ok {
			return halt, true
		}
	}
	for _, kind := range artifacts.flow.WorkflowKinds {
		if !seen[kind.ID] {
			return certificationGateHalt(CertificationGateRequest{}, "workflow/"+connector+"/"+kind.ID+"/missing", "input", "workflow matrix omits a declared kind"), true
		}
	}
	return CertificationGateVerdict{}, false
}

func (artifacts certificationArtifacts) evaluateSyncModes(connector string, gate ConnectorCertificationGate, retries *[]CertificationGateFailure) (CertificationGateVerdict, bool) {
	entry, _ := findSyncSet(artifacts.flow.SyncModeCells, connector)
	modes := make(map[string]bool, len(artifacts.flow.SyncModeKinds))
	for _, mode := range artifacts.flow.SyncModeKinds {
		modes[mode.ID] = true
	}
	primitives := make(map[string]bool, len(artifacts.flow.SyncPrimitives))
	for _, primitive := range artifacts.flow.SyncPrimitives {
		primitives[primitive.ID] = true
	}
	seen := make(map[string]bool, len(entry.Cells))
	for _, cell := range entry.Cells {
		cellID := "sync_mode/" + connector + "/" + cell.SyncMode + "/" + cell.Primitive
		if !modes[cell.SyncMode] || !primitives[cell.Primitive] {
			return certificationGateHalt(CertificationGateRequest{}, cellID+"/unknown", "input", "sync-mode cell names an unknown mode or primitive"), true
		}
		key := cell.SyncMode + "\x00" + cell.Primitive
		if seen[key] {
			return certificationGateHalt(CertificationGateRequest{}, cellID+"/duplicate", "input", "sync-mode cell is duplicated"), true
		}
		seen[key] = true
		if halt, ok := artifacts.evaluateCell(cellID, certificationEvidenceBinding{Scope: "sync_mode", Connector: connector, SyncMode: cell.SyncMode, Primitive: cell.Primitive}, cell.certificationFacts, gate, retries); ok {
			return halt, true
		}
	}
	if len(entry.Cells) == 0 {
		return certificationGateHalt(CertificationGateRequest{}, "sync_mode/"+connector+"/missing", "input", "sync-mode matrix has no cells"), true
	}
	return CertificationGateVerdict{}, false
}

func (artifacts certificationArtifacts) evaluateFlowPairs(connector string, gate ConnectorCertificationGate, retries *[]CertificationGateFailure) (CertificationGateVerdict, bool) {
	pairs, err := artifacts.flow.flowPairsForConnector(connector)
	if err != nil {
		return certificationGateHalt(CertificationGateRequest{}, "flow/"+connector+"/invalid", "input", err.Error()), true
	}
	roles, _ := findConnectorRoles(artifacts.flow.ConnectorRoles, connector)
	hasApplicableRole := false
	for _, role := range roles.Roles {
		if role.Applicable {
			hasApplicableRole = true
			break
		}
	}
	if hasApplicableRole && len(pairs) == 0 {
		return certificationGateHalt(CertificationGateRequest{}, "flow/"+connector+"/missing", "input", "connector has applicable flow roles but no pair cells"), true
	}
	for _, pair := range pairs {
		cellID := "flow/" + pair.FlowKind + "/" + pair.Source + "/" + pair.Destination
		binding := certificationEvidenceBinding{Scope: "flow", Source: pair.Source, Destination: pair.Destination, FlowKind: pair.FlowKind}
		if halt, ok := artifacts.evaluateCell(cellID, binding, pair.Cell.certificationFacts, gate, retries); ok {
			return halt, true
		}
	}
	return CertificationGateVerdict{}, false
}

func (artifacts certificationArtifacts) evaluateStatus(connector string, retries *[]CertificationGateFailure) (CertificationGateVerdict, bool) {
	status, _ := findStatus(artifacts.status.Connectors, connector)
	flowStatus, _ := findStatus(artifacts.flow.ConnectorStatuses, connector)
	if status != flowStatus {
		return certificationGateHalt(CertificationGateRequest{}, "status/"+connector+"/disagreement", "input", "flow and status artifacts disagree"), true
	}
	if !status.Certified {
		*retries = append(*retries, CertificationGateFailure{
			ID:      "status/" + connector + "/certified",
			Class:   "status",
			CellID:  "status/" + connector,
			Message: "generated connector status is not certified",
		})
	}
	return CertificationGateVerdict{}, false
}

func (artifacts certificationArtifacts) evaluateCell(cellID string, binding certificationEvidenceBinding, facts certificationFacts, gate ConnectorCertificationGate, retries *[]CertificationGateFailure) (CertificationGateVerdict, bool) {
	if !facts.Applicable {
		if facts.NotApplicable == nil || strings.TrimSpace(facts.NotApplicable.Code) == "" || strings.TrimSpace(facts.NotApplicable.Reason) == "" {
			return certificationGateHaltAt(CertificationGateRequest{}, cellID+"/not_applicable", "input", cellID, "", "non-applicable cell has no explicit reason"), true
		}
		if facts.Declared || facts.Implemented || facts.FixtureTested || facts.LiveTested || len(facts.LiveEvidence) != 0 {
			return certificationGateHaltAt(CertificationGateRequest{}, cellID+"/not_applicable", "input", cellID, "", "non-applicable cell carries certification claims"), true
		}
		return CertificationGateVerdict{}, false
	}
	if facts.NotApplicable != nil {
		return certificationGateHaltAt(CertificationGateRequest{}, cellID+"/not_applicable", "input", cellID, "", "applicable cell carries a not_applicable reason"), true
	}

	criteria := []struct {
		name string
		pass bool
	}{
		{name: "declared", pass: facts.Declared},
		{name: "implemented", pass: facts.Implemented},
		{name: "fixture_tested", pass: facts.FixtureTested && len(facts.FixtureEvidence) != 0},
		{name: "live_tested", pass: facts.LiveTested},
		{name: "live_evidence", pass: len(facts.LiveEvidence) != 0},
	}
	for _, criterion := range criteria {
		if criterion.pass {
			continue
		}
		*retries = append(*retries, CertificationGateFailure{
			ID:      cellID + "/" + criterion.name,
			Class:   "binding",
			CellID:  cellID,
			Message: "binding criterion is not proven by the generated artifact",
		})
	}
	if len(facts.LiveEvidence) == 0 {
		return CertificationGateVerdict{}, false
	}
	if err := artifacts.validateCellEvidence(certificationEvidenceCell{cellID: cellID, binding: binding, facts: facts}, gate); err != nil {
		return certificationEvidenceHalt(CertificationGateRequest{}, err), true
	}
	return CertificationGateVerdict{}, false
}

// certificationFacts is shared by capability, workflow, sync-mode, and flow cells. It is fully
// typed so strict decoding rejects additions instead of allowing a proof schema to drift silently.
type certificationFacts struct {
	Applicable      bool                              `json:"applicable"`
	Declared        bool                              `json:"declared"`
	Implemented     bool                              `json:"implemented"`
	FixtureTested   bool                              `json:"fixture_tested"`
	LiveTested      bool                              `json:"live_tested"`
	FixtureEvidence []string                          `json:"fixture_evidence"`
	LiveEvidence    []certificationEvidencePointer    `json:"live_evidence"`
	NotApplicable   *certificationNotApplicableReason `json:"not_applicable,omitempty"`
}

type certificationNotApplicableReason struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

type certificationEvidencePointer struct {
	Record               string                     `json:"record"`
	Provider             string                     `json:"provider"`
	ExecutedAt           string                     `json:"executed_at"`
	RunID                string                     `json:"run_id"`
	CredentialScope      string                     `json:"credential_scope"`
	CredentialNote       string                     `json:"credential_note"`
	CredentialScopeProof string                     `json:"credential_scope_proof"`
	Proof                certificationEvidenceProof `json:"proof"`
}

type certificationAcceptedEvidence struct {
	SchemaVersion        int                        `json:"schema_version"`
	Scope                string                     `json:"scope"`
	Status               string                     `json:"status"`
	CredentialScope      string                     `json:"credential_scope"`
	CredentialNote       string                     `json:"credential_note"`
	CredentialScopeProof string                     `json:"credential_scope_proof"`
	Connector            string                     `json:"connector,omitempty"`
	FunctionKind         string                     `json:"function_kind,omitempty"`
	WorkflowKind         string                     `json:"workflow_kind,omitempty"`
	SyncMode             string                     `json:"sync_mode,omitempty"`
	Primitive            string                     `json:"primitive,omitempty"`
	Source               string                     `json:"source,omitempty"`
	Destination          string                     `json:"destination,omitempty"`
	FlowKind             string                     `json:"flow_kind,omitempty"`
	Provider             string                     `json:"provider"`
	ExecutedAt           string                     `json:"executed_at"`
	RunID                string                     `json:"run_id"`
	Proof                certificationEvidenceProof `json:"proof"`
}

type certificationEvidenceProof struct {
	RedactionStrategy      string                          `json:"redaction_strategy"`
	PMBinarySHA256         string                          `json:"pm_binary_sha256"`
	PMCommandFingerprint   string                          `json:"pm_command_fingerprint"`
	CredentialFingerprints []string                        `json:"credential_fingerprints"`
	HTTPExchanges          []certificationHTTPExchange     `json:"http_exchanges"`
	DatabaseExchanges      []certificationDatabaseExchange `json:"database_exchanges"`
	Flow                   *certificationFlowRoundTrip     `json:"flow,omitempty"`
}

type certificationHTTPExchange struct {
	Operation string                    `json:"operation"`
	Request   certificationHTTPRequest  `json:"request"`
	Response  certificationHTTPResponse `json:"response"`
}

type certificationHTTPRequest struct {
	Method  string                   `json:"method"`
	Target  string                   `json:"target"`
	Query   []certificationQuery     `json:"query"`
	Headers []certificationHTTPField `json:"headers"`
	Body    certificationHTTPBody    `json:"body"`
}

type certificationHTTPResponse struct {
	Status  int                      `json:"status"`
	Headers []certificationHTTPField `json:"headers"`
	Body    certificationHTTPBody    `json:"body"`
}

type certificationHTTPField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type certificationQuery struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type certificationHTTPBody struct {
	Encoding      string          `json:"encoding"`
	Value         json.RawMessage `json:"value"`
	OriginalBytes int             `json:"original_bytes"`
	Truncated     bool            `json:"truncated"`
}

type certificationDatabaseExchange struct {
	Operation string                        `json:"operation"`
	Protocol  string                        `json:"protocol"`
	Request   certificationDatabaseRequest  `json:"request"`
	Response  certificationDatabaseResponse `json:"response"`
}

type certificationDatabaseRequest struct {
	Statement  string   `json:"statement"`
	Parameters []string `json:"parameters"`
}

type certificationDatabaseResponse struct {
	Status string                `json:"status"`
	Body   certificationHTTPBody `json:"body"`
}

type certificationFlowRoundTrip struct {
	PMCommandFingerprint         string                          `json:"pm_command_fingerprint"`
	Mediator                     string                          `json:"mediator"`
	WarehouseReadbackOperation   string                          `json:"warehouse_readback_operation"`
	DestinationReadbackOperation string                          `json:"destination_readback_operation"`
	Delivery                     certificationDeliveryGuarantees `json:"delivery"`
}

type certificationDeliveryGuarantees struct {
	Resumable              bool                              `json:"resumable"`
	ReceiptBacked          bool                              `json:"receipt_backed"`
	Checkpointed           bool                              `json:"checkpointed"`
	ReplayIdentity         bool                              `json:"replay_identity"`
	ProviderIdempotencyKey bool                              `json:"provider_idempotency_key"`
	Limitations            []certificationDeliveryLimitation `json:"limitations"`
}

type certificationDeliveryLimitation struct {
	Guarantee string `json:"guarantee"`
	Code      string `json:"code"`
	Reason    string `json:"reason"`
}

type certificationFunctionKind struct {
	ID              string `json:"id"`
	Category        string `json:"category"`
	Name            string `json:"name"`
	DiscoverySource string `json:"discovery_source"`
	ExecutorSource  string `json:"executor_source,omitempty"`
}

type certificationCapabilityCell struct {
	FunctionKind string `json:"function_kind"`
	certificationFacts
}

type certificationCapabilityConnector struct {
	Name               string                        `json:"name"`
	IntegrationType    string                        `json:"integration_type"`
	CapabilityComplete bool                          `json:"capability_complete"`
	Cells              []certificationCapabilityCell `json:"cells"`
}

type certificationLegacyFile struct {
	Path           string `json:"path"`
	Classification string `json:"classification"`
}

type certificationLegacyInventory struct {
	Ignored bool                      `json:"ignored"`
	Files   []certificationLegacyFile `json:"files"`
}

type certificationKindBaseline struct {
	FunctionKind  string `json:"function_kind"`
	Connectors    int    `json:"connectors"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

type certificationCapabilityBaseline struct {
	Connectors         int                         `json:"connectors"`
	CapabilityComplete int                         `json:"capability_complete"`
	PerKind            []certificationKindBaseline `json:"per_kind"`
}

type certificationCapabilityMatrix struct {
	SchemaVersion             int                                `json:"schema_version"`
	GeneratedCommand          string                             `json:"generated_command"`
	FunctionKinds             []certificationFunctionKind        `json:"function_kinds"`
	Connectors                []certificationCapabilityConnector `json:"connectors"`
	LegacyCertificationInputs certificationLegacyInventory       `json:"legacy_certification_inputs"`
	Baseline                  certificationCapabilityBaseline    `json:"baseline"`
}

type certificationFlowKind struct {
	ID              string `json:"id"`
	SourceRole      string `json:"source_role"`
	DestinationRole string `json:"destination_role"`
}

type certificationWorkflowKind struct {
	ID              string `json:"id"`
	DiscoverySource string `json:"discovery_source"`
}

type certificationWorkflowCell struct {
	WorkflowKind string `json:"workflow_kind"`
	certificationFacts
}

type certificationWorkflowSet struct {
	Connector string                      `json:"connector"`
	Complete  bool                        `json:"complete"`
	Cells     []certificationWorkflowCell `json:"cells"`
}

type certificationSyncModeKind struct {
	ID              string `json:"id"`
	DiscoverySource string `json:"discovery_source"`
}

type certificationSyncPrimitive struct {
	ID                 string `json:"id"`
	IntegrationType    string `json:"integration_type"`
	Capability         string `json:"capability"`
	WarehouseDirection string `json:"warehouse_direction"`
	DiscoverySource    string `json:"discovery_source"`
}

type certificationSyncModeCell struct {
	SyncMode  string `json:"sync_mode"`
	Primitive string `json:"primitive"`
	certificationFacts
}

type certificationSyncModeSet struct {
	Connector string                      `json:"connector"`
	Complete  bool                        `json:"complete"`
	Cells     []certificationSyncModeCell `json:"cells"`
}

type certificationFlowRole struct {
	Role          string                            `json:"role"`
	Applicable    bool                              `json:"applicable"`
	Declared      bool                              `json:"declared"`
	Implemented   bool                              `json:"implemented"`
	NotApplicable *certificationNotApplicableReason `json:"not_applicable,omitempty"`
}

type certificationConnectorRoles struct {
	Connector string                  `json:"connector"`
	Roles     []certificationFlowRole `json:"roles"`
}

type certificationFlowCell struct {
	certificationFacts
}

type certificationFlowPairSet struct {
	FlowKind              string                `json:"flow_kind"`
	Mediator              string                `json:"mediator"`
	SourceConnectors      []string              `json:"source_connectors"`
	DestinationConnectors []string              `json:"destination_connectors"`
	Cell                  certificationFlowCell `json:"cell"`
}

type certificationFlowPairOverride struct {
	FlowKind    string                `json:"flow_kind"`
	Source      string                `json:"source"`
	Destination string                `json:"destination"`
	Mediator    string                `json:"mediator"`
	Cell        certificationFlowCell `json:"cell"`
}

type certificationWorkflowBaseline struct {
	WorkflowKind  string `json:"workflow_kind"`
	Connectors    int    `json:"connectors"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

type certificationSyncModeBaseline struct {
	SyncMode      string `json:"sync_mode"`
	Primitive     string `json:"primitive"`
	Connectors    int    `json:"connectors"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

type certificationFlowBaselineKind struct {
	FlowKind      string `json:"flow_kind"`
	Pairs         int    `json:"pairs"`
	Applicable    int    `json:"applicable"`
	Declared      int    `json:"declared"`
	Implemented   int    `json:"implemented"`
	FixtureTested int    `json:"fixture_tested"`
	LiveTested    int    `json:"live_tested"`
	Complete      int    `json:"complete"`
}

type certificationFlowBaseline struct {
	Connectors int                             `json:"connectors"`
	Certified  int                             `json:"certified"`
	Workflows  []certificationWorkflowBaseline `json:"workflows"`
	SyncModes  []certificationSyncModeBaseline `json:"sync_modes"`
	PerKind    []certificationFlowBaselineKind `json:"per_kind"`
}

type certificationConnectorStatus struct {
	Connector string `json:"connector"`
	Certified bool   `json:"certified"`
	Label     string `json:"label"`
	Warning   string `json:"warning,omitempty"`
}

type certificationStatusArtifact struct {
	SchemaVersion      int                            `json:"schema_version"`
	GeneratedCommand   string                         `json:"generated_command"`
	CertificationScope []string                       `json:"certification_scope,omitempty"`
	Connectors         []certificationConnectorStatus `json:"connectors"`
}

type certificationFlowMatrix struct {
	SchemaVersion     int                             `json:"schema_version"`
	GeneratedCommand  string                          `json:"generated_command"`
	Mediator          string                          `json:"mediator"`
	FlowKinds         []certificationFlowKind         `json:"flow_kinds"`
	WorkflowKinds     []certificationWorkflowKind     `json:"workflow_kinds"`
	Workflows         []certificationWorkflowSet      `json:"workflows"`
	SyncModeKinds     []certificationSyncModeKind     `json:"sync_mode_kinds"`
	SyncPrimitives    []certificationSyncPrimitive    `json:"sync_primitives"`
	SyncModeCells     []certificationSyncModeSet      `json:"sync_mode_cells"`
	ConnectorRoles    []certificationConnectorRoles   `json:"connector_roles"`
	PairSets          []certificationFlowPairSet      `json:"pair_sets"`
	PairOverrides     []certificationFlowPairOverride `json:"pair_overrides"`
	ConnectorStatuses []certificationConnectorStatus  `json:"connector_statuses"`
	Baseline          certificationFlowBaseline       `json:"baseline"`
}

type certificationShard struct {
	SchemaVersion    int                              `json:"schema_version"`
	GeneratedCommand string                           `json:"generated_command"`
	FunctionKinds    []certificationFunctionKind      `json:"function_kinds"`
	Connector        certificationCapabilityConnector `json:"connector"`
	FlowKinds        []certificationFlowKind          `json:"flow_kinds"`
	WorkflowKinds    []certificationWorkflowKind      `json:"workflow_kinds"`
	Workflow         certificationWorkflowSet         `json:"workflow"`
	SyncModeKinds    []certificationSyncModeKind      `json:"sync_mode_kinds"`
	SyncPrimitives   []certificationSyncPrimitive     `json:"sync_primitives"`
	SyncModeCells    certificationSyncModeSet         `json:"sync_mode_cells"`
	ConnectorRoles   certificationConnectorRoles      `json:"connector_roles"`
	PairOverrides    []certificationFlowPairOverride  `json:"pair_overrides"`
}

func loadCertificationShardMatrices(root string, gate ConnectorCertificationGate, status certificationStatusArtifact) (certificationCapabilityMatrix, certificationFlowMatrix, error) {
	capabilities := certificationCapabilityMatrix{
		SchemaVersion:    gate.SchemaVersion,
		GeneratedCommand: gate.GeneratedCommand,
		LegacyCertificationInputs: certificationLegacyInventory{
			Ignored: true,
			Files:   []certificationLegacyFile{},
		},
	}
	flow := certificationFlowMatrix{
		SchemaVersion:     gate.SchemaVersion,
		GeneratedCommand:  gate.GeneratedCommand,
		Mediator:          localParquetWarehouse,
		ConnectorStatuses: slices.Clone(status.Connectors),
		PairOverrides:     []certificationFlowPairOverride{},
	}
	if status.SchemaVersion != gate.SchemaVersion || status.GeneratedCommand != certificationAllCommand || len(status.Connectors) == 0 {
		return capabilities, flow, errors.New("certification status does not identify a supported shard set")
	}

	names := make([]string, 0, len(status.Connectors))
	seen := make(map[string]bool, len(status.Connectors))
	claimedCapabilityComplete := make(map[string]bool, len(status.Connectors))
	statusByConnector := make(map[string]certificationConnectorStatus, len(status.Connectors))
	for _, item := range status.Connectors {
		if !safeCertificationID(item.Connector) || seen[item.Connector] {
			return capabilities, flow, fmt.Errorf("certification status connector %q is unsafe or duplicated", item.Connector)
		}
		seen[item.Connector] = true
		names = append(names, item.Connector)
		statusByConnector[item.Connector] = item
	}
	sort.Strings(names)

	for index, name := range names {
		var shard certificationShard
		shardPath := pathpkg.Join(gate.Inputs.CertificationShards, name, "certification-matrix.json")
		if err := readCertificationFile(root, shardPath, &shard); err != nil {
			return capabilities, flow, fmt.Errorf("read connector %q certification shard: %w", name, err)
		}
		if shard.SchemaVersion != gate.SchemaVersion || shard.GeneratedCommand != "go run ./cmd/connectorgen certification-matrix --connector "+name {
			return capabilities, flow, fmt.Errorf("connector %q certification shard has an unsupported schema or command", name)
		}
		if shard.Connector.Name != name || shard.Workflow.Connector != name || shard.SyncModeCells.Connector != name || shard.ConnectorRoles.Connector != name {
			return capabilities, flow, fmt.Errorf("connector %q certification shard has mismatched owned records", name)
		}
		if index == 0 {
			capabilities.FunctionKinds = slices.Clone(shard.FunctionKinds)
			flow.FlowKinds = slices.Clone(shard.FlowKinds)
			flow.WorkflowKinds = slices.Clone(shard.WorkflowKinds)
			flow.SyncModeKinds = slices.Clone(shard.SyncModeKinds)
			flow.SyncPrimitives = slices.Clone(shard.SyncPrimitives)
		} else if !reflect.DeepEqual(capabilities.FunctionKinds, shard.FunctionKinds) ||
			!reflect.DeepEqual(flow.FlowKinds, shard.FlowKinds) ||
			!reflect.DeepEqual(flow.WorkflowKinds, shard.WorkflowKinds) ||
			!reflect.DeepEqual(flow.SyncModeKinds, shard.SyncModeKinds) ||
			!reflect.DeepEqual(flow.SyncPrimitives, shard.SyncPrimitives) {
			return capabilities, flow, fmt.Errorf("connector %q certification shard has a divergent shared inventory", name)
		}
		for _, override := range shard.PairOverrides {
			if override.Source != name {
				return capabilities, flow, fmt.Errorf("connector %q certification shard owns an override for source %q", name, override.Source)
			}
		}
		capabilities.Connectors = append(capabilities.Connectors, shard.Connector)
		claimedCapabilityComplete[name] = shard.Connector.CapabilityComplete
		flow.Workflows = append(flow.Workflows, shard.Workflow)
		flow.SyncModeCells = append(flow.SyncModeCells, shard.SyncModeCells)
		flow.ConnectorRoles = append(flow.ConnectorRoles, shard.ConnectorRoles)
		flow.PairOverrides = append(flow.PairOverrides, shard.PairOverrides...)
	}

	rolesByConnector := make(map[string]map[string]certificationFlowRole, len(flow.ConnectorRoles))
	for _, declaration := range flow.ConnectorRoles {
		roles := make(map[string]certificationFlowRole, len(declaration.Roles))
		for _, role := range declaration.Roles {
			roles[role.Role] = role
		}
		rolesByConnector[declaration.Connector] = roles
	}
	for _, kind := range flow.FlowKinds {
		for _, source := range names {
			for _, destination := range names {
				sourceRole, sourceOK := rolesByConnector[source][kind.SourceRole]
				destinationRole, destinationOK := rolesByConnector[destination][kind.DestinationRole]
				if !sourceOK || !destinationOK {
					return capabilities, flow, fmt.Errorf("flow kind %q cannot resolve endpoint roles for %s -> %s", kind.ID, source, destination)
				}
				flow.PairSets = append(flow.PairSets, certificationFlowPairSet{
					FlowKind:              kind.ID,
					Mediator:              localParquetWarehouse,
					SourceConnectors:      []string{source},
					DestinationConnectors: []string{destination},
					Cell: certificationFlowCell{
						certificationFacts: certificationFlowPairRoleFacts(sourceRole, destinationRole),
					},
				})
			}
		}
	}
	capabilities.Baseline = deriveCertificationCapabilityBaseline(capabilities, claimedCapabilityComplete)
	flow.Baseline = deriveCertificationFlowBaseline(flow, statusByConnector)
	return capabilities, flow, nil
}

func loadCapabilityMatrix(root string, gate ConnectorCertificationGate) (certificationCapabilityMatrix, error) {
	var matrix certificationCapabilityMatrix
	if err := readCertificationFile(root, gate.Inputs.CapabilityMatrix, &matrix); err != nil {
		return matrix, err
	}
	return matrix, nil
}

func loadFlowMatrix(root string, gate ConnectorCertificationGate) (certificationFlowMatrix, error) {
	var matrix certificationFlowMatrix
	if err := readCertificationFile(root, gate.Inputs.FlowMatrix, &matrix); err != nil {
		return matrix, err
	}
	return matrix, nil
}

func loadStatusArtifact(root string, gate ConnectorCertificationGate) (certificationStatusArtifact, error) {
	var artifact certificationStatusArtifact
	if err := readCertificationFile(root, gate.Inputs.Status, &artifact); err != nil {
		return artifact, err
	}
	return artifact, nil
}

func readCertificationFile(root, relativePath string, destination any) error {
	reader, err := openCertificationRoot(root)
	if err != nil {
		return err
	}
	raw, readErr := reader.readFile(relativePath)
	closeErr := reader.Close()
	if readErr != nil {
		return readErr
	}
	if closeErr != nil {
		return fmt.Errorf("close certification root: %w", closeErr)
	}
	if err := decodeCertificationJSON(raw, destination); err != nil {
		return err
	}
	return nil
}

func loadAcceptedCertificationEvidence(root string, gate ConnectorCertificationGate) (map[string]certificationAcceptedEvidence, error) {
	directory := gate.Inputs.EvidenceDirectory
	entries, err := readCertificationDirectory(root, directory)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]certificationAcceptedEvidence{}, nil
	}
	if err != nil {
		return nil, err
	}
	records := make(map[string]certificationAcceptedEvidence, len(entries))
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if !safeCertificationID(strings.TrimSuffix(entry.Name(), ".json")) {
			return nil, fmt.Errorf("evidence record %q is unsafe", entry.Name())
		}
		recordPath := pathpkg.Join(directory, entry.Name())
		var record certificationAcceptedEvidence
		if err := readCertificationFile(root, recordPath, &record); err != nil {
			return nil, fmt.Errorf("read evidence record %s: %w", recordPath, err)
		}
		if err := validateAcceptedEvidence(record, gate); err != nil {
			return nil, fmt.Errorf("validate evidence record %s: %w", recordPath, err)
		}
		if _, exists := records[recordPath]; exists {
			return nil, fmt.Errorf("evidence record %s is duplicated", recordPath)
		}
		records[recordPath] = record
	}
	return records, nil
}

type certificationRootReader struct {
	root *os.Root
}

func openCertificationRoot(path string) (*certificationRootReader, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&fs.ModeSymlink != 0 || !info.IsDir() {
		return nil, fmt.Errorf("certification root %q must be a directory, not a symlink or special file", path)
	}
	root, err := os.OpenRoot(path)
	if err != nil {
		return nil, err
	}
	return &certificationRootReader{root: root}, nil
}

func readCertificationDirectory(root, directory string) ([]fs.DirEntry, error) {
	reader, err := openCertificationRoot(root)
	if err != nil {
		return nil, err
	}
	entries, readErr := reader.readDirectory(directory)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close certification root: %w", closeErr)
	}
	return entries, nil
}

func (reader *certificationRootReader) Close() error {
	return reader.root.Close()
}

func (reader *certificationRootReader) readFile(relativePath string) ([]byte, error) {
	if err := reader.inspectPath(relativePath, true); err != nil {
		return nil, err
	}
	return reader.root.ReadFile(filepath.FromSlash(relativePath))
}

func (reader *certificationRootReader) readDirectory(relativePath string) ([]fs.DirEntry, error) {
	if err := reader.inspectPath(relativePath, false); err != nil {
		return nil, err
	}
	return fs.ReadDir(reader.root.FS(), relativePath)
}

func (reader *certificationRootReader) inspectPath(relativePath string, regular bool) error {
	if !fs.ValidPath(relativePath) || pathpkg.Clean(relativePath) != relativePath || pathpkg.IsAbs(relativePath) {
		return fmt.Errorf("certification input path %q is not local", relativePath)
	}
	components := strings.Split(relativePath, "/")
	for index := range components {
		path := filepath.FromSlash(strings.Join(components[:index+1], "/"))
		info, err := reader.root.Lstat(path)
		if err != nil {
			return err
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			return fmt.Errorf("certification input path %q contains a symlink", relativePath)
		}
		if index < len(components)-1 && !info.IsDir() {
			return fmt.Errorf("certification input path %q has a non-directory ancestor", relativePath)
		}
		if index == len(components)-1 {
			if regular && !info.Mode().IsRegular() {
				return fmt.Errorf("certification input path %q is not a regular file", relativePath)
			}
			if !regular && !info.IsDir() {
				return fmt.Errorf("certification input path %q is not a directory", relativePath)
			}
		}
	}
	return nil
}

func validateCapabilityMatrix(matrix certificationCapabilityMatrix, gate ConnectorCertificationGate) error {
	if matrix.SchemaVersion != gate.SchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", matrix.SchemaVersion)
	}
	if matrix.GeneratedCommand != gate.GeneratedCommand {
		return fmt.Errorf("generated_command %q is unsupported", matrix.GeneratedCommand)
	}
	if len(matrix.FunctionKinds) == 0 || len(matrix.Connectors) == 0 {
		return errors.New("function kinds and connectors are required")
	}
	if err := validateUniqueIDs("capability kind", matrix.FunctionKinds, func(kind certificationFunctionKind) string { return kind.ID }); err != nil {
		return err
	}
	if err := validateUniqueIDs("capability connector", matrix.Connectors, func(connector certificationCapabilityConnector) string { return connector.Name }); err != nil {
		return err
	}
	kinds := make(map[string]bool, len(matrix.FunctionKinds))
	for _, kind := range matrix.FunctionKinds {
		kinds[kind.ID] = true
	}
	for _, connector := range matrix.Connectors {
		if len(connector.Cells) != len(kinds) {
			return fmt.Errorf("capability connector %q has %d cells for %d kinds", connector.Name, len(connector.Cells), len(kinds))
		}
		seen := make(map[string]bool, len(connector.Cells))
		for _, cell := range connector.Cells {
			if !kinds[cell.FunctionKind] || seen[cell.FunctionKind] {
				return fmt.Errorf("capability connector %q has an unknown or duplicate kind %q", connector.Name, cell.FunctionKind)
			}
			cellID := "capability/" + connector.Name + "/" + cell.FunctionKind
			if err := validateCertificationFacts(cell.certificationFacts, gate, cellID); err != nil {
				return fmt.Errorf("capability connector %q kind %q is invalid: %w", connector.Name, cell.FunctionKind, err)
			}
			seen[cell.FunctionKind] = true
		}
	}
	return nil
}

func validateFlowMatrix(matrix certificationFlowMatrix, gate ConnectorCertificationGate, capabilities certificationCapabilityMatrix, status certificationStatusArtifact) error {
	if matrix.SchemaVersion != gate.SchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", matrix.SchemaVersion)
	}
	if matrix.GeneratedCommand != gate.GeneratedCommand || matrix.Mediator != localParquetWarehouse {
		return errors.New("generated command or warehouse mediator is unsupported")
	}
	if len(matrix.FlowKinds) == 0 || len(matrix.WorkflowKinds) == 0 || len(matrix.Workflows) == 0 || len(matrix.SyncModeKinds) == 0 || len(matrix.SyncPrimitives) == 0 || len(matrix.SyncModeCells) == 0 || len(matrix.ConnectorRoles) == 0 {
		return errors.New("flow kinds, workflows, sync modes, primitives, sync-mode cells, and connector roles are required")
	}
	if err := validateUniqueIDs("flow kind", matrix.FlowKinds, func(kind certificationFlowKind) string { return kind.ID }); err != nil {
		return err
	}
	if err := validateCertificationFlowKindCatalog(matrix.FlowKinds); err != nil {
		return err
	}
	flowKinds := make(map[string]certificationFlowKind, len(matrix.FlowKinds))
	for _, kind := range matrix.FlowKinds {
		if !safeCertificationID(kind.SourceRole) || !safeCertificationID(kind.DestinationRole) {
			return fmt.Errorf("flow kind %q has an unsafe role", kind.ID)
		}
		flowKinds[kind.ID] = kind
	}
	if err := validateUniqueIDs("workflow kind", matrix.WorkflowKinds, func(kind certificationWorkflowKind) string { return kind.ID }); err != nil {
		return err
	}
	workflowKinds := make(map[string]bool, len(matrix.WorkflowKinds))
	for _, kind := range matrix.WorkflowKinds {
		if strings.TrimSpace(kind.DiscoverySource) == "" {
			return fmt.Errorf("workflow kind %q has no discovery source", kind.ID)
		}
		workflowKinds[kind.ID] = true
	}
	if err := validateUniqueIDs("sync mode", matrix.SyncModeKinds, func(kind certificationSyncModeKind) string { return kind.ID }); err != nil {
		return err
	}
	syncModes := make(map[string]bool, len(matrix.SyncModeKinds))
	for _, mode := range matrix.SyncModeKinds {
		if strings.TrimSpace(mode.DiscoverySource) == "" {
			return fmt.Errorf("sync mode %q has no discovery source", mode.ID)
		}
		syncModes[mode.ID] = true
	}
	if err := validateUniqueIDs("sync primitive", matrix.SyncPrimitives, func(primitive certificationSyncPrimitive) string { return primitive.ID }); err != nil {
		return err
	}
	syncPrimitives := make(map[string]certificationSyncPrimitive, len(matrix.SyncPrimitives))
	for _, primitive := range matrix.SyncPrimitives {
		syncPrimitives[primitive.ID] = primitive
	}
	if err := validateRequiredCertificationSyncPrimitives(syncPrimitives); err != nil {
		return err
	}
	primitiveIDs := make(map[string]bool, len(syncPrimitives))
	for id := range syncPrimitives {
		primitiveIDs[id] = true
	}
	if err := validateUniqueIDs("flow-role connector", matrix.ConnectorRoles, func(roles certificationConnectorRoles) string { return roles.Connector }); err != nil {
		return err
	}
	connectors := make(map[string]bool, len(matrix.ConnectorRoles))
	connectorRoles := make(map[string]map[string]certificationFlowRole, len(matrix.ConnectorRoles))
	for _, declaration := range matrix.ConnectorRoles {
		connectors[declaration.Connector] = true
		roles := make(map[string]certificationFlowRole, len(declaration.Roles))
		for _, role := range declaration.Roles {
			if _, exists := roles[role.Role]; !safeCertificationID(role.Role) || exists {
				return fmt.Errorf("flow-role connector %q has an unsafe or duplicate role %q", declaration.Connector, role.Role)
			}
			roles[role.Role] = role
			if !role.Applicable {
				if role.NotApplicable == nil || role.Declared || role.Implemented {
					return fmt.Errorf("flow-role connector %q role %q is invalid", declaration.Connector, role.Role)
				}
				if err := validateCertificationReason(role.NotApplicable.Code, role.NotApplicable.Reason); err != nil {
					return fmt.Errorf("flow-role connector %q role %q is invalid: %w", declaration.Connector, role.Role, err)
				}
			} else if role.NotApplicable != nil {
				return fmt.Errorf("flow-role connector %q role %q is invalid", declaration.Connector, role.Role)
			}
		}
		for _, kind := range matrix.FlowKinds {
			if _, ok := roles[kind.SourceRole]; !ok {
				return fmt.Errorf("flow-role connector %q omits a required role", declaration.Connector)
			}
			if _, ok := roles[kind.DestinationRole]; !ok {
				return fmt.Errorf("flow-role connector %q omits a required role", declaration.Connector)
			}
		}
		connectorRoles[declaration.Connector] = roles
	}
	if err := validateWorkflowTopology(matrix.Workflows, connectors, workflowKinds, gate); err != nil {
		return err
	}
	if err := validateSyncModeTopology(matrix.SyncModeCells, connectors, syncModes, primitiveIDs, gate); err != nil {
		return err
	}
	if err := validateFlowPairTopology(matrix, flowKinds, connectors, connectorRoles, gate); err != nil {
		return err
	}
	if err := validateUniqueIDs("flow status", matrix.ConnectorStatuses, func(status certificationConnectorStatus) string { return status.Connector }); err != nil {
		return err
	}
	flowStatuses := make([]string, 0, len(matrix.ConnectorStatuses))
	for _, item := range matrix.ConnectorStatuses {
		flowStatuses = append(flowStatuses, item.Connector)
	}
	if !sameCertificationConnectorSet(connectors, flowStatuses) {
		return errors.New("flow connector statuses do not match connector roles")
	}
	capabilityConnectors := make([]string, 0, len(capabilities.Connectors))
	for _, connector := range capabilities.Connectors {
		capabilityConnectors = append(capabilityConnectors, connector.Name)
	}
	if !sameCertificationConnectorSet(connectors, capabilityConnectors) {
		return errors.New("flow connector roles do not match capability connectors")
	}
	statusConnectors := make([]string, 0, len(status.Connectors))
	for _, connector := range status.Connectors {
		statusConnectors = append(statusConnectors, connector.Connector)
	}
	if !sameCertificationConnectorSet(connectors, statusConnectors) {
		return errors.New("flow connector roles do not match status connectors")
	}
	return nil
}

func validateCertificationFlowKindCatalog(kinds []certificationFlowKind) error {
	expected := certificationcatalog.FlowKinds()
	if len(expected) == 0 {
		return errors.New("generated flow-kind catalog is missing or invalid")
	}
	if len(kinds) != len(expected) {
		return fmt.Errorf("flow kind inventory has %d entries, want %d producer-defined kinds", len(kinds), len(expected))
	}
	actual := make(map[string]certificationFlowKind, len(kinds))
	for _, kind := range kinds {
		actual[kind.ID] = kind
	}
	for _, want := range expected {
		got, ok := actual[want.ID]
		if !ok || got.SourceRole != want.SourceRole || got.DestinationRole != want.DestinationRole {
			return fmt.Errorf("flow kind %q is missing or has an invalid producer mapping", want.ID)
		}
	}
	return nil
}

func validateStatusArtifact(artifact certificationStatusArtifact, gate ConnectorCertificationGate) error {
	if artifact.SchemaVersion != gate.SchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", artifact.SchemaVersion)
	}
	wantCommand := gate.GeneratedCommand
	if gate.Inputs.CertificationShards != "" {
		wantCommand = certificationAllCommand
	}
	if artifact.GeneratedCommand != wantCommand || len(artifact.Connectors) == 0 {
		return errors.New("generated command or connector statuses are invalid")
	}
	if err := validateUniqueIDs("status connector", artifact.Connectors, func(status certificationConnectorStatus) string { return status.Connector }); err != nil {
		return err
	}
	if gate.Inputs.CertificationShards != "" {
		connectors := make([]string, 0, len(artifact.Connectors))
		for _, item := range artifact.Connectors {
			connectors = append(connectors, item.Connector)
		}
		if !slices.Equal(artifact.CertificationScope, connectors) {
			return errors.New("certification scope does not match status connectors")
		}
	} else if len(artifact.CertificationScope) != 0 {
		return errors.New("aggregate certification status cannot declare a shard scope")
	}
	for _, status := range artifact.Connectors {
		if status.Certified {
			if status.Label != "CERTIFIED" || status.Warning != "" {
				return fmt.Errorf("certified status %q is malformed", status.Connector)
			}
			continue
		}
		if status.Label != "COMMUNITY BUILD, UNCERTIFIED" || status.Warning != "This connector is reachable but is a COMMUNITY BUILD, UNCERTIFIED." {
			return fmt.Errorf("uncertified status %q is malformed", status.Connector)
		}
	}
	return nil
}

func validateWorkflowTopology(sets []certificationWorkflowSet, connectors, kinds map[string]bool, gate ConnectorCertificationGate) error {
	if len(sets) != len(connectors) {
		return errors.New("workflow sets omit one or more connectors")
	}
	seenSets := make(map[string]bool, len(sets))
	for _, set := range sets {
		if !connectors[set.Connector] || seenSets[set.Connector] {
			return fmt.Errorf("workflow connector %q is unknown or duplicated", set.Connector)
		}
		seenSets[set.Connector] = true
		if len(set.Cells) != len(kinds) {
			return fmt.Errorf("workflow connector %q has %d cells for %d kinds", set.Connector, len(set.Cells), len(kinds))
		}
		seenCells := make(map[string]bool, len(set.Cells))
		for _, cell := range set.Cells {
			if !kinds[cell.WorkflowKind] || seenCells[cell.WorkflowKind] {
				return fmt.Errorf("workflow connector %q has an unknown or duplicate kind %q", set.Connector, cell.WorkflowKind)
			}
			cellID := "workflow/" + set.Connector + "/" + cell.WorkflowKind
			if err := validateCertificationFacts(cell.certificationFacts, gate, cellID); err != nil {
				return fmt.Errorf("workflow connector %q kind %q is invalid: %w", set.Connector, cell.WorkflowKind, err)
			}
			seenCells[cell.WorkflowKind] = true
		}
		if set.Complete != certificationWorkflowCellsComplete(set.Cells) {
			return fmt.Errorf("workflow connector %q complete disagrees with cells", set.Connector)
		}
	}
	return nil
}

func validateSyncModeTopology(sets []certificationSyncModeSet, connectors, modes, primitives map[string]bool, gate ConnectorCertificationGate) error {
	if len(sets) != len(connectors) {
		return errors.New("sync-mode sets omit one or more connectors")
	}
	expected := len(modes) * len(primitives)
	seenSets := make(map[string]bool, len(sets))
	for _, set := range sets {
		if !connectors[set.Connector] || seenSets[set.Connector] {
			return fmt.Errorf("sync-mode connector %q is unknown or duplicated", set.Connector)
		}
		seenSets[set.Connector] = true
		if len(set.Cells) != expected {
			return fmt.Errorf("sync-mode connector %q has %d cells for %d mode/primitive combinations", set.Connector, len(set.Cells), expected)
		}
		seenCells := make(map[string]bool, len(set.Cells))
		for _, cell := range set.Cells {
			if !modes[cell.SyncMode] || !primitives[cell.Primitive] {
				return fmt.Errorf("sync-mode connector %q has an unknown mode or primitive", set.Connector)
			}
			key := cell.SyncMode + "\x00" + cell.Primitive
			if seenCells[key] {
				return fmt.Errorf("sync-mode connector %q duplicates %q", set.Connector, key)
			}
			cellID := "sync_mode/" + set.Connector + "/" + cell.SyncMode + "/" + cell.Primitive
			if err := validateCertificationFacts(cell.certificationFacts, gate, cellID); err != nil {
				return fmt.Errorf("sync-mode connector %q %q is invalid: %w", set.Connector, key, err)
			}
			seenCells[key] = true
		}
		if set.Complete != certificationSyncModeCellsComplete(set.Cells) {
			return fmt.Errorf("sync-mode connector %q complete disagrees with cells", set.Connector)
		}
	}
	return nil
}

func validateFlowPairTopology(matrix certificationFlowMatrix, kinds map[string]certificationFlowKind, connectors map[string]bool, connectorRoles map[string]map[string]certificationFlowRole, gate ConnectorCertificationGate) error {
	for _, set := range matrix.PairSets {
		kind, ok := kinds[set.FlowKind]
		if !ok || set.Mediator != matrix.Mediator || !sortedUniqueCertificationIDs(set.SourceConnectors) || !sortedUniqueCertificationIDs(set.DestinationConnectors) {
			return errors.New("flow pair set is invalid")
		}
		for _, connector := range append(append([]string{}, set.SourceConnectors...), set.DestinationConnectors...) {
			if !connectors[connector] {
				return fmt.Errorf("flow pair set names unknown connector %q", connector)
			}
		}
		for _, source := range set.SourceConnectors {
			sourceRole, ok := connectorRoles[source][kind.SourceRole]
			if !ok {
				return fmt.Errorf("flow pair set %q source %q omits role %q", set.FlowKind, source, kind.SourceRole)
			}
			for _, destination := range set.DestinationConnectors {
				destinationRole, ok := connectorRoles[destination][kind.DestinationRole]
				if !ok {
					return fmt.Errorf("flow pair set %q destination %q omits role %q", set.FlowKind, destination, kind.DestinationRole)
				}
				cellID := "flow/" + set.FlowKind + "/" + source + "/" + destination
				if err := validateCertificationFacts(set.Cell.certificationFacts, gate, cellID); err != nil {
					return fmt.Errorf("flow pair set %q is invalid: %w", set.FlowKind, err)
				}
				if err := validateCertificationFlowPairRoleInvariant(set.Cell.certificationFacts, sourceRole, destinationRole, cellID); err != nil {
					return err
				}
			}
		}
	}
	names := make([]string, 0, len(connectors))
	for name := range connectors {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, kind := range matrix.FlowKinds {
		for _, source := range names {
			for _, destination := range names {
				if countCertificationFlowPairSets(matrix.PairSets, kind.ID, source, destination) != 1 {
					return fmt.Errorf("flow pair coverage %s %s -> %s is incomplete or duplicated", kind.ID, source, destination)
				}
			}
		}
	}
	seenOverrides := make(map[string]bool, len(matrix.PairOverrides))
	for _, override := range matrix.PairOverrides {
		key := flowPairKey(override.FlowKind, override.Source, override.Destination)
		if seenOverrides[key] || override.Mediator != matrix.Mediator || kinds[override.FlowKind].ID == "" || !connectors[override.Source] || !connectors[override.Destination] {
			return errors.New("flow pair override is invalid")
		}
		seenOverrides[key] = true
		base, ok := certificationFlowPairSetFor(matrix.PairSets, override.FlowKind, override.Source, override.Destination)
		if !ok || !base.Cell.Applicable {
			return errors.New("flow pair override has no default pair")
		}
		cellID := "flow/" + override.FlowKind + "/" + override.Source + "/" + override.Destination
		if err := validateCertificationFacts(override.Cell.certificationFacts, gate, cellID); err != nil {
			return fmt.Errorf("flow pair override %q is invalid: %w", key, err)
		}
		if !override.Cell.Applicable {
			return errors.New("flow pair override is invalid")
		}
		if !sameCertificationFlowOverrideFacts(base.Cell.certificationFacts, override.Cell.certificationFacts) {
			return &certificationInvalidCellError{id: cellID + "/override_immutable", cellID: cellID, message: "flow pair override changes immutable facts"}
		}
	}
	return nil
}

func validateCertificationFlowPairRoleInvariant(facts certificationFacts, sourceRole, destinationRole certificationFlowRole, cellID string) error {
	expected := certificationFlowPairRoleFacts(sourceRole, destinationRole)
	if facts.Applicable == expected.Applicable &&
		facts.Declared == expected.Declared &&
		facts.Implemented == expected.Implemented &&
		reflect.DeepEqual(facts.NotApplicable, expected.NotApplicable) {
		return nil
	}
	return &certificationInvalidCellError{id: cellID + "/role_invariant", cellID: cellID, message: "flow pair cell disagrees with endpoint roles"}
}

func certificationFlowPairRoleFacts(sourceRole, destinationRole certificationFlowRole) certificationFacts {
	if !sourceRole.Applicable && !destinationRole.Applicable {
		return certificationFacts{NotApplicable: &certificationNotApplicableReason{
			Code:   "source_and_destination_roles_inapplicable",
			Reason: "source " + sourceRole.Role + " and destination " + destinationRole.Role + " roles are not applicable",
		}}
	}
	if !sourceRole.Applicable {
		return certificationFacts{NotApplicable: &certificationNotApplicableReason{
			Code:   "source_" + sourceRole.NotApplicable.Code,
			Reason: "source " + sourceRole.NotApplicable.Reason,
		}}
	}
	if !destinationRole.Applicable {
		return certificationFacts{NotApplicable: &certificationNotApplicableReason{
			Code:   "destination_" + destinationRole.NotApplicable.Code,
			Reason: "destination " + destinationRole.NotApplicable.Reason,
		}}
	}
	return certificationFacts{
		Applicable:  true,
		Declared:    sourceRole.Declared && destinationRole.Declared,
		Implemented: sourceRole.Implemented && destinationRole.Implemented,
	}
}

func sameCertificationFlowOverrideFacts(base, override certificationFacts) bool {
	return base.Applicable == override.Applicable &&
		base.Declared == override.Declared &&
		base.Implemented == override.Implemented &&
		base.FixtureTested == override.FixtureTested &&
		slices.Equal(base.FixtureEvidence, override.FixtureEvidence) &&
		reflect.DeepEqual(base.NotApplicable, override.NotApplicable)
}

func validateRequiredCertificationSyncPrimitives(primitives map[string]certificationSyncPrimitive) error {
	want := map[string]certificationSyncPrimitive{
		"api_read_into_warehouse":       {ID: "api_read_into_warehouse", IntegrationType: "api", Capability: "read", WarehouseDirection: "into_warehouse"},
		"api_write_from_warehouse":      {ID: "api_write_from_warehouse", IntegrationType: "api", Capability: "write", WarehouseDirection: "from_warehouse"},
		"database_read_into_warehouse":  {ID: "database_read_into_warehouse", IntegrationType: "database", Capability: "read", WarehouseDirection: "into_warehouse"},
		"database_write_from_warehouse": {ID: "database_write_from_warehouse", IntegrationType: "database", Capability: "write", WarehouseDirection: "from_warehouse"},
	}
	if len(primitives) != len(want) {
		return fmt.Errorf("sync primitive inventory has %d entries, want four required warehouse-facing primitives", len(primitives))
	}
	for id, expected := range want {
		actual, ok := primitives[id]
		if !ok || actual.IntegrationType != expected.IntegrationType || actual.Capability != expected.Capability || actual.WarehouseDirection != expected.WarehouseDirection || strings.TrimSpace(actual.DiscoverySource) == "" {
			return fmt.Errorf("sync primitive %q is missing or has an invalid warehouse-facing mapping", id)
		}
	}
	return nil
}

type certificationInvalidPointerError struct {
	cellID     string
	evidenceID string
}

func (err *certificationInvalidPointerError) Error() string {
	return "invalid_pointer"
}

type certificationInvalidCellError struct {
	id      string
	cellID  string
	message string
}

func (err *certificationInvalidCellError) Error() string {
	return err.message
}

func validateCertificationFacts(facts certificationFacts, gate ConnectorCertificationGate, cellID string) error {
	if !facts.Applicable {
		if facts.NotApplicable == nil || facts.Declared || facts.Implemented || facts.FixtureTested || facts.LiveTested || len(facts.FixtureEvidence) != 0 || len(facts.LiveEvidence) != 0 {
			return &certificationInvalidCellError{id: cellID + "/not_applicable", cellID: cellID, message: "non-applicable cell is invalid"}
		}
		if err := validateCertificationReason(facts.NotApplicable.Code, facts.NotApplicable.Reason); err != nil {
			return &certificationInvalidCellError{id: cellID + "/not_applicable", cellID: cellID, message: err.Error()}
		}
		return nil
	}
	if facts.NotApplicable != nil {
		return &certificationInvalidCellError{id: cellID + "/not_applicable", cellID: cellID, message: "applicable cell has a not_applicable reason"}
	}
	if facts.FixtureTested && len(facts.FixtureEvidence) == 0 {
		return &certificationInvalidCellError{id: cellID + "/fixture_evidence", cellID: cellID, message: "fixture_tested cell requires fixture evidence"}
	}
	if facts.LiveTested && len(facts.LiveEvidence) == 0 {
		return &certificationInvalidCellError{id: cellID + "/live_evidence", cellID: cellID, message: "live_tested cell requires live evidence"}
	}
	if !facts.LiveTested && len(facts.LiveEvidence) != 0 {
		return &certificationInvalidCellError{id: cellID + "/live_evidence", cellID: cellID, message: "live evidence requires live_tested=true"}
	}
	for _, pointer := range facts.LiveEvidence {
		if err := validateEvidencePointer(pointer, gate); err != nil {
			evidenceID := ""
			if safeEvidenceRecordPath(pointer.Record, gate.Inputs.EvidenceDirectory) {
				evidenceID = pointer.Record
			}
			return &certificationInvalidPointerError{cellID: cellID, evidenceID: evidenceID}
		}
	}
	return nil
}

func certificationWorkflowCellsComplete(cells []certificationWorkflowCell) bool {
	applicable := 0
	for _, cell := range cells {
		if !cell.Applicable {
			continue
		}
		applicable++
		if !certificationFactsComplete(cell.certificationFacts) {
			return false
		}
	}
	return applicable != 0
}

func certificationSyncModeCellsComplete(cells []certificationSyncModeCell) bool {
	applicable := 0
	for _, cell := range cells {
		if !cell.Applicable {
			continue
		}
		applicable++
		if !certificationFactsComplete(cell.certificationFacts) {
			return false
		}
	}
	return applicable != 0
}

func certificationFactsComplete(facts certificationFacts) bool {
	return facts.Applicable && facts.Declared && facts.Implemented && facts.FixtureTested && facts.LiveTested && len(facts.LiveEvidence) != 0
}

func certificationFlowPairSetFor(sets []certificationFlowPairSet, kind, source, destination string) (certificationFlowPairSet, bool) {
	for _, set := range sets {
		if set.FlowKind == kind && slices.Contains(set.SourceConnectors, source) && slices.Contains(set.DestinationConnectors, destination) {
			return set, true
		}
	}
	return certificationFlowPairSet{}, false
}

func sortedUniqueCertificationIDs(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for index, value := range values {
		if !safeCertificationID(value) || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}

func countCertificationFlowPairSets(sets []certificationFlowPairSet, kind, source, destination string) int {
	count := 0
	for _, set := range sets {
		if set.FlowKind == kind && slices.Contains(set.SourceConnectors, source) && slices.Contains(set.DestinationConnectors, destination) {
			count++
		}
	}
	return count
}

func sameCertificationConnectorSet(expected map[string]bool, actual []string) bool {
	if len(expected) != len(actual) {
		return false
	}
	for _, name := range actual {
		if !expected[name] {
			return false
		}
	}
	return true
}

func validateUniqueIDs[T any](label string, values []T, identifier func(T) string) error {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		id := identifier(value)
		if !safeCertificationID(id) || seen[id] {
			return fmt.Errorf("%s %q is missing, unsafe, or duplicated", label, id)
		}
		seen[id] = true
	}
	return nil
}

func validateAcceptedEvidence(record certificationAcceptedEvidence, gate ConnectorCertificationGate) error {
	if record.SchemaVersion != gate.AcceptedEvidenceSchemaVersion || record.Status != "passed" {
		return errors.New("evidence schema_version or status is unsupported")
	}
	if err := validateEvidenceIdentity(record.Provider, record.ExecutedAt, record.RunID, record.CredentialScope, record.CredentialNote, record.CredentialScopeProof); err != nil {
		return err
	}
	if err := validateEvidenceProof(record.Proof, gate); err != nil {
		return err
	}
	switch record.Scope {
	case "capability":
		if !safeCertificationID(record.Connector) || !safeCertificationID(record.FunctionKind) || record.Proof.Flow != nil {
			return errors.New("capability evidence binding is invalid")
		}
	case "workflow":
		if !safeCertificationID(record.Connector) || !safeCertificationID(record.WorkflowKind) || record.Proof.Flow != nil {
			return errors.New("workflow evidence binding is invalid")
		}
	case "sync_mode":
		if !safeCertificationID(record.Connector) || !safeCertificationID(record.SyncMode) || !safeCertificationID(record.Primitive) || record.Proof.Flow != nil {
			return errors.New("sync-mode evidence binding is invalid")
		}
	case "flow":
		if !safeCertificationID(record.Source) || !safeCertificationID(record.Destination) || !safeCertificationID(record.FlowKind) || record.Proof.Flow == nil {
			return errors.New("flow evidence binding is invalid")
		}
	default:
		return fmt.Errorf("evidence scope %q is unsupported", record.Scope)
	}
	return nil
}

func validateEvidencePointer(pointer certificationEvidencePointer, gate ConnectorCertificationGate) error {
	if !safeEvidenceRecordPath(pointer.Record, gate.Inputs.EvidenceDirectory) {
		return errors.New("evidence record path is unsafe")
	}
	if err := validateEvidenceIdentity(pointer.Provider, pointer.ExecutedAt, pointer.RunID, pointer.CredentialScope, pointer.CredentialNote, pointer.CredentialScopeProof); err != nil {
		return err
	}
	return validateEvidenceProof(pointer.Proof, gate)
}

func validateEvidenceIdentity(provider, executedAt, runID, credentialScope, credentialNote, credentialScopeProof string) error {
	if !safeCertificationID(provider) || !safeCertificationID(runID) {
		return errors.New("provider and run_id must be safe identifiers")
	}
	if _, err := time.Parse(time.RFC3339, executedAt); err != nil {
		return fmt.Errorf("executed_at must be RFC3339: %w", err)
	}
	switch credentialScope {
	case fullParityCredentialScope:
		if credentialNote != fullParityCredentialNote || credentialScopeProof != fullParityCredentialScopeProof {
			return errors.New("full-parity credential scope proof is invalid")
		}
	case observedOperationsCredentialScope:
		if credentialNote != observedOperationsCredentialNote || credentialScopeProof != observedOperationsCredentialScopeProof {
			return errors.New("observed-operations credential scope proof is invalid")
		}
	default:
		return errors.New("evidence credential scope is unsupported")
	}
	return nil
}

func validateEvidenceProof(proof certificationEvidenceProof, gate ConnectorCertificationGate) error {
	if proof.RedactionStrategy != gate.ProofRedactionStrategy || !sha256Digest.MatchString(proof.PMBinarySHA256) || !certificationFingerprintSequence(proof.PMCommandFingerprint) {
		return errors.New("evidence proof version or fingerprints are invalid")
	}
	if len(proof.CredentialFingerprints) == 0 || !sortedUniqueFingerprints(proof.CredentialFingerprints) {
		return errors.New("evidence proof credential fingerprints are invalid")
	}
	if len(proof.HTTPExchanges)+len(proof.DatabaseExchanges) == 0 {
		return errors.New("evidence proof requires a protocol exchange")
	}
	operations := make(map[string]bool, len(proof.HTTPExchanges)+len(proof.DatabaseExchanges))
	for index, exchange := range proof.HTTPExchanges {
		if !safeCertificationID(exchange.Operation) || operations[exchange.Operation] ||
			!safeHTTPMethod(exchange.Request.Method) || !certificationFingerprintSequence(exchange.Request.Target) ||
			exchange.Response.Status < 100 || exchange.Response.Status > 599 {
			return fmt.Errorf("http_exchanges[%d] is invalid", index)
		}
		if err := validateHTTPFields(exchange.Request.Headers); err != nil {
			return fmt.Errorf("http_exchanges[%d].request.headers: %w", index, err)
		}
		if err := validateQueries(exchange.Request.Query); err != nil {
			return fmt.Errorf("http_exchanges[%d].request.query: %w", index, err)
		}
		if err := validateHTTPFields(exchange.Response.Headers); err != nil {
			return fmt.Errorf("http_exchanges[%d].response.headers: %w", index, err)
		}
		if err := validateHTTPBody(exchange.Request.Body); err != nil {
			return fmt.Errorf("http_exchanges[%d].request.body: %w", index, err)
		}
		if err := validateHTTPBody(exchange.Response.Body); err != nil {
			return fmt.Errorf("http_exchanges[%d].response.body: %w", index, err)
		}
		operations[exchange.Operation] = true
	}
	for index, exchange := range proof.DatabaseExchanges {
		if !safeCertificationID(exchange.Operation) || operations[exchange.Operation] || !safeCertificationID(exchange.Protocol) ||
			!certificationFingerprintSequence(exchange.Request.Statement) || !safeCertificationID(exchange.Response.Status) {
			return fmt.Errorf("database_exchanges[%d] is invalid", index)
		}
		for _, parameter := range exchange.Request.Parameters {
			if !certificationFingerprintSequence(parameter) {
				return fmt.Errorf("database_exchanges[%d] has an invalid parameter fingerprint", index)
			}
		}
		if err := validateHTTPBody(exchange.Response.Body); err != nil {
			return fmt.Errorf("database_exchanges[%d].response.body: %w", index, err)
		}
		operations[exchange.Operation] = true
	}
	if proof.Flow != nil {
		flow := proof.Flow
		if !certificationFingerprintSequence(flow.PMCommandFingerprint) || flow.Mediator != localParquetWarehouse ||
			!operations[flow.WarehouseReadbackOperation] || !operations[flow.DestinationReadbackOperation] ||
			flow.WarehouseReadbackOperation == flow.DestinationReadbackOperation {
			return errors.New("flow proof lacks independent warehouse and destination readback exchanges")
		}
		if err := validateCertificationDeliveryGuarantees(flow.Delivery); err != nil {
			return fmt.Errorf("flow proof delivery guarantees: %w", err)
		}
	}
	return nil
}

func validateCertificationDeliveryGuarantees(delivery certificationDeliveryGuarantees) error {
	guarantees := []struct {
		name    string
		isFalse bool
	}{
		{name: "resumable", isFalse: !delivery.Resumable},
		{name: "receipt_backed", isFalse: !delivery.ReceiptBacked},
		{name: "checkpointed", isFalse: !delivery.Checkpointed},
		{name: "replay_identity", isFalse: !delivery.ReplayIdentity},
		{name: "provider_idempotency_key", isFalse: !delivery.ProviderIdempotencyKey},
	}
	falseGuarantees := make(map[string]bool, len(guarantees))
	for _, guarantee := range guarantees {
		falseGuarantees[guarantee.name] = guarantee.isFalse
	}
	covered := make(map[string]bool, len(delivery.Limitations))
	for _, limitation := range delivery.Limitations {
		if !falseGuarantees[limitation.Guarantee] {
			return fmt.Errorf("limitation %q does not correspond to a false guarantee", limitation.Guarantee)
		}
		if covered[limitation.Guarantee] {
			return fmt.Errorf("limitation %q is duplicated", limitation.Guarantee)
		}
		if err := validateCertificationLimitation(limitation); err != nil {
			return fmt.Errorf("limitation %q: %w", limitation.Guarantee, err)
		}
		covered[limitation.Guarantee] = true
	}
	for _, guarantee := range guarantees {
		if guarantee.isFalse && !covered[guarantee.name] {
			return fmt.Errorf("false guarantee %q requires a named limitation", guarantee.name)
		}
	}
	return nil
}

func validateCertificationLimitation(limitation certificationDeliveryLimitation) error {
	return validateCertificationReason(limitation.Code, limitation.Reason)
}

func validateCertificationReason(code, reason string) error {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "", "n/a", "na", "blocked", "not_applicable", "not-applicable":
		return fmt.Errorf("reason code %q is generic", code)
	}
	if strings.TrimSpace(reason) == "" {
		return errors.New("reason explanation is required")
	}
	return nil
}

func validateHTTPFields(fields []certificationHTTPField) error {
	for _, field := range fields {
		if (!safeProofFieldName(field.Name) && !certificationFingerprintSequence(field.Name)) || !certificationFingerprintSequence(field.Value) {
			return errors.New("field name or fingerprint is invalid")
		}
	}
	return nil
}

func validateQueries(queries []certificationQuery) error {
	for _, query := range queries {
		if !safeProofFieldName(query.Name) || !certificationFingerprintSequence(query.Value) {
			return errors.New("query name or fingerprint is invalid")
		}
	}
	return nil
}

func validateHTTPBody(body certificationHTTPBody) error {
	if body.OriginalBytes < 0 {
		return errors.New("body length is invalid")
	}
	switch body.Encoding {
	case "none":
		if string(body.Value) != "null" || body.OriginalBytes != 0 || body.Truncated {
			return errors.New("none body is invalid")
		}
		return nil
	case "opaque":
		var value string
		if err := decodeCertificationJSON(body.Value, &value); err != nil || !certificationFingerprintSequence(value) {
			return errors.New("opaque body contains an unproved value")
		}
		return nil
	case "json":
		decoder := json.NewDecoder(bytes.NewReader(body.Value))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return errors.New("json body is invalid")
		}
		if err := decoder.Decode(&struct{}{}); err != io.EOF {
			return errors.New("json body is invalid")
		}
		if !sanitizedProofJSONValue(value) {
			return errors.New("json body contains an unproved value")
		}
		return nil
	default:
		return fmt.Errorf("body encoding %q is unsupported", body.Encoding)
	}
}

func sortedUniqueFingerprints(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for index, value := range values {
		if !certificationFingerprintSequence(value) || index > 0 && values[index-1] >= value {
			return false
		}
	}
	return true
}

func safeHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS":
		return true
	default:
		return false
	}
}

func safeCertificationID(value string) bool {
	return len(value) <= 200 && safeCertificationIdentifier.MatchString(value) && strings.IndexFunc(value, unicode.IsControl) < 0
}

func safeProofFieldName(value string) bool {
	if value == "" || len(value) > 200 {
		return false
	}
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' || character == '.' {
			continue
		}
		return false
	}
	return true
}

func certificationFingerprintSequence(value string) bool {
	if value == "" {
		return false
	}
	for value != "" {
		if !strings.HasPrefix(value, "{{pmcertfp:v1:") {
			return false
		}
		end := strings.Index(value[len("{{pmcertfp:v1:"):], "}}")
		if end < 0 {
			return false
		}
		digest := value[len("{{pmcertfp:v1:") : len("{{pmcertfp:v1:")+end]
		if !sha256Digest.MatchString(digest) {
			return false
		}
		value = value[len("{{pmcertfp:v1:")+end+len("}}"):]
	}
	return true
}

func sanitizedProofJSONValue(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if !safeProofFieldName(key) && !certificationFingerprintSequence(key) {
				return false
			}
			if !sanitizedProofJSONValue(child) {
				return false
			}
		}
		return true
	case []any:
		for _, child := range typed {
			if !sanitizedProofJSONValue(child) {
				return false
			}
		}
		return true
	case nil:
		return true
	case string:
		return certificationFingerprintSequence(typed)
	default:
		return false
	}
}

func safeEvidenceRecordPath(record, directory string) bool {
	if !fs.ValidPath(record) || pathpkg.Clean(record) != record || pathpkg.IsAbs(record) || !strings.HasPrefix(record, directory+"/") || !strings.HasSuffix(record, ".json") {
		return false
	}
	return safeCertificationID(strings.TrimSuffix(pathpkg.Base(record), ".json"))
}

func evidencePointerMatchesRecord(pointer certificationEvidencePointer, record certificationAcceptedEvidence) bool {
	return pointer.Provider == record.Provider && pointer.ExecutedAt == record.ExecutedAt &&
		pointer.RunID == record.RunID && pointer.CredentialScope == record.CredentialScope &&
		pointer.CredentialNote == record.CredentialNote && pointer.CredentialScopeProof == record.CredentialScopeProof && certificationProofEqual(pointer.Proof, record.Proof)
}

func certificationProofEqual(left, right certificationEvidenceProof) bool {
	leftValue, leftErr := certificationProofValue(left)
	rightValue, rightErr := certificationProofValue(right)
	return leftErr == nil && rightErr == nil && reflect.DeepEqual(leftValue, rightValue)
}

func certificationProofValue(proof certificationEvidenceProof) (any, error) {
	raw, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return nil, errors.New("multiple JSON values")
		}
		return nil, err
	}
	return value, nil
}

type certificationEvidenceBinding struct {
	Scope        string
	Connector    string
	FunctionKind string
	WorkflowKind string
	SyncMode     string
	Primitive    string
	Source       string
	Destination  string
	FlowKind     string
}

func (binding certificationEvidenceBinding) matches(record certificationAcceptedEvidence) error {
	if record.Scope != binding.Scope {
		return fmt.Errorf("evidence scope %q does not match %q", record.Scope, binding.Scope)
	}
	switch binding.Scope {
	case "capability":
		if record.Connector != binding.Connector || record.FunctionKind != binding.FunctionKind {
			return errors.New("capability evidence names a different cell")
		}
	case "workflow":
		if record.Connector != binding.Connector || record.WorkflowKind != binding.WorkflowKind {
			return errors.New("workflow evidence names a different cell")
		}
	case "sync_mode":
		if record.Connector != binding.Connector || record.SyncMode != binding.SyncMode || record.Primitive != binding.Primitive {
			return errors.New("sync-mode evidence names a different cell")
		}
	case "flow":
		if record.Source != binding.Source || record.Destination != binding.Destination || record.FlowKind != binding.FlowKind {
			return errors.New("flow evidence names a different pair")
		}
	default:
		return errors.New("evidence binding scope is unsupported")
	}
	return nil
}

func findCapabilityConnector(connectors []certificationCapabilityConnector, name string) (certificationCapabilityConnector, bool) {
	for _, connector := range connectors {
		if connector.Name == name {
			return connector, true
		}
	}
	return certificationCapabilityConnector{}, false
}

func findWorkflowSet(workflows []certificationWorkflowSet, connector string) (certificationWorkflowSet, bool) {
	for _, workflow := range workflows {
		if workflow.Connector == connector {
			return workflow, true
		}
	}
	return certificationWorkflowSet{}, false
}

func findSyncSet(sets []certificationSyncModeSet, connector string) (certificationSyncModeSet, bool) {
	for _, set := range sets {
		if set.Connector == connector {
			return set, true
		}
	}
	return certificationSyncModeSet{}, false
}

func findConnectorRoles(roles []certificationConnectorRoles, connector string) (certificationConnectorRoles, bool) {
	for _, set := range roles {
		if set.Connector == connector {
			return set, true
		}
	}
	return certificationConnectorRoles{}, false
}

func findStatus(statuses []certificationConnectorStatus, connector string) (certificationConnectorStatus, bool) {
	for _, status := range statuses {
		if status.Connector == connector {
			return status, true
		}
	}
	return certificationConnectorStatus{}, false
}

type certificationFlowPair struct {
	FlowKind    string
	Source      string
	Destination string
	Cell        certificationFlowCell
}

func (matrix certificationFlowMatrix) resolvedFlowPair(flowKind, source, destination string) (certificationFlowPair, bool) {
	for _, override := range matrix.PairOverrides {
		if override.FlowKind == flowKind && override.Source == source && override.Destination == destination {
			return certificationFlowPair{FlowKind: flowKind, Source: source, Destination: destination, Cell: override.Cell}, true
		}
	}
	set, ok := certificationFlowPairSetFor(matrix.PairSets, flowKind, source, destination)
	if !ok {
		return certificationFlowPair{}, false
	}
	return certificationFlowPair{FlowKind: flowKind, Source: source, Destination: destination, Cell: set.Cell}, true
}

func (matrix certificationFlowMatrix) flowPairsForConnector(connector string) ([]certificationFlowPair, error) {
	knownKinds := make(map[string]bool, len(matrix.FlowKinds))
	for _, kind := range matrix.FlowKinds {
		knownKinds[kind.ID] = true
	}
	pairs := make(map[string]certificationFlowPair)
	for _, set := range matrix.PairSets {
		if !knownKinds[set.FlowKind] || set.Mediator != matrix.Mediator || len(set.SourceConnectors) == 0 || len(set.DestinationConnectors) == 0 {
			return nil, errors.New("flow pair set is invalid")
		}
		if slices.Contains(set.SourceConnectors, connector) {
			for _, destination := range set.DestinationConnectors {
				if err := addFlowPair(pairs, certificationFlowPair{FlowKind: set.FlowKind, Source: connector, Destination: destination, Cell: set.Cell}); err != nil {
					return nil, err
				}
			}
		}
		if slices.Contains(set.DestinationConnectors, connector) {
			for _, source := range set.SourceConnectors {
				if err := addFlowPair(pairs, certificationFlowPair{FlowKind: set.FlowKind, Source: source, Destination: connector, Cell: set.Cell}); err != nil {
					return nil, err
				}
			}
		}
	}
	overrides := make(map[string]certificationFlowPairOverride, len(matrix.PairOverrides))
	for _, override := range matrix.PairOverrides {
		key := flowPairKey(override.FlowKind, override.Source, override.Destination)
		if !knownKinds[override.FlowKind] || override.Mediator != matrix.Mediator || !safeCertificationID(override.Source) || !safeCertificationID(override.Destination) || overrides[key].FlowKind != "" {
			return nil, errors.New("flow pair override is invalid")
		}
		overrides[key] = override
	}
	for key, pair := range pairs {
		if override, ok := overrides[key]; ok {
			pair.Cell = override.Cell
			pairs[key] = pair
		}
	}
	result := make([]certificationFlowPair, 0, len(pairs))
	for _, pair := range pairs {
		result = append(result, pair)
	}
	sort.Slice(result, func(left, right int) bool {
		return flowPairKey(result[left].FlowKind, result[left].Source, result[left].Destination) < flowPairKey(result[right].FlowKind, result[right].Source, result[right].Destination)
	})
	return result, nil
}

func addFlowPair(pairs map[string]certificationFlowPair, pair certificationFlowPair) error {
	if !safeCertificationID(pair.FlowKind) || !safeCertificationID(pair.Source) || !safeCertificationID(pair.Destination) {
		return errors.New("flow pair contains an unsafe identifier")
	}
	key := flowPairKey(pair.FlowKind, pair.Source, pair.Destination)
	if existing, exists := pairs[key]; exists {
		// A self-pair appears once through the source expansion and once through the
		// destination expansion of the same compact set. It is one exact cell, not
		// two requirements. Any non-identical overlap remains unsafe.
		if reflect.DeepEqual(existing, pair) {
			return nil
		}
		return fmt.Errorf("flow pair %s is duplicated", key)
	}
	pairs[key] = pair
	return nil
}

func flowPairKey(kind, source, destination string) string {
	return kind + "\x00" + source + "\x00" + destination
}

const (
	certificationGateBeginMarker = "<!-- BEGIN POLYMETRICS CONNECTOR CERTIFICATION SHEPHERD GATE -->"
	certificationGateEndMarker   = "<!-- END POLYMETRICS CONNECTOR CERTIFICATION SHEPHERD GATE -->"
)

// RenderCertificationGateIO is the single adapter-neutral instruction block. Every generated
// harness embeds this exact text, which makes a missing input or verdict field projection drift.
func RenderCertificationGateIO(contract *Contract) ([]byte, error) {
	if contract == nil {
		return nil, errors.New("canonical contract is required")
	}
	if err := contract.Validate(); err != nil {
		return nil, err
	}
	gate := contract.CertificationGate
	commandArgv, err := marshalArgv(gate.Command.Argv)
	if err != nil {
		return nil, err
	}
	var output bytes.Buffer
	fmt.Fprintln(&output, certificationGateBeginMarker)
	fmt.Fprintln(&output, "## Connector certification Shepherd gate")
	fmt.Fprintln(&output)
	if gate.Inputs.CertificationShards != "" {
		fmt.Fprintf(&output, "This is the versioned, read-only `%s` gate. It reads only definition-owned certification shards below `%s`, `%s`, and accepted records below `%s`; it never creates evidence, loads credentials, invokes a provider, or mutates provider/production state.\n\n", gate.Name, gate.Inputs.CertificationShards, gate.Inputs.Status, gate.Inputs.EvidenceDirectory)
	} else {
		fmt.Fprintf(&output, "This is the versioned, read-only `%s` gate. It reads only `%s`, `%s`, `%s`, and accepted records below `%s`; it never creates evidence, loads credentials, invokes a provider, or mutates provider/production state.\n\n", gate.Name, gate.Inputs.CapabilityMatrix, gate.Inputs.FlowMatrix, gate.Inputs.Status, gate.Inputs.EvidenceDirectory)
	}
	fmt.Fprintf(&output, "- Enforce a `PROCEED` verdict before `%s`.\n", joinNatural(gate.EnforcedTransitions))
	fmt.Fprintf(&output, "- Run argv `%s` at the transition boundary. %s\n", commandArgv, gate.Command.Instruction)
	fmt.Fprintf(&output, "- Input schema v%d requires every field: %s.\n", gate.InputSchemaVersion, joinInlineCode(gate.InputFields))
	fmt.Fprintf(&output, "- The nested `inputs` values must exactly equal the canonical paths above; no adapter-local default or replacement is allowed.\n")
	fmt.Fprintf(&output, "- Every applicable capability, workflow, sync-mode primitive, and flow pair needs %s plus a matching accepted live-evidence record. File presence, reachability, or `implemented` alone cannot pass.\n", joinNatural(gate.BindingCriteria))
	fmt.Fprintf(&output, "- Verdict schema v%d contains every field: %s. Allowed decisions: %s.\n", gate.VerdictSchemaVersion, joinInlineCode(gate.VerdictFields), joinInlineCode(gate.Verdicts))
	fmt.Fprintln(&output, "- Preserve every exact `failures[].id`, `cell_id`, and `evidence_id` in a `RETRY` or `HALT` handoff; do not replace them with prose.")
	fmt.Fprintf(&output, "- An unknown/missing matrix, evidence schema, proof redaction strategy, or adapter field is `%s`; do not invent #3989 proof fields.\n", gate.UnsupportedProofSchemaBehavior)
	fmt.Fprintln(&output, certificationGateEndMarker)
	fmt.Fprintln(&output)
	return output.Bytes(), nil
}
