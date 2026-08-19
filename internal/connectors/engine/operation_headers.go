package engine

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

// maxOperationHeaderBytes is an absolute ceiling in addition to each
// declaration's smaller cap. It keeps a malformed bundle from converting an
// otherwise bounded fixed operation into an oversized-header transport path.
const maxOperationHeaderBytes = 16 << 10

// protectedOperationHeaderNames are credential- or transport-owned fields.
// This is deliberately lower case: HTTP field names are case-insensitive, and
// callers must not get around the boundary through a normalization variant.
var protectedOperationHeaderNames = map[string]struct{}{
	"accept": {}, "accept-charset": {}, "accept-encoding": {}, "accept-language": {},
	"authorization": {}, "connection": {}, "content-length": {}, "content-type": {},
	"cookie": {}, "expect": {}, "forwarded": {}, "host": {}, "keep-alive": {},
	"max-forwards": {}, "proxy-authorization": {}, "proxy-connection": {}, "set-cookie": {},
	"te": {}, "trailer": {}, "transfer-encoding": {}, "upgrade": {}, "user-agent": {}, "via": {},
	"x-forwarded-for": {}, "x-forwarded-host": {}, "x-forwarded-proto": {},
}

// maskedOperationResponseHeaderNames mirrors the established fixed
// credential/transport header masking used by certification capture. A value
// is never silently omitted: a declared header remains observable as
// {"redacted":true}. Ordinary provider headers, including unusual or
// paid-tier metadata, are preserved without a scope-based filter.
var maskedOperationResponseHeaderNames = map[string]struct{}{
	"authorization":       {},
	"cookie":              {},
	"proxy-authorization": {},
	"set-cookie":          {},
	"www-authenticate":    {},
	"x-api-key":           {},
	"x-auth-token":        {},
}

// validateOperationHeaderParameters admits the closed declaration shape. Only
// REST-like operation blocks have parameters; GraphQL variables remain the
// existing fixed-document contract and cannot gain a header escape hatch.
func validateOperationHeaderParameters(op OperationSpec) error {
	seen := make(map[string]struct{})
	for _, parameter := range operationParameters(op) {
		if parameter.In != "header" {
			continue
		}
		name := strings.TrimSpace(parameter.Name)
		if name == "" || name != parameter.Name || !httpHeaderNamePattern.MatchString(name) {
			return fmt.Errorf("header parameter name %q is not a valid HTTP field name", parameter.Name)
		}
		canonical := strings.ToLower(name)
		if _, protected := protectedOperationHeaderNames[canonical]; protected {
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

func validateOperationResponseContract(op OperationSpec) error {
	response := operationResponseSpec(op)
	if response == nil {
		return nil
	}
	seen := make(map[string]struct{}, len(response.Headers))
	for _, header := range response.Headers {
		name := strings.TrimSpace(header.Name)
		if name == "" || name != header.Name || !httpHeaderNamePattern.MatchString(name) {
			return fmt.Errorf("response header name %q is not a valid HTTP field name", header.Name)
		}
		canonical := strings.ToLower(name)
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
func operationRequestHeaders(b Bundle, op OperationSpec, values map[string]string) (map[string]string, error) {
	if err := validateOperationHeaderParameters(op); err != nil {
		return nil, fmt.Errorf("operation %q request headers: %w", op.ID, err)
	}
	declared := make(map[string]OperationParameter)
	for _, parameter := range operationParameters(op) {
		if parameter.In == "header" {
			declared[strings.ToLower(parameter.Name)] = parameter
		}
	}
	protected := operationRuntimeHeaderNames(b)
	provided := make(map[string]struct{}, len(values))
	resolved := make(map[string]string, len(values))
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, suppliedName := range keys {
		value := values[suppliedName]
		if strings.TrimSpace(suppliedName) != suppliedName || !httpHeaderNamePattern.MatchString(suppliedName) {
			return nil, fmt.Errorf("operation %q request header %q is malformed", op.ID, suppliedName)
		}
		name := strings.ToLower(suppliedName)
		if _, blocked := protected[name]; blocked {
			return nil, fmt.Errorf("operation %q request header %q is protected and runtime-owned", op.ID, suppliedName)
		}
		parameter, ok := declared[name]
		if !ok {
			return nil, fmt.Errorf("operation %q has unknown declared request header %q", op.ID, suppliedName)
		}
		if _, duplicate := provided[name]; duplicate {
			return nil, fmt.Errorf("operation %q supplied duplicate request header %q", op.ID, suppliedName)
		}
		if err := safety.RejectDangerousChars(value, fmt.Sprintf("operation %q request header %q", op.ID, suppliedName)); err != nil {
			return nil, err
		}
		if len(value) > parameter.MaxBytes {
			return nil, fmt.Errorf("operation %q request header %q exceeds declared byte cap %d", op.ID, suppliedName, parameter.MaxBytes)
		}
		schema, err := CompileSchema(parameter.Schema)
		if err != nil {
			return nil, fmt.Errorf("operation %q request header %q declaration: %w", op.ID, suppliedName, err)
		}
		if err := schema.Validate(value); err != nil {
			return nil, fmt.Errorf("operation %q request header %q does not satisfy declared schema: %w", op.ID, suppliedName, err)
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
				return nil, fmt.Errorf("operation %q request header %q is not one of the declared values", op.ID, suppliedName)
			}
		}
		provided[name] = struct{}{}
		resolved[parameter.Name] = value
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
// channel. Each admitted ordinary value is preserved exactly and a known
// credential/transport value is represented by an explicit redaction marker.
func operationResponseHeaders(op OperationSpec, headers http.Header) (map[string]connectors.OperationResponseHeader, error) {
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
		if _, masked := maskedOperationResponseHeaderNames[strings.ToLower(declared.Name)]; masked {
			result[declared.Name] = connectors.OperationResponseHeader{Redacted: true}
			continue
		}
		result[declared.Name] = connectors.OperationResponseHeader{Values: append([]string(nil), values...)}
	}
	return result, nil
}

func operationRuntimeHeaderNames(b Bundle) map[string]struct{} {
	protected := make(map[string]struct{}, len(protectedOperationHeaderNames)+len(b.HTTP.Headers)+len(b.HTTP.Auth))
	for name := range protectedOperationHeaderNames {
		protected[name] = struct{}{}
	}
	for name := range b.HTTP.Headers {
		protected[strings.ToLower(name)] = struct{}{}
	}
	for _, auth := range b.HTTP.Auth {
		if name := strings.TrimSpace(auth.Header); name != "" {
			protected[strings.ToLower(name)] = struct{}{}
		}
	}
	return protected
}

// requesterWithOperationHeaders returns a shallow Requester clone with the
// already-admitted declaration-owned header values. The original rate-limited
// requester is never mutated, so retries, redirects, sibling operations, and
// later calls cannot inherit a caller value.
func requesterWithOperationHeaders(requester *connsdk.Requester, headers map[string]string) *connsdk.Requester {
	if len(headers) == 0 {
		return requester
	}
	clone := *requester
	clone.DefaultHeaders = make(map[string]string, len(requester.DefaultHeaders)+len(headers))
	for key, value := range requester.DefaultHeaders {
		clone.DefaultHeaders[key] = value
	}
	for key, value := range headers {
		clone.DefaultHeaders[key] = value
	}
	return &clone
}

func cloneOperationHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	clone := make(map[string]string, len(headers))
	for key, value := range headers {
		clone[key] = value
	}
	return clone
}
