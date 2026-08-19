package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/certify"
)

const (
	proofRedactionStrategy = "repository_salted_hmac_sha256_v1"
	fingerprintPrefix      = "{{pmcertfp:v1:"
	fingerprintSuffix      = "}}"

	acceptedEvidenceSchemaVersion         = 2
	credentialScopeFullParity             = "full_parity"
	credentialScopeObservedOperations     = "observed_operations"
	credentialScopeProofFullParityStage   = "full_parity_stage"
	credentialScopeProofProtocolExchanges = "protocol_exchanges"
	fullParityCredentialNote              = "Certification used a full-parity credential; a narrower credential exposes a subset of this certified surface."
	observedOperationsCredentialNote      = "Only the credential use documented by this record's protocol exchanges was verified; no broader credential scope is claimed."
	localWarehouseMediator                = "local_parquet_warehouse"

	// repositoryFingerprintSaltPath is intentionally ignored by git.  A clone
	// creates a new value, while replay on the same checkout uses the same
	// value.  That makes a fingerprint useful for local re-verification without
	// making it comparable between installations.
	repositoryFingerprintSaltPath = "internal/connectors/certifications/.fingerprint-salt"

	maxProofBodyBytes = 1 << 20
)

// embeddedEvidenceProof is the publishable, proof-bearing portion of an
// accepted live result. It contains the complete request/response transcript
// shape, but every actual value is transformed before JSON serialization. The
// repository salt and the prepared credential values are deliberately absent.
type embeddedEvidenceProof struct {
	RedactionStrategy    string `json:"redaction_strategy"`
	PMBinarySHA256       string `json:"pm_binary_sha256"`
	PMCommandFingerprint string `json:"pm_command_fingerprint"`
	// CredentialFingerprints are the exact prepared credential values, hashed
	// individually.  They let a local replay compare its credential without
	// ever writing the value (or an encrypted form of it) to the record.
	CredentialFingerprints []string                    `json:"credential_fingerprints"`
	HTTPExchanges          []certifiedHTTPExchange     `json:"http_exchanges"`
	DatabaseExchanges      []certifiedDatabaseExchange `json:"database_exchanges"`
	Flow                   *certifiedFlowRoundTrip     `json:"flow,omitempty"`
}

type certifiedHTTPExchange struct {
	Operation string                `json:"operation"`
	Request   certifiedHTTPRequest  `json:"request"`
	Response  certifiedHTTPResponse `json:"response"`
}

// certifiedHTTPRequest retains the request's actual method, endpoint shape,
// headers, query, and body. Values are a deterministic sequence of repository
// fingerprints; parameter/header names remain readable structural labels.
type certifiedHTTPRequest struct {
	Method  string               `json:"method"`
	Target  string               `json:"target"`
	Query   []certifiedQuery     `json:"query"`
	Headers []certifiedHTTPField `json:"headers"`
	Body    certifiedHTTPBody    `json:"body"`
}

type certifiedHTTPResponse struct {
	Status  int                  `json:"status"`
	Headers []certifiedHTTPField `json:"headers"`
	Body    certifiedHTTPBody    `json:"body"`
}

type certifiedHTTPField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type certifiedQuery struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// certifiedHTTPBody preserves a JSON body's object/array shape and field names
// while fingerprinting every scalar value. Opaque bytes become one fingerprint
// value. The raw body never reaches a serialization API.
type certifiedHTTPBody struct {
	Encoding      string          `json:"encoding"`
	Value         json.RawMessage `json:"value"`
	OriginalBytes int             `json:"original_bytes"`
	Truncated     bool            `json:"truncated"`
}

// certifiedDatabaseExchange gives non-HTTP connectors the same proof-bearing
// path as API connectors. Statements, parameter values, and result bodies are
// transformed before the exchange can be serialized.
type certifiedDatabaseExchange struct {
	Operation string                    `json:"operation"`
	Protocol  string                    `json:"protocol"`
	Request   certifiedDatabaseRequest  `json:"request"`
	Response  certifiedDatabaseResponse `json:"response"`
}

type certifiedDatabaseRequest struct {
	Statement  string   `json:"statement"`
	Parameters []string `json:"parameters"`
}

type certifiedDatabaseResponse struct {
	Status string            `json:"status"`
	Body   certifiedHTTPBody `json:"body"`
}

// certifiedFlowRoundTrip points at independently captured transcript
// operations while retaining the delivery facts that must not be inferred from
// a successful one-shot mutation.
type certifiedFlowRoundTrip struct {
	PMCommandFingerprint         string             `json:"pm_command_fingerprint"`
	Mediator                     string             `json:"mediator"`
	WarehouseReadbackOperation   string             `json:"warehouse_readback_operation"`
	DestinationReadbackOperation string             `json:"destination_readback_operation"`
	Delivery                     deliveryGuarantees `json:"delivery"`
}

type deliveryGuarantees struct {
	Resumable              bool                 `json:"resumable"`
	ReceiptBacked          bool                 `json:"receipt_backed"`
	Checkpointed           bool                 `json:"checkpointed"`
	ReplayIdentity         bool                 `json:"replay_identity"`
	ProviderIdempotencyKey bool                 `json:"provider_idempotency_key"`
	Limitations            []deliveryLimitation `json:"limitations"`
}

type deliveryLimitation struct {
	Guarantee string `json:"guarantee"`
	Code      string `json:"code"`
	Reason    string `json:"reason"`
}

// completedLiveEvidence is intentionally an in-memory boundary. A live runner
// builds it only after the pm invocation completes. RepositorySalt and
// PreparedValues carry sensitive material only long enough to derive the
// publishable proof; they are never JSON fields and are never written directly.
type completedLiveEvidence struct {
	SchemaVersion  int
	Scope          string
	Connector      string
	FunctionKind   string
	WorkflowKind   string
	SyncMode       string
	Primitive      string
	Source         string
	Destination    string
	FlowKind       string
	Provider       string
	ExecutedAt     string
	RunID          string
	PMBinarySHA256 string
	PMCommand      string
	Passed         bool

	RepositorySalt    []byte
	PreparedValues    []string
	HTTPExchanges     []completedHTTPExchange
	DatabaseExchanges []completedDatabaseExchange
	Flow              *completedFlowRoundTrip
}

// credentialScopeClaim is intentionally constructed only from proof-bearing
// run state. Callers cannot provide arbitrary scope strings, notes, or proof
// discriminators to an accepted-evidence record.
type credentialScopeClaim struct {
	scope string
	note  string
	proof string
}

func observedOperationsCredentialScope() credentialScopeClaim {
	return credentialScopeClaim{
		scope: credentialScopeObservedOperations,
		note:  observedOperationsCredentialNote,
		proof: credentialScopeProofProtocolExchanges,
	}
}

func fullParityCredentialScope(report certify.Report) (credentialScopeClaim, error) {
	if !report.FullParityVerified() {
		return credentialScopeClaim{}, errors.New("completed certification report did not pass the full-parity stage")
	}
	return credentialScopeClaim{
		scope: credentialScopeFullParity,
		note:  fullParityCredentialNote,
		proof: credentialScopeProofFullParityStage,
	}, nil
}

type completedHTTPExchange struct {
	Operation string
	Request   completedHTTPRequest
	Response  completedHTTPResponse
}

type completedHTTPRequest struct {
	Method  string
	Target  string
	Headers map[string][]string
	Body    []byte
}

type completedHTTPResponse struct {
	Status  int
	Headers map[string][]string
	Body    []byte
}

type completedDatabaseExchange struct {
	Operation      string
	Protocol       string
	Statement      string
	Parameters     []string
	ResponseStatus string
	ResponseBody   []byte
}

type completedFlowRoundTrip struct {
	PMCommand                    string
	WarehouseReadbackOperation   string
	DestinationReadbackOperation string
	Delivery                     deliveryGuarantees
}

// newProofBearingEvidence is the only construction path for a live result.
// It refuses incomplete/failed runs and converts raw exchange material before
// it can be persisted. A caller never receives a serializable type containing
// a raw credential or response body.
func newProofBearingEvidence(completed completedLiveEvidence) (acceptedEvidence, error) {
	return newProofBearingEvidenceForCredentialScope(completed, observedOperationsCredentialScope())
}

// newFullParityProofBearingEvidence is the sole construction path for a
// full-parity claim. The claim is derived from the executed report's passing
// stage, not from a caller attestation or command-line spelling.
func newFullParityProofBearingEvidence(completed completedLiveEvidence, report certify.Report) (acceptedEvidence, error) {
	claim, err := fullParityCredentialScope(report)
	if err != nil {
		return acceptedEvidence{}, err
	}
	return newProofBearingEvidenceForCredentialScope(completed, claim)
}

func newProofBearingEvidenceForCredentialScope(completed completedLiveEvidence, claim credentialScopeClaim) (acceptedEvidence, error) {
	if !completed.Passed {
		return acceptedEvidence{}, errors.New("completed live test did not pass")
	}
	if len(completed.RepositorySalt) < 16 {
		return acceptedEvidence{}, errors.New("repository fingerprint salt must be at least 16 bytes")
	}
	if len(normalizedPreparedValues(completed.PreparedValues)) == 0 {
		return acceptedEvidence{}, errors.New("completed live test has no prepared credential values to fingerprint")
	}
	if len(completed.HTTPExchanges)+len(completed.DatabaseExchanges) == 0 {
		return acceptedEvidence{}, errors.New("completed live test has no protocol exchanges")
	}
	if strings.TrimSpace(completed.PMCommand) == "" {
		return acceptedEvidence{}, errors.New("completed live test has no pm command")
	}

	proof := embeddedEvidenceProof{
		RedactionStrategy:      proofRedactionStrategy,
		PMBinarySHA256:         completed.PMBinarySHA256,
		PMCommandFingerprint:   fingerprintText(completed.PMCommand, completed.PreparedValues, completed.RepositorySalt),
		CredentialFingerprints: fingerprintPreparedValues(completed.PreparedValues, completed.RepositorySalt),
		HTTPExchanges:          make([]certifiedHTTPExchange, 0, len(completed.HTTPExchanges)),
		DatabaseExchanges:      make([]certifiedDatabaseExchange, 0, len(completed.DatabaseExchanges)),
	}
	for _, exchange := range completed.HTTPExchanges {
		sanitized, err := sanitizeHTTPExchange(exchange, completed.PreparedValues, completed.RepositorySalt)
		if err != nil {
			return acceptedEvidence{}, err
		}
		proof.HTTPExchanges = append(proof.HTTPExchanges, sanitized)
	}
	for _, exchange := range completed.DatabaseExchanges {
		sanitized, err := sanitizeDatabaseExchange(exchange, completed.PreparedValues, completed.RepositorySalt)
		if err != nil {
			return acceptedEvidence{}, err
		}
		proof.DatabaseExchanges = append(proof.DatabaseExchanges, sanitized)
	}
	if completed.Flow != nil {
		flow, err := sanitizeFlowRoundTrip(*completed.Flow, completed.PreparedValues, completed.RepositorySalt)
		if err != nil {
			return acceptedEvidence{}, err
		}
		proof.Flow = &flow
	}

	evidence := acceptedEvidence{
		SchemaVersion:        acceptedEvidenceSchemaVersion,
		Scope:                completed.Scope,
		Status:               evidenceStatusPassed,
		CredentialScope:      claim.scope,
		CredentialNote:       claim.note,
		CredentialScopeProof: claim.proof,
		Connector:            completed.Connector,
		FunctionKind:         completed.FunctionKind,
		WorkflowKind:         completed.WorkflowKind,
		SyncMode:             completed.SyncMode,
		Primitive:            completed.Primitive,
		Source:               completed.Source,
		Destination:          completed.Destination,
		FlowKind:             completed.FlowKind,
		Provider:             completed.Provider,
		ExecutedAt:           completed.ExecutedAt,
		RunID:                completed.RunID,
		Proof:                proof,
	}
	if err := validateAcceptedEvidence(evidence); err != nil {
		return acceptedEvidence{}, err
	}
	return evidence, nil
}

// preparedAcceptedEvidence is a fully rendered and validated record whose
// destination is safe for publication. Preparing every record before calling
// publishPreparedAcceptedEvidence keeps validation failures from producing a
// prefix of a multi-record import.
type preparedAcceptedEvidence struct {
	outputPath string
	evidence   acceptedEvidence
	payload    []byte
}

// prepareProofBearingEvidence is deliberately fed only a completed run. It
// owns the repository-local salt and renders the sanitized result in memory;
// raw request, response, or credential data has no persistence path.
func prepareProofBearingEvidence(repoRoot, path string, completed completedLiveEvidence) (preparedAcceptedEvidence, error) {
	outputPath, err := acceptedEvidenceOutputPath(repoRoot, path)
	if err != nil {
		return preparedAcceptedEvidence{}, err
	}
	salt, err := repositoryFingerprintSalt(repoRoot)
	if err != nil {
		return preparedAcceptedEvidence{}, err
	}
	completed.RepositorySalt = salt
	evidence, err := newProofBearingEvidence(completed)
	if err != nil {
		return preparedAcceptedEvidence{}, err
	}
	return prepareAcceptedEvidence(outputPath, evidence, "render proof-bearing evidence")
}

func writeProofBearingEvidence(repoRoot, path string, completed completedLiveEvidence) (acceptedEvidence, error) {
	prepared, err := prepareProofBearingEvidence(repoRoot, path, completed)
	if err != nil {
		return acceptedEvidence{}, err
	}
	if err := publishPreparedAcceptedEvidence([]preparedAcceptedEvidence{prepared}, nil); err != nil {
		return acceptedEvidence{}, err
	}
	return prepared.evidence, nil
}

// importedLiveEvidence contains only values that ReadExternalProof has already
// proved are fingerprints or structural metadata. Unlike completedLiveEvidence
// it deliberately has no raw credentials, request values, or response bodies.
type importedLiveEvidence struct {
	SchemaVersion          int
	Scope                  string
	Connector              string
	FunctionKind           string
	WorkflowKind           string
	SyncMode               string
	Primitive              string
	Source                 string
	Destination            string
	FlowKind               string
	Provider               string
	ExecutedAt             string
	RunID                  string
	PMBinarySHA256         string
	PMCommandFingerprint   string
	CredentialFingerprints []string
	HTTPExchanges          []certifiedHTTPExchange
}

// newImportedProofBearingEvidence accepts the safe projection from a completed
// external proof. It still validates the whole accepted-record contract before
// a file can be opened, so a future importer cannot turn an unredacted proof
// field into committed evidence by bypassing the raw-value sanitizer.
func newImportedProofBearingEvidence(completed importedLiveEvidence) (acceptedEvidence, error) {
	proof := embeddedEvidenceProof{
		RedactionStrategy:      proofRedactionStrategy,
		PMBinarySHA256:         completed.PMBinarySHA256,
		PMCommandFingerprint:   completed.PMCommandFingerprint,
		CredentialFingerprints: append([]string(nil), completed.CredentialFingerprints...),
		HTTPExchanges:          append([]certifiedHTTPExchange(nil), completed.HTTPExchanges...),
		DatabaseExchanges:      []certifiedDatabaseExchange{},
	}
	evidence := acceptedEvidence{
		SchemaVersion:        acceptedEvidenceSchemaVersion,
		Scope:                completed.Scope,
		Status:               evidenceStatusPassed,
		CredentialScope:      credentialScopeObservedOperations,
		CredentialNote:       observedOperationsCredentialNote,
		CredentialScopeProof: credentialScopeProofProtocolExchanges,
		Connector:            completed.Connector,
		FunctionKind:         completed.FunctionKind,
		WorkflowKind:         completed.WorkflowKind,
		SyncMode:             completed.SyncMode,
		Primitive:            completed.Primitive,
		Source:               completed.Source,
		Destination:          completed.Destination,
		FlowKind:             completed.FlowKind,
		Provider:             completed.Provider,
		ExecutedAt:           completed.ExecutedAt,
		RunID:                completed.RunID,
		Proof:                proof,
	}
	if err := validateAcceptedEvidence(evidence); err != nil {
		return acceptedEvidence{}, err
	}
	return evidence, nil
}

func prepareImportedProofBearingEvidence(repoRoot, path string, completed importedLiveEvidence) (preparedAcceptedEvidence, error) {
	outputPath, err := acceptedEvidenceOutputPath(repoRoot, path)
	if err != nil {
		return preparedAcceptedEvidence{}, err
	}
	evidence, err := newImportedProofBearingEvidence(completed)
	if err != nil {
		return preparedAcceptedEvidence{}, err
	}
	return prepareAcceptedEvidence(outputPath, evidence, "render imported proof-bearing evidence")
}

func prepareAcceptedEvidence(outputPath string, evidence acceptedEvidence, renderContext string) (preparedAcceptedEvidence, error) {
	if err := validateAcceptedEvidence(evidence); err != nil {
		return preparedAcceptedEvidence{}, err
	}
	payload, err := marshalGeneratedJSON(evidence)
	if err != nil {
		return preparedAcceptedEvidence{}, fmt.Errorf("%s: %w", renderContext, err)
	}
	return preparedAcceptedEvidence{outputPath: outputPath, evidence: evidence, payload: payload}, nil
}

// publishPreparedAcceptedEvidence writes each complete payload to a private
// file on the target filesystem, fsyncs it, and atomically links it into the
// evidence directory. os.Link refuses an existing destination, so publication
// is both no-replace and invisible to matrix readers until the complete JSON is
// durable. Callers must prepare the entire batch before publishing it.
//
// beforePublish is only used by the concurrent-reader regression test. It runs
// after every record has been staged but before any final name exists.
func publishPreparedAcceptedEvidence(records []preparedAcceptedEvidence, beforePublish func() error) error {
	if len(records) == 0 {
		return errors.New("no prepared evidence records to publish")
	}
	seen := make(map[string]struct{}, len(records))
	for _, record := range records {
		if record.outputPath == "" || len(record.payload) == 0 {
			return errors.New("prepared evidence requires a destination and complete payload")
		}
		if _, exists := seen[record.outputPath]; exists {
			return fmt.Errorf("prepared evidence has duplicate destination %q", filepath.ToSlash(record.outputPath))
		}
		seen[record.outputPath] = struct{}{}
		if err := os.MkdirAll(filepath.Dir(record.outputPath), 0o755); err != nil {
			return fmt.Errorf("create proof evidence directory: %w", err)
		}
		if _, err := os.Lstat(record.outputPath); err == nil {
			return fmt.Errorf("publish proof-bearing evidence %q: destination already exists", filepath.ToSlash(record.outputPath))
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect proof-bearing evidence destination %q: %w", filepath.ToSlash(record.outputPath), err)
		}
	}

	staged := make([]string, len(records))
	for index, record := range records {
		path, err := stageEvidencePayload(record.outputPath, record.payload)
		if err != nil {
			for _, stagedPath := range staged[:index] {
				_ = os.Remove(stagedPath)
			}
			return err
		}
		staged[index] = path
	}
	defer func() {
		for _, stagedPath := range staged {
			_ = os.Remove(stagedPath)
		}
	}()
	if beforePublish != nil {
		if err := beforePublish(); err != nil {
			return err
		}
	}
	for index, record := range records {
		if err := os.Link(staged[index], record.outputPath); err != nil {
			return fmt.Errorf("publish proof-bearing evidence %q without replacement: %w", filepath.ToSlash(record.outputPath), err)
		}
		if err := syncEvidenceDirectory(filepath.Dir(record.outputPath)); err != nil {
			return err
		}
	}
	return nil
}

func stageEvidencePayload(outputPath string, payload []byte) (string, error) {
	file, err := os.CreateTemp(filepath.Dir(outputPath), "."+filepath.Base(outputPath)+".tmp-*")
	if err != nil {
		return "", fmt.Errorf("stage proof-bearing evidence: %w", err)
	}
	path := file.Name()
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return "", fmt.Errorf("protect staged proof-bearing evidence: %w", err)
	}
	written, err := file.Write(payload)
	if err == nil && written != len(payload) {
		err = io.ErrShortWrite
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("write staged proof-bearing evidence: %w", err)
	}
	return path, nil
}

func syncEvidenceDirectory(directory string) error {
	file, err := os.Open(directory)
	if err != nil {
		return fmt.Errorf("open proof evidence directory: %w", err)
	}
	err = file.Sync()
	closeErr := file.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("sync proof evidence directory: %w", err)
	}
	return nil
}

func acceptedEvidenceOutputPath(repoRoot, path string) (string, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repository root for proof evidence: %w", err)
	}
	dir := filepath.Join(root, acceptedEvidenceDirectory)
	output, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve proof evidence path: %w", err)
	}
	recordName := strings.TrimSuffix(filepath.Base(output), ".json")
	if filepath.Dir(output) != dir || filepath.Ext(output) != ".json" || !isSafeProofIdentifier(recordName) {
		return "", fmt.Errorf("proof evidence must be a new JSON record directly under %q", filepath.ToSlash(dir))
	}
	return output, nil
}

// repositoryFingerprintSalt loads or creates the installation-local HMAC key.
// It deliberately lives beside the generated certification artifacts but is
// excluded from git, so the committed proof contains only HMAC output.  The
// credential itself never reaches this function or this file.
func repositoryFingerprintSalt(repoRoot string) ([]byte, error) {
	path := filepath.Join(repoRoot, repositoryFingerprintSaltPath)
	current, err := os.ReadFile(path)
	if err == nil {
		if len(current) < 16 {
			return nil, fmt.Errorf("repository fingerprint salt %q is too short", filepath.ToSlash(path))
		}
		return append([]byte(nil), current...), nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("read repository fingerprint salt %q: %w", filepath.ToSlash(path), err)
	}

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate repository fingerprint salt: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create repository fingerprint salt directory: %w", err)
	}
	// O_EXCL makes a concurrent first run retry with the winner's value rather
	// than silently replacing it and invalidating existing local fingerprints.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return repositoryFingerprintSalt(repoRoot)
	}
	if err != nil {
		return nil, fmt.Errorf("create repository fingerprint salt %q: %w", filepath.ToSlash(path), err)
	}
	if _, err := file.Write(salt); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("write repository fingerprint salt %q: %w", filepath.ToSlash(path), err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close repository fingerprint salt %q: %w", filepath.ToSlash(path), err)
	}
	return salt, nil
}

func sanitizeHTTPExchange(exchange completedHTTPExchange, preparedValues []string, repositorySalt []byte) (certifiedHTTPExchange, error) {
	operation := strings.TrimSpace(exchange.Operation)
	if !isSafeProofIdentifier(operation) {
		return certifiedHTTPExchange{}, errors.New("HTTP exchange operation must be a safe identifier")
	}
	request, err := sanitizeHTTPRequest(exchange.Request, preparedValues, repositorySalt)
	if err != nil {
		return certifiedHTTPExchange{}, fmt.Errorf("sanitize request: %w", err)
	}
	response, err := sanitizeHTTPResponse(exchange.Response, preparedValues, repositorySalt)
	if err != nil {
		return certifiedHTTPExchange{}, fmt.Errorf("sanitize response: %w", err)
	}
	return certifiedHTTPExchange{Operation: operation, Request: request, Response: response}, nil
}

func sanitizeHTTPRequest(request completedHTTPRequest, preparedValues []string, repositorySalt []byte) (certifiedHTTPRequest, error) {
	method := strings.ToUpper(strings.TrimSpace(request.Method))
	if !isSafeHTTPMethod(method) {
		return certifiedHTTPRequest{}, errors.New("HTTP request method is invalid")
	}
	parsed, err := url.Parse(request.Target)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return certifiedHTTPRequest{}, errors.New("HTTP request target is invalid")
	}
	if parsed.User != nil {
		return certifiedHTTPRequest{}, errors.New("HTTP request target must not contain userinfo")
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	query, err := sanitizeQuery(request.Target, preparedValues, repositorySalt)
	if err != nil {
		return certifiedHTTPRequest{}, err
	}
	body, err := sanitizeHTTPBody(request.Body, preparedValues, repositorySalt)
	if err != nil {
		return certifiedHTTPRequest{}, err
	}
	return certifiedHTTPRequest{
		Method:  method,
		Target:  fingerprintText(parsed.String(), preparedValues, repositorySalt),
		Query:   query,
		Headers: sanitizeHTTPFields(request.Headers, preparedValues, repositorySalt),
		Body:    body,
	}, nil
}

func sanitizeHTTPResponse(response completedHTTPResponse, preparedValues []string, repositorySalt []byte) (certifiedHTTPResponse, error) {
	if response.Status < 100 || response.Status > 599 {
		return certifiedHTTPResponse{}, errors.New("HTTP response status is invalid")
	}
	body, err := sanitizeHTTPBody(response.Body, preparedValues, repositorySalt)
	if err != nil {
		return certifiedHTTPResponse{}, err
	}
	return certifiedHTTPResponse{
		Status:  response.Status,
		Headers: sanitizeHTTPFields(response.Headers, preparedValues, repositorySalt),
		Body:    body,
	}, nil
}

func sanitizeDatabaseExchange(exchange completedDatabaseExchange, preparedValues []string, repositorySalt []byte) (certifiedDatabaseExchange, error) {
	operation := strings.TrimSpace(exchange.Operation)
	if !isSafeProofIdentifier(operation) {
		return certifiedDatabaseExchange{}, errors.New("database exchange operation must be a safe identifier")
	}
	protocol := strings.ToLower(strings.TrimSpace(exchange.Protocol))
	if !isSafeProofIdentifier(protocol) {
		return certifiedDatabaseExchange{}, errors.New("database exchange protocol must be a safe identifier")
	}
	status := strings.ToLower(strings.TrimSpace(exchange.ResponseStatus))
	if !isSafeProofIdentifier(status) {
		return certifiedDatabaseExchange{}, errors.New("database exchange response status must be a safe identifier")
	}
	body, err := sanitizeHTTPBody(exchange.ResponseBody, preparedValues, repositorySalt)
	if err != nil {
		return certifiedDatabaseExchange{}, fmt.Errorf("sanitize database response body: %w", err)
	}
	parameters := make([]string, 0, len(exchange.Parameters))
	for _, parameter := range exchange.Parameters {
		parameters = append(parameters, fingerprintText(parameter, preparedValues, repositorySalt))
	}
	return certifiedDatabaseExchange{
		Operation: operation,
		Protocol:  protocol,
		Request: certifiedDatabaseRequest{
			Statement:  fingerprintText(exchange.Statement, preparedValues, repositorySalt),
			Parameters: parameters,
		},
		Response: certifiedDatabaseResponse{Status: status, Body: body},
	}, nil
}

func sanitizeFlowRoundTrip(flow completedFlowRoundTrip, preparedValues []string, repositorySalt []byte) (certifiedFlowRoundTrip, error) {
	if strings.TrimSpace(flow.PMCommand) == "" {
		return certifiedFlowRoundTrip{}, errors.New("flow proof pm command is required")
	}
	if !isSafeProofIdentifier(flow.WarehouseReadbackOperation) {
		return certifiedFlowRoundTrip{}, errors.New("flow proof warehouse readback operation is invalid")
	}
	if !isSafeProofIdentifier(flow.DestinationReadbackOperation) {
		return certifiedFlowRoundTrip{}, errors.New("flow proof destination readback operation is invalid")
	}
	if flow.WarehouseReadbackOperation == flow.DestinationReadbackOperation {
		return certifiedFlowRoundTrip{}, errors.New("flow proof readbacks must use independent operations")
	}
	return certifiedFlowRoundTrip{
		PMCommandFingerprint:         fingerprintText(flow.PMCommand, preparedValues, repositorySalt),
		Mediator:                     localWarehouseMediator,
		WarehouseReadbackOperation:   flow.WarehouseReadbackOperation,
		DestinationReadbackOperation: flow.DestinationReadbackOperation,
		Delivery:                     flow.Delivery,
	}, nil
}

func sanitizeQuery(target string, preparedValues []string, repositorySalt []byte) ([]certifiedQuery, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return nil, errors.New("HTTP request query is invalid")
	}
	values, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return nil, errors.New("HTTP request query is invalid")
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		if !isSafeProofFieldName(key) {
			return nil, errors.New("HTTP request query name is invalid")
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	query := make([]certifiedQuery, 0)
	for _, key := range keys {
		for _, value := range values[key] {
			query = append(query, certifiedQuery{Name: key, Value: fingerprintText(value, preparedValues, repositorySalt)})
		}
	}
	return query, nil
}

func sanitizeHTTPFields(fields map[string][]string, preparedValues []string, repositorySalt []byte) []certifiedHTTPField {
	out := make([]certifiedHTTPField, 0)
	for originalName, values := range fields {
		name := strings.ToLower(strings.TrimSpace(originalName))
		if !isSafeProofFieldName(name) {
			// A malformed header name cannot be treated as a structural label;
			// retain only its deterministic fingerprint.
			name = fingerprintText(name, preparedValues, repositorySalt)
		}
		for _, value := range values {
			out = append(out, certifiedHTTPField{Name: name, Value: fingerprintText(value, preparedValues, repositorySalt)})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name == out[j].Name {
			return out[i].Value < out[j].Value
		}
		return out[i].Name < out[j].Name
	})
	return out
}

func sanitizeHTTPBody(body []byte, preparedValues []string, repositorySalt []byte) (certifiedHTTPBody, error) {
	originalBytes := len(body)
	truncated := false
	if len(body) > maxProofBodyBytes {
		body = body[:maxProofBodyBytes]
		truncated = true
	}
	if len(body) == 0 {
		return certifiedHTTPBody{Encoding: "none", Value: json.RawMessage("null"), OriginalBytes: originalBytes, Truncated: truncated}, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err == nil && ensureJSONEOF(decoder) == nil {
		sanitized := fingerprintJSONValue(value, preparedValues, repositorySalt)
		raw, err := json.Marshal(sanitized)
		if err != nil {
			return certifiedHTTPBody{}, errors.New("render sanitized JSON body")
		}
		return certifiedHTTPBody{Encoding: "json", Value: raw, OriginalBytes: originalBytes, Truncated: truncated}, nil
	}

	fingerprinted, err := json.Marshal(fingerprintText(string(body), preparedValues, repositorySalt))
	if err != nil {
		return certifiedHTTPBody{}, errors.New("render sanitized opaque body")
	}
	return certifiedHTTPBody{Encoding: "opaque", Value: fingerprinted, OriginalBytes: originalBytes, Truncated: truncated}, nil
}

func fingerprintJSONValue(value any, preparedValues []string, repositorySalt []byte) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			// JSON property names are structural labels. If a prepared value was
			// incorrectly used as one, it is not safe and is fingerprinted too.
			if containsPreparedValue(key, preparedValues) || !isSafeProofFieldName(key) {
				key = fingerprintText(key, preparedValues, repositorySalt)
			}
			out[key] = fingerprintJSONValue(child, preparedValues, repositorySalt)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			out[i] = fingerprintJSONValue(child, preparedValues, repositorySalt)
		}
		return out
	case nil:
		return nil
	case string:
		return fingerprintText(typed, preparedValues, repositorySalt)
	case json.Number:
		return fingerprintText(typed.String(), preparedValues, repositorySalt)
	case bool:
		return fingerprintText(fmt.Sprintf("%t", typed), preparedValues, repositorySalt)
	case float64:
		return fingerprintText(fmt.Sprintf("%v", typed), preparedValues, repositorySalt)
	default:
		return fingerprintText(fmt.Sprint(typed), preparedValues, repositorySalt)
	}
}

// fingerprintText first identifies exact prepared values, then fingerprints
// every remaining unproved fragment. Therefore a credential embedded inside a
// header, URL, request body, or response body receives its own stable marker,
// while surrounding unknown text cannot leak as a supposedly harmless value.
func fingerprintText(value string, preparedValues []string, repositorySalt []byte) string {
	values := normalizedPreparedValues(preparedValues)
	if value == "" {
		return fingerprintValue(repositorySalt, value)
	}
	var out strings.Builder
	remaining := value
	for len(remaining) > 0 {
		index, matched := nextPreparedValue(remaining, values)
		if index < 0 {
			out.WriteString(fingerprintValue(repositorySalt, remaining))
			break
		}
		if index > 0 {
			out.WriteString(fingerprintValue(repositorySalt, remaining[:index]))
		}
		out.WriteString(fingerprintValue(repositorySalt, matched))
		remaining = remaining[index+len(matched):]
	}
	return out.String()
}

func fingerprintPreparedValues(values []string, repositorySalt []byte) []string {
	prepared := normalizedPreparedValues(values)
	fingerprints := make([]string, 0, len(prepared))
	for _, value := range prepared {
		fingerprints = append(fingerprints, fingerprintValue(repositorySalt, value))
	}
	sort.Strings(fingerprints)
	return fingerprints
}

func normalizedPreparedValues(values []string) []string {
	unique := make(map[string]bool, len(values))
	for _, value := range values {
		if value != "" {
			unique[value] = true
		}
	}
	out := make([]string, 0, len(unique))
	for value := range unique {
		out = append(out, value)
	}
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) == len(out[j]) {
			return out[i] < out[j]
		}
		return len(out[i]) > len(out[j])
	})
	return out
}

func nextPreparedValue(value string, preparedValues []string) (int, string) {
	bestIndex := -1
	bestValue := ""
	for _, prepared := range preparedValues {
		index := strings.Index(value, prepared)
		if index < 0 {
			continue
		}
		if bestIndex < 0 || index < bestIndex || (index == bestIndex && len(prepared) > len(bestValue)) {
			bestIndex = index
			bestValue = prepared
		}
	}
	return bestIndex, bestValue
}

func containsPreparedValue(value string, preparedValues []string) bool {
	for _, prepared := range preparedValues {
		if prepared != "" && strings.Contains(value, prepared) {
			return true
		}
	}
	return false
}

func fingerprintValue(repositorySalt []byte, value string) string {
	hash := hmac.New(sha256.New, repositorySalt)
	_, _ = hash.Write([]byte(value))
	return fingerprintPrefix + hex.EncodeToString(hash.Sum(nil)) + fingerprintSuffix
}

func validateEmbeddedEvidenceProof(proof embeddedEvidenceProof) error {
	if proof.RedactionStrategy != proofRedactionStrategy {
		return fmt.Errorf("redaction_strategy %q is unsupported", proof.RedactionStrategy)
	}
	if !isSHA256(proof.PMBinarySHA256) {
		return errors.New("pm_binary_sha256 must be a lowercase SHA-256 digest")
	}
	if !isFingerprintSequence(proof.PMCommandFingerprint) {
		return errors.New("pm_command_fingerprint must contain only fingerprints")
	}
	if len(proof.CredentialFingerprints) == 0 || !isSortedUniqueStrings(proof.CredentialFingerprints) {
		return errors.New("credential_fingerprints must be a non-empty sorted unique list")
	}
	for index, fingerprint := range proof.CredentialFingerprints {
		if !isFingerprintSequence(fingerprint) {
			return fmt.Errorf("credential_fingerprints[%d] must contain only fingerprints", index)
		}
	}
	if len(proof.HTTPExchanges)+len(proof.DatabaseExchanges) == 0 {
		return errors.New("at least one protocol exchange is required")
	}
	operations := make(map[string]bool, len(proof.HTTPExchanges)+len(proof.DatabaseExchanges))
	for i, exchange := range proof.HTTPExchanges {
		if !isSafeProofIdentifier(exchange.Operation) {
			return fmt.Errorf("http_exchanges[%d].operation is invalid", i)
		}
		if operations[exchange.Operation] {
			return fmt.Errorf("protocol exchange operation %q is duplicated", exchange.Operation)
		}
		operations[exchange.Operation] = true
		if !isSafeHTTPMethod(exchange.Request.Method) {
			return fmt.Errorf("http_exchanges[%d].request.method is invalid", i)
		}
		if !isFingerprintSequence(exchange.Request.Target) {
			return fmt.Errorf("http_exchanges[%d].request.target must contain only fingerprints", i)
		}
		if err := validateCertifiedQuery(exchange.Request.Query); err != nil {
			return fmt.Errorf("http_exchanges[%d].request.query: %w", i, err)
		}
		if err := validateCertifiedHTTPFields(exchange.Request.Headers); err != nil {
			return fmt.Errorf("http_exchanges[%d].request.headers: %w", i, err)
		}
		if err := validateCertifiedHTTPBody(exchange.Request.Body); err != nil {
			return fmt.Errorf("http_exchanges[%d].request.body: %w", i, err)
		}
		if exchange.Response.Status < 100 || exchange.Response.Status > 599 {
			return fmt.Errorf("http_exchanges[%d].response.status is invalid", i)
		}
		if err := validateCertifiedHTTPFields(exchange.Response.Headers); err != nil {
			return fmt.Errorf("http_exchanges[%d].response.headers: %w", i, err)
		}
		if err := validateCertifiedHTTPBody(exchange.Response.Body); err != nil {
			return fmt.Errorf("http_exchanges[%d].response.body: %w", i, err)
		}
	}
	for i, exchange := range proof.DatabaseExchanges {
		if !isSafeProofIdentifier(exchange.Operation) {
			return fmt.Errorf("database_exchanges[%d].operation is invalid", i)
		}
		if operations[exchange.Operation] {
			return fmt.Errorf("protocol exchange operation %q is duplicated", exchange.Operation)
		}
		operations[exchange.Operation] = true
		if !isSafeProofIdentifier(exchange.Protocol) {
			return fmt.Errorf("database_exchanges[%d].protocol is invalid", i)
		}
		if !isFingerprintSequence(exchange.Request.Statement) {
			return fmt.Errorf("database_exchanges[%d].request.statement must contain only fingerprints", i)
		}
		for parameterIndex, parameter := range exchange.Request.Parameters {
			if !isFingerprintSequence(parameter) {
				return fmt.Errorf("database_exchanges[%d].request.parameters[%d] must contain only fingerprints", i, parameterIndex)
			}
		}
		if !isSafeProofIdentifier(exchange.Response.Status) {
			return fmt.Errorf("database_exchanges[%d].response.status is invalid", i)
		}
		if err := validateCertifiedHTTPBody(exchange.Response.Body); err != nil {
			return fmt.Errorf("database_exchanges[%d].response.body: %w", i, err)
		}
	}
	if proof.Flow != nil {
		if err := validateCertifiedFlowRoundTrip(*proof.Flow, operations); err != nil {
			return fmt.Errorf("flow: %w", err)
		}
	}
	return nil
}

func isSortedUniqueStrings(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for index := 1; index < len(values); index++ {
		if values[index-1] >= values[index] {
			return false
		}
	}
	return true
}

func validateCertifiedFlowRoundTrip(flow certifiedFlowRoundTrip, operations map[string]bool) error {
	if !isFingerprintSequence(flow.PMCommandFingerprint) {
		return errors.New("pm_command_fingerprint must contain only fingerprints")
	}
	if flow.Mediator != localWarehouseMediator {
		return fmt.Errorf("mediator %q is unsupported", flow.Mediator)
	}
	if !operations[flow.WarehouseReadbackOperation] {
		return errors.New("warehouse_readback_operation must name an embedded exchange")
	}
	if !operations[flow.DestinationReadbackOperation] {
		return errors.New("destination_readback_operation must name an embedded exchange")
	}
	if flow.WarehouseReadbackOperation == flow.DestinationReadbackOperation {
		return errors.New("warehouse and destination readbacks must be independent")
	}
	return validateDeliveryGuarantees(flow.Delivery)
}

func validateDeliveryGuarantees(delivery deliveryGuarantees) error {
	falseGuarantees := map[string]bool{
		"resumable":                !delivery.Resumable,
		"receipt_backed":           !delivery.ReceiptBacked,
		"checkpointed":             !delivery.Checkpointed,
		"replay_identity":          !delivery.ReplayIdentity,
		"provider_idempotency_key": !delivery.ProviderIdempotencyKey,
	}
	covered := make(map[string]bool, len(delivery.Limitations))
	for _, limitation := range delivery.Limitations {
		if !falseGuarantees[limitation.Guarantee] {
			return fmt.Errorf("limitation %q does not correspond to a false guarantee", limitation.Guarantee)
		}
		if covered[limitation.Guarantee] {
			return fmt.Errorf("limitation %q is duplicated", limitation.Guarantee)
		}
		if err := validateNotApplicableReason(notApplicableReason{Code: limitation.Code, Reason: limitation.Reason}); err != nil {
			return fmt.Errorf("limitation %q: %w", limitation.Guarantee, err)
		}
		covered[limitation.Guarantee] = true
	}
	for guarantee, isFalse := range falseGuarantees {
		if isFalse && !covered[guarantee] {
			return fmt.Errorf("false guarantee %q requires a named limitation", guarantee)
		}
	}
	return nil
}

func validateCertifiedQuery(query []certifiedQuery) error {
	for _, item := range query {
		if !isSafeProofFieldName(item.Name) && !isFingerprintSequence(item.Name) {
			return errors.New("query name is invalid")
		}
		if !isFingerprintSequence(item.Value) {
			return errors.New("query value must contain only fingerprints")
		}
	}
	return nil
}

func validateCertifiedHTTPFields(fields []certifiedHTTPField) error {
	for _, field := range fields {
		if !isSafeProofFieldName(field.Name) && !isFingerprintSequence(field.Name) {
			return errors.New("header name is invalid")
		}
		if !isFingerprintSequence(field.Value) {
			return errors.New("header value must contain only fingerprints")
		}
	}
	return nil
}

func validateCertifiedHTTPBody(body certifiedHTTPBody) error {
	if body.OriginalBytes < 0 {
		return errors.New("original_bytes is invalid")
	}
	switch body.Encoding {
	case "none":
		if string(body.Value) != "null" || body.OriginalBytes != 0 || body.Truncated {
			return errors.New("none body must be null with zero bytes")
		}
		return nil
	case "opaque":
		var value string
		if err := decodeStrictJSON(body.Value, &value); err != nil || !isFingerprintSequence(value) {
			return errors.New("opaque body must contain only fingerprints")
		}
		return nil
	case "json":
		decoder := json.NewDecoder(bytes.NewReader(body.Value))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil {
			return errors.New("json body is invalid")
		}
		if err := ensureJSONEOF(decoder); err != nil {
			return errors.New("json body is invalid")
		}
		if !isSanitizedJSONValue(value) {
			return errors.New("json body contains an unproved value")
		}
		return nil
	default:
		return fmt.Errorf("encoding %q is unsupported", body.Encoding)
	}
}

func isSanitizedJSONValue(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if !isSafeProofFieldName(key) && !isFingerprintSequence(key) {
				return false
			}
			if !isSanitizedJSONValue(child) {
				return false
			}
		}
		return true
	case []any:
		for _, child := range typed {
			if !isSanitizedJSONValue(child) {
				return false
			}
		}
		return true
	case nil:
		return true
	case string:
		return isFingerprintSequence(typed)
	default:
		return false
	}
}

func validateEvidenceProvider(provider string) error {
	if !isSafeProofIdentifier(provider) {
		return errors.New("must be a safe provider identifier")
	}
	return nil
}

func validateEvidenceRunID(runID string) error {
	if !isSafeProofIdentifier(runID) {
		return errors.New("must be a safe run identifier")
	}
	return nil
}

func isSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func isFingerprintSequence(value string) bool {
	if value == "" {
		return false
	}
	for value != "" {
		if !strings.HasPrefix(value, fingerprintPrefix) {
			return false
		}
		end := strings.Index(value[len(fingerprintPrefix):], fingerprintSuffix)
		if end < 0 {
			return false
		}
		digest := value[len(fingerprintPrefix) : len(fingerprintPrefix)+end]
		if !isSHA256(digest) {
			return false
		}
		value = value[len(fingerprintPrefix)+end+len(fingerprintSuffix):]
	}
	return true
}

func isSafeHTTPMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		return true
	default:
		return false
	}
}

func isSafeProofIdentifier(value string) bool {
	if value == "" || len(value) > 200 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '.' || char == '_' || char == '-' || char == ':' {
			continue
		}
		return false
	}
	return true
}

func isSafeProofFieldName(value string) bool {
	if value == "" || len(value) > 200 {
		return false
	}
	for _, char := range value {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' || char == '.' {
			continue
		}
		return false
	}
	return true
}

// decodeStrictJSON is used for evidence and generated artifact parsing so an
// unrecognized field can never hide a raw provider response or credential.
func decodeStrictJSON(raw []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return ensureJSONEOF(decoder)
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return errors.New("multiple JSON values")
	}
	return err
}
