package engine

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
)

// MissingOperationRouteError is the fail-closed routing diagnostic returned
// before a provider request can be built. It carries the source URL so the
// command runner can retain its blocked-command presentation without parsing
// an opaque URL construction error.
type MissingOperationRouteError struct {
	Connector string
	Operation string
	Route     string
	SourceURL string
	Reason    string
}

func (e *MissingOperationRouteError) Error() string {
	parts := []string{fmt.Sprintf("connector %q operation %q is blocked", e.Connector, e.Operation), "missing route foundation"}
	if e.Route != "" {
		parts = append(parts, fmt.Sprintf("route=%q", e.Route))
	}
	if e.Reason != "" {
		parts = append(parts, e.Reason)
	}
	if e.SourceURL != "" {
		parts = append(parts, "source="+e.SourceURL)
	}
	return strings.Join(parts, ": ")
}

func operationRouteFailure(b Bundle, operation, route, sourceURL, reason string) error {
	return &MissingOperationRouteError{
		Connector: b.Name,
		Operation: operation,
		Route:     route,
		SourceURL: sourceURL,
		Reason:    reason,
	}
}

// validateOperationRoutes keeps route names unambiguous at load time. A
// duplicate name is a conflict: choosing either declaration would make the
// provider destination depend on declaration order.
func validateOperationRoutes(b Bundle, streams []StreamSpec, writes []WriteAction, operations []OperationSpec) error {
	routes, err := declaredOperationRoutes(b)
	if err != nil {
		return err
	}

	for _, stream := range streams {
		if err := validateOperationRouteSelection(b, routes, stream.Route, stream.Name, stream.Path, ""); err != nil {
			return err
		}
	}
	for _, action := range writes {
		if strings.TrimSpace(action.Route) != "" && strings.TrimSpace(action.BaseURL) != "" {
			return fmt.Errorf("write action %q declares both route and base_url", action.Name)
		}
		if err := validateOperationRouteSelection(b, routes, action.Route, action.Name, action.Path, ""); err != nil {
			return err
		}
	}
	for _, operation := range operations {
		path := operationRoutePath(operation)
		if err := validateOperationRouteSelection(b, routes, operation.Route, operation.ID, path, operation.SourceURL); err != nil {
			return err
		}
	}
	return nil
}

// declaredOperationRoutes validates and indexes only connector-controlled
// route declarations. It is used both while loading a bundle and by the
// no-network preflight helpers, so a manually assembled Bundle cannot bypass
// the same closed routing contract the JSON loader enforces.
func declaredOperationRoutes(b Bundle) (map[string]OperationRouteSpec, error) {
	routes := make(map[string]OperationRouteSpec, len(b.HTTP.Routes))
	for i, route := range b.HTTP.Routes {
		name := strings.TrimSpace(route.Name)
		if name == "" {
			return nil, fmt.Errorf("base route %d requires name", i)
		}
		if name != route.Name {
			return nil, fmt.Errorf("base route %d name %q must not contain surrounding whitespace", i, route.Name)
		}
		if err := validateOperationRouteBase(route.BaseURL); err != nil {
			return nil, fmt.Errorf("base route %q: %w", name, err)
		}
		if err := validateOperationRouteVersion(route.Version); err != nil {
			return nil, fmt.Errorf("base route %q: %w", name, err)
		}
		if prior, exists := routes[name]; exists {
			if prior.BaseURL != route.BaseURL || prior.Version != route.Version {
				return nil, fmt.Errorf("base route %q declares conflicting bases", name)
			}
			return nil, fmt.Errorf("base route %q is declared more than once", name)
		}
		routes[name] = route
	}

	return routes, nil
}

func validateOperationRouteSelection(b Bundle, routes map[string]OperationRouteSpec, selection, operation, path, sourceURL string) error {
	selection = strings.TrimSpace(selection)
	if selection == "" {
		return nil
	}
	route, ok := routes[selection]
	if !ok {
		return operationRouteFailure(b, operation, selection, sourceURL, "route is not declared by streams.base.routes")
	}
	if version := strings.TrimSpace(route.Version); version != "" && !strings.HasPrefix(path, "/"+version+"/") && path != "/"+version {
		return operationRouteFailure(b, operation, selection, sourceURL, fmt.Sprintf("version %q does not match declared path %q", version, path))
	}
	return nil
}

func validateOperationRouteForOperation(b Bundle, selection, operation, path, sourceURL string) error {
	routes, err := declaredOperationRoutes(b)
	if err != nil {
		return err
	}
	return validateOperationRouteSelection(b, routes, selection, operation, path, sourceURL)
}

func validateOperationRouteBase(baseURL string) error {
	if baseURL == "{{ config.base_url }}" {
		return nil
	}
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(baseURL) != baseURL || strings.Contains(baseURL, "{{") {
		return fmt.Errorf("base_url must be one fixed absolute HTTP origin or {{ config.base_url }}")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return fmt.Errorf("base_url must be one fixed absolute HTTP origin or {{ config.base_url }}")
	}
	return nil
}

func validateOperationRouteVersion(version string) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return nil
	}
	if strings.Contains(version, "/") || strings.ContainsAny(version, "?#") {
		return fmt.Errorf("version must be one path segment")
	}
	return nil
}

func operationRoutePath(operation OperationSpec) string {
	if operation.REST != nil {
		return operation.REST.Path
	}
	if operation.GraphQL != nil {
		return operation.GraphQL.Path
	}
	if operation.Binary != nil {
		return operation.Binary.Path
	}
	return ""
}

func resolveOperationRoute(b Bundle, cfg connectors.RuntimeConfig, selection, operation, path, sourceURL string) (string, error) {
	selection = strings.TrimSpace(selection)
	if selection == "" {
		baseURL, err := Interpolate(b.HTTP.URL, requestVars(cfg, nil, ""))
		if err != nil {
			return "", operationRouteFailure(b, operation, "", sourceURL, "resolve declared default base URL")
		}
		if !isAbsoluteHTTPURL(baseURL) {
			return "", operationRouteFailure(b, operation, "", sourceURL, "declared default base URL is invalid")
		}
		return baseURL, nil
	}

	routes, err := declaredOperationRoutes(b)
	if err != nil {
		return "", operationRouteFailure(b, operation, selection, sourceURL, err.Error())
	}
	if err := validateOperationRouteSelection(b, routes, selection, operation, path, sourceURL); err != nil {
		return "", err
	}
	route := routes[selection]
	baseURL, err := Interpolate(route.BaseURL, requestVars(cfg, nil, ""))
	if err != nil {
		return "", operationRouteFailure(b, operation, selection, sourceURL, "resolve declared route base URL")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", operationRouteFailure(b, operation, selection, sourceURL, "declared route base URL is invalid")
	}
	// A version route deliberately selects a provider origin. Its version stays
	// in the source-locked operation path, so a configured v2 base can never
	// become /v2/v3 when a v3 operation is selected.
	if strings.TrimSpace(route.Version) != "" {
		parsed.Path = ""
		parsed.RawPath = ""
		parsed.RawQuery = ""
		parsed.Fragment = ""
	}
	return strings.TrimRight(parsed.String(), "/"), nil
}

func resolveStreamRoute(b Bundle, cfg connectors.RuntimeConfig, stream StreamSpec) (string, error) {
	return resolveOperationRoute(b, cfg, stream.Route, stream.Name, stream.Path, "")
}

func resolveWriteActionRoute(b Bundle, cfg connectors.RuntimeConfig, action WriteAction) (string, error) {
	if strings.TrimSpace(action.Route) != "" && strings.TrimSpace(action.BaseURL) != "" {
		return "", operationRouteFailure(b, action.Name, action.Route, "", "write action declares both route and base_url")
	}
	if strings.TrimSpace(action.Route) != "" {
		return resolveOperationRoute(b, cfg, action.Route, action.Name, action.Path, "")
	}
	if strings.TrimSpace(action.BaseURL) != "" {
		if err := validateWriteActionBaseURL(-1, action); err != nil {
			return "", err
		}
		baseURL, err := Interpolate(action.BaseURL, requestVars(cfg, nil, ""))
		if err != nil {
			return "", fmt.Errorf("engine: resolve write action base URL: %w", err)
		}
		return baseURL, nil
	}
	return resolveOperationRoute(b, cfg, "", action.Name, action.Path, "")
}

func newRuntimeForOperationRoute(ctx context.Context, b Bundle, cfg connectors.RuntimeConfig, h Hooks, selection, operation, path, sourceURL string) (*Runtime, error) {
	if err := ValidateExplicitEmptyRequiredSecretFields(b, cfg.Secrets); err != nil {
		return nil, fmt.Errorf("engine: %w", err)
	}
	baseURL, err := resolveOperationRoute(b, cfg, selection, operation, path, sourceURL)
	if err != nil {
		return nil, err
	}
	headers, err := resolveHeaders(b.HTTP.Headers, cfg, b.Spec)
	if err != nil {
		return nil, err
	}
	return newRuntimeWithResolvedHTTP(ctx, b, cfg, h, baseURL, headers)
}
