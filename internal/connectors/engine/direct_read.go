package engine

import (
	"bytes"
	"context"
	"encoding/base64"
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
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

const (
	defaultDirectReadMaxBytes                  = 1 << 20
	defaultDirectReadTimeout                   = 30 * time.Second
	directReadPolicyGitHubContentsFileMetadata = "github_contents_file_metadata"
	directReadPolicyGitHubContentsDirectory    = "github_contents_directory"
	directReadPolicyBitbucketJSONObject        = "bitbucket_json_object"
	directReadPolicyBitbucketJSONCollection    = "bitbucket_json_collection"
	directReadPolicyBitbucketBinaryBase64      = "bitbucket_binary_base64"
)

var surfacePathVarPattern = regexp.MustCompile(`\{([A-Za-z_][A-Za-z0-9_]*)\}`)

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
	if err := validateDirectReadOutputPolicy(req.OutputPolicy, req.PathParams); err != nil {
		return connectors.DirectReadResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDirectReadTimeout)
	defer cancel()

	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	maxBytes := clampDirectReadMaxBytes(req.MaxBytes)
	resp, err := rt.Requester.DoLimited(ctx, method, resolvedPath, query, nil, maxBytes)
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

	var body any
	if req.OutputPolicy == directReadPolicyBitbucketBinaryBase64 {
		body = applyBinaryDirectReadOutputPolicy(resp)
	} else {
		dec := json.NewDecoder(io.LimitReader(bytes.NewReader(resp.Body), int64(maxBytes)+1))
		dec.UseNumber()
		if err := dec.Decode(&body); err != nil {
			return connectors.DirectReadResult{}, fmt.Errorf("direct read response is not JSON: %w", err)
		}
		body, err = applyDirectReadOutputPolicy(req.OutputPolicy, body)
		if err != nil {
			return connectors.DirectReadResult{}, err
		}
	}
	return connectors.DirectReadResult{
		Connector: b.Name,
		Method:    method,
		Path:      resolvedPath,
		Status:    resp.Status,
		Body:      body,
	}, nil
}

func clampDirectReadMaxBytes(maxBytes int) int {
	if maxBytes <= 0 || maxBytes > defaultDirectReadMaxBytes {
		return defaultDirectReadMaxBytes
	}
	return maxBytes
}

func validateDirectReadOutputPolicy(policy string, pathParams map[string]string) error {
	switch policy {
	case directReadPolicyGitHubContentsFileMetadata, directReadPolicyGitHubContentsDirectory:
		if err := rejectSensitiveRepositoryPath(pathParams["path"]); err != nil {
			return err
		}
		return nil
	case directReadPolicyBitbucketJSONObject, directReadPolicyBitbucketJSONCollection,
		directReadPolicyBitbucketBinaryBase64:
		return nil
	default:
		return fmt.Errorf("direct read output policy %q is not supported", policy)
	}
}

func applyDirectReadOutputPolicy(policy string, body any) (any, error) {
	switch policy {
	case directReadPolicyGitHubContentsFileMetadata:
		obj, ok := body.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("direct read output policy %q requires a file metadata object", policy)
		}
		if typ, _ := obj["type"].(string); typ == "dir" {
			return nil, fmt.Errorf("direct read output policy %q received a directory response", policy)
		}
		return redactGitHubContentsObject(obj), nil
	case directReadPolicyGitHubContentsDirectory:
		items, ok := body.([]any)
		if !ok {
			return nil, fmt.Errorf("direct read output policy %q requires a directory listing array", policy)
		}
		out := make([]any, 0, len(items))
		for _, item := range items {
			if obj, ok := item.(map[string]any); ok {
				out = append(out, redactGitHubContentsObject(obj))
				continue
			}
			out = append(out, item)
		}
		return out, nil
	case directReadPolicyBitbucketJSONObject:
		obj, ok := body.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("direct read output policy %q requires a JSON object", policy)
		}
		return redactBitbucketJSON(obj).(map[string]any), nil
	case directReadPolicyBitbucketJSONCollection:
		return redactBitbucketJSON(body), nil
	default:
		return nil, fmt.Errorf("direct read output policy %q is not supported", policy)
	}
}

func applyBinaryDirectReadOutputPolicy(resp *connsdk.Response) map[string]any {
	contentType := resp.Header.Get("Content-Type")
	if idx := strings.Index(contentType, ";"); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}
	return map[string]any{
		"encoding":       "base64",
		"bytes":          len(resp.Body),
		"content_type":   contentType,
		"content_base64": base64.StdEncoding.EncodeToString(resp.Body),
	}
}

func redactGitHubContentsObject(in map[string]any) map[string]any {
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

func redactBitbucketJSON(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed)+2)
		for k, v := range typed {
			lower := strings.ToLower(k)
			if shouldRedactBitbucketKey(lower) {
				out[k+"_redacted"] = true
				continue
			}
			if lower == "links" {
				if links, ok := v.(map[string]any); ok {
					out[k] = redactBitbucketLinks(links)
					continue
				}
			}
			out[k] = redactBitbucketJSON(v)
		}
		return out
	case []any:
		out := make([]any, 0, len(typed))
		for _, item := range typed {
			out = append(out, redactBitbucketJSON(item))
		}
		return out
	default:
		return value
	}
}

func redactBitbucketLinks(in map[string]any) map[string]any {
	out := make(map[string]any, len(in)+2)
	for k, v := range in {
		switch strings.ToLower(k) {
		case "clone", "download", "downloads":
			out[k+"_redacted"] = true
		default:
			out[k] = redactBitbucketJSON(v)
		}
	}
	return out
}

func shouldRedactBitbucketKey(lower string) bool {
	for _, marker := range []string{"token", "secret", "password", "private_key", "access_token", "client_secret"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	switch lower {
	case "content", "raw", "patch", "diff", "key", "key_pair", "private_key", "value":
		return true
	default:
		return false
	}
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
