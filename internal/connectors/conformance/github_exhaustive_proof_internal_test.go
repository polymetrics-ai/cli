package conformance

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

var (
	githubProviderPathParam   = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	githubProviderGraphQLRoot = regexp.MustCompile(`\{\s*([A-Za-z_][A-Za-z0-9_]*)`)
)

type githubProviderDoubleReport struct {
	SchemaVersion  int                             `json:"schema_version"`
	Connector      string                          `json:"connector"`
	Streams        int                             `json:"streams"`
	WriteActions   int                             `json:"write_actions"`
	Operations     int                             `json:"operations"`
	GenericStreams int                             `json:"generic_streams"`
	GenericWrites  int                             `json:"generic_write_actions"`
	Exercised      int                             `json:"exercised"`
	Untestable     int                             `json:"untestable"`
	Failed         int                             `json:"failed"`
	Failures       []string                        `json:"failures,omitempty"`
	Rows           []githubProviderDoubleReportRow `json:"rows"`
}

type githubProviderDoubleReportRow struct {
	Kind         string                             `json:"kind"`
	Name         string                             `json:"name"`
	State        string                             `json:"state"`
	Route        string                             `json:"route"`
	RequestCount int                                `json:"request_count,omitempty"`
	Requests     []githubProviderDoubleRequestProof `json:"requests,omitempty"`
	Response     *githubProviderDoubleResponseProof `json:"response,omitempty"`
	Reason       string                             `json:"reason,omitempty"`
}

type githubProviderDoubleRequestProof struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	QueryKeys   []string `json:"query_keys,omitempty"`
	QuerySHA256 string   `json:"query_sha256,omitempty"`
	HeaderNames []string `json:"header_names,omitempty"`
	ContentType string   `json:"content_type,omitempty"`
	BodyKeys    []string `json:"body_keys,omitempty"`
	BodySHA256  string   `json:"body_sha256,omitempty"`
}

type githubProviderDoubleResponseProof struct {
	Status     int    `json:"status"`
	Bytes      int    `json:"bytes,omitempty"`
	BodySHA256 string `json:"body_sha256,omitempty"`
}

type githubProviderCapturedRequest struct {
	Method      string
	Path        string
	Query       map[string][]string
	HeaderNames []string
	ContentType string
	Body        map[string]any
	BodyBytes   []byte
}

type githubProviderCapture struct {
	*httptest.Server
	mu       sync.Mutex
	requests []githubProviderCapturedRequest
	response func(*http.Request) (int, string, []byte)
}

func newGitHubProviderCapture(response func(*http.Request) (int, string, []byte)) *githubProviderCapture {
	capture := &githubProviderCapture{response: response}
	capture.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(io.LimitReader(r.Body, 2<<20))
		body := map[string]any(nil)
		if len(bodyBytes) > 0 {
			decoder := json.NewDecoder(strings.NewReader(string(bodyBytes)))
			decoder.UseNumber()
			_ = decoder.Decode(&body)
		}
		headerNames := make([]string, 0, len(r.Header))
		for name := range r.Header {
			headerNames = append(headerNames, name)
		}
		sort.Strings(headerNames)
		capture.mu.Lock()
		capture.requests = append(capture.requests, githubProviderCapturedRequest{
			Method:      r.Method,
			Path:        r.URL.Path,
			Query:       r.URL.Query(),
			HeaderNames: headerNames,
			ContentType: r.Header.Get("Content-Type"),
			Body:        body,
			BodyBytes:   append([]byte(nil), bodyBytes...),
		})
		capture.mu.Unlock()

		status, contentType, responseBody := 200, "application/json", []byte(`{}`)
		if capture.response != nil {
			status, contentType, responseBody = capture.response(r)
		}
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.WriteHeader(status)
		_, _ = w.Write(responseBody)
	}))
	return capture
}

func (c *githubProviderCapture) captured() []githubProviderCapturedRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]githubProviderCapturedRequest(nil), c.requests...)
}

func githubProviderDoubleBundle(b engine.Bundle, baseURL string) engine.Bundle {
	b.HTTP.URL = baseURL
	for index := range b.Writes {
		if b.Writes[index].BaseURL != "" {
			b.Writes[index].BaseURL = baseURL
		}
		if len(b.Writes[index].AllowedBaseURLOrigins) > 0 {
			// The double replaces both the connector and action endpoints with its
			// local origin. Preserve the installed action's origin-bound security
			// path by replacing its declared source origin with that same local
			// origin; production declarations retain their real allow-list.
			b.Writes[index].AllowedBaseURLOrigins = []string{baseURL}
		}
	}
	// The provider double deliberately exercises transport and engine request
	// construction, not GitHub authentication. Removing the bundle's auth
	// candidates prevents synthetic values from ever being mistaken for a
	// credential and makes every request's header set inspectable.
	b.HTTP.Auth = nil
	b.HTTP.Headers = nil
	b.HTTP.RateLimit = nil
	b.RateLimits = nil
	return b
}

func githubProviderDoubleConfig(b engine.Bundle) connectors.RuntimeConfig {
	cfg := runtimeConfigForEngine(b)
	if cfg.Config == nil {
		cfg.Config = map[string]string{}
	}
	cfg.Config["owner"] = "provider-double-owner"
	cfg.Config["repo"] = "provider-double-repo"
	cfg.Config["org"] = "provider-double-org"
	cfg.Config["username"] = "provider-double-user"
	cfg.Config["base_url"] = "http://provider-double.invalid"
	return cfg
}

func runGitHubExhaustiveProviderDouble(t *testing.T) (githubProviderDoubleReport, error) {
	t.Helper()
	b, err := engine.Load(os.DirFS("../defs"), "github")
	if err != nil {
		return githubProviderDoubleReport{}, fmt.Errorf("load github bundle: %w", err)
	}
	report := githubProviderDoubleReport{
		SchemaVersion: 1,
		Connector:     "github",
		Streams:       len(b.Streams),
		WriteActions:  len(b.Writes),
		Operations:    len(b.Operations),
		Rows:          make([]githubProviderDoubleReportRow, 0, len(b.Streams)+len(b.Writes)+len(b.Operations)),
	}
	streamRows := make(map[string]githubProviderDoubleReportRow, len(b.Streams))
	for _, stream := range b.Streams {
		check := checkReadFixtureNonempty(b, stream.Name, true)
		row := githubProviderDoubleReportRow{
			Kind:  "stream",
			Name:  stream.Name,
			Route: "engine.Read against the declared fixture replay",
			State: "exercised",
		}
		if !githubStreamHasCommand(b, stream.Name) {
			row.Route = "pm etl generic route -> engine.Read against the declared fixture replay"
			report.GenericStreams++
		}
		if !check.Passed {
			row.State = "failed"
			report.fail(fmt.Sprintf("stream:%s", stream.Name), check.Error)
		}
		streamRows[stream.Name] = row
		report.Rows = append(report.Rows, row)
	}

	for _, action := range b.Writes {
		row := runGitHubWriteProviderDouble(t, b, action)
		if !githubWriteActionHasCommand(b, action.Name) {
			row.Route = "pm reverse generic route -> engine.Write with plan, preview, approval, execute"
			report.GenericWrites++
		}
		report.Rows = append(report.Rows, row)
		if row.State == "failed" {
			report.fail(fmt.Sprintf("write:%s", action.Name), row.Reason)
		}
	}

	for _, operation := range b.Operations {
		row := runGitHubOperationProviderDouble(t, b, operation, b.Streams, streamRows)
		report.Rows = append(report.Rows, row)
		switch row.State {
		case "failed":
			report.fail(fmt.Sprintf("operation:%s", operation.ID), row.Reason)
		case "untestable":
			report.Untestable++
		}
	}
	report.Exercised = 0
	for _, row := range report.Rows {
		if row.State == "exercised" {
			report.Exercised++
		}
	}
	if reportPath := strings.TrimSpace(os.Getenv("POLYMETRICS_GITHUB_PROVIDER_DOUBLE_REPORT")); reportPath != "" {
		if err := writeGitHubProviderDoubleReport(reportPath, report); err != nil {
			return report, err
		}
	}
	return report, nil
}

func githubStreamHasCommand(b engine.Bundle, name string) bool {
	if b.CLISurface == nil {
		return false
	}
	for _, command := range b.CLISurface.Commands {
		if command.Stream == name {
			return true
		}
	}
	return false
}

func githubWriteActionHasCommand(b engine.Bundle, name string) bool {
	if b.CLISurface == nil {
		return false
	}
	for _, command := range b.CLISurface.Commands {
		if command.Write == name {
			return true
		}
	}
	return false
}

func (r *githubProviderDoubleReport) fail(name, reason string) {
	r.Failed++
	// Keep report failures bounded and free of request bodies or credential-like
	// values. The Go test log retains the detailed assertion for local repair.
	if len(r.Failures) < 32 {
		r.Failures = append(r.Failures, name+": "+safeProviderDoubleReason(reason))
	}
}

func safeProviderDoubleReason(reason string) string {
	lower := strings.ToLower(strings.TrimSpace(reason))
	for _, sensitive := range []string{"token", "credential", "secret", "authorization", "bearer", "password"} {
		if strings.Contains(lower, sensitive) {
			return "provider-double assertion failed (sensitive diagnostic redacted)"
		}
	}
	return "provider-double assertion failed"
}

func writeGitHubProviderDoubleReport(name string, report githubProviderDoubleReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal github provider-double report: %w", err)
	}
	if err := os.WriteFile(filepath.Clean(name), append(data, '\n'), 0o600); err != nil {
		return fmt.Errorf("write github provider-double report: %w", err)
	}
	return nil
}

func runGitHubWriteProviderDouble(t *testing.T, b engine.Bundle, action engine.WriteAction) githubProviderDoubleReportRow {
	t.Helper()
	row := githubProviderDoubleReportRow{
		Kind:  "write_action",
		Name:  action.Name,
		Route: "engine.Write with plan, preview, approval, execute",
		State: "failed",
	}
	fixture, fixtureErr := loadWriteFixture(b.Fixtures, action.Name)
	record, synthesisErr := syntheticGitHubRecord(action)
	if synthesisErr != nil {
		row.Reason = synthesisErr.Error()
		return row
	}
	if fixtureErr == nil {
		record = connectors.Record(fixture.Record)
	}
	var expected *writeExpectation
	if fixtureErr == nil {
		expected = &fixture.Expect
	}
	capture := newGitHubProviderCapture(func(_ *http.Request) (int, string, []byte) {
		if action.Name == "create_pull_request" {
			return http.StatusOK, "application/json", []byte(`{"number":301}`)
		}
		if fixture.Response != nil {
			status := fixture.Response.Status
			if status == 0 {
				status = http.StatusOK
			}
			contentType := "application/json"
			if len(fixture.Response.Headers["Content-Type"]) > 0 {
				contentType = fixture.Response.Headers["Content-Type"][0]
			}
			return status, contentType, fixture.Response.Body
		}
		if len(action.SuccessStatuses) > 0 {
			// An unfixtured action can still declare an exact successful provider
			// receipt. The double must return that receipt instead of assuming
			// generic 200 success, so narrowed write-status policies stay tested.
			return action.SuccessStatuses[0], "application/json", []byte(`{}`)
		}
		return http.StatusOK, "application/json", []byte(`{}`)
	})
	defer capture.Close()
	doubleBundle := githubProviderDoubleBundle(b, capture.URL)
	cfg := runtimeConfigForEngine(doubleBundle)
	if fixtureErr != nil {
		cfg = githubProviderDoubleConfig(doubleBundle)
	}
	if action.BinaryUpload != nil {
		cfg.ProjectDir = t.TempDir()
		fileName := "provider-double-upload.bin"
		if err := os.WriteFile(filepath.Join(cfg.ProjectDir, fileName), []byte("provider-double"), 0o600); err != nil {
			row.Reason = err.Error()
			return row
		}
		record[action.BinaryUpload.SourceField] = fileName
	}
	request, err := approvedFixtureWriteRequest(context.Background(), doubleBundle, action.Name, cfg, []connectors.Record{record}, engine.HooksFor(doubleBundle.Name))
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	result, err := engine.Write(context.Background(), doubleBundle, request, []connectors.Record{record}, engine.HooksFor(doubleBundle.Name))
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	if result.RecordsWritten != 1 || result.RecordsFailed != 0 {
		row.Reason = fmt.Sprintf("write result recorded %d written and %d failed", result.RecordsWritten, result.RecordsFailed)
		return row
	}
	requests := capture.captured()
	if len(requests) == 0 {
		row.Reason = "approved write emitted no provider-double request"
		return row
	}
	if expected != nil {
		last := requests[len(requests)-1]
		if mismatch := compareWriteExpectation(capturedRequest{Method: last.Method, Path: last.Path, Query: valuesFromMap(last.Query), Body: last.Body}, *expected); mismatch != "" {
			row.Reason = mismatch
			return row
		}
	}
	row.State = "exercised"
	row.RequestCount = len(requests)
	row.Requests = githubProviderRequestProofs(requests)
	return row
}

func valuesFromMap(values map[string][]string) map[string][]string { return values }

func syntheticGitHubRecord(action engine.WriteAction) (connectors.Record, error) {
	record := connectors.Record{}
	if len(action.RecordSchema) > 0 {
		var schema struct {
			Properties map[string]json.RawMessage `json:"properties"`
			Required   []string                   `json:"required"`
		}
		if json.Unmarshal(action.RecordSchema, &schema) == nil {
			for _, name := range schema.Required {
				if property, exists := schema.Properties[name]; exists {
					value, err := syntheticSchemaValue(property, name)
					if err != nil {
						return nil, fmt.Errorf("synthetic %s.%s: %w", action.Name, name, err)
					}
					record[name] = value
				}
			}
		}
	}
	for _, field := range action.PathFields {
		if _, ok := record[field]; !ok {
			record[field] = syntheticFieldValue(field)
		}
	}
	for _, match := range regexp.MustCompile(`\{\{\s*record\.([^}\s]+)\s*\}\}`).FindAllStringSubmatch(action.Path, -1) {
		field := match[1]
		if _, ok := record[field]; !ok {
			record[field] = syntheticFieldValue(field)
		}
	}
	switch action.Name {
	case "close_issue", "reopen_issue":
		record["issue_number"] = 101
	case "close_pull_request", "reopen_pull_request", "update_pull_request":
		record["pull_number"] = 101
	case "create_pull_request":
		record["head"] = "provider-double-branch"
		record["base"] = "main"
		record["title"] = "provider-double"
	case "create_label", "update_label":
		record["name"] = "provider-double-label"
		record["color"] = "ffffff"
	}
	return record, nil
}

func syntheticSchemaValue(raw json.RawMessage, field string) (any, error) {
	var schema map[string]json.RawMessage
	if json.Unmarshal(raw, &schema) != nil {
		return syntheticFieldValue(field), nil
	}
	var enum []any
	if json.Unmarshal(schema["enum"], &enum) == nil && len(enum) > 0 {
		return enum[0], nil
	}
	var union []json.RawMessage
	for _, key := range []string{"oneOf", "anyOf"} {
		if json.Unmarshal(schema[key], &union) == nil && len(union) > 0 {
			return syntheticSchemaValue(union[0], field)
		}
	}
	kind := syntheticSchemaType(schema["type"])
	var format string
	_ = json.Unmarshal(schema["format"], &format)
	switch kind {
	case "integer":
		return 1, nil
	case "number":
		return 1, nil
	case "boolean":
		return true, nil
	case "string":
		// GitHub secret-set actions accept a caller-sealed ciphertext, not a
		// plaintext value. The provider-double must exercise that declared
		// encrypted_value schema without inventing or retaining secret material.
		if strings.Contains(strings.ToLower(field), "encrypted_value") {
			return base64.StdEncoding.EncodeToString([]byte("provider-double")), nil
		}
		if format == "uri" {
			return "https://provider-double.invalid/resource", nil
		}
		return syntheticPatternString(schema, field)
	case "array":
		var items json.RawMessage
		count := 1
		_ = json.Unmarshal(schema["minItems"], &count)
		if count < 1 {
			count = 1
		}
		if count > 64 {
			return nil, fmt.Errorf("unsupported minItems %d exceeds deterministic witness limit", count)
		}
		if json.Unmarshal(schema["items"], &items) == nil && len(items) > 0 {
			values := make([]any, count)
			for index := range values {
				value, err := syntheticSchemaValue(items, field)
				if err != nil {
					return nil, err
				}
				values[index] = value
			}
			return values, nil
		}
		return make([]any, count), nil
	case "object":
		var properties map[string]json.RawMessage
		_ = json.Unmarshal(schema["properties"], &properties)
		out := map[string]any{}
		var required []string
		_ = json.Unmarshal(schema["required"], &required)
		for _, name := range required {
			if property, exists := properties[name]; exists {
				value, err := syntheticSchemaValue(property, name)
				if err != nil {
					return nil, err
				}
				out[name] = value
			}
		}
		return out, nil
	default:
		return syntheticFieldValue(field), nil
	}
}

func syntheticSchemaType(raw json.RawMessage) string {
	var single string
	if json.Unmarshal(raw, &single) == nil {
		return single
	}
	var union []string
	if json.Unmarshal(raw, &union) == nil {
		for _, candidate := range union {
			if candidate != "null" {
				return candidate
			}
		}
	}
	return ""
}

func syntheticPatternString(schema map[string]json.RawMessage, _ string) (string, error) {
	fallback := "provider-double"
	minLength := 0
	_ = json.Unmarshal(schema["minLength"], &minLength)
	var pattern string
	if json.Unmarshal(schema["pattern"], &pattern) != nil || pattern == "" {
		return syntheticMinimumLengthString(fallback, minLength), nil
	}
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		return "", fmt.Errorf("unsupported invalid pattern %q", pattern)
	}
	candidates := []string{
		fallback,
		"1.0.0",
		strings.Repeat("0", 40),
		strings.Repeat("0", 64),
		"sha256:" + strings.Repeat("0", 64),
		"refs/heads/provider-double",
		"https://provider-double.invalid/registry",
		"ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEZha2VQcm92aWRlckRvdWJsZUtleQ==",
		"ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTY=",
		base64.StdEncoding.EncodeToString([]byte("provider-double")),
	}
	for _, candidate := range candidates {
		if len(candidate) >= minLength && compiled.MatchString(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("unsupported pattern %q with minLength %d", pattern, minLength)
}

func syntheticMinimumLengthString(value string, minLength int) string {
	if minLength <= len(value) {
		return value
	}
	return value + strings.Repeat("x", minLength-len(value))
}

// syntheticGraphQLVariables materializes required caller inputs and one
// declaration-owned forward pagination direction. The direct-read executor
// refuses neither/both directions, so the provider double must make an
// explicit, minimal choice rather than omit paging variables entirely.
func syntheticGraphQLVariables(operation engine.OperationSpec) (map[string]any, error) {
	if operation.GraphQL == nil {
		return nil, fmt.Errorf("GraphQL operation has no declaration")
	}
	var root struct {
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(operation.GraphQL.VariablesSchema, &root); err != nil {
		return nil, fmt.Errorf("decode fixed GraphQL variables schema: %w", err)
	}
	variables := make(map[string]any, len(root.Required))
	names := append([]string(nil), root.Required...)
	sort.Strings(names)
	for _, name := range names {
		variables[name] = syntheticGraphQLSchemaValue(root.Properties[name], name)
	}
	if pagination := operation.GraphQL.Pagination; pagination != nil && pagination.PageSizeVariable != "" {
		variables[pagination.PageSizeVariable] = 1
	}
	return variables, nil
}

// syntheticGraphQLSchemaValue differs from the older write-fixture helper in
// one important way: an empty closed object stays empty.  Adding a fallback
// "name" property would be an undeclared input and must fail the exact schema
// which this provider-double path is meant to exercise.
func syntheticGraphQLSchemaValue(raw json.RawMessage, field string) any {
	var schema map[string]json.RawMessage
	if json.Unmarshal(raw, &schema) != nil {
		return syntheticFieldValue(field)
	}
	var enum []any
	if json.Unmarshal(schema["enum"], &enum) == nil && len(enum) > 0 {
		return enum[0]
	}
	var union []json.RawMessage
	for _, key := range []string{"oneOf", "anyOf"} {
		if json.Unmarshal(schema[key], &union) == nil && len(union) > 0 {
			return syntheticGraphQLSchemaValue(union[0], field)
		}
	}
	kind := syntheticSchemaType(schema["type"])
	var format string
	_ = json.Unmarshal(schema["format"], &format)
	switch kind {
	case "integer":
		return 1
	case "number":
		return 1
	case "boolean":
		return true
	case "string":
		if format == "uri" {
			return "https://provider-double.invalid/resource"
		}
		return "provider-double"
	case "array":
		var items json.RawMessage
		if json.Unmarshal(schema["items"], &items) == nil && len(items) > 0 {
			return []any{syntheticGraphQLSchemaValue(items, field)}
		}
		return []any{}
	case "object":
		var properties map[string]json.RawMessage
		_ = json.Unmarshal(schema["properties"], &properties)
		var required []string
		_ = json.Unmarshal(schema["required"], &required)
		out := make(map[string]any, len(required))
		names := append([]string(nil), required...)
		sort.Strings(names)
		for _, name := range names {
			out[name] = syntheticGraphQLSchemaValue(properties[name], name)
		}
		return out
	default:
		return syntheticFieldValue(field)
	}
}

func syntheticFieldValue(field string) any {
	lower := strings.ToLower(field)
	if strings.Contains(lower, "force") || lower == "archived" {
		return true
	}
	if strings.Contains(lower, "labels") || strings.Contains(lower, "assignees") || strings.Contains(lower, "reviewers") || strings.Contains(lower, "digests") || strings.Contains(lower, "ids") || strings.Contains(lower, "options") {
		return []any{"provider-double-item"}
	}
	if strings.Contains(lower, "number") || strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id") || strings.Contains(lower, "port") {
		return 1
	}
	return "provider-double"
}

func githubProviderRequestProofs(requests []githubProviderCapturedRequest) []githubProviderDoubleRequestProof {
	proofs := make([]githubProviderDoubleRequestProof, 0, len(requests))
	for _, request := range requests {
		queryKeys := make([]string, 0, len(request.Query))
		for key := range request.Query {
			queryKeys = append(queryKeys, key)
		}
		sort.Strings(queryKeys)
		proof := githubProviderDoubleRequestProof{
			Method:      request.Method,
			Path:        request.Path,
			QueryKeys:   queryKeys,
			QuerySHA256: hashJSON(request.Query),
			HeaderNames: append([]string(nil), request.HeaderNames...),
			ContentType: request.ContentType,
			BodyKeys:    sortedMapKeys(request.Body),
			BodySHA256:  hashBytes(request.BodyBytes),
		}
		proofs = append(proofs, proof)
	}
	return proofs
}

func sortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func hashJSON(value any) string {
	data, _ := json.Marshal(value)
	return hashBytes(data)
}

func hashBytes(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func runGitHubOperationProviderDouble(t *testing.T, b engine.Bundle, operation engine.OperationSpec, streams []engine.StreamSpec, streamRows map[string]githubProviderDoubleReportRow) githubProviderDoubleReportRow {
	t.Helper()
	row := githubProviderDoubleReportRow{
		Kind:  "operation",
		Name:  operation.ID,
		Route: "declared operation executor against provider double",
		State: "failed",
	}
	switch operation.Kind {
	case "graphql_query":
		// A declared transport path marks a self-contained fixed query. Source-
		// generated roots use github.graphql.query.* IDs, but supplemental fixed
		// documents such as github.repo.list intentionally do not impersonate a
		// source root and must exercise the same direct-operation executor.
		if operation.GraphQL != nil && operation.GraphQL.Path != "" {
			return runGitHubGraphQLQueryProviderDouble(t, b, operation)
		}
		for _, stream := range streams {
			if stream.GraphQL != nil && stream.GraphQL.OperationName == operation.GraphQL.OperationName {
				if streamRows[stream.Name].State == "exercised" {
					row.State = "exercised"
					row.Route = "engine.Read GraphQL stream replay: " + stream.Name
					return row
				}
				row.Reason = "the matching GraphQL stream provider-double row failed"
				return row
			}
		}
		row.Reason = "no declared GraphQL stream binds operation " + operation.ID
		return row
	case "graphql_mutation":
		if strings.HasPrefix(operation.ID, "github.graphql.mutation.") {
			return runGitHubGraphQLMutationProviderDouble(t, b, operation)
		}
		row.State = "untestable"
		row.Reason = "GraphQL mutation has no fixed generated operation contract"
		return row
	case "local_git":
		row.State = "untestable"
		row.Reason = "local_git is a local workflow operation, not a provider operation executor"
		return row
	case "binary_download":
		return runGitHubBinaryProviderDouble(t, b, operation)
	case "rest_read":
		return runGitHubRestReadProviderDouble(t, b, operation)
	case "rest_write":
		for _, action := range b.Writes {
			if action.Name == strings.TrimPrefix(operation.ID, "github.") {
				row := runGitHubWriteProviderDouble(t, b, action)
				row.Kind = "operation"
				row.Name = operation.ID
				row.Route = "engine.Write for the operation's bound semantic write action"
				return row
			}
		}
		row.State = "untestable"
		row.Reason = "sensitive rest_write declaration has no executable semantic write action"
		return row
	default:
		row.Reason = "operation kind has no declared provider-double executor"
		return row
	}
}

func githubGraphQLProviderDoubleResponse(operation engine.OperationSpec) ([]byte, error) {
	if operation.GraphQL == nil {
		return nil, fmt.Errorf("GraphQL operation has no declaration")
	}
	match := githubProviderGraphQLRoot.FindStringSubmatch(operation.GraphQL.Document)
	if len(match) != 2 {
		return nil, fmt.Errorf("fixed GraphQL document has no root field")
	}
	root := match[1]
	value := any(map[string]any{"__typename": "ProviderDouble"})
	if operation.GraphQL.Pagination != nil {
		connection := any(map[string]any{
			"__typename": "ProviderDoubleConnection",
			"nodes":      []any{},
			"pageInfo":   map[string]any{"hasNextPage": false, "endCursor": nil, "hasPreviousPage": false, "startCursor": nil},
		})
		path := strings.Split(operation.GraphQL.Pagination.ConnectionPath, ".")
		if len(path) == 0 || path[0] != root {
			return nil, fmt.Errorf("GraphQL pagination connection path %q does not start at document root %q", operation.GraphQL.Pagination.ConnectionPath, root)
		}
		value = connection
		for index := len(path) - 1; index >= 1; index-- {
			value = map[string]any{
				"__typename": "ProviderDouble",
				path[index]:  value,
			}
		}
	}
	if root == "rateLimit" {
		value = map[string]any{"limit": 5000, "cost": 1, "remaining": 4999, "resetAt": "2026-08-09T00:00:00Z"}
	}
	data := map[string]any{root: value}
	if operation.Kind == "graphql_query" && root != "rateLimit" {
		data["rateLimit"] = map[string]any{"limit": 5000, "cost": 1, "remaining": 4999, "resetAt": "2026-08-09T00:00:00Z"}
	}
	return json.Marshal(map[string]any{"data": data})
}

func runGitHubGraphQLQueryProviderDouble(t *testing.T, b engine.Bundle, operation engine.OperationSpec) githubProviderDoubleReportRow {
	t.Helper()
	row := githubProviderDoubleReportRow{Kind: "operation", Name: operation.ID, Route: "engine.OperationDirectRead fixed GraphQL query", State: "failed"}
	variables, err := syntheticGraphQLVariables(operation)
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	responseBody, err := githubGraphQLProviderDoubleResponse(operation)
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	capture := newGitHubProviderCapture(func(_ *http.Request) (int, string, []byte) {
		return http.StatusOK, "application/json", responseBody
	})
	defer capture.Close()
	doubleBundle := githubProviderDoubleBundle(b, capture.URL)
	result, err := engine.OperationDirectRead(context.Background(), doubleBundle, connectors.OperationDirectReadRequest{
		Operation: operation.ID, Config: githubProviderDoubleConfig(doubleBundle), Body: variables,
		MaxBytes: operation.GraphQL.MaxBytes, OutputPolicy: operation.OutputPolicy,
	}, engine.HooksFor(doubleBundle.Name))
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	if result.GraphQL == nil || result.GraphQL.RateLimit == nil {
		row.Reason = "fixed GraphQL query did not report rate-limit metadata"
		return row
	}
	requests := capture.captured()
	if len(requests) != 1 {
		row.Reason = fmt.Sprintf("fixed GraphQL query made %d provider-double requests, want 1", len(requests))
		return row
	}
	row.State = "exercised"
	row.RequestCount = len(requests)
	row.Requests = githubProviderRequestProofs(requests)
	row.Response = &githubProviderDoubleResponseProof{Status: http.StatusOK, Bytes: len(responseBody), BodySHA256: hashBytes(responseBody)}
	return row
}

func approvedGitHubGraphQLMutationRequest(ctx context.Context, t *testing.T, b engine.Bundle, operation engine.OperationSpec, cfg connectors.RuntimeConfig, variables map[string]any) (connectors.OperationDirectWriteRequest, error) {
	t.Helper()
	authority, err := connectors.NewFixtureWriteApprovalAuthority()
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	cfg.CredentialRevision, err = authority.CredentialRevision("conformance:"+b.Name, cfg.Secrets)
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	cfg.ConfigurationDigest, err = authority.ConfigurationDigest("conformance:"+b.Name, cfg.Config)
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	cfg.WriteApprovalScope = connectors.WriteApprovalScopeFixture
	req := connectors.OperationDirectWriteRequest{Operation: operation.ID, Config: cfg, Body: variables, OutputPolicy: operation.OutputPolicy}
	preview, err := engine.PreviewOperationDirectWrite(ctx, b, req, engine.HooksFor(b.Name))
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.PreviewDigest = preview.Digest
	if preview.ApprovalTarget.Confirmation.Kind == "" {
		return req, nil
	}
	const approvalToken = "conformance-graphql-fixture-approval"
	grant, err := authority.IssueWriteGrant(connectors.WriteApprovalGrantRequest{
		PlanID:        "rplan_conformance_graphql_fixture",
		PlanHash:      strings.Repeat("a", 64),
		PreviewDigest: preview.Digest,
		ApprovalToken: approvalToken,
		Target:        preview.ApprovalTarget,
		Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
	})
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	req.Approval, err = authority.VerifyWriteGrant(grant, connectors.WriteApprovalExpectation{
		PlanID:        grant.PlanID,
		PlanHash:      grant.PlanHash,
		PreviewDigest: grant.PreviewDigest,
		ApprovalToken: approvalToken,
		Target:        grant.Target,
		Confirmation:  grant.Confirmation,
	})
	if err != nil {
		return connectors.OperationDirectWriteRequest{}, err
	}
	return req, nil
}

func runGitHubGraphQLMutationProviderDouble(t *testing.T, b engine.Bundle, operation engine.OperationSpec) githubProviderDoubleReportRow {
	t.Helper()
	row := githubProviderDoubleReportRow{Kind: "operation", Name: operation.ID, Route: "engine.PreviewOperationDirectWrite and OperationDirectWrite fixed GraphQL mutation", State: "failed"}
	variables, err := syntheticGraphQLVariables(operation)
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	responseBody, err := githubGraphQLProviderDoubleResponse(operation)
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	capture := newGitHubProviderCapture(func(_ *http.Request) (int, string, []byte) {
		return http.StatusOK, "application/json", responseBody
	})
	defer capture.Close()
	doubleBundle := githubProviderDoubleBundle(b, capture.URL)
	req, err := approvedGitHubGraphQLMutationRequest(context.Background(), t, doubleBundle, operation, githubProviderDoubleConfig(doubleBundle), variables)
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	result, err := engine.OperationDirectWrite(context.Background(), doubleBundle, req, engine.HooksFor(doubleBundle.Name))
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	if result.GraphQL == nil || len(result.GraphQL.Errors) != 0 {
		row.Reason = "fixed GraphQL mutation did not return a clean GraphQL result"
		return row
	}
	requests := capture.captured()
	if len(requests) != 1 {
		row.Reason = fmt.Sprintf("fixed GraphQL mutation made %d provider-double requests, want 1", len(requests))
		return row
	}
	row.State = "exercised"
	row.RequestCount = len(requests)
	row.Requests = githubProviderRequestProofs(requests)
	row.Response = &githubProviderDoubleResponseProof{Status: http.StatusOK, Bytes: len(responseBody), BodySHA256: hashBytes(responseBody)}
	return row
}

func operationPathParams(path string) map[string]string {
	params := map[string]string{}
	for _, match := range githubProviderPathParam.FindAllStringSubmatch(path, -1) {
		name := match[1]
		value := syntheticFieldValue(name)
		switch typed := value.(type) {
		case int:
			params[name] = fmt.Sprintf("%d", typed)
		default:
			params[name] = fmt.Sprint(typed)
		}
	}
	return params
}

func operationParameterValue(parameter engine.OperationParameter) string {
	if len(parameter.Values) > 0 {
		return parameter.Values[0]
	}
	switch parameter.Type {
	case "integer", "number":
		return "1"
	case "boolean":
		return "true"
	default:
		return "provider-double"
	}
}

func runGitHubRestReadProviderDouble(t *testing.T, b engine.Bundle, operation engine.OperationSpec) githubProviderDoubleReportRow {
	t.Helper()
	row := githubProviderDoubleReportRow{Kind: "operation", Name: operation.ID, Route: "engine.OperationDirectRead", State: "failed"}
	responseBody := []byte(`{}`)
	contentType := "application/json"
	status := http.StatusOK
	switch operation.OutputPolicy {
	case "none":
		responseBody = nil
		status = http.StatusNoContent
	case "text":
		responseBody = []byte("provider-double")
		contentType = "text/plain"
	}
	capture := newGitHubProviderCapture(func(_ *http.Request) (int, string, []byte) { return status, contentType, responseBody })
	defer capture.Close()
	doubleBundle := githubProviderDoubleBundle(b, capture.URL)
	cfg := githubProviderDoubleConfig(doubleBundle)
	pathParams := operationPathParams(operation.REST.Path)
	query := map[string]string{}
	body := map[string]any{}
	for _, parameter := range operation.REST.Parameters {
		value := operationParameterValue(parameter)
		switch parameter.In {
		case "path":
			pathParams[parameter.Name] = value
		case "query":
			query[parameter.Name] = value
		case "body":
			body[parameter.Name] = value
		}
	}
	if operation.ID == "github.orgs_list_attestations_bulk" || operation.ID == "github.users_list_attestations_bulk" {
		body["subject_digests"] = []string{"provider-double-digest"}
	}
	if operation.ID == "github.markdown" {
		body["text"] = "provider-double"
	}
	var rawBody *string
	if operation.ID == "github.markdown_raw" {
		text := "provider-double"
		rawBody = &text
	}
	outputPolicy := operation.OutputPolicy
	if outputPolicy == "json" {
		// The operation ledger keeps the provider's generic JSON declaration;
		// executable CLI contracts tighten it to the shared redacting policy.
		outputPolicy = "json_redacted"
	}
	_, err := engine.OperationDirectRead(context.Background(), doubleBundle, connectors.OperationDirectReadRequest{
		Operation: operation.ID, Config: cfg, PathParams: pathParams, Query: query, Body: body,
		RawBody: rawBody, MaxBytes: operation.REST.MaxBytes, OutputPolicy: outputPolicy,
	}, engine.HooksFor(doubleBundle.Name))
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	requests := capture.captured()
	if len(requests) != 1 {
		row.Reason = fmt.Sprintf("direct read made %d provider-double requests, want 1", len(requests))
		return row
	}
	row.State = "exercised"
	row.RequestCount = len(requests)
	row.Requests = githubProviderRequestProofs(requests)
	row.Response = &githubProviderDoubleResponseProof{Status: status, Bytes: len(responseBody), BodySHA256: hashBytes(responseBody)}
	return row
}

func runGitHubBinaryProviderDouble(t *testing.T, b engine.Bundle, operation engine.OperationSpec) githubProviderDoubleReportRow {
	t.Helper()
	row := githubProviderDoubleReportRow{Kind: "operation", Name: operation.ID, Route: "engine.OperationBinaryDownload", State: "failed"}
	if operation.Binary.ExtractArchives {
		capture := newGitHubProviderCapture(nil)
		defer capture.Close()
		doubleBundle := githubProviderDoubleBundle(b, capture.URL)
		temp := t.TempDir()
		_, err := engine.OperationBinaryDownload(context.Background(), doubleBundle, engine.BinaryDownloadRequest{
			Operation: operation.ID, Config: githubProviderDoubleConfig(doubleBundle),
			PathParams: operationPathParams(operation.Binary.Path), DestRoot: temp, FileName: "artifact.bin",
		}, engine.HooksFor(doubleBundle.Name))
		if err == nil || len(capture.captured()) != 0 {
			row.Reason = "archive extraction was not rejected before provider dispatch"
			return row
		}
		row.State = "untestable"
		row.Reason = "declares archive extraction, which the bounded binary executor refuses"
		return row
	}
	body := []byte("provider-double-binary")
	contentType := githubBinaryProviderContentType(operation.Binary.ContentTypes)
	capture := newGitHubProviderCapture(func(_ *http.Request) (int, string, []byte) { return http.StatusOK, contentType, body })
	defer capture.Close()
	doubleBundle := githubProviderDoubleBundle(b, capture.URL)
	temp := t.TempDir()
	result, err := engine.OperationBinaryDownload(context.Background(), doubleBundle, engine.BinaryDownloadRequest{
		Operation: operation.ID, Config: githubProviderDoubleConfig(doubleBundle),
		PathParams: operationPathParams(operation.Binary.Path), DestRoot: temp, FileName: "artifact.bin",
	}, engine.HooksFor(doubleBundle.Name))
	if err != nil {
		row.Reason = err.Error()
		return row
	}
	if result.Record == nil {
		row.Reason = "binary download returned no file manifest"
		return row
	}
	if len(capture.captured()) != 1 {
		row.Reason = fmt.Sprintf("binary download made %d provider-double requests, want 1", len(capture.captured()))
		return row
	}
	row.State = "exercised"
	row.RequestCount = 1
	row.Requests = githubProviderRequestProofs(capture.captured())
	row.Response = &githubProviderDoubleResponseProof{Status: http.StatusOK, Bytes: len(body), BodySHA256: hashBytes(body)}
	return row
}

func githubBinaryProviderContentType(contentTypes []string) string {
	for _, declared := range contentTypes {
		mediaType := strings.TrimSpace(strings.SplitN(declared, ";", 2)[0])
		parts := strings.Split(mediaType, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		if parts[1] == "*" {
			switch parts[0] {
			case "text":
				return "text/plain"
			case "image":
				return "image/png"
			default:
				return parts[0] + "/octet-stream"
			}
		}
		return mediaType
	}
	return "application/octet-stream"
}
