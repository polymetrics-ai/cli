package certify

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	externalProofVersion               = 2
	externalProofFingerprintID         = "repository_salted_hmac_sha256_v1"
	externalProofMarkerPrefix          = "{{pmcertfp:v1:"
	externalProofMarkerSuffix          = "}}"
	externalProofMaxExchangesPerTarget = 16
	externalProofMaxTotalExchanges     = 4096

	externalProofCredentialScopeFullParity         = "full_parity"
	externalProofCredentialScopeObservedOperations = "observed_operations"
	externalProofScopeProofFullParityStage         = "full_parity_stage"
	externalProofScopeProofProtocolExchanges       = "protocol_exchanges"
)

var externalProofIdentifier = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
var externalProofFieldName = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)
var externalProofSHA256 = regexp.MustCompile(`^[a-f0-9]{64}$`)

// ExternalProofInput is the process-memory-only boundary between a completed
// external binary run and its serialized proof. PreparedValues and exchanges
// contain raw values only long enough to generate HMAC fingerprints; neither
// field has a JSON representation or a persistence path.
type ExternalProofInput struct {
	Connector               string
	RunID                   string
	BinarySHA256            string
	Command                 []string
	Stdout                  string
	Stderr                  string
	ExitCode                int
	Passed                  bool
	FullParity              bool
	PreparedValues          []string
	HTTPExchanges           []ObservedHTTPExchange
	FlowRoundTripReferences []string
}

// externalProofArtifact is intentionally private so callers cannot construct
// a serializable artifact containing raw credential or transcript values.
type externalProofArtifact struct {
	Version                 int                         `json:"version"`
	RedactionStrategy       string                      `json:"redaction_strategy"`
	Connector               string                      `json:"connector"`
	RunID                   string                      `json:"run_id"`
	PMBinarySHA256          string                      `json:"pm_binary_sha256"`
	CredentialScope         string                      `json:"credential_scope,omitempty"`
	CredentialScopeProof    string                      `json:"credential_scope_proof,omitempty"`
	CredentialFingerprints  []string                    `json:"credential_fingerprints"`
	Process                 externalProofProcess        `json:"process"`
	HTTPExchanges           []externalProofHTTPExchange `json:"http_exchanges"`
	FlowRoundTripReferences []string                    `json:"flow_round_trip_references"`
}

type externalProofProcess struct {
	Command           []string `json:"command"`
	ExitCode          int      `json:"exit_code"`
	StdoutFingerprint string   `json:"stdout_fingerprint"`
	StderrFingerprint string   `json:"stderr_fingerprint"`
}

type externalProofHTTPExchange struct {
	Request  externalProofHTTPRequest  `json:"request"`
	Response externalProofHTTPResponse `json:"response"`
}

type externalProofHTTPRequest struct {
	Method  string                    `json:"method"`
	Target  string                    `json:"target"`
	Query   []externalProofField      `json:"query"`
	Headers []externalProofField      `json:"headers"`
	Body    externalProofObservedBody `json:"body"`
}

type externalProofHTTPResponse struct {
	Status  int                       `json:"status"`
	Headers []externalProofField      `json:"headers"`
	Body    externalProofObservedBody `json:"body"`
}

type externalProofField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type externalProofObservedBody struct {
	Encoding      string          `json:"encoding"`
	Value         json.RawMessage `json:"value"`
	OriginalBytes int             `json:"original_bytes"`
}

// ImportedExternalProof is the validated, redacted projection an evidence
// publisher may consume. It intentionally omits raw process argv, stdout, and
// stderr: those values never need to cross the certification-package boundary
// after their fingerprints have been verified.
type ImportedExternalProof struct {
	Connector               string
	RunID                   string
	PMBinarySHA256          string
	PMCommandFingerprint    string
	CredentialFingerprints  []string
	HTTPExchanges           []ImportedExternalHTTPExchange
	FlowRoundTripReferences []string
}

// ImportedExternalHTTPExchange is already redacted. Every query/header/body
// value has passed the stored-proof fingerprint validation in ReadExternalProof.
type ImportedExternalHTTPExchange struct {
	Request  ImportedExternalHTTPRequest
	Response ImportedExternalHTTPResponse
}

type ImportedExternalHTTPRequest struct {
	Method  string
	Target  string
	Query   []ImportedExternalProofField
	Headers []ImportedExternalProofField
	Body    ImportedExternalProofBody
}

type ImportedExternalHTTPResponse struct {
	Status  int
	Headers []ImportedExternalProofField
	Body    ImportedExternalProofBody
}

type ImportedExternalProofField struct {
	Name  string
	Value string
}

type ImportedExternalProofBody struct {
	Encoding      string
	Value         json.RawMessage
	OriginalBytes int
}

// WriteExternalProof creates a schema-v2 artifact after all bounded-capture
// and sanitization checks pass. Its credential scope is derived here, not
// supplied by a caller: a verified full-parity result retains that claim;
// every other completed run can claim only the successful protocol exchanges
// it actually observed. A rejected input performs no artifact or salt writes.
func WriteExternalProof(root string, input ExternalProofInput) (string, error) {
	prepared, scope, err := validateExternalProofInput(input)
	if err != nil {
		return "", err
	}

	salt, err := externalProofSalt(root)
	if err != nil {
		return "", err
	}
	artifact, err := sanitizeExternalProof(input, prepared, salt, scope)
	if err != nil {
		return "", err
	}
	payload, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return "", fmt.Errorf("render external proof: %w", err)
	}
	if hits := ScanForSecrets(string(payload), prepared); len(hits) != 0 {
		return "", errors.New("sanitized external proof retained credential material")
	}
	outputPath, err := externalProofOutputPath(root, input.Connector, input.RunID)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o700); err != nil {
		return "", fmt.Errorf("create external proof directory: %w", err)
	}
	file, err := os.OpenFile(outputPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("create external proof: %w", err)
	}
	if _, err := file.Write(append(payload, '\n')); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("write external proof: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close external proof: %w", err)
	}
	return outputPath, nil
}

// VerifyExternalProofTranscript proves that stdout and stderr are exactly the
// bytes observed from the external process, without placing either stream into
// a proof artifact. It is useful for smoke tests and post-run acceptance.
func VerifyExternalProofTranscript(root, proofPath, stdout, stderr string) (bool, error) {
	salt, err := readExternalProofSalt(root)
	if err != nil {
		return false, err
	}
	raw, err := os.ReadFile(proofPath)
	if err != nil {
		return false, fmt.Errorf("read external proof: %w", err)
	}
	var artifact externalProofArtifact
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return false, fmt.Errorf("parse external proof: %w", err)
	}
	if (artifact.Version != 1 && artifact.Version != externalProofVersion) || artifact.RedactionStrategy != externalProofFingerprintID {
		return false, errors.New("external proof has an unsupported schema")
	}
	if artifact.Version == externalProofVersion && !validExternalProofCredentialScope(artifact.CredentialScope, artifact.CredentialScopeProof) {
		return false, errors.New("external proof has an unsupported credential scope claim")
	}
	stdoutFingerprint := externalProofFingerprint(salt, stdout)
	stderrFingerprint := externalProofFingerprint(salt, stderr)
	return subtle.ConstantTimeCompare([]byte(artifact.Process.StdoutFingerprint), []byte(stdoutFingerprint)) == 1 &&
		subtle.ConstantTimeCompare([]byte(artifact.Process.StderrFingerprint), []byte(stderrFingerprint)) == 1, nil
}

// ReadExternalProof loads a proof only after checking that all serialized
// transcript values are fingerprints. It is the narrow bridge from an
// ephemeral, sanitized external-proof artifact to a committed accepted
// evidence record; callers cannot obtain the captured command or process
// streams from it.
func ReadExternalProof(root, proofPath string) (ImportedExternalProof, error) {
	salt, err := readExternalProofSalt(root)
	if err != nil {
		return ImportedExternalProof{}, err
	}
	raw, err := os.ReadFile(proofPath)
	if err != nil {
		return ImportedExternalProof{}, fmt.Errorf("read external proof: %w", err)
	}
	var artifact externalProofArtifact
	if err := decodeExternalProofArtifact(raw, &artifact); err != nil {
		return ImportedExternalProof{}, fmt.Errorf("parse external proof: %w", err)
	}
	if err := validateStoredExternalProof(artifact); err != nil {
		return ImportedExternalProof{}, fmt.Errorf("validate external proof: %w", err)
	}
	proof := ImportedExternalProof{
		Connector:               artifact.Connector,
		RunID:                   artifact.RunID,
		PMBinarySHA256:          artifact.PMBinarySHA256,
		PMCommandFingerprint:    externalProofFingerprint(salt, strings.Join(artifact.Process.Command, "\x00")),
		CredentialFingerprints:  append([]string(nil), artifact.CredentialFingerprints...),
		HTTPExchanges:           make([]ImportedExternalHTTPExchange, 0, len(artifact.HTTPExchanges)),
		FlowRoundTripReferences: append([]string(nil), artifact.FlowRoundTripReferences...),
	}
	for _, exchange := range artifact.HTTPExchanges {
		proof.HTTPExchanges = append(proof.HTTPExchanges, importedExternalHTTPExchange(exchange))
	}
	return proof, nil
}

func importedExternalHTTPExchange(exchange externalProofHTTPExchange) ImportedExternalHTTPExchange {
	return ImportedExternalHTTPExchange{
		Request: ImportedExternalHTTPRequest{
			Method: exchange.Request.Method, Target: exchange.Request.Target,
			Query: importedExternalProofFields(exchange.Request.Query), Headers: importedExternalProofFields(exchange.Request.Headers),
			Body: importedExternalProofBody(exchange.Request.Body),
		},
		Response: ImportedExternalHTTPResponse{
			Status: exchange.Response.Status, Headers: importedExternalProofFields(exchange.Response.Headers),
			Body: importedExternalProofBody(exchange.Response.Body),
		},
	}
}

func importedExternalProofFields(fields []externalProofField) []ImportedExternalProofField {
	result := make([]ImportedExternalProofField, len(fields))
	for i, field := range fields {
		result[i] = ImportedExternalProofField(field)
	}
	return result
}

func importedExternalProofBody(body externalProofObservedBody) ImportedExternalProofBody {
	return ImportedExternalProofBody{Encoding: body.Encoding, Value: append(json.RawMessage(nil), body.Value...), OriginalBytes: body.OriginalBytes}
}

// RedactExternalProofDiagnostic preserves a readable diagnostic while
// replacing every detected representation of a prepared credential with the
// external-proof HMAC marker. Its salt is random, in-memory, and per-call:
// diagnostics never create a proof, a root salt, or another persisted
// credential-bearing artifact.
func RedactExternalProofDiagnostic(text string, secretValues []string) (string, error) {
	prepared := normalizedExternalProofValues(secretValues)
	if len(prepared) == 0 || text == "" {
		return text, nil
	}
	salt := make([]byte, sha256.Size)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate external-proof diagnostic salt: %w", err)
	}
	for _, secret := range prepared {
		marker := externalProofFingerprint(salt, secret)
		for _, form := range externalProofDiagnosticSecretForms(secret) {
			text = strings.ReplaceAll(text, form, marker)
		}
	}
	if len(ScanForSecrets(text, prepared)) != 0 {
		return "", errors.New("external-proof diagnostic retains credential material")
	}
	return text, nil
}

func externalProofDiagnosticSecretForms(secret string) []string {
	forms := []string{
		secret,
		base64.StdEncoding.EncodeToString([]byte(secret)),
		base64.RawStdEncoding.EncodeToString([]byte(secret)),
		url.QueryEscape(secret),
	}
	sort.Slice(forms, func(left, right int) bool {
		return len(forms[left]) > len(forms[right])
	})
	return forms
}

type externalProofCredentialScope struct {
	credentialScope string
	proof           string
	fullParity      bool
}

func validateExternalProofInput(input ExternalProofInput) ([]string, externalProofCredentialScope, error) {
	if input.ExitCode < 0 {
		return nil, externalProofCredentialScope{}, errors.New("external proof requires a completed process exit code")
	}
	scope := externalProofCredentialScope{
		credentialScope: externalProofCredentialScopeObservedOperations,
		proof:           externalProofScopeProofProtocolExchanges,
	}
	if input.Passed && input.ExitCode == 0 && input.FullParity {
		scope = externalProofCredentialScope{
			credentialScope: externalProofCredentialScopeFullParity,
			proof:           externalProofScopeProofFullParityStage,
			fullParity:      true,
		}
	}
	if !externalProofIdentifier.MatchString(input.Connector) || !externalProofIdentifier.MatchString(input.RunID) {
		return nil, externalProofCredentialScope{}, errors.New("external proof connector and run id must be safe identifiers")
	}
	if !externalProofSHA256.MatchString(input.BinarySHA256) {
		return nil, externalProofCredentialScope{}, errors.New("external proof requires a lowercase SHA-256 built-binary fingerprint")
	}
	if len(input.Command) == 0 {
		return nil, externalProofCredentialScope{}, errors.New("external proof requires the exact observed process command")
	}
	prepared := normalizedExternalProofValues(input.PreparedValues)
	if len(prepared) == 0 {
		return nil, externalProofCredentialScope{}, errors.New("external proof requires prepared credential values")
	}
	if len(ScanForSecrets(strings.Join(input.Command, "\x00"), prepared)) != 0 {
		return nil, externalProofCredentialScope{}, errors.New("external proof refuses credential material in process arguments")
	}
	if len(ScanForSecrets(input.Stdout, prepared)) != 0 || len(ScanForSecrets(input.Stderr, prepared)) != 0 {
		return nil, externalProofCredentialScope{}, errors.New("external proof refuses credential material in observed process output")
	}
	if len(input.HTTPExchanges) == 0 {
		return nil, externalProofCredentialScope{}, errors.New("external proof requires at least one observed HTTPS exchange")
	}
	if len(input.HTTPExchanges) > externalProofMaxTotalExchanges {
		return nil, externalProofCredentialScope{}, fmt.Errorf("external proof observed %d exchanges, exceeding bounded whole-surface limit %d", len(input.HTTPExchanges), externalProofMaxTotalExchanges)
	}
	exchangesPerTarget := make(map[string]int)
	observedProviderSuccess := false
	for index, exchange := range input.HTTPExchanges {
		if err := validateObservedProofExchange(exchange); err != nil {
			return nil, externalProofCredentialScope{}, fmt.Errorf("external proof exchange %d: %w", index+1, err)
		}
		if exchange.Response.Status >= 200 && exchange.Response.Status < 300 {
			observedProviderSuccess = true
		}
		targetKey := strings.ToUpper(exchange.Request.Method) + "\x00" + exchange.Request.Target
		exchangesPerTarget[targetKey]++
		if exchangesPerTarget[targetKey] > externalProofMaxExchangesPerTarget {
			return nil, externalProofCredentialScope{}, fmt.Errorf("external proof observed more than %d exchanges for one request target, exceeding bounded redirect/retry limit", externalProofMaxExchangesPerTarget)
		}
	}
	if !observedProviderSuccess {
		return nil, externalProofCredentialScope{}, errors.New("external proof requires an observed successful provider response")
	}
	if scope.fullParity {
		if err := validateExternalProofFlowRoundTripReferences(input.FlowRoundTripReferences); err != nil {
			return nil, externalProofCredentialScope{}, err
		}
	} else if len(input.FlowRoundTripReferences) != 0 {
		return nil, externalProofCredentialScope{}, errors.New("observed-operations external proof must not include full-parity flow references")
	}
	return prepared, scope, nil
}

func validExternalProofCredentialScope(scope, proof string) bool {
	return (scope == externalProofCredentialScopeFullParity && proof == externalProofScopeProofFullParityStage) ||
		(scope == externalProofCredentialScopeObservedOperations && proof == externalProofScopeProofProtocolExchanges)
}

func decodeExternalProofArtifact(raw []byte, target *externalProofArtifact) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return ensureExternalProofJSONEOF(decoder)
}

func validateStoredExternalProof(artifact externalProofArtifact) error {
	if (artifact.Version != 1 && artifact.Version != externalProofVersion) || artifact.RedactionStrategy != externalProofFingerprintID {
		return errors.New("external proof has an unsupported schema")
	}
	if artifact.Version == externalProofVersion && !validExternalProofCredentialScope(artifact.CredentialScope, artifact.CredentialScopeProof) {
		return errors.New("external proof has an unsupported credential scope claim")
	}
	if !externalProofIdentifier.MatchString(artifact.Connector) || !externalProofIdentifier.MatchString(artifact.RunID) {
		return errors.New("external proof connector and run id must be safe identifiers")
	}
	if !externalProofSHA256.MatchString(artifact.PMBinarySHA256) {
		return errors.New("external proof requires a lowercase SHA-256 built-binary fingerprint")
	}
	if len(artifact.CredentialFingerprints) == 0 || !sortedUniqueExternalProofFingerprints(artifact.CredentialFingerprints) {
		return errors.New("external proof credential fingerprints must be a non-empty sorted unique list")
	}
	if len(artifact.Process.Command) == 0 || artifact.Process.ExitCode < 0 ||
		!isExternalProofFingerprintSequence(artifact.Process.StdoutFingerprint) ||
		!isExternalProofFingerprintSequence(artifact.Process.StderrFingerprint) {
		return errors.New("external proof process record is not a completed redacted process")
	}
	if len(artifact.HTTPExchanges) == 0 || len(artifact.HTTPExchanges) > externalProofMaxTotalExchanges {
		return errors.New("external proof must contain a bounded non-empty HTTPS exchange set")
	}
	for i, exchange := range artifact.HTTPExchanges {
		if err := validateStoredExternalProofExchange(exchange); err != nil {
			return fmt.Errorf("http_exchanges[%d]: %w", i, err)
		}
	}
	if artifact.Version == 1 {
		if artifact.Process.ExitCode != 0 {
			return errors.New("version-1 external proof process record is not successful")
		}
		if len(artifact.FlowRoundTripReferences) != 0 {
			return validateExternalProofFlowRoundTripReferences(artifact.FlowRoundTripReferences)
		}
		return nil
	}
	observedProviderSuccess := false
	for _, exchange := range artifact.HTTPExchanges {
		if exchange.Response.Status >= 200 && exchange.Response.Status < 300 {
			observedProviderSuccess = true
			break
		}
	}
	if !observedProviderSuccess {
		return errors.New("external proof requires an observed successful provider response")
	}
	if artifact.CredentialScope == externalProofCredentialScopeFullParity {
		if artifact.Process.ExitCode != 0 {
			return errors.New("full-parity external proof process record is not successful")
		}
		return validateExternalProofFlowRoundTripReferences(artifact.FlowRoundTripReferences)
	}
	if len(artifact.FlowRoundTripReferences) != 0 {
		return errors.New("observed-operations external proof must not include full-parity flow references")
	}
	return nil
}

func sortedUniqueExternalProofFingerprints(values []string) bool {
	for i, value := range values {
		if !isExternalProofFingerprintSequence(value) {
			return false
		}
		if i > 0 && values[i-1] >= value {
			return false
		}
	}
	return true
}

func validateStoredExternalProofExchange(exchange externalProofHTTPExchange) error {
	if !isExternalProofMethod(exchange.Request.Method) || !isExternalProofFingerprintSequence(exchange.Request.Target) {
		return errors.New("request method or target is invalid")
	}
	if exchange.Response.Status < 100 || exchange.Response.Status > 599 {
		return errors.New("response status is invalid")
	}
	for _, fields := range [][]externalProofField{exchange.Request.Query, exchange.Request.Headers, exchange.Response.Headers} {
		for _, field := range fields {
			if (!externalProofFieldName.MatchString(field.Name) && !isExternalProofFingerprintSequence(field.Name)) || !isExternalProofFingerprintSequence(field.Value) {
				return errors.New("field retains an unredacted name or value")
			}
		}
	}
	if err := validateStoredExternalProofBody(exchange.Request.Body); err != nil {
		return fmt.Errorf("request body: %w", err)
	}
	if err := validateStoredExternalProofBody(exchange.Response.Body); err != nil {
		return fmt.Errorf("response body: %w", err)
	}
	return nil
}

func validateStoredExternalProofBody(body externalProofObservedBody) error {
	if body.OriginalBytes < 0 {
		return errors.New("original_bytes is invalid")
	}
	switch body.Encoding {
	case "none":
		if string(body.Value) != "null" || body.OriginalBytes != 0 {
			return errors.New("none body must be null with zero bytes")
		}
		return nil
	case "opaque":
		var value string
		if err := json.Unmarshal(body.Value, &value); err != nil || !isExternalProofFingerprintSequence(value) {
			return errors.New("opaque body retains an unredacted value")
		}
		return nil
	case "json":
		decoder := json.NewDecoder(bytes.NewReader(body.Value))
		decoder.UseNumber()
		var value any
		if err := decoder.Decode(&value); err != nil || ensureExternalProofJSONEOF(decoder) != nil || !isStoredExternalProofJSON(value) {
			return errors.New("JSON body retains an unredacted value")
		}
		return nil
	default:
		return fmt.Errorf("encoding %q is unsupported", body.Encoding)
	}
}

func isStoredExternalProofJSON(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if (!externalProofFieldName.MatchString(key) && !isExternalProofFingerprintSequence(key)) || !isStoredExternalProofJSON(child) {
				return false
			}
		}
		return true
	case []any:
		for _, child := range typed {
			if !isStoredExternalProofJSON(child) {
				return false
			}
		}
		return true
	case nil:
		return true
	case string:
		return isExternalProofFingerprintSequence(typed)
	default:
		return false
	}
}

func isExternalProofFingerprintSequence(value string) bool {
	if value == "" {
		return false
	}
	for value != "" {
		if !strings.HasPrefix(value, externalProofMarkerPrefix) {
			return false
		}
		end := strings.Index(value[len(externalProofMarkerPrefix):], externalProofMarkerSuffix)
		if end < 0 {
			return false
		}
		digest := value[len(externalProofMarkerPrefix) : len(externalProofMarkerPrefix)+end]
		if !externalProofSHA256.MatchString(digest) {
			return false
		}
		value = value[len(externalProofMarkerPrefix)+end+len(externalProofMarkerSuffix):]
	}
	return true
}

func isExternalProofMethod(method string) bool {
	switch method {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS":
		return true
	default:
		return false
	}
}

func validateExternalProofFlowRoundTripReferences(references []string) error {
	required := []string{"flow_plan", "flow_preview", "flow_run", "flow_status"}
	present := make(map[string]bool, len(required))
	for _, reference := range references {
		if !externalProofIdentifier.MatchString(reference) {
			return fmt.Errorf("external proof has unsafe flow round-trip reference %q", reference)
		}
		for _, requiredReference := range required {
			if reference == requiredReference {
				present[reference] = true
			}
		}
	}
	for _, reference := range required {
		if !present[reference] {
			return fmt.Errorf("external proof requires successful flow round-trip reference %q", reference)
		}
	}
	return nil
}

func validateObservedProofExchange(exchange ObservedHTTPExchange) error {
	parsed, err := url.Parse(exchange.Request.Target)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return errors.New("requires an observed HTTPS request without userinfo")
	}
	if exchange.Response.Status < 100 || exchange.Response.Status > 599 {
		return errors.New("has an invalid response status")
	}
	for _, body := range []ObservedBody{exchange.Request.Body, exchange.Response.Body} {
		if !body.Complete {
			return errors.New("has an incomplete body capture")
		}
		if body.Truncated {
			return errors.New("has a truncated body capture")
		}
	}
	return nil
}

func sanitizeExternalProof(input ExternalProofInput, prepared []string, salt []byte, scope externalProofCredentialScope) (externalProofArtifact, error) {
	artifact := externalProofArtifact{
		Version:                externalProofVersion,
		RedactionStrategy:      externalProofFingerprintID,
		Connector:              input.Connector,
		RunID:                  input.RunID,
		PMBinarySHA256:         input.BinarySHA256,
		CredentialScope:        scope.credentialScope,
		CredentialScopeProof:   scope.proof,
		CredentialFingerprints: fingerprintExternalPreparedValues(salt, prepared),
		Process: externalProofProcess{
			Command:           append([]string(nil), input.Command...),
			ExitCode:          input.ExitCode,
			StdoutFingerprint: externalProofFingerprint(salt, input.Stdout),
			StderrFingerprint: externalProofFingerprint(salt, input.Stderr),
		},
		HTTPExchanges:           make([]externalProofHTTPExchange, 0, len(input.HTTPExchanges)),
		FlowRoundTripReferences: append([]string(nil), input.FlowRoundTripReferences...),
	}
	for _, exchange := range input.HTTPExchanges {
		sanitized, err := sanitizeObservedProofExchange(exchange, prepared, salt)
		if err != nil {
			return externalProofArtifact{}, err
		}
		artifact.HTTPExchanges = append(artifact.HTTPExchanges, sanitized)
	}
	return artifact, nil
}

func sanitizeObservedProofExchange(exchange ObservedHTTPExchange, prepared []string, salt []byte) (externalProofHTTPExchange, error) {
	parsed, err := url.Parse(exchange.Request.Target)
	if err != nil {
		return externalProofHTTPExchange{}, errors.New("parse observed request target")
	}
	query := parsed.Query()
	queryNames := make([]string, 0, len(query))
	for name := range query {
		queryNames = append(queryNames, name)
	}
	sort.Strings(queryNames)
	fields := make([]externalProofField, 0)
	for _, name := range queryNames {
		for _, value := range query[name] {
			fields = append(fields, externalProofField{
				Name:  sanitizeExternalProofFieldName(name, prepared, salt),
				Value: externalProofFingerprintText(salt, value, prepared),
			})
		}
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	requestBody, err := sanitizeExternalProofBody(exchange.Request.Body, prepared, salt)
	if err != nil {
		return externalProofHTTPExchange{}, err
	}
	responseBody, err := sanitizeExternalProofBody(exchange.Response.Body, prepared, salt)
	if err != nil {
		return externalProofHTTPExchange{}, err
	}
	return externalProofHTTPExchange{
		Request: externalProofHTTPRequest{
			Method:  strings.ToUpper(exchange.Request.Method),
			Target:  externalProofFingerprintText(salt, parsed.String(), prepared),
			Query:   fields,
			Headers: sanitizeExternalProofHeaders(exchange.Request.Headers, prepared, salt),
			Body:    requestBody,
		},
		Response: externalProofHTTPResponse{
			Status:  exchange.Response.Status,
			Headers: sanitizeExternalProofHeaders(exchange.Response.Headers, prepared, salt),
			Body:    responseBody,
		},
	}, nil
}

func sanitizeExternalProofHeaders(headers map[string][]string, prepared []string, salt []byte) []externalProofField {
	fields := make([]externalProofField, 0)
	for name, values := range headers {
		for _, value := range values {
			fields = append(fields, externalProofField{
				Name:  sanitizeExternalProofFieldName(name, prepared, salt),
				Value: externalProofFingerprintText(salt, value, prepared),
			})
		}
	}
	sort.Slice(fields, func(left, right int) bool {
		if fields[left].Name == fields[right].Name {
			return fields[left].Value < fields[right].Value
		}
		return fields[left].Name < fields[right].Name
	})
	return fields
}

func sanitizeExternalProofFieldName(name string, prepared []string, salt []byte) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if externalProofFieldName.MatchString(name) && !containsExternalPreparedValue(name, prepared) {
		return name
	}
	return externalProofFingerprintText(salt, name, prepared)
}

func sanitizeExternalProofBody(body ObservedBody, prepared []string, salt []byte) (externalProofObservedBody, error) {
	if !body.Complete || body.Truncated {
		return externalProofObservedBody{}, errors.New("refuse incomplete or truncated body at serialization boundary")
	}
	if len(body.Bytes) == 0 {
		return externalProofObservedBody{Encoding: "none", Value: json.RawMessage("null"), OriginalBytes: body.OriginalBytes}, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(body.Bytes))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err == nil && ensureExternalProofJSONEOF(decoder) == nil {
		raw, marshalErr := json.Marshal(fingerprintExternalProofJSON(salt, value, prepared))
		if marshalErr != nil {
			return externalProofObservedBody{}, errors.New("render fingerprinted JSON body")
		}
		return externalProofObservedBody{Encoding: "json", Value: raw, OriginalBytes: body.OriginalBytes}, nil
	}
	raw, err := json.Marshal(externalProofFingerprintText(salt, string(body.Bytes), prepared))
	if err != nil {
		return externalProofObservedBody{}, errors.New("render fingerprinted opaque body")
	}
	return externalProofObservedBody{Encoding: "opaque", Value: raw, OriginalBytes: body.OriginalBytes}, nil
}

func fingerprintExternalProofJSON(salt []byte, value any, prepared []string) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			out[sanitizeExternalProofFieldName(key, prepared, salt)] = fingerprintExternalProofJSON(salt, child, prepared)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, child := range typed {
			out[index] = fingerprintExternalProofJSON(salt, child, prepared)
		}
		return out
	case nil:
		return nil
	case json.Number:
		return externalProofFingerprintText(salt, typed.String(), prepared)
	case string:
		return externalProofFingerprintText(salt, typed, prepared)
	case bool:
		return externalProofFingerprintText(salt, fmt.Sprintf("%t", typed), prepared)
	default:
		return externalProofFingerprintText(salt, fmt.Sprint(typed), prepared)
	}
}

func ensureExternalProofJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return errors.New("extra JSON value")
	} else if !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func normalizedExternalProofValues(values []string) []string {
	unique := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value != "" {
			unique[value] = struct{}{}
		}
	}
	prepared := make([]string, 0, len(unique))
	for value := range unique {
		prepared = append(prepared, value)
	}
	sort.Slice(prepared, func(left, right int) bool {
		if len(prepared[left]) == len(prepared[right]) {
			return prepared[left] < prepared[right]
		}
		return len(prepared[left]) > len(prepared[right])
	})
	return prepared
}

func fingerprintExternalPreparedValues(salt []byte, prepared []string) []string {
	values := make([]string, 0, len(prepared))
	for _, value := range prepared {
		values = append(values, externalProofFingerprint(salt, value))
	}
	sort.Strings(values)
	return values
}

func externalProofFingerprintText(salt []byte, value string, prepared []string) string {
	if value == "" {
		return externalProofFingerprint(salt, "")
	}
	var out strings.Builder
	for remaining := value; remaining != ""; {
		index, matched := nextExternalPreparedValue(remaining, prepared)
		if index < 0 {
			out.WriteString(externalProofFingerprint(salt, remaining))
			break
		}
		if index > 0 {
			out.WriteString(externalProofFingerprint(salt, remaining[:index]))
		}
		out.WriteString(externalProofFingerprint(salt, matched))
		remaining = remaining[index+len(matched):]
	}
	return out.String()
}

func nextExternalPreparedValue(value string, prepared []string) (int, string) {
	index, matched := -1, ""
	for _, candidate := range prepared {
		candidateIndex := strings.Index(value, candidate)
		if candidateIndex < 0 || (index >= 0 && candidateIndex >= index) {
			continue
		}
		index, matched = candidateIndex, candidate
	}
	return index, matched
}

func containsExternalPreparedValue(value string, prepared []string) bool {
	for _, candidate := range prepared {
		if strings.Contains(value, candidate) {
			return true
		}
	}
	return false
}

func externalProofFingerprint(salt []byte, value string) string {
	mac := hmac.New(sha256.New, salt)
	_, _ = mac.Write([]byte(value))
	return externalProofMarkerPrefix + hex.EncodeToString(mac.Sum(nil)) + externalProofMarkerSuffix
}

func externalProofOutputPath(root, connector, runID string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", errors.New("external proof root is required")
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve external proof root: %w", err)
	}
	return filepath.Join(absRoot, ".polymetrics", "certifications", "external-proof", connector, runID+".json"), nil
}

func externalProofSalt(root string) ([]byte, error) {
	if salt, err := readExternalProofSalt(root); err == nil {
		return salt, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	path, err := externalProofSaltPath(root)
	if err != nil {
		return nil, err
	}
	salt := make([]byte, sha256.Size)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("generate external proof salt: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create external proof salt directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return readExternalProofSalt(root)
	}
	if err != nil {
		return nil, fmt.Errorf("create external proof salt: %w", err)
	}
	if _, err := file.Write(salt); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("write external proof salt: %w", err)
	}
	if err := file.Close(); err != nil {
		return nil, fmt.Errorf("close external proof salt: %w", err)
	}
	return salt, nil
}

func readExternalProofSalt(root string) ([]byte, error) {
	path, err := externalProofSaltPath(root)
	if err != nil {
		return nil, err
	}
	salt, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(salt) != sha256.Size {
		return nil, errors.New("external proof salt has invalid length")
	}
	return salt, nil
}

func externalProofSaltPath(root string) (string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve external proof root: %w", err)
	}
	return filepath.Join(absRoot, ".polymetrics", "certifications", ".external-proof-salt"), nil
}
