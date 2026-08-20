package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

const (
	graphQLCursorStrategy      = "graphql_cursor"
	maxGraphQLOperationErrors  = 3
	maxGraphQLErrorMessageSize = 512
)

// validateGraphQLOperationDeclaration checks the fields every GraphQL
// operation has always promised: a fixed, named document of the declared kind.
// The executable-only endpoint and variable fields are deliberately checked
// later, at operation preflight, so legacy stream bindings stay metadata-only
// until their generated command has a complete typed contract.
func validateGraphQLOperationDeclaration(op OperationSpec, kind string) error {
	if op.GraphQL == nil {
		return fmt.Errorf("graphql is required")
	}
	return validateGraphQLSpec(&GraphQLRequestSpec{
		Document:      op.GraphQL.Document,
		OperationName: op.GraphQL.OperationName,
	}, kind)
}

// validateGraphQLOperationDirectContract is the common no-network admission
// check for fixed GraphQL queries and mutations. It has intentionally no
// caller-provided document, endpoint, selection, or headers: all four facts
// are declaration-owned and are therefore preview-bindable.
func validateGraphQLOperationDirectContract(op OperationSpec, kind string) error {
	if op.Kind != "graphql_"+kind {
		return fmt.Errorf("fixed GraphQL %s requires graphql_%s operation, got %q", kind, kind, op.Kind)
	}
	if err := validateGraphQLOperationDeclaration(op, kind); err != nil {
		return err
	}
	if err := validateSingleNamedFixedGraphQLOperation(op.GraphQL.Document, kind, op.GraphQL.OperationName); err != nil {
		return err
	}
	if err := validateFixedGraphQLOperationPath(op.GraphQL.Path); err != nil {
		return err
	}
	if op.GraphQL.MaxBytes <= 0 {
		return fmt.Errorf("fixed GraphQL operation requires positive max_bytes")
	}
	if len(op.GraphQL.VariablesSchema) == 0 {
		return fmt.Errorf("fixed GraphQL operation requires variables_schema")
	}
	variablesSchema, variablesNode, err := graphQLOperationVariablesSchema(op)
	if err != nil {
		return err
	}
	_ = variablesSchema // compilation is part of the admission check above.
	if err := validateGraphQLOperationPagination(op, kind, variablesNode); err != nil {
		return err
	}
	return nil
}

type graphQLFixedOperationDefinition struct {
	kind string
	name string
}

// validateSingleNamedFixedGraphQLOperation makes the fixed-document promise
// executable rather than merely descriptive. validateGraphQLSpec deliberately
// supports legacy metadata bindings and only checks a document prefix; a
// direct_read must additionally prove that operationName cannot select a
// second mutation embedded after an otherwise valid query.
func validateSingleNamedFixedGraphQLOperation(document, wantKind, wantName string) error {
	definitions, err := graphQLFixedOperationDefinitions(document)
	if err != nil {
		return fmt.Errorf("fixed GraphQL document: %w", err)
	}
	if len(definitions) != 1 || definitions[0].kind != wantKind || definitions[0].name != wantName {
		return fmt.Errorf("fixed GraphQL document must contain exactly one named %s operation %q", wantKind, wantName)
	}
	return nil
}

// graphQLFixedOperationDefinitions is a deliberately small lexical scanner,
// not a general GraphQL parser. It only recognizes top-level executable
// definitions, skips comments and string literals, and consumes each balanced
// selection before looking for another definition. That is enough to reject a
// query document whose operationName could select an appended mutation while
// retaining fixed fragments and ordinary argument/default-value syntax.
func graphQLFixedOperationDefinitions(document string) ([]graphQLFixedOperationDefinition, error) {
	var definitions []graphQLFixedOperationDefinition
	for offset := 0; ; {
		var err error
		offset, err = skipGraphQLIgnored(document, offset)
		if err != nil {
			return nil, err
		}
		if offset == len(document) {
			return definitions, nil
		}
		keyword, next, ok := readGraphQLName(document, offset)
		if !ok {
			return nil, fmt.Errorf("expected a top-level definition at byte %d", offset)
		}
		offset, err = skipGraphQLIgnored(document, next)
		if err != nil {
			return nil, err
		}
		switch keyword {
		case "query", "mutation", "subscription":
			name, afterName, named := readGraphQLName(document, offset)
			if !named {
				return nil, fmt.Errorf("%s operation must be named", keyword)
			}
			offset, err = skipGraphQLDefinitionSelection(document, afterName)
			if err != nil {
				return nil, fmt.Errorf("%s operation %q: %w", keyword, name, err)
			}
			definitions = append(definitions, graphQLFixedOperationDefinition{kind: keyword, name: name})
		case "fragment":
			offset, err = skipGraphQLDefinitionSelection(document, offset)
			if err != nil {
				return nil, fmt.Errorf("fragment definition: %w", err)
			}
		default:
			return nil, fmt.Errorf("unsupported top-level definition %q", keyword)
		}
	}
}

func skipGraphQLDefinitionSelection(document string, offset int) (int, error) {
	parenDepth, bracketDepth := 0, 0
	for offset < len(document) {
		var err error
		offset, err = skipGraphQLIgnored(document, offset)
		if err != nil {
			return 0, err
		}
		if offset == len(document) {
			break
		}
		switch document[offset] {
		case '"':
			offset, err = skipGraphQLString(document, offset)
			if err != nil {
				return 0, err
			}
			continue
		case '(':
			parenDepth++
		case ')':
			parenDepth--
			if parenDepth < 0 {
				return 0, fmt.Errorf("unmatched closing parenthesis")
			}
		case '[':
			bracketDepth++
		case ']':
			bracketDepth--
			if bracketDepth < 0 {
				return 0, fmt.Errorf("unmatched closing bracket")
			}
		case '{':
			if parenDepth == 0 && bracketDepth == 0 {
				return skipGraphQLBalancedBraces(document, offset)
			}
			// Input object literals can appear in a variable default or directive
			// argument. They are not the operation selection and must be skipped
			// as a complete unit before continuing the header scan.
			offset, err = skipGraphQLBalancedBraces(document, offset)
			if err != nil {
				return 0, err
			}
			continue
		case '}':
			return 0, fmt.Errorf("unexpected closing brace before selection")
		}
		offset++
	}
	return 0, fmt.Errorf("has no selection set")
}

func skipGraphQLBalancedBraces(document string, offset int) (int, error) {
	if offset >= len(document) || document[offset] != '{' {
		return 0, fmt.Errorf("expected opening brace")
	}
	depth := 0
	for offset < len(document) {
		switch document[offset] {
		case '"':
			next, err := skipGraphQLString(document, offset)
			if err != nil {
				return 0, err
			}
			offset = next
			continue
		case '#':
			offset = skipGraphQLComment(document, offset)
			continue
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return offset + 1, nil
			}
		}
		offset++
	}
	return 0, fmt.Errorf("unclosed selection set")
}

func skipGraphQLIgnored(document string, offset int) (int, error) {
	for offset < len(document) {
		switch document[offset] {
		case ' ', '\t', '\n', '\r', ',':
			offset++
		case '#':
			offset = skipGraphQLComment(document, offset)
		default:
			return offset, nil
		}
	}
	return offset, nil
}

func skipGraphQLComment(document string, offset int) int {
	for offset < len(document) && document[offset] != '\n' && document[offset] != '\r' {
		offset++
	}
	return offset
}

func skipGraphQLString(document string, offset int) (int, error) {
	if offset >= len(document) || document[offset] != '"' {
		return 0, fmt.Errorf("expected opening string quote")
	}
	if strings.HasPrefix(document[offset:], `"""`) {
		offset += 3
		for offset+2 < len(document) {
			if document[offset] == '\\' {
				offset += 2
				continue
			}
			if strings.HasPrefix(document[offset:], `"""`) {
				return offset + 3, nil
			}
			offset++
		}
		return 0, fmt.Errorf("unterminated block string")
	}
	offset++
	for offset < len(document) {
		switch document[offset] {
		case '\\':
			offset += 2
		case '"':
			return offset + 1, nil
		case '\n', '\r':
			return 0, fmt.Errorf("unterminated string")
		default:
			offset++
		}
	}
	return 0, fmt.Errorf("unterminated string")
}

func readGraphQLName(document string, offset int) (string, int, bool) {
	if offset >= len(document) || !isGraphQLNameStart(document[offset]) {
		return "", offset, false
	}
	end := offset + 1
	for end < len(document) && isGraphQLNameContinue(document[end]) {
		end++
	}
	return document[offset:end], end, true
}

func isGraphQLNameStart(value byte) bool {
	return (value >= 'A' && value <= 'Z') || (value >= 'a' && value <= 'z') || value == '_'
}

func isGraphQLNameContinue(value byte) bool {
	return isGraphQLNameStart(value) || (value >= '0' && value <= '9')
}

func validateFixedGraphQLOperationPath(raw string) error {
	path := strings.TrimSpace(raw)
	if path == "" || path != raw || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || isAbsoluteHTTPURL(path) {
		return fmt.Errorf("fixed GraphQL operation requires a connector-relative path")
	}
	if err := safety.RejectDangerousChars(path, "fixed GraphQL operation path"); err != nil {
		return err
	}
	// Fixed GraphQL endpoints never need a query string, fragment, encoded
	// separator, or path traversal. Refusing all of them keeps the declared
	// api_surface path byte-for-byte equal to the request target the preview
	// binds, rather than relying on URL normalization after admission.
	if strings.ContainsAny(path, "?#\\") || strings.Contains(path, "%") {
		return fmt.Errorf("fixed GraphQL operation path must not contain query, fragment, escape, or backslash characters")
	}
	for _, segment := range strings.Split(strings.TrimPrefix(path, "/"), "/") {
		if segment == "" || segment == "." || segment == ".." {
			return fmt.Errorf("fixed GraphQL operation path must not contain empty or traversal segments")
		}
	}
	return nil
}

// graphQLOperationVariablesSchema compiles the closed input object used by a
// fixed GraphQL operation. A closed, recursively bounded schema is required
// because GraphQL variables are otherwise an easy way to smuggle a raw query's
// worth of arbitrary nested input through a fixed document.
func graphQLOperationVariablesSchema(op OperationSpec) (*Schema, map[string]any, error) {
	var node map[string]any
	if err := json.Unmarshal(op.GraphQL.VariablesSchema, &node); err != nil {
		return nil, nil, fmt.Errorf("operation %q graphql.variables_schema is not an object: %w", op.ID, err)
	}
	if !isObjectType(node) {
		return nil, nil, fmt.Errorf("operation %q graphql.variables_schema must be an object", op.ID)
	}
	if closed, ok := node["additionalProperties"].(bool); !ok || closed {
		return nil, nil, fmt.Errorf("operation %q graphql.variables_schema must declare additionalProperties: false", op.ID)
	}
	sch, err := CompileSchema(op.GraphQL.VariablesSchema)
	if err != nil {
		return nil, nil, fmt.Errorf("operation %q graphql.variables_schema: %w", op.ID, err)
	}
	if err := requireClosedBoundedGraphQLVariables(op.ID, node, "variables_schema"); err != nil {
		return nil, nil, err
	}
	if err := requireGraphQLVariablesSchemaReferences(op.ID, op.GraphQL.Document, node); err != nil {
		return nil, nil, err
	}
	return sch, node, nil
}

// ValidateGraphQLOperationStructuredJSONVariable is the declaration-owned
// admission check for the one place a connector command may parse structured
// JSON for a fixed GraphQL operation. It deliberately proves only a named,
// top-level closed object/array variable from the operation's own schema; it
// is not a generic GraphQL request-body validator.
//
// Commandrunner and connectorgen both call this exact function. That keeps a
// hand-authored `json` CLI flag from becoming executable merely because one
// layer remembered the GraphQL contract while the other did not.
func ValidateGraphQLOperationStructuredJSONVariable(op OperationSpec, variable string) error {
	kind := strings.TrimPrefix(op.Kind, "graphql_")
	if kind != "query" && kind != "mutation" {
		return fmt.Errorf("operation %q structured JSON variables require a fixed GraphQL operation, got %q", op.ID, op.Kind)
	}
	if err := validateGraphQLOperationDirectContract(op, kind); err != nil {
		return fmt.Errorf("operation %q structured JSON variable contract: %w", op.ID, err)
	}
	variable = strings.TrimSpace(variable)
	if !graphQLNamePattern.MatchString(variable) {
		return fmt.Errorf("operation %q structured JSON variable %q must be a top-level GraphQL variable", op.ID, variable)
	}
	_, schema, err := graphQLOperationVariablesSchema(op)
	if err != nil {
		return err
	}
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("operation %q structured JSON variable %q is not declared", op.ID, variable)
	}
	raw, ok := properties[variable]
	if !ok {
		return fmt.Errorf("operation %q structured JSON variable %q is not declared", op.ID, variable)
	}
	node, ok := raw.(map[string]any)
	if !ok || (!isObjectType(node) && !isArrayType(node)) {
		return fmt.Errorf("operation %q structured JSON variable %q must be a closed object or array", op.ID, variable)
	}
	if err := requireClosedBoundedGraphQLVariables(op.ID, node, "variables_schema/"+variable); err != nil {
		return err
	}
	return nil
}

// requireGraphQLVariablesSchemaReferences makes the typed variables schema a
// contract for this fixed document, rather than merely a general JSON object
// carried alongside it. An unreferenced property cannot change a selection or
// endpoint, but accepting it creates a false impression that arbitrary caller
// input is part of the declared operation and postpones a local error until a
// provider request.
func requireGraphQLVariablesSchemaReferences(operation, document string, node map[string]any) error {
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return nil
	}
	references := graphQLDocumentVariableReferences(document)
	for _, name := range sortedMapKeys(properties) {
		if _, ok := references[name]; !ok {
			return fmt.Errorf("operation %q graphql.variables_schema property %q is not referenced by its fixed GraphQL document", operation, name)
		}
	}
	return nil
}

func graphQLDocumentVariableReferences(document string) map[string]struct{} {
	references := make(map[string]struct{})
	for offset := 0; offset < len(document); offset++ {
		if document[offset] != '$' || offset+1 == len(document) {
			continue
		}
		end := offset + 1
		for end < len(document) {
			c := document[end]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || (c >= '0' && c <= '9' && end > offset+1) {
				end++
				continue
			}
			break
		}
		if end == offset+1 {
			continue
		}
		name := document[offset+1 : end]
		if graphQLNamePattern.MatchString(name) {
			references[name] = struct{}{}
		}
		offset = end - 1
	}
	return references
}

func requireClosedBoundedGraphQLVariables(operation string, node map[string]any, path string) error {
	if isArrayType(node) {
		if _, ok := node["maxItems"]; !ok {
			return fmt.Errorf("operation %q graphql.%s declares an array without maxItems", operation, path)
		}
	}
	if isObjectType(node) {
		if closed, ok := node["additionalProperties"].(bool); !ok || closed {
			return fmt.Errorf("operation %q graphql.%s is an object and must declare additionalProperties: false", operation, path)
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		if err := requireClosedBoundedGraphQLVariables(operation, items, path+"/items"); err != nil {
			return err
		}
	}
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return nil
	}
	for _, name := range sortedMapKeys(properties) {
		child, ok := properties[name].(map[string]any)
		if !ok {
			continue
		}
		if err := requireClosedBoundedGraphQLVariables(operation, child, path+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

func sortedMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	// Small local insertion sort avoids creating another dependency for the
	// deterministic error ordering this recursive validator promises.
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}

func validateGraphQLOperationPagination(op OperationSpec, kind string, variables map[string]any) error {
	pagination := op.GraphQL.Pagination
	if pagination == nil {
		return nil
	}
	if kind != "query" {
		return fmt.Errorf("operation %q graphql pagination is only valid for graphql_query", op.ID)
	}
	if err := validateGraphQLFieldPath(pagination.ConnectionPath, "graphql.pagination.connection_path"); err != nil {
		return fmt.Errorf("operation %q: %w", op.ID, err)
	}
	if !graphQLNamePattern.MatchString(pagination.CursorVariable) {
		return fmt.Errorf("operation %q graphql.pagination.cursor_variable %q is not a valid GraphQL name", op.ID, pagination.CursorVariable)
	}
	if !graphQLVariablesSchemaHasProperty(variables, pagination.CursorVariable) {
		return fmt.Errorf("operation %q graphql.pagination.cursor_variable %q is absent from variables_schema", op.ID, pagination.CursorVariable)
	}
	if pagination.PageSizeVariable == "" && pagination.MaxPageSize != 0 {
		return fmt.Errorf("operation %q graphql.pagination.max_page_size requires page_size_variable", op.ID)
	}
	if pagination.PageSizeVariable == "" {
		return nil
	}
	if !graphQLNamePattern.MatchString(pagination.PageSizeVariable) {
		return fmt.Errorf("operation %q graphql.pagination.page_size_variable %q is not a valid GraphQL name", op.ID, pagination.PageSizeVariable)
	}
	if pagination.MaxPageSize <= 0 {
		return fmt.Errorf("operation %q graphql.pagination.page_size_variable requires positive max_page_size", op.ID)
	}
	if !graphQLVariablesSchemaHasProperty(variables, pagination.PageSizeVariable) {
		return fmt.Errorf("operation %q graphql.pagination.page_size_variable %q is absent from variables_schema", op.ID, pagination.PageSizeVariable)
	}
	return nil
}

func graphQLVariablesSchemaHasProperty(root map[string]any, name string) bool {
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return false
	}
	_, ok = properties[name]
	return ok
}

func validateGraphQLFieldPath(path, field string) error {
	if strings.TrimSpace(path) == "" || strings.TrimSpace(path) != path {
		return fmt.Errorf("%s is required", field)
	}
	for _, segment := range strings.Split(path, ".") {
		if !graphQLNamePattern.MatchString(segment) {
			return fmt.Errorf("%s has invalid GraphQL field %q", field, segment)
		}
	}
	return nil
}

// graphQLOperationVariables validates exactly the closed caller input object,
// then inserts the only opaque navigation value the direct-read contract
// supports. A caller may never smuggle a cursor through Body: `page-cursor` is
// the one navigation channel and is reflected in the result's NextCursor.
func graphQLOperationVariables(op OperationSpec, supplied map[string]any, page int, pageCursor string) (map[string]any, error) {
	if page != 0 {
		return nil, fmt.Errorf("operation %q GraphQL pagination accepts --page-cursor, not --page", op.ID)
	}
	variables := cloneAnyMap(supplied)
	pagination := op.GraphQL.Pagination
	if pagination == nil {
		if pageCursor != "" {
			return nil, fmt.Errorf("operation %q has no declared GraphQL cursor pagination", op.ID)
		}
	} else {
		if _, suppliedCursor := variables[pagination.CursorVariable]; suppliedCursor {
			return nil, fmt.Errorf("operation %q GraphQL cursor variable %q must be supplied with --page-cursor", op.ID, pagination.CursorVariable)
		}
		if pageCursor != "" {
			if err := safety.RejectDangerousChars(pageCursor, "page cursor"); err != nil {
				return nil, err
			}
			variables[pagination.CursorVariable] = pageCursor
		}
	}
	if pagination != nil && pagination.PageSizeVariable != "" {
		if value, ok := variables[pagination.PageSizeVariable]; ok {
			n, ok := graphQLPositiveInt(value)
			if !ok || n > pagination.MaxPageSize {
				return nil, fmt.Errorf("operation %q GraphQL page size variable %q must be a positive integer at most %d", op.ID, pagination.PageSizeVariable, pagination.MaxPageSize)
			}
		}
	}
	sch, _, err := graphQLOperationVariablesSchema(op)
	if err != nil {
		return nil, err
	}
	if err := sch.Validate(variables); err != nil {
		return nil, fmt.Errorf("operation %q graphql.variables_schema: %w", op.ID, err)
	}
	return variables, nil
}

func operationGraphQLDirectRead(ctx context.Context, b Bundle, op OperationSpec, req connectors.OperationDirectReadRequest, h Hooks) (connectors.DirectReadResult, error) {
	if len(req.PathParams) != 0 {
		return connectors.DirectReadResult{}, fmt.Errorf("operation %q fixed GraphQL query does not accept path parameters", op.ID)
	}
	if len(req.Query) != 0 {
		return connectors.DirectReadResult{}, fmt.Errorf("operation %q fixed GraphQL query does not accept query overrides", op.ID)
	}
	if req.RawBody != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation %q fixed GraphQL query does not accept a raw body", op.ID)
	}
	policy := strings.TrimSpace(req.OutputPolicy)
	if policy == "" {
		policy = op.OutputPolicy
	}
	if policy != op.OutputPolicy {
		return connectors.DirectReadResult{}, fmt.Errorf("operation %q output_policy must match declared policy %q", op.ID, op.OutputPolicy)
	}
	if err := validateDirectReadOutputPolicy(policy, op.GraphQL.Path, nil, connectors.RuntimeConfig{}); err != nil {
		return connectors.DirectReadResult{}, err
	}
	maxBytes := clampOperationDirectReadMaxBytes(req.MaxBytes, op.GraphQL.MaxBytes)
	variables, err := graphQLOperationVariables(op, req.Body, req.Page, req.PageCursor)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	payload, _, err := buildGraphQLOperationPayload(op, variables, maxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDirectReadTimeout)
	defer cancel()
	cfg := materializeConfigDefaults(b, req.Config)
	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	requester, err := rt.requesterFor(http.MethodPost, op.GraphQL.Path)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	requestPath := normalizeDirectReadPathForBaseURL(op.GraphQL.Path, directReadBaseURL(b, cfg))
	response, err := requester.DoLimited(ctx, http.MethodPost, requestPath, nil, payload, maxBytes)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		// Fixed GraphQL calls expose only the bounded, sanitized GraphQL error
		// metadata on their successful HTTP path. A non-2xx response has no
		// GraphQL envelope to sanitize, so do not let its raw provider body
		// bypass that same redaction boundary.
		message := safety.RedactErrorText(completeEngineErrorText(err))
		if hint != "" {
			message += ": " + hint
		}
		if class != "" {
			message = class + ": " + message
		}
		return connectors.DirectReadResult{}, formatResponseError(fmt.Sprintf("operation direct read POST %s: %s", op.GraphQL.Path, message), err)
	}
	data, metadata, err := graphQLOperationResponse(response.Body, maxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read GraphQL response: %w", err)
	}
	observeGraphQLRateLimit(ctx, requester, response, data)
	page, err := graphQLOperationPage(data, op.GraphQL.Pagination, metadata.PartialData)
	if err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read GraphQL pagination: %w", err)
	}
	body, err := applyDirectReadOutputPolicy(policy, data)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	redactFields := append([]string(nil), req.RedactFields...)
	if op.SensitivePolicy != nil {
		redactFields = append(redactFields, op.SensitivePolicy.RedactFields...)
	}
	if len(redactFields) != 0 {
		body = redactNamedJSONFields(body, redactFields)
	}
	return connectors.DirectReadResult{
		Connector: b.Name,
		Method:    http.MethodPost,
		Path:      op.GraphQL.Path,
		Status:    response.Status,
		Body:      body,
		GraphQL:   metadata,
		Page:      page,
	}, nil
}

func graphQLPositiveInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, v > 0
	case int64:
		return int(v), v > 0 && int64(int(v)) == v
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil && n > 0 && int64(int(n)) == n
	case float64:
		n := int(v)
		return n, v == float64(n) && n > 0
	default:
		return 0, false
	}
}

func buildGraphQLOperationPayload(op OperationSpec, variables map[string]any, maxBytes int) (map[string]any, string, error) {
	payload := map[string]any{
		"query":         op.GraphQL.Document,
		"operationName": op.GraphQL.OperationName,
		"variables":     variables,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, "", fmt.Errorf("operation %q encode GraphQL payload: %w", op.ID, err)
	}
	if len(raw) > maxBytes {
		return nil, "", fmt.Errorf("operation %q GraphQL request body too large: %d bytes exceeds limit %d", op.ID, len(raw), maxBytes)
	}
	return payload, string(raw), nil
}

// graphQLOperationResponse preserves a bounded data object and a deliberately
// tiny error/rate-limit summary. GraphQL returns HTTP 200 for resolver errors;
// queries surface partial data truthfully, while mutations use the same parser
// but their caller fails closed if Errors is non-empty.
func graphQLOperationResponse(raw []byte, maxBytes int) (map[string]any, *connectors.GraphQLResponseMetadata, error) {
	return graphQLOperationResponseWithRuntimeErrorPolicy(raw, maxBytes, false)
}

// graphQLOperationResponseWithRuntimeErrorPolicy decodes a bounded GraphQL response.
func graphQLOperationResponseWithRuntimeErrorPolicy(raw []byte, maxBytes int, retainRuntimeErrors bool) (map[string]any, *connectors.GraphQLResponseMetadata, error) {
	if len(raw) > maxBytes {
		return nil, nil, fmt.Errorf("GraphQL response too large: %d bytes exceeds limit %d", len(raw), maxBytes)
	}
	var envelope struct {
		Data   json.RawMessage `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		return nil, nil, fmt.Errorf("GraphQL response is not JSON: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return nil, nil, fmt.Errorf("GraphQL response contains multiple JSON values")
	} else if err != io.EOF {
		return nil, nil, fmt.Errorf("GraphQL response has trailing data: %w", err)
	}

	metadata := &connectors.GraphQLResponseMetadata{Errors: boundedGraphQLErrorMetadata(envelope.Errors, retainRuntimeErrors)}
	var data map[string]any
	if len(bytes.TrimSpace(envelope.Data)) != 0 && string(bytes.TrimSpace(envelope.Data)) != "null" {
		decoded, err := decodeDirectReadBody(envelope.Data, maxBytes)
		if err != nil {
			return nil, nil, fmt.Errorf("GraphQL data is not JSON: %w", err)
		}
		var ok bool
		data, ok = decoded.(map[string]any)
		if !ok {
			return nil, nil, fmt.Errorf("GraphQL data must be an object")
		}
		metadata.RateLimit = graphQLRateLimit(data)
	}
	metadata.PartialData = data != nil && len(metadata.Errors) > 0
	return data, metadata, nil
}

func boundedGraphQLErrorMetadata(items []struct {
	Message string `json:"message"`
}, retainRuntimeErrors bool) []connectors.GraphQLResultError {
	if len(items) == 0 {
		return nil
	}
	out := make([]connectors.GraphQLResultError, 0, min(len(items), maxGraphQLOperationErrors))
	for _, item := range items {
		message := sanitizeGraphQLErrorMessage(item.Message)
		if retainRuntimeErrors {
			message = item.Message
		}
		out = append(out, connectors.GraphQLResultError{Message: message})
		if len(out) == maxGraphQLOperationErrors {
			break
		}
	}
	return out
}

func sanitizeGraphQLErrorMessage(value string) string {
	value = strings.TrimSpace(safety.RedactErrorText(value))
	if value == "" {
		return "provider returned a GraphQL error without a message"
	}
	lower := strings.ToLower(value)
	for _, marker := range []string{"token", "secret", "password", "authorization", "credential", "private key"} {
		if strings.Contains(lower, marker) {
			return "provider GraphQL error message redacted"
		}
	}
	value = safety.SanitizeTerminal(value)
	if utf8.RuneCountInString(value) <= maxGraphQLErrorMessageSize {
		return value
	}
	return string([]rune(value)[:maxGraphQLErrorMessageSize]) + "…"
}

func graphQLErrorSummary(metadata *connectors.GraphQLResponseMetadata) string {
	if metadata == nil || len(metadata.Errors) == 0 {
		return ""
	}
	parts := make([]string, 0, len(metadata.Errors))
	for _, item := range metadata.Errors {
		parts = append(parts, item.Message)
	}
	return strings.Join(parts, "; ")
}

func graphQLRateLimit(data map[string]any) *connectors.GraphQLRateLimit {
	raw, ok := data["rateLimit"].(map[string]any)
	if !ok {
		return nil
	}
	limit, hasLimit := graphQLInt(raw["limit"])
	cost, hasCost := graphQLInt(raw["cost"])
	remaining, hasRemaining := graphQLInt(raw["remaining"])
	resetAt, hasReset := raw["resetAt"].(string)
	if hasReset {
		if _, err := time.Parse(time.RFC3339, resetAt); err != nil {
			hasReset = false
			resetAt = ""
		}
	}
	if !hasLimit && !hasCost && !hasRemaining && !hasReset {
		return nil
	}
	return &connectors.GraphQLRateLimit{Limit: limit, Cost: cost, Remaining: remaining, ResetAt: resetAt}
}

// observeGraphQLRateLimit bridges the fixed response selection declared by a
// GraphQL operation into the requester's typed observation seam. The data map
// has already passed the operation's bounded JSON parser, and no caller can
// supply another response path or raw provider body.
func observeGraphQLRateLimit(ctx context.Context, requester *connsdk.Requester, response *connsdk.Response, data map[string]any) {
	if requester == nil || response == nil {
		return
	}
	observation, ok := graphQLRateLimitObservation(data)
	if !ok {
		return
	}
	requester.ObserveRateLimit(ctx, response, observation)
}

func graphQLRateLimitObservation(data map[string]any) (connsdk.RateLimitObservation, bool) {
	raw, ok := data["rateLimit"].(map[string]any)
	if !ok {
		return connsdk.RateLimitObservation{}, false
	}
	limit, hasLimit := graphQLInt(raw["limit"])
	cost, hasCost := graphQLInt(raw["cost"])
	remaining, hasRemaining := graphQLInt(raw["remaining"])
	resetAt, hasReset := raw["resetAt"].(string)
	var parsedResetAt time.Time
	if hasReset {
		var err error
		parsedResetAt, err = time.Parse(time.RFC3339, resetAt)
		if err != nil {
			hasReset = false
		}
	}
	if !hasLimit && !hasCost && !hasRemaining && !hasReset {
		return connsdk.RateLimitObservation{}, false
	}
	observation := connsdk.RateLimitObservation{
		Source:       connsdk.RateLimitObservationSourceBody,
		Limit:        int64(limit),
		HasLimit:     hasLimit,
		Remaining:    int64(remaining),
		HasRemaining: hasRemaining,
		ResetAt:      parsedResetAt,
		HasReset:     hasReset,
		// GitHub's resetAt is an absolute timestamp.
		ResetAtAbsolute: hasReset,
	}
	if hasCost {
		observation.Cost = float64(cost)
		observation.HasCost = true
		observation.CostSource = connsdk.RateLimitCostSourceGraphQLRateLimit
	}
	return observation, true
}

func graphQLInt(value any) (int, bool) {
	switch v := value.(type) {
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil && int64(int(n)) == n
	case float64:
		n := int(v)
		return n, v == float64(n)
	case int:
		return v, true
	case int64:
		return int(v), int64(int(v)) == v
	default:
		return 0, false
	}
}

func graphQLOperationPage(data map[string]any, pagination *GraphQLOperationPaginationSpec, partial bool) (connectors.DirectReadPage, error) {
	if pagination == nil {
		return connectors.DirectReadPage{
			Strategy: graphQLCursorStrategy,
			Complete: false,
			Reason:   connectors.DirectReadPageReasonNoPagination,
		}, nil
	}
	if data == nil {
		return connectors.DirectReadPage{Strategy: graphQLCursorStrategy, Complete: false, Reason: connectors.DirectReadPageReasonAmbiguous}, nil
	}
	connection, found := graphQLValueAtPath(data, pagination.ConnectionPath)
	if !found {
		if partial {
			return connectors.DirectReadPage{Strategy: graphQLCursorStrategy, Complete: false, Reason: connectors.DirectReadPageReasonAmbiguous}, nil
		}
		return connectors.DirectReadPage{}, fmt.Errorf("GraphQL response omits declared connection path %q", pagination.ConnectionPath)
	}
	object, ok := connection.(map[string]any)
	if !ok {
		return connectors.DirectReadPage{}, fmt.Errorf("GraphQL declared connection path %q is not an object", pagination.ConnectionPath)
	}
	nodes, ok := object["nodes"].([]any)
	if !ok {
		if partial {
			return connectors.DirectReadPage{Strategy: graphQLCursorStrategy, Complete: false, Reason: connectors.DirectReadPageReasonAmbiguous}, nil
		}
		return connectors.DirectReadPage{}, fmt.Errorf("GraphQL declared connection path %q has no nodes array", pagination.ConnectionPath)
	}
	pageInfo, ok := object["pageInfo"].(map[string]any)
	if !ok {
		if partial {
			return connectors.DirectReadPage{Strategy: graphQLCursorStrategy, Records: len(nodes), Complete: false, Reason: connectors.DirectReadPageReasonAmbiguous}, nil
		}
		return connectors.DirectReadPage{}, fmt.Errorf("GraphQL declared connection path %q has no pageInfo object", pagination.ConnectionPath)
	}
	hasMore, ok := pageInfo["hasNextPage"].(bool)
	if !ok {
		return connectors.DirectReadPage{}, fmt.Errorf("GraphQL declared connection path %q pageInfo.hasNextPage is not boolean", pagination.ConnectionPath)
	}
	page := connectors.DirectReadPage{Strategy: graphQLCursorStrategy, Records: len(nodes), HasMore: hasMore, Complete: !hasMore}
	if !hasMore {
		return page, nil
	}
	next, ok := pageInfo["endCursor"].(string)
	if !ok || strings.TrimSpace(next) == "" {
		return connectors.DirectReadPage{}, fmt.Errorf("GraphQL declared connection path %q reports another page without endCursor", pagination.ConnectionPath)
	}
	page.NextCursor = next
	page.Reason = connectors.DirectReadPageReasonMorePages
	return page, nil
}

func graphQLValueAtPath(root map[string]any, path string) (any, bool) {
	var value any = root
	for _, segment := range strings.Split(path, ".") {
		object, ok := value.(map[string]any)
		if !ok {
			return nil, false
		}
		value, ok = object[segment]
		if !ok {
			return nil, false
		}
	}
	return value, true
}
