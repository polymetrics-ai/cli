package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
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
	defaultSourceImportArtifactBytes   = int64(16 << 20)
	defaultSourceImportSchemaBytes     = int64(1 << 20)
	defaultSourceImportDescriptorBytes = int64(32 << 20)
	defaultSourceImportOperations      = 10_000
	defaultSourceImportReferences      = 50_000
	defaultSourceImportReferenceDepth  = 32
	defaultSourceImportSchemaNodes     = 100_000
)

type sourceImportLimits struct {
	MaxArtifactBytes           int64
	MaxSchemaBytes             int64
	MaxResolvedDescriptorBytes int64
	MaxOperations              int
	MaxReferences              int
	MaxReferenceDepth          int
	MaxSchemaNodes             int
}

type sourceDocumentForm struct {
	Family  string
	Version string
}

func (form sourceDocumentForm) isOpenAPI() bool {
	return form.Family == "openapi"
}

func (form sourceDocumentForm) isSwagger2() bool {
	return form.Family == "swagger" && form.Version == "2.0"
}

func (form sourceDocumentForm) allowsReferenceSiblings() bool {
	if !form.isOpenAPI() {
		return false
	}
	major, minor, ok := sourceOpenAPIMajorMinor(form.Version)
	return ok && major == 3 && minor == 1
}

func (form sourceDocumentForm) isOpenAPI31() bool {
	return form.allowsReferenceSiblings()
}

func defaultSourceImportLimits() sourceImportLimits {
	return sourceImportLimits{
		MaxArtifactBytes:           defaultSourceImportArtifactBytes,
		MaxSchemaBytes:             defaultSourceImportSchemaBytes,
		MaxResolvedDescriptorBytes: defaultSourceImportDescriptorBytes,
		MaxOperations:              defaultSourceImportOperations,
		MaxReferences:              defaultSourceImportReferences,
		MaxReferenceDepth:          defaultSourceImportReferenceDepth,
		MaxSchemaNodes:             defaultSourceImportSchemaNodes,
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
	Form     string `json:"form"`
	Version  string `json:"version"`
}

type sourceParameterDescriptor struct {
	Name     string                        `json:"name"`
	Required bool                          `json:"required"`
	Schema   any                           `json:"schema,omitempty"`
	Content  any                           `json:"content,omitempty"`
	Wire     sourceParameterWireDescriptor `json:"wire"`
}

type sourceContractGap struct {
	Foundation string `json:"foundation"`
	Location   string `json:"location"`
	Reason     string `json:"reason"`
}

type sourceRuntimeReachability struct {
	MergeBlocked bool                `json:"merge_blocked"`
	Gaps         []sourceContractGap `json:"gaps,omitempty"`
}

type sourceParameterWireDescriptor struct {
	Style            string              `json:"style,omitempty"`
	Explode          *bool               `json:"explode,omitempty"`
	AllowReserved    *bool               `json:"allow_reserved,omitempty"`
	AllowEmptyValue  *bool               `json:"allow_empty_value,omitempty"`
	CollectionFormat string              `json:"collection_format,omitempty"`
	Gaps             []sourceContractGap `json:"gaps,omitempty"`
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
	Status      string                          `json:"status"`
	Declaration any                             `json:"declaration"`
	Media       []sourceResponseMediaDescriptor `json:"media"`
}

type sourceResponseMediaDescriptor struct {
	MediaType string            `json:"media_type"`
	Class     sourceOutputClass `json:"class"`
}

type sourceOutputClass string

const (
	sourceOutputJSON   sourceOutputClass = "json"
	sourceOutputBinary sourceOutputClass = "binary"
	sourceOutputStatus sourceOutputClass = "status"
	sourceOutputText   sourceOutputClass = "text"
)

type sourceOutputDescriptor struct {
	Class      sourceOutputClass     `json:"class,omitempty"`
	MediaTypes []string              `json:"media_types,omitempty"`
	Success    []sourceOutputVariant `json:"success"`
}

type sourceOutputVariant struct {
	Status    string            `json:"status"`
	MediaType string            `json:"media_type,omitempty"`
	Class     sourceOutputClass `json:"class"`
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
	Servers             sourceServerOverrides      `json:"servers"`
	Runtime             sourceRuntimeReachability  `json:"runtime"`
}

type sourceServerLayer struct {
	Declared bool `json:"declared"`
	Servers  any  `json:"servers,omitempty"`
}

type sourceServerOverrides struct {
	Root       sourceServerLayer          `json:"root"`
	PathItem   sourceServerLayer          `json:"path_item"`
	Operation  sourceServerLayer          `json:"operation"`
	Swagger    *sourceSwaggerRouteBinding `json:"swagger,omitempty"`
	Precedence []string                   `json:"precedence"`
	Gaps       []sourceContractGap        `json:"gaps,omitempty"`
}

type sourceSwaggerRouteBinding struct {
	Declared         bool     `json:"declared"`
	Host             string   `json:"host,omitempty"`
	BasePath         string   `json:"base_path,omitempty"`
	RootSchemes      []string `json:"root_schemes,omitempty"`
	OperationSchemes []string `json:"operation_schemes,omitempty"`
	Schemes          []string `json:"schemes,omitempty"`
	EffectivePath    string   `json:"effective_path"`
	Precedence       []string `json:"precedence"`
}

type sourceInboundEventDescriptor struct {
	Connector      string                    `json:"connector"`
	SourceID       string                    `json:"source_id"`
	ParentSourceID string                    `json:"parent_source_id,omitempty"`
	Kind           string                    `json:"kind"`
	Name           string                    `json:"name"`
	Source         sourceImportSource        `json:"source"`
	Declaration    any                       `json:"declaration"`
	Runtime        sourceRuntimeReachability `json:"runtime"`
}

type sourceExtensionDescriptor struct {
	Location string `json:"location"`
	Value    any    `json:"value"`
}

type sourceImportResult struct {
	Operations    []sourceOperationDescriptor    `json:"operations"`
	InboundEvents []sourceInboundEventDescriptor `json:"inbound_events,omitempty"`
	Extensions    []sourceExtensionDescriptor    `json:"extensions,omitempty"`
}

type sourceImportDescriptorDocument struct {
	SchemaVersion int                            `json:"schema_version"`
	Operations    []sourceOperationDescriptor    `json:"operations"`
	InboundEvents []sourceInboundEventDescriptor `json:"inbound_events,omitempty"`
	Extensions    []sourceExtensionDescriptor    `json:"extensions,omitempty"`
	MergeBlocked  bool                           `json:"merge_blocked"`
	Gaps          []sourceContractGap            `json:"gaps,omitempty"`
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
	if err := decodeSourceJSON(raw, &lock); err != nil {
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
	if artifact.OpenAPI != "" && artifact.Swagger != "" {
		return fmt.Errorf("source lock has ambiguous OpenAPI and Swagger form pins")
	}
	if artifact.OpenAPI != "" {
		major, minor, ok := sourceOpenAPIMajorMinor(artifact.OpenAPI)
		if !ok || major != 3 || (minor != 0 && minor != 1) {
			return fmt.Errorf("source lock has unsupported OpenAPI form pin %q", artifact.OpenAPI)
		}
	}
	if artifact.Swagger != "" && artifact.Swagger != "2.0" {
		return fmt.Errorf("source lock has unsupported Swagger form pin %q", artifact.Swagger)
	}
	return nil
}

func importSourceLocks(ctx context.Context, locks []sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceOperationDescriptor, error) {
	result, err := importSourceLockResults(ctx, locks, fetcher, limits)
	if err != nil {
		return nil, err
	}
	return result.Operations, nil
}

func importSourceLockResults(ctx context.Context, locks []sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) (sourceImportResult, error) {
	if fetcher == nil {
		return sourceImportResult{}, fmt.Errorf("source importer has no fetcher")
	}
	if err := validateSourceImportLimits(limits); err != nil {
		return sourceImportResult{}, err
	}
	orderedLocks := append([]sourceImportLock(nil), locks...)
	sort.Slice(orderedLocks, func(i, j int) bool { return orderedLocks[i].Connector < orderedLocks[j].Connector })
	seenConnectors := map[string]bool{}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{}}
	budget := sourceImportBudget{limit: sourceResolvedDescriptorLimit(limits)}
	for _, lock := range orderedLocks {
		if seenConnectors[lock.Connector] {
			return sourceImportResult{}, fmt.Errorf("duplicate source-lock connector %q", lock.Connector)
		}
		seenConnectors[lock.Connector] = true
		imported, err := importSourceLockResultWithBudget(ctx, lock, fetcher, limits, &budget)
		if err != nil {
			return sourceImportResult{}, err
		}
		if len(result.Operations)+len(imported.Operations) > limits.MaxOperations {
			return sourceImportResult{}, fmt.Errorf("operation count limit exceeded")
		}
		if len(result.InboundEvents)+len(imported.InboundEvents) > limits.MaxOperations {
			return sourceImportResult{}, fmt.Errorf("inbound event count limit exceeded")
		}
		result.Operations = append(result.Operations, imported.Operations...)
		result.InboundEvents = append(result.InboundEvents, imported.InboundEvents...)
		result.Extensions = append(result.Extensions, imported.Extensions...)
	}
	sortSourceOperationDescriptors(result.Operations)
	sortSourceInboundEventDescriptors(result.InboundEvents)
	sortSourceExtensions(result.Extensions)
	if err := validateSourceImportResultIdentities(result); err != nil {
		return sourceImportResult{}, err
	}
	return result, nil
}

func validateSourceImportLimits(limits sourceImportLimits) error {
	if limits.MaxArtifactBytes <= 0 || limits.MaxSchemaBytes <= 0 || limits.MaxOperations <= 0 || limits.MaxReferences <= 0 || limits.MaxReferenceDepth <= 0 || limits.MaxResolvedDescriptorBytes < 0 {
		return fmt.Errorf("source import limits must all be positive")
	}
	return nil
}

func importSourceLock(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) ([]sourceOperationDescriptor, error) {
	result, err := importSourceLockResult(ctx, lock, fetcher, limits)
	if err != nil {
		return nil, err
	}
	return result.Operations, nil
}

func importSourceLockResult(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits) (sourceImportResult, error) {
	budget := sourceImportBudget{limit: sourceResolvedDescriptorLimit(limits)}
	return importSourceLockResultWithBudget(ctx, lock, fetcher, limits, &budget)
}

func importSourceLockResultWithBudget(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits, budget *sourceImportBudget) (sourceImportResult, error) {
	if err := validateSourceImportLimits(limits); err != nil {
		return sourceImportResult{}, err
	}
	if fetcher == nil {
		return sourceImportResult{}, fmt.Errorf("source importer has no fetcher")
	}
	if budget == nil {
		return sourceImportResult{}, fmt.Errorf("source importer has no descriptor budget")
	}
	if lock.SchemaVersion <= 0 {
		return sourceImportResult{}, fmt.Errorf("source lock has invalid schema version")
	}
	if err := validateSourceImportConnector(lock.Connector); err != nil {
		return sourceImportResult{}, err
	}
	if err := validateSourceImportArtifact(lock.Rest); err != nil {
		return sourceImportResult{}, err
	}
	if lock.Rest.Bytes > limits.MaxArtifactBytes {
		return sourceImportResult{}, fmt.Errorf("artifact byte limit exceeded by source lock")
	}
	raw, err := fetcher.Fetch(ctx, lock.Rest.SourceURL)
	if err != nil {
		return sourceImportResult{}, fmt.Errorf("fetch locked source artifact: %w", err)
	}
	if int64(len(raw)) > limits.MaxArtifactBytes {
		return sourceImportResult{}, fmt.Errorf("artifact byte limit exceeded")
	}
	actualDigest := sha256.Sum256(raw)
	if int64(len(raw)) != lock.Rest.Bytes || !strings.EqualFold(hex.EncodeToString(actualDigest[:]), lock.Rest.SHA256) {
		return sourceImportResult{}, fmt.Errorf("source-lock refresh required: fetched artifact does not match locked bytes and SHA-256")
	}
	doc, form, err := parseSourceImportDocument(raw)
	if err != nil {
		return sourceImportResult{}, err
	}
	if err := validateSourceImportArtifactForm(lock.Rest, form); err != nil {
		return sourceImportResult{}, err
	}
	resolver := sourceReferenceResolver{root: doc, limits: limits, form: form}
	result, err := importSourceDocumentResult(lock, doc, form, &resolver, limits, budget)
	if err != nil {
		return sourceImportResult{}, err
	}
	sortSourceOperationDescriptors(result.Operations)
	sortSourceInboundEventDescriptors(result.InboundEvents)
	sortSourceExtensions(result.Extensions)
	if err := validateSourceImportResultIdentities(result); err != nil {
		return sourceImportResult{}, err
	}
	return result, nil
}

func validateSourceImportResultIdentities(result sourceImportResult) error {
	seen := map[string]string{}
	for _, descriptor := range result.Operations {
		key := descriptor.Connector + "\x00" + descriptor.SourceID
		if existing, exists := seen[key]; exists {
			return fmt.Errorf("duplicate source identity %q for connector %q (%s and operation)", descriptor.SourceID, descriptor.Connector, existing)
		}
		seen[key] = "operation"
	}
	for _, event := range result.InboundEvents {
		key := event.Connector + "\x00" + event.SourceID
		if existing, exists := seen[key]; exists {
			return fmt.Errorf("duplicate source identity %q for connector %q (%s and inbound event)", event.SourceID, event.Connector, existing)
		}
		seen[key] = "inbound event"
	}
	return nil
}

func sourceResolvedDescriptorLimit(limits sourceImportLimits) int64 {
	if limits.MaxResolvedDescriptorBytes > 0 {
		return limits.MaxResolvedDescriptorBytes
	}
	return defaultSourceImportDescriptorBytes
}

type sourceImportBudget struct {
	limit int64
	used  int64
}

type sourceResponseExpansionBudget struct {
	limit int64
	used  int64
}

func (budget *sourceImportBudget) add(value any, label string) error {
	raw, err := sourceMarshalCompact(value)
	if err != nil {
		return fmt.Errorf("encode %s descriptor: %w", label, err)
	}
	return budget.reserve(int64(len(raw)), label)
}

func (budget *sourceImportBudget) reserve(bytes int64, label string) error {
	if bytes < 0 || bytes > budget.limit || budget.used > budget.limit-bytes {
		return fmt.Errorf("resolved descriptor byte limit exceeded while retaining %s", label)
	}
	budget.used += bytes
	return nil
}

func (budget *sourceImportBudget) remaining() int64 {
	if budget == nil || budget.used >= budget.limit {
		return 0
	}
	return budget.limit - budget.used
}

func (budget *sourceResponseExpansionBudget) check(bytes int64) error {
	if budget == nil || bytes < 0 || bytes > budget.limit || budget.used > budget.limit-bytes {
		return fmt.Errorf("resolved descriptor byte limit exceeded while retaining response")
	}
	return nil
}

func (budget *sourceResponseExpansionBudget) reserve(bytes int64) error {
	if err := budget.check(bytes); err != nil {
		return err
	}
	budget.used += bytes
	return nil
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

func parseSourceImportDocument(raw []byte) (map[string]any, sourceDocumentForm, error) {
	var value any
	if err := decodeSourceJSON(raw, &value); err != nil {
		var yamlValue any
		if yamlErr := decodeSourceYAML(raw, &yamlValue); yamlErr != nil {
			return nil, sourceDocumentForm{}, fmt.Errorf("parse source artifact as JSON or YAML: JSON: %v; YAML: %w", err, yamlErr)
		}
		canonical, yamlErr := json.Marshal(normalizeSourceYAML(yamlValue))
		if yamlErr != nil {
			return nil, sourceDocumentForm{}, fmt.Errorf("normalize YAML source artifact: %w", yamlErr)
		}
		if err := decodeSourceJSON(canonical, &value); err != nil {
			return nil, sourceDocumentForm{}, fmt.Errorf("parse normalized YAML source artifact: %w", err)
		}
	}
	doc, ok := value.(map[string]any)
	if !ok {
		return nil, sourceDocumentForm{}, fmt.Errorf("source artifact root must be an object")
	}
	_, hasOpenAPI := doc["openapi"]
	_, hasSwagger := doc["swagger"]
	if hasOpenAPI && hasSwagger {
		return nil, sourceDocumentForm{}, fmt.Errorf("ambiguous source artifact form: OpenAPI and Swagger declarations cannot be combined")
	}
	if hasOpenAPI {
		version, ok := doc["openapi"].(string)
		if !ok {
			return nil, sourceDocumentForm{}, fmt.Errorf("OpenAPI version must be a string")
		}
		major, minor, ok := sourceOpenAPIMajorMinor(version)
		if !ok || major != 3 || (minor != 0 && minor != 1) {
			return nil, sourceDocumentForm{}, fmt.Errorf("unsupported OpenAPI version %q", version)
		}
		return doc, sourceDocumentForm{Family: "openapi", Version: version}, nil
	}
	if hasSwagger {
		version, ok := doc["swagger"].(string)
		if !ok || version != "2.0" {
			return nil, sourceDocumentForm{}, fmt.Errorf("unsupported Swagger version %v", doc["swagger"])
		}
		return doc, sourceDocumentForm{Family: "swagger", Version: version}, nil
	}
	return nil, sourceDocumentForm{}, fmt.Errorf("unsupported source artifact form: require OpenAPI 3.0/3.1 or Swagger 2.0")
}

func sourceOpenAPIMajorMinor(version string) (int, int, bool) {
	if version == "" || strings.TrimSpace(version) != version {
		return 0, 0, false
	}
	core := version
	if buildIndex := strings.IndexByte(core, '+'); buildIndex >= 0 {
		if buildIndex == len(core)-1 || strings.Contains(core[buildIndex+1:], "+") || !sourceSemverIdentifiers(core[buildIndex+1:]) {
			return 0, 0, false
		}
		core = core[:buildIndex]
	}
	if prereleaseIndex := strings.IndexByte(core, '-'); prereleaseIndex >= 0 {
		if prereleaseIndex == len(core)-1 || !sourceSemverIdentifiers(core[prereleaseIndex+1:]) {
			return 0, 0, false
		}
		core = core[:prereleaseIndex]
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return 0, 0, false
	}
	major, ok := sourceSemverInteger(parts[0])
	if !ok {
		return 0, 0, false
	}
	minor, ok := sourceSemverInteger(parts[1])
	if !ok {
		return 0, 0, false
	}
	if _, ok := sourceSemverInteger(parts[2]); !ok {
		return 0, 0, false
	}
	return major, minor, true
}

func sourceSemverInteger(value string) (int, bool) {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0, false
		}
	}
	parsed, err := strconv.Atoi(value)
	return parsed, err == nil
}

func sourceSemverIdentifiers(value string) bool {
	if value == "" {
		return false
	}
	for _, identifier := range strings.Split(value, ".") {
		if identifier == "" {
			return false
		}
		numeric := true
		for _, character := range identifier {
			if (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') && (character < '0' || character > '9') && character != '-' {
				return false
			}
			if character < '0' || character > '9' {
				numeric = false
			}
		}
		if numeric && len(identifier) > 1 && identifier[0] == '0' {
			return false
		}
	}
	return true
}

func validateSourceImportArtifactForm(artifact sourceImportArtifact, form sourceDocumentForm) error {
	if artifact.OpenAPI != "" && artifact.Swagger != "" {
		return fmt.Errorf("source lock has ambiguous OpenAPI and Swagger form pins")
	}
	if artifact.OpenAPI != "" && (!form.isOpenAPI() || artifact.OpenAPI != form.Version) {
		return fmt.Errorf("source lock OpenAPI pin %q does not match source form %s %q", artifact.OpenAPI, form.Family, form.Version)
	}
	if artifact.Swagger != "" && (!form.isSwagger2() || artifact.Swagger != form.Version) {
		return fmt.Errorf("source lock Swagger pin %q does not match source form %s %q", artifact.Swagger, form.Family, form.Version)
	}
	return nil
}

func decodeSourceJSON(raw []byte, target any) error {
	validator := json.NewDecoder(bytes.NewReader(raw))
	validator.UseNumber()
	if err := validateSourceJSONValue(validator, ""); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return sourceJSONEOF(decoder)
}

func validateSourceJSONValue(decoder *json.Decoder, pointer string) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, isDelimiter := token.(json.Delim)
	if !isDelimiter {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			rawKey, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := rawKey.(string)
			if !ok {
				return fmt.Errorf("JSON object key at %s is not a string", pointer)
			}
			childPointer := sourceJSONPointer(pointer, key)
			if seen[key] {
				return fmt.Errorf("duplicate JSON object member at %s", childPointer)
			}
			seen[key] = true
			if err := validateSourceJSONValue(decoder, childPointer); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim('}') {
			return fmt.Errorf("JSON object at %s is not closed", pointer)
		}
	case '[':
		for index := 0; decoder.More(); index++ {
			if err := validateSourceJSONValue(decoder, sourceJSONPointer(pointer, strconv.Itoa(index))); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil {
			return err
		}
		if closing != json.Delim(']') {
			return fmt.Errorf("JSON array at %s is not closed", pointer)
		}
	default:
		return fmt.Errorf("unexpected JSON delimiter %q at %s", delimiter, pointer)
	}
	return nil
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

func decodeSourceYAML(raw []byte, target any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return err
	}
	if err := validateSourceYAMLNode(&document, ""); err != nil {
		return err
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("multiple YAML documents are unsupported")
	}
	value, err := sourceYAMLNodeValue(&document, "")
	if err != nil {
		return err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("encode YAML source artifact: %w", err)
	}
	return decodeSourceJSON(canonical, target)
}

func validateSourceYAMLNode(node *yaml.Node, pointer string) error {
	if node == nil {
		return fmt.Errorf("YAML source artifact is empty")
	}
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return fmt.Errorf("YAML source artifact has no single document root")
		}
		return validateSourceYAMLNode(node.Content[0], pointer)
	case yaml.MappingNode:
		seen := map[string]bool{}
		for index := 0; index < len(node.Content); index += 2 {
			if index+1 >= len(node.Content) {
				return fmt.Errorf("YAML mapping at %s has an incomplete entry", pointer)
			}
			key := node.Content[index]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return fmt.Errorf("YAML mapping key at %s must be a string", pointer)
			}
			childPointer := sourceJSONPointer(pointer, key.Value)
			if seen[key.Value] {
				return fmt.Errorf("duplicate YAML mapping key at %s", childPointer)
			}
			seen[key.Value] = true
			if err := validateSourceYAMLNode(node.Content[index+1], childPointer); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for index, child := range node.Content {
			if err := validateSourceYAMLNode(child, sourceJSONPointer(pointer, strconv.Itoa(index))); err != nil {
				return err
			}
		}
	case yaml.AliasNode:
		return fmt.Errorf("YAML aliases are unsupported at %s", pointer)
	}
	return nil
}

func sourceYAMLNodeValue(node *yaml.Node, pointer string) (any, error) {
	if node == nil {
		return nil, fmt.Errorf("YAML source artifact is empty")
	}
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return nil, fmt.Errorf("YAML source artifact has no single document root")
		}
		return sourceYAMLNodeValue(node.Content[0], pointer)
	case yaml.MappingNode:
		out := make(map[string]any, len(node.Content)/2)
		for index := 0; index < len(node.Content); index += 2 {
			key := node.Content[index]
			if key.Kind != yaml.ScalarNode || key.Tag != "!!str" {
				return nil, fmt.Errorf("YAML mapping key at %s must be a string", pointer)
			}
			childPointer := sourceJSONPointer(pointer, key.Value)
			child, err := sourceYAMLNodeValue(node.Content[index+1], childPointer)
			if err != nil {
				return nil, err
			}
			out[key.Value] = child
		}
		return out, nil
	case yaml.SequenceNode:
		out := make([]any, len(node.Content))
		for index, child := range node.Content {
			value, err := sourceYAMLNodeValue(child, sourceJSONPointer(pointer, strconv.Itoa(index)))
			if err != nil {
				return nil, err
			}
			out[index] = value
		}
		return out, nil
	case yaml.AliasNode:
		return nil, fmt.Errorf("YAML aliases are unsupported at %s", pointer)
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!null":
			return nil, nil
		case "!!bool":
			value, err := strconv.ParseBool(strings.ToLower(node.Value))
			if err != nil {
				return nil, fmt.Errorf("YAML boolean at %s is invalid", pointer)
			}
			return value, nil
		case "!!int", "!!float":
			var parsed any
			if err := decodeSourceJSON([]byte(node.Value), &parsed); err != nil {
				return nil, fmt.Errorf("YAML number at %s is not a finite JSON number", pointer)
			}
			number, ok := parsed.(json.Number)
			if !ok {
				return nil, fmt.Errorf("YAML number at %s is not a JSON number", pointer)
			}
			if _, ok := sourceExactDecimal(string(number)); !ok {
				return nil, fmt.Errorf("YAML number at %s is not a finite exact decimal", pointer)
			}
			return number, nil
		default:
			return node.Value, nil
		}
	default:
		return nil, fmt.Errorf("YAML node at %s is unsupported", pointer)
	}
}

func sourceJSONPointer(parent, segment string) string {
	segment = strings.ReplaceAll(strings.ReplaceAll(segment, "~", "~0"), "/", "~1")
	return parent + "/" + segment
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
	root              map[string]any
	limits            sourceImportLimits
	form              sourceDocumentForm
	referenceIndex    *sourceReferenceIndex
	references        int
	expansion         sourceSchemaExpansionBudget
	responseExpansion sourceResponseExpansionBudget
	responseScope     *sourceResponseExpansionBudget
}

type sourceReferenceKind string

const (
	sourceReferencePathItem    sourceReferenceKind = "path item"
	sourceReferenceParameter   sourceReferenceKind = "parameter"
	sourceReferenceRequestBody sourceReferenceKind = "request body"
	sourceReferenceResponse    sourceReferenceKind = "response"
	sourceReferenceHeader      sourceReferenceKind = "header"
	sourceReferenceSchema      sourceReferenceKind = "schema"
	sourceReferenceCallback    sourceReferenceKind = "callback"
	sourceReferenceLink        sourceReferenceKind = "link"
	sourceReferenceExample     sourceReferenceKind = "example"
	sourceReferenceSecurity    sourceReferenceKind = "security scheme"
)

type sourceReferenceIndex struct {
	positions     map[string]sourceReferenceKind
	limits        sourceImportLimits
	positionBytes int64
}

func (index *sourceReferenceIndex) add(pointer string, kind sourceReferenceKind) error {
	pointer = sourceReferenceIndexPointer(pointer)
	if existing, exists := index.positions[pointer]; exists && existing != kind {
		return fmt.Errorf("ambiguous source grammar position %q is both %s and %s", pointer, existing, kind)
	}
	if _, exists := index.positions[pointer]; exists {
		return nil
	}
	bytes := sourceReferenceIndexEntryBytes(pointer)
	if err := index.checkAddition(1, bytes); err != nil {
		return err
	}
	index.positions[pointer] = kind
	index.positionBytes += bytes
	return nil
}

func sourceReferenceIndexPointer(pointer string) string {
	if strings.HasPrefix(pointer, "#") {
		return pointer
	}
	return "#" + pointer
}

func sourceReferenceIndexEntryBytes(pointer string) int64 {
	bytes := sourceStructuralStringBytes(pointer)
	if bytes > math.MaxInt64-32 {
		return math.MaxInt64
	}
	return bytes + 32
}

func sourceReferenceIndexByteLimit(limits sourceImportLimits) int64 {
	limit := limits.MaxArtifactBytes
	if limit <= 0 {
		limit = defaultSourceImportArtifactBytes
	}
	if descriptorLimit := sourceResolvedDescriptorLimit(limits); descriptorLimit > 0 && descriptorLimit < limit {
		limit = descriptorLimit
	}
	return limit
}

func (index *sourceReferenceIndex) checkAddition(count int, bytes int64) error {
	if count < 0 || count > sourceSchemaNodeLimit(index.limits)-len(index.positions) {
		return fmt.Errorf("source grammar position limit exceeded")
	}
	limit := sourceReferenceIndexByteLimit(index.limits)
	if bytes < 0 || bytes > limit || index.positionBytes > limit-bytes {
		return fmt.Errorf("source grammar position byte limit exceeded")
	}
	return nil
}

func (index *sourceReferenceIndex) preflightEntries(pointer string, entries map[string]any, skipExtensions bool) error {
	remainingPositions := sourceSchemaNodeLimit(index.limits) - len(index.positions)
	if remainingPositions < 0 {
		return fmt.Errorf("source grammar position limit exceeded")
	}
	count := 0
	for name := range entries {
		if skipExtensions && strings.HasPrefix(name, "x-") {
			continue
		}
		childPointer := sourceReferenceIndexPointer(sourceJSONPointer(pointer, name))
		if _, exists := index.positions[childPointer]; exists {
			continue
		}
		if count == remainingPositions {
			return fmt.Errorf("source grammar position limit exceeded")
		}
		count++
	}
	remainingBytes := sourceReferenceIndexByteLimit(index.limits) - index.positionBytes
	if remainingBytes < 0 {
		return fmt.Errorf("source grammar position byte limit exceeded")
	}
	used := int64(0)
	for name := range entries {
		if skipExtensions && strings.HasPrefix(name, "x-") {
			continue
		}
		childPointer := sourceReferenceIndexPointer(sourceJSONPointer(pointer, name))
		if _, exists := index.positions[childPointer]; exists {
			continue
		}
		entryBytes := sourceReferenceIndexEntryBytes(childPointer)
		if entryBytes == math.MaxInt64 || entryBytes > remainingBytes-used {
			return fmt.Errorf("source grammar position byte limit exceeded")
		}
		used += entryBytes
	}
	return nil
}

func (index *sourceReferenceIndex) preflightArrayEntries(pointer string, count int) error {
	if count < 0 || count > sourceSchemaNodeLimit(index.limits)-len(index.positions) {
		return fmt.Errorf("source grammar position limit exceeded")
	}
	remainingBytes := sourceReferenceIndexByteLimit(index.limits) - index.positionBytes
	if remainingBytes < 0 {
		return fmt.Errorf("source grammar position byte limit exceeded")
	}
	used := int64(0)
	for position := 0; position < count; position++ {
		entryBytes := sourceReferenceIndexEntryBytes(sourceReferenceIndexPointer(sourceJSONPointer(pointer, strconv.Itoa(position))))
		if entryBytes == math.MaxInt64 || entryBytes > remainingBytes-used {
			return fmt.Errorf("source grammar position byte limit exceeded")
		}
		used += entryBytes
	}
	return nil
}

type sourceSchemaExpansionBudget struct {
	used  int64
	nodes int
}

type sourceSchemaExpansionLimitError struct {
	Scope string
	Label string
}

type sourceSchemaStructureLimitError struct {
	Scope string
}

func (err *sourceSchemaStructureLimitError) Error() string {
	if err.Scope == "node" {
		return "schema node limit exceeded"
	}
	return "schema structural byte limit exceeded"
}

func (err *sourceSchemaExpansionLimitError) Error() string {
	if err.Scope == "object" {
		return fmt.Sprintf("schema byte limit exceeded during resolved expansion while retaining %s", err.Label)
	}
	if err.Scope == "node" {
		return fmt.Sprintf("resolved schema expansion node limit exceeded while retaining %s", err.Label)
	}
	return fmt.Sprintf("resolved schema expansion %s byte limit exceeded while retaining %s", err.Scope, err.Label)
}

type sourceSchemaGrammar struct {
	mapChildren        []string
	schemaChildren     []string
	arrayChildren      []string
	dependencyChildren bool
}

func sourceSchemaGrammarFor(form sourceDocumentForm) sourceSchemaGrammar {
	if form.isSwagger2() {
		return sourceSchemaGrammar{
			mapChildren:    []string{"properties"},
			schemaChildren: []string{"items", "additionalProperties"},
			arrayChildren:  []string{"allOf"},
		}
	}
	if form.isOpenAPI() && !form.isOpenAPI31() {
		return sourceSchemaGrammar{
			mapChildren:    []string{"properties"},
			schemaChildren: []string{"items", "additionalProperties", "not"},
			arrayChildren:  []string{"allOf", "anyOf", "oneOf"},
		}
	}
	return sourceSchemaGrammar{
		mapChildren:        []string{"$defs", "definitions", "properties", "patternProperties", "dependentSchemas"},
		schemaChildren:     []string{"items", "additionalProperties", "contains", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems", "contentSchema"},
		arrayChildren:      []string{"allOf", "anyOf", "oneOf", "prefixItems"},
		dependencyChildren: true,
	}
}

func (grammar sourceSchemaGrammar) handles(key string) bool {
	for _, children := range [][]string{grammar.mapChildren, grammar.schemaChildren, grammar.arrayChildren} {
		for _, candidate := range children {
			if key == candidate {
				return true
			}
		}
	}
	return grammar.dependencyChildren && key == "dependencies"
}

var sourceSchemaBearingKeywords = []string{
	"$defs", "definitions", "properties", "patternProperties", "dependentSchemas", "dependencies",
	"items", "additionalProperties", "contains", "not", "if", "then", "else", "propertyNames", "unevaluatedProperties", "unevaluatedItems", "contentSchema",
	"allOf", "anyOf", "oneOf", "prefixItems",
}

func sourceBuildReferenceIndex(root map[string]any, form sourceDocumentForm, limits sourceImportLimits) (*sourceReferenceIndex, error) {
	index := &sourceReferenceIndex{positions: map[string]sourceReferenceKind{}, limits: limits}
	if rawPaths, exists := root["paths"]; exists {
		if err := sourceIndexPathItems(index, rawPaths, sourceJSONPointer("", "paths"), form); err != nil {
			return nil, err
		}
	}
	if form.isOpenAPI() {
		if rawWebhooks, exists := root["webhooks"]; exists && form.isOpenAPI31() {
			if err := sourceIndexPathItems(index, rawWebhooks, sourceJSONPointer("", "webhooks"), form); err != nil {
				return nil, err
			}
		}
		if rawComponents, exists := root["components"]; exists {
			if err := sourceIndexOpenAPIComponents(index, rawComponents, sourceJSONPointer("", "components"), form); err != nil {
				return nil, err
			}
		}
	} else if form.isSwagger2() {
		if err := sourceIndexSwaggerComponents(index, root, form); err != nil {
			return nil, err
		}
	}
	return index, nil
}

func sourceIndexOpenAPIComponents(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	components, err := sourceReferenceObject(value, "components")
	if err != nil {
		return err
	}
	entries := []struct {
		name  string
		index func(any, string) error
	}{
		{name: "schemas", index: func(child any, childPointer string) error { return sourceIndexSchema(index, child, childPointer, form) }},
		{name: "responses", index: func(child any, childPointer string) error {
			return sourceIndexResponse(index, child, childPointer, form)
		}},
		{name: "parameters", index: func(child any, childPointer string) error {
			return sourceIndexParameter(index, child, childPointer, form)
		}},
		{name: "examples", index: func(child any, childPointer string) error {
			return sourceIndexExample(index, child, childPointer, form)
		}},
		{name: "requestBodies", index: func(child any, childPointer string) error {
			return sourceIndexRequestBody(index, child, childPointer, form)
		}},
		{name: "headers", index: func(child any, childPointer string) error { return sourceIndexHeader(index, child, childPointer, form) }},
		{name: "securitySchemes", index: func(child any, childPointer string) error {
			return sourceIndexSecurityScheme(index, child, childPointer)
		}},
		{name: "links", index: func(child any, childPointer string) error { return sourceIndexLink(index, child, childPointer) }},
		{name: "callbacks", index: func(child any, childPointer string) error {
			return sourceIndexCallback(index, child, childPointer, form)
		}},
	}
	if form.isOpenAPI31() {
		entries = append(entries, struct {
			name  string
			index func(any, string) error
		}{name: "pathItems", index: func(child any, childPointer string) error {
			return sourceIndexPathItem(index, child, childPointer, form)
		}})
	}
	for _, entry := range entries {
		raw, exists := components[entry.name]
		if !exists {
			continue
		}
		if err := sourceIndexEntries(index, raw, sourceJSONPointer(pointer, entry.name), "components."+entry.name, false, entry.index); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexSwaggerComponents(index *sourceReferenceIndex, root map[string]any, form sourceDocumentForm) error {
	entries := []struct {
		name  string
		index func(any, string) error
	}{
		{name: "definitions", index: func(child any, childPointer string) error { return sourceIndexSchema(index, child, childPointer, form) }},
		{name: "parameters", index: func(child any, childPointer string) error {
			return sourceIndexParameter(index, child, childPointer, form)
		}},
		{name: "responses", index: func(child any, childPointer string) error {
			return sourceIndexResponse(index, child, childPointer, form)
		}},
		{name: "securityDefinitions", index: func(child any, childPointer string) error {
			return sourceIndexSecurityScheme(index, child, childPointer)
		}},
	}
	for _, entry := range entries {
		raw, exists := root[entry.name]
		if !exists {
			continue
		}
		if err := sourceIndexEntries(index, raw, sourceJSONPointer("", entry.name), entry.name, false, entry.index); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexEntries(index *sourceReferenceIndex, value any, pointer, label string, skipExtensions bool, indexChild func(any, string) error) error {
	entries, err := sourceReferenceObject(value, label)
	if err != nil {
		return err
	}
	if err := index.preflightEntries(pointer, entries, skipExtensions); err != nil {
		return err
	}
	for _, name := range sortedSourceMapKeys(entries) {
		if skipExtensions && strings.HasPrefix(name, "x-") {
			continue
		}
		if err := indexChild(entries[name], sourceJSONPointer(pointer, name)); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexPathItems(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	return sourceIndexEntries(index, value, pointer, "path items", true, func(child any, childPointer string) error {
		return sourceIndexPathItem(index, child, childPointer, form)
	})
}

func sourceIndexPathItem(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferencePathItem); err != nil {
		return err
	}
	pathItem, err := sourceReferenceObject(value, "path item")
	if err != nil {
		return err
	}
	if parameters, exists := pathItem["parameters"]; exists {
		if err := sourceIndexParameters(index, parameters, sourceJSONPointer(pointer, "parameters"), form); err != nil {
			return err
		}
	}
	if form.isOpenAPI() {
		if callbacks, exists := pathItem["callbacks"]; exists {
			if err := sourceIndexEntries(index, callbacks, sourceJSONPointer(pointer, "callbacks"), "path item callbacks", true, func(child any, childPointer string) error {
				return sourceIndexCallback(index, child, childPointer, form)
			}); err != nil {
				return err
			}
		}
	}
	for _, method := range sourceHTTPMethods {
		operation, exists := pathItem[method]
		if !exists {
			continue
		}
		if err := sourceIndexOperation(index, operation, sourceJSONPointer(pointer, method), form); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexOperation(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	operation, err := sourceReferenceObject(value, "operation")
	if err != nil {
		return err
	}
	if parameters, exists := operation["parameters"]; exists {
		if err := sourceIndexParameters(index, parameters, sourceJSONPointer(pointer, "parameters"), form); err != nil {
			return err
		}
	}
	if form.isOpenAPI() {
		if requestBody, exists := operation["requestBody"]; exists {
			if err := sourceIndexRequestBody(index, requestBody, sourceJSONPointer(pointer, "requestBody"), form); err != nil {
				return err
			}
		}
	}
	if responses, exists := operation["responses"]; exists {
		if err := sourceIndexEntries(index, responses, sourceJSONPointer(pointer, "responses"), "operation responses", true, func(child any, childPointer string) error {
			return sourceIndexResponse(index, child, childPointer, form)
		}); err != nil {
			return err
		}
	}
	if form.isOpenAPI() {
		if callbacks, exists := operation["callbacks"]; exists {
			if err := sourceIndexEntries(index, callbacks, sourceJSONPointer(pointer, "callbacks"), "operation callbacks", true, func(child any, childPointer string) error {
				return sourceIndexCallback(index, child, childPointer, form)
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

func sourceIndexParameters(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	parameters, ok := value.([]any)
	if !ok {
		return fmt.Errorf("parameters must be an array")
	}
	if err := index.preflightArrayEntries(pointer, len(parameters)); err != nil {
		return err
	}
	for position, parameter := range parameters {
		if err := sourceIndexParameter(index, parameter, sourceJSONPointer(pointer, strconv.Itoa(position)), form); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexParameter(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceParameter); err != nil {
		return err
	}
	parameter, err := sourceReferenceObject(value, "parameter")
	if err != nil {
		return err
	}
	if schema, exists := parameter["schema"]; exists {
		if err := sourceIndexSchema(index, schema, sourceJSONPointer(pointer, "schema"), form); err != nil {
			return err
		}
	}
	if form.isOpenAPI() {
		if content, exists := parameter["content"]; exists {
			if err := sourceIndexContent(index, content, sourceJSONPointer(pointer, "content"), form, false); err != nil {
				return err
			}
		}
		if examples, exists := parameter["examples"]; exists {
			if err := sourceIndexExamples(index, examples, sourceJSONPointer(pointer, "examples"), form); err != nil {
				return err
			}
		}
	} else if items, exists := parameter["items"]; exists {
		if err := sourceIndexSchema(index, items, sourceJSONPointer(pointer, "items"), form); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexRequestBody(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceRequestBody); err != nil {
		return err
	}
	requestBody, err := sourceReferenceObject(value, "request body")
	if err != nil {
		return err
	}
	if content, exists := requestBody["content"]; exists {
		return sourceIndexContent(index, content, sourceJSONPointer(pointer, "content"), form, true)
	}
	return nil
}

func sourceIndexResponse(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceResponse); err != nil {
		return err
	}
	response, err := sourceReferenceObject(value, "response")
	if err != nil {
		return err
	}
	if form.isSwagger2() {
		if schema, exists := response["schema"]; exists {
			if err := sourceIndexSchema(index, schema, sourceJSONPointer(pointer, "schema"), form); err != nil {
				return err
			}
		}
		if headers, exists := response["headers"]; exists {
			if err := sourceIndexSwaggerHeaders(index, headers, sourceJSONPointer(pointer, "headers"), form); err != nil {
				return err
			}
		}
		return nil
	}
	if content, exists := response["content"]; exists {
		if err := sourceIndexContent(index, content, sourceJSONPointer(pointer, "content"), form, false); err != nil {
			return err
		}
	}
	if headers, exists := response["headers"]; exists {
		if err := sourceIndexEntries(index, headers, sourceJSONPointer(pointer, "headers"), "response headers", false, func(child any, childPointer string) error {
			return sourceIndexHeader(index, child, childPointer, form)
		}); err != nil {
			return err
		}
	}
	if links, exists := response["links"]; exists {
		if err := sourceIndexLinks(index, links, sourceJSONPointer(pointer, "links")); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexHeader(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceHeader); err != nil {
		return err
	}
	header, err := sourceReferenceObject(value, "header")
	if err != nil {
		return err
	}
	if schema, exists := header["schema"]; exists {
		if err := sourceIndexSchema(index, schema, sourceJSONPointer(pointer, "schema"), form); err != nil {
			return err
		}
	}
	if content, exists := header["content"]; exists {
		if err := sourceIndexContent(index, content, sourceJSONPointer(pointer, "content"), form, false); err != nil {
			return err
		}
	}
	if examples, exists := header["examples"]; exists {
		return sourceIndexExamples(index, examples, sourceJSONPointer(pointer, "examples"), form)
	}
	return nil
}

func sourceIndexSwaggerHeaders(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	return sourceIndexEntries(index, value, pointer, "Swagger response headers", false, func(child any, childPointer string) error {
		header, err := sourceReferenceObject(child, "Swagger response header")
		if err != nil {
			return err
		}
		if items, exists := header["items"]; exists {
			return sourceIndexSchema(index, items, sourceJSONPointer(childPointer, "items"), form)
		}
		return nil
	})
}

func sourceIndexContent(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm, resolveEncoding bool) error {
	content, err := sourceReferenceObject(value, "content")
	if err != nil {
		return err
	}
	if err := index.preflightEntries(pointer, content, false); err != nil {
		return err
	}
	for _, mediaType := range sortedSourceMapKeys(content) {
		mediaPointer := sourceJSONPointer(pointer, mediaType)
		media, err := sourceReferenceObject(content[mediaType], "content media")
		if err != nil {
			return err
		}
		if schema, exists := media["schema"]; exists {
			if err := sourceIndexSchema(index, schema, sourceJSONPointer(mediaPointer, "schema"), form); err != nil {
				return err
			}
		}
		if examples, exists := media["examples"]; exists {
			if err := sourceIndexExamples(index, examples, sourceJSONPointer(mediaPointer, "examples"), form); err != nil {
				return err
			}
		}
		if encoding, exists := media["encoding"]; exists {
			if !resolveEncoding {
				return fmt.Errorf("unsupported encoding on non-request content media %q", mediaType)
			}
			if err := sourceIndexEncoding(index, encoding, sourceJSONPointer(mediaPointer, "encoding"), form); err != nil {
				return err
			}
		}
	}
	return nil
}

func sourceIndexEncoding(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	return sourceIndexEntries(index, value, pointer, "encoding", false, func(child any, childPointer string) error {
		encoding, err := sourceReferenceObject(child, "encoding entry")
		if err != nil {
			return err
		}
		if headers, exists := encoding["headers"]; exists {
			return sourceIndexEntries(index, headers, sourceJSONPointer(childPointer, "headers"), "encoding headers", false, func(header any, headerPointer string) error {
				return sourceIndexHeader(index, header, headerPointer, form)
			})
		}
		return nil
	})
}

func sourceIndexLinks(index *sourceReferenceIndex, value any, pointer string) error {
	return sourceIndexEntries(index, value, pointer, "links", false, func(child any, childPointer string) error {
		return sourceIndexLink(index, child, childPointer)
	})
}

func sourceIndexLink(index *sourceReferenceIndex, value any, pointer string) error {
	if err := index.add(pointer, sourceReferenceLink); err != nil {
		return err
	}
	_, err := sourceReferenceObject(value, "link")
	return err
}

func sourceIndexExamples(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	return sourceIndexEntries(index, value, pointer, "examples", false, func(child any, childPointer string) error {
		return sourceIndexExample(index, child, childPointer, form)
	})
}

func sourceIndexExample(index *sourceReferenceIndex, value any, pointer string, _ sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceExample); err != nil {
		return err
	}
	_, err := sourceReferenceObject(value, "example")
	return err
}

func sourceIndexCallback(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceCallback); err != nil {
		return err
	}
	callback, err := sourceReferenceObject(value, "callback")
	if err != nil {
		return err
	}
	if _, hasReference := callback["$ref"]; hasReference {
		return nil
	}
	if err := index.preflightEntries(pointer, callback, true); err != nil {
		return err
	}
	for _, expression := range sortedSourceMapKeys(callback) {
		if strings.HasPrefix(expression, "x-") {
			continue
		}
		if err := sourceIndexPathItem(index, callback[expression], sourceJSONPointer(pointer, expression), form); err != nil {
			return err
		}
	}
	return nil
}

func sourceIndexSecurityScheme(index *sourceReferenceIndex, value any, pointer string) error {
	if err := index.add(pointer, sourceReferenceSecurity); err != nil {
		return err
	}
	_, err := sourceReferenceObject(value, "security scheme")
	return err
}

func sourceIndexSchema(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if _, err := sourceSchemaStructuralBytes(value, 0, index.limits); err != nil {
		return err
	}
	if err := index.add(pointer, sourceReferenceSchema); err != nil {
		return err
	}
	if _, isBoolean := value.(bool); isBoolean {
		if !form.isOpenAPI31() {
			return fmt.Errorf("unsupported %s boolean schema", form.Family)
		}
		return nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("schema at %s must be an object or boolean", pointer)
	}
	grammar := sourceSchemaGrammarFor(form)
	for _, key := range grammar.mapChildren {
		raw, exists := object[key]
		if !exists {
			continue
		}
		children, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("schema %s must be an object", key)
		}
		if err := index.preflightEntries(sourceJSONPointer(pointer, key), children, false); err != nil {
			return err
		}
		for _, name := range sortedSourceMapKeys(children) {
			if err := sourceIndexSchema(index, children[name], sourceJSONPointer(sourceJSONPointer(pointer, key), name), form); err != nil {
				return err
			}
		}
	}
	if grammar.dependencyChildren {
		if raw, exists := object["dependencies"]; exists {
			children, ok := raw.(map[string]any)
			if !ok {
				return fmt.Errorf("schema dependencies must be an object")
			}
			if err := index.preflightEntries(sourceJSONPointer(pointer, "dependencies"), children, false); err != nil {
				return err
			}
			for _, name := range sortedSourceMapKeys(children) {
				switch child := children[name].(type) {
				case map[string]any, bool:
					if err := sourceIndexSchema(index, child, sourceJSONPointer(sourceJSONPointer(pointer, "dependencies"), name), form); err != nil {
						return err
					}
				case []any:
				default:
					return fmt.Errorf("schema dependencies[%q] must be a schema or string array", name)
				}
			}
		}
	}
	for _, key := range grammar.schemaChildren {
		raw, exists := object[key]
		if !exists {
			continue
		}
		if key == "additionalProperties" && !form.isOpenAPI31() {
			if _, isBoolean := raw.(bool); isBoolean {
				continue
			}
		}
		if err := sourceIndexSchema(index, raw, sourceJSONPointer(pointer, key), form); err != nil {
			return err
		}
	}
	for _, key := range grammar.arrayChildren {
		raw, exists := object[key]
		if !exists {
			continue
		}
		items, ok := raw.([]any)
		if !ok {
			return fmt.Errorf("schema %s must be an array", key)
		}
		if err := index.preflightArrayEntries(sourceJSONPointer(pointer, key), len(items)); err != nil {
			return err
		}
		for position, item := range items {
			if err := sourceIndexSchema(index, item, sourceJSONPointer(sourceJSONPointer(pointer, key), strconv.Itoa(position)), form); err != nil {
				return err
			}
		}
	}
	return nil
}

func sourceReferenceObject(value any, label string) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an object", label)
	}
	return object, nil
}

func sourcePrepareSourceDocument(doc map[string]any, form sourceDocumentForm, limits sourceImportLimits, resolver *sourceReferenceResolver) error {
	if resolver == nil {
		return fmt.Errorf("source importer has no reference resolver")
	}
	index, err := sourceBuildReferenceIndex(doc, form, limits)
	if err != nil {
		return fmt.Errorf("index source grammar positions: %w", err)
	}
	preflight := sourceReferenceResolver{
		root:              doc,
		limits:            limits,
		form:              form,
		referenceIndex:    index,
		responseExpansion: sourceResponseExpansionBudget{limit: sourceResolvedDescriptorLimit(limits)},
	}
	if err := preflight.preflightDocument(); err != nil {
		return fmt.Errorf("preflight source grammar: %w", err)
	}
	resolver.root = doc
	resolver.limits = limits
	resolver.form = form
	resolver.referenceIndex = index
	resolver.references = 0
	resolver.expansion = sourceSchemaExpansionBudget{}
	resolver.responseExpansion = sourceResponseExpansionBudget{limit: sourceResolvedDescriptorLimit(limits)}
	resolver.responseScope = nil
	return nil
}

func (r *sourceReferenceResolver) preflightDocument() error {
	if rawPaths, declared := r.root["paths"]; declared {
		if err := r.preflightPathItems(rawPaths); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if rawWebhooks, declared := r.root["webhooks"]; declared && r.form.isOpenAPI31() {
			if err := r.preflightPathItems(rawWebhooks); err != nil {
				return err
			}
		}
		if rawComponents, declared := r.root["components"]; declared {
			if err := r.preflightOpenAPIComponents(rawComponents); err != nil {
				return err
			}
		}
		return nil
	}
	return r.preflightSwaggerComponents(r.root)
}

func (r *sourceReferenceResolver) preflightPathItems(value any) error {
	items, err := sourceReferenceObject(value, "path items")
	if err != nil {
		return err
	}
	for _, name := range sortedSourceMapKeys(items) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		if err := r.preflightPathItem(items[name]); err != nil {
			return fmt.Errorf("path item %q: %w", name, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightPathItem(value any) error {
	pathItem, err := r.resolvePathItem(value, nil, 0)
	if err != nil {
		return err
	}
	if parameters, declared := pathItem["parameters"]; declared {
		if err := r.preflightParameters(parameters); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if callbacks, declared := pathItem["callbacks"]; declared {
			if err := r.preflightCallbacks(callbacks); err != nil {
				return err
			}
		}
	}
	for _, method := range sourceHTTPMethods {
		operation, declared := pathItem[method]
		if !declared {
			continue
		}
		if err := r.preflightOperation(operation); err != nil {
			return fmt.Errorf("%s operation: %w", method, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightOperation(value any) error {
	operation, err := sourceReferenceObject(value, "operation")
	if err != nil {
		return err
	}
	if parameters, declared := operation["parameters"]; declared {
		if err := r.preflightParameters(parameters); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if requestBody, declared := operation["requestBody"]; declared {
			if err := r.preflightRequestBody(requestBody); err != nil {
				return err
			}
		}
	}
	if responses, declared := operation["responses"]; declared {
		if err := r.preflightResponses(responses); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if callbacks, declared := operation["callbacks"]; declared {
			if err := r.preflightCallbacks(callbacks); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightParameters(value any) error {
	parameters, ok := value.([]any)
	if !ok {
		return fmt.Errorf("parameters must be an array")
	}
	for index, parameter := range parameters {
		if err := r.preflightParameter(parameter); err != nil {
			return fmt.Errorf("parameter %d: %w", index, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightParameter(value any) error {
	parameter, err := r.resolveParameter(value, nil, 0)
	if err != nil {
		return err
	}
	schema, content, err := sourceParameterRepresentation(parameter, r.form)
	if err != nil {
		return err
	}
	if content != nil {
		return validateBoundedParameterContent("source preflight", content, r.form, r.limits)
	}
	if r.form.isSwagger2() {
		if _, declared := parameter["schema"]; !declared {
			schema, err = r.resolveSchema(schema, nil, 0)
			if err != nil {
				return err
			}
		}
	}
	return validateBoundedRequestSchema(schema, r.form, r.limits, 0)
}

func (r *sourceReferenceResolver) preflightRequestBody(value any) error {
	body, err := r.resolveRequestBody(value, nil, 0)
	if err != nil {
		return err
	}
	content, declared := body["content"]
	if !declared {
		return nil
	}
	media, err := sourceReferenceObject(content, "request body content")
	if err != nil {
		return err
	}
	for _, mediaType := range sortedSourceMapKeys(media) {
		declaration, err := sourceReferenceObject(media[mediaType], "request body media")
		if err != nil {
			return err
		}
		schema, declared := declaration["schema"]
		if !declared {
			continue
		}
		if err := validateBoundedRequestSchema(schema, r.form, r.limits, 0); err != nil {
			return fmt.Errorf("request body media %q: %w", mediaType, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightResponses(value any) error {
	responses, err := sourceReferenceObject(value, "responses")
	if err != nil {
		return err
	}
	for _, status := range sortedSourceMapKeys(responses) {
		if strings.HasPrefix(status, "x-") {
			continue
		}
		if _, err := r.resolveResponse(responses[status], nil, 0); err != nil {
			return fmt.Errorf("response %q: %w", status, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightCallbacks(value any) error {
	callbacks, err := sourceReferenceObject(value, "callbacks")
	if err != nil {
		return err
	}
	for _, name := range sortedSourceMapKeys(callbacks) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		if err := r.preflightCallback(callbacks[name]); err != nil {
			return fmt.Errorf("callback %q: %w", name, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightCallback(value any) error {
	callback, err := r.resolveCallback(value, nil, 0)
	if err != nil {
		return err
	}
	for _, expression := range sortedSourceMapKeys(callback) {
		if strings.HasPrefix(expression, "x-") {
			continue
		}
		if err := r.preflightPathItem(callback[expression]); err != nil {
			return fmt.Errorf("callback expression %q: %w", expression, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightOpenAPIComponents(value any) error {
	components, err := sourceReferenceObject(value, "components")
	if err != nil {
		return err
	}
	entries := []struct {
		name      string
		preflight func(any) error
	}{
		{name: "schemas", preflight: func(child any) error { _, err := r.resolveSchema(child, nil, 0); return err }},
		{name: "parameters", preflight: r.preflightParameter},
		{name: "responses", preflight: func(child any) error { _, err := r.resolveResponse(child, nil, 0); return err }},
		{name: "requestBodies", preflight: r.preflightRequestBody},
		{name: "headers", preflight: func(child any) error { _, err := r.resolveHeader(child, nil, 0); return err }},
		{name: "securitySchemes", preflight: func(child any) error { _, err := r.resolveSecurityScheme(child, nil, 0); return err }},
		{name: "links", preflight: func(child any) error { _, err := r.resolveLink(child, nil, 0); return err }},
		{name: "examples", preflight: func(child any) error { _, err := r.resolveExample(child, nil, 0); return err }},
		{name: "callbacks", preflight: r.preflightCallback},
	}
	if r.form.isOpenAPI31() {
		entries = append(entries, struct {
			name      string
			preflight func(any) error
		}{name: "pathItems", preflight: r.preflightPathItem})
	}
	for _, entry := range entries {
		raw, declared := components[entry.name]
		if !declared {
			continue
		}
		items, err := sourceReferenceObject(raw, "components."+entry.name)
		if err != nil {
			return err
		}
		for _, name := range sortedSourceMapKeys(items) {
			if err := entry.preflight(items[name]); err != nil {
				return fmt.Errorf("components.%s[%q]: %w", entry.name, name, err)
			}
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightSwaggerComponents(root map[string]any) error {
	entries := []struct {
		name      string
		preflight func(any) error
	}{
		{name: "definitions", preflight: func(child any) error { _, err := r.resolveSchema(child, nil, 0); return err }},
		{name: "parameters", preflight: r.preflightParameter},
		{name: "responses", preflight: func(child any) error { _, err := r.resolveResponse(child, nil, 0); return err }},
		{name: "securityDefinitions", preflight: func(child any) error { _, err := r.resolveSecurityScheme(child, nil, 0); return err }},
	}
	for _, entry := range entries {
		raw, declared := root[entry.name]
		if !declared {
			continue
		}
		items, err := sourceReferenceObject(raw, entry.name)
		if err != nil {
			return err
		}
		for _, name := range sortedSourceMapKeys(items) {
			if err := entry.preflight(items[name]); err != nil {
				return fmt.Errorf("%s[%q]: %w", entry.name, name, err)
			}
		}
	}
	return nil
}

func (r *sourceReferenceResolver) resolve(value any) (any, error) {
	return sourceCloneLiteral(value), nil
}

func (r *sourceReferenceResolver) resolvePathItem(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("path item must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferencePathItem, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolvePathItem(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		return sourceMergeReferenceObject(resolved, reference), nil
	}
	return sourceCloneMap(object), nil
}

func (r *sourceReferenceResolver) resolveInboundPathItem(value any, stack map[string]bool, depth int) (map[string]any, error) {
	pathItem, err := r.resolvePathItem(value, stack, depth)
	if err != nil {
		return nil, err
	}
	out := sourceCloneMap(pathItem)
	if parameters, exists := pathItem["parameters"]; exists {
		items, ok := parameters.([]any)
		if !ok {
			return nil, fmt.Errorf("path item parameters must be an array")
		}
		resolvedItems := make([]any, len(items))
		for index, item := range items {
			resolved, err := r.resolveParameter(item, nil, 0)
			if err != nil {
				return nil, err
			}
			resolvedItems[index] = resolved
		}
		out["parameters"] = resolvedItems
	}
	for _, method := range sourceHTTPMethods {
		rawOperation, exists := pathItem[method]
		if !exists {
			continue
		}
		operation, ok := rawOperation.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("inbound %s operation must be an object", method)
		}
		resolvedOperation := sourceCloneMap(operation)
		if parameters, exists := operation["parameters"]; exists {
			items, ok := parameters.([]any)
			if !ok {
				return nil, fmt.Errorf("inbound %s parameters must be an array", method)
			}
			resolvedItems := make([]any, len(items))
			for index, item := range items {
				resolved, err := r.resolveParameter(item, nil, 0)
				if err != nil {
					return nil, err
				}
				resolvedItems[index] = resolved
			}
			resolvedOperation["parameters"] = resolvedItems
		}
		if r.form.isOpenAPI() {
			if requestBody, exists := operation["requestBody"]; exists {
				resolved, err := r.resolveRequestBody(requestBody, nil, 0)
				if err != nil {
					return nil, err
				}
				resolvedOperation["requestBody"] = resolved
			}
		}
		if responses, exists := operation["responses"]; exists {
			responseMap, ok := responses.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("inbound %s responses must be an object", method)
			}
			resolvedResponses := make(map[string]any, len(responseMap))
			for _, status := range sortedSourceMapKeys(responseMap) {
				if strings.HasPrefix(status, "x-") {
					resolvedResponses[status] = sourceCloneLiteral(responseMap[status])
					continue
				}
				resolved, err := r.resolveResponse(responseMap[status], nil, 0)
				if err != nil {
					return nil, err
				}
				resolvedResponses[status] = resolved
			}
			resolvedOperation["responses"] = resolvedResponses
		}
		if r.form.isOpenAPI() {
			if callbacks, exists := operation["callbacks"]; exists {
				callbackMap, ok := callbacks.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("inbound %s callbacks must be an object", method)
				}
				resolvedCallbacks := make(map[string]any, len(callbackMap))
				for _, name := range sortedSourceMapKeys(callbackMap) {
					resolved, err := r.resolveCallback(callbackMap[name], nil, 0)
					if err != nil {
						return nil, err
					}
					resolvedCallbacks[name] = resolved
				}
				resolvedOperation["callbacks"] = resolvedCallbacks
			}
		}
		out[method] = resolvedOperation
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveParameter(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parameter must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceParameter, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveParameter(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		object = sourceMergeReferenceObject(resolved, reference)
	}
	out := sourceCloneMap(object)
	if schema, exists := object["schema"]; exists {
		resolved, err := r.resolveSchema(schema, nil, 0)
		if err != nil {
			return nil, err
		}
		out["schema"] = resolved
	}
	if r.form.isOpenAPI() {
		if content, exists := object["content"]; exists {
			resolved, err := r.resolveContent(content, stack, depth, false)
			if err != nil {
				return nil, err
			}
			out["content"] = resolved
		}
		if examples, exists := object["examples"]; exists {
			resolved, err := r.resolveExamples(examples, stack, depth)
			if err != nil {
				return nil, err
			}
			out["examples"] = resolved
		}
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveRequestBody(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("request body must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceRequestBody, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveRequestBody(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		object = sourceMergeReferenceObject(resolved, reference)
	}
	out := sourceCloneMap(object)
	if content, exists := object["content"]; exists {
		resolved, err := r.resolveContent(content, stack, depth, true)
		if err != nil {
			return nil, err
		}
		out["content"] = resolved
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveResponse(value any, stack map[string]bool, depth int) (map[string]any, error) {
	if err := r.reserveResponseEntry(value, stack, depth); err != nil {
		return nil, err
	}
	return r.resolveResponseUnchecked(value, stack, depth)
}

func (r *sourceReferenceResolver) reserveResponseEntry(value any, stack map[string]bool, depth int) error {
	bytes, err := r.responseExpansionBytes(value, stack, depth)
	if err != nil {
		return err
	}
	if err := r.responseExpansion.check(bytes); err != nil {
		return err
	}
	if r.responseScope != nil {
		if err := r.responseScope.check(bytes); err != nil {
			return err
		}
	}
	if err := r.responseExpansion.reserve(bytes); err != nil {
		return err
	}
	if r.responseScope != nil {
		r.responseScope.used += bytes
	}
	return nil
}

func (r *sourceReferenceResolver) responseExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("response must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceResponse, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if !hasReference {
		return sourceResponseStructuralBytes(object, r.limits)
	}
	targetBytes, err := r.responseExpansionBytes(target, next, depth+1)
	if err != nil {
		return 0, err
	}
	referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
	if err != nil {
		return 0, err
	}
	return sourceStructuralAdd(targetBytes, referenceBytes)
}

func (r *sourceReferenceResolver) resolveResponseUnchecked(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("response must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceResponse, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveResponseUnchecked(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		object = sourceMergeReferenceObject(resolved, reference)
	}
	out := sourceCloneMap(object)
	if r.form.isSwagger2() {
		if schema, exists := object["schema"]; exists {
			resolved, err := r.resolveSchema(schema, nil, 0)
			if err != nil {
				return nil, err
			}
			out["schema"] = resolved
		}
		if headers, exists := object["headers"]; exists {
			resolved, err := r.resolveSwaggerHeaders(headers, stack, depth)
			if err != nil {
				return nil, err
			}
			out["headers"] = resolved
		}
	}
	if r.form.isOpenAPI() {
		if content, exists := object["content"]; exists {
			resolved, err := r.resolveContent(content, stack, depth, false)
			if err != nil {
				return nil, err
			}
			out["content"] = resolved
		}
	}
	if r.form.isOpenAPI() {
		if headers, exists := object["headers"]; exists {
			headerMap, ok := headers.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("response headers must be an object")
			}
			resolvedHeaders := make(map[string]any, len(headerMap))
			for _, name := range sortedSourceMapKeys(headerMap) {
				resolved, err := r.resolveHeader(headerMap[name], nil, 0)
				if err != nil {
					return nil, fmt.Errorf("header %q: %w", name, err)
				}
				resolvedHeaders[name] = resolved
			}
			out["headers"] = resolvedHeaders
		}
	}
	if r.form.isOpenAPI() {
		if links, exists := object["links"]; exists {
			resolved, err := r.resolveLinks(links, stack, depth)
			if err != nil {
				return nil, err
			}
			out["links"] = resolved
		}
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveSwaggerHeaders(value any, stack map[string]bool, depth int) (map[string]any, error) {
	headers, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("Swagger response headers must be an object")
	}
	out := make(map[string]any, len(headers))
	for _, name := range sortedSourceMapKeys(headers) {
		header, ok := headers[name].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("Swagger response header %q must be an object", name)
		}
		resolvedHeader := sourceCloneMap(header)
		if items, exists := header["items"]; exists {
			resolved, err := r.resolveSchema(items, stack, depth)
			if err != nil {
				return nil, fmt.Errorf("Swagger response header %q items: %w", name, err)
			}
			resolvedHeader["items"] = resolved
		}
		out[name] = resolvedHeader
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveHeader(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("header must be an object")
	}
	if !r.form.isOpenAPI() {
		return sourceCloneMap(object), nil
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceHeader, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveHeader(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		object = sourceMergeReferenceObject(resolved, reference)
	}
	out := sourceCloneMap(object)
	if r.form.isOpenAPI() {
		if schema, exists := object["schema"]; exists {
			resolved, err := r.resolveSchema(schema, nil, 0)
			if err != nil {
				return nil, err
			}
			out["schema"] = resolved
		}
		if content, exists := object["content"]; exists {
			resolved, err := r.resolveContent(content, stack, depth, false)
			if err != nil {
				return nil, err
			}
			out["content"] = resolved
		}
		if examples, exists := object["examples"]; exists {
			resolved, err := r.resolveExamples(examples, stack, depth)
			if err != nil {
				return nil, err
			}
			out["examples"] = resolved
		}
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveContent(value any, stack map[string]bool, depth int, resolveEncoding bool) (any, error) {
	content, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("content must be an object")
	}
	out := make(map[string]any, len(content))
	for _, mediaType := range sortedSourceMapKeys(content) {
		media, ok := content[mediaType].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("content media %q must be an object", mediaType)
		}
		if _, exists := media["encoding"]; exists && !resolveEncoding {
			return nil, fmt.Errorf("unsupported encoding on non-request content media %q", mediaType)
		}
		resolvedMedia := sourceCloneMap(media)
		if schema, exists := media["schema"]; exists {
			resolved, err := r.resolveSchema(schema, stack, depth)
			if err != nil {
				return nil, fmt.Errorf("content media %q schema: %w", mediaType, err)
			}
			resolvedMedia["schema"] = resolved
		}
		if examples, exists := media["examples"]; exists {
			resolved, err := r.resolveExamples(examples, stack, depth)
			if err != nil {
				return nil, fmt.Errorf("content media %q examples: %w", mediaType, err)
			}
			resolvedMedia["examples"] = resolved
		}
		if encoding, exists := media["encoding"]; exists {
			resolved, err := r.resolveEncoding(encoding, stack, depth)
			if err != nil {
				return nil, fmt.Errorf("content media %q encoding: %w", mediaType, err)
			}
			resolvedMedia["encoding"] = resolved
		}
		out[mediaType] = resolvedMedia
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveEncoding(value any, stack map[string]bool, depth int) (map[string]any, error) {
	encodings, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("encoding must be an object")
	}
	out := make(map[string]any, len(encodings))
	for _, property := range sortedSourceMapKeys(encodings) {
		encoding, ok := encodings[property].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("encoding %q must be an object", property)
		}
		resolvedEncoding := sourceCloneMap(encoding)
		if headers, exists := encoding["headers"]; exists {
			headerMap, ok := headers.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("encoding %q headers must be an object", property)
			}
			resolvedHeaders := make(map[string]any, len(headerMap))
			for _, name := range sortedSourceMapKeys(headerMap) {
				resolved, err := r.resolveHeader(headerMap[name], stack, depth)
				if err != nil {
					return nil, fmt.Errorf("encoding %q header %q: %w", property, name, err)
				}
				resolvedHeaders[name] = resolved
			}
			resolvedEncoding["headers"] = resolvedHeaders
		}
		out[property] = resolvedEncoding
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveLinks(value any, stack map[string]bool, depth int) (map[string]any, error) {
	links, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("response links must be an object")
	}
	out := make(map[string]any, len(links))
	for _, name := range sortedSourceMapKeys(links) {
		resolved, err := r.resolveLink(links[name], stack, depth)
		if err != nil {
			return nil, fmt.Errorf("link %q: %w", name, err)
		}
		out[name] = resolved
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveLink(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("link must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceLink, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveLink(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		return sourceMergeReferenceObject(resolved, reference), nil
	}
	return sourceCloneMap(object), nil
}

func (r *sourceReferenceResolver) resolveExamples(value any, stack map[string]bool, depth int) (map[string]any, error) {
	examples, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("examples must be an object")
	}
	out := make(map[string]any, len(examples))
	for _, name := range sortedSourceMapKeys(examples) {
		resolved, err := r.resolveExample(examples[name], stack, depth)
		if err != nil {
			return nil, fmt.Errorf("example %q: %w", name, err)
		}
		out[name] = resolved
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveExample(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("example must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceExample, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveExample(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		return sourceMergeReferenceObject(resolved, reference), nil
	}
	return sourceCloneMap(object), nil
}

func (r *sourceReferenceResolver) resolveSecurityScheme(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("security scheme must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceSecurity, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveSecurityScheme(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		return sourceMergeReferenceObject(resolved, reference), nil
	}
	return sourceCloneMap(object), nil
}

func (r *sourceReferenceResolver) resolveSchema(value any, stack map[string]bool, depth int) (any, error) {
	if err := r.reserveSchemaEntry(value, "schema"); err != nil {
		return nil, err
	}
	if booleanSchema, ok := value.(bool); ok {
		if !r.form.isOpenAPI31() {
			return nil, fmt.Errorf("unsupported %s boolean schema", r.form.Family)
		}
		return booleanSchema, nil
	}
	object, ok := value.(map[string]any)
	if !ok {
		return sourceCloneLiteral(value), nil
	}
	if err := sourceRejectDynamicSchemaKeywords(object); err != nil {
		return nil, err
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceSchema, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveSchema(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		targetObject, ok := resolved.(map[string]any)
		if !ok {
			if len(reference) == 1 {
				return resolved, nil
			}
			return nil, fmt.Errorf("schema reference with siblings does not resolve to an object")
		}
		object = sourceOverlayReferenceObject(targetObject, reference)
	}
	if err := sourceValidateSchemaForm(object, r.form); err != nil {
		return nil, err
	}
	grammar := sourceSchemaGrammarFor(r.form)
	for _, key := range sourceSchemaBearingKeywords {
		if _, exists := object[key]; exists && !grammar.handles(key) {
			return nil, fmt.Errorf("unsupported %s schema keyword %s", r.form.Family, key)
		}
	}
	var objectUsed int64
	out := make(map[string]any, len(object))
	special := map[string]bool{}
	for _, key := range grammar.mapChildren {
		special[key] = true
	}
	for _, key := range grammar.schemaChildren {
		special[key] = true
	}
	for _, key := range grammar.arrayChildren {
		special[key] = true
	}
	if grammar.dependencyChildren {
		special["dependencies"] = true
	}
	for _, key := range sortedSourceMapKeys(object) {
		if special[key] {
			continue
		}
		if err := r.reserveSchemaValue(object[key], &objectUsed, "schema field "+key); err != nil {
			return nil, err
		}
		out[key] = sourceCloneLiteral(object[key])
	}
	for _, key := range grammar.mapChildren {
		if raw, exists := object[key]; exists {
			children, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("schema %s must be an object", key)
			}
			resolvedChildren := make(map[string]any, len(children))
			for _, name := range sortedSourceMapKeys(children) {
				resolved, err := r.resolveSchemaChild(children[name], stack, depth+1, fmt.Sprintf("schema %s[%q]", key, name))
				if err != nil {
					return nil, err
				}
				if err := r.reserveSchemaValue(resolved, &objectUsed, fmt.Sprintf("schema %s[%q]", key, name)); err != nil {
					return nil, err
				}
				resolvedChildren[name] = resolved
			}
			out[key] = resolvedChildren
		}
	}
	if grammar.dependencyChildren {
		if raw, exists := object["dependencies"]; exists {
			children, ok := raw.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("schema dependencies must be an object")
			}
			resolvedChildren := make(map[string]any, len(children))
			for _, name := range sortedSourceMapKeys(children) {
				child := children[name]
				switch child.(type) {
				case map[string]any, bool:
					resolved, err := r.resolveSchemaChild(child, stack, depth+1, fmt.Sprintf("schema dependencies[%q]", name))
					if err != nil {
						return nil, err
					}
					if err := r.reserveSchemaValue(resolved, &objectUsed, fmt.Sprintf("schema dependencies[%q]", name)); err != nil {
						return nil, err
					}
					resolvedChildren[name] = resolved
				case []any:
					if err := r.reserveSchemaValue(child, &objectUsed, fmt.Sprintf("schema dependencies[%q]", name)); err != nil {
						return nil, err
					}
					resolvedChildren[name] = sourceCloneLiteral(child)
				default:
					return nil, fmt.Errorf("schema dependencies[%q] must be a schema or string array", name)
				}
			}
			out["dependencies"] = resolvedChildren
		}
	}
	for _, key := range grammar.schemaChildren {
		if raw, exists := object[key]; exists {
			if key == "additionalProperties" && !r.form.isOpenAPI31() {
				if booleanValue, isBoolean := raw.(bool); isBoolean {
					if err := r.reserveSchemaValue(booleanValue, &objectUsed, "schema field "+key); err != nil {
						return nil, err
					}
					out[key] = booleanValue
					continue
				}
			}
			resolved, err := r.resolveSchemaChild(raw, stack, depth+1, "schema field "+key)
			if err != nil {
				return nil, err
			}
			if err := r.reserveSchemaValue(resolved, &objectUsed, "schema field "+key); err != nil {
				return nil, err
			}
			out[key] = resolved
		}
	}
	for _, key := range grammar.arrayChildren {
		if raw, exists := object[key]; exists {
			items, ok := raw.([]any)
			if !ok {
				return nil, fmt.Errorf("schema %s must be an array", key)
			}
			resolvedItems := make([]any, len(items))
			for index, item := range items {
				resolved, err := r.resolveSchemaChild(item, stack, depth+1, fmt.Sprintf("schema %s[%d]", key, index))
				if err != nil {
					return nil, err
				}
				if err := r.reserveSchemaValue(resolved, &objectUsed, fmt.Sprintf("schema %s[%d]", key, index)); err != nil {
					return nil, err
				}
				resolvedItems[index] = resolved
			}
			out[key] = resolvedItems
		}
	}
	return out, nil
}

func (r *sourceReferenceResolver) resolveSchemaChild(value any, stack map[string]bool, depth int, label string) (any, error) {
	switch value.(type) {
	case map[string]any, bool:
		return r.resolveSchema(value, stack, depth)
	default:
		return nil, fmt.Errorf("%s must be a schema object or boolean", label)
	}
}

func (r *sourceReferenceResolver) reserveSchemaEntry(value any, label string) error {
	bytes, err := sourceSchemaStructuralBytes(value, 0, r.limits)
	if err != nil {
		if limit, ok := err.(*sourceSchemaStructureLimitError); ok {
			return &sourceSchemaExpansionLimitError{Scope: limit.Scope, Label: label}
		}
		return err
	}
	if bytes > sourceSchemaByteLimit(r.limits) {
		return &sourceSchemaExpansionLimitError{Scope: "object", Label: label}
	}
	if r.expansion.nodes >= sourceSchemaNodeLimit(r.limits) {
		return &sourceSchemaExpansionLimitError{Scope: "node", Label: label}
	}
	documentLimit := sourceResolvedDescriptorLimit(r.limits)
	if bytes > documentLimit || r.expansion.used > documentLimit-bytes {
		return &sourceSchemaExpansionLimitError{Scope: "document", Label: label}
	}
	r.expansion.nodes++
	r.expansion.used += bytes
	return nil
}

func sourceSchemaByteLimit(limits sourceImportLimits) int64 {
	if limits.MaxSchemaBytes > 0 {
		return limits.MaxSchemaBytes
	}
	return defaultSourceImportSchemaBytes
}

func sourceSchemaNodeLimit(limits sourceImportLimits) int {
	if limits.MaxSchemaNodes > 0 {
		return limits.MaxSchemaNodes
	}
	return defaultSourceImportSchemaNodes
}

func (r *sourceReferenceResolver) reserveSchemaValue(value any, objectUsed *int64, label string) error {
	bytes, err := sourceSchemaStructuralBytes(value, 0, r.limits)
	if err != nil {
		if limit, ok := err.(*sourceSchemaStructureLimitError); ok {
			return &sourceSchemaExpansionLimitError{Scope: limit.Scope, Label: label}
		}
		return err
	}
	if bytes > sourceSchemaByteLimit(r.limits) || *objectUsed > sourceSchemaByteLimit(r.limits)-bytes {
		return &sourceSchemaExpansionLimitError{Scope: "object", Label: label}
	}
	*objectUsed += bytes
	return nil
}

func sourceSchemaStructuralBytes(value any, depth int, limits sourceImportLimits) (int64, error) {
	return sourceStructuralBytes(value, depth, limits.MaxReferenceDepth, "schema depth limit exceeded", limits)
}

func sourceResponseStructuralBytes(value any, limits sourceImportLimits) (int64, error) {
	depthLimit := limits.MaxReferenceDepth
	if depthLimit < defaultSourceImportReferenceDepth {
		depthLimit = defaultSourceImportReferenceDepth
	}
	return sourceStructuralBytes(value, 0, depthLimit, "response structural depth limit exceeded", limits)
}

func sourceStructuralBytes(value any, depth, depthLimit int, depthError string, limits sourceImportLimits) (int64, error) {
	nodes := 0
	var walk func(any, int) (int64, error)
	walk = func(current any, currentDepth int) (int64, error) {
		if currentDepth > depthLimit {
			return 0, fmt.Errorf("%s", depthError)
		}
		nodes++
		if nodes > sourceSchemaNodeLimit(limits) {
			return 0, &sourceSchemaStructureLimitError{Scope: "node"}
		}
		switch typed := current.(type) {
		case nil:
			return 4, nil
		case bool:
			if typed {
				return 4, nil
			}
			return 5, nil
		case string:
			return sourceStructuralStringBytes(typed), nil
		case json.Number:
			return int64(len(typed)), nil
		case float64:
			return int64(len(strconv.FormatFloat(typed, 'g', -1, 64))), nil
		case float32:
			return int64(len(strconv.FormatFloat(float64(typed), 'g', -1, 32))), nil
		case int:
			return int64(len(strconv.Itoa(typed))), nil
		case int64:
			return int64(len(strconv.FormatInt(typed, 10))), nil
		case uint:
			return int64(len(strconv.FormatUint(uint64(typed), 10))), nil
		case uint64:
			return int64(len(strconv.FormatUint(typed, 10))), nil
		case map[string]any:
			used := int64(2)
			for key, value := range typed {
				child, err := walk(value, currentDepth+1)
				if err != nil {
					return 0, err
				}
				used, err = sourceStructuralAdd(used, sourceStructuralStringBytes(key))
				if err != nil {
					return 0, err
				}
				used, err = sourceStructuralAdd(used, child)
				if err != nil {
					return 0, err
				}
				used, err = sourceStructuralAdd(used, 2)
				if err != nil {
					return 0, err
				}
			}
			return used, nil
		case []any:
			used := int64(2)
			for _, childValue := range typed {
				child, err := walk(childValue, currentDepth+1)
				if err != nil {
					return 0, err
				}
				used, err = sourceStructuralAdd(used, child)
				if err != nil {
					return 0, err
				}
				used, err = sourceStructuralAdd(used, 1)
				if err != nil {
					return 0, err
				}
			}
			return used, nil
		default:
			return 0, fmt.Errorf("schema contains unsupported value type %T", current)
		}
	}
	return walk(value, depth)
}

func sourceStructuralStringBytes(value string) int64 {
	if len(value) > math.MaxInt64/6-2 {
		return math.MaxInt64
	}
	return int64(len(value))*6 + 2
}

func sourceStructuralAdd(left, right int64) (int64, error) {
	if right < 0 || left > math.MaxInt64-right {
		return 0, &sourceSchemaStructureLimitError{Scope: "object"}
	}
	return left + right, nil
}

func (r *sourceReferenceResolver) resolveCallback(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("callback must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceCallback, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveCallback(target, next, depth+1)
		if err != nil {
			return nil, err
		}
		object = sourceMergeReferenceObject(resolved, reference)
	}
	out := sourceCloneMap(object)
	for _, expression := range sortedSourceMapKeys(object) {
		if strings.HasPrefix(expression, "x-") {
			continue
		}
		pathItem, err := r.resolveInboundPathItem(object[expression], nil, 0)
		if err != nil {
			return nil, fmt.Errorf("callback expression %q: %w", expression, err)
		}
		out[expression] = pathItem
	}
	return out, nil
}

func (r *sourceReferenceResolver) referenceTarget(object map[string]any, kind sourceReferenceKind, stack map[string]bool, depth int) (any, map[string]any, map[string]bool, bool, error) {
	return r.referenceTargetWithCount(object, kind, stack, depth, true)
}

func (r *sourceReferenceResolver) referenceTargetWithCount(object map[string]any, kind sourceReferenceKind, stack map[string]bool, depth int, countReference bool) (any, map[string]any, map[string]bool, bool, error) {
	rawRef, hasReference := object["$ref"]
	if !hasReference {
		return nil, object, stack, false, nil
	}
	if depth >= r.limits.MaxReferenceDepth {
		return nil, nil, nil, false, fmt.Errorf("reference depth limit exceeded")
	}
	for key := range object {
		if !r.referenceSiblingAllowed(kind, key) {
			return nil, nil, nil, false, fmt.Errorf("ambiguous %s reference with sibling field %q", kind, key)
		}
	}
	ref, ok := rawRef.(string)
	if !ok || (ref != "#" && !strings.HasPrefix(ref, "#/")) {
		return nil, nil, nil, false, fmt.Errorf("external reference %q is unsupported", rawRef)
	}
	if countReference {
		r.references++
		if r.references > r.limits.MaxReferences {
			return nil, nil, nil, false, fmt.Errorf("reference count limit exceeded")
		}
	}
	if stack != nil && stack[ref] {
		return nil, nil, nil, false, fmt.Errorf("reference cycle at %q", ref)
	}
	target, err := sourcePointer(r.root, ref)
	if err != nil {
		return nil, nil, nil, false, err
	}
	if err := r.validateReferenceTargetKind(target, kind, ref); err != nil {
		return nil, nil, nil, false, err
	}
	next := make(map[string]bool, len(stack)+1)
	for item := range stack {
		next[item] = true
	}
	next[ref] = true
	return target, object, next, true, nil
}

func (r *sourceReferenceResolver) referenceSiblingAllowed(kind sourceReferenceKind, key string) bool {
	if key == "$ref" || strings.HasPrefix(key, "x-") {
		return true
	}
	if !r.form.allowsReferenceSiblings() {
		return false
	}
	if kind == sourceReferenceSchema {
		return true
	}
	return key == "summary" || key == "description"
}

func (r *sourceReferenceResolver) validateReferenceTargetKind(target any, kind sourceReferenceKind, ref string) error {
	if r.referenceIndex == nil {
		return fmt.Errorf("reference grammar index is unavailable")
	}
	actual, indexed := r.referenceIndex.positions[ref]
	if !indexed {
		return fmt.Errorf("%s reference %q does not resolve to a grammar-defined %s position", kind, ref, kind)
	}
	if actual != kind {
		return fmt.Errorf("%s reference %q resolves to %s rather than the expected kind", kind, ref, actual)
	}
	if kind == sourceReferenceSchema {
		if _, isObject := target.(map[string]any); isObject {
			return nil
		}
		if _, isBooleanSchema := target.(bool); isBooleanSchema {
			return nil
		}
		return fmt.Errorf("%s reference does not resolve to a schema", kind)
	}
	object, ok := target.(map[string]any)
	if !ok {
		return fmt.Errorf("%s reference does not resolve to an object", kind)
	}
	if _, chained := object["$ref"]; chained {
		return nil
	}
	switch kind {
	case sourceReferenceParameter:
		if _, ok := object["name"].(string); !ok {
			return fmt.Errorf("parameter reference does not resolve to a parameter object")
		}
		if _, ok := object["in"].(string); !ok {
			return fmt.Errorf("parameter reference does not resolve to a parameter object")
		}
	case sourceReferenceRequestBody:
		if _, ok := object["content"].(map[string]any); !ok {
			return fmt.Errorf("request body reference does not resolve to a request body object")
		}
	case sourceReferenceResponse:
		if _, ok := object["description"].(string); !ok {
			return fmt.Errorf("response reference does not resolve to a response object")
		}
	case sourceReferenceLink:
		operationRef, hasOperationRef := object["operationRef"].(string)
		operationID, hasOperationID := object["operationId"].(string)
		if (hasOperationRef && operationRef == "") || (hasOperationID && operationID == "") || hasOperationRef == hasOperationID {
			return fmt.Errorf("link reference does not resolve to a link object")
		}
	case sourceReferenceExample:
		if err := sourceValidateReferenceObjectFields(object, map[string]bool{"summary": true, "description": true, "value": true, "externalValue": true}); err != nil {
			return fmt.Errorf("example reference does not resolve to an example object: %w", err)
		}
	case sourceReferencePathItem:
		allowed := map[string]bool{"$ref": true, "summary": true, "description": true, "servers": true, "parameters": true}
		for _, method := range sourceHTTPMethods {
			allowed[method] = true
		}
		if err := sourceValidateReferenceObjectFields(object, allowed); err != nil {
			return fmt.Errorf("path item reference does not resolve to a path item object: %w", err)
		}
	case sourceReferenceHeader:
		allowed := map[string]bool{"description": true, "required": true, "deprecated": true, "allowEmptyValue": true, "style": true, "explode": true, "allowReserved": true, "schema": true, "example": true, "examples": true, "content": true, "type": true, "format": true, "items": true, "collectionFormat": true, "default": true, "maximum": true, "exclusiveMaximum": true, "minimum": true, "exclusiveMinimum": true, "maxLength": true, "minLength": true, "pattern": true, "maxItems": true, "minItems": true, "uniqueItems": true, "maxProperties": true, "minProperties": true, "enum": true, "multipleOf": true}
		if err := sourceValidateReferenceObjectFields(object, allowed); err != nil {
			return fmt.Errorf("header reference does not resolve to a header object: %w", err)
		}
	case sourceReferenceCallback:
		for key, value := range object {
			if strings.HasPrefix(key, "x-") {
				continue
			}
			if _, ok := value.(map[string]any); !ok {
				return fmt.Errorf("callback reference does not resolve to a callback object")
			}
		}
	case sourceReferenceSecurity:
		if _, ok := object["type"].(string); !ok {
			return fmt.Errorf("security scheme reference does not resolve to a security scheme object")
		}
	}
	return nil
}

func (r *sourceReferenceResolver) referencePointerCategory(ref string) string {
	segments := strings.Split(strings.TrimPrefix(ref, "#/"), "/")
	if r.form.isOpenAPI() && len(segments) == 3 && segments[0] == "components" {
		return segments[1]
	}
	if r.form.isSwagger2() && len(segments) == 2 {
		switch segments[0] {
		case "definitions", "parameters", "responses":
			return segments[0]
		}
	}
	return ""
}

func sourceReferenceCategoryMatches(kind sourceReferenceKind, category string) bool {
	return map[sourceReferenceKind]string{
		sourceReferencePathItem:    "pathItems",
		sourceReferenceParameter:   "parameters",
		sourceReferenceRequestBody: "requestBodies",
		sourceReferenceResponse:    "responses",
		sourceReferenceHeader:      "headers",
		sourceReferenceSchema:      "schemas",
		sourceReferenceCallback:    "callbacks",
		sourceReferenceLink:        "links",
		sourceReferenceExample:     "examples",
		sourceReferenceSecurity:    "securitySchemes",
	}[kind] == category || (kind == sourceReferenceSchema && category == "definitions")
}

func sourceValidateReferenceObjectFields(object map[string]any, allowed map[string]bool) error {
	for key := range object {
		if allowed[key] || strings.HasPrefix(key, "x-") {
			continue
		}
		return fmt.Errorf("unexpected field %q", key)
	}
	return nil
}

func sourceMergeReferenceObject(target, reference map[string]any) map[string]any {
	out := sourceCloneMap(target)
	for _, key := range sortedSourceMapKeys(reference) {
		if key != "$ref" {
			out[key] = sourceCloneLiteral(reference[key])
		}
	}
	return out
}

func sourceOverlayReferenceObject(target, reference map[string]any) map[string]any {
	out := make(map[string]any, len(target)+len(reference))
	for key, value := range target {
		out[key] = value
	}
	for key, value := range reference {
		if key != "$ref" {
			out[key] = value
		}
	}
	return out
}

func sourceCloneMap(value map[string]any) map[string]any {
	out := make(map[string]any, len(value))
	for _, key := range sortedSourceMapKeys(value) {
		out[key] = sourceCloneLiteral(value[key])
	}
	return out
}

func sourceCloneLiteral(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return sourceCloneMap(typed)
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = sourceCloneLiteral(item)
		}
		return out
	default:
		return value
	}
}

func sourcePointer(root map[string]any, ref string) (any, error) {
	if ref == "#" {
		return root, nil
	}
	var current any = root
	for _, segment := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		segment = strings.ReplaceAll(strings.ReplaceAll(segment, "~1", "/"), "~0", "~")
		switch typed := current.(type) {
		case map[string]any:
			var exists bool
			current, exists = typed[segment]
			if !exists {
				return nil, fmt.Errorf("unresolved reference %q", ref)
			}
		case []any:
			index, err := strconv.Atoi(segment)
			if err != nil || index < 0 || index >= len(typed) {
				return nil, fmt.Errorf("unresolved reference %q", ref)
			}
			current = typed[index]
		default:
			return nil, fmt.Errorf("reference %q does not resolve to an object member", ref)
		}
	}
	return current, nil
}

func importSourceDocument(lock sourceImportLock, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, limits sourceImportLimits) ([]sourceOperationDescriptor, error) {
	budget := sourceImportBudget{limit: sourceResolvedDescriptorLimit(limits)}
	result, err := importSourceDocumentResult(lock, doc, form, resolver, limits, &budget)
	if err != nil {
		return nil, err
	}
	return result.Operations, nil
}

func importSourceDocumentResult(lock sourceImportLock, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, limits sourceImportLimits, budget *sourceImportBudget) (sourceImportResult, error) {
	if budget == nil {
		return sourceImportResult{}, fmt.Errorf("source importer has no descriptor budget")
	}
	if err := sourcePrepareSourceDocument(doc, form, limits, resolver); err != nil {
		return sourceImportResult{}, err
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{}}
	rootServers := sourceServerLayer{}
	if form.isOpenAPI() {
		var err error
		rootServers, err = sourceServerLayerFrom(doc)
		if err != nil {
			return sourceImportResult{}, fmt.Errorf("root servers: %w", err)
		}
	}
	webhooks, err := sourceWebhookEvents(lock, doc, form, resolver)
	if err != nil {
		return sourceImportResult{}, err
	}
	for _, event := range webhooks {
		if len(result.InboundEvents) >= limits.MaxOperations {
			return sourceImportResult{}, fmt.Errorf("inbound event count limit exceeded")
		}
		if err := budget.add(event, "inbound event"); err != nil {
			return sourceImportResult{}, err
		}
		result.InboundEvents = append(result.InboundEvents, event)
	}
	rawPaths, declaredPaths := doc["paths"]
	if !declaredPaths {
		if len(result.InboundEvents) == 0 {
			return sourceImportResult{}, fmt.Errorf("source artifact has no paths object")
		}
		sortSourceInboundEventDescriptors(result.InboundEvents)
		if err := validateSourceImportResultIdentities(result); err != nil {
			return sourceImportResult{}, err
		}
		return result, nil
	}
	paths, ok := rawPaths.(map[string]any)
	if !ok {
		return sourceImportResult{}, fmt.Errorf("source artifact paths must be an object")
	}
	for _, path := range sortedSourceMapKeys(paths) {
		if strings.HasPrefix(path, "x-") {
			extension := sourceExtensionDescriptor{Location: fmt.Sprintf("paths[%q]", path), Value: sourceCloneLiteral(paths[path])}
			if err := budget.add(extension, "extension"); err != nil {
				return sourceImportResult{}, err
			}
			result.Extensions = append(result.Extensions, extension)
			continue
		}
		if err := validateSourceImportPath(path); err != nil {
			return sourceImportResult{}, fmt.Errorf("path %q is not a connector-relative path template", path)
		}
		pathItem, err := resolver.resolvePathItem(paths[path], nil, 0)
		if err != nil {
			return sourceImportResult{}, fmt.Errorf("path %q: %w", path, err)
		}
		pathServers := sourceServerLayer{}
		if form.isOpenAPI() {
			pathServers, err = sourceServerLayerFrom(pathItem)
			if err != nil {
				return sourceImportResult{}, fmt.Errorf("path %q servers: %w", path, err)
			}
		}
		for _, key := range sortedSourceMapKeys(pathItem) {
			if !strings.HasPrefix(key, "x-") {
				continue
			}
			extension := sourceExtensionDescriptor{Location: fmt.Sprintf("paths[%q].%s", path, key), Value: sourceCloneLiteral(pathItem[key])}
			if err := budget.add(extension, "extension"); err != nil {
				return sourceImportResult{}, err
			}
			result.Extensions = append(result.Extensions, extension)
		}
		if form.isOpenAPI() {
			if rawCallbacks, hasCallbacks := pathItem["callbacks"]; hasCallbacks {
				events, err := sourceCallbackEvents(lock, form, "", fmt.Sprintf("paths[%q].callbacks", path), rawCallbacks, resolver)
				if err != nil {
					return sourceImportResult{}, err
				}
				for _, event := range events {
					if len(result.InboundEvents) >= limits.MaxOperations {
						return sourceImportResult{}, fmt.Errorf("inbound event count limit exceeded")
					}
					if err := budget.add(event, "inbound event"); err != nil {
						return sourceImportResult{}, err
					}
					result.InboundEvents = append(result.InboundEvents, event)
				}
			}
		}
		pathParameters, err := sourceParameterValues(pathItem["parameters"], resolver, form)
		if err != nil {
			return sourceImportResult{}, fmt.Errorf("path %q parameters: %w", path, err)
		}
		for _, method := range sourceHTTPMethods {
			rawOperation, ok := pathItem[method]
			if !ok {
				continue
			}
			operation, ok := rawOperation.(map[string]any)
			if !ok {
				return sourceImportResult{}, fmt.Errorf("operation %s %s must be an object", method, path)
			}
			descriptor, err := importSourceOperation(lock, doc, form, resolver, path, method, pathParameters, rootServers, pathServers, operation, limits, budget.remaining())
			if err != nil {
				return sourceImportResult{}, err
			}
			if len(result.Operations) >= limits.MaxOperations {
				return sourceImportResult{}, fmt.Errorf("operation count limit exceeded")
			}
			if err := budget.add(descriptor, "operation"); err != nil {
				return sourceImportResult{}, err
			}
			result.Operations = append(result.Operations, descriptor)
			if form.isOpenAPI() {
				if rawCallbacks, hasCallbacks := operation["callbacks"]; hasCallbacks {
					events, err := sourceCallbackEvents(lock, form, descriptor.SourceID, fmt.Sprintf("paths[%q].%s.callbacks", path, method), rawCallbacks, resolver)
					if err != nil {
						return sourceImportResult{}, err
					}
					for _, event := range events {
						if len(result.InboundEvents) >= limits.MaxOperations {
							return sourceImportResult{}, fmt.Errorf("inbound event count limit exceeded")
						}
						if err := budget.add(event, "inbound event"); err != nil {
							return sourceImportResult{}, err
						}
						result.InboundEvents = append(result.InboundEvents, event)
					}
				}
			}
		}
	}
	if len(result.Operations) == 0 && len(result.InboundEvents) == 0 {
		return sourceImportResult{}, fmt.Errorf("source artifact has no provider operations or inbound events")
	}
	sortSourceOperationDescriptors(result.Operations)
	sortSourceInboundEventDescriptors(result.InboundEvents)
	sortSourceExtensions(result.Extensions)
	if err := validateSourceImportResultIdentities(result); err != nil {
		return sourceImportResult{}, err
	}
	return result, nil
}

var sourceHTTPMethods = []string{"delete", "get", "head", "options", "patch", "post", "put", "trace"}

func validateSourceImportPath(path string) error {
	if path == "" || path != strings.TrimSpace(path) || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "://") || strings.ContainsAny(path, "\r\n?#") {
		return fmt.Errorf("not connector-relative")
	}
	return nil
}

func importSourceOperation(lock sourceImportLock, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, path, method string, pathParameters []sourceParameterValue, rootServers, pathServers sourceServerLayer, operation map[string]any, limits sourceImportLimits, remainingDescriptorBytes int64) (sourceOperationDescriptor, error) {
	location := fmt.Sprintf("paths[%q].%s", path, method)
	operationParameters, err := sourceParameterValues(operation["parameters"], resolver, form)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s parameters: %w", location, err)
	}
	request, err := sourceRequestDescriptorFrom(path, pathParameters, operationParameters, operation, doc, form, resolver, limits)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s request: %w", location, err)
	}
	responses, _, err := sourceResponses(operation, doc, form, resolver, limits, remainingDescriptorBytes)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s responses: %w", location, err)
	}
	output, err := sourceOutputForResponses(method, responses)
	if err != nil {
		return sourceOperationDescriptor{}, fmt.Errorf("%s output: %w", location, err)
	}
	providerID := ""
	if rawProviderID, declared := operation["operationId"]; declared {
		var ok bool
		providerID, ok = rawProviderID.(string)
		if !ok {
			return sourceOperationDescriptor{}, fmt.Errorf("%s.operationId must be a string", location)
		}
	}
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
	operationServers := sourceServerLayer{}
	if form.isOpenAPI() {
		operationServers, err = sourceServerLayerFrom(operation)
		if err != nil {
			return sourceOperationDescriptor{}, fmt.Errorf("%s servers: %w", location, err)
		}
	}
	servers := sourceServerOverrides{
		Root:       rootServers,
		PathItem:   pathServers,
		Operation:  operationServers,
		Precedence: []string{"operation", "path_item", "root"},
	}
	if rootServers.Declared || pathServers.Declared || operationServers.Declared {
		servers.Gaps = []sourceContractGap{sourceContractGapFor("cli-operation-route-override-foundation-r1", location+".servers", "provider-declared server routing requires runtime route-override support")}
	}
	if form.isSwagger2() {
		binding, err := sourceSwaggerRouteBindingFrom(doc, operation, path)
		if err != nil {
			return sourceOperationDescriptor{}, fmt.Errorf("%s Swagger route binding: %w", location, err)
		}
		servers.Swagger = &binding
		servers.Precedence = []string{"swagger_operation_schemes", "swagger_root"}
		if binding.Declared {
			servers.Gaps = append(servers.Gaps, sourceContractGapFor("cli-operation-route-override-foundation-r1", location+".swagger", "Swagger host, basePath, or schemes require runtime route-override support"))
		}
	}
	servers.Gaps = sourceSortedGaps(servers.Gaps)
	runtimeGaps := append(sourceRequestGaps(request), servers.Gaps...)
	runtime := sourceRuntimeReachability{MergeBlocked: len(runtimeGaps) > 0, Gaps: runtimeGaps}
	return sourceOperationDescriptor{
		Connector:           lock.Connector,
		SourceID:            sourceID,
		ProviderOperationID: providerID,
		Source:              sourceImportProvenance(lock, form, location),
		Method:              method,
		Path:                path,
		Request:             request,
		Responses:           responses,
		Output:              output,
		Pagination:          pagination,
		ByteLimits:          byteLimits,
		AuthScopes:          authScopes,
		Servers:             servers,
		Runtime:             runtime,
	}, nil
}

type sourceParameterValue struct {
	Name     string
	In       string
	Required bool
	Schema   any
	Content  any
	Wire     sourceParameterWireDescriptor
}

func sourceParameterValues(raw any, resolver *sourceReferenceResolver, form sourceDocumentForm) ([]sourceParameterValue, error) {
	if raw == nil {
		return nil, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("parameters must be an array")
	}
	values := make([]sourceParameterValue, 0, len(items))
	for _, item := range items {
		parameter, err := resolver.resolveParameter(item, nil, 0)
		if err != nil {
			return nil, err
		}
		name, _ := parameter["name"].(string)
		in, _ := parameter["in"].(string)
		allowedLocation := in == "path" || in == "query" || in == "header" || (form.isSwagger2() && in == "body")
		if name == "" || name != strings.TrimSpace(name) || !allowedLocation {
			return nil, fmt.Errorf("parameter %q has unsupported location %q", name, in)
		}
		schema, content, err := sourceParameterRepresentation(parameter, form)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}
		resolvedSchema := schema
		_, declaredSchema := parameter["schema"]
		if form.isSwagger2() && !declaredSchema && resolvedSchema != nil {
			resolvedSchema, err = resolver.resolveSchema(resolvedSchema, nil, 0)
			if err != nil {
				return nil, fmt.Errorf("parameter %q schema: %w", name, err)
			}
		}
		wire, err := sourceParameterWire(parameter, form, in)
		if err != nil {
			return nil, fmt.Errorf("parameter %q: %w", name, err)
		}
		if content != nil {
			wire.Gaps = append(wire.Gaps, sourceContractGapFor("cli-parameter-serialization-foundation-r1", "parameter "+name, "content-based parameter serialization requires runtime serialization support"))
		}
		values = append(values, sourceParameterValue{Name: name, In: in, Required: sourceBool(parameter["required"]), Schema: resolvedSchema, Content: sourceCloneLiteral(content), Wire: wire})
	}
	return values, nil
}

func sourceParameterSchema(parameter map[string]any, form sourceDocumentForm) (any, error) {
	schema, content, err := sourceParameterRepresentation(parameter, form)
	if err != nil {
		return nil, err
	}
	if content != nil {
		return nil, fmt.Errorf("uses content rather than schema")
	}
	return schema, nil
}

func sourceParameterRepresentation(parameter map[string]any, form sourceDocumentForm) (any, any, error) {
	schema, hasSchema := parameter["schema"]
	content, hasContent := parameter["content"]
	if form.isOpenAPI() {
		if hasSchema == hasContent {
			return nil, nil, fmt.Errorf("requires exactly one of schema or content")
		}
		if hasSchema {
			return schema, nil, nil
		}
		return nil, content, nil
	}
	if hasContent {
		return nil, nil, fmt.Errorf("Swagger parameter content is unsupported")
	}
	if hasSchema {
		return schema, nil, nil
	}
	excluded := map[string]bool{
		"name": true, "in": true, "description": true, "required": true, "collectionFormat": true,
		"allowEmptyValue": true, "allowReserved": true, "style": true, "explode": true,
	}
	legacySchema := map[string]any{}
	for _, key := range sortedSourceMapKeys(parameter) {
		if !excluded[key] {
			legacySchema[key] = sourceCloneLiteral(parameter[key])
		}
	}
	if len(legacySchema) == 0 {
		return nil, nil, fmt.Errorf("missing schema")
	}
	return legacySchema, nil, nil
}

func sourceRequestDescriptorFrom(path string, pathParameters, operationParameters []sourceParameterValue, operation, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, limits sourceImportLimits) (sourceRequestDescriptor, error) {
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
		if parameter.Content == nil {
			if err := validateBoundedRequestSchema(parameter.Schema, form, limits, 0); err != nil {
				return sourceRequestDescriptor{}, fmt.Errorf("parameter %q: %w", parameter.Name, err)
			}
		} else if err := validateBoundedParameterContent(parameter.Name, parameter.Content, form, limits); err != nil {
			return sourceRequestDescriptor{}, err
		}
		descriptor := sourceParameterDescriptor{Name: parameter.Name, Required: parameter.Required, Schema: parameter.Schema, Content: parameter.Content, Wire: parameter.Wire}
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
	if form.isSwagger2() {
		if len(bodyParameters) == 0 {
			return request, nil
		}
		if len(bodyParameters) != 1 {
			return sourceRequestDescriptor{}, fmt.Errorf("Swagger request body is ambiguous")
		}
		body := bodyParameters[0]
		if err := validateBoundedRequestSchema(body.Schema, form, limits, 0); err != nil {
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
	body, err := resolver.resolveRequestBody(rawBody, nil, 0)
	if err != nil {
		return sourceRequestDescriptor{}, err
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
	if err := validateBoundedRequestSchema(schema, form, limits, 0); err != nil {
		return sourceRequestDescriptor{}, err
	}
	request.Body = &sourceRequestBodyDescriptor{Required: sourceBool(body["required"]), Schema: schema}
	request.MediaType = mediaType
	return request, nil
}

func validateBoundedParameterContent(name string, content any, form sourceDocumentForm, limits sourceImportLimits) error {
	media, ok := content.(map[string]any)
	if !ok || len(media) == 0 {
		return fmt.Errorf("parameter %q content must be a non-empty object", name)
	}
	for _, mediaType := range sortedSourceMapKeys(media) {
		declaration, ok := media[mediaType].(map[string]any)
		if !ok {
			return fmt.Errorf("parameter %q content media %q must be an object", name, mediaType)
		}
		schema, exists := declaration["schema"]
		if !exists {
			return fmt.Errorf("parameter %q content media %q is missing schema", name, mediaType)
		}
		if err := validateBoundedRequestSchema(schema, form, limits, 0); err != nil {
			return fmt.Errorf("parameter %q content media %q: %w", name, mediaType, err)
		}
	}
	if len(media) != 1 {
		return fmt.Errorf("parameter %q content requires exactly one unambiguous media type", name)
	}
	return nil
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

func sourceParameterWire(parameter map[string]any, form sourceDocumentForm, location string) (sourceParameterWireDescriptor, error) {
	wire := sourceParameterWireDescriptor{Gaps: []sourceContractGap{}}
	if raw, exists := parameter["style"]; exists {
		style, ok := raw.(string)
		if !ok || style == "" || style != strings.TrimSpace(style) {
			return sourceParameterWireDescriptor{}, fmt.Errorf("style must be a non-empty string")
		}
		wire.Style = style
	}
	if raw, exists := parameter["explode"]; exists {
		value, ok := raw.(bool)
		if !ok {
			return sourceParameterWireDescriptor{}, fmt.Errorf("explode must be a boolean")
		}
		wire.Explode = &value
	}
	if raw, exists := parameter["allowReserved"]; exists {
		value, ok := raw.(bool)
		if !ok {
			return sourceParameterWireDescriptor{}, fmt.Errorf("allowReserved must be a boolean")
		}
		wire.AllowReserved = &value
	}
	if raw, exists := parameter["allowEmptyValue"]; exists {
		value, ok := raw.(bool)
		if !ok {
			return sourceParameterWireDescriptor{}, fmt.Errorf("allowEmptyValue must be a boolean")
		}
		wire.AllowEmptyValue = &value
	}
	if raw, exists := parameter["collectionFormat"]; exists {
		format, ok := raw.(string)
		if !ok || format == "" || format != strings.TrimSpace(format) {
			return sourceParameterWireDescriptor{}, fmt.Errorf("collectionFormat must be a non-empty string")
		}
		wire.CollectionFormat = format
	}
	if form.isOpenAPI() {
		supportedStyle := wire.Style == "" || wire.Style == "form" || wire.Style == "simple"
		if !supportedStyle {
			wire.Gaps = append(wire.Gaps, sourceContractGapFor("cli-parameter-serialization-foundation-r1", "parameter serialization", fmt.Sprintf("%s parameter style %q requires runtime serialization support", location, wire.Style)))
		}
	}
	if form.isSwagger2() && wire.CollectionFormat != "" && wire.CollectionFormat != "csv" {
		wire.Gaps = append(wire.Gaps, sourceContractGapFor("cli-parameter-serialization-foundation-r1", "parameter serialization", fmt.Sprintf("%s parameter collectionFormat %q requires runtime serialization support", location, wire.CollectionFormat)))
	}
	return wire, nil
}

func sourceRequestGaps(request sourceRequestDescriptor) []sourceContractGap {
	var gaps []sourceContractGap
	for _, group := range [][]sourceParameterDescriptor{request.Path, request.Query, request.Header} {
		for _, parameter := range group {
			gaps = append(gaps, parameter.Wire.Gaps...)
		}
	}
	return sourceSortedGaps(gaps)
}

func sourceContractGapFor(foundation, location, reason string) sourceContractGap {
	return sourceContractGap{Foundation: foundation, Location: location, Reason: reason}
}

func sourceSortedGaps(gaps []sourceContractGap) []sourceContractGap {
	if len(gaps) == 0 {
		return nil
	}
	out := append([]sourceContractGap(nil), gaps...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Foundation != out[j].Foundation {
			return out[i].Foundation < out[j].Foundation
		}
		if out[i].Location != out[j].Location {
			return out[i].Location < out[j].Location
		}
		return out[i].Reason < out[j].Reason
	})
	return out
}

func validateBoundedRequestSchema(schema any, form sourceDocumentForm, limits sourceImportLimits, depth int) error {
	return validateBoundedRequestSchemaWithinEnum(schema, form, limits, depth, false)
}

func validateBoundedRequestSchemaWithinEnum(schema any, form sourceDocumentForm, limits sourceImportLimits, depth int, boundedByFiniteSet bool) error {
	if depth > limits.MaxReferenceDepth {
		return fmt.Errorf("schema depth limit exceeded")
	}
	bytes, err := sourceSchemaStructuralBytes(schema, depth, limits)
	if err != nil {
		return err
	}
	if bytes > sourceSchemaByteLimit(limits) {
		return fmt.Errorf("schema byte limit exceeded")
	}
	if booleanSchema, ok := schema.(bool); ok {
		if !form.isOpenAPI31() {
			return fmt.Errorf("unsupported %s boolean schema", form.Family)
		}
		if !booleanSchema {
			return nil
		}
		if boundedByFiniteSet {
			return nil
		}
		return fmt.Errorf("unbounded request schema boolean true accepts arbitrary values")
	}
	object, ok := schema.(map[string]any)
	if !ok {
		return fmt.Errorf("unbounded request schema must be an object")
	}
	if err := sourceRejectDynamicSchemaKeywords(object); err != nil {
		return err
	}
	if err := sourceValidateSchemaForm(object, form); err != nil {
		return err
	}
	for _, composition := range []string{"oneOf", "anyOf", "allOf", "not"} {
		if _, exists := object[composition]; exists {
			return fmt.Errorf("ambiguous request schema uses %s", composition)
		}
	}
	for _, keyword := range []string{"contains", "if", "then", "else", "unevaluatedItems", "contentSchema", "$defs", "definitions", "dependencies"} {
		if _, exists := object[keyword]; exists {
			return fmt.Errorf("unsupported request schema keyword %s", keyword)
		}
	}
	for _, keyword := range []string{"patternProperties", "propertyNames", "unevaluatedProperties", "dependentSchemas", "dependentRequired"} {
		if _, exists := object[keyword]; exists {
			return fmt.Errorf("unbounded request schema object uses dynamic %s", keyword)
		}
	}
	finiteEnum, err := sourceFiniteEnum(object)
	if err != nil {
		return err
	}
	finiteConst, err := sourceFiniteConst(object, form)
	if err != nil {
		return err
	}
	bounded := boundedByFiniteSet || finiteEnum || finiteConst
	typeNames, err := sourceRequestSchemaTypes(object, form)
	if err != nil {
		return err
	}
	if len(typeNames) == 0 {
		if bounded {
			return nil
		}
		return fmt.Errorf("unbounded request schema has no type")
	}
	for _, typeName := range typeNames {
		if err := validateBoundedRequestSchemaType(object, typeName, form, limits, depth, bounded); err != nil {
			return err
		}
	}
	return nil
}

func validateBoundedRequestSchemaType(object map[string]any, typeName string, form sourceDocumentForm, limits sourceImportLimits, depth int, bounded bool) error {
	switch typeName {
	case "boolean", "null":
		return nil
	case "string":
		if err := sourceValidateLengthBounds(object, bounded); err != nil {
			return err
		}
	case "integer", "number":
		if err := sourceValidateNumericBounds(object, form, !bounded); err != nil {
			return err
		}
	case "array":
		prefixCount := 0
		if rawPrefix, exists := object["prefixItems"]; exists {
			if !form.isOpenAPI31() {
				return fmt.Errorf("unsupported request schema keyword prefixItems")
			}
			prefixItems, ok := rawPrefix.([]any)
			if !ok {
				return fmt.Errorf("request schema prefixItems must be an array")
			}
			prefixCount = len(prefixItems)
			for index, item := range prefixItems {
				if err := validateBoundedRequestSchemaWithinEnum(item, form, limits, depth+1, bounded); err != nil {
					return fmt.Errorf("prefix item %d: %w", index, err)
				}
			}
		}
		items, exists := object["items"]
		itemsFalse := false
		if exists {
			if booleanItems, isBoolean := items.(bool); isBoolean && !form.isOpenAPI31() {
				return fmt.Errorf("request schema items must be an object")
			} else if isBoolean && !booleanItems {
				itemsFalse = true
			} else if err := validateBoundedRequestSchemaWithinEnum(items, form, limits, depth+1, bounded); err != nil {
				return err
			}
		}
		var closedTupleLimit *int64
		if itemsFalse {
			limit := int64(prefixCount)
			closedTupleLimit = &limit
		}
		if err := sourceValidateArrayBounds(object, bounded, closedTupleLimit); err != nil {
			return err
		}
		if !exists && !bounded && closedTupleLimit == nil {
			maxItems, hasMaxItems, err := sourceOptionalNonNegativeInteger(object, "maxItems")
			if err != nil {
				return err
			}
			if !hasMaxItems || int64(prefixCount) < maxItems {
				return fmt.Errorf("unbounded request schema array has no items")
			}
		}
	case "object":
		additional, exists := object["additionalProperties"]
		if !bounded && (!exists || additional != false) {
			return fmt.Errorf("unbounded request schema object has dynamic additionalProperties")
		}
		properties, exists := object["properties"]
		if !exists {
			if !bounded {
				return fmt.Errorf("unbounded request schema object has no fixed properties")
			}
			return nil
		}
		propertyMap, ok := properties.(map[string]any)
		if !ok {
			return fmt.Errorf("unbounded request schema object properties are invalid")
		}
		for _, name := range sortedSourceMapKeys(propertyMap) {
			if err := validateBoundedRequestSchemaWithinEnum(propertyMap[name], form, limits, depth+1, bounded); err != nil {
				return fmt.Errorf("property %q: %w", name, err)
			}
		}
	default:
		return fmt.Errorf("unsupported request schema type %q", typeName)
	}
	return nil
}

func sourceRequestSchemaTypes(object map[string]any, form sourceDocumentForm) ([]string, error) {
	raw, declared := object["type"]
	if !declared {
		return nil, nil
	}
	var values []string
	switch typed := raw.(type) {
	case string:
		values = []string{typed}
	case []any:
		if !form.isOpenAPI31() {
			return nil, fmt.Errorf("request schema type union requires OpenAPI 3.1")
		}
		values = make([]string, 0, len(typed))
		for _, item := range typed {
			value, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("request schema type union must contain strings")
			}
			values = append(values, value)
		}
	case []string:
		if !form.isOpenAPI31() {
			return nil, fmt.Errorf("request schema type union requires OpenAPI 3.1")
		}
		values = append([]string(nil), typed...)
	default:
		return nil, fmt.Errorf("request schema type must be a string or string array")
	}
	if len(values) == 0 {
		return nil, fmt.Errorf("request schema type union must not be empty")
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			return nil, fmt.Errorf("request schema type union must contain distinct non-empty strings")
		}
		seen[value] = true
	}
	return values, nil
}

func sourceValidateSchemaForm(object map[string]any, form sourceDocumentForm) error {
	if _, exists := object["const"]; exists && !form.isOpenAPI31() {
		return fmt.Errorf("unsupported %s schema keyword const", form.Family)
	}
	if raw, exists := object["type"]; exists {
		if _, isArray := raw.([]any); isArray && !form.isOpenAPI31() {
			return fmt.Errorf("request schema type union requires OpenAPI 3.1")
		}
	}
	if raw, exists := object["nullable"]; exists {
		if !form.isOpenAPI() || form.isOpenAPI31() {
			return fmt.Errorf("unsupported %s schema keyword nullable", form.Family)
		}
		if _, ok := raw.(bool); !ok {
			return fmt.Errorf("OpenAPI 3.0 schema nullable must be a boolean")
		}
	}
	if raw, exists := object["items"]; exists {
		if _, isBoolean := raw.(bool); isBoolean && !form.isOpenAPI31() {
			return fmt.Errorf("unsupported %s boolean schema items", form.Family)
		}
	}
	for _, key := range []string{"exclusiveMinimum", "exclusiveMaximum"} {
		raw, exists := object[key]
		if !exists {
			continue
		}
		if form.isOpenAPI31() {
			if _, ok := sourceFiniteNumber(raw); !ok {
				return fmt.Errorf("OpenAPI 3.1 schema %s must be a finite number", key)
			}
			continue
		}
		flag, ok := raw.(bool)
		if !ok {
			return fmt.Errorf("%s schema %s must be a boolean", form.Family, key)
		}
		if flag {
			bound := "minimum"
			if key == "exclusiveMaximum" {
				bound = "maximum"
			}
			if _, exists := object[bound]; !exists {
				return fmt.Errorf("%s schema %s requires %s", form.Family, key, bound)
			}
		}
	}
	return nil
}

func sourceRejectDynamicSchemaKeywords(object map[string]any) error {
	for _, keyword := range []string{"$dynamicRef", "$dynamicAnchor", "$recursiveRef", "$recursiveAnchor"} {
		if _, exists := object[keyword]; exists {
			return fmt.Errorf("cli-openapi-dynamic-ref-foundation-r1: request schema uses %s", keyword)
		}
	}
	return nil
}

func sourceFiniteEnum(object map[string]any) (bool, error) {
	raw, exists := object["enum"]
	if !exists {
		return false, nil
	}
	values, ok := raw.([]any)
	if !ok || len(values) == 0 {
		return false, fmt.Errorf("request schema enum must be a non-empty array")
	}
	for _, value := range values {
		if !sourceFiniteSchemaLiteral(value) {
			return false, fmt.Errorf("request schema enum contains a non-finite value")
		}
	}
	return true, nil
}

func sourceFiniteConst(object map[string]any, form sourceDocumentForm) (bool, error) {
	raw, exists := object["const"]
	if !exists {
		return false, nil
	}
	if !form.isOpenAPI31() {
		return false, fmt.Errorf("unsupported %s schema keyword const", form.Family)
	}
	if !sourceFiniteSchemaLiteral(raw) {
		return false, fmt.Errorf("request schema const contains a non-finite value")
	}
	return true, nil
}

func sourceFiniteSchemaLiteral(value any) bool {
	switch typed := value.(type) {
	case float64:
		return !math.IsNaN(typed) && !math.IsInf(typed, 0)
	case json.Number:
		_, ok := sourceFiniteNumber(typed)
		return ok
	case []any:
		for _, item := range typed {
			if !sourceFiniteSchemaLiteral(item) {
				return false
			}
		}
	case map[string]any:
		for _, key := range sortedSourceMapKeys(typed) {
			if !sourceFiniteSchemaLiteral(typed[key]) {
				return false
			}
		}
	}
	return true
}

func sourceValidateLengthBounds(object map[string]any, finiteEnum bool) error {
	minimum, hasMinimum, err := sourceOptionalNonNegativeInteger(object, "minLength")
	if err != nil {
		return err
	}
	maximum, hasMaximum, err := sourceOptionalNonNegativeInteger(object, "maxLength")
	if err != nil {
		return err
	}
	if hasMinimum && hasMaximum && minimum > maximum {
		return fmt.Errorf("contradictory request schema string length bounds")
	}
	if !finiteEnum && !hasMaximum {
		return fmt.Errorf("unbounded request schema string has no maxLength")
	}
	return nil
}

func sourceValidateArrayBounds(object map[string]any, finiteEnum bool, closedTupleLimit *int64) error {
	minimum, hasMinimum, err := sourceOptionalNonNegativeInteger(object, "minItems")
	if err != nil {
		return err
	}
	maximum, hasMaximum, err := sourceOptionalNonNegativeInteger(object, "maxItems")
	if err != nil {
		return err
	}
	if hasMinimum && hasMaximum && minimum > maximum {
		return fmt.Errorf("contradictory request schema array item bounds")
	}
	if closedTupleLimit != nil {
		if hasMinimum && minimum > *closedTupleLimit {
			return fmt.Errorf("contradictory request schema array item bounds")
		}
		return nil
	}
	if !finiteEnum && !hasMaximum {
		return fmt.Errorf("unbounded request schema array has no maxItems")
	}
	return nil
}

type sourceNumericBound struct {
	value     *big.Rat
	inclusive bool
	present   bool
}

func sourceValidateNumericBounds(object map[string]any, form sourceDocumentForm, requireBoth bool) error {
	lower, err := sourceNumericBoundFor(object, form, "minimum", "exclusiveMinimum", true)
	if err != nil {
		return err
	}
	upper, err := sourceNumericBoundFor(object, form, "maximum", "exclusiveMaximum", false)
	if err != nil {
		return err
	}
	if requireBoth && !lower.present {
		return fmt.Errorf("unbounded request schema number has no minimum")
	}
	if requireBoth && !upper.present {
		return fmt.Errorf("unbounded request schema number has no maximum")
	}
	if lower.present && upper.present && (lower.value.Cmp(upper.value) > 0 || (lower.value.Cmp(upper.value) == 0 && (!lower.inclusive || !upper.inclusive))) {
		return fmt.Errorf("contradictory request schema numeric bounds")
	}
	return nil
}

func sourceNumericBoundFor(object map[string]any, form sourceDocumentForm, inclusiveKey, exclusiveKey string, lower bool) (sourceNumericBound, error) {
	bound := sourceNumericBound{inclusive: true}
	if raw, exists := object[inclusiveKey]; exists {
		value, ok := sourceFiniteNumber(raw)
		if !ok {
			return sourceNumericBound{}, fmt.Errorf("request schema %s must be a finite number", inclusiveKey)
		}
		bound = sourceNumericBound{value: value, inclusive: true, present: true}
	}
	rawExclusive, exists := object[exclusiveKey]
	if !exists {
		return bound, nil
	}
	if form.isOpenAPI31() {
		exclusive, ok := sourceFiniteNumber(rawExclusive)
		if !ok {
			return sourceNumericBound{}, fmt.Errorf("request schema %s must be a finite number", exclusiveKey)
		}
		if !bound.present || (lower && exclusive.Cmp(bound.value) > 0) || (!lower && exclusive.Cmp(bound.value) < 0) {
			return sourceNumericBound{value: exclusive, inclusive: false, present: true}, nil
		}
		if exclusive.Cmp(bound.value) == 0 {
			bound.inclusive = false
		}
		return bound, nil
	}
	if flag, isFlag := rawExclusive.(bool); isFlag {
		if flag {
			if !bound.present {
				return sourceNumericBound{}, fmt.Errorf("request schema %s requires %s", exclusiveKey, inclusiveKey)
			}
			bound.inclusive = false
		}
		return bound, nil
	}
	return sourceNumericBound{}, fmt.Errorf("request schema %s must be a boolean", exclusiveKey)
}

func sourceOptionalNonNegativeInteger(object map[string]any, key string) (int64, bool, error) {
	raw, exists := object[key]
	if !exists {
		return 0, false, nil
	}
	value, ok := sourceNonNegativeInteger(raw)
	if !ok {
		return 0, false, fmt.Errorf("request schema %s must be a non-negative integer", key)
	}
	return value, true, nil
}

func sourceNonNegativeInteger(value any) (int64, bool) {
	switch typed := value.(type) {
	case json.Number:
		integer, err := typed.Int64()
		return integer, err == nil && integer >= 0
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) || typed < 0 || typed > math.MaxInt64 || math.Trunc(typed) != typed {
			return 0, false
		}
		return int64(typed), true
	case int:
		return int64(typed), typed >= 0
	case int64:
		return typed, typed >= 0
	case uint:
		if uint64(typed) > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	case uint64:
		if typed > math.MaxInt64 {
			return 0, false
		}
		return int64(typed), true
	default:
		return 0, false
	}
}

func sourceFiniteNumber(value any) (*big.Rat, bool) {
	var encoded string
	switch typed := value.(type) {
	case json.Number:
		encoded = string(typed)
	case float64:
		if math.IsNaN(typed) || math.IsInf(typed, 0) {
			return nil, false
		}
		encoded = strconv.FormatFloat(typed, 'g', -1, 64)
	case float32:
		if math.IsNaN(float64(typed)) || math.IsInf(float64(typed), 0) {
			return nil, false
		}
		encoded = strconv.FormatFloat(float64(typed), 'g', -1, 32)
	case int:
		encoded = strconv.Itoa(typed)
	case int64:
		encoded = strconv.FormatInt(typed, 10)
	case uint:
		encoded = strconv.FormatUint(uint64(typed), 10)
	case uint64:
		encoded = strconv.FormatUint(typed, 10)
	default:
		return nil, false
	}
	return sourceExactDecimal(encoded)
}

const sourceMaximumExactNumberScale = 100_000

func sourceExactDecimal(value string) (*big.Rat, bool) {
	if value == "" {
		return nil, false
	}
	sign := 1
	if value[0] == '-' {
		sign = -1
		value = value[1:]
	}
	if value == "" || value[0] == '+' {
		return nil, false
	}
	mantissa, exponentText, hasExponent := value, "", false
	if index := strings.IndexAny(value, "eE"); index >= 0 {
		if strings.IndexAny(value[index+1:], "eE") >= 0 {
			return nil, false
		}
		mantissa, exponentText, hasExponent = value[:index], value[index+1:], true
	}
	if mantissa == "" || (hasExponent && exponentText == "") {
		return nil, false
	}
	whole, fraction := mantissa, ""
	if index := strings.IndexByte(mantissa, '.'); index >= 0 {
		if strings.IndexByte(mantissa[index+1:], '.') >= 0 {
			return nil, false
		}
		whole, fraction = mantissa[:index], mantissa[index+1:]
		if whole == "" || fraction == "" {
			return nil, false
		}
	}
	for _, part := range []string{whole, fraction} {
		for _, character := range part {
			if character < '0' || character > '9' {
				return nil, false
			}
		}
	}
	exponent := int64(0)
	if hasExponent {
		parsed, err := strconv.ParseInt(exponentText, 10, 64)
		if err != nil || parsed > sourceMaximumExactNumberScale || parsed < -sourceMaximumExactNumberScale {
			return nil, false
		}
		exponent = parsed
	}
	scale := exponent - int64(len(fraction))
	if scale > sourceMaximumExactNumberScale || scale < -sourceMaximumExactNumberScale {
		return nil, false
	}
	numerator, ok := new(big.Int).SetString(whole+fraction, 10)
	if !ok {
		return nil, false
	}
	if sign < 0 {
		numerator.Neg(numerator)
	}
	if scale >= 0 {
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(scale), nil)
		numerator.Mul(numerator, factor)
		return new(big.Rat).SetInt(numerator), true
	}
	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(-scale), nil)
	return new(big.Rat).SetFrac(numerator, denominator), true
}

func sourceResponses(operation, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, limits sourceImportLimits, remainingDescriptorBytes int64) ([]sourceResponseDescriptor, []string, error) {
	rawResponses, ok := operation["responses"].(map[string]any)
	if !ok || len(rawResponses) == 0 {
		return nil, nil, fmt.Errorf("missing responses")
	}
	responses := make([]sourceResponseDescriptor, 0, len(rawResponses))
	responseLimit := sourceResolvedDescriptorLimit(limits)
	if remainingDescriptorBytes < responseLimit {
		responseLimit = remainingDescriptorBytes
	}
	responseBudget := sourceImportBudget{limit: responseLimit}
	responseExpansion := sourceResponseExpansionBudget{limit: responseLimit}
	previousScope := resolver.responseScope
	resolver.responseScope = &responseExpansion
	defer func() { resolver.responseScope = previousScope }()
	mediaSet := map[string]bool{}
	var swaggerProduces []string
	if form.isSwagger2() {
		produces, declared := operation["produces"]
		if !declared {
			produces, declared = doc["produces"]
		}
		if declared {
			var err error
			swaggerProduces, err = sourceStringArray(produces, "Swagger produces")
			if err != nil {
				return nil, nil, err
			}
		}
	}
	for _, status := range sortedSourceMapKeys(rawResponses) {
		if strings.HasPrefix(status, "x-") {
			continue
		}
		resolved, err := resolver.resolveResponse(rawResponses[status], nil, 0)
		if err != nil {
			return nil, nil, err
		}
		encoded, err := sourceMarshalCompact(resolved)
		if err != nil {
			return nil, nil, fmt.Errorf("encode response %q: %w", status, err)
		}
		if int64(len(encoded)) > limits.MaxSchemaBytes {
			return nil, nil, fmt.Errorf("schema byte limit exceeded for response %q", status)
		}
		media, err := sourceResponseMedia(resolved, swaggerProduces)
		if err != nil {
			return nil, nil, fmt.Errorf("response %q: %w", status, err)
		}
		if err := responseBudget.reserve(sourceResponseDescriptorEstimate(status, int64(len(encoded)), media), "response"); err != nil {
			return nil, nil, err
		}
		responses = append(responses, sourceResponseDescriptor{Status: status, Declaration: resolved, Media: media})
		for _, item := range media {
			mediaSet[item.MediaType] = true
		}
	}
	mediaTypes := make([]string, 0, len(mediaSet))
	for mediaType := range mediaSet {
		mediaTypes = append(mediaTypes, mediaType)
	}
	sort.Strings(mediaTypes)
	return responses, mediaTypes, nil
}

func sourceResponseDescriptorEstimate(status string, declarationBytes int64, media []sourceResponseMediaDescriptor) int64 {
	used := declarationBytes
	for _, value := range append([]string{status}, sourceResponseMediaEstimateValues(media)...) {
		if used > math.MaxInt64-sourceStructuralStringBytes(value) {
			return math.MaxInt64
		}
		used += sourceStructuralStringBytes(value)
	}
	if used > math.MaxInt64-96 {
		return math.MaxInt64
	}
	return used + 96
}

func sourceResponseMediaEstimateValues(media []sourceResponseMediaDescriptor) []string {
	values := make([]string, 0, len(media)*2)
	for _, item := range media {
		values = append(values, item.MediaType, string(item.Class))
	}
	return values
}

func sourceResponseMedia(response map[string]any, fallback []string) ([]sourceResponseMediaDescriptor, error) {
	mediaSet := map[string]bool{}
	if content, exists := response["content"]; exists {
		contentMap, ok := content.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("response content must be an object")
		}
		for mediaType := range contentMap {
			mediaSet[mediaType] = true
		}
	}
	if len(mediaSet) == 0 {
		for _, mediaType := range fallback {
			mediaSet[mediaType] = true
		}
	}
	mediaTypes := make([]string, 0, len(mediaSet))
	for mediaType := range mediaSet {
		mediaTypes = append(mediaTypes, mediaType)
	}
	sort.Strings(mediaTypes)
	media := make([]sourceResponseMediaDescriptor, 0, len(mediaTypes))
	for _, mediaType := range mediaTypes {
		class, err := sourceOutputClassForMediaType(mediaType)
		if err != nil {
			return nil, err
		}
		media = append(media, sourceResponseMediaDescriptor{MediaType: mediaType, Class: class})
	}
	return media, nil
}

func sourceOutputForResponses(method string, responses []sourceResponseDescriptor) (sourceOutputDescriptor, error) {
	output := sourceOutputDescriptor{Success: []sourceOutputVariant{}}
	mediaSet := map[string]bool{}
	classSet := map[sourceOutputClass]bool{}
	for _, response := range responses {
		if !sourceSuccessfulResponseStatus(response.Status) {
			continue
		}
		if method == "head" || len(response.Media) == 0 {
			output.Success = append(output.Success, sourceOutputVariant{Status: response.Status, Class: sourceOutputStatus})
			classSet[sourceOutputStatus] = true
			continue
		}
		for _, media := range response.Media {
			output.Success = append(output.Success, sourceOutputVariant{Status: response.Status, MediaType: media.MediaType, Class: media.Class})
			mediaSet[media.MediaType] = true
			classSet[media.Class] = true
		}
	}
	if len(output.Success) == 0 {
		output.Class = sourceOutputStatus
		return output, nil
	}
	sort.Slice(output.Success, func(i, j int) bool {
		if output.Success[i].Status != output.Success[j].Status {
			return output.Success[i].Status < output.Success[j].Status
		}
		return output.Success[i].MediaType < output.Success[j].MediaType
	})
	if len(classSet) == 1 {
		for class := range classSet {
			output.Class = class
		}
	}
	for mediaType := range mediaSet {
		output.MediaTypes = append(output.MediaTypes, mediaType)
	}
	sort.Strings(output.MediaTypes)
	return output, nil
}

func sourceSuccessfulResponseStatus(status string) bool {
	if strings.EqualFold(status, "2XX") {
		return true
	}
	code, err := strconv.Atoi(status)
	return err == nil && code >= 200 && code < 300
}

func sourceOutputClassFor(method string, mediaTypes []string) (sourceOutputClass, error) {
	if method == "head" || len(mediaTypes) == 0 {
		return sourceOutputStatus, nil
	}
	var class sourceOutputClass
	for _, mediaType := range mediaTypes {
		mediaClass, err := sourceOutputClassForMediaType(mediaType)
		if err != nil {
			return "", err
		}
		if class != "" && class != mediaClass {
			return "", fmt.Errorf("ambiguous response output media types")
		}
		class = mediaClass
	}
	return class, nil
}

func sourceOutputClassForMediaType(mediaType string) (sourceOutputClass, error) {
	normalizedMediaType := strings.ToLower(strings.TrimSpace(strings.Split(mediaType, ";")[0]))
	if normalizedMediaType == "" {
		return "", fmt.Errorf("response output media type is empty")
	}
	if sourceJSONMediaType(mediaType) {
		return sourceOutputJSON, nil
	}
	if strings.HasPrefix(normalizedMediaType, "text/") {
		return sourceOutputText, nil
	}
	return sourceOutputBinary, nil
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

func sourceServerLayerFrom(object map[string]any) (sourceServerLayer, error) {
	raw, declared := object["servers"]
	if !declared {
		return sourceServerLayer{}, nil
	}
	servers, ok := raw.([]any)
	if !ok {
		return sourceServerLayer{}, fmt.Errorf("servers must be an array")
	}
	return sourceServerLayer{Declared: true, Servers: sourceCloneLiteral(servers)}, nil
}

func sourceSwaggerRouteBindingFrom(root, operation map[string]any, path string) (sourceSwaggerRouteBinding, error) {
	binding := sourceSwaggerRouteBinding{EffectivePath: path, Precedence: []string{"operation_schemes", "root"}}
	if host, declared := root["host"]; declared {
		value, ok := host.(string)
		if !ok || value == "" || value != strings.TrimSpace(value) {
			return sourceSwaggerRouteBinding{}, fmt.Errorf("host must be a non-empty string")
		}
		binding.Declared = true
		binding.Host = value
	}
	if basePath, declared := root["basePath"]; declared {
		value, ok := basePath.(string)
		if !ok || value == "" || value != strings.TrimSpace(value) || !strings.HasPrefix(value, "/") || strings.HasPrefix(value, "//") || strings.ContainsAny(value, "\r\n?#") {
			return sourceSwaggerRouteBinding{}, fmt.Errorf("basePath must be a fixed relative path")
		}
		binding.Declared = true
		binding.BasePath = value
		binding.EffectivePath = sourceSwaggerEffectivePath(value, path)
	}
	if schemes, declared := root["schemes"]; declared {
		values, err := sourceOrderedStringArray(schemes, "Swagger schemes")
		if err != nil {
			return sourceSwaggerRouteBinding{}, err
		}
		binding.Declared = true
		binding.RootSchemes = values
		binding.Schemes = values
	}
	if schemes, declared := operation["schemes"]; declared {
		values, err := sourceOrderedStringArray(schemes, "Swagger operation schemes")
		if err != nil {
			return sourceSwaggerRouteBinding{}, err
		}
		binding.Declared = true
		binding.OperationSchemes = values
		binding.Schemes = values
	}
	return binding, nil
}

func sourceSwaggerEffectivePath(basePath, path string) string {
	if basePath == "" || basePath == "/" {
		return path
	}
	return strings.TrimSuffix(basePath, "/") + path
}

func sourceOrderedStringArray(value any, field string) ([]string, error) {
	items, ok := value.([]any)
	if !ok || len(items) == 0 {
		return nil, fmt.Errorf("%s must be a non-empty array", field)
	}
	result := make([]string, len(items))
	for index, item := range items {
		stringValue, ok := item.(string)
		if !ok || stringValue == "" || stringValue != strings.TrimSpace(stringValue) {
			return nil, fmt.Errorf("%s must contain non-empty strings", field)
		}
		result[index] = stringValue
	}
	return result, nil
}

func sourceImportProvenance(lock sourceImportLock, form sourceDocumentForm, location string) sourceImportSource {
	return sourceImportSource{
		URL:      lock.Rest.SourceURL,
		SHA256:   strings.ToLower(lock.Rest.SHA256),
		Bytes:    lock.Rest.Bytes,
		Location: location,
		Form:     form.Family,
		Version:  form.Version,
	}
}

func sourceWebhookEvents(lock sourceImportLock, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver) ([]sourceInboundEventDescriptor, error) {
	if !form.isOpenAPI31() {
		return nil, nil
	}
	raw, declared := doc["webhooks"]
	if !declared {
		return nil, nil
	}
	webhooks, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("webhooks must be an object")
	}
	events := make([]sourceInboundEventDescriptor, 0, len(webhooks))
	for _, name := range sortedSourceMapKeys(webhooks) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		if name == "" || name != strings.TrimSpace(name) {
			return nil, fmt.Errorf("webhook has an invalid name")
		}
		declaration, err := resolver.resolveInboundPathItem(webhooks[name], nil, 0)
		if err != nil {
			return nil, fmt.Errorf("webhook %q: %w", name, err)
		}
		location := fmt.Sprintf("webhooks[%q]", name)
		gap := sourceContractGapFor("cli-webhook-event-surface-foundation-r1", location, "provider inbound event requires webhook event-surface support")
		events = append(events, sourceInboundEventDescriptor{
			Connector:   lock.Connector,
			SourceID:    fmt.Sprintf("%s.inbound.webhook.%s", lock.Connector, name),
			Kind:        "webhook",
			Name:        name,
			Source:      sourceImportProvenance(lock, form, location),
			Declaration: declaration,
			Runtime:     sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{gap}},
		})
	}
	sortSourceInboundEventDescriptors(events)
	return events, nil
}

func sourceCallbackEvents(lock sourceImportLock, form sourceDocumentForm, parentSourceID, location string, raw any, resolver *sourceReferenceResolver) ([]sourceInboundEventDescriptor, error) {
	callbacks, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("callbacks must be an object")
	}
	events := make([]sourceInboundEventDescriptor, 0, len(callbacks))
	for _, name := range sortedSourceMapKeys(callbacks) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		if name == "" || name != strings.TrimSpace(name) {
			return nil, fmt.Errorf("callback has an invalid name")
		}
		declaration, err := resolver.resolveCallback(callbacks[name], nil, 0)
		if err != nil {
			return nil, fmt.Errorf("callback %q: %w", name, err)
		}
		eventLocation := fmt.Sprintf("%s[%q]", location, name)
		identity := parentSourceID
		if identity == "" {
			identity = eventLocation
		}
		gap := sourceContractGapFor("cli-webhook-event-surface-foundation-r1", eventLocation, "provider callback requires webhook event-surface support")
		events = append(events, sourceInboundEventDescriptor{
			Connector:      lock.Connector,
			SourceID:       fmt.Sprintf("%s.inbound.callback.%s.%s", lock.Connector, identity, name),
			ParentSourceID: parentSourceID,
			Kind:           "callback",
			Name:           name,
			Source:         sourceImportProvenance(lock, form, eventLocation),
			Declaration:    declaration,
			Runtime:        sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{gap}},
		})
	}
	sortSourceInboundEventDescriptors(events)
	return events, nil
}

func sortSourceInboundEventDescriptors(events []sourceInboundEventDescriptor) {
	sort.Slice(events, func(i, j int) bool {
		if events[i].Connector != events[j].Connector {
			return events[i].Connector < events[j].Connector
		}
		if events[i].Kind != events[j].Kind {
			return events[i].Kind < events[j].Kind
		}
		if events[i].Source.Location != events[j].Source.Location {
			return events[i].Source.Location < events[j].Source.Location
		}
		return events[i].SourceID < events[j].SourceID
	})
}

func sortSourceExtensions(extensions []sourceExtensionDescriptor) {
	sort.Slice(extensions, func(i, j int) bool { return extensions[i].Location < extensions[j].Location })
}

func sourceMarshalCompact(value any) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buffer.Bytes(), []byte("\n")), nil
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
	integer, ok := sourceNonNegativeInteger(value)
	if !ok || integer <= 0 {
		return 0
	}
	return integer
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
	return marshalSourceImportResult(sourceImportResult{Operations: descriptors})
}

func marshalSourceImportResult(result sourceImportResult) ([]byte, error) {
	copyResult := sourceImportResult{
		Operations:    append([]sourceOperationDescriptor(nil), result.Operations...),
		InboundEvents: append([]sourceInboundEventDescriptor(nil), result.InboundEvents...),
		Extensions:    append([]sourceExtensionDescriptor(nil), result.Extensions...),
	}
	sortSourceOperationDescriptors(copyResult.Operations)
	sortSourceInboundEventDescriptors(copyResult.InboundEvents)
	sortSourceExtensions(copyResult.Extensions)
	var gaps []sourceContractGap
	for _, operation := range copyResult.Operations {
		gaps = append(gaps, operation.Runtime.Gaps...)
	}
	for _, event := range copyResult.InboundEvents {
		gaps = append(gaps, event.Runtime.Gaps...)
	}
	gaps = sourceSortedGaps(gaps)
	document := sourceImportDescriptorDocument{
		SchemaVersion: 1,
		Operations:    copyResult.Operations,
		InboundEvents: copyResult.InboundEvents,
		Extensions:    copyResult.Extensions,
		MergeBlocked:  len(gaps) > 0,
		Gaps:          gaps,
	}
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(document); err != nil {
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
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		logln(stderr, "connectorgen source-import:", err)
		return 1
	}
	raw, err := marshalSourceImportResult(result)
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
		logln(stdout, fmt.Sprintf("connectorgen source-import: %s, %d operation(s), %d inbound event(s) verified", opts.Connector, len(result.Operations), len(result.InboundEvents)))
		return 0
	}
	if err := os.WriteFile(opts.Output, raw, 0o644); err != nil {
		logln(stderr, "connectorgen source-import: write descriptors:", err)
		return 1
	}
	logln(stdout, fmt.Sprintf("connectorgen source-import: %s, %d operation(s), %d inbound event(s) imported", opts.Connector, len(result.Operations), len(result.InboundEvents)))
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
