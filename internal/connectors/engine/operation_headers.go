package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

// maxOperationHeaderBytes is an absolute ceiling in addition to each
// declaration's smaller cap. It keeps a malformed bundle from converting an
// otherwise bounded fixed operation into an oversized-header transport path.
const maxOperationHeaderBytes = 16 << 10

// validateOperationHeaderParameters admits the closed declaration shape. Only
// REST-like operation blocks have parameters; GraphQL variables remain the
// existing fixed-document contract and cannot gain a header escape hatch.
func validateOperationHeaderParameters(op OperationSpec) error {
	if op.REST != nil {
		for _, parameter := range op.REST.PaginationParameters {
			if parameter.Repeatable {
				return fmt.Errorf("pagination parameter %q cannot be repeatable", parameter.Name)
			}
		}
	}
	seen := make(map[string]struct{})
	for _, parameter := range operationParameters(op) {
		location := strings.ToLower(strings.TrimSpace(parameter.In))
		if err := validateOperationParameterCLIName(parameter, location); err != nil {
			return err
		}
		switch location {
		case "path", "query":
			if err := safety.ValidateIdentifier(parameter.Name, "operation "+location+" parameter"); err != nil {
				return err
			}
			if parameter.MaxBytes < 0 || parameter.MaxBytes > maxOperationParameterMaxBytes {
				return fmt.Errorf("%s parameter %q max_bytes must be omitted or between 1 and %d", location, parameter.Name, maxOperationParameterMaxBytes)
			}
		case "header":
		default:
			return fmt.Errorf("parameter %q has unsupported location %q", parameter.Name, parameter.In)
		}
		if parameter.Repeatable && parameter.In != "header" {
			return fmt.Errorf("parameter %q repeatable is supported only for request headers", parameter.Name)
		}
		if parameter.In != "header" {
			continue
		}
		name := strings.TrimSpace(parameter.Name)
		if name == "" || name != parameter.Name || !httpHeaderNamePattern.MatchString(name) {
			return fmt.Errorf("header parameter name %q is not a valid HTTP field name", parameter.Name)
		}
		canonical, err := connectors.CanonicalOperationHeaderName(name)
		if err != nil {
			return fmt.Errorf("header parameter name %q is not a valid HTTP field name", parameter.Name)
		}
		if connectors.IsProtectedOperationHeaderName(canonical) {
			return fmt.Errorf("header parameter %q is protected and runtime-owned", parameter.Name)
		}
		if _, duplicate := seen[canonical]; duplicate {
			return fmt.Errorf("header parameter %q duplicates another header ignoring case", parameter.Name)
		}
		seen[canonical] = struct{}{}
		if parameter.Type != "" && parameter.Type != "string" {
			return fmt.Errorf("header parameter %q type must be string", parameter.Name)
		}
		if len(parameter.Schema) == 0 {
			return fmt.Errorf("header parameter %q requires a bounded string schema", parameter.Name)
		}
		if parameter.MaxBytes <= 0 || parameter.MaxBytes > maxOperationHeaderBytes {
			return fmt.Errorf("header parameter %q max_bytes must be between 1 and %d", parameter.Name, maxOperationHeaderBytes)
		}
		if !operationHeaderSchemaIsString(parameter.Schema) {
			return fmt.Errorf("header parameter %q schema type must be string", parameter.Name)
		}
		if _, err := CompileSchema(parameter.Schema); err != nil {
			return fmt.Errorf("header parameter %q schema: %w", parameter.Name, err)
		}
	}
	return nil
}

func validateOperationParameterCLIName(parameter OperationParameter, location string) error {
	cliName := strings.TrimSpace(parameter.CLIName)
	if cliName == "" {
		return nil
	}
	if cliName != parameter.CLIName || location != "path" {
		return fmt.Errorf("parameter %q cli_name is supported only as an exact path flag name", parameter.Name)
	}
	if err := safety.ValidateIdentifier(cliName, "path parameter cli_name"); err != nil {
		return err
	}
	if cliName != strings.ToLower(cliName) || strings.Contains(cliName, ".") || strings.HasPrefix(cliName, "-") || strings.HasSuffix(cliName, "-") || strings.Contains(cliName, "--") {
		return fmt.Errorf("path parameter cli_name %q must be a lowercase hyphenated flag name", parameter.CLIName)
	}
	return nil
}

func validateOperationResponseContract(op OperationSpec) error {
	response := operationResponseSpec(op)
	if response == nil {
		return nil
	}
	for _, declared := range response.SuccessStatuses {
		if _, err := parseOperationSuccessStatus(declared); err != nil {
			return fmt.Errorf("response success status %q: %w", declared, err)
		}
	}
	seen := make(map[string]struct{}, len(response.Headers))
	for _, header := range response.Headers {
		name := strings.TrimSpace(header.Name)
		if name == "" || name != header.Name || !httpHeaderNamePattern.MatchString(name) {
			return fmt.Errorf("response header name %q is not a valid HTTP field name", header.Name)
		}
		canonical, err := connectors.CanonicalOperationHeaderName(name)
		if err != nil {
			return fmt.Errorf("response header name %q is not a valid HTTP field name", header.Name)
		}
		if _, duplicate := seen[canonical]; duplicate {
			return fmt.Errorf("response header %q duplicates another header ignoring case", header.Name)
		}
		seen[canonical] = struct{}{}
		if header.MaxBytes <= 0 || header.MaxBytes > maxOperationHeaderBytes {
			return fmt.Errorf("response header %q max_bytes must be between 1 and %d", header.Name, maxOperationHeaderBytes)
		}
	}
	return nil
}

func requireOperationSuccessStatusPolicy(op OperationSpec) ([]connsdk.StatusRange, error) {
	response := operationResponseSpec(op)
	if response == nil || len(response.SuccessStatuses) == 0 {
		return nil, fmt.Errorf("operation %q requires declared response success_statuses", op.ID)
	}
	ranges := make([]connsdk.StatusRange, 0, len(response.SuccessStatuses))
	for _, declared := range response.SuccessStatuses {
		status, err := parseOperationSuccessStatus(declared)
		if err != nil {
			return nil, fmt.Errorf("operation %q response success status %q: %w", op.ID, declared, err)
		}
		ranges = append(ranges, status)
	}
	return ranges, nil
}

func operationSuccessStatusRanges(op OperationSpec) ([]connsdk.StatusRange, error) {
	response := operationResponseSpec(op)
	if response == nil || len(response.SuccessStatuses) == 0 {
		return nil, nil
	}
	return requireOperationSuccessStatusPolicy(op)
}

func parseOperationSuccessStatus(raw string) (connsdk.StatusRange, error) {
	text := strings.TrimSpace(raw)
	if text != raw || text == "" {
		return connsdk.StatusRange{}, fmt.Errorf("must be an exact 2xx status or inclusive 2xx range")
	}
	if len(text) == 3 {
		status, err := strconv.Atoi(text)
		if err != nil || status < 200 || status > 299 {
			return connsdk.StatusRange{}, fmt.Errorf("must be an exact 2xx status or inclusive 2xx range")
		}
		return connsdk.StatusRange{Min: status, Max: status}, nil
	}
	start, end, ok := strings.Cut(text, "-")
	if !ok || len(start) != 3 || len(end) != 3 {
		return connsdk.StatusRange{}, fmt.Errorf("must be an exact 2xx status or inclusive 2xx range")
	}
	min, minErr := strconv.Atoi(start)
	max, maxErr := strconv.Atoi(end)
	if minErr != nil || maxErr != nil || min < 200 || max > 299 || min > max {
		return connsdk.StatusRange{}, fmt.Errorf("must be an exact 2xx status or inclusive 2xx range")
	}
	return connsdk.StatusRange{Min: min, Max: max}, nil
}

func operationHeaderSchemaIsString(raw json.RawMessage) bool {
	var schema struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return false
	}
	return schema.Type == "string"
}

// operationRequestHeaders validates one operation's caller-supplied headers
// before credentials/runtime construction. It returns canonical declaration
// names only, which prevents unknown, duplicate, cross-operation, and
// normalization variants from ever reaching connsdk.
func operationRequestHeaders(b Bundle, op OperationSpec, values map[string]string, repeated map[string][]string) (http.Header, error) {
	if err := validateOperationHeaderParameters(op); err != nil {
		return nil, fmt.Errorf("operation %q request headers: %w", op.ID, err)
	}
	declared := make(map[string]OperationParameter)
	for _, parameter := range operationParameters(op) {
		if parameter.In == "header" {
			canonical, err := connectors.CanonicalOperationHeaderName(parameter.Name)
			if err != nil {
				return nil, fmt.Errorf("operation %q request header %q declaration: %w", op.ID, parameter.Name, err)
			}
			declared[canonical] = parameter
		}
	}
	protected := operationRuntimeHeaderNames(b)
	provided := make(map[string]struct{}, len(values)+len(repeated))
	resolved := make(http.Header, len(values)+len(repeated))
	resolve := func(suppliedName string, suppliedValues []string) error {
		if strings.TrimSpace(suppliedName) != suppliedName || !httpHeaderNamePattern.MatchString(suppliedName) {
			return fmt.Errorf("operation %q request header %q is malformed", op.ID, suppliedName)
		}
		name, err := connectors.CanonicalOperationHeaderName(suppliedName)
		if err != nil {
			return fmt.Errorf("operation %q request header %q is malformed", op.ID, suppliedName)
		}
		if connectors.IsProtectedOperationHeaderName(name) {
			return fmt.Errorf("operation %q request header %q is protected and runtime-owned", op.ID, suppliedName)
		}
		if _, blocked := protected[name]; blocked {
			return fmt.Errorf("operation %q request header %q is protected and runtime-owned", op.ID, suppliedName)
		}
		parameter, ok := declared[name]
		if !ok {
			return fmt.Errorf("operation %q has unknown declared request header %q", op.ID, suppliedName)
		}
		if _, duplicate := provided[name]; duplicate {
			return fmt.Errorf("operation %q supplied duplicate request header %q", op.ID, suppliedName)
		}
		if len(suppliedValues) == 0 {
			return fmt.Errorf("operation %q request header %q has no values", op.ID, suppliedName)
		}
		if len(suppliedValues) > 1 && !parameter.Repeatable {
			return fmt.Errorf("operation %q request header %q accepts exactly one value", op.ID, suppliedName)
		}
		schema, err := CompileSchema(parameter.Schema)
		if err != nil {
			return fmt.Errorf("operation %q request header %q declaration: %w", op.ID, suppliedName, err)
		}
		totalBytes := 0
		for _, value := range suppliedValues {
			if err := safety.RejectDangerousChars(value, fmt.Sprintf("operation %q request header %q", op.ID, suppliedName)); err != nil {
				return err
			}
			totalBytes += len(value)
			if totalBytes > parameter.MaxBytes {
				return fmt.Errorf("operation %q request header %q exceeds declared byte cap %d", op.ID, suppliedName, parameter.MaxBytes)
			}
			if err := schema.Validate(value); err != nil {
				return fmt.Errorf("operation %q request header %q does not satisfy declared schema: %w", op.ID, suppliedName, err)
			}
			if len(parameter.Values) > 0 {
				matched := false
				for _, allowed := range parameter.Values {
					if value == allowed {
						matched = true
						break
					}
				}
				if !matched {
					return fmt.Errorf("operation %q request header %q is not one of the declared values", op.ID, suppliedName)
				}
			}
		}
		provided[name] = struct{}{}
		resolved[http.CanonicalHeaderKey(parameter.Name)] = append([]string(nil), suppliedValues...)
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, suppliedName := range keys {
		if err := resolve(suppliedName, []string{values[suppliedName]}); err != nil {
			return nil, err
		}
	}
	repeatedKeys := make([]string, 0, len(repeated))
	for key := range repeated {
		repeatedKeys = append(repeatedKeys, key)
	}
	sort.Strings(repeatedKeys)
	for _, suppliedName := range repeatedKeys {
		if err := resolve(suppliedName, repeated[suppliedName]); err != nil {
			return nil, err
		}
	}
	for name, parameter := range declared {
		if parameter.Required {
			if _, ok := provided[name]; !ok {
				return nil, fmt.Errorf("operation %q requires declared request header %q", op.ID, parameter.Name)
			}
		}
	}
	return resolved, nil
}

func operationParameters(op OperationSpec) []OperationParameter {
	if op.REST != nil {
		return op.REST.Parameters
	}
	if op.Binary != nil {
		return op.Binary.Parameters
	}
	return nil
}

func operationResponseSpec(op OperationSpec) *OperationResponseSpec {
	if op.REST != nil {
		return op.REST.Response
	}
	if op.Binary != nil {
		return op.Binary.Response
	}
	return nil
}

// operationResponseHeaders materializes the declaration's bounded metadata
// projection. It intentionally iterates the declaration, not the provider
// map, so an endpoint cannot turn response metadata into an arbitrary output
// channel. Each admitted provider header value is preserved verbatim, even
// when it equals configured credential bytes. Declared output secret fields
// are sanitized at the public projection boundary.
func operationResponseHeaders(_ Bundle, op OperationSpec, headers http.Header) (map[string]connectors.OperationResponseHeader, error) {
	if err := validateOperationResponseContract(op); err != nil {
		return nil, fmt.Errorf("operation %q response headers: %w", op.ID, err)
	}
	response := operationResponseSpec(op)
	if response == nil || len(response.Headers) == 0 {
		return nil, nil
	}
	result := make(map[string]connectors.OperationResponseHeader, len(response.Headers))
	for _, declared := range response.Headers {
		values, present := headers[http.CanonicalHeaderKey(declared.Name)]
		if !present {
			// http.Header normally canonicalizes keys, but provider transports
			// are not required to build this map through Header.Set.
			for name, candidate := range headers {
				if strings.EqualFold(name, declared.Name) {
					values, present = candidate, true
					break
				}
			}
		}
		if !present {
			continue
		}
		total := 0
		for _, value := range values {
			total += len(value)
		}
		if total > declared.MaxBytes {
			return nil, fmt.Errorf("operation %q response header %q exceeds declared byte cap %d", op.ID, declared.Name, declared.MaxBytes)
		}
		result[declared.Name] = connectors.OperationResponseHeader{Values: append([]string(nil), values...)}
	}
	return result, nil
}

func operationRuntimeHeaderNames(b Bundle) map[string]struct{} {
	return operationRuntimeHeaderNamesForBase(b.HTTP)
}

func operationRuntimeHeaderNamesForBase(base HTTPBase) map[string]struct{} {
	protected := operationRuntimeAuthHeaderNames(base)
	for name := range base.Headers {
		if canonical, err := connectors.CanonicalOperationHeaderName(name); err == nil {
			protected[canonical] = struct{}{}
		}
	}
	return protected
}

func operationRuntimeAuthHeaderNames(base HTTPBase) map[string]struct{} {
	protected := make(map[string]struct{}, len(base.Auth))
	for _, auth := range base.Auth {
		if name := strings.TrimSpace(auth.Header); name != "" {
			if canonical, err := connectors.CanonicalOperationHeaderName(name); err == nil {
				protected[canonical] = struct{}{}
			}
		}
	}
	return protected
}

func validateOperationRuntimeHeaderIsolation(base HTTPBase, operations []OperationSpec) error {
	runtimeHeaders := operationRuntimeHeaderNamesForBase(base)
	if len(runtimeHeaders) == 0 {
		return nil
	}
	for _, operation := range operations {
		for _, parameter := range operationParameters(operation) {
			if parameter.In != "header" {
				continue
			}
			canonical, err := connectors.CanonicalOperationHeaderName(parameter.Name)
			if err != nil {
				continue
			}
			if _, protected := runtimeHeaders[canonical]; protected {
				return fmt.Errorf("operation %q header parameter %q is protected and runtime-owned", operation.ID, parameter.Name)
			}
		}
	}
	return nil
}

// requesterWithOperationHeaders returns a shallow Requester clone with the
// already-admitted declaration-owned header values. The original rate-limited
// requester is never mutated, so retries, redirects, sibling operations, and
// later calls cannot inherit a caller value.
func requesterWithOperationHeaders(requester *connsdk.Requester, op OperationSpec, headers http.Header) (*connsdk.Requester, error) {
	statuses, err := operationSuccessStatusRanges(op)
	if err != nil {
		return nil, err
	}
	redirect, err := operationRedirectPolicy(op)
	if err != nil {
		return nil, err
	}
	clone := *requester
	clone.DefaultHeaders = make(map[string]string, len(requester.DefaultHeaders))
	for key, value := range requester.DefaultHeaders {
		clone.DefaultHeaders[key] = value
	}
	clone.DefaultHeaderValues = cloneOperationHeaders(requester.DefaultHeaderValues)
	if clone.DefaultHeaderValues == nil {
		clone.DefaultHeaderValues = make(http.Header, len(headers))
	}
	for key, values := range headers {
		clone.DefaultHeaderValues[key] = append([]string(nil), values...)
	}
	clone.AcceptedStatuses = statuses
	clone.RedirectPolicy = redirect
	return &clone, nil
}

func operationRedirectPolicy(op OperationSpec) (*connsdk.RedirectPolicy, error) {
	var declared *OperationRedirectSpec
	if op.REST != nil {
		declared = op.REST.Redirect
	}
	if op.Binary != nil {
		if op.Binary.AllowCrossHost {
			return nil, fmt.Errorf("operation %q uses unbounded legacy cross-host redirect metadata", op.ID)
		}
		declared = op.Binary.Redirect
		if declared == nil && len(op.Binary.AllowedHosts) != 0 {
			declared = &OperationRedirectSpec{MaxHops: 1, AllowedHosts: append([]string(nil), op.Binary.AllowedHosts...)}
		}
	}
	if declared == nil {
		return &connsdk.RedirectPolicy{}, nil
	}
	if declared.MaxHops < 1 || declared.MaxHops > 10 {
		return nil, fmt.Errorf("operation %q redirect max_hops must be between 1 and 10", op.ID)
	}
	allowed := make([]string, 0, len(declared.AllowedHosts))
	seen := map[string]struct{}{}
	for _, raw := range declared.AllowedHosts {
		host := strings.ToLower(strings.TrimSpace(raw))
		if host == "" || strings.ContainsAny(host, "/?#@") {
			return nil, fmt.Errorf("operation %q redirect allowed host %q is invalid", op.ID, raw)
		}
		if _, duplicate := seen[host]; duplicate {
			return nil, fmt.Errorf("operation %q redirect allowed host %q is duplicated", op.ID, raw)
		}
		seen[host] = struct{}{}
		allowed = append(allowed, host)
	}
	if !declared.AllowSameOrigin && len(allowed) == 0 {
		return nil, fmt.Errorf("operation %q redirect policy permits no redirect target", op.ID)
	}
	return &connsdk.RedirectPolicy{MaxHops: declared.MaxHops, AllowSameOrigin: declared.AllowSameOrigin, AllowedHosts: allowed}, nil
}

func cloneOperationHeaders(headers http.Header) http.Header {
	if len(headers) == 0 {
		return nil
	}
	clone := make(http.Header, len(headers))
	for key, values := range headers {
		clone[key] = append([]string(nil), values...)
	}
	return clone
}

func operationSingleHeaders(headers http.Header) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	result := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) == 1 {
			result[key] = values[0]
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func operationRepeatedHeaders(headers http.Header) map[string][]string {
	var result map[string][]string
	for key, values := range headers {
		if len(values) <= 1 {
			continue
		}
		if result == nil {
			result = make(map[string][]string)
		}
		result[key] = append([]string(nil), values...)
	}
	return result
}
