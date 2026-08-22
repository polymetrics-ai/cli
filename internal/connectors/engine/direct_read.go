package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	stdpath "path"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

const (
	defaultDirectReadMaxBytes                      = 1 << 20
	maxOperationDirectReadBytes                    = 16 << 20
	defaultDirectReadTimeout                       = 30 * time.Second
	directReadPolicyRepositoryContentsFileMetadata = "repository_contents_file_metadata"
	directReadPolicyRepositoryContentsDirectory    = "repository_contents_directory"
	directReadPolicyJSONRedacted                   = "json_redacted"
	directReadPolicyClinicalJSONRedacted           = "clinical_json_redacted"
	// directReadPolicyNone is an explicitly status-only operation contract.
	// It accepts a successful empty response such as GitHub's membership checks;
	// it is not a generic way to discard an otherwise meaningful response.
	directReadPolicyNone = "none"
	// directReadPolicyText is a bounded UTF-8 response contract for declared
	// text endpoints such as GitHub's Markdown renderer and /zen. Binary output
	// remains in the binary_download executor.
	directReadPolicyText = "text"
)

var surfacePathVarPattern = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// completeEngineErrorText renders only safe transport facts. The typed cause
// remains available through errors.As for classification and rate parking, but
// provider URLs, queries, headers, and bodies never cross the printable error
// boundary.
func completeEngineErrorText(err error) string {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return safety.RedactErrorText(err.Error())
	}
	if statusText := http.StatusText(httpErr.Status); statusText != "" {
		return fmt.Sprintf("http %d (%s)", httpErr.Status, statusText)
	}
	return fmt.Sprintf("http %d", httpErr.Status)
}

func OperationDirectRead(ctx context.Context, b Bundle, req connectors.OperationDirectReadRequest, h Hooks) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	op, err := operationDirectReadSpec(b, req.Operation)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	commandQueryFields := map[string]struct{}(nil)
	if req.CommandBindings != nil {
		if !operationDirectReadBindingsDeclaredByCommand(b, op.ID, req.CommandBindings.Path, req.CommandBindings.Query, req.CommandBindings.Body, req.CommandBindings.RawBody) {
			return connectors.DirectReadResult{}, fmt.Errorf("operation %q command bindings do not match a declared implemented command", op.ID)
		}
		if err := validateOperationDirectReadRequestBindings(req); err != nil {
			return connectors.DirectReadResult{}, fmt.Errorf("operation %q command bindings: %w", op.ID, err)
		}
		commandQueryFields = bindingFieldSet(req.CommandBindings.Query)
	}
	if op.Kind == "graphql_query" {
		if len(req.Headers) != 0 || len(req.HeaderValues) != 0 {
			return connectors.DirectReadResult{}, fmt.Errorf("operation %q fixed GraphQL query does not accept request header overrides", op.ID)
		}
		return operationGraphQLDirectRead(ctx, b, op, req, h)
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	cfg := materializeConfigDefaults(b, req.Config)
	if err := validateOperationDirectReadPathParams(op, req.PathParams); err != nil {
		return connectors.DirectReadResult{}, err
	}
	resolvedPath, err := resolveSurfaceEndpointPath(op.REST.Path, cfg, req.PathParams)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	queryMap, err := operationDirectReadQuery(op, req.Query, commandQueryFields)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if err := requireOperationQueryGroups(op, queryMap); err != nil {
		return connectors.DirectReadResult{}, err
	}
	query, err := directReadQuery(queryMap)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	policy := req.OutputPolicy
	if policy == "" {
		policy = op.OutputPolicy
	}
	if err := validateDirectReadOutputPolicy(policy, op.REST.Path, req.PathParams, cfg); err != nil {
		return connectors.DirectReadResult{}, err
	}
	maxBytes := clampOperationDirectReadMaxBytes(req.MaxBytes, op.REST.MaxBytes)
	body, err := operationReadBody(op, req.Body, req.RawBody, maxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	headers, err := operationRequestHeaders(b, op, req.Headers, req.HeaderValues)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDirectReadTimeout)
	defer cancel()
	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, directReadBaseURL(b, cfg))
	decoded, pageInfo, resp, err := readDirectPage(ctx, b, rt, directReadWalk{
		method:          method,
		declaredPat:     op.REST.Path,
		requestPath:     requestPath,
		query:           query,
		body:            body,
		bodyContentType: operationDirectReadContentType(op),
		headers:         headers,
		operation:       &op,
		outputPolicy:    policy,
		maxBytes:        maxBytes,
		page:            req.Page,
		pageCursor:      req.PageCursor,
		pagination:      op.REST.Pagination,
	})
	readResult := connectors.DirectReadResult{Connector: b.Name, Operation: op.ID, Method: method, Path: resolvedPath, Page: pageInfo}
	readResult.Receipt = providerResponseReceiptFromResponse(b, resp, cfg.Secrets)
	if readResult.Receipt == nil && err != nil {
		readResult.Receipt = providerResponseReceiptFromHTTPError(b, err, cfg.Secrets)
	}
	if readResult.Receipt != nil {
		readResult.Status = readResult.Receipt.Status
	}
	if err != nil {
		var tooLarge errDirectReadTooLarge
		if errors.As(err, &tooLarge) {
			return readResult, fmt.Errorf("operation direct read %s", tooLarge.Error())
		}
		// A pagination failure arrives with the response already fetched and
		// decoded, so it must not fall through to the not-JSON branch below.
		var pageErr errDirectReadPagination
		if errors.As(err, &pageErr) {
			return readResult, fmt.Errorf("operation direct read pagination: %w", pageErr.err)
		}
		var providerHTTPError *connsdk.HTTPError
		if resp == nil || errors.As(err, &providerHTTPError) {
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			msg := completeEngineErrorText(err)
			if hint != "" {
				msg = msg + ": " + hint
			}
			if class != "" {
				msg = class + ": " + msg
			}
			return readResult, formatResponseError(fmt.Sprintf("operation direct read %s %s: %s", method, op.REST.Path, msg), err)
		}
		return readResult, fmt.Errorf("operation direct read response is not JSON: %w", err)
	}
	decoded, err = applyDirectReadOutputPolicy(policy, decoded)
	if err != nil {
		return readResult, err
	}
	redactFields := append([]string(nil), req.RedactFields...)
	if op.SensitivePolicy != nil {
		redactFields = append(redactFields, op.SensitivePolicy.RedactFields...)
	}
	if len(redactFields) > 0 {
		decoded = redactNamedJSONFields(decoded, redactFields)
	}
	decoded = connectors.SanitizeProviderOutputForOutput(decoded, req.Config.Secrets)
	responseHeaders, err := operationResponseHeaders(b, op, resp.Header)
	if err != nil {
		return readResult, err
	}
	readResult.Body = decoded
	readResult.Headers = responseHeaders
	readResult.Page = connectors.SanitizeDirectReadPageForOutput(readResult.Page, cfg.Secrets)
	return readResult, nil
}

// PreflightOperationDirectRead proves a command's named operation can reach
// the bounded direct-read executor with its declared endpoint, response cap,
// and output policy. It shares operationDirectReadSpec with execution and is
// deliberately no-network: preflight must be safe to run across every bundle.
//
// pmcert:executes rest_read,provider_search,graphql_query
func PreflightOperationDirectRead(b Bundle, operation, method, endpointPath string, maxBytes int, outputPolicy string) error {
	op, err := operationDirectReadSpec(b, operation)
	if err != nil {
		return err
	}
	if op.Kind == "graphql_query" {
		if !strings.EqualFold(strings.TrimSpace(method), http.MethodPost) {
			return fmt.Errorf("operation direct read method %s does not match declared GraphQL method POST", strings.ToUpper(strings.TrimSpace(method)))
		}
		if endpointPath != op.GraphQL.Path {
			return fmt.Errorf("operation direct read path %q does not match declared GraphQL path %q", endpointPath, op.GraphQL.Path)
		}
		if maxBytes <= 0 {
			return fmt.Errorf("operation direct read command requires positive max_bytes")
		}
		if clampOperationDirectReadMaxBytes(maxBytes, op.GraphQL.MaxBytes) <= 0 {
			return fmt.Errorf("operation direct read has no executable response cap")
		}
		return validateDirectReadOutputPolicy(outputPolicy, op.GraphQL.Path, nil, connectors.RuntimeConfig{})
	}
	if !strings.EqualFold(strings.TrimSpace(method), op.REST.Method) {
		return fmt.Errorf("operation direct read method %s does not match declared operation method %s", strings.ToUpper(strings.TrimSpace(method)), strings.ToUpper(strings.TrimSpace(op.REST.Method)))
	}
	if endpointPath != op.REST.Path {
		return fmt.Errorf("operation direct read path %q does not match declared operation path %q", endpointPath, op.REST.Path)
	}
	if maxBytes <= 0 {
		return fmt.Errorf("operation direct read command requires positive max_bytes")
	}
	if clampOperationDirectReadMaxBytes(maxBytes, op.REST.MaxBytes) <= 0 {
		return fmt.Errorf("operation direct read has no executable response cap")
	}
	return validateDirectReadOutputPolicy(outputPolicy, op.REST.Path, nil, connectors.RuntimeConfig{})
}

// PreflightOperationDirectReadBindings proves that every command-controlled
// REST input is owned by the selected operation declaration. It is kept
// separate from endpoint preflight so older native adapters can delegate both
// checks through Base without learning engine internals.
func PreflightOperationDirectReadBindings(b Bundle, operation string, pathFields, queryFields, bodyFields []string, rawBody bool) error {
	op, err := operationDirectReadSpec(b, operation)
	if err != nil {
		return err
	}
	if op.Kind == "graphql_query" {
		if len(pathFields) != 0 || len(queryFields) != 0 || rawBody {
			return fmt.Errorf("operation %q fixed GraphQL query accepts only declared variables", op.ID)
		}
		_, schema, err := graphQLOperationVariablesSchema(op)
		if err != nil {
			return err
		}
		properties, _ := schema["properties"].(map[string]any)
		seen := make(map[string]struct{}, len(bodyFields))
		for _, field := range bodyFields {
			if !graphQLNamePattern.MatchString(field) {
				return fmt.Errorf("operation %q GraphQL variable %q is not a top-level declared variable", op.ID, field)
			}
			if _, duplicate := seen[field]; duplicate {
				return fmt.Errorf("operation %q maps more than one command flag to GraphQL variable %q", op.ID, field)
			}
			seen[field] = struct{}{}
			if _, declared := properties[field]; !declared {
				return fmt.Errorf("operation %q GraphQL variable %q is not declared", op.ID, field)
			}
		}
		return nil
	}
	if err := validateOperationDirectReadPathFields(op, pathFields); err != nil {
		return err
	}
	var commandQueryFields map[string]struct{}
	if operationDirectReadBindingsDeclaredByCommand(b, op.ID, pathFields, queryFields, bodyFields, rawBody) {
		commandQueryFields = bindingFieldSet(queryFields)
	}
	if err := validateOperationDirectReadQueryFields(op, queryFields, commandQueryFields); err != nil {
		return err
	}
	contentType := operationDirectReadContentType(op)
	if rawBody {
		if contentType != "text/plain" || len(bodyFields) != 0 {
			return fmt.Errorf("operation %q raw body requires the exact declared text/plain root string", op.ID)
		}
		return validateOperationDirectReadTextPlainContract(op)
	}
	if contentType == "text/plain" {
		if len(bodyFields) != 0 {
			return fmt.Errorf("operation %q text/plain request does not accept JSON body fields", op.ID)
		}
		return nil
	}
	return validateOperationDirectReadBodyFields(op, bodyFields)
}

func operationDirectReadBindingsDeclaredByCommand(b Bundle, operation string, pathFields, queryFields, bodyFields []string, rawBody bool) bool {
	if b.CLISurface == nil {
		return false
	}
	for _, command := range b.CLISurface.Commands {
		if command.Operation != operation || command.Intent != "direct_read" || command.Availability != "implemented" {
			continue
		}
		var declaredPath, declaredQuery, declaredBody []string
		declaredRawBody := false
		valid := true
		for _, flag := range command.Flags {
			mapsTo := strings.TrimSpace(flag.MapsTo)
			switch {
			case strings.HasPrefix(mapsTo, "path."):
				declaredPath = append(declaredPath, strings.TrimPrefix(mapsTo, "path."))
			case strings.HasPrefix(mapsTo, "query."):
				declaredQuery = append(declaredQuery, strings.TrimPrefix(mapsTo, "query."))
			case strings.HasPrefix(mapsTo, "body."):
				declaredBody = append(declaredBody, strings.TrimPrefix(mapsTo, "body."))
			case strings.HasPrefix(mapsTo, "header."):
			case mapsTo == "body" && !declaredRawBody:
				declaredRawBody = true
			default:
				valid = false
			}
		}
		if valid && declaredRawBody == rawBody && sameBindingFields(declaredPath, pathFields) && sameBindingFields(declaredQuery, queryFields) && sameBindingFields(declaredBody, bodyFields) {
			return true
		}
	}
	return false
}

func sameBindingFields(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, field := range left {
		counts[field]++
	}
	for _, field := range right {
		if counts[field] == 0 {
			return false
		}
		counts[field]--
	}
	return true
}

func bindingFieldSet(fields []string) map[string]struct{} {
	set := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		set[field] = struct{}{}
	}
	return set
}

func validateOperationDirectReadRequestBindings(req connectors.OperationDirectReadRequest) error {
	if req.CommandBindings == nil {
		return nil
	}
	for field := range req.PathParams {
		if _, declared := bindingFieldSet(req.CommandBindings.Path)[field]; !declared {
			return fmt.Errorf("path field %q is not bound by the selected command", field)
		}
	}
	for field := range req.Query {
		if _, declared := bindingFieldSet(req.CommandBindings.Query)[field]; !declared {
			return fmt.Errorf("query field %q is not bound by the selected command", field)
		}
	}
	for field := range req.Body {
		if !bindingOwnsTopLevelField(req.CommandBindings.Body, field) {
			return fmt.Errorf("body field %q is not bound by the selected command", field)
		}
	}
	if req.RawBody != nil && !req.CommandBindings.RawBody {
		return fmt.Errorf("raw body is not bound by the selected command")
	}
	return nil
}

func bindingOwnsTopLevelField(bindings []string, field string) bool {
	for _, binding := range bindings {
		if binding == field || strings.HasPrefix(binding, field+".") {
			return true
		}
	}
	return false
}

// operationDirectReadSpec is the static, no-network part of the bounded
// direct-read executor. Both execution and command preflight call it so their
// eligibility rules cannot drift apart.
func operationDirectReadSpec(b Bundle, operation string) (OperationSpec, error) {
	op, err := findOperation(b, operation)
	if err != nil {
		return OperationSpec{}, err
	}
	if op.Kind == "graphql_query" {
		if err := validateGraphQLOperationDirectContract(op, "query"); err != nil {
			return OperationSpec{}, err
		}
		if err := requireOperationDirectReadLedgerEndpoint(b, op.ID, op.Kind, http.MethodPost, op.GraphQL.Path, op.GraphQL.MaxBytes); err != nil {
			return OperationSpec{}, err
		}
		return op, nil
	}
	if err := validateRESTOperationPagination(op); err != nil {
		return OperationSpec{}, fmt.Errorf("operation %q pagination: %w", op.ID, err)
	}
	// provider_search shares this executor deliberately: its response bounding,
	// clamping, redaction and output-policy handling are the same as any other
	// bounded read. What differs is its stricter bundle-load contract.
	if (op.Kind != "rest_read" && op.Kind != "provider_search") || op.REST == nil {
		return OperationSpec{}, fmt.Errorf("operation direct read requires rest_read or provider_search operation, got %q", op.Kind)
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if method != http.MethodGet && method != http.MethodPost {
		return OperationSpec{}, fmt.Errorf("operation direct read requires GET or POST, got %s", method)
	}
	if op.Kind == "provider_search" && method != http.MethodPost {
		return OperationSpec{}, fmt.Errorf("provider search requires POST, got %s", method)
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		return OperationSpec{}, fmt.Errorf("operation direct read endpoint must be connector-relative, got absolute URL")
	}
	if method == http.MethodPost {
		if err := validateOperationDirectReadPOSTContract(op); err != nil {
			return OperationSpec{}, err
		}
	}
	if op.REST.MaxBytes <= 0 {
		return OperationSpec{}, fmt.Errorf("operation direct read requires positive max_bytes")
	}
	if err := requireOperationDirectReadLedgerEndpoint(b, "", op.Kind, method, op.REST.Path, op.REST.MaxBytes); err != nil {
		return OperationSpec{}, err
	}
	return op, nil
}

func DirectRead(ctx context.Context, b Bundle, req connectors.DirectReadRequest, h Hooks) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method != http.MethodGet {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read requires GET, got %s", method)
	}
	if isAbsoluteHTTPURL(req.Path) {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read endpoint must be connector-relative, got absolute URL")
	}
	if err := requireDirectReadEndpoint(b, method, req.Path); err != nil {
		return connectors.DirectReadResult{}, err
	}
	if isOperationOnlyDirectReadOutputPolicy(req.OutputPolicy) {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read output policy %q requires an operation-backed direct read", req.OutputPolicy)
	}
	cfg := materializeConfigDefaults(b, req.Config)
	resolvedPath, err := resolveSurfaceEndpointPath(req.Path, cfg, req.PathParams)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	query, err := directReadQuery(req.Query)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if err := validateDirectReadOutputPolicy(req.OutputPolicy, req.Path, req.PathParams, cfg); err != nil {
		return connectors.DirectReadResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDirectReadTimeout)
	defer cancel()

	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	maxBytes := clampDirectReadMaxBytes(req.MaxBytes)
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, directReadBaseURL(b, cfg))
	body, pageInfo, resp, err := readDirectPage(ctx, b, rt, directReadWalk{
		method:       method,
		declaredPat:  req.Path,
		requestPath:  requestPath,
		query:        query,
		outputPolicy: req.OutputPolicy,
		maxBytes:     maxBytes,
		page:         req.Page,
		pageCursor:   req.PageCursor,
	})
	readResult := connectors.DirectReadResult{Connector: b.Name, Method: method, Path: resolvedPath, Page: pageInfo}
	readResult.Receipt = providerResponseReceiptFromResponse(b, resp, cfg.Secrets)
	if readResult.Receipt == nil && err != nil {
		readResult.Receipt = providerResponseReceiptFromHTTPError(b, err, cfg.Secrets)
	}
	if readResult.Receipt != nil {
		readResult.Status = readResult.Receipt.Status
	}
	if err != nil {
		var tooLarge errDirectReadTooLarge
		if errors.As(err, &tooLarge) {
			return readResult, fmt.Errorf("direct read %s", tooLarge.Error())
		}
		// A pagination failure arrives with the response already fetched and
		// decoded, so it must not fall through to the not-JSON branch below.
		var pageErr errDirectReadPagination
		if errors.As(err, &pageErr) {
			return readResult, fmt.Errorf("direct read pagination: %w", pageErr.err)
		}
		var providerHTTPError *connsdk.HTTPError
		if resp == nil || errors.As(err, &providerHTTPError) {
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			msg := completeEngineErrorText(err)
			if hint != "" {
				msg = msg + ": " + hint
			}
			if class != "" {
				msg = class + ": " + msg
			}
			return readResult, formatResponseError(fmt.Sprintf("direct read %s %s: %s", method, req.Path, msg), err)
		}
		return readResult, fmt.Errorf("direct read response is not JSON: %w", err)
	}
	body, err = applyDirectReadOutputPolicy(req.OutputPolicy, body)
	if err != nil {
		return readResult, err
	}
	if len(req.RedactFields) > 0 {
		body = redactNamedJSONFields(body, req.RedactFields)
	}
	body = connectors.SanitizeProviderOutputForOutput(body, req.Config.Secrets)
	readResult.Body = body
	readResult.Page = connectors.SanitizeDirectReadPageForOutput(readResult.Page, cfg.Secrets)
	return readResult, nil
}

func findOperation(b Bundle, id string) (OperationSpec, error) {
	for _, op := range b.Operations {
		if op.ID == id {
			return op, nil
		}
	}
	return OperationSpec{}, fmt.Errorf("operation %q not found in bundle %q", id, b.Name)
}

func requireOperationSurfaceEndpoint(b Bundle, method, endpointPath string) error {
	if b.Surface == nil {
		return nil
	}
	for _, ep := range b.Surface.Endpoints {
		if strings.EqualFold(ep.Method, method) && ep.Path == endpointPath {
			if ep.Operation == nil && (ep.CoveredBy == nil || (ep.CoveredBy.DirectRead == "" && len(ep.CoveredBy.DirectReads) == 0)) {
				return fmt.Errorf("api_surface endpoint %s %s is not declared as an operation or direct_read command", method, endpointPath)
			}
			return nil
		}
	}
	return fmt.Errorf("api_surface endpoint %s %s not found", method, endpointPath)
}

func requireOperationDirectReadLedgerEndpoint(b Bundle, operation, kind, method, endpointPath string, maxBytes int) error {
	ledger := operationDirectReadEndpointLedger(b)
	if ledger == nil {
		return fmt.Errorf("runtime operation endpoint ledger is unavailable for bundle %q", b.Name)
	}
	for _, entry := range ledger.entries {
		if strings.EqualFold(entry.Method, method) && entry.Path == endpointPath && entry.Kind == kind && entry.Operation == operation && entry.MaxBytes == maxBytes {
			return nil
		}
	}
	if operation != "" {
		return fmt.Errorf("runtime operation endpoint ledger does not contain %s %s operation %q kind %q with max_bytes %d", method, endpointPath, operation, kind, maxBytes)
	}
	return fmt.Errorf("runtime operation endpoint ledger does not contain %s %s kind %q with max_bytes %d", method, endpointPath, kind, maxBytes)
}

// requireOperationQueryGroups enforces rest.required_query against the merged
// query (the operation's own declared values plus the caller's) BEFORE any
// network call. Some endpoints answer an unfiltered request with the entire
// tenant; declaring the filter mandatory is what lets such an endpoint be
// executable at all instead of permanently blocked.
//
// A parameter counts as supplied only when its value is non-blank: a
// present-but-empty value produces exactly the unfiltered request the
// constraint exists to prevent.
func requireOperationQueryGroups(op OperationSpec, query map[string]string) error {
	if op.REST == nil {
		return nil
	}
	for _, group := range op.REST.RequiredQuery {
		if queryGroupSatisfied(group, query) {
			continue
		}
		return fmt.Errorf("operation %q requires at least one of query parameters %s", op.ID, strings.Join(group.AnyOf, ", "))
	}
	return nil
}

func queryGroupSatisfied(group RequiredQueryGroup, query map[string]string) bool {
	for _, name := range group.AnyOf {
		if strings.TrimSpace(query[strings.TrimSpace(name)]) != "" {
			return true
		}
	}
	return false
}

func validateOperationDirectReadPathFields(op OperationSpec, pathFields []string) error {
	declared, err := operationDirectWritePathParameterNames(op)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(pathFields))
	for _, field := range pathFields {
		if err := safety.ValidateIdentifier(field, "operation path field"); err != nil {
			return fmt.Errorf("operation %q path field: %w", op.ID, err)
		}
		if _, duplicate := seen[field]; duplicate {
			return fmt.Errorf("operation %q maps more than one command flag to path parameter %q", op.ID, field)
		}
		seen[field] = struct{}{}
		if _, ok := declared[field]; !ok {
			return fmt.Errorf("operation %q path parameter %q is not declared by rest.path", op.ID, field)
		}
	}
	return nil
}

func validateOperationDirectReadPathParams(op OperationSpec, pathParams map[string]string) error {
	fields := make([]string, 0, len(pathParams))
	for field := range pathParams {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	if err := validateOperationDirectReadPathFields(op, fields); err != nil {
		return err
	}
	parameters, err := operationParametersForLocation(op, "path")
	if err != nil {
		return err
	}
	for _, field := range fields {
		parameter, declared := parameters[field]
		if !declared {
			parameter = OperationParameter{Name: field, In: "path", Type: "string"}
		}
		if err := validateOperationParameterWireValue(op, parameter, "path", pathParams[field]); err != nil {
			return err
		}
	}
	return nil
}

func operationDirectReadQueryParameters(op OperationSpec) (map[string]OperationParameter, error) {
	return operationParametersForLocation(op, "query")
}

func validateOperationDirectReadQueryFields(op OperationSpec, queryFields []string, commandFields map[string]struct{}) error {
	parameters, err := operationDirectReadQueryParameters(op)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(queryFields))
	for _, field := range queryFields {
		if err := safety.ValidateIdentifier(field, "operation query parameter"); err != nil {
			return fmt.Errorf("operation %q query field: %w", op.ID, err)
		}
		if _, duplicate := seen[field]; duplicate {
			return fmt.Errorf("operation %q maps more than one command flag to query parameter %q", op.ID, field)
		}
		seen[field] = struct{}{}
		if _, fixed := op.REST.Query[field]; fixed {
			return fmt.Errorf("operation %q query parameter %q is fixed by rest.query and cannot be caller-bound", op.ID, field)
		}
		if _, declared := parameters[field]; !declared {
			if _, commandDeclared := commandFields[field]; commandDeclared {
				continue
			}
			return fmt.Errorf("operation %q query parameter %q is not source-declared in rest.parameters", op.ID, field)
		}
	}
	for name, parameter := range parameters {
		if !parameter.Required {
			continue
		}
		if _, fixed := op.REST.Query[name]; fixed {
			continue
		}
		if _, bound := seen[name]; !bound {
			return fmt.Errorf("operation %q requires query parameter %q", op.ID, name)
		}
	}
	return nil
}

func operationDirectReadQuery(op OperationSpec, requested map[string]string, commandFields map[string]struct{}) (map[string]string, error) {
	parameters, err := operationDirectReadQueryParameters(op)
	if err != nil {
		return nil, err
	}
	requestedNames := make([]string, 0, len(requested))
	for name := range requested {
		requestedNames = append(requestedNames, name)
	}
	sort.Strings(requestedNames)
	if err := validateOperationDirectReadQueryFields(op, requestedNames, commandFields); err != nil {
		return nil, err
	}
	query := make(map[string]string, len(op.REST.Query)+len(requested))
	for name, value := range op.REST.Query {
		parameter, declared := parameters[name]
		if !declared {
			parameter = OperationParameter{Name: name, In: "query", Type: "string"}
		}
		if err := validateOperationParameterWireValue(op, parameter, "query", value); err != nil {
			return nil, fmt.Errorf("fixed rest.query: %w", err)
		}
		query[name] = value
	}
	for _, name := range requestedNames {
		parameter, declared := parameters[name]
		if !declared {
			parameter = OperationParameter{Name: name, In: "query", Type: "string", MaxBytes: defaultOperationParameterMaxBytes}
		}
		if err := validateOperationParameterWireValue(op, parameter, "query", requested[name]); err != nil {
			return nil, err
		}
		query[name] = requested[name]
	}
	return query, nil
}

func validateOperationDirectReadBodyFields(op OperationSpec, bodyFields []string) error {
	if op.REST == nil || strings.ToUpper(strings.TrimSpace(op.REST.Method)) != http.MethodPost {
		if len(bodyFields) == 0 {
			return nil
		}
		return fmt.Errorf("operation %q does not accept request body fields", op.ID)
	}
	if operationDirectReadContentType(op) != "application/json" {
		if len(bodyFields) == 0 {
			return nil
		}
		return fmt.Errorf("operation %q does not accept JSON body fields", op.ID)
	}
	if len(bodyFields) == 0 {
		_, err := CompileSchema(op.REST.BodySchema)
		return err
	}
	normalized, _, err := compileOperationDirectReadBodySchema(op)
	if err != nil {
		return err
	}
	return validateOperationDirectWriteBodyFields(normalized, bodyFields)
}

func operationReadBody(op OperationSpec, overrides map[string]any, rawBody *string, maxBytes int) (any, error) {
	if op.REST == nil || strings.ToUpper(strings.TrimSpace(op.REST.Method)) != http.MethodPost {
		if rawBody != nil {
			return nil, fmt.Errorf("operation %q raw body requires a POST operation", op.ID)
		}
		return nil, nil
	}
	contentType := operationDirectReadContentType(op)
	if contentType == "text/plain" {
		if len(overrides) != 0 {
			return nil, fmt.Errorf("operation %q text/plain request cannot mix a raw body with JSON body fields", op.ID)
		}
		if rawBody == nil {
			return nil, fmt.Errorf("operation %q text/plain request requires a raw body", op.ID)
		}
		if len(*rawBody) > maxBytes {
			return nil, fmt.Errorf("operation %q request body too large: %d bytes exceeds limit %d", op.ID, len(*rawBody), maxBytes)
		}
		sch, err := CompileSchema(op.REST.BodySchema)
		if err != nil {
			return nil, fmt.Errorf("operation %q: compile body_schema: %w", op.ID, err)
		}
		if err := sch.Validate(*rawBody); err != nil {
			return nil, fmt.Errorf("operation %q: body_schema: %w", op.ID, err)
		}
		return *rawBody, nil
	}
	if rawBody != nil {
		return nil, fmt.Errorf("operation %q raw body requires declared text/plain content_type", op.ID)
	}
	bodyFields := make([]string, 0, len(overrides))
	for field := range overrides {
		bodyFields = append(bodyFields, field)
	}
	sort.Strings(bodyFields)
	if err := validateOperationDirectReadBodyFields(op, bodyFields); err != nil {
		return nil, err
	}
	body := cloneAnyMap(op.REST.Body)
	for key, value := range overrides {
		body[key] = value
	}
	if len(op.REST.BodySchema) > 0 {
		var schema *Schema
		var err error
		if len(overrides) == 0 {
			schema, err = CompileSchema(op.REST.BodySchema)
		} else {
			_, compiled, compileErr := compileOperationDirectReadBodySchema(op)
			err = compileErr
			if compiled != nil {
				schema = compiled.schema
			}
		}
		if err != nil {
			return nil, fmt.Errorf("operation %q: compile body_schema: %w", op.ID, err)
		}
		if err := schema.Validate(body); err != nil {
			return nil, fmt.Errorf("operation %q: body_schema: %w", op.ID, err)
		}
	}
	return body, nil
}

// validateOperationDirectReadPOSTContract is shared by bundle loading and
// runtime preflight. Keeping this contract in one function prevents a bundle
// from validating as implemented while the executor refuses its POST format.
func validateOperationDirectReadPOSTContract(op OperationSpec) error {
	if op.REST == nil {
		return fmt.Errorf("operation direct read POST requires rest declaration")
	}
	if len(op.REST.BodySchema) == 0 {
		return fmt.Errorf("operation direct read POST requires body_schema")
	}
	contentType, _, err := mime.ParseMediaType(strings.TrimSpace(op.REST.ContentType))
	if err != nil {
		return fmt.Errorf("operation direct read POST has invalid content_type %q: %w", op.REST.ContentType, err)
	}
	contentType = strings.ToLower(contentType)
	switch contentType {
	case "application/json":
		if _, err := CompileSchema(op.REST.BodySchema); err != nil {
			return fmt.Errorf("operation direct read body_schema: %w", err)
		}
		return nil
	case "text/plain":
		return validateOperationDirectReadTextPlainContract(op)
	default:
		return fmt.Errorf("operation direct read POST requires application/json or text/plain content_type")
	}
}

// compileOperationDirectReadBodySchema applies the engine's conservative
// closure defaults before compiling a REST read body. Imported provider
// schemas frequently omit JSON Schema's permissive-by-default keywords; an
// executable command must not turn that omission into caller authority. An
// explicit open object remains invalid, while omitted object/array bounds are
// filled with the same finite engine ceilings used by structured writes.
func compileOperationDirectReadBodySchema(op OperationSpec) (OperationSpec, *structuredRESTBodySchemaCompilation, error) {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return OperationSpec{}, nil, fmt.Errorf("operation %q structured JSON body requires body_schema", op.ID)
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return OperationSpec{}, nil, fmt.Errorf("operation %q body_schema is not an object: %w", op.ID, err)
	}
	if err := closeOperationDirectReadSchemaNode(op.ID, root, "body_schema", 1); err != nil {
		return OperationSpec{}, nil, err
	}
	normalizedRaw, err := json.Marshal(root)
	if err != nil {
		return OperationSpec{}, nil, fmt.Errorf("operation %q normalize body_schema: %w", op.ID, err)
	}
	normalized := op
	rest := *op.REST
	rest.BodySchema = normalizedRaw
	normalized.REST = &rest
	compiled, err := compileStructuredRESTBodySchema(normalized)
	if err != nil {
		return OperationSpec{}, nil, err
	}
	return normalized, compiled, nil
}

func closeOperationDirectReadSchemaNode(operation string, node map[string]any, path string, depth int) error {
	if depth > maxStructuredRESTBodyDepth {
		return fmt.Errorf("operation %q %s exceeds structured body depth limit %d", operation, path, maxStructuredRESTBodyDepth)
	}
	object, array := operationDirectWriteBodyNodeKinds(node)
	if object {
		if open, declared := node["additionalProperties"].(bool); declared && open {
			return fmt.Errorf("operation %q %s explicitly permits additionalProperties", operation, path)
		}
		node["additionalProperties"] = false
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			properties = map[string]any{}
			node["properties"] = properties
		}
		for _, name := range sortedMapKeys(properties) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return fmt.Errorf("operation %q %s/%s must be a schema object", operation, path, name)
			}
			if err := closeOperationDirectReadSchemaNode(operation, child, path+"/"+name, depth+1); err != nil {
				return err
			}
		}
	}
	if array {
		if _, declared := node["maxItems"]; !declared {
			node["maxItems"] = float64(maxStructuredRESTBodyItems)
		}
		if items, ok := node["items"].(map[string]any); ok {
			if err := closeOperationDirectReadSchemaNode(operation, items, path+"/items", depth+1); err != nil {
				return err
			}
		}
		if prefix, ok := node["prefixItems"].([]any); ok {
			for index, raw := range prefix {
				child, ok := raw.(map[string]any)
				if !ok {
					return fmt.Errorf("operation %q %s/prefixItems/%d must be a schema object", operation, path, index)
				}
				if err := closeOperationDirectReadSchemaNode(operation, child, fmt.Sprintf("%s/prefixItems/%d", path, index), depth+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// validateOperationDirectReadTextPlainContract is the loader-safe half of
// the literal-body contract. It is separate from JSON body validation because
// generated metadata for blocked operations may use schema annotations the
// executor does not implement; only a named operation command reaches the
// full validateOperationDirectReadPOSTContract preflight.
func validateOperationDirectReadTextPlainContract(op OperationSpec) error {
	if op.REST == nil {
		return fmt.Errorf("operation direct read POST requires rest declaration")
	}
	if op.Kind != "rest_read" {
		return fmt.Errorf("operation direct read POST text/plain content_type requires rest_read")
	}
	if len(op.REST.Body) != 0 {
		return fmt.Errorf("operation direct read POST text/plain content_type must not declare rest.body")
	}
	if !bodySchemaHasRootString(op.REST.BodySchema) {
		return fmt.Errorf("operation direct read POST text/plain content_type requires a root string body_schema")
	}
	if _, err := CompileSchema(op.REST.BodySchema); err != nil {
		return fmt.Errorf("operation direct read body_schema: %w", err)
	}
	return nil
}

func operationDirectReadContentType(op OperationSpec) string {
	if op.REST == nil {
		return ""
	}
	contentType, _, err := mime.ParseMediaType(strings.TrimSpace(op.REST.ContentType))
	if err != nil {
		return ""
	}
	return strings.ToLower(contentType)
}

func bodySchemaHasRootString(raw json.RawMessage) bool {
	var body map[string]json.RawMessage
	if err := json.Unmarshal(raw, &body); err != nil {
		return false
	}
	var typeName string
	if err := json.Unmarshal(body["type"], &typeName); err != nil {
		return false
	}
	return typeName == "string"
}

func cloneAnyMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func clampOperationDirectReadMaxBytes(requested, operationMax int) int {
	maxBytes := requested
	if maxBytes <= 0 {
		maxBytes = operationMax
	}
	if maxBytes <= 0 || maxBytes > maxOperationDirectReadBytes {
		maxBytes = maxOperationDirectReadBytes
	}
	if operationMax > 0 && maxBytes > operationMax {
		return operationMax
	}
	return maxBytes
}

func decodeDirectReadBody(raw []byte, maxBytes int) (any, error) {
	if maxBytes >= 0 && len(raw) > maxBytes {
		return nil, fmt.Errorf("response body exceeds limit %d", maxBytes)
	}
	var body any
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		return nil, err
	}
	var extra any
	if err := dec.Decode(&extra); err == nil {
		return nil, fmt.Errorf("response contains multiple JSON values")
	} else if err != io.EOF {
		return nil, err
	}
	return body, nil
}

// decodeDirectReadResponse is the single decoder for a declared direct-read
// response policy. It deliberately selects only between the bounded response
// contracts the operation declared; Requester has already captured at most the
// operation cap plus one byte, and readDirectPage rejects that sentinel byte
// before this function runs.
func decodeDirectReadResponse(policy string, raw []byte, maxBytes int) (any, error) {
	switch policy {
	case directReadPolicyNone:
		if len(raw) != 0 {
			return nil, fmt.Errorf("status-only response must be empty, got %d bytes", len(raw))
		}
		return nil, nil
	case directReadPolicyText:
		if !utf8.Valid(raw) {
			return nil, fmt.Errorf("text response must be valid UTF-8")
		}
		return string(raw), nil
	default:
		return decodeDirectReadBody(raw, maxBytes)
	}
}

func clampDirectReadMaxBytes(maxBytes int) int {
	if maxBytes <= 0 || maxBytes > defaultDirectReadMaxBytes {
		return defaultDirectReadMaxBytes
	}
	return maxBytes
}

func validateDirectReadOutputPolicy(policy, endpointPath string, pathParams map[string]string, cfg connectors.RuntimeConfig) error {
	switch policy {
	case directReadPolicyRepositoryContentsFileMetadata, directReadPolicyRepositoryContentsDirectory:
		path, err := repositoryDirectReadPathValue(policy, endpointPath, pathParams, cfg)
		if err != nil {
			return err
		}
		if err := rejectSensitiveRepositoryPath(path); err != nil {
			return err
		}
		return nil
	case directReadPolicyJSONRedacted, directReadPolicyClinicalJSONRedacted,
		directReadPolicyNone, directReadPolicyText:
		return nil
	default:
		return fmt.Errorf("direct read output policy %q is not supported", policy)
	}
}

// isOperationOnlyDirectReadOutputPolicy identifies response contracts that need
// an operation's reviewed REST declaration. The legacy DirectRead path has only
// a caller-supplied endpoint and must not become a way to discard arbitrary
// provider responses.
func isOperationOnlyDirectReadOutputPolicy(policy string) bool {
	return policy == directReadPolicyNone || policy == directReadPolicyText
}

func repositoryDirectReadPathValue(policy, endpointPath string, pathParams map[string]string, cfg connectors.RuntimeConfig) (string, error) {
	if !surfacePathHasVariable(endpointPath, "path") {
		return "", fmt.Errorf("direct read output policy %q requires endpoint path variable {path}", policy)
	}
	if pathParams != nil && pathParams["path"] != "" {
		return pathParams["path"], nil
	}
	if cfg.Config != nil {
		return cfg.Config["path"], nil
	}
	return "", nil
}

func surfacePathHasVariable(template, name string) bool {
	for _, match := range surfacePathVarPattern.FindAllStringSubmatch(template, -1) {
		if len(match) == 2 && match[1] == name {
			return true
		}
	}
	return false
}

func applyDirectReadOutputPolicy(policy string, body any) (any, error) {
	switch policy {
	case directReadPolicyRepositoryContentsFileMetadata:
		obj, ok := body.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("direct read output policy %q requires a file metadata object", policy)
		}
		if typ, _ := obj["type"].(string); typ == "dir" {
			return nil, fmt.Errorf("direct read output policy %q received a directory response", policy)
		}
		return redactRepositoryContentsObject(obj), nil
	case directReadPolicyRepositoryContentsDirectory:
		items, ok := body.([]any)
		if !ok {
			return nil, fmt.Errorf("direct read output policy %q requires a directory listing array", policy)
		}
		out := make([]any, 0, len(items))
		for _, item := range items {
			if obj, ok := item.(map[string]any); ok {
				out = append(out, redactRepositoryContentsObject(obj))
				continue
			}
			out = append(out, item)
		}
		return out, nil
	case directReadPolicyJSONRedacted, directReadPolicyClinicalJSONRedacted:
		return cloneJSONValue(body), nil
	case directReadPolicyNone:
		if body != nil {
			return nil, fmt.Errorf("direct read output policy %q received a response body", policy)
		}
		return nil, nil
	case directReadPolicyText:
		text, ok := body.(string)
		if !ok {
			return nil, fmt.Errorf("direct read output policy %q requires a text response", policy)
		}
		return text, nil
	default:
		return nil, fmt.Errorf("direct read output policy %q is not supported", policy)
	}
}

func directReadBaseURL(b Bundle, cfg connectors.RuntimeConfig) string {
	baseURL, err := Interpolate(b.HTTP.URL, Vars{Config: cfg.Config, Secrets: cfg.Secrets})
	if err != nil || strings.TrimSpace(baseURL) == "" {
		if cfg.Config != nil && cfg.Config["base_url"] != "" {
			return cfg.Config["base_url"]
		}
		return b.HTTP.URL
	}
	return baseURL
}

func normalizeDirectReadPathForBaseURL(resolvedPath, baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return resolvedPath
	}
	basePath := strings.TrimRight(parsed.EscapedPath(), "/")
	if basePath == "" || basePath == "." {
		return resolvedPath
	}
	if resolvedPath == basePath {
		return "/"
	}
	prefix := basePath + "/"
	if strings.HasPrefix(resolvedPath, prefix) {
		return "/" + strings.TrimPrefix(resolvedPath, prefix)
	}
	return resolvedPath
}

func cloneJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			out[key] = cloneJSONValue(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = cloneJSONValue(item)
		}
		return out
	default:
		return value
	}
}

func redactNamedJSONFields(value any, fields []string) any {
	if len(fields) == 0 {
		return value
	}
	fieldSet := make(map[string]bool, len(fields))
	for _, field := range fields {
		fieldSet[normalizeJSONFieldName(field)] = true
	}
	return redactJSONFieldsBySet(value, fieldSet)
}

func redactJSONFieldsBySet(value any, fields map[string]bool) any {
	return redactJSONFieldsByPredicate(value, func(name string) bool { return fields[normalizeJSONFieldName(name)] })
}

func redactJSONFieldsByPredicate(value any, shouldRedact func(string) bool) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if item != nil && shouldRedact(key) {
				out[key+"_redacted"] = true
				continue
			}
			out[key] = redactJSONFieldsByPredicate(item, shouldRedact)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactJSONFieldsByPredicate(item, shouldRedact)
		}
		return out
	default:
		return value
	}
}

func normalizeJSONFieldName(name string) string {
	return strings.ToLower(strings.NewReplacer("-", "_", " ", "_", ".", "_").Replace(name))
}

func redactRepositoryContentsObject(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+2)
	for k, v := range in {
		switch k {
		case "content":
			out["content_redacted"] = true
		case "download_url":
			if v != nil {
				out["download_url_redacted"] = true
			}
		default:
			out[k] = v
		}
	}
	return out
}

func rejectSensitiveRepositoryPath(value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	clean := stdpath.Clean(value)
	for _, part := range strings.Split(clean, "/") {
		lower := strings.ToLower(part)
		if isSensitiveRepositoryPathPart(lower) {
			return fmt.Errorf("repository path %q is blocked by direct read output policy", value)
		}
	}
	return nil
}

func isSensitiveRepositoryPathPart(part string) bool {
	switch part {
	case ".env", ".npmrc", ".pypirc", ".netrc", ".pgpass", ".ssh", ".gnupg",
		"id_rsa", "id_dsa", "id_ecdsa", "id_ed25519",
		"credentials", "credentials.json", "secrets.json", "secret.json":
		return true
	}
	if strings.HasPrefix(part, ".env.") {
		return true
	}
	for _, suffix := range []string{".pem", ".key", ".p12", ".pfx"} {
		if strings.HasSuffix(part, suffix) {
			return true
		}
	}
	return false
}

func requireDirectReadEndpoint(b Bundle, method, endpointPath string) error {
	if b.Surface != nil {
		return requireDirectReadSurfaceEndpoint(b.Surface, method, endpointPath)
	}
	if b.CLISurface != nil {
		for _, cmd := range b.CLISurface.Commands {
			if cmd.Intent != "direct_read" || cmd.Availability != "implemented" {
				continue
			}
			for _, ref := range cmd.APISurface {
				if strings.EqualFold(ref.Method, method) && ref.Path == endpointPath {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("direct read endpoint %s %s is not declared in command metadata", method, endpointPath)
}

func requireDirectReadSurfaceEndpoint(surface *APISurface, method, endpointPath string) error {
	for _, ep := range surface.Endpoints {
		if strings.EqualFold(ep.Method, method) && ep.Path == endpointPath {
			if ep.CoveredBy == nil || (ep.CoveredBy.DirectRead == "" && len(ep.CoveredBy.DirectReads) == 0) {
				return fmt.Errorf("api_surface endpoint %s %s is not covered by a direct_read command", method, endpointPath)
			}
			return nil
		}
	}
	return fmt.Errorf("api_surface endpoint %s %s not found", method, endpointPath)
}

func resolveSurfaceEndpointPath(template string, cfg connectors.RuntimeConfig, pathParams map[string]string) (string, error) {
	if strings.TrimSpace(template) == "" {
		return "", fmt.Errorf("direct read endpoint path is required")
	}
	if isAbsoluteHTTPURL(template) {
		return "", fmt.Errorf("direct read endpoint must be connector-relative, got absolute URL")
	}
	var firstErr error
	resolved := surfacePathVarPattern.ReplaceAllStringFunc(template, func(match string) string {
		if firstErr != nil {
			return ""
		}
		name := strings.Trim(match, "{}")
		value, ok := pathParams[name]
		if !ok || value == "" {
			value, ok = cfg.Config[name]
		}
		if !ok || value == "" {
			firstErr = fmt.Errorf("missing path variable %q", name)
			return ""
		}
		encoded, err := encodeSurfacePathValue(name, value)
		if err != nil {
			firstErr = err
			return ""
		}
		if len(encoded) > maxOperationParameterMaxBytes {
			firstErr = fmt.Errorf("path variable %q encoded value exceeds byte cap %d", name, maxOperationParameterMaxBytes)
			return ""
		}
		return encoded
	})
	if firstErr != nil {
		return "", firstErr
	}
	if strings.Contains(resolved, "{") || strings.Contains(resolved, "}") {
		return "", fmt.Errorf("unresolved path template %q", template)
	}
	return resolved, nil
}

func encodeSurfacePathValue(name, value string) (string, error) {
	if name == "path" || name == "ref" {
		if strings.Contains(value, "\\") {
			return "", fmt.Errorf("path variable %q must use forward slashes", name)
		}
		if name == "path" {
			if err := safety.ValidateRelativePath(value, "path variable "+name); err != nil {
				return "", err
			}
		} else {
			if strings.TrimSpace(value) == "" {
				return "", fmt.Errorf("path variable %q is required", name)
			}
			for _, part := range strings.Split(value, "/") {
				if err := safety.ValidateIdentifier(part, "path variable "+name); err != nil {
					return "", err
				}
			}
		}
		clean := stdpath.Clean(value)
		if clean == "." {
			return "", fmt.Errorf("path variable %q is required", name)
		}
		parts := strings.Split(clean, "/")
		for i, part := range parts {
			parts[i] = url.PathEscape(part)
		}
		return strings.Join(parts, "/"), nil
	}
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("path variable %q is required", name)
	}
	if err := safety.RejectDangerousChars(value, "path variable "+name); err != nil {
		return "", err
	}
	return url.PathEscape(value), nil
}

func directReadQuery(query map[string]string) (url.Values, error) {
	values := url.Values{}
	for name, value := range query {
		if err := safety.ValidateIdentifier(name, "query parameter"); err != nil {
			return nil, err
		}
		if err := safety.RejectDangerousChars(value, "query parameter "+name); err != nil {
			return nil, err
		}
		if encoded := url.QueryEscape(value); len(encoded) > maxOperationParameterMaxBytes {
			return nil, fmt.Errorf("query parameter %q encoded value exceeds byte cap %d", name, maxOperationParameterMaxBytes)
		}
		values.Set(name, value)
	}
	return values, nil
}

func isAbsoluteHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
