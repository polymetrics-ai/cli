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
)

var surfacePathVarPattern = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)

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

func OperationDirectRead(ctx context.Context, b Bundle, req connectors.OperationDirectReadRequest, h Hooks) (connectors.DirectReadResult, error) {
	if err := ctx.Err(); err != nil {
		return connectors.DirectReadResult{}, err
	}
	op, err := operationDirectReadSpec(b, req.Operation)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	cfg := materializeConfigDefaults(b, req.Config)
	resolvedPath, err := resolveSurfaceEndpointPath(op.REST.Path, cfg, req.PathParams)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	queryMap := map[string]string{}
	for key, value := range op.REST.Query {
		queryMap[key] = value
	}
	for key, value := range req.Query {
		queryMap[key] = value
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
	body, err := operationReadBody(op, req.Body)
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
	requestPath := normalizeDirectReadPathForBaseURL(resolvedPath, directReadBaseURL(b, cfg))
	decoded, pageInfo, resp, err := readDirectPage(ctx, b, rt, directReadWalk{
		method:      method,
		declaredPat: op.REST.Path,
		requestPath: requestPath,
		query:       query,
		body:        body,
		maxBytes:    maxBytes,
		page:        req.Page,
		pageCursor:  req.PageCursor,
	})
	if err != nil {
		var tooLarge errDirectReadTooLarge
		if errors.As(err, &tooLarge) {
			return connectors.DirectReadResult{}, fmt.Errorf("operation direct read %s", tooLarge.Error())
		}
		// A pagination failure arrives with the response already fetched and
		// decoded, so it must not fall through to the not-JSON branch below.
		var pageErr errDirectReadPagination
		if errors.As(err, &pageErr) {
			return connectors.DirectReadResult{}, fmt.Errorf("operation direct read pagination: %w", pageErr.err)
		}
		if resp == nil {
			class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
			msg := completeEngineErrorText(err)
			if hint != "" {
				msg = msg + ": " + hint
			}
			if class != "" {
				msg = class + ": " + msg
			}
			return connectors.DirectReadResult{}, fmt.Errorf("operation direct read %s %s: %s", method, op.REST.Path, msg)
		}
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
	return connectors.DirectReadResult{Connector: b.Name, Method: method, Path: resolvedPath, Status: resp.Status, Body: decoded, Page: pageInfo}, nil
}

// PreflightOperationDirectRead proves a command's named operation can reach
// the bounded direct-read executor with its declared endpoint, response cap,
// and output policy. It shares operationDirectReadSpec with execution and is
// deliberately no-network: preflight must be safe to run across every bundle.
func PreflightOperationDirectRead(b Bundle, operation, method, endpointPath string, maxBytes int, outputPolicy string) error {
	op, err := operationDirectReadSpec(b, operation)
	if err != nil {
		return err
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

// operationDirectReadSpec is the static, no-network part of the bounded
// direct-read executor. Both execution and command preflight call it so their
// eligibility rules cannot drift apart.
func operationDirectReadSpec(b Bundle, operation string) (OperationSpec, error) {
	op, err := findOperation(b, operation)
	if err != nil {
		return OperationSpec{}, err
	}
	// provider_search shares this executor deliberately: its response bounding,
	// clamping, redaction and output-policy handling are the same as any other
	// bounded read. What differs is its stricter bundle-load contract.
	if (op.Kind != "rest_read" && op.Kind != "provider_search") || op.REST == nil {
		return OperationSpec{}, fmt.Errorf("operation direct read requires rest_read or provider_search operation, got %q", op.Kind)
	}
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if method != http.MethodGet && method != http.MethodPost && method != http.MethodHead {
		return OperationSpec{}, fmt.Errorf("operation direct read requires GET, POST, or HEAD, got %s", method)
	}
	if op.Kind == "provider_search" && method != http.MethodPost {
		return OperationSpec{}, fmt.Errorf("provider search requires POST, got %s", method)
	}
	if op.Kind == "rest_read" && method == http.MethodHead && op.OutputPolicy != directReadPolicyJSONRedacted {
		return OperationSpec{}, fmt.Errorf("operation direct read HEAD (status-only, no response body) requires output_policy %q, got %q", directReadPolicyJSONRedacted, op.OutputPolicy)
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		return OperationSpec{}, fmt.Errorf("operation direct read endpoint must be connector-relative, got absolute URL")
	}
	if method == http.MethodPost && !strings.EqualFold(strings.TrimSpace(op.REST.ContentType), "application/json") {
		return OperationSpec{}, fmt.Errorf("operation direct read POST requires application/json content_type")
	}
	if method == http.MethodPost && len(op.REST.BodySchema) == 0 {
		return OperationSpec{}, fmt.Errorf("operation direct read POST requires body_schema")
	}
	if method == http.MethodPost {
		if _, err := CompileSchema(op.REST.BodySchema); err != nil {
			return OperationSpec{}, fmt.Errorf("operation direct read body_schema: %w", err)
		}
	}
	if op.REST.MaxBytes <= 0 {
		return OperationSpec{}, fmt.Errorf("operation direct read requires positive max_bytes")
	}
	if err := requireOperationDirectReadLedgerEndpoint(b, op.Kind, method, op.REST.Path, op.REST.MaxBytes); err != nil {
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
		method:      method,
		declaredPat: req.Path,
		requestPath: requestPath,
		query:       query,
		maxBytes:    maxBytes,
		page:        req.Page,
		pageCursor:  req.PageCursor,
	})
	if err != nil {
		var tooLarge errDirectReadTooLarge
		if errors.As(err, &tooLarge) {
			return connectors.DirectReadResult{}, fmt.Errorf("direct read %s", tooLarge.Error())
		}
		// A pagination failure arrives with the response already fetched and
		// decoded, so it must not fall through to the not-JSON branch below.
		var pageErr errDirectReadPagination
		if errors.As(err, &pageErr) {
			return connectors.DirectReadResult{}, fmt.Errorf("direct read pagination: %w", pageErr.err)
		}
		if resp == nil {
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
		Page:      pageInfo,
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

func requireOperationDirectReadLedgerEndpoint(b Bundle, kind, method, endpointPath string, maxBytes int) error {
	ledger := operationDirectReadEndpointLedger(b)
	if ledger == nil {
		return fmt.Errorf("runtime operation endpoint ledger is unavailable for bundle %q", b.Name)
	}
	for _, entry := range ledger.entries {
		if strings.EqualFold(entry.Method, method) && entry.Path == endpointPath && entry.Kind == kind && entry.MaxBytes == maxBytes {
			return nil
		}
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

// decodeDirectReadPageBody turns one fetched page into the value the direct-read
// pipeline works over. A HEAD response never carries a body (RFC 9110 §9.3.2) —
// there is nothing to JSON-decode. The provider's status code is itself the
// entire signal a status-only "check" command exists to surface;
// operationDirectReadSpec already requires a HEAD-shaped operation to declare
// output_policy "json_redacted", so the redaction/policy pipeline still runs
// uniformly over this small synthetic object rather than special-casing HEAD
// out of it. It lives beside the fetch rather than in OperationDirectRead
// because decoding is what readDirectPage owns.
func decodeDirectReadPageBody(method string, resp *connsdk.Response, maxBytes int) (any, error) {
	if strings.EqualFold(strings.TrimSpace(method), http.MethodHead) {
		return map[string]any{"status_code": resp.Status}, nil
	}
	return decodeDirectReadBody(resp.Body, maxBytes)
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
	if err := safety.ValidateURLPathSegment(value, "path variable "+name); err != nil {
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
