package certify

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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
	externalProofVersion       = 1
	externalProofFingerprintID = "repository_salted_hmac_sha256_v1"
	externalProofMarkerPrefix  = "{{pmcertfp:v1:"
	externalProofMarkerSuffix  = "}}"
	externalProofMaxExchanges  = 16
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

// WriteExternalProof accepts only a completed full-parity external run and
// creates its artifact after all bounded-capture and sanitization checks pass.
// A rejected input performs no artifact or salt writes.
func WriteExternalProof(root string, input ExternalProofInput) (string, error) {
	prepared, err := validateExternalProofInput(input)
	if err != nil {
		return "", err
	}

	salt, err := externalProofSalt(root)
	if err != nil {
		return "", err
	}
	artifact, err := sanitizeExternalProof(input, prepared, salt)
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
	if artifact.Version != externalProofVersion || artifact.RedactionStrategy != externalProofFingerprintID {
		return false, errors.New("external proof has an unsupported schema")
	}
	stdoutFingerprint := externalProofFingerprint(salt, stdout)
	stderrFingerprint := externalProofFingerprint(salt, stderr)
	return subtle.ConstantTimeCompare([]byte(artifact.Process.StdoutFingerprint), []byte(stdoutFingerprint)) == 1 &&
		subtle.ConstantTimeCompare([]byte(artifact.Process.StderrFingerprint), []byte(stderrFingerprint)) == 1, nil
}

func validateExternalProofInput(input ExternalProofInput) ([]string, error) {
	if !input.Passed || input.ExitCode != 0 {
		return nil, errors.New("external proof requires a completed successful process")
	}
	if !input.FullParity {
		return nil, errors.New("external proof requires a full-parity credential")
	}
	if !externalProofIdentifier.MatchString(input.Connector) || !externalProofIdentifier.MatchString(input.RunID) {
		return nil, errors.New("external proof connector and run id must be safe identifiers")
	}
	if !externalProofSHA256.MatchString(input.BinarySHA256) {
		return nil, errors.New("external proof requires a lowercase SHA-256 built-binary fingerprint")
	}
	if len(input.Command) == 0 {
		return nil, errors.New("external proof requires the exact observed process command")
	}
	prepared := normalizedExternalProofValues(input.PreparedValues)
	if len(prepared) == 0 {
		return nil, errors.New("external proof requires prepared credential values")
	}
	if len(ScanForSecrets(strings.Join(input.Command, "\x00"), prepared)) != 0 {
		return nil, errors.New("external proof refuses credential material in process arguments")
	}
	if len(ScanForSecrets(input.Stdout, prepared)) != 0 || len(ScanForSecrets(input.Stderr, prepared)) != 0 {
		return nil, errors.New("external proof refuses credential material in observed process output")
	}
	if len(input.HTTPExchanges) == 0 {
		return nil, errors.New("external proof requires at least one observed HTTPS exchange")
	}
	if len(input.HTTPExchanges) > externalProofMaxExchanges {
		return nil, fmt.Errorf("external proof observed %d exchanges, exceeding bounded redirect/retry limit %d", len(input.HTTPExchanges), externalProofMaxExchanges)
	}
	for index, exchange := range input.HTTPExchanges {
		if err := validateObservedProofExchange(exchange); err != nil {
			return nil, fmt.Errorf("external proof exchange %d: %w", index+1, err)
		}
	}
	if err := validateExternalProofFlowRoundTripReferences(input.FlowRoundTripReferences); err != nil {
		return nil, err
	}
	return prepared, nil
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

func sanitizeExternalProof(input ExternalProofInput, prepared []string, salt []byte) (externalProofArtifact, error) {
	artifact := externalProofArtifact{
		Version:                externalProofVersion,
		RedactionStrategy:      externalProofFingerprintID,
		Connector:              input.Connector,
		RunID:                  input.RunID,
		PMBinarySHA256:         input.BinarySHA256,
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
