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
	defaultDirectReadMaxBytes = 1 << 20
	defaultDirectReadTimeout  = 30 * time.Second
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
	resolvedPath, err := resolveSurfaceEndpointPath(req.Path, req.Config, req.PathParams)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}
	query, err := directReadQuery(req.Query)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	cfg := materializeConfigDefaults(b, req.Config)
	rt, err := newRuntime(ctx, b, cfg, h)
	if err != nil {
		return connectors.DirectReadResult{}, err
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDirectReadTimeout)
	defer cancel()

	resp, err := rt.Requester.Do(ctx, method, resolvedPath, query, nil)
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

	maxBytes := req.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultDirectReadMaxBytes
	}
	if len(resp.Body) > maxBytes {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read response too large: %d bytes exceeds limit %d", len(resp.Body), maxBytes)
	}

	var body any
	dec := json.NewDecoder(io.LimitReader(bytes.NewReader(resp.Body), int64(maxBytes)+1))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		return connectors.DirectReadResult{}, fmt.Errorf("direct read response is not JSON: %w", err)
	}
	return connectors.DirectReadResult{
		Connector: b.Name,
		Method:    method,
		Path:      resolvedPath,
		Status:    resp.Status,
		Body:      body,
	}, nil
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
			if ep.Operation == nil || ep.Operation.Model != "direct_read" {
				return fmt.Errorf("api_surface endpoint %s %s is not a direct_read operation", method, endpointPath)
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
