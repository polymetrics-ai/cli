package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	stdpath "path"
	"regexp"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors"
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
	resp, err := rt.Requester.DoLimited(ctx, method, requestPath, query, body, maxBytes)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		msg := safety.RedactErrorText(err.Error())
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
	resp, err := rt.Requester.DoLimited(ctx, method, requestPath, query, nil, maxBytes)
	if err != nil {
		class, hint := applyErrorMap(b.HTTP.ErrorMap, err)
		msg := safety.RedactErrorText(err.Error())
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
