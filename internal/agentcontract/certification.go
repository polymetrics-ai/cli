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
)

const (
	certificationGateSchemaVersion = 1
	certificationDecisionProceed   = CertificationGateDecision("PROCEED")
	certificationDecisionRetry     = CertificationGateDecision("RETRY")
	certificationDecisionHalt      = CertificationGateDecision("HALT")

	fullParityCredentialScope = "full_parity"
	fullParityCredentialNote  = "Certification used a full-parity credential; a narrower credential exposes a subset of this certified surface."
	localParquetWarehouse     = "local_parquet_warehouse"
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

	capabilities, err := loadCapabilityMatrix(root, gate)
	if err != nil {
		return certificationGateHalt(request, certificationInputFailureID("capability_matrix", err), "input", err.Error()), nil
	}
	flow, err := loadFlowMatrix(root, gate)
	if err != nil {
		return certificationGateHalt(request, certificationInputFailureID("flow_matrix", err), "input", err.Error()), nil
	}
	status, err := loadStatusArtifact(root, gate)
	if err != nil {
		return certificationGateHalt(request, certificationInputFailureID("status", err), "input", err.Error()), nil
	}
	evidence, err := loadAcceptedCertificationEvidence(root, gate)
	if err != nil {
		return certificationGateHalt(request, certificationInputFailureID("evidence", err), "input", err.Error()), nil
	}

	if err := validateCapabilityMatrix(capabilities, gate); err != nil {
		return certificationGateHalt(request, "capability_matrix/invalid", "input", err.Error()), nil
	}
	if err := validateFlowMatrix(flow, gate); err != nil {
		return certificationGateHalt(request, "flow_matrix/invalid", "input", err.Error()), nil
	}
	if err := validateStatusArtifact(status, gate); err != nil {
		return certificationGateHalt(request, "status/invalid", "input", err.Error()), nil
	}

	artifacts := certificationArtifacts{
		capabilities: capabilities,
		flow:         flow,
		status:       status,
		evidence:     evidence,
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

	seen := make(map[string]bool, len(facts.LiveEvidence))
	for _, pointer := range facts.LiveEvidence {
		if pointer.Record == "" || seen[pointer.Record] {
			return certificationGateHaltAt(CertificationGateRequest{}, cellID+"/live_evidence", "evidence", cellID, pointer.Record, "live evidence pointer is empty or duplicated"), true
		}
		seen[pointer.Record] = true
		record, ok := artifacts.evidence[pointer.Record]
		if !ok {
			return certificationGateHaltAt(CertificationGateRequest{}, "evidence/"+pointer.Record+"/missing", "evidence", cellID, pointer.Record, "referenced live evidence record is missing"), true
		}
		if err := validateEvidencePointer(pointer, gate); err != nil {
			return certificationGateHaltAt(CertificationGateRequest{}, "evidence/"+pointer.Record+"/invalid", "evidence", cellID, pointer.Record, err.Error()), true
		}
		if !evidencePointerMatchesRecord(pointer, record) {
			return certificationGateHaltAt(CertificationGateRequest{}, "evidence/"+pointer.Record+"/mismatch", "evidence", cellID, pointer.Record, "matrix pointer and accepted evidence record differ"), true
		}
		if err := binding.matches(record); err != nil {
			return certificationGateHaltAt(CertificationGateRequest{}, "evidence/"+pointer.Record+"/binding", "evidence", cellID, pointer.Record, err.Error()), true
		}
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
	Record          string                     `json:"record"`
	Provider        string                     `json:"provider"`
	ExecutedAt      string                     `json:"executed_at"`
	RunID           string                     `json:"run_id"`
	CredentialScope string                     `json:"credential_scope"`
	CredentialNote  string                     `json:"credential_note"`
	Proof           certificationEvidenceProof `json:"proof"`
}

type certificationAcceptedEvidence struct {
	SchemaVersion   int                        `json:"schema_version"`
	Scope           string                     `json:"scope"`
	Status          string                     `json:"status"`
	CredentialScope string                     `json:"credential_scope"`
	CredentialNote  string                     `json:"credential_note"`
	Connector       string                     `json:"connector,omitempty"`
	FunctionKind    string                     `json:"function_kind,omitempty"`
	WorkflowKind    string                     `json:"workflow_kind,omitempty"`
	SyncMode        string                     `json:"sync_mode,omitempty"`
	Primitive       string                     `json:"primitive,omitempty"`
	Source          string                     `json:"source,omitempty"`
	Destination     string                     `json:"destination,omitempty"`
	FlowKind        string                     `json:"flow_kind,omitempty"`
	Provider        string                     `json:"provider"`
	ExecutedAt      string                     `json:"executed_at"`
	RunID           string                     `json:"run_id"`
	Proof           certificationEvidenceProof `json:"proof"`
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
	SchemaVersion    int                            `json:"schema_version"`
	GeneratedCommand string                         `json:"generated_command"`
	Connectors       []certificationConnectorStatus `json:"connectors"`
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
	if !fs.ValidPath(relativePath) || pathpkg.Clean(relativePath) != relativePath || pathpkg.IsAbs(relativePath) {
		return fmt.Errorf("certification input path %q is not local", relativePath)
	}
	raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relativePath)))
	if err != nil {
		return err
	}
	if err := decodeCertificationJSON(raw, destination); err != nil {
		return err
	}
	return nil
}

func loadAcceptedCertificationEvidence(root string, gate ConnectorCertificationGate) (map[string]certificationAcceptedEvidence, error) {
	directory := gate.Inputs.EvidenceDirectory
	if !fs.ValidPath(directory) || pathpkg.Clean(directory) != directory || pathpkg.IsAbs(directory) {
		return nil, fmt.Errorf("evidence directory %q is not local", directory)
	}
	entries, err := os.ReadDir(filepath.Join(root, filepath.FromSlash(directory)))
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]certificationAcceptedEvidence{}, nil
	}
	if err != nil {
		return nil, err
	}
	records := make(map[string]certificationAcceptedEvidence, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		if entry.Type()&fs.ModeSymlink != 0 || !safeCertificationID(strings.TrimSuffix(entry.Name(), ".json")) {
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
	return validateUniqueIDs("capability connector", matrix.Connectors, func(connector certificationCapabilityConnector) string { return connector.Name })
}

func validateFlowMatrix(matrix certificationFlowMatrix, gate ConnectorCertificationGate) error {
	if matrix.SchemaVersion != gate.SchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", matrix.SchemaVersion)
	}
	if matrix.GeneratedCommand != gate.GeneratedCommand || matrix.Mediator != localParquetWarehouse {
		return errors.New("generated command or warehouse mediator is unsupported")
	}
	if len(matrix.FlowKinds) == 0 || len(matrix.WorkflowKinds) == 0 || len(matrix.Workflows) == 0 || len(matrix.SyncModeCells) == 0 || len(matrix.ConnectorRoles) == 0 {
		return errors.New("flow kinds, workflows, sync-mode cells, and connector roles are required")
	}
	if err := validateUniqueIDs("flow kind", matrix.FlowKinds, func(kind certificationFlowKind) string { return kind.ID }); err != nil {
		return err
	}
	if err := validateUniqueIDs("workflow kind", matrix.WorkflowKinds, func(kind certificationWorkflowKind) string { return kind.ID }); err != nil {
		return err
	}
	if err := validateUniqueIDs("workflow connector", matrix.Workflows, func(set certificationWorkflowSet) string { return set.Connector }); err != nil {
		return err
	}
	if err := validateUniqueIDs("sync-mode connector", matrix.SyncModeCells, func(set certificationSyncModeSet) string { return set.Connector }); err != nil {
		return err
	}
	if err := validateUniqueIDs("flow-role connector", matrix.ConnectorRoles, func(roles certificationConnectorRoles) string { return roles.Connector }); err != nil {
		return err
	}
	return validateUniqueIDs("flow status", matrix.ConnectorStatuses, func(status certificationConnectorStatus) string { return status.Connector })
}

func validateStatusArtifact(artifact certificationStatusArtifact, gate ConnectorCertificationGate) error {
	if artifact.SchemaVersion != gate.SchemaVersion {
		return fmt.Errorf("schema_version %d is unsupported", artifact.SchemaVersion)
	}
	if artifact.GeneratedCommand != gate.GeneratedCommand || len(artifact.Connectors) == 0 {
		return errors.New("generated command or connector statuses are invalid")
	}
	return validateUniqueIDs("status connector", artifact.Connectors, func(status certificationConnectorStatus) string { return status.Connector })
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
	if err := validateEvidenceIdentity(record.Provider, record.ExecutedAt, record.RunID, record.CredentialScope, record.CredentialNote); err != nil {
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
	if err := validateEvidenceIdentity(pointer.Provider, pointer.ExecutedAt, pointer.RunID, pointer.CredentialScope, pointer.CredentialNote); err != nil {
		return err
	}
	return validateEvidenceProof(pointer.Proof, gate)
}

func validateEvidenceIdentity(provider, executedAt, runID, credentialScope, credentialNote string) error {
	if !safeCertificationID(provider) || !safeCertificationID(runID) {
		return errors.New("provider and run_id must be safe identifiers")
	}
	if _, err := time.Parse(time.RFC3339, executedAt); err != nil {
		return fmt.Errorf("executed_at must be RFC3339: %w", err)
	}
	if credentialScope != fullParityCredentialScope || credentialNote != fullParityCredentialNote {
		return errors.New("evidence credential scope is not full parity")
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
		if !flow.Delivery.Resumable || !flow.Delivery.ReceiptBacked || !flow.Delivery.Checkpointed || !flow.Delivery.ReplayIdentity || !flow.Delivery.ProviderIdempotencyKey {
			return errors.New("flow proof delivery guarantees are incomplete")
		}
		for _, limitation := range flow.Delivery.Limitations {
			if !safeCertificationID(limitation.Guarantee) || !safeCertificationID(limitation.Code) || strings.TrimSpace(limitation.Reason) == "" {
				return errors.New("flow proof limitation is invalid")
			}
		}
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
		pointer.CredentialNote == record.CredentialNote && certificationProofEqual(pointer.Proof, record.Proof)
}

func certificationProofEqual(left, right certificationEvidenceProof) bool {
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
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
	var output bytes.Buffer
	fmt.Fprintln(&output, certificationGateBeginMarker)
	fmt.Fprintln(&output, "## Connector certification Shepherd gate")
	fmt.Fprintln(&output)
	fmt.Fprintf(&output, "This is the versioned, read-only `%s` gate. It reads only `%s`, `%s`, `%s`, and accepted records below `%s`; it never creates evidence, loads credentials, invokes a provider, or mutates provider/production state.\n\n", gate.Name, gate.Inputs.CapabilityMatrix, gate.Inputs.FlowMatrix, gate.Inputs.Status, gate.Inputs.EvidenceDirectory)
	fmt.Fprintf(&output, "- Enforce a `PROCEED` verdict before `%s`.\n", joinNatural(gate.EnforcedTransitions))
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
