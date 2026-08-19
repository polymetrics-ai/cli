package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// source-import turns a verified, connector-owned provider source into a
// canonical intermediate descriptor. It intentionally owns no execution
// controls: a connector name selects its checked-in lock, and that lock alone
// supplies the artifact URL, byte count, and digest.
const sourceImportUsage = `connectorgen source-import <connector> --out <path> [--defs <dir>] [--check]

Verifies the connector-owned source lock, retrieves only its fixed public
artifact URL, and writes canonical provider operation descriptors for later
declaration generators.

  <connector>     connector whose sources/<connector>-operation-source-lock.json is used
  --out <path>    descriptor output path
  --defs <dir>    connector defs root (default internal/connectors/defs)
  --check         compare generated descriptors with --out; do not write

The source lock is authoritative. A byte or SHA-256 mismatch requires a
source-lock refresh; this command never accepts a replacement URL, method,
path, header, body, credential, or generic request input.`

const (
	defaultSourceImportArtifactBytes  = int64(16 << 20)
	defaultSourceImportSchemaBytes    = int64(1 << 20)
	defaultSourceImportOperations     = 10_000
	defaultSourceImportReferences     = 50_000
	defaultSourceImportReferenceDepth = 32
)

type sourceImportLimits struct {
	MaxArtifactBytes  int64
	MaxSchemaBytes    int64
	MaxOperations     int
	MaxReferences     int
	MaxReferenceDepth int
}

func defaultSourceImportLimits() sourceImportLimits {
	return sourceImportLimits{
		MaxArtifactBytes:  defaultSourceImportArtifactBytes,
		MaxSchemaBytes:    defaultSourceImportSchemaBytes,
		MaxOperations:     defaultSourceImportOperations,
		MaxReferences:     defaultSourceImportReferences,
		MaxReferenceDepth: defaultSourceImportReferenceDepth,
	}
}

type sourceImportArtifact struct {
	SourceURL string `json:"source_url"`
	SHA256    string `json:"sha256"`
	Bytes     int64  `json:"bytes"`
	OpenAPI   string `json:"openapi,omitempty"`
	Swagger   string `json:"swagger,omitempty"`
}

type sourceImportLock struct {
	SchemaVersion int                  `json:"schema_version"`
	Connector     string               `json:"connector"`
	Rest          sourceImportArtifact `json:"rest"`
}

type sourceImportSource struct {
	URL      string `json:"url"`
	SHA256   string `json:"sha256"`
	Bytes    int64  `json:"bytes"`
	Location string `json:"location"`
}

type sourceParameterDescriptor struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Schema   any    `json:"schema"`
}

type sourceRequestBodyDescriptor struct {
	Required bool `json:"required"`
	Schema   any  `json:"schema"`
}

type sourceRequestDescriptor struct {
	Path      []sourceParameterDescriptor  `json:"path"`
	Query     []sourceParameterDescriptor  `json:"query"`
	Header    []sourceParameterDescriptor  `json:"header"`
	Body      *sourceRequestBodyDescriptor `json:"body,omitempty"`
	MediaType string                       `json:"media_type,omitempty"`
}

type sourceResponseDescriptor struct {
	Status      string `json:"status"`
	Declaration any    `json:"declaration"`
}

type sourceOutputClass string

const (
	sourceOutputJSON   sourceOutputClass = "json"
	sourceOutputBinary sourceOutputClass = "binary"
	sourceOutputStatus sourceOutputClass = "status"
	sourceOutputText   sourceOutputClass = "text"
)

type sourceOutputDescriptor struct {
	Class      sourceOutputClass `json:"class"`
	MediaTypes []string          `json:"media_types"`
}

type sourceByteLimits struct {
	Request  int64 `json:"request,omitempty"`
	Response int64 `json:"response,omitempty"`
}

type sourceAuthScope struct {
	Scheme string   `json:"scheme"`
	Scopes []string `json:"scopes"`
}

type sourceAuthRequirementGroup struct {
	AllOf []sourceAuthScope `json:"all_of"`
}

type sourceAuthDescriptor struct {
	Declared bool                         `json:"declared"`
	AnyOf    []sourceAuthRequirementGroup `json:"any_of"`
}

// sourceOperationDescriptor is the immutable bridge from a verified provider
// source to later declaration materializers. Responses intentionally retain
// their resolved provider declaration wholesale; output classification is
// additive metadata, never a response-field filter.
type sourceOperationDescriptor struct {
	Connector           string                     `json:"connector"`
	SourceID            string                     `json:"source_id"`
	ProviderOperationID string                     `json:"operation_id"`
	Source              sourceImportSource         `json:"source"`
	Method              string                     `json:"method"`
	Path                string                     `json:"path"`
	Request             sourceRequestDescriptor    `json:"request"`
	Responses           []sourceResponseDescriptor `json:"responses"`
	Output              sourceOutputDescriptor     `json:"output"`
	Pagination          any                        `json:"pagination,omitempty"`
	ByteLimits          sourceByteLimits           `json:"byte_limits"`
	AuthScopes          sourceAuthDescriptor       `json:"auth_scopes"`
}

type sourceImportDescriptorDocument struct {
	SchemaVersion int                         `json:"schema_version"`
	Operations    []sourceOperationDescriptor `json:"operations"`
}

type sourceImportFetcher interface {
	Fetch(context.Context, string) ([]byte, error)
}

type sourceImportFetchFunc func(context.Context, string) ([]byte, error)

func (f sourceImportFetchFunc) Fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	return f(ctx, sourceURL)
}

func parseSourceImportLock(raw []byte, expectedConnector string) (sourceImportLock, error) {
	var lock sourceImportLock
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&lock); err != nil {
		return sourceImportLock{}, fmt.Errorf("parse source lock: %w", err)
	}
	if err := sourceJSONEOF(decoder); err != nil {
		return sourceImportLock{}, fmt.Errorf("parse source lock: %w", err)
	}
	if lock.SchemaVersion <= 0 {
		return sourceImportLock{}, fmt.Errorf("source lock has invalid schema version")
	}
	if lock.Connector == "" {
		return sourceImportLock{}, fmt.Errorf("source lock has no connector")
	}
	if expectedConnector != "" && lock.Connector != expectedConnector {
		return sourceImportLock{}, fmt.Errorf("source lock connector %q does not match requested connector %q", lock.Connector, expectedConnector)
	}
	if err := validateSourceImportConnector(lock.Connector); err != nil {
		return sourceImportLock{}, err
	}
	if err := validateSourceImportArtifact(lock.Rest); err != nil {
		return sourceImportLock{}, err
	}
	return lock, nil
}

func validateSourceImportConnector(connector string) error {
	if connector == "" || len(connector) > 80 {
		return fmt.Errorf("invalid connector name")
	}
	for i, r := range connector {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || (r == '-' && i > 0) {
			continue
		}
		return fmt.Errorf("invalid connector name %q", connector)
	}
	return nil
}

func validateSourceImportArtifact(artifact sourceImportArtifact) error {
	if _, err := parseBatchArtifactURL(artifact.SourceURL); err != nil {
		return fmt.Errorf("source lock has invalid public artifact URL: %w", err)
	}
	if artifact.Bytes <= 0 {
		return fmt.Errorf("source lock has invalid artifact byte count")
	}
	if len(artifact.SHA256) != sha256.Size*2 {
		return fmt.Errorf("source lock has invalid SHA-256")
	}
	if _, err := hex.DecodeString(artifact.SHA256); err != nil {
		return fmt.Errorf("source lock has invalid SHA-256: %w", err)
	}
	return nil
}

func importSourceLocks(ctx context.Context, locks []sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceOperationDescriptor, error) {
	if fetcher == nil {
		return nil, fmt.Errorf("source importer has no fetcher")
	}
	if err := validateSourceImportLimits(limits); err != nil {
		return nil, err
	}
	orderedLocks := append([]sourceImportLock(nil), locks...)
	sort.Slice(orderedLocks, func(i, j int) bool { return orderedLocks[i].Connector < orderedLocks[j].Connector })
	seenConnectors := map[string]bool{}
	var descriptors []sourceOperationDescriptor
	for _, lock := range orderedLocks {
		if seenConnectors[lock.Connector] {
			return nil, fmt.Errorf("duplicate source-lock connector %q", lock.Connector)
		}
		seenConnectors[lock.Connector] = true
		imported, err := importSourceLock(ctx, lock, fetcher, limits)
		if err != nil {
			return nil, err
		}
		descriptors = append(descriptors, imported...)
	}
	sortSourceOperationDescriptors(descriptors)
	seen := map[string]bool{}
	for _, descriptor := range descriptors {
		key := descriptor.Connector + "\x00" + descriptor.SourceID
		if seen[key] {
			return nil, fmt.Errorf("duplicate source identity %q for connector %q", descriptor.SourceID, descriptor.Connector)
		}
		seen[key] = true
	}
	return descriptors, nil
}

func validateSourceImportLimits(limits sourceImportLimits) error {
	if limits.MaxArtifactBytes <= 0 || limits.MaxSchemaBytes <= 0 || limits.MaxOperations <= 0 || limits.MaxReferences <= 0 || limits.MaxReferenceDepth <= 0 {
		return fmt.Errorf("source import limits must all be positive")
	}
	return nil
}

func importSourceLock(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceOperationDescriptor, error) {
	if err := validateSourceImportLimits(limits); err != nil {
		return nil, err
	}
	if err := validateSourceImportConnector(lock.Connector); err != nil {
		return nil, err
	}
	if err := validateSourceImportArtifact(lock.Rest); err != nil {
		return nil, err
	}
	if lock.Rest.Bytes > limits.MaxArtifactBytes {
		return nil, fmt.Errorf("artifact byte limit exceeded by source lock")
	}
	raw, err := fetcher.Fetch(ctx, lock.Rest.SourceURL)
	if err != nil {
		return nil, fmt.Errorf("fetch locked source artifact: %w", err)
	}
	if int64(len(raw)) > limits.MaxArtifactBytes {
		return nil, fmt.Errorf("artifact byte limit exceeded")
	}
	actualDigest := sha256.Sum256(raw)
	if int64(len(raw)) != lock.Rest.Bytes || !strings.EqualFold(hex.EncodeToString(actualDigest[:]), lock.Rest.SHA256) {
		return nil, fmt.Errorf("source-lock refresh required: fetched artifact does not match locked bytes and SHA-256")
	}
	doc, version, err := parseSourceImportDocument(raw)
	if err != nil {
		return nil, err
	}
	resolver := sourceReferenceResolver{root: doc, limits: limits}
	descriptors, err := importSourceDocument(lock, doc, version, &resolver, limits)
	if err != nil {
		return nil, err
	}
	sortSourceOperationDescriptors(descriptors)
	seen := map[string]bool{}
	for _, descriptor := range descriptors {
		if seen[descriptor.SourceID] {
			return nil, fmt.Errorf("duplicate source identity %q", descriptor.SourceID)
		}
		seen[descriptor.SourceID] = true
	}
	return descriptors, nil
}

func sortSourceOperationDescriptors(descriptors []sourceOperationDescriptor) {
	sort.Slice(descriptors, func(i, j int) bool {
		if descriptors[i].Connector != descriptors[j].Connector {
			return descriptors[i].Connector < descriptors[j].Connector
		}
		if descriptors[i].Method != descriptors[j].Method {
			return descriptors[i].Method < descriptors[j].Method
		}
		if descriptors[i].Path != descriptors[j].Path {
			return descriptors[i].Path < descriptors[j].Path
		}
		return descriptors[i].SourceID < descriptors[j].SourceID
	})
}

func parseSourceImportDocument(raw []byte) (map[string]any, string, error) {
	var value any
	if err := decodeSourceJSON(raw, &value); err != nil {
		var yamlValue any
		if yamlErr := decodeSourceYAML(raw, &yamlValue); yamlErr != nil {
			return nil, "", fmt.Errorf("parse source artifact as JSON or YAML: JSON: %v; YAML: %w", err, yamlErr)
		}
		canonical, yamlErr := json.Marshal(normalizeSourceYAML(yamlValue))
		if yamlErr != nil {
			return nil, "", fmt.Errorf("normalize YAML source artifact: %w", yamlErr)
		}
		if err := decodeSourceJSON(canonical, &value); err != nil {
			return nil, "", fmt.Errorf("parse normalized YAML source artifact: %w", err)
		}
	}
	doc, ok := value.(map[string]any)
	if !ok {
		return nil, "", fmt.Errorf("source artifact root must be an object")
	}
	if version, _ := doc["openapi"].(string); strings.HasPrefix(version, "3.") {
		return doc, "openapi3", nil
	}
	if version, _ := doc["swagger"].(string); version == "2.0" {
		return doc, "swagger2", nil
	}
	return nil, "", fmt.Errorf("unsupported source artifact form: require OpenAPI 3 or Swagger 2")
}

func decodeSourceJSON(raw []byte, target *any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return sourceJSONEOF(decoder)
}

func sourceJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("unexpected trailing JSON value")
	}
	return err
}

func decodeSourceYAML(raw []byte, target *any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("multiple YAML documents are unsupported")
	}
	return nil
}

func normalizeSourceYAML(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			out[key] = normalizeSourceYAML(child)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, child := range typed {
			out[fmt.Sprint(key)] = normalizeSourceYAML(child)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			out[i] = normalizeSourceYAML(child)
		}
		return out
	default:
		return value
	}
}

type sourceReferenceResolver struct {
	root       map[string]any
	limits     sourceImportLimits
	references int
}

func (r *sourceReferenceResolver) resolve(value any) (any, error) {
	return r.resolveAt(value, nil, 0)
}

func (r *sourceReferenceResolver) resolveAt(value any, stack map[string]bool, depth int) (any, error) {
	if depth > r.limits.MaxReferenceDepth {
		return nil, fmt.Errorf("reference depth limit exceeded")
	}
	switch typed := value.(type) {
	case map[string]any:
		if rawRef, ok := typed["$ref"]; ok {
			if len(typed) != 1 {
				return nil, fmt.Errorf("ambiguous reference with sibling fields")
			}
			ref, ok := rawRef.(string)
			if !ok || !strings.HasPrefix(ref, "#/") {
				return nil, fmt.Errorf("external reference %q is unsupported", rawRef)
			}
			r.references++
			if r.references > r.limits.MaxReferences {
				return nil, fmt.Errorf("reference count limit exceeded")
			}
			if stack != nil && stack[ref] {
				return nil, fmt.Errorf("reference cycle at %q", ref)
			}
			target, err := sourcePointer(r.root, ref)
			if err != nil {
				return nil, err
			}
			if stack == nil {
				stack = map[string]bool{}
			} else {
				stack = copySourceRefStack(stack)
			}
			stack[ref] = true
			return r.resolveAt(target, stack, depth+1)
		}
		out := make(map[string]any, len(typed))
		for _, key := range sortedSourceMapKeys(typed) {
			child, err := r.resolveAt(typed[key], stack, depth)
			if err != nil {
				return nil, err
			}
			out[key] = child
		}
		return out, nil
	case []any:
		out := make([]any, len(typed))
		for i, child := range typed {
			resolved, err := r.resolveAt(child, stack, depth)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	default:
		return value, nil
	}
}

func copySourceRefStack(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in)+1)
	for key, value := range in {
		out[key] = value
	}
	return out
}

func sourcePointer(root map[string]any, ref string) (any, error) {
	var current any = root
	for _, segment := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		segment = strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~")
		object, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("reference %q does not resolve to an object member", ref)
		}
		var exists bool
		current, exists = object[segment]
		if !exists {
			return nil, fmt.Errorf("unresolved reference %q", ref)
		}
	}
	return current, nil
}

func importSourceDocument(lock sourceImportLock, doc map[string]any, version string, resolver *sourceReferenceResolver, limits sourceImportLimits) ([]sourceOperationDescriptor, error) {
	paths, ok := doc["paths"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("source artifact has no paths object")
	}
	var descriptors []sourceOperationDescriptor
	for _, path := range sortedSourceMapKeys(paths) {
		if err := validateSourceImportPath(path); err != nil {
			return nil, fmt.Errorf("path %q is not a connector-relative path template", path)
		}
		resolvedPathItem, err := resolver.resolve(paths[path])
		if err != nil {
			return nil, fmt.Errorf("path %q: %w", path, err)
		}
		pathItem, ok := resolvedPathItem.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("path %q must be an object", path)
		}
		if _, hasCallbacks := pathItem["callbacks"]; hasCallbacks {
			return nil, fmt.Errorf("callback-only route %q is unsupported", path)
		}
		pathParameters, err := sourceParameterValues(pathItem["parameters"], resolver, version)
		if err != nil {
			return nil, fmt.Errorf("path %q parameters: %w", path, err)
		}
		for _, method := range sourceHTTPMethods {
			rawOperation, ok := pathItem[method]
			if !ok {
				continue
			}
			operation, ok := rawOperation.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("operation %s %s must be an object", method, path)
			}
			if _, hasCallbacks := operation["callbacks"]; hasCallbacks {
				return nil, fmt.Errorf("callback-only route %s %s is unsupported", method, path)
			}
			descriptor, err := importSourceOperation(lock, doc, version, resolver, path, method, pathParameters, operation, limits)
			if err != nil {
				return nil, err
			}
			descriptors = append(descriptors, descriptor)
			if len(descriptors) > limits.MaxOperations {
				return nil, fmt.Errorf("operation count limit exceeded")
			}
		}
	}
	if len(descriptors) == 0 {
		return nil, fmt.Errorf("source artifact has no provider operations")
	}
	return descriptors, nil
}

var sourceHTTPMethods = []string{"delete", "get", "head", "options", "patch", "post", "put", "trace"}

func validateSourceImportPath(path string) error {
	if path == "" || path != strings.TrimSpace(path) || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "://") || strings.ContainsAny(path, "\r\n?#") {
		return fmt.Errorf("not connector-relative")
	}
	return nil
}

func importSourceOperation(lock sourceImportLock, doc map[string]any, version string, resolver *sourceReferenceResolver, path, method string, pathParameters []sourceParameterValue, operation map[string]any, limits sourceImportLimits) (sourceOperationDescriptor, error) {
	location := fmt.Sprintf("paths[%q].%s", path, method)
	operationParameters, err := sourceParameterValues(operation["parameters"], resolver, version)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s parameters: %w", location, err)
	}
	request, err := sourceRequestDescriptorFrom(path, pathParameters, operationParameters, operation, doc, version, resolver, limits)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s request: %w", location, err)
	}
	responses, mediaTypes, err := sourceResponses(operation, doc, version, resolver, limits)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s responses: %w", location, err)
	}
	outputClass, err := sourceOutputClassFor(method, mediaTypes)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s output: %w", location, err)
	}
	providerID, _ := operation["operationId"].(string)
	sourceID := providerID
	if sourceID == "" {
		sourceID = fmt.Sprintf("%s.rest.%s_%s", lock.Connector, method, path)
	}
	pagination, err := sourcePagination(operation, resolver)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s: %w", location, err)
	}
	byteLimits, err := sourceOperationByteLimits(operation)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s: %w", location, err)
	}
	authScopes, err := sourceAuthRequirements(operation, doc)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s auth scopes: %w", location, err)
	}
	return sourceOperationDescriptor{
		Connector:           lock.Connector,
		SourceID:            sourceID,
		ProviderOperationID: providerID,
		Source:              sourceImportSource{URL: lock.Rest.SourceURL, SHA256: strings.ToLower(lock.Rest.SHA256), Bytes: lock.Rest.Bytes, Location: location},
		Method:              method,
		Path:                path,
		Request:             request,
		Responses:           responses,
		Output:              sourceOutputDescriptor{Class: outputClass, MediaTypes: mediaTypes},
		Pagination:          pagination,
		ByteLimits:          byteLimits,
		AuthScopes:          authScopes,
	}, nil
}

type sourceParameterValue struct {
	Name     string
	In       string
	Required bool
	Schema   any
}

func sourceParameterValues(raw any, resolver *sourceReferenceResolver, version string) ([]sourceParameterValue, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("parameters must be an array")
	}
	values := make([]sourceParameterValue, 0, len(items))
	for _, item := range items {
		resolved, err := resolver.resolve(item)
		if err != nil {
			return nil, err
		}
		parameter, ok := resolved.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("parameter must be an object")
		}
		name, _ := parameter["name"].(string)
		in, _ := parameter["in"].(string)
		allowedLocation := in == "path" || in == "query" || in == "header" || (version == "swagger2" && in == "body")
		if name == "" || name != strings.TrimSpace(name) || !allowedLocation {
			return nil, fmt.Errorf("parameter %q has unsupported location %q", name, in)
		}
		schema, err := sourceParameterSchema(parameter, version)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}
		resolvedSchema, err := resolver.resolve(schema)
		if err != nil {
			return nil, fmt.Errorf("parameter %q schema: %w", name, err)
		}
		values = append(values, sourceParameterValue{Name: name, In: in, Required: sourceBool(parameter["required"]), Schema: resolvedSchema})
	}
	return values, nil
}

func sourceParameterSchema(parameter map[string]any, version string) (any, error) {
	if version == "openapi3" {
		schema, ok := parameter["schema"]
		if !ok {
			return nil, fmt.Errorf("missing schema")
		}
		return schema, nil
	}
	if schema, ok := parameter["schema"]; ok {
		return schema, nil
	}
	schema := map[string]any{}
	for _, key := range []string{"type", "format", "enum", "items", "maximum", "minimum", "maxLength", "maxItems", "maxProperties", "properties", "additionalProperties", "required"} {
		if value, ok := parameter[key]; ok {
			schema[key] = value
		}
	}
	if len(schema) == 0 {
		return nil, fmt.Errorf("missing schema")
	}
	return schema, nil
}

func sourceRequestDescriptorFrom(path string, pathParameters, operationParameters []sourceParameterValue, operation, doc map[string]any, version string, resolver *sourceReferenceResolver, limits sourceImportLimits) (sourceRequestDescriptor, error) {
	parameters, err := sourceEffectiveParameters(pathParameters, operationParameters)
	if err != nil {
		return sourceRequestDescriptor{}, err
	}
	request := sourceRequestDescriptor{Path: []sourceParameterDescriptor{}, Query: []sourceParameterDescriptor{}, Header: []sourceParameterDescriptor{}}
	var bodyParameters []sourceParameterValue
	for _, parameter := range parameters {
		if parameter.In == "body" {
			bodyParameters = append(bodyParameters, parameter)
			continue
		}
		if err := validateBoundedRequestSchema(parameter.Schema, limits, 0); err != nil {
			return sourceRequestDescriptor{}, fmt.Errorf("parameter %q: %w", parameter.Name, err)
		}
		descriptor := sourceParameterDescriptor{Name: parameter.Name, Required: parameter.Required, Schema: parameter.Schema}
		switch parameter.In {
		case "path":
			request.Path = append(request.Path, descriptor)
		case "query":
			request.Query = append(request.Query, descriptor)
		case "header":
			request.Header = append(request.Header, descriptor)
		}
	}
	if err := validateSourcePathParameters(path, parameters); err != nil {
		return sourceRequestDescriptor{}, err
	}
	for _, group := range [][]sourceParameterDescriptor{request.Path, request.Query, request.Header} {
		sort.Slice(group, func(i, j int) bool { return group[i].Name < group[j].Name })
	}
	if version == "swagger2" {
		if len(bodyParameters) == 0 {
			return request, nil
		}
		if len(bodyParameters) != 1 {
			return sourceRequestDescriptor{}, fmt.Errorf("Swagger request body is ambiguous")
		}
		body := bodyParameters[0]
		if err := validateBoundedRequestSchema(body.Schema, limits, 0); err != nil {
			return sourceRequestDescriptor{}, fmt.Errorf("request body: %w", err)
		}
		mediaType, err := sourceSwaggerRequestMediaType(operation, doc)
		if err != nil {
			return sourceRequestDescriptor{}, err
		}
		request.Body = &sourceRequestBodyDescriptor{Required: body.Required, Schema: body.Schema}
		request.MediaType = mediaType
		return request, nil
	}
	if len(bodyParameters) > 0 {
		return sourceRequestDescriptor{}, fmt.Errorf("request body parameter is only supported by Swagger 2")
	}
	rawBody, ok := operation["requestBody"]
	if !ok {
		return request, nil
	}
	resolvedBody, err := resolver.resolve(rawBody)
	if err != nil {
		return sourceRequestDescriptor{}, err
	}
	body, ok := resolvedBody.(map[string]any)
	if !ok {
		return sourceRequestDescriptor{}, fmt.Errorf("request body must be an object")
	}
	content, ok := body["content"].(map[string]any)
	if !ok || len(content) != 1 {
		return sourceRequestDescriptor{}, fmt.Errorf("request body requires exactly one unambiguous media type")
	}
	mediaType := sortedSourceMapKeys(content)[0]
	if !sourceJSONMediaType(mediaType) {
		return sourceRequestDescriptor{}, fmt.Errorf("unsupported request encoding %q", mediaType)
	}
	media, ok := content[mediaType].(map[string]any)
	if !ok {
		return sourceRequestDescriptor{}, fmt.Errorf("request media declaration must be an object")
	}
	schema, ok := media["schema"]
	if !ok {
		return sourceRequestDescriptor{}, fmt.Errorf("request body is missing schema")
	}
	resolvedSchema, err := resolver.resolve(schema)
	if err != nil {
		return sourceRequestDescriptor{}, err
	}
	if err := validateBoundedRequestSchema(resolvedSchema, limits, 0); err != nil {
		return sourceRequestDescriptor{}, err
	}
	request.Body = &sourceRequestBodyDescriptor{Required: sourceBool(body["required"]), Schema: resolvedSchema}
	request.MediaType = mediaType
	return request, nil
}

func sourceEffectiveParameters(pathParameters, operationParameters []sourceParameterValue) ([]sourceParameterValue, error) {
	effective := make(map[string]sourceParameterValue, len(pathParameters)+len(operationParameters))
	for _, parameters := range [][]sourceParameterValue{pathParameters, operationParameters} {
		seen := make(map[string]bool, len(parameters))
		for _, parameter := range parameters {
			key := parameter.In + "\x00" + parameter.Name
			if seen[key] {
				return nil, fmt.Errorf("ambiguous parameter %q in %q", parameter.Name, parameter.In)
			}
			seen[key] = true
			effective[key] = parameter
		}
	}
	parameters := make([]sourceParameterValue, 0, len(effective))
	for _, parameter := range effective {
		parameters = append(parameters, parameter)
	}
	sort.Slice(parameters, func(i, j int) bool {
		if parameters[i].In != parameters[j].In {
			return parameters[i].In < parameters[j].In
		}
		return parameters[i].Name < parameters[j].Name
	})
	return parameters, nil
}

func validateSourcePathParameters(path string, parameters []sourceParameterValue) error {
	placeholders, err := sourcePathTemplateParameters(path)
	if err != nil {
		return err
	}
	pathParameters := make(map[string]bool)
	for _, parameter := range parameters {
		if parameter.In != "path" {
			continue
		}
		if !parameter.Required {
			return fmt.Errorf("path parameter %q must be required", parameter.Name)
		}
		pathParameters[parameter.Name] = true
	}
	for _, placeholder := range placeholders {
		if !pathParameters[placeholder] {
			return fmt.Errorf("path placeholder %q has no required path parameter", placeholder)
		}
	}
	for _, parameter := range parameters {
		if parameter.In == "path" && !containsSourceString(placeholders, parameter.Name) {
			return fmt.Errorf("path parameter %q is not present in the path template", parameter.Name)
		}
	}
	return nil
}

func sourcePathTemplateParameters(path string) ([]string, error) {
	seen := map[string]bool{}
	var parameters []string
	for remaining := path; remaining != ""; {
		open := strings.IndexByte(remaining, '{')
		close := strings.IndexByte(remaining, '}')
		if open == -1 {
			if close >= 0 {
				return nil, fmt.Errorf("invalid path placeholder")
			}
			break
		}
		if close >= 0 && close < open {
			return nil, fmt.Errorf("invalid path placeholder")
		}
		remaining = remaining[open+1:]
		close = strings.IndexByte(remaining, '}')
		if close < 0 {
			return nil, fmt.Errorf("invalid path placeholder")
		}
		name := remaining[:close]
		if name == "" || name != strings.TrimSpace(name) || strings.ContainsAny(name, "{}") {
			return nil, fmt.Errorf("invalid path placeholder")
		}
		if !seen[name] {
			seen[name] = true
			parameters = append(parameters, name)
		}
		remaining = remaining[close+1:]
	}
	sort.Strings(parameters)
	return parameters, nil
}

func containsSourceString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sourceSwaggerRequestMediaType(operation, doc map[string]any) (string, error) {
	rawConsumes, exists := operation["consumes"]
	if !exists {
		rawConsumes, exists = doc["consumes"]
	}
	if !exists {
		return "", fmt.Errorf("Swagger request body has no declared consumes media type")
	}
	mediaTypes, err := sourceStringArray(rawConsumes, "Swagger consumes")
	if err != nil {
		return "", err
	}
	if len(mediaTypes) != 1 {
		return "", fmt.Errorf("Swagger request body requires exactly one unambiguous media type")
	}
	if !sourceJSONMediaType(mediaTypes[0]) {
		return "", fmt.Errorf("unsupported request encoding %q", mediaTypes[0])
	}
	return mediaTypes[0], nil
}

func validateBoundedRequestSchema(schema any, limits sourceImportLimits, depth int) error {
	if depth > limits.MaxReferenceDepth {
		return fmt.Errorf("schema depth limit exceeded")
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("encode request schema: %w", err)
	}
	if int64(len(raw)) > limits.MaxSchemaBytes {
		return fmt.Errorf("schema byte limit exceeded")
	}
	object, ok := schema.(map[string]any)
	if !ok {
		return fmt.Errorf("unbounded request schema must be an object")
	}
	for _, composition := range []string{"oneOf", "anyOf", "allOf", "not"} {
		if _, exists := object[composition]; exists {
			return fmt.Errorf("ambiguous request schema uses %s", composition)
		}
	}
	typeName, _ := object["type"].(string)
	if typeName == "" {
		if enum, ok := object["enum"].([]any); ok && len(enum) > 0 {
			return nil
		}
		return fmt.Errorf("unbounded request schema has no type")
	}
	switch typeName {
	case "boolean", "null":
		return nil
	case "string":
		if sourcePositiveInteger(object["maxLength"]) <= 0 {
			return fmt.Errorf("unbounded request schema string has no maxLength")
		}
	case "integer", "number":
		if _, hasMinimum := object["minimum"]; !hasMinimum {
			return fmt.Errorf("unbounded request schema number has no minimum")
		}
		if _, hasMaximum := object["maximum"]; !hasMaximum {
			return fmt.Errorf("unbounded request schema number has no maximum")
		}
	case "array":
		if sourcePositiveInteger(object["maxItems"]) <= 0 {
			return fmt.Errorf("unbounded request schema array has no maxItems")
		}
		items, exists := object["items"]
		if !exists {
			return fmt.Errorf("unbounded request schema array has no items")
		}
		return validateBoundedRequestSchema(items, limits, depth+1)
	case "object":
		additional, exists := object["additionalProperties"]
		if !exists || additional != false {
			return fmt.Errorf("unbounded request schema object has dynamic additionalProperties")
		}
		for _, keyword := range []string{"patternProperties", "propertyNames", "unevaluatedProperties", "dependentSchemas", "dependentRequired"} {
			if _, exists := object[keyword]; exists {
				return fmt.Errorf("unbounded request schema object uses dynamic %s", keyword)
			}
		}
		properties, exists := object["properties"]
		if !exists {
			return fmt.Errorf("unbounded request schema object has no fixed properties")
		}
		propertyMap, ok := properties.(map[string]any)
		if !ok {
			return fmt.Errorf("unbounded request schema object properties are invalid")
		}
		for _, name := range sortedSourceMapKeys(propertyMap) {
			if err := validateBoundedRequestSchema(propertyMap[name], limits, depth+1); err != nil {
				return fmt.Errorf("property %q: %w", name, err)
			}
		}
	default:
		return fmt.Errorf("unsupported request schema type %q", typeName)
	}
	return nil
}

func sourceResponses(operation, doc map[string]any, version string, resolver *sourceReferenceResolver, limits sourceImportLimits) ([]sourceResponseDescriptor, []string, error) {
	rawResponses, ok := operation["responses"].(map[string]any)
	if !ok || len(rawResponses) == 0 {
		return nil, nil, fmt.Errorf("missing responses")
	}
	responses := make([]sourceResponseDescriptor, 0, len(rawResponses))
	mediaSet := map[string]bool{}
	for _, status := range sortedSourceMapKeys(rawResponses) {
		resolved, err := resolver.resolve(rawResponses[status])
		if err != nil {
			return nil, nil, err
		}
		encoded, err := json.Marshal(resolved)
		if err != nil {
			return nil, nil, fmt.Errorf("encode response %q: %w", status, err)
		}
		if int64(len(encoded)) > limits.MaxSchemaBytes {
			return nil, nil, fmt.Errorf("schema byte limit exceeded for response %q", status)
		}
		responses = append(responses, sourceResponseDescriptor{Status: status, Declaration: resolved})
		if response, ok := resolved.(map[string]any); ok {
			if content, ok := response["content"].(map[string]any); ok {
				for mediaType := range content {
					mediaSet[mediaType] = true
				}
			}
		}
	}
	if version == "swagger2" && len(mediaSet) == 0 {
		produces, declared := operation["produces"]
		if !declared {
			produces, declared = doc["produces"]
		}
		if declared {
			mediaTypes, err := sourceStringArray(produces, "Swagger produces")
			if err != nil {
				return nil, nil, err
			}
			for _, mediaType := range mediaTypes {
				mediaSet[mediaType] = true
			}
		}
	}
	mediaTypes := make([]string, 0, len(mediaSet))
	for mediaType := range mediaSet {
		mediaTypes = append(mediaTypes, mediaType)
	}
	sort.Strings(mediaTypes)
	return responses, mediaTypes, nil
}

func sourceOutputClassFor(method string, mediaTypes []string) (sourceOutputClass, error) {
	if method == "head" || len(mediaTypes) == 0 {
		return sourceOutputStatus, nil
	}
	var class sourceOutputClass
	for _, mediaType := range mediaTypes {
		normalizedMediaType := strings.ToLower(strings.TrimSpace(strings.Split(mediaType, ";")[0]))
		if normalizedMediaType == "" {
			return "", fmt.Errorf("response output media type is empty")
		}
		mediaClass := sourceOutputBinary
		if sourceJSONMediaType(mediaType) {
			mediaClass = sourceOutputJSON
		} else if strings.HasPrefix(normalizedMediaType, "text/") {
			mediaClass = sourceOutputText
		}
		if class != "" && class != mediaClass {
			return "", fmt.Errorf("ambiguous response output media types")
		}
		class = mediaClass
	}
	return class, nil
}

func sourcePagination(operation map[string]any, resolver *sourceReferenceResolver) (any, error) {
	var keys []string
	for key := range operation {
		if strings.Contains(strings.ToLower(key), "pagination") {
			keys = append(keys, key)
		}
	}
	if len(keys) == 0 {
		return nil, nil
	}
	sort.Strings(keys)
	if len(keys) > 1 {
		return nil, fmt.Errorf("ambiguous pagination metadata")
	}
	return resolver.resolve(operation[keys[0]])
}

func sourceOperationByteLimits(operation map[string]any) (sourceByteLimits, error) {
	limits := sourceByteLimits{}
	for _, entry := range []struct {
		key  string
		dest *int64
	}{{"x-max-request-bytes", &limits.Request}, {"x-max-response-bytes", &limits.Response}} {
		if raw, exists := operation[entry.key]; exists {
			value := sourcePositiveInteger(raw)
			if value <= 0 {
				return sourceByteLimits{}, fmt.Errorf("%s must be a positive integer", entry.key)
			}
			*entry.dest = value
		}
	}
	return limits, nil
}

func sourceAuthRequirements(operation, doc map[string]any) (sourceAuthDescriptor, error) {
	descriptor := sourceAuthDescriptor{AnyOf: []sourceAuthRequirementGroup{}}
	security, declared := operation["security"]
	if !declared {
		security, declared = doc["security"]
	}
	if !declared {
		return descriptor, nil
	}
	descriptor.Declared = true
	items, ok := security.([]any)
	if !ok {
		return sourceAuthDescriptor{}, fmt.Errorf("security must be an array")
	}
	groups := make([]sourceAuthRequirementGroup, 0, len(items))
	for _, item := range items {
		requirement, ok := item.(map[string]any)
		if !ok {
			return sourceAuthDescriptor{}, fmt.Errorf("security requirement must be an object")
		}
		group := sourceAuthRequirementGroup{AllOf: []sourceAuthScope{}}
		for _, scheme := range sortedSourceMapKeys(requirement) {
			if scheme == "" {
				return sourceAuthDescriptor{}, fmt.Errorf("security requirement has an empty scheme")
			}
			scopes, err := sourceStringArray(requirement[scheme], "security requirement scopes")
			if err != nil {
				return sourceAuthDescriptor{}, err
			}
			group.AllOf = append(group.AllOf, sourceAuthScope{Scheme: scheme, Scopes: scopes})
		}
		groups = append(groups, group)
	}
	sort.Slice(groups, func(i, j int) bool {
		left, right := groups[i].AllOf, groups[j].AllOf
		for index := 0; index < len(left) && index < len(right); index++ {
			if left[index].Scheme != right[index].Scheme {
				return left[index].Scheme < right[index].Scheme
			}
			for scopeIndex := 0; scopeIndex < len(left[index].Scopes) && scopeIndex < len(right[index].Scopes); scopeIndex++ {
				if left[index].Scopes[scopeIndex] != right[index].Scopes[scopeIndex] {
					return left[index].Scopes[scopeIndex] < right[index].Scopes[scopeIndex]
				}
			}
			if len(left[index].Scopes) != len(right[index].Scopes) {
				return len(left[index].Scopes) < len(right[index].Scopes)
			}
		}
		return len(left) < len(right)
	})
	descriptor.AnyOf = groups
	return descriptor, nil
}

func sourceJSONMediaType(mediaType string) bool {
	mediaType = strings.ToLower(strings.TrimSpace(strings.Split(mediaType, ";")[0]))
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func sourcePositiveInteger(value any) int64 {
	switch typed := value.(type) {
	case json.Number:
		integer, err := typed.Int64()
		if err == nil {
			return integer
		}
	case float64:
		if typed == float64(int64(typed)) {
			return int64(typed)
		}
	case int:
		return int64(typed)
	case int64:
		return typed
	case string:
		integer, err := strconv.ParseInt(typed, 10, 64)
		if err == nil {
			return integer
		}
	}
	return 0
}

func sourceBool(value any) bool {
	result, _ := value.(bool)
	return result
}

func sourceStringArray(value any, field string) ([]string, error) {
	items, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array", field)
	}
	stringsOut := make([]string, 0, len(items))
	for _, item := range items {
		stringValue, ok := item.(string)
		if !ok || stringValue == "" || stringValue != strings.TrimSpace(stringValue) {
			return nil, fmt.Errorf("%s must contain non-empty strings", field)
		}
		stringsOut = append(stringsOut, stringValue)
	}
	sort.Strings(stringsOut)
	return stringsOut, nil
}

func sortedSourceMapKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func marshalSourceImportDescriptors(descriptors []sourceOperationDescriptor) ([]byte, error) {
	copyDescriptors := append([]sourceOperationDescriptor(nil), descriptors...)
	sortSourceOperationDescriptors(copyDescriptors)
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(sourceImportDescriptorDocument{SchemaVersion: 1, Operations: copyDescriptors}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

type sourceImportOptions struct {
	Connector string
	DefsDir   string
	Output    string
	Check     bool
}

func runSourceImport(args []string, stdout, stderr io.Writer) int {
	return runSourceImportWithFetcher(args, stdout, stderr, newHTTPSourceImportFetcher(defaultSourceImportLimits()))
}

func runSourceImportWithFetcher(args []string, stdout, stderr io.Writer, fetcher sourceImportFetcher) int {
	for _, arg := range args[1:] {
		if arg == "--help" || arg == "-h" {
			logln(stdout, sourceImportUsage)
			return 0
		}
	}
	opts, err := parseSourceImportOptions(args[1:])
	if err != nil {
		logln(stderr, "connectorgen source-import:", err)
		logln(stderr, sourceImportUsage)
		return 2
	}
	lock, err := loadConnectorSourceImportLock(opts.DefsDir, opts.Connector)
	if err != nil {
		logln(stderr, "connectorgen source-import:", err)
		return 1
	}
	descriptors, err := importSourceLock(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		logln(stderr, "connectorgen source-import:", err)
		return 1
	}
	raw, err := marshalSourceImportDescriptors(descriptors)
	if err != nil {
		logln(stderr, "connectorgen source-import: encode descriptors:", err)
		return 1
	}
	existing, readErr := os.ReadFile(opts.Output)
	if opts.Check {
		if readErr != nil || !bytes.Equal(existing, raw) {
			logln(stderr, "connectorgen source-import: descriptor output has drifted; rerun without --check after source-lock verification")
			return 1
		}
		logln(stdout, fmt.Sprintf("connectorgen source-import: %s, %d operation(s) verified", opts.Connector, len(descriptors)))
		return 0
	}
	if err := os.WriteFile(opts.Output, raw, 0o644); err != nil {
		logln(stderr, "connectorgen source-import: write descriptors:", err)
		return 1
	}
	logln(stdout, fmt.Sprintf("connectorgen source-import: %s, %d operation(s) imported", opts.Connector, len(descriptors)))
	return 0
}

func parseSourceImportOptions(args []string) (sourceImportOptions, error) {
	opts := sourceImportOptions{DefsDir: filepath.Join("internal", "connectors", "defs")}
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--check":
			opts.Check = true
		case "--defs", "--out":
			if i+1 >= len(args) {
				return sourceImportOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			i++
			if arg == "--defs" {
				opts.DefsDir = args[i]
			} else {
				opts.Output = args[i]
			}
		default:
			if strings.HasPrefix(arg, "-") {
				return sourceImportOptions{}, fmt.Errorf("unknown flag %q", arg)
			}
			if opts.Connector != "" {
				return sourceImportOptions{}, fmt.Errorf("only one connector may be imported at a time")
			}
			opts.Connector = arg
		}
	}
	if opts.Connector == "" {
		return sourceImportOptions{}, fmt.Errorf("a connector name is required")
	}
	if err := validateSourceImportConnector(opts.Connector); err != nil {
		return sourceImportOptions{}, err
	}
	if opts.Output == "" {
		return sourceImportOptions{}, fmt.Errorf("--out is required")
	}
	return opts, nil
}

func loadConnectorSourceImportLock(defsDir, connector string) (sourceImportLock, error) {
	if err := validateSourceImportConnector(connector); err != nil {
		return sourceImportLock{}, err
	}
	absDefsDir, err := filepath.Abs(defsDir)
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("resolve connector definitions directory: %w", err)
	}
	resolvedDefsDir, err := filepath.EvalSymlinks(absDefsDir)
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("resolve connector definitions directory: %w", err)
	}
	bundleDir := filepath.Join(absDefsDir, connector)
	sourcesDir := filepath.Join(bundleDir, "sources")
	path := filepath.Join(sourcesDir, connector+"-operation-source-lock.json")
	resolvedBundleDir, err := filepath.EvalSymlinks(bundleDir)
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("resolve connector bundle directory: %w", err)
	}
	if !sourceImportPathWithin(resolvedDefsDir, resolvedBundleDir) {
		return sourceImportLock{}, fmt.Errorf("connector bundle is outside definitions directory")
	}
	resolvedSourcesDir, err := filepath.EvalSymlinks(sourcesDir)
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("resolve connector-owned source directory: %w", err)
	}
	if !sourceImportPathWithin(resolvedBundleDir, resolvedSourcesDir) {
		return sourceImportLock{}, fmt.Errorf("source directory is outside connector-owned bundle")
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("resolve connector-owned source lock: %w", err)
	}
	if !sourceImportPathWithin(resolvedSourcesDir, resolvedPath) {
		return sourceImportLock{}, fmt.Errorf("source lock is outside connector-owned sources directory")
	}
	raw, err := os.ReadFile(resolvedPath)
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("read connector-owned source lock: %w", err)
	}
	return parseSourceImportLock(raw, connector)
}

func sourceImportPathWithin(root, path string) bool {
	relativePath, err := filepath.Rel(root, path)
	return err == nil && relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) && !filepath.IsAbs(relativePath)
}

type httpSourceImportFetcher struct {
	limits sourceImportLimits
	lookup batchArtifactLookupIPAddr
}

func newHTTPSourceImportFetcher(limits sourceImportLimits) sourceImportFetcher {
	return httpSourceImportFetcher{limits: limits, lookup: batchArtifactLookupIPAddr(net.DefaultResolver.LookupIPAddr)}
}

func (fetcher httpSourceImportFetcher) Fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	if err := validateSourceImportLimits(fetcher.limits); err != nil {
		return nil, err
	}
	if fetcher.lookup == nil {
		return nil, fmt.Errorf("source importer has no public address resolver")
	}
	parsed, err := parseBatchArtifactURL(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("validate locked source artifact URL: %w", err)
	}
	if err := validateBatchArtifactRequestURL(ctx, parsed, fetcher.lookup); err != nil {
		return nil, fmt.Errorf("validate locked source artifact destination: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	client := newSourceImportHTTPClient(fetcher.lookup)
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		if closeErr := response.Body.Close(); closeErr != nil {
			return nil, fmt.Errorf("close source-lock artifact response after HTTP %d: %w", response.StatusCode, closeErr)
		}
		return nil, fmt.Errorf("source-lock artifact returned HTTP %d", response.StatusCode)
	}
	raw, readErr := io.ReadAll(io.LimitReader(response.Body, fetcher.limits.MaxArtifactBytes+1))
	closeErr := response.Body.Close()
	if readErr != nil {
		return nil, readErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close source-lock artifact response: %w", closeErr)
	}
	return raw, nil
}

func newSourceImportHTTPClient(lookup batchArtifactLookupIPAddr) *http.Client {
	client := newBatchArtifactHTTPClient(lookup)
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return fmt.Errorf("source-lock artifact redirects are not permitted")
	}
	return client
}
