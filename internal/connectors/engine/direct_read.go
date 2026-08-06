package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	stdpath "path"
	"regexp"
	"strings"
	"time"

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
	callerSuppliedIdentifierSetTargetSegmentMarker = "polymetrics_identifier_set_target_segment"
	callerSuppliedIdentifierSetTargetPathMarker    = "polymetrics_identifier_set_target_path"
)

var surfacePathVarPattern = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)
var chainAddressIdentifierSetPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*:0x[0-9a-fA-F]{40}$`)

// CallerSuppliedIdentifierSetError is the typed, value-redacted rejection for
// a caller-supplied identifier collection. It deliberately carries only the
// declared parameter contract, never an identifier value.
type CallerSuppliedIdentifierSetError struct {
	Parameter     string
	Reason        CallerSuppliedIdentifierSetErrorReason
	Limit         int
	Position      int
	ExpectedShape string
}

// CallerSuppliedIdentifierSetErrorReason classifies a safe, caller-correctable
// identifier-set rejection without retaining any supplied value.
type CallerSuppliedIdentifierSetErrorReason string

const (
	// CallerSuppliedIdentifierSetUndeclared means input named no declared set.
	CallerSuppliedIdentifierSetUndeclared CallerSuppliedIdentifierSetErrorReason = "undeclared"
	// CallerSuppliedIdentifierSetMissing means a declared set was absent.
	CallerSuppliedIdentifierSetMissing CallerSuppliedIdentifierSetErrorReason = "missing"
	// CallerSuppliedIdentifierSetBelowMin means the set has too few items.
	CallerSuppliedIdentifierSetBelowMin CallerSuppliedIdentifierSetErrorReason = "below_minimum"
	// CallerSuppliedIdentifierSetAboveMax means the set has too many items.
	CallerSuppliedIdentifierSetAboveMax CallerSuppliedIdentifierSetErrorReason = "above_maximum"
	// CallerSuppliedIdentifierSetMalformed means an item failed its shape.
	CallerSuppliedIdentifierSetMalformed          CallerSuppliedIdentifierSetErrorReason = "malformed_element"
	CallerSuppliedIdentifierSetQueryConflict      CallerSuppliedIdentifierSetErrorReason = "query_conflict"
	CallerSuppliedIdentifierSetEmptyRequiredQuery CallerSuppliedIdentifierSetErrorReason = "required_query_empty"
)

func (e *CallerSuppliedIdentifierSetError) Error() string {
	switch e.Reason {
	case CallerSuppliedIdentifierSetUndeclared:
		return "caller-supplied identifier set is not declared for this operation"
	case CallerSuppliedIdentifierSetMissing:
		return fmt.Sprintf("caller-supplied identifier set %q is required", e.Parameter)
	case CallerSuppliedIdentifierSetBelowMin:
		return fmt.Sprintf("caller-supplied identifier set %q is below the minimum of %d elements", e.Parameter, e.Limit)
	case CallerSuppliedIdentifierSetAboveMax:
		return fmt.Sprintf("caller-supplied identifier set %q exceeds the maximum of %d elements", e.Parameter, e.Limit)
	case CallerSuppliedIdentifierSetMalformed:
		return fmt.Sprintf("caller-supplied identifier set %q element %d must match %s", e.Parameter, e.Position, e.ExpectedShape)
	case CallerSuppliedIdentifierSetQueryConflict:
		return fmt.Sprintf("caller-supplied identifier set %q cannot be combined with generic query input", e.Parameter)
	case CallerSuppliedIdentifierSetEmptyRequiredQuery:
		return fmt.Sprintf("caller-supplied identifier set %q cannot be empty because it is required to filter this operation", e.Parameter)
	default:
		return "caller-supplied identifier set is invalid"
	}
}

// completeEngineErrorText preserves the bounded diagnostic content captured by
// connsdk for direct-read callers. HTTPError.Error deliberately applies its
// own presentation policy, so the engine reads the typed fields here to retain
// the request URL and provider body that the operator needs to diagnose a
// declared connector action.
func completeEngineErrorText(err error) string {
	var httpErr *connsdk.HTTPError
	if !errors.As(err, &httpErr) {
		return err.Error()
	}
	message := strings.TrimSpace(httpErr.Body)
	if message == "" {
		message = http.StatusText(httpErr.Status)
	}
	return fmt.Sprintf("http %d for %s: %s", httpErr.Status, httpErr.URL, message)
}

func completeOperationDirectReadErrorText(err error, identifierSets map[string][]string) string {
	values := make([]string, 0)
	for _, identifiers := range identifierSets {
		values = append(values, identifiers...)
	}
	return redactSensitiveLiterals(completeEngineErrorText(err), values)
}

func OperationDirectRead(ctx context.Context, b Bundle, req connectors.OperationDirectReadRequest, h Hooks) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	op, err := findOperation(b, req.Operation)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	// provider_search shares this executor deliberately: its response bounding,
	// clamping, redaction and output-policy handling are the same as any other
	// bounded read. What differs is the stricter front half, which is enforced at
	// bundle load (validateProviderSearchSemantics), not here.
	if (op.Kind != "rest_read" && op.Kind != "provider_search") || op.REST == nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read requires rest_read or provider_search operation, got %q", op.Kind)
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if method != http.MethodGet && method != http.MethodPost {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read requires GET or POST, got %s", method)
	}
	if op.Kind == "provider_search" && method != http.MethodPost {
		return connectors.DirectReadResult{}, fmt.Errorf("provider search requires POST, got %s", method)
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read endpoint must be connector-relative, got absolute URL")
	}
	if method == http.MethodPost && !strings.EqualFold(strings.TrimSpace(op.REST.ContentType), "application/json") {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read POST requires application/json content_type")
	}
	if method == http.MethodPost && len(op.REST.BodySchema) == 0 {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read POST requires body_schema")
	}
	if op.REST.MaxBytes <= 0 {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read requires positive max_bytes")
	}
	if err := requireOperationDirectReadEndpoint(b, method, op.REST.Path); err != nil {
		return connectors.DirectReadResult{}, err
	}
	cfg := materializeConfigDefaults(b, req.Config)
	if err := rejectCallerSuppliedIdentifierSetQueryCollisions(op, req.Query); err != nil {
		return connectors.DirectReadResult{}, err
	}
	pathParams := cloneOperationPathParams(req.PathParams)
	queryMap := map[string]string{}
	for key, value := range op.REST.Query {
		queryMap[key] = value
	}
	for key, value := range req.Query {
		queryMap[key] = value
	}
	bodyOverrides := cloneAnyMap(req.Body)
	repeatedQuery, identifierSetPathNames, err := applyCallerSuppliedIdentifierSets(op, req.IdentifierSets, pathParams, queryMap, bodyOverrides)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if err := rejectEmptyCallerSuppliedIdentifierSetRequiredQueries(op, req.IdentifierSets, queryMap); err != nil {
		return connectors.DirectReadResult{}, err
	}
	resolvedPath, err := resolveOperationDirectReadPath(op.REST.Path, cfg, pathParams, identifierSetPathNames)
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
	for name, values := range repeatedQuery {
		if len(values) == 0 {
			continue
		}
		query[name] = cloneIdentifierSetValues(values)
	}
	policy := req.OutputPolicy
	if policy == "" {
		policy = op.OutputPolicy
	}
	if err := validateDirectReadOutputPolicy(policy, op.REST.Path, pathParams, cfg); err != nil {
		return connectors.DirectReadResult{}, err
	}
	body, err := operationReadBody(op, bodyOverrides)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDirectReadTimeout)
	defer cancel()
	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	maxBytes := clampOperationDirectReadMaxBytes(req.MaxBytes, op.REST.MaxBytes)
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, rt.Requester.BaseURL)
	if err := rejectCallerSuppliedIdentifierSetTargetBypass(b, cfg, rt.Requester.BaseURL, method, requestPath, pathParams, op.ID); err != nil {
		return connectors.DirectReadResult{}, err
	}
	resp, err := rt.Requester.DoLimited(ctx, method, requestPath, query, body, maxBytes)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		msg := completeOperationDirectReadErrorText(err, req.IdentifierSets)
		if hint != "" {
			msg = msg + ": " + hint
		}
		if class != "" {
			msg = class + ": " + msg
		}
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read %s %s: %s", method, op.REST.Path, msg)
	}
	if len(resp.Body) > maxBytes {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read response too large: %d bytes exceeds limit %d", len(resp.Body), maxBytes)
	}
	decoded, err := decodeDirectReadBody(resp.Body, maxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("operation direct read response is not JSON: %w", err)
	}
	decoded, err = applyDirectReadOutputPolicy(policy, decoded)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	redactFields := append([]string(nil), req.RedactFields...)
	if op.SensitivePolicy != nil {
		redactFields = append(redactFields, op.SensitivePolicy.RedactFields...)
	}
	if len(redactFields) > 0 {
		decoded = redactNamedJSONFields(decoded, redactFields)
	}
	return connectors.DirectReadResult{Connector: b.Name, Method: method, Path: resolvedPath, Status: resp.Status, Body: decoded}, nil
}

func cloneOperationPathParams(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for name, value := range in {
		out[name] = value
	}
	return out
}

func cloneIdentifierSetValues(values []string) []string {
	clone := make([]string, len(values))
	copy(clone, values)
	return clone
}

// applyCallerSuppliedIdentifierSets verifies each closed declaration before a
// request path, query, or body is assembled. It intentionally reports names,
// positions, shapes, and bounds only: identifier values can be account or
// wallet data and must never be reflected in a diagnostic.
func applyCallerSuppliedIdentifierSets(
	op OperationSpec,
	supplied map[string][]string,
	pathParams map[string]string,
	query map[string]string,
	body map[string]any,
) (map[string][]string, map[string]bool, error) {
	if op.REST == nil {
		return nil, nil, nil
	}
	declared := make(map[string]CallerSuppliedIdentifierSetSpec, len(op.REST.CallerSuppliedIdentifierSets))
	for _, set := range op.REST.CallerSuppliedIdentifierSets {
		declared[set.Name] = set
	}
	for name := range supplied {
		if _, ok := declared[name]; !ok {
			return nil, nil, &CallerSuppliedIdentifierSetError{Reason: CallerSuppliedIdentifierSetUndeclared}
		}
	}

	repeated := make(map[string][]string)
	pathNames := make(map[string]bool)
	for _, set := range op.REST.CallerSuppliedIdentifierSets {
		values, present := supplied[set.Name]
		if !present {
			return nil, nil, &CallerSuppliedIdentifierSetError{Parameter: set.Name, Reason: CallerSuppliedIdentifierSetMissing}
		}
		if len(values) < set.MinItems {
			return nil, nil, &CallerSuppliedIdentifierSetError{Parameter: set.Name, Reason: CallerSuppliedIdentifierSetBelowMin, Limit: set.MinItems}
		}
		if len(values) > set.MaxItems {
			return nil, nil, &CallerSuppliedIdentifierSetError{Parameter: set.Name, Reason: CallerSuppliedIdentifierSetAboveMax, Limit: set.MaxItems}
		}
		for position, value := range values {
			if err := validateCallerSuppliedIdentifierElement(set, position+1, value); err != nil {
				return nil, nil, err
			}
		}

		switch set.Wire {
		case "query_comma_separated":
			query[set.Name] = strings.Join(values, ",")
		case "query_repeated":
			repeated[set.Name] = cloneIdentifierSetValues(values)
			if len(values) > 0 {
				query[set.Name] = values[0]
			} else {
				query[set.Name] = ""
			}
		case "body_json_array":
			body[set.Name] = cloneIdentifierSetValues(values)
		case "path_segment":
			pathParams[set.Name] = values[0]
			pathNames[set.Name] = true
		default:
			return nil, nil, fmt.Errorf("operation %q has unsupported caller-supplied identifier wire", op.ID)
		}
	}
	return repeated, pathNames, nil
}

func rejectCallerSuppliedIdentifierSetQueryCollisions(op OperationSpec, query map[string]string) error {
	if op.REST == nil {
		return nil
	}
	for _, set := range op.REST.CallerSuppliedIdentifierSets {
		if set.Wire != "query_comma_separated" && set.Wire != "query_repeated" {
			continue
		}
		if _, present := query[set.Name]; present {
			return &CallerSuppliedIdentifierSetError{Parameter: set.Name, Reason: CallerSuppliedIdentifierSetQueryConflict}
		}
	}
	return nil
}

func rejectEmptyCallerSuppliedIdentifierSetRequiredQueries(op OperationSpec, supplied map[string][]string, query map[string]string) error {
	if op.REST == nil {
		return nil
	}
	for _, set := range op.REST.CallerSuppliedIdentifierSets {
		if set.Wire != "query_comma_separated" && set.Wire != "query_repeated" {
			continue
		}
		values, present := supplied[set.Name]
		if !present || len(values) != 0 {
			continue
		}
		for _, group := range op.REST.RequiredQuery {
			if !requiredQueryGroupIncludes(group, set.Name) || queryGroupSatisfiedExcluding(group, query, set.Name) {
				continue
			}
			return &CallerSuppliedIdentifierSetError{Parameter: set.Name, Reason: CallerSuppliedIdentifierSetEmptyRequiredQuery}
		}
	}
	return nil
}

func requiredQueryGroupIncludes(group RequiredQueryGroup, name string) bool {
	for _, candidate := range group.AnyOf {
		if strings.TrimSpace(candidate) == name {
			return true
		}
	}
	return false
}

func queryGroupSatisfiedExcluding(group RequiredQueryGroup, query map[string]string, excluded string) bool {
	for _, name := range group.AnyOf {
		name = strings.TrimSpace(name)
		if name == excluded {
			continue
		}
		if strings.TrimSpace(query[name]) != "" {
			return true
		}
	}
	return false
}

func validateCallerSuppliedIdentifierElement(set CallerSuppliedIdentifierSetSpec, position int, value string) error {
	invalid := func() error {
		return &CallerSuppliedIdentifierSetError{Parameter: set.Name, Reason: CallerSuppliedIdentifierSetMalformed, Position: position, ExpectedShape: set.ElementShape}
	}
	switch set.ElementShape {
	case "opaque_string":
		if strings.TrimSpace(value) == "" ||
			safety.RejectDangerousChars(value, "identifier set element") != nil ||
			(set.Wire == "query_comma_separated" && strings.Contains(value, ",")) {
			return invalid()
		}
	case "chain_address":
		if !chainAddressIdentifierSetPattern.MatchString(value) {
			return invalid()
		}
	default:
		return invalid()
	}
	return nil
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
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, rt.Requester.BaseURL)
	if err := rejectCallerSuppliedIdentifierSetTargetBypass(b, cfg, rt.Requester.BaseURL, method, requestPath, req.PathParams, ""); err != nil {
		return connectors.DirectReadResult{}, err
	}
	resp, err := rt.Requester.DoLimited(ctx, method, requestPath, query, nil, maxBytes)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		msg := completeEngineErrorText(err)
		if hint != "" {
			msg = msg + ": " + hint
		}
		if class != "" {
			msg = class + ": " + msg
		}
		return connectors.DirectReadResult{}, fmt.Errorf("direct read %s %s: %s", method, req.Path, msg)
	}

	if len(resp.Body) > maxBytes {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read response too large: %d bytes exceeds limit %d", len(resp.Body), maxBytes)
	}

	body, err := decodeDirectReadBody(resp.Body, maxBytes)
	if err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read response is not JSON: %w", err)
	}
	body, err = applyDirectReadOutputPolicy(req.OutputPolicy, body)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	if len(req.RedactFields) > 0 {
		body = redactNamedJSONFields(body, req.RedactFields)
	}
	return connectors.DirectReadResult{
		Connector: b.Name,
		Method:    method,
		Path:      resolvedPath,
		Status:    resp.Status,
		Body:      body,
	}, nil
}

func findOperation(b Bundle, id string) (OperationSpec, error) {
	for _, op := range b.Operations {
		if op.ID == id {
			return op, nil
		}
	}
	return OperationSpec{}, fmt.Errorf("operation %q not found in bundle %q", id, b.Name)
}

func operationDirectReadIdentifierSetWires(b Bundle, operation string) (map[string]string, error) {
	op, err := findOperation(b, operation)
	if err != nil {
		return nil, err
	}
	if op.Kind != "rest_read" || op.REST == nil {
		return nil, fmt.Errorf("operation %q does not declare a rest_read identifier-set contract", operation)
	}
	wires := make(map[string]string, len(op.REST.CallerSuppliedIdentifierSets))
	for _, set := range op.REST.CallerSuppliedIdentifierSets {
		wires[set.Name] = set.Wire
	}
	return wires, nil
}

func requireOperationDirectReadEndpoint(b Bundle, method, endpointPath string) error {
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

func operationReadBody(op OperationSpec, overrides map[string]any) (any, error) {
	if op.REST == nil || strings.ToUpper(strings.TrimSpace(op.REST.Method)) != http.MethodPost {
		return nil, nil
	}
	body := cloneAnyMap(op.REST.Body)
	for key, value := range overrides {
		body[key] = value
	}
	if len(op.REST.BodySchema) > 0 {
		sch, err := CompileSchema(op.REST.BodySchema)
		if err != nil {
			return nil, fmt.Errorf("operation %q: compile body_schema: %w", op.ID, err)
		}
		if err := sch.Validate(body); err != nil {
			return nil, fmt.Errorf("operation %q: body_schema: %w", op.ID, err)
		}
	}
	return body, nil
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
	var body any
	dec := json.NewDecoder(io.LimitReader(bytes.NewReader(raw), int64(maxBytes)+1))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
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
	case directReadPolicyJSONRedacted, directReadPolicyClinicalJSONRedacted:
		return nil
	default:
		return fmt.Errorf("direct read output policy %q is not supported", policy)
	}
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
		return redactJSONValue(body), nil
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

func rejectCallerSuppliedIdentifierSetTargetBypass(b Bundle, cfg connectors.RuntimeConfig, baseURL, method, requestPath string, candidatePathParams map[string]string, allowedOperation string) error {
	target, err := normalizedRequesterTargetPath(baseURL, requestPath)
	if err != nil {
		return nil
	}
	for _, op := range b.Operations {
		if op.ID == allowedOperation || op.Kind != "rest_read" || op.REST == nil || len(op.REST.CallerSuppliedIdentifierSets) == 0 || !strings.EqualFold(op.REST.Method, method) {
			continue
		}
		pattern, err := callerSuppliedIdentifierSetOperationTargetPattern(op, cfg, baseURL, candidatePathParams)
		if err != nil {
			continue
		}
		if callerSuppliedIdentifierSetTargetMatches(pattern, target) {
			return fmt.Errorf("request target is reserved for caller-supplied identifier-set operation %q", op.ID)
		}
	}
	return nil
}

func callerSuppliedIdentifierSetOperationTargetPattern(op OperationSpec, cfg connectors.RuntimeConfig, baseURL string, candidatePathParams map[string]string) (string, error) {
	identifierSetPathNames := map[string]bool{}
	for _, set := range op.REST.CallerSuppliedIdentifierSets {
		if set.Wire == "path_segment" {
			identifierSetPathNames[set.Name] = true
		}
	}
	pathParams := map[string]string{}
	for _, match := range surfacePathVarPattern.FindAllStringSubmatch(op.REST.Path, -1) {
		name := match[1]
		if identifierSetPathNames[name] {
			pathParams[name] = callerSuppliedIdentifierSetTargetSegmentMarker
			continue
		}
		if value := candidatePathParams[name]; value != "" {
			pathParams[name] = value
			continue
		}
		if value := cfg.Config[name]; value != "" {
			pathParams[name] = value
			continue
		}
		if name == "path" {
			pathParams[name] = callerSuppliedIdentifierSetTargetPathMarker
			continue
		}
		pathParams[name] = callerSuppliedIdentifierSetTargetSegmentMarker
	}
	resolved, err := resolveOperationDirectReadPath(op.REST.Path, cfg, pathParams, identifierSetPathNames)
	if err != nil {
		return "", err
	}
	return normalizedRequesterTargetPath(baseURL, normalizeDirectReadPathForBaseURL(resolved, baseURL))
}

func normalizedRequesterTargetPath(baseURL, requestPath string) (string, error) {
	raw := requestPath
	if !strings.HasPrefix(requestPath, "http://") && !strings.HasPrefix(requestPath, "https://") {
		raw = strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(requestPath, "/")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	path := parsed.EscapedPath()
	if path == "" {
		path = "/"
	}
	target := canonicalizeRequestTargetPercentEscapes(path)
	if parsed.Scheme == "" && parsed.Host == "" {
		return target, nil
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host) + target, nil
}

func callerSuppliedIdentifierSetTargetMatches(pattern, target string) bool {
	expression := regexp.QuoteMeta(pattern)
	expression = strings.ReplaceAll(expression, regexp.QuoteMeta(callerSuppliedIdentifierSetTargetSegmentMarker), `[^/]+`)
	expression = strings.ReplaceAll(expression, regexp.QuoteMeta(callerSuppliedIdentifierSetTargetPathMarker), `.+`)
	return regexp.MustCompile("^" + expression + "$").MatchString(target)
}

func canonicalizeRequestTargetPercentEscapes(path string) string {
	var canonical strings.Builder
	canonical.Grow(len(path))
	for i := 0; i < len(path); i++ {
		if path[i] != '%' || i+2 >= len(path) {
			canonical.WriteByte(path[i])
			continue
		}
		high, highOK := requestTargetHexValue(path[i+1])
		low, lowOK := requestTargetHexValue(path[i+2])
		if !highOK || !lowOK {
			canonical.WriteByte(path[i])
			continue
		}
		value := high<<4 | low
		if isRequestTargetUnreserved(value) {
			canonical.WriteByte(value)
		} else {
			canonical.WriteByte('%')
			canonical.WriteByte(uppercaseRequestTargetHex(path[i+1]))
			canonical.WriteByte(uppercaseRequestTargetHex(path[i+2]))
		}
		i += 2
	}
	return canonical.String()
}

func requestTargetHexValue(value byte) (byte, bool) {
	switch {
	case value >= '0' && value <= '9':
		return value - '0', true
	case value >= 'a' && value <= 'f':
		return value - 'a' + 10, true
	case value >= 'A' && value <= 'F':
		return value - 'A' + 10, true
	default:
		return 0, false
	}
}

func isRequestTargetUnreserved(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z') || (value >= '0' && value <= '9') || strings.ContainsRune("-._~", rune(value))
}

func uppercaseRequestTargetHex(value byte) byte {
	if value >= 'a' && value <= 'f' {
		return value - ('a' - 'A')
	}
	return value
}

func redactJSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, item := range v {
			if item != nil && shouldRedactJSONField(key) {
				out[key+"_redacted"] = true
				continue
			}
			out[key] = redactJSONValue(item)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = redactJSONValue(item)
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

func shouldRedactJSONField(name string) bool {
	normalized := normalizeJSONFieldName(name)
	switch normalized {
	case "content", "body", "payload", "raw", "download_url", "download_media_url", "clone_url", "api_key", "apikey", "access_key", "private_key", "authorization", "credential", "credentials":
		return true
	}
	if strings.Contains(normalized, "download") && strings.Contains(normalized, "url") {
		return true
	}
	if strings.Contains(normalized, "clone") && strings.Contains(normalized, "url") {
		return true
	}
	for _, marker := range []string{"token", "secret", "password", "private_key", "api_key", "apikey", "access_key", "authorization", "credential"} {
		if strings.Contains(normalized, marker) {
			return true
		}
	}
	return false
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
			return errors.New("repository path is blocked by direct read output policy")
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
	return resolveSurfaceEndpointPathWithIdentifierSets(template, cfg, pathParams, nil)
}

func resolveOperationDirectReadPath(template string, cfg connectors.RuntimeConfig, pathParams map[string]string, identifierSetPathNames map[string]bool) (string, error) {
	return resolveSurfaceEndpointPathWithIdentifierSets(template, cfg, pathParams, identifierSetPathNames)
}

func resolveSurfaceEndpointPathWithIdentifierSets(template string, cfg connectors.RuntimeConfig, pathParams map[string]string, identifierSetPathNames map[string]bool) (string, error) {
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
		if identifierSetPathNames[name] {
			encoded = url.PathEscape(value)
			err = nil
		}
		if err != nil {
			firstErr = err
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
	if containsDotDotSegment(resolved) {
		return "", errors.New("direct read endpoint path contains path traversal")
	}
	return resolved, nil
}

func encodeSurfacePathValue(name, value string) (string, error) {
	if name == "path" {
		if strings.Contains(value, "\\") {
			return "", fmt.Errorf("path variable %q must use forward slashes", name)
		}
		if err := safety.ValidateRelativePath(value, "path variable "+name); err != nil {
			return "", err
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
	if err := safety.ValidateIdentifier(value, "path variable "+name); err != nil {
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
		values.Set(name, value)
	}
	return values, nil
}

func isAbsoluteHTTPURL(raw string) bool {
	lower := strings.ToLower(strings.TrimSpace(raw))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}
