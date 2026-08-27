package engine

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"polymetrics.ai/internal/connectors"
)

var (
	commandBindingTemplateRE   = regexp.MustCompile(`\{\{\s*(?:config|record|fanout)\.([-A-Za-z0-9_]+)(?:\s*\|[^}]*)?\s*\}\}`)
	commandBindingSlotRE       = regexp.MustCompile(`\{[-A-Za-z0-9_]+\}`)
	commandBindingAnnotationRE = regexp.MustCompile(`\s+\((?:body|object|parent)=[-A-Za-z0-9_]+\)$`)
	commandBindingBaseURLRE    = regexp.MustCompile(`\{\{\s*config\.base_url\s*\}\}`)
)

const (
	CommandEndpointExact                         = "exact"
	CommandEndpointBasePath                      = "declared_base_path"
	CommandEndpointPlaceholder                   = "placeholder_identity"
	CommandEndpointHookTransport                 = "registered_hook_transport"
	CommandEndpointGraphQLTransport              = "graphql_operation_transport"
	CommandEndpointAbsoluteTransport             = "absolute_url_transport"
	CommandEndpointQueryTransport                = "declared_query_transport"
	CommandEndpointSuffixTransport               = "provider_suffix_transport"
	CommandEndpointAnnotationIdentity            = "operation_annotation_identity"
	CommandEndpointCompositeProviderPathIdentity = "composite_provider_path_identity"
)

// ResolvedCommandBinding is the one declaration and provider target selected
// by the runtime for an implemented CLI command. Method/Path are the canonical
// discovery identity. TransportMethod/TransportPath retain what the executor
// actually sends; Equivalence names the closed proof relating the two.
type ResolvedCommandBinding struct {
	Binding         connectors.CommandBindingIdentity
	Method          string
	Path            string
	TransportMethod string
	TransportPath   string
	Equivalence     string
	Destructive     DestructiveTarget
}

type commandRuntimeEndpoint struct {
	method           string
	path             string
	route            string
	baseURL          string
	graphqlOperation string
}

// ResolveImplementedCommandBinding is shared by runtime preflight and source
// admission. It is the sole lane-to-declaration resolver; certification must
// not copy its own incomplete stream/write/operation switch.
func ResolveImplementedCommandBinding(b Bundle, cmd connectors.CommandSurfaceCommand) (ResolvedCommandBinding, error) {
	return resolveImplementedCommandBinding(b, cmd, HooksFor(b.Name))
}

func resolveImplementedCommandBinding(b Bundle, cmd connectors.CommandSurfaceCommand, hooks Hooks) (ResolvedCommandBinding, error) {
	bindings := 0
	for _, value := range []string{cmd.Stream, cmd.Write, cmd.Operation} {
		if strings.TrimSpace(value) != "" {
			bindings++
		}
	}
	if bindings > 1 {
		return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references more than one runtime binding", cmd.Path)
	}

	var (
		resolved ResolvedCommandBinding
		runtime  commandRuntimeEndpoint
	)
	switch {
	case cmd.Stream != "":
		stream, ok := commandBindingStream(b, cmd.Stream)
		if !ok {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references missing stream %q", cmd.Path, cmd.Stream)
		}
		method := strings.ToUpper(strings.TrimSpace(stream.Method))
		if method == "" {
			method = http.MethodGet
		}
		resolved.Binding = connectors.CommandBindingIdentity{Kind: connectors.CommandBindingStream, ID: stream.Name}
		runtime = commandRuntimeEndpoint{method: method, path: normalizeBindingPathTemplate(stream.Path), route: stream.Route}
		if stream.GraphQL != nil {
			runtime.graphqlOperation = strings.TrimSpace(stream.GraphQL.OperationName)
		}
	case cmd.Write != "":
		action, ok := commandBindingWrite(b, cmd.Write)
		if !ok {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references missing write action %q", cmd.Path, cmd.Write)
		}
		resolved.Binding = connectors.CommandBindingIdentity{Kind: connectors.CommandBindingWrite, ID: action.Name}
		resolved.Destructive = DestructiveTargetForWrite(b.Name, action)
		runtime = commandRuntimeEndpoint{
			method: strings.ToUpper(strings.TrimSpace(action.Method)), path: normalizeBindingPathTemplate(action.Path),
			route: action.Route, baseURL: action.BaseURL,
		}
	case cmd.Operation != "":
		operation, ok := commandBindingOperation(b, cmd.Operation)
		if !ok {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q references missing operation %q", cmd.Path, cmd.Operation)
		}
		method, path, err := commandBindingOperationEndpoint(operation)
		if err != nil {
			return ResolvedCommandBinding{}, err
		}
		resolved.Binding = connectors.CommandBindingIdentity{Kind: connectors.CommandBindingOperation, ID: operation.ID}
		resolved.Destructive = DestructiveTargetForOperation(b.Name, operation)
		runtime = commandRuntimeEndpoint{method: method, path: normalizeBindingPathTemplate(path), route: operation.Route}
		if operation.GraphQL != nil {
			runtime.graphqlOperation = strings.TrimSpace(operation.GraphQL.OperationName)
		}
	default:
		if cmd.Intent != "direct_read" || len(cmd.APISurface) != 1 {
			return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q has no resolvable runtime binding", cmd.Path)
		}
		reference := cmd.APISurface[0]
		resolved.Binding = connectors.CommandBindingIdentity{Kind: connectors.CommandBindingCommand, ID: cmd.Path}
		resolved.Method = strings.ToUpper(strings.TrimSpace(reference.Method))
		resolved.Path = reference.Path
		resolved.TransportMethod = resolved.Method
		resolved.TransportPath = resolved.Path
		resolved.Equivalence = CommandEndpointExact
		return resolved, nil
	}

	if len(cmd.APISurface) != 1 {
		resolved.Method = runtime.method
		resolved.Path = runtime.path
		resolved.TransportMethod = runtime.method
		resolved.TransportPath = runtime.path
		resolved.Equivalence = CommandEndpointExact
		return resolved, nil
	}
	proofPrefix := ""
	if transportHook, ok := hooks.(CommandBindingTransportHook); ok {
		method, path, handled := transportHook.CommandBindingTransport(resolved.Binding)
		if handled {
			if strings.TrimSpace(method) == "" || strings.TrimSpace(path) == "" {
				return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q registered hook returned an empty transport endpoint", cmd.Path)
			}
			runtime.method = strings.ToUpper(strings.TrimSpace(method))
			runtime.path = normalizeBindingPathTemplate(path)
			proofPrefix = CommandEndpointHookTransport
		}
	}

	reference := cmd.APISurface[0]
	canonicalMethod := strings.ToUpper(strings.TrimSpace(reference.Method))
	canonicalPath := reference.Path
	proof, err := proveCommandEndpointEquivalence(b, cmd, resolved.Binding, runtime, canonicalMethod, canonicalPath, proofPrefix != "")
	if err != nil {
		return ResolvedCommandBinding{}, fmt.Errorf("implemented command %q binding %s/%s: %w", cmd.Path, resolved.Binding.Kind, resolved.Binding.ID, err)
	}
	if proofPrefix != "" {
		proof = proofPrefix
	}
	resolved.Method = canonicalMethod
	resolved.Path = canonicalPath
	resolved.TransportMethod = runtime.method
	resolved.TransportPath = runtime.path
	resolved.Equivalence = proof
	return resolved, nil
}

func proveCommandEndpointEquivalence(b Bundle, cmd connectors.CommandSurfaceCommand, binding connectors.CommandBindingIdentity, runtime commandRuntimeEndpoint, canonicalMethod, canonicalPath string, hookTransport bool) (string, error) {
	if canonicalMethod == "GRAPHQL" {
		if runtime.graphqlOperation == "" || runtime.graphqlOperation != canonicalPath {
			return "", fmt.Errorf("runtime GraphQL operation %q does not match canonical operation %q", runtime.graphqlOperation, canonicalPath)
		}
		if strings.ToUpper(runtime.method) != http.MethodPost || commandBindingComparablePath(runtime.path) != "/graphql" {
			return "", fmt.Errorf("GraphQL operation %q does not use the declared POST /graphql transport", canonicalPath)
		}
		return CommandEndpointGraphQLTransport, nil
	}
	if strings.ToUpper(runtime.method) != canonicalMethod {
		return "", fmt.Errorf("runtime method %s does not match canonical method %s", runtime.method, canonicalMethod)
	}

	runtimeComparable, runtimeChange := commandBindingComparablePathWithProof(runtime.path)
	declaredComparable, declaredChange := commandBindingComparablePathWithProof(canonicalPath)
	if proof, matched, err := proveCircleCICompositeProviderPathIdentity(b, cmd, binding, runtime, canonicalMethod, canonicalPath, runtimeChange, declaredChange, hookTransport); err != nil {
		return "", err
	} else if matched {
		return proof, nil
	}
	if commandBindingSlots(runtimeComparable) == commandBindingSlots(declaredComparable) {
		return commandBindingEquivalenceProof(runtime.path, canonicalPath, runtimeChange, declaredChange), nil
	}
	for _, basePath := range commandBindingBasePaths(b, runtime.route, runtime.baseURL) {
		combined := strings.TrimRight(basePath, "/") + "/" + strings.TrimLeft(runtimeComparable, "/")
		if strings.TrimRight(commandBindingSlots(combined), "/") == strings.TrimRight(commandBindingSlots(declaredComparable), "/") {
			return CommandEndpointBasePath, nil
		}
	}
	if commandBindingUsesConfigurableBase(b, runtime.route, runtime.baseURL) &&
		strings.HasSuffix(commandBindingSlots(declaredComparable), commandBindingSlots(runtimeComparable)) {
		return CommandEndpointBasePath, nil
	}
	return "", fmt.Errorf("runtime endpoint %s %s is not canonically equivalent to %s %s", runtime.method, runtime.path, canonicalMethod, canonicalPath)
}

// proveCircleCICompositeProviderPathIdentity is deliberately a closed proof
// for the retained CircleCI OpenAPI project's project-slug identity. It cannot
// be configured for another connector or another shape: validation requires
// the complete source-cited manifest, and this function requires the one exact
// runtime expansion to vcs_type/org/repo.
func proveCircleCICompositeProviderPathIdentity(b Bundle, cmd connectors.CommandSurfaceCommand, binding connectors.CommandBindingIdentity, runtime commandRuntimeEndpoint, canonicalMethod, canonicalPath, runtimeChange, declaredChange string, hookTransport bool) (string, bool, error) {
	if b.CompositeProviderPathIdentity == nil {
		return "", false, nil
	}
	if err := validateCompositeProviderPathIdentity(b.Name, b.CompositeProviderPathIdentity); err != nil {
		return "", false, err
	}

	var expected *CompositeProviderPathBinding
	for index := range circleCICompositeProviderPathBindings {
		candidate := &circleCICompositeProviderPathBindings[index]
		if candidate.Intent == cmd.Intent && candidate.BindingKind == binding.Kind && candidate.BindingID == binding.ID &&
			candidate.Method == canonicalMethod && candidate.Path == canonicalPath {
			expected = candidate
			break
		}
	}
	if expected == nil {
		return "", false, nil
	}
	if hookTransport {
		return "", false, fmt.Errorf("CircleCI composite provider path identity does not permit hook transport")
	}
	if runtimeChange != "" || declaredChange != "" || runtime.route != "" || runtime.baseURL != "" {
		return "", false, fmt.Errorf("CircleCI composite provider path identity requires the declared relative transport path without query, suffix, annotation, route, or base override")
	}
	if strings.Count(canonicalPath, "{project-slug}") != 1 {
		return "", false, fmt.Errorf("CircleCI composite provider path identity requires exactly one {project-slug} placeholder")
	}
	expectedRuntimePath := strings.Replace(canonicalPath, "{project-slug}", "{vcs_type}/{org}/{repo}", 1)
	if runtime.path != expectedRuntimePath || strings.ToUpper(runtime.method) != expected.Method {
		return "", false, fmt.Errorf("CircleCI composite provider path identity requires %s %s transport, got %s %s", expected.Method, expectedRuntimePath, runtime.method, runtime.path)
	}
	return CommandEndpointCompositeProviderPathIdentity, true, nil
}

func commandBindingEquivalenceProof(runtimePath, canonicalPath, runtimeChange, canonicalChange string) string {
	switch {
	case runtimeChange == CommandEndpointQueryTransport:
		return CommandEndpointQueryTransport
	case runtimeChange == CommandEndpointAbsoluteTransport:
		return CommandEndpointAbsoluteTransport
	case runtimeChange == CommandEndpointSuffixTransport || canonicalChange == CommandEndpointSuffixTransport:
		return CommandEndpointSuffixTransport
	case canonicalChange == CommandEndpointAnnotationIdentity:
		return CommandEndpointAnnotationIdentity
	case normalizeBindingPathTemplate(runtimePath) != normalizeBindingPathTemplate(canonicalPath):
		return CommandEndpointPlaceholder
	default:
		return CommandEndpointExact
	}
}

func commandBindingComparablePath(path string) string {
	value, _ := commandBindingComparablePathWithProof(path)
	return value
}

func commandBindingComparablePathWithProof(path string) (string, string) {
	value := strings.TrimSpace(normalizeBindingPathTemplate(path))
	proof := ""
	if parsed, err := url.Parse(value); err == nil && parsed.IsAbs() && parsed.Host != "" {
		value = parsed.Path
		proof = CommandEndpointAbsoluteTransport
		if value == "" {
			value = "/"
		}
		if parsed.RawQuery != "" {
			proof = CommandEndpointQueryTransport
		}
	} else if query := strings.IndexByte(value, '?'); query >= 0 {
		value = value[:query]
		proof = CommandEndpointQueryTransport
	}
	if commandBindingAnnotationRE.MatchString(value) {
		value = commandBindingAnnotationRE.ReplaceAllString(value, "")
		proof = CommandEndpointAnnotationIdentity
	}
	if strings.HasSuffix(value, ".json") {
		value = strings.TrimSuffix(value, ".json")
		proof = CommandEndpointSuffixTransport
	}
	return value, proof
}

func commandBindingSlots(path string) string {
	return commandBindingSlotRE.ReplaceAllString(path, "{}")
}

func commandBindingBasePaths(b Bundle, routeName, override string) []string {
	candidates := []string{}
	if strings.TrimSpace(override) != "" {
		candidates = append(candidates, override)
	} else if strings.TrimSpace(routeName) != "" {
		for _, route := range b.HTTP.Routes {
			if route.Name == routeName {
				if strings.TrimSpace(route.Version) == "" {
					candidates = append(candidates, route.BaseURL)
				}
				break
			}
		}
	} else {
		candidates = append(candidates, b.HTTP.URL)
	}

	seen := map[string]bool{}
	paths := []string{}
	for _, candidate := range candidates {
		path := commandBindingBasePath(b, candidate)
		if path == "" || path == "/" || seen[path] {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	return paths
}

func commandBindingUsesConfigurableBase(b Bundle, routeName, override string) bool {
	if strings.TrimSpace(override) != "" {
		return commandBindingBaseURLRE.MatchString(override)
	}
	if strings.TrimSpace(routeName) != "" {
		for _, route := range b.HTTP.Routes {
			if route.Name == routeName {
				return commandBindingBaseURLRE.MatchString(route.BaseURL)
			}
		}
		return false
	}
	return commandBindingBaseURLRE.MatchString(b.HTTP.URL)
}

func commandBindingBasePath(b Bundle, raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if b.Spec != nil {
		if baseURL := strings.TrimSpace(b.Spec.Defaults()["base_url"]); baseURL != "" {
			value = commandBindingBaseURLRE.ReplaceAllString(value, baseURL)
		}
	}
	value = commandBindingBaseURLRE.ReplaceAllString(value, "https://declared-base.invalid")
	value = normalizeBindingPathTemplate(value)
	parsed, err := url.Parse(value)
	if err != nil {
		return ""
	}
	if parsed.IsAbs() && parsed.Host != "" {
		return strings.TrimRight(parsed.Path, "/")
	}
	if strings.HasPrefix(value, "/") {
		return strings.TrimRight(value, "/")
	}
	return ""
}

// ResolveImplementedCommandPath resolves one bundle command through the same
// projected command surface commandrunner receives.
func ResolveImplementedCommandPath(b Bundle, path string) (ResolvedCommandBinding, error) {
	surface := synthesizeCommandSurface(b)
	if surface == nil {
		return ResolvedCommandBinding{}, fmt.Errorf("bundle %q has no command surface", b.Name)
	}
	matches := 0
	var command connectors.CommandSurfaceCommand
	for _, candidate := range surface.Commands {
		if candidate.Path == path {
			matches++
			command = candidate
		}
	}
	if matches != 1 {
		return ResolvedCommandBinding{}, fmt.Errorf("bundle %q command %q resolves %d times", b.Name, path, matches)
	}
	return ResolveImplementedCommandBinding(b, command)
}

// PreflightImplementedCommand lets commandrunner use the same resolver as
// admission before selecting an executor.
func (c *Connector) PreflightImplementedCommand(cmd connectors.CommandSurfaceCommand) error {
	_, err := resolveImplementedCommandBinding(c.bundle, cmd, c.hooks)
	return err
}

func (b Base) PreflightImplementedCommand(cmd connectors.CommandSurfaceCommand) error {
	_, err := ResolveImplementedCommandBinding(b.bundle, cmd)
	return err
}

func commandBindingStream(b Bundle, name string) (StreamSpec, bool) {
	for _, stream := range b.Streams {
		if stream.Name == name {
			return stream, true
		}
	}
	return StreamSpec{}, false
}

func commandBindingWrite(b Bundle, name string) (WriteAction, bool) {
	for _, action := range b.Writes {
		if action.Name == name {
			return action, true
		}
	}
	return WriteAction{}, false
}

func commandBindingOperation(b Bundle, id string) (OperationSpec, bool) {
	for _, operation := range b.Operations {
		if operation.ID == id {
			return operation, true
		}
	}
	return OperationSpec{}, false
}

func commandBindingOperationEndpoint(operation OperationSpec) (string, string, error) {
	switch {
	case operation.REST != nil:
		return strings.ToUpper(strings.TrimSpace(operation.REST.Method)), operation.REST.Path, nil
	case operation.Binary != nil:
		return strings.ToUpper(strings.TrimSpace(operation.Binary.Method)), operation.Binary.Path, nil
	case operation.GraphQL != nil:
		return http.MethodPost, operation.GraphQL.Path, nil
	default:
		return "", "", fmt.Errorf("operation %q has no command-addressable endpoint", operation.ID)
	}
}

func normalizeBindingPathTemplate(path string) string {
	return commandBindingTemplateRE.ReplaceAllString(path, `{$1}`)
}
