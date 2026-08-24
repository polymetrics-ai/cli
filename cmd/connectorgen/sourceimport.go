package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"

	"polymetrics.ai/internal/connectors/engine"
)

// source-import turns a verified, connector-owned provider source into a
// canonical intermediate descriptor. It intentionally owns no execution
// controls: a connector name selects its checked-in lock, and that lock alone
// supplies the artifact URL, byte count, and digest.
const sourceImportUsage = `connectorgen source-import <connector> [--out <path>] [--defs <dir>] [--cache-dir <dir>] [--check]

Verifies the connector-owned source lock, retrieves only its fixed public
artifact URL or document artifact URLs, and writes canonical provider
operation descriptors for later declaration generators.

  <connector>     connector whose sources/<connector>-operation-source-lock.json is used
  --out <path>    descriptor output path (default: connector-owned canonical descriptor)
  --defs <dir>    connector defs root (default internal/connectors/defs)
  --cache-dir <dir> existing isolated verified-artifact cache root (default: platform cache)
  --check         compare generated descriptors with --out; do not write

The source lock is authoritative. A byte or SHA-256 mismatch requires a
source-lock refresh; this command never accepts a replacement URL, method,
path, header, body, credential, or generic request input. Version 3 locks may
cite provider URLs with bounded queries, but the importer never fetches those
citations: it retrieves only their stable queryless document artifacts. A cold
fetch is bounded to three minutes; later checks re-verify a content-addressed
cache before using it.`

const (
	defaultSourceImportArtifactBytes = int64(16 << 20)
	// The pinned GitHub REST description contains a few legitimate resolved
	// response unions larger than 1 MiB. This is a per-schema expansion bound,
	// independent from both artifact bytes and the aggregate descriptor budget.
	defaultSourceImportSchemaBytes = int64(32 << 20)
	// Aggregate descriptor accounting is deliberately independent from both
	// artifact and index size. The pinned GitHub inventory expands referenced
	// request/response declarations well beyond the compressed source bytes.
	defaultSourceImportDescriptorBytes    = int64(128 << 20)
	defaultSourceImportOperations         = 10_000
	defaultSourceImportReferences         = 50_000
	defaultSourceImportReferenceDepth     = 32
	defaultSourceImportSchemaNodes        = 100_000
	defaultSourceImportDocuments          = 256
	defaultSourceImportTotalArtifactBytes = int64(256 << 20)
	// A locked provider artifact may be large enough that a cold public fetch
	// takes longer than an ordinary interactive command. Keep the request
	// bounded; subsequent checks use the verified content-addressed cache.
	defaultSourceImportFetchTimeout  = 3 * time.Minute
	defaultSourceImportCorpusTimeout = 10 * time.Minute
	defaultSourceImportFetchWorkers  = 4
)

type sourceImportLimits struct {
	MaxArtifactBytes           int64
	MaxIndexBytes              int64
	MaxSchemaBytes             int64
	MaxResolvedDescriptorBytes int64
	MaxOperations              int
	MaxReferences              int
	MaxReferenceDepth          int
	MaxSchemaNodes             int
	MaxDocuments               int
	MaxTotalArtifactBytes      int64
	AllowSourceContractGaps    bool
	// UseExecutionEnvelopes is enabled only for v3 immutable source
	// descriptors. Legacy v1/v2 descriptors retain their byte identity until
	// their locks are deliberately migrated rather than silently rewriting the
	// 41 MiB checked-in GitHub descriptor.
	UseExecutionEnvelopes bool
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
		MaxIndexBytes:              64 << 20,
		MaxSchemaBytes:             defaultSourceImportSchemaBytes,
		MaxResolvedDescriptorBytes: defaultSourceImportDescriptorBytes,
		MaxOperations:              defaultSourceImportOperations,
		MaxReferences:              defaultSourceImportReferences,
		MaxReferenceDepth:          defaultSourceImportReferenceDepth,
		MaxSchemaNodes:             defaultSourceImportSchemaNodes,
		MaxDocuments:               defaultSourceImportDocuments,
		MaxTotalArtifactBytes:      defaultSourceImportTotalArtifactBytes,
	}
}

type sourceImportArtifact struct {
	SourceURL     string `json:"source_url"`
	SHA256        string `json:"sha256"`
	Bytes         int64  `json:"bytes"`
	OpenAPI       string `json:"openapi,omitempty"`
	Swagger       string `json:"swagger,omitempty"`
	IdentityQuery bool   `json:"identity_query,omitempty"`
}

type sourceImportRESTOperation struct {
	ID             string `json:"id"`
	Protocol       string `json:"protocol"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	OperationID    string `json:"operation_id"`
	Deprecated     bool   `json:"deprecated"`
	SourceLocation string `json:"source_location"`
}

type sourceImportREST struct {
	sourceImportArtifact
	Commit          string                      `json:"commit,omitempty"`
	InfoVersion     string                      `json:"info_version,omitempty"`
	Operations      []sourceImportRESTOperation `json:"operations,omitempty"`
	Retrieval       string                      `json:"-"`
	OpenAPIVersions []string                    `json:"-"`
	SourceDocuments []sourceImportRESTDocument  `json:"-"`
}

// sourceImportPublishedSource records the provider document cited by an
// immutable artifact. Its source URL is provenance only: source import never
// fetches it, so a provider's short-lived capture query cannot widen the
// importer request surface.
type sourceImportPublishedSource struct {
	SourceURL  string `json:"source_url"`
	CaptureURL string `json:"capture_url"`
	SHA256     string `json:"sha256"`
	Bytes      int64  `json:"bytes"`
	Adapter    string `json:"adapter"`
}

// sourceImportRESTDocument is a v3 document-owned REST inventory. A document
// retains both the stable bytes parsed by the importer and the provider
// publication from which a capture adapter derived them.
type sourceImportRESTDocument struct {
	ID              string                      `json:"id"`
	Artifact        sourceImportArtifact        `json:"artifact"`
	PublishedSource sourceImportPublishedSource `json:"published_source"`
	InfoVersion     string                      `json:"info_version,omitempty"`
	Operations      []sourceImportRESTOperation `json:"operations"`
}

type sourceGraphQLTypeRef struct {
	Kind    string                `json:"kind"`
	Name    string                `json:"name,omitempty"`
	NonNull bool                  `json:"non_null"`
	OfType  *sourceGraphQLTypeRef `json:"of_type,omitempty"`
}

type sourceGraphQLArgument struct {
	Name string               `json:"name"`
	Type sourceGraphQLTypeRef `json:"type"`
}

type sourceGraphQLField struct {
	Root       string                  `json:"root"`
	Name       string                  `json:"name"`
	Line       int                     `json:"line"`
	Signature  string                  `json:"signature"`
	Arguments  []sourceGraphQLArgument `json:"arguments"`
	ReturnType sourceGraphQLTypeRef    `json:"return_type"`
	Deprecated bool                    `json:"deprecated"`
	Preview    bool                    `json:"preview"`
}

type sourceGraphQLNamedField struct {
	Name      string                  `json:"name"`
	Type      sourceGraphQLTypeRef    `json:"type"`
	Arguments []sourceGraphQLArgument `json:"arguments,omitempty"`
}

type sourceGraphQLNamedType struct {
	Name          string                    `json:"name"`
	Fields        []sourceGraphQLNamedField `json:"fields,omitempty"`
	Interfaces    []string                  `json:"interfaces,omitempty"`
	PossibleTypes []string                  `json:"possible_types,omitempty"`
	Values        []string                  `json:"values,omitempty"`
}

type sourceGraphQLTypeSystem struct {
	Enums        []sourceGraphQLNamedType `json:"enums"`
	InputObjects []sourceGraphQLNamedType `json:"input_objects"`
	Interfaces   []sourceGraphQLNamedType `json:"interfaces"`
	Objects      []sourceGraphQLNamedType `json:"objects"`
	Scalars      []string                 `json:"scalars"`
	Unions       []sourceGraphQLNamedType `json:"unions"`
}

type sourceImportGraphQL struct {
	sourceImportArtifact
	ProjectionSHA256 string                  `json:"projection_sha256,omitempty"`
	ProjectionBytes  int64                   `json:"projection_bytes,omitempty"`
	QueryFields      []sourceGraphQLField    `json:"query_fields"`
	MutationFields   []sourceGraphQLField    `json:"mutation_fields"`
	TypeSystem       sourceGraphQLTypeSystem `json:"type_system"`
}

type sourceImportCounts struct {
	REST            int `json:"rest"`
	GraphQLQuery    int `json:"graphql_query"`
	GraphQLMutation int `json:"graphql_mutation"`
	Total           int `json:"total"`
}

type sourceImportLock struct {
	SchemaVersion int                 `json:"schema_version"`
	Connector     string              `json:"connector"`
	CapturedAt    string              `json:"captured_at,omitempty"`
	Rest          sourceImportREST    `json:"rest"`
	GraphQL       sourceImportGraphQL `json:"graphql,omitempty"`
	Counts        sourceImportCounts  `json:"counts,omitempty"`
}

// sourceImportLockLegacy and sourceImportLockV3 keep strict wire decoding
// versioned. In particular, a future schema must not inherit v2 semantics just
// because its version number happens to be greater than two.
type sourceImportLockLegacy struct {
	SchemaVersion int                 `json:"schema_version"`
	Connector     string              `json:"connector"`
	CapturedAt    string              `json:"captured_at,omitempty"`
	Rest          sourceImportREST    `json:"rest"`
	GraphQL       sourceImportGraphQL `json:"graphql,omitempty"`
	Counts        sourceImportCounts  `json:"counts,omitempty"`
}

type sourceImportRESTV3 struct {
	Retrieval       string                     `json:"retrieval"`
	OpenAPIVersions []string                   `json:"openapi"`
	SourceDocuments []sourceImportRESTDocument `json:"source_documents"`
}

type sourceImportLockV3 struct {
	SchemaVersion int                 `json:"schema_version"`
	Connector     string              `json:"connector"`
	CapturedAt    string              `json:"captured_at,omitempty"`
	Rest          sourceImportRESTV3  `json:"rest"`
	GraphQL       sourceImportGraphQL `json:"graphql,omitempty"`
	Counts        sourceImportCounts  `json:"counts,omitempty"`
}

func (lock *sourceImportLock) UnmarshalJSON(raw []byte) error {
	var header struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return err
	}
	switch header.SchemaVersion {
	case 1, 2:
		var legacy sourceImportLockLegacy
		if err := decodeSourceStrictJSON(raw, &legacy); err != nil {
			return err
		}
		*lock = sourceImportLock(legacy)
		return nil
	case 3:
		var v3 sourceImportLockV3
		if err := decodeSourceStrictJSON(raw, &v3); err != nil {
			return err
		}
		*lock = sourceImportLock{
			SchemaVersion: v3.SchemaVersion,
			Connector:     v3.Connector,
			CapturedAt:    v3.CapturedAt,
			Rest: sourceImportREST{
				Retrieval:       v3.Rest.Retrieval,
				OpenAPIVersions: v3.Rest.OpenAPIVersions,
				SourceDocuments: v3.Rest.SourceDocuments,
			},
			GraphQL: v3.GraphQL,
			Counts:  v3.Counts,
		}
		return nil
	default:
		return fmt.Errorf("source lock has unsupported schema version %d", header.SchemaVersion)
	}
}

type sourceImportSource struct {
	URL                 string `json:"url"`
	SHA256              string `json:"sha256"`
	Bytes               int64  `json:"bytes"`
	Location            string `json:"location"`
	Form                string `json:"form"`
	Version             string `json:"version"`
	DocumentID          string `json:"document_id,omitempty"`
	PublishedURL        string `json:"published_url,omitempty"`
	PublishedCaptureURL string `json:"published_capture_url,omitempty"`
	PublishedSHA256     string `json:"published_sha256,omitempty"`
	PublishedBytes      int64  `json:"published_bytes,omitempty"`
	PublishedAdapter    string `json:"published_adapter,omitempty"`
}

type sourceParameterDescriptor struct {
	Name              string                        `json:"name"`
	Required          bool                          `json:"required"`
	Schema            any                           `json:"schema,omitempty"`
	Content           any                           `json:"content,omitempty"`
	Wire              sourceParameterWireDescriptor `json:"wire"`
	ExecutionEnvelope *sourceExecutionEnvelope      `json:"execution_envelope,omitempty"`
}

type sourceExecutionEnvelope struct {
	PolicyVersion  string                 `json:"policy_version"`
	Origin         string                 `json:"origin"`
	SourceLocation string                 `json:"source_location"`
	Limits         []sourceExecutionLimit `json:"limits"`
}

type sourceExecutionLimit struct {
	Kind        string `json:"kind"`
	Unit        string `json:"unit"`
	Default     int    `json:"default,omitempty"`
	HardCeiling int    `json:"hard_ceiling"`
	Effective   int    `json:"effective"`
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
	Required          bool                     `json:"required"`
	Schema            any                      `json:"schema"`
	Encoding          any                      `json:"encoding,omitempty"`
	ExecutionEnvelope *sourceExecutionEnvelope `json:"execution_envelope,omitempty"`
}

type sourceRequestMediaDescriptor struct {
	MediaType         string                   `json:"media_type"`
	Required          bool                     `json:"required"`
	Schema            any                      `json:"schema"`
	Encoding          any                      `json:"encoding,omitempty"`
	ExecutionEnvelope *sourceExecutionEnvelope `json:"execution_envelope,omitempty"`
}

type sourceRequestDescriptor struct {
	Path      []sourceParameterDescriptor    `json:"path"`
	Query     []sourceParameterDescriptor    `json:"query"`
	Header    []sourceParameterDescriptor    `json:"header"`
	Body      *sourceRequestBodyDescriptor   `json:"body,omitempty"`
	MediaType string                         `json:"media_type,omitempty"`
	Media     []sourceRequestMediaDescriptor `json:"media,omitempty"`
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
	Connector           string                            `json:"connector"`
	Protocol            string                            `json:"protocol,omitempty"`
	SourceID            string                            `json:"source_id"`
	ProviderOperationID string                            `json:"operation_id"`
	Source              sourceImportSource                `json:"source"`
	Method              string                            `json:"method"`
	Path                string                            `json:"path"`
	Request             sourceRequestDescriptor           `json:"request"`
	Responses           []sourceResponseDescriptor        `json:"responses"`
	Output              sourceOutputDescriptor            `json:"output"`
	Pagination          any                               `json:"pagination,omitempty"`
	ByteLimits          sourceByteLimits                  `json:"byte_limits"`
	AuthScopes          sourceAuthDescriptor              `json:"auth_scopes"`
	Servers             sourceServerOverrides             `json:"servers"`
	Runtime             sourceRuntimeReachability         `json:"runtime"`
	GraphQL             *sourceGraphQLOperationDescriptor `json:"graphql,omitempty"`
}

type sourceGraphQLOperationDescriptor struct {
	Root       string                  `json:"root"`
	Name       string                  `json:"name"`
	Line       int                     `json:"line"`
	Signature  string                  `json:"signature"`
	Arguments  []sourceGraphQLArgument `json:"arguments"`
	ReturnType sourceGraphQLTypeRef    `json:"return_type"`
	Deprecated bool                    `json:"deprecated"`
	Preview    bool                    `json:"preview"`
}

type sourceGraphQLSchemaDescriptor struct {
	Connector  string                  `json:"connector"`
	Source     sourceImportSource      `json:"source"`
	TypeSystem sourceGraphQLTypeSystem `json:"type_system"`
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
	DescriptorSchemaVersion int                             `json:"-"`
	Operations              []sourceOperationDescriptor     `json:"operations"`
	InboundEvents           []sourceInboundEventDescriptor  `json:"inbound_events,omitempty"`
	Extensions              []sourceExtensionDescriptor     `json:"extensions,omitempty"`
	GraphQLSchemas          []sourceGraphQLSchemaDescriptor `json:"graphql_schemas,omitempty"`
	Gaps                    []sourceContractGap             `json:"-"`
}

type sourceImportDescriptorDocument struct {
	SchemaVersion  int                             `json:"schema_version"`
	Operations     []sourceOperationDescriptor     `json:"operations"`
	InboundEvents  []sourceInboundEventDescriptor  `json:"inbound_events,omitempty"`
	Extensions     []sourceExtensionDescriptor     `json:"extensions,omitempty"`
	GraphQLSchemas []sourceGraphQLSchemaDescriptor `json:"graphql_schemas,omitempty"`
	MergeBlocked   bool                            `json:"merge_blocked"`
	Gaps           []sourceContractGap             `json:"gaps,omitempty"`
}

type sourceImportFetcher interface {
	Fetch(context.Context, string) ([]byte, error)
}

type sourceImportFetchFunc func(context.Context, string) ([]byte, error)

func (f sourceImportFetchFunc) Fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	return f(ctx, sourceURL)
}

// sourceImportVerifiedArtifactFetcher can validate a cache against the
// immutable lock before returning bytes to the importer. The legacy Fetch
// method remains the narrow seam used by hermetic source-import tests.
type sourceImportVerifiedArtifactFetcher interface {
	FetchArtifact(context.Context, sourceImportArtifact) ([]byte, error)
}

func fetchSourceImportArtifact(ctx context.Context, fetcher sourceImportFetcher, artifact sourceImportArtifact) ([]byte, error) {
	if verified, ok := fetcher.(sourceImportVerifiedArtifactFetcher); ok {
		return verified.FetchArtifact(ctx, artifact)
	}
	return fetcher.Fetch(ctx, artifact.SourceURL)
}

func parseSourceImportLock(raw []byte, expectedConnector string) (sourceImportLock, error) {
	var lock sourceImportLock
	if err := decodeSourceStrictJSON(raw, &lock); err != nil {
		return sourceImportLock{}, fmt.Errorf("parse source lock: %w", err)
	}
	if lock.SchemaVersion != 1 && lock.SchemaVersion != 2 && lock.SchemaVersion != 3 {
		return sourceImportLock{}, fmt.Errorf("source lock has unsupported schema version %d", lock.SchemaVersion)
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
	if lock.SchemaVersion < 3 {
		if lock.Rest.IdentityQuery || lock.GraphQL.IdentityQuery {
			return sourceImportLock{}, fmt.Errorf("source lock identity query declaration requires a v3 REST source document")
		}
		if err := validateSourceImportArtifact(lock.Rest.sourceImportArtifact); err != nil {
			return sourceImportLock{}, err
		}
	}
	if err := validateSourceImportLockInventory(lock); err != nil {
		return sourceImportLock{}, err
	}
	return lock, nil
}

func decodeSourceStrictJSON(raw []byte, target any) error {
	validator := json.NewDecoder(bytes.NewReader(raw))
	validator.UseNumber()
	if err := validateSourceJSONValue(validator, ""); err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return sourceJSONEOF(decoder)
}

func validateSourceImportLockInventory(lock sourceImportLock) error {
	if lock.SchemaVersion == 3 {
		return validateSourceImportV3LockInventory(lock)
	}
	restCount := len(lock.Rest.Operations)
	if lock.Counts.REST != restCount {
		return fmt.Errorf("source lock REST count %d does not match %d operations", lock.Counts.REST, restCount)
	}
	graphqlCount := len(lock.GraphQL.QueryFields) + len(lock.GraphQL.MutationFields)
	if lock.Counts.Total != lock.Counts.REST+graphqlCount {
		return fmt.Errorf("source lock total count does not match REST and GraphQL inventories")
	}
	if restCount > 0 {
		seen := map[string]bool{}
		for _, operation := range lock.Rest.Operations {
			if operation.ID == "" || operation.Protocol != "rest" || operation.Method == "" || operation.Path == "" || operation.SourceLocation == "" {
				return fmt.Errorf("source lock has incomplete REST operation identity")
			}
			if seen[operation.ID] {
				return fmt.Errorf("source lock duplicates REST operation identity %q", operation.ID)
			}
			seen[operation.ID] = true
		}
	}
	if graphqlCount == 0 {
		if lock.Counts.GraphQLQuery != 0 || lock.Counts.GraphQLMutation != 0 {
			return fmt.Errorf("source lock GraphQL counts require a GraphQL inventory")
		}
		return nil
	}
	if err := validateSourceImportArtifact(lock.GraphQL.sourceImportArtifact); err != nil {
		return fmt.Errorf("source lock has invalid GraphQL artifact: %w", err)
	}
	if lock.Counts.GraphQLQuery != len(lock.GraphQL.QueryFields) || lock.Counts.GraphQLMutation != len(lock.GraphQL.MutationFields) {
		return fmt.Errorf("source lock GraphQL counts do not match root fields")
	}
	seen := map[string]bool{}
	for _, group := range []struct {
		root   string
		fields []sourceGraphQLField
	}{{"Query", lock.GraphQL.QueryFields}, {"Mutation", lock.GraphQL.MutationFields}} {
		for _, field := range group.fields {
			identity := group.root + "." + field.Name
			if field.Root != group.root || field.Name == "" || field.Line <= 0 || field.Signature == "" {
				return fmt.Errorf("source lock has incomplete GraphQL root identity %q", identity)
			}
			if seen[identity] {
				return fmt.Errorf("source lock duplicates GraphQL root identity %q", identity)
			}
			seen[identity] = true
		}
	}
	return nil
}

func validateSourceImportV3LockInventory(lock sourceImportLock) error {
	if lock.Rest.Retrieval == "" || lock.Rest.Retrieval != strings.TrimSpace(lock.Rest.Retrieval) || len(lock.Rest.Retrieval) > 1024 || strings.ContainsAny(lock.Rest.Retrieval, "\r\n") {
		return fmt.Errorf("source lock has invalid v3 REST retrieval metadata")
	}
	if len(lock.Rest.SourceDocuments) == 0 {
		return fmt.Errorf("source lock has no v3 REST source documents")
	}
	if len(lock.Rest.SourceDocuments) > defaultSourceImportDocuments {
		return fmt.Errorf("source lock v3 document count exceeds %d", defaultSourceImportDocuments)
	}
	if len(lock.Rest.OpenAPIVersions) == 0 {
		return fmt.Errorf("source lock has no v3 REST OpenAPI versions")
	}
	if !sort.StringsAreSorted(lock.Rest.OpenAPIVersions) {
		return fmt.Errorf("source lock v3 REST OpenAPI versions are not sorted")
	}
	versions := make(map[string]bool, len(lock.Rest.OpenAPIVersions))
	for _, version := range lock.Rest.OpenAPIVersions {
		major, minor, ok := sourceOpenAPIMajorMinor(version)
		if !ok || major != 3 || (minor != 0 && minor != 1) || versions[version] {
			return fmt.Errorf("source lock has invalid or duplicate v3 REST OpenAPI version %q", version)
		}
		versions[version] = true
	}

	seenDocuments := make(map[string]bool, len(lock.Rest.SourceDocuments))
	seenOperations := map[string]bool{}
	seenRoutes := map[string]string{}
	restCount := 0
	for index, document := range lock.Rest.SourceDocuments {
		if document.ID == "" || document.ID != strings.ToLower(document.ID) || document.ID != strings.TrimSpace(document.ID) || !sourceImportDocumentID(document.ID) {
			return fmt.Errorf("source lock has invalid v3 REST document ID %q", document.ID)
		}
		if seenDocuments[document.ID] {
			return fmt.Errorf("source lock duplicates v3 REST document ID %q", document.ID)
		}
		if index > 0 && lock.Rest.SourceDocuments[index-1].ID >= document.ID {
			return fmt.Errorf("source lock v3 REST source documents are not sorted")
		}
		seenDocuments[document.ID] = true
		if err := validateSourceImportArtifact(document.Artifact); err != nil {
			return fmt.Errorf("source lock v3 REST document %q has invalid artifact: %w", document.ID, err)
		}
		if document.Artifact.OpenAPI == "" || !versions[document.Artifact.OpenAPI] {
			return fmt.Errorf("source lock v3 REST document %q has an OpenAPI version outside the aggregate inventory", document.ID)
		}
		if err := validateSourceImportPublishedSource(document.PublishedSource); err != nil {
			return fmt.Errorf("source lock v3 REST document %q has invalid published source: %w", document.ID, err)
		}
		if len(document.Operations) == 0 {
			return fmt.Errorf("source lock v3 REST document %q has no operations", document.ID)
		}
		for _, operation := range document.Operations {
			restCount++
			if operation.ID == "" || operation.Protocol != "rest" || operation.Method == "" || operation.Path == "" || operation.SourceLocation == "" {
				return fmt.Errorf("source lock has incomplete v3 REST operation identity")
			}
			if seenOperations[operation.ID] {
				return fmt.Errorf("source lock duplicates v3 REST operation identity %q", operation.ID)
			}
			if err := validateSourceImportPath(operation.Path); err != nil {
				return fmt.Errorf("source lock v3 REST operation %q has invalid path", operation.ID)
			}
			route := strings.ToUpper(operation.Method) + "\x00" + operation.Path
			if existing, exists := seenRoutes[route]; exists {
				return fmt.Errorf("source lock v3 REST route %s %s occurs in both %q and %q", operation.Method, operation.Path, existing, document.ID)
			}
			seenOperations[operation.ID] = true
			seenRoutes[route] = document.ID
		}
	}
	if lock.Counts.REST != restCount || lock.Counts.Total != restCount+len(lock.GraphQL.QueryFields)+len(lock.GraphQL.MutationFields) {
		return fmt.Errorf("source lock v3 counts do not match document inventories")
	}
	return validateSourceImportGraphQLInventory(lock)
}

func sourceImportDocumentID(value string) bool {
	for index, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') || (character == '-' && index > 0) {
			continue
		}
		return false
	}
	return true
}

func validateSourceImportGraphQLInventory(lock sourceImportLock) error {
	if lock.GraphQL.IdentityQuery {
		return fmt.Errorf("source lock identity query declaration requires a v3 REST source document")
	}
	graphqlCount := len(lock.GraphQL.QueryFields) + len(lock.GraphQL.MutationFields)
	if graphqlCount == 0 {
		if lock.Counts.GraphQLQuery != 0 || lock.Counts.GraphQLMutation != 0 {
			return fmt.Errorf("source lock GraphQL counts require a GraphQL inventory")
		}
		return nil
	}
	if err := validateSourceImportArtifact(lock.GraphQL.sourceImportArtifact); err != nil {
		return fmt.Errorf("source lock has invalid GraphQL artifact: %w", err)
	}
	if lock.Counts.GraphQLQuery != len(lock.GraphQL.QueryFields) || lock.Counts.GraphQLMutation != len(lock.GraphQL.MutationFields) {
		return fmt.Errorf("source lock GraphQL counts do not match root fields")
	}
	seen := map[string]bool{}
	for _, group := range []struct {
		root   string
		fields []sourceGraphQLField
	}{{"Query", lock.GraphQL.QueryFields}, {"Mutation", lock.GraphQL.MutationFields}} {
		for _, field := range group.fields {
			identity := group.root + "." + field.Name
			if field.Root != group.root || field.Name == "" || field.Line <= 0 || field.Signature == "" {
				return fmt.Errorf("source lock has incomplete GraphQL root identity %q", identity)
			}
			if seen[identity] {
				return fmt.Errorf("source lock duplicates GraphQL root identity %q", identity)
			}
			seen[identity] = true
		}
	}
	return nil
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
	policy := batchArtifactURLPolicy{allowIdentityQuery: artifact.IdentityQuery}
	parsed, err := parseBatchArtifactURLWithPolicy(artifact.SourceURL, policy)
	if err != nil {
		return fmt.Errorf("source lock has invalid public artifact URL: %w", err)
	}
	if artifact.IdentityQuery {
		if err := validateSourceImportIdentityArtifactQuery(parsed); err != nil {
			return fmt.Errorf("source lock has invalid identity artifact query: %w", err)
		}
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

const (
	maxSourceImportPublishedQueryBytes = 1024
	maxSourceImportPublishedQueryKeys  = 16
	maxSourceImportPublishedQueryKey   = 64
	maxSourceImportPublishedQueryValue = 256
)

func validateSourceImportPublishedSource(source sourceImportPublishedSource) error {
	if err := validateSourceImportPublishedURL(source.SourceURL); err != nil {
		return err
	}
	if _, err := parseBatchArtifactURL(source.CaptureURL); err != nil {
		return fmt.Errorf("capture URL: %w", err)
	}
	if source.Bytes <= 0 {
		return fmt.Errorf("published source has invalid byte count")
	}
	if len(source.SHA256) != sha256.Size*2 {
		return fmt.Errorf("published source has invalid SHA-256")
	}
	if _, err := hex.DecodeString(source.SHA256); err != nil {
		return fmt.Errorf("published source has invalid SHA-256: %w", err)
	}
	if source.Adapter == "" || source.Adapter != strings.TrimSpace(source.Adapter) || len(source.Adapter) > 128 || strings.ContainsAny(source.Adapter, "\r\n") {
		return fmt.Errorf("published source has invalid adapter")
	}
	return nil
}

// validateSourceImportPublishedURL permits a bounded query only for a v3
// provider citation. Unlike an artifact URL it is never given to Fetch, so the
// query cannot become request authority or credential transport.
func validateSourceImportPublishedURL(raw string) error {
	if raw == "" || raw != strings.TrimSpace(raw) || strings.ContainsAny(raw, "\r\n") {
		return fmt.Errorf("published source URL must be a non-empty absolute HTTPS URL")
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Scheme != "https" || parsed.Host == "" {
		return fmt.Errorf("published source URL must be an absolute HTTPS URL")
	}
	if parsed.User != nil {
		return fmt.Errorf("published source URL must not include userinfo")
	}
	if parsed.Fragment != "" || strings.Contains(raw, "#") {
		return fmt.Errorf("published source URL must not include a fragment")
	}
	artifactURL := *parsed
	artifactURL.RawQuery = ""
	artifactURL.ForceQuery = false
	if _, err := parseBatchArtifactURL(artifactURL.String()); err != nil {
		return fmt.Errorf("published source URL: %w", err)
	}
	return validateSourceImportBoundedQuery(parsed, "published source URL", false)
}

func validateSourceImportIdentityArtifactQuery(parsed *url.URL) error {
	return validateSourceImportBoundedQuery(parsed, "identity artifact query", true)
}

func validateSourceImportBoundedQuery(parsed *url.URL, subject string, requireQuery bool) error {
	if parsed.RawQuery == "" && !parsed.ForceQuery {
		if requireQuery {
			return fmt.Errorf("%s is missing", subject)
		}
		return nil
	}
	if parsed.ForceQuery || len(parsed.RawQuery) > maxSourceImportPublishedQueryBytes {
		return fmt.Errorf("%s exceeds the citation policy", subject)
	}
	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil || len(query) == 0 || len(query) > maxSourceImportPublishedQueryKeys {
		return fmt.Errorf("%s violates the citation policy", subject)
	}
	for key, values := range query {
		lowerKey := strings.ToLower(key)
		if key == "" || len(key) > maxSourceImportPublishedQueryKey || len(values) != 1 || len(values[0]) > maxSourceImportPublishedQueryValue || strings.ContainsAny(key, "\r\n") || strings.ContainsAny(values[0], "\r\n") || sourceImportCredentialLikeQueryKey(lowerKey) {
			return fmt.Errorf("%s violates the citation policy", subject)
		}
	}
	return nil
}

func sourceImportCredentialLikeQueryKey(key string) bool {
	key = strings.ReplaceAll(key, "-", "_")
	for _, prohibited := range []string{"token", "secret", "password", "credential", "authorization", "api_key", "apikey", "signature", "sig", "key"} {
		if key == prohibited || strings.HasSuffix(key, "_"+prohibited) || strings.HasPrefix(key, prohibited+"_") {
			return true
		}
	}
	return false
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
	budget := sourceImportBudget{
		limit:  sourceResolvedDescriptorLimit(limits),
		counts: &sourceImportCountBudget{limit: limits.MaxOperations},
	}
	for _, lock := range orderedLocks {
		if seenConnectors[lock.Connector] {
			return sourceImportResult{}, fmt.Errorf("duplicate source-lock connector %q", lock.Connector)
		}
		seenConnectors[lock.Connector] = true
		imported, err := importSourceLockResultWithBudget(ctx, lock, fetcher, limits, &budget)
		if err != nil {
			return sourceImportResult{}, err
		}
		result.Operations = append(result.Operations, imported.Operations...)
		result.InboundEvents = append(result.InboundEvents, imported.InboundEvents...)
		result.Extensions = append(result.Extensions, imported.Extensions...)
		result.GraphQLSchemas = append(result.GraphQLSchemas, imported.GraphQLSchemas...)
		result.Gaps = append(result.Gaps, imported.Gaps...)
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
	if limits.MaxArtifactBytes <= 0 || limits.MaxSchemaBytes <= 0 || limits.MaxOperations <= 0 || limits.MaxReferences <= 0 || limits.MaxReferenceDepth <= 0 || limits.MaxDocuments < 0 || limits.MaxTotalArtifactBytes < 0 || limits.MaxResolvedDescriptorBytes < 0 {
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
	if lock.SchemaVersion != 1 && lock.SchemaVersion != 2 && lock.SchemaVersion != 3 {
		return sourceImportResult{}, fmt.Errorf("source lock has unsupported schema version %d", lock.SchemaVersion)
	}
	if err := validateSourceImportConnector(lock.Connector); err != nil {
		return sourceImportResult{}, err
	}
	if lock.SchemaVersion == 3 {
		return importSourceLockResultV3(ctx, lock, fetcher, limits, budget)
	}
	if err := validateSourceImportArtifact(lock.Rest.sourceImportArtifact); err != nil {
		return sourceImportResult{}, err
	}
	// Version 2 locks are provider-authoritative inventories. They may contain
	// open schemas or encodings that the current runtime cannot safely expose;
	// import retains those contracts and emits source-bound gaps instead of
	// making the provider operation disappear from the canonical descriptor.
	if lock.SchemaVersion >= 2 {
		limits.AllowSourceContractGaps = true
	}
	if lock.Rest.Bytes > limits.MaxArtifactBytes {
		return sourceImportResult{}, fmt.Errorf("artifact byte limit exceeded by source lock")
	}
	raw, err := fetchSourceImportArtifact(ctx, fetcher, lock.Rest.sourceImportArtifact)
	if err != nil {
		return sourceImportResult{}, fmt.Errorf("fetch locked source artifact: %w", err)
	}
	if int64(len(raw)) > limits.MaxArtifactBytes {
		return sourceImportResult{}, fmt.Errorf("artifact byte limit exceeded")
	}
	if err := validateSourceImportArtifactBytes(raw, lock.Rest.sourceImportArtifact); err != nil {
		return sourceImportResult{}, err
	}
	doc, form, err := parseSourceImportDocument(raw)
	if err != nil {
		return sourceImportResult{}, err
	}
	if err := validateSourceImportArtifactForm(lock.Rest.sourceImportArtifact, form); err != nil {
		return sourceImportResult{}, err
	}
	resolver := sourceReferenceResolver{root: doc, limits: limits, form: form}
	result, err := importSourceDocumentResult(sourceImportDocumentContext{Lock: lock, Artifact: lock.Rest.sourceImportArtifact}, doc, form, &resolver, limits, budget)
	if err != nil {
		return sourceImportResult{}, err
	}
	if err := validateLockedRESTProjection(lock, result.Operations); err != nil {
		return sourceImportResult{}, err
	}
	if err := appendLockedGraphQLProjection(ctx, lock, fetcher, limits, budget, &result); err != nil {
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

type sourceImportDocumentContext struct {
	Lock     sourceImportLock
	Artifact sourceImportArtifact
	Document *sourceImportRESTDocument
}

func (context sourceImportDocumentContext) lockedRESTOperation(method, path string) (sourceImportRESTOperation, bool) {
	if context.Document == nil {
		return sourceImportRESTOperation{}, false
	}
	for _, operation := range context.Document.Operations {
		if strings.EqualFold(operation.Method, method) && operation.Path == path {
			return operation, true
		}
	}
	return sourceImportRESTOperation{}, false
}

func importSourceLockResultV3(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits, budget *sourceImportBudget) (sourceImportResult, error) {
	if err := validateSourceImportLockInventory(lock); err != nil {
		return sourceImportResult{}, err
	}
	if len(lock.Rest.SourceDocuments) > sourceImportDocumentLimit(limits) {
		return sourceImportResult{}, fmt.Errorf("source document count limit exceeded")
	}
	remainingArtifactBytes := sourceImportTotalArtifactLimit(limits)
	for _, document := range lock.Rest.SourceDocuments {
		if document.Artifact.Bytes > remainingArtifactBytes {
			return sourceImportResult{}, fmt.Errorf("source artifact corpus byte limit exceeded by document %q", document.ID)
		}
		remainingArtifactBytes -= document.Artifact.Bytes
	}
	corpusContext, cancel := context.WithTimeout(ctx, defaultSourceImportCorpusTimeout)
	defer cancel()
	rawDocuments, err := fetchSourceImportV3Documents(corpusContext, lock.Rest.SourceDocuments, fetcher)
	if err != nil {
		return sourceImportResult{}, err
	}
	// Version 3 is the immutable document-owned contract and therefore the
	// first descriptor version that can make PM execution policy explicit
	// without rewriting legacy canonical descriptor bytes.
	limits.UseExecutionEnvelopes = true
	result := sourceImportResult{DescriptorSchemaVersion: 3, Operations: []sourceOperationDescriptor{}}
	for _, document := range lock.Rest.SourceDocuments {
		raw := rawDocuments[document.ID]
		doc, form, err := parseSourceImportDocument(raw)
		if err != nil {
			return sourceImportResult{}, fmt.Errorf("parse source document %q: %w", document.ID, err)
		}
		if err := validateSourceImportArtifactForm(document.Artifact, form); err != nil {
			return sourceImportResult{}, fmt.Errorf("validate source document %q form: %w", document.ID, err)
		}
		resolver := sourceReferenceResolver{root: doc, limits: limits, form: form}
		documentContext := sourceImportDocumentContext{Lock: lock, Artifact: document.Artifact, Document: &document}
		imported, err := importSourceDocumentResult(documentContext, doc, form, &resolver, limits, budget)
		if err != nil {
			return sourceImportResult{}, fmt.Errorf("import source document %q: %w", document.ID, err)
		}
		if err := validateLockedRESTDocumentProjection(document, imported.Operations); err != nil {
			return sourceImportResult{}, err
		}
		result.Operations = append(result.Operations, imported.Operations...)
		result.InboundEvents = append(result.InboundEvents, imported.InboundEvents...)
		result.Extensions = append(result.Extensions, imported.Extensions...)
		result.Gaps = append(result.Gaps, imported.Gaps...)
	}
	if err := appendLockedGraphQLProjection(ctx, lock, fetcher, limits, budget, &result); err != nil {
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

func sourceImportDocumentLimit(limits sourceImportLimits) int {
	if limits.MaxDocuments == 0 {
		return defaultSourceImportDocuments
	}
	return limits.MaxDocuments
}

func sourceImportTotalArtifactLimit(limits sourceImportLimits) int64 {
	if limits.MaxTotalArtifactBytes == 0 {
		return defaultSourceImportTotalArtifactBytes
	}
	return limits.MaxTotalArtifactBytes
}

type sourceImportArtifactFetchResult struct {
	raw []byte
	err error
}

// sourceImportArtifactDeduplicator provides corpus-local single-flight by the
// content-addressed digest. It protects both plain hermetic fetchers and the
// on-disk cache fetcher from duplicate-digest races.
type sourceImportArtifactDeduplicator struct {
	fetcher  sourceImportFetcher
	mu       sync.Mutex
	results  map[string]sourceImportArtifactFetchResult
	inFlight map[string]chan struct{}
}

func newSourceImportArtifactDeduplicator(fetcher sourceImportFetcher) *sourceImportArtifactDeduplicator {
	return &sourceImportArtifactDeduplicator{
		fetcher:  fetcher,
		results:  map[string]sourceImportArtifactFetchResult{},
		inFlight: map[string]chan struct{}{},
	}
}

func (deduplicator *sourceImportArtifactDeduplicator) fetch(ctx context.Context, artifact sourceImportArtifact) ([]byte, error) {
	key := strings.ToLower(artifact.SHA256)
	deduplicator.mu.Lock()
	if result, ok := deduplicator.results[key]; ok {
		deduplicator.mu.Unlock()
		if result.err != nil {
			return nil, result.err
		}
		if err := validateSourceImportArtifactBytes(result.raw, artifact); err != nil {
			return nil, err
		}
		return result.raw, nil
	}
	if done, waiting := deduplicator.inFlight[key]; waiting {
		deduplicator.mu.Unlock()
		select {
		case <-done:
			return deduplicator.fetch(ctx, artifact)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	done := make(chan struct{})
	deduplicator.inFlight[key] = done
	deduplicator.mu.Unlock()

	raw, err := fetchSourceImportArtifact(ctx, deduplicator.fetcher, artifact)
	if err == nil {
		err = validateSourceImportArtifactBytes(raw, artifact)
	}
	deduplicator.mu.Lock()
	deduplicator.results[key] = sourceImportArtifactFetchResult{raw: raw, err: err}
	delete(deduplicator.inFlight, key)
	close(done)
	deduplicator.mu.Unlock()
	return raw, err
}

func fetchSourceImportV3Documents(ctx context.Context, documents []sourceImportRESTDocument, fetcher sourceImportFetcher) (map[string][]byte, error) {
	type documentResult struct {
		raw []byte
		err error
	}
	deduplicator := newSourceImportArtifactDeduplicator(fetcher)
	results := make([]documentResult, len(documents))
	jobs := make(chan int)
	workers := defaultSourceImportFetchWorkers
	if workers > len(documents) {
		workers = len(documents)
	}
	var workersDone sync.WaitGroup
	for range workers {
		workersDone.Add(1)
		go func() {
			defer workersDone.Done()
			for index := range jobs {
				results[index].raw, results[index].err = deduplicator.fetch(ctx, documents[index].Artifact)
			}
		}()
	}
	for index := range documents {
		jobs <- index
	}
	close(jobs)
	workersDone.Wait()
	rawDocuments := make(map[string][]byte, len(documents))
	for index, document := range documents {
		if results[index].err != nil {
			return nil, fmt.Errorf("fetch source document %q: %w", document.ID, results[index].err)
		}
		rawDocuments[document.ID] = results[index].raw
	}
	return rawDocuments, nil
}

func validateLockedRESTProjection(lock sourceImportLock, descriptors []sourceOperationDescriptor) error {
	if len(lock.Rest.Operations) == 0 {
		return nil
	}
	if len(descriptors) != len(lock.Rest.Operations) {
		return fmt.Errorf("source lock REST inventory has %d operations but imported artifact has %d", len(lock.Rest.Operations), len(descriptors))
	}
	actual := make(map[string]sourceOperationDescriptor, len(descriptors))
	for _, descriptor := range descriptors {
		key := strings.ToUpper(descriptor.Method) + "\x00" + descriptor.Path
		if _, exists := actual[key]; exists {
			return fmt.Errorf("imported REST artifact duplicates route %s %s", descriptor.Method, descriptor.Path)
		}
		actual[key] = descriptor
	}
	for _, locked := range lock.Rest.Operations {
		key := strings.ToUpper(locked.Method) + "\x00" + locked.Path
		descriptor, exists := actual[key]
		if !exists {
			return fmt.Errorf("source lock REST identity %q is absent from imported artifact", locked.ID)
		}
		if descriptor.ProviderOperationID != locked.OperationID || descriptor.Source.Location != locked.SourceLocation {
			return fmt.Errorf("source lock REST identity %q disagrees with imported provider operation", locked.ID)
		}
	}
	return nil
}

func validateLockedRESTDocumentProjection(document sourceImportRESTDocument, descriptors []sourceOperationDescriptor) error {
	if len(descriptors) != len(document.Operations) {
		return fmt.Errorf("source document %q inventory has %d operations but imported artifact has %d", document.ID, len(document.Operations), len(descriptors))
	}
	actual := make(map[string]sourceOperationDescriptor, len(descriptors))
	for _, descriptor := range descriptors {
		key := strings.ToUpper(descriptor.Method) + "\x00" + descriptor.Path
		if _, exists := actual[key]; exists {
			return fmt.Errorf("imported source document %q duplicates route %s %s", document.ID, descriptor.Method, descriptor.Path)
		}
		actual[key] = descriptor
	}
	for _, locked := range document.Operations {
		key := strings.ToUpper(locked.Method) + "\x00" + locked.Path
		descriptor, exists := actual[key]
		if !exists {
			return fmt.Errorf("source document %q identity %q is absent from imported artifact", document.ID, locked.ID)
		}
		if descriptor.SourceID != locked.ID || descriptor.ProviderOperationID != locked.OperationID || descriptor.Source.Location != locked.SourceLocation || descriptor.Source.DocumentID != document.ID {
			return fmt.Errorf("source document %q identity %q disagrees with imported provider operation", document.ID, locked.ID)
		}
	}
	return nil
}

func appendLockedGraphQLProjection(ctx context.Context, lock sourceImportLock, fetcher sourceImportFetcher, limits sourceImportLimits, budget *sourceImportBudget, result *sourceImportResult) error {
	fields := append(append([]sourceGraphQLField(nil), lock.GraphQL.QueryFields...), lock.GraphQL.MutationFields...)
	if len(fields) == 0 {
		return nil
	}
	if lock.GraphQL.Bytes > limits.MaxArtifactBytes {
		return fmt.Errorf("GraphQL artifact byte limit exceeded by source lock")
	}
	projection, err := canonicalSourceGraphQLProjection(lock.GraphQL)
	if err != nil {
		return fmt.Errorf("canonicalize embedded GraphQL projection: %w", err)
	}
	projectionDigest := sha256.Sum256(projection)
	if lock.SchemaVersion >= 2 {
		if lock.GraphQL.ProjectionBytes <= 0 || len(lock.GraphQL.ProjectionSHA256) != sha256.Size*2 {
			return fmt.Errorf("GraphQL projection digest and byte count are required for source-lock version %d", lock.SchemaVersion)
		}
		if int64(len(projection)) != lock.GraphQL.ProjectionBytes || !strings.EqualFold(hex.EncodeToString(projectionDigest[:]), lock.GraphQL.ProjectionSHA256) {
			return fmt.Errorf("GraphQL projection drift: embedded type-system bytes do not match projection_bytes/projection_sha256")
		}
	}
	// Version 2 embeds the complete, normalized GraphQL field and type-system
	// projection in the strict checked-in lock. Import from those reviewed bytes
	// instead of requiring an unversioned public URL to continue serving the
	// historical capture. Legacy locks still verify their external artifact.
	if lock.SchemaVersion < 2 {
		raw, err := fetchSourceImportArtifact(ctx, fetcher, lock.GraphQL.sourceImportArtifact)
		if err != nil {
			return fmt.Errorf("fetch locked GraphQL source artifact: %w", err)
		}
		if int64(len(raw)) > limits.MaxArtifactBytes {
			return fmt.Errorf("GraphQL artifact byte limit exceeded")
		}
		if err := validateSourceImportArtifactBytes(raw, lock.GraphQL.sourceImportArtifact); err != nil {
			return fmt.Errorf("locked GraphQL source artifact: %w", err)
		}
	}
	countBudget, err := budget.countBudget(limits)
	if err != nil {
		return err
	}
	if err := countBudget.reserveOperations(len(fields)); err != nil {
		return err
	}
	schema := sourceGraphQLSchemaDescriptor{
		Connector: lock.Connector,
		Source: sourceImportSource{
			URL:      lock.GraphQL.SourceURL,
			SHA256:   strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)),
			Bytes:    firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes),
			Location: "graphql.type_system",
			Form:     "graphql",
			Version:  strconv.Itoa(lock.SchemaVersion),
		},
		TypeSystem: lock.GraphQL.TypeSystem,
	}
	if err := budget.add(schema, "GraphQL schema"); err != nil {
		return err
	}
	result.GraphQLSchemas = append(result.GraphQLSchemas, schema)
	for _, field := range fields {
		kind := "query"
		if field.Root == "Mutation" {
			kind = "mutation"
		}
		location := fmt.Sprintf("graphql.%s_fields[%q]@line:%d", kind, field.Name, field.Line)
		descriptor := sourceOperationDescriptor{
			Connector: lock.Connector,
			Protocol:  "graphql",
			SourceID:  fmt.Sprintf("%s.graphql.%s.%s", lock.Connector, kind, field.Name),
			Source: sourceImportSource{
				URL:      lock.GraphQL.SourceURL,
				SHA256:   strings.ToLower(firstNonEmpty(lock.GraphQL.ProjectionSHA256, lock.GraphQL.SHA256)),
				Bytes:    firstPositiveInt64(lock.GraphQL.ProjectionBytes, lock.GraphQL.Bytes),
				Location: location,
				Form:     "graphql",
				Version:  strconv.Itoa(lock.SchemaVersion),
			},
			Method:     "post",
			Path:       "/graphql",
			Request:    sourceRequestDescriptor{Path: []sourceParameterDescriptor{}, Query: []sourceParameterDescriptor{}, Header: []sourceParameterDescriptor{}},
			Responses:  []sourceResponseDescriptor{},
			Output:     sourceOutputDescriptor{Class: sourceOutputJSON, Success: []sourceOutputVariant{{Status: "200", MediaType: "application/json", Class: sourceOutputJSON}}},
			ByteLimits: sourceByteLimits{Response: defaultSourceImportSchemaBytes},
			AuthScopes: sourceAuthDescriptor{Declared: true},
			Servers:    sourceServerOverrides{Precedence: []string{"graphql_source_lock"}},
			Runtime:    sourceRuntimeReachability{},
			GraphQL: &sourceGraphQLOperationDescriptor{
				Root: field.Root, Name: field.Name, Line: field.Line, Signature: field.Signature,
				Arguments: field.Arguments, ReturnType: field.ReturnType, Deprecated: field.Deprecated, Preview: field.Preview,
			},
		}
		if err := budget.add(descriptor, "GraphQL operation"); err != nil {
			return err
		}
		result.Operations = append(result.Operations, descriptor)
	}
	return nil
}

func firstPositiveInt64(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func canonicalSourceGraphQLProjection(graphql sourceImportGraphQL) ([]byte, error) {
	raw, err := json.Marshal(struct {
		QueryFields    []sourceGraphQLField    `json:"query_fields"`
		MutationFields []sourceGraphQLField    `json:"mutation_fields"`
		TypeSystem     sourceGraphQLTypeSystem `json:"type_system"`
	}{graphql.QueryFields, graphql.MutationFields, graphql.TypeSystem})
	if err != nil {
		return nil, err
	}
	var projection struct {
		QueryFields    []sourceGraphQLField    `json:"query_fields"`
		MutationFields []sourceGraphQLField    `json:"mutation_fields"`
		TypeSystem     sourceGraphQLTypeSystem `json:"type_system"`
	}
	if err := json.Unmarshal(raw, &projection); err != nil {
		return nil, err
	}
	sortFields := func(fields []sourceGraphQLField) {
		sort.Slice(fields, func(i, j int) bool {
			if fields[i].Root != fields[j].Root {
				return fields[i].Root < fields[j].Root
			}
			return fields[i].Name < fields[j].Name
		})
		for index := range fields {
			sort.Slice(fields[index].Arguments, func(i, j int) bool { return fields[index].Arguments[i].Name < fields[index].Arguments[j].Name })
		}
	}
	sortNamed := func(types []sourceGraphQLNamedType) {
		sort.Slice(types, func(i, j int) bool { return types[i].Name < types[j].Name })
		for index := range types {
			sort.Slice(types[index].Fields, func(i, j int) bool { return types[index].Fields[i].Name < types[index].Fields[j].Name })
			for fieldIndex := range types[index].Fields {
				sort.Slice(types[index].Fields[fieldIndex].Arguments, func(i, j int) bool {
					return types[index].Fields[fieldIndex].Arguments[i].Name < types[index].Fields[fieldIndex].Arguments[j].Name
				})
			}
			sort.Strings(types[index].Interfaces)
			sort.Strings(types[index].PossibleTypes)
			sort.Strings(types[index].Values)
		}
	}
	sortFields(projection.QueryFields)
	sortFields(projection.MutationFields)
	sortNamed(projection.TypeSystem.Enums)
	sortNamed(projection.TypeSystem.InputObjects)
	sortNamed(projection.TypeSystem.Interfaces)
	sortNamed(projection.TypeSystem.Objects)
	sort.Strings(projection.TypeSystem.Scalars)
	sortNamed(projection.TypeSystem.Unions)
	return json.Marshal(map[string]any{
		"query_fields":    projection.QueryFields,
		"mutation_fields": projection.MutationFields,
		"type_system":     projection.TypeSystem,
	})
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
	limit  int64
	used   int64
	counts *sourceImportCountBudget
}

type sourceImportCountBudget struct {
	limit         int
	operations    int
	inboundEvents int
}

type sourceResponseExpansionBudget struct {
	limit int64
	used  int64
}

type sourceRetainedExpansionBudget struct {
	limit int64
	used  int64
	label string
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

func (budget *sourceImportBudget) countBudget(limits sourceImportLimits) (*sourceImportCountBudget, error) {
	if budget == nil {
		return nil, fmt.Errorf("source importer has no descriptor budget")
	}
	if budget.counts == nil {
		budget.counts = &sourceImportCountBudget{limit: limits.MaxOperations}
	}
	if budget.counts.limit != limits.MaxOperations {
		return nil, fmt.Errorf("source importer has inconsistent operation count budget")
	}
	return budget.counts, nil
}

func (budget *sourceImportCountBudget) reserveOperations(count int) error {
	if budget == nil || count < 0 || count > budget.limit-budget.operations {
		return fmt.Errorf("operation count limit exceeded")
	}
	budget.operations += count
	return nil
}

func (budget *sourceImportCountBudget) reserveInboundEvents(count int) error {
	if budget == nil || count < 0 || count > budget.limit-budget.inboundEvents {
		return fmt.Errorf("inbound event count limit exceeded")
	}
	budget.inboundEvents += count
	return nil
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

func (budget *sourceRetainedExpansionBudget) check(bytes int64) error {
	if budget == nil || bytes < 0 || bytes > budget.limit || budget.used > budget.limit-bytes {
		return fmt.Errorf("resolved descriptor byte limit exceeded while retaining %s", budget.label)
	}
	return nil
}

func (budget *sourceRetainedExpansionBudget) reserve(bytes int64) error {
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
			key, err := sourceYAMLMappingKey(node.Content[index], pointer)
			if err != nil {
				return err
			}
			childPointer := sourceJSONPointer(pointer, key)
			if seen[key] {
				return fmt.Errorf("duplicate YAML mapping key at %s", childPointer)
			}
			seen[key] = true
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
			key, err := sourceYAMLMappingKey(node.Content[index], pointer)
			if err != nil {
				return nil, err
			}
			childPointer := sourceJSONPointer(pointer, key)
			child, err := sourceYAMLNodeValue(node.Content[index+1], childPointer)
			if err != nil {
				return nil, err
			}
			out[key] = child
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

// sourceYAMLMappingKey turns a YAML scalar mapping key into the JSON object
// member name the strict JSON decoder will receive. YAML permits scalar keys
// that JSON object syntax writes as strings, such as an unquoted OpenAPI
// response code. Standard JSON scalar tags are deliberately accepted:
// strings remain literal, while integers, floats, booleans, and null use the
// yaml.v3-decoded JSON spelling. YAML-only/custom scalar tags and compound
// keys remain unsupported so normalization cannot invent a JSON contract.
func sourceYAMLMappingKey(node *yaml.Node, pointer string) (string, error) {
	if node.Kind != yaml.ScalarNode {
		return "", fmt.Errorf("YAML mapping key at %s must be a scalar", pointer)
	}
	if node.Tag == "!!str" {
		return node.Value, nil
	}
	switch node.Tag {
	case "!!int", "!!float", "!!bool", "!!null":
		var value any
		if err := node.Decode(&value); err != nil {
			return "", fmt.Errorf("decode YAML mapping key at %s: %w", pointer, err)
		}
		canonical, err := json.Marshal(value)
		if err != nil {
			return "", fmt.Errorf("canonicalize YAML mapping key at %s: %w", pointer, err)
		}
		return string(canonical), nil
	default:
		return "", fmt.Errorf("YAML mapping key at %s must use a JSON scalar type", pointer)
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
	root                       map[string]any
	limits                     sourceImportLimits
	form                       sourceDocumentForm
	referenceIndex             *sourceReferenceIndex
	schemaCycles               map[string]struct{}
	references                 int
	expansion                  sourceSchemaExpansionBudget
	responseExpansion          sourceResponseExpansionBudget
	responseScope              *sourceResponseExpansionBudget
	mediaExpansion             sourceRetainedExpansionBudget
	inboundExpansion           sourceRetainedExpansionBudget
	referenceExpansion         sourceRetainedExpansionBudget
	schemaReferenceSiblingGaps []sourceContractGap
}

const (
	sourceRecursiveSchemaFoundation           = "cli-recursive-schema-foundation-r1"
	sourceOpenAPI30ReferenceSiblingFoundation = "cli-openapi30-reference-sibling-foundation-r1"
)

type sourceSchemaReferenceCycleError struct {
	Reference string
}

func (err *sourceSchemaReferenceCycleError) Error() string {
	return fmt.Sprintf("reference cycle at %q", err.Reference)
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
	sourceReferenceOperation   sourceReferenceKind = "operation"
)

type sourceReferenceIndex struct {
	positions                   map[string]sourceReferenceKind
	reachableOperationIDs       map[string]int
	reachableOperationPositions map[string]int
	extensions                  map[string]struct{}
	limits                      sourceImportLimits
	positionBytes               int64
	extensionBytes              int64
}

func (index *sourceReferenceIndex) add(pointer string, kind sourceReferenceKind) error {
	pointer = sourceReferenceIndexPointer(pointer)
	if existing, exists := index.positions[pointer]; exists && existing != kind {
		return fmt.Errorf("ambiguous source grammar position %q is both %s and %s", pointer, existing, kind)
	}
	if _, exists := index.positions[pointer]; exists {
		return nil
	}
	if _, exists := index.extensions[pointer]; exists {
		return fmt.Errorf("ambiguous source grammar position %q is both an extension and %s", pointer, kind)
	}
	bytes := sourceReferenceIndexEntryBytes(pointer)
	if err := index.checkAddition(1, bytes); err != nil {
		return err
	}
	index.positions[pointer] = kind
	index.positionBytes += bytes
	return nil
}

func (index *sourceReferenceIndex) addExtension(pointer string) error {
	pointer = sourceReferenceIndexPointer(pointer)
	if _, exists := index.positions[pointer]; exists {
		return nil
	}
	if _, exists := index.extensions[pointer]; exists {
		return nil
	}
	bytes := sourceReferenceIndexEntryBytes(pointer)
	if err := index.checkAddition(1, bytes); err != nil {
		return err
	}
	index.extensions[pointer] = struct{}{}
	index.extensionBytes += bytes
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
	limit := limits.MaxIndexBytes
	if limit <= 0 {
		limit = 64 << 20
	}
	return limit
}

func (index *sourceReferenceIndex) checkAddition(count int, bytes int64) error {
	if count < 0 || count > sourceSchemaNodeLimit(index.limits)-index.entryCount() {
		return fmt.Errorf("source grammar position limit exceeded")
	}
	limit := sourceReferenceIndexByteLimit(index.limits)
	if bytes < 0 || bytes > limit || index.entryBytes() > limit-bytes {
		return fmt.Errorf("source grammar position byte limit exceeded")
	}
	return nil
}

func (index *sourceReferenceIndex) entryCount() int {
	return len(index.positions) + len(index.extensions)
}

func (index *sourceReferenceIndex) entryBytes() int64 {
	return index.positionBytes + index.extensionBytes
}

func (index *sourceReferenceIndex) preflightObjectExtensions(pointer string, object map[string]any) error {
	for key := range object {
		if strings.HasPrefix(key, "x-") {
			if err := index.addExtension(sourceJSONPointer(pointer, key)); err != nil {
				return err
			}
		}
	}
	return nil
}

func (index *sourceReferenceIndex) preflightEntries(pointer string, entries map[string]any, skipExtensions bool) error {
	if skipExtensions {
		for name := range entries {
			if strings.HasPrefix(name, "x-") {
				if err := index.addExtension(sourceJSONPointer(pointer, name)); err != nil {
					return err
				}
			}
		}
	}
	remainingPositions := sourceSchemaNodeLimit(index.limits) - index.entryCount()
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
	remainingBytes := sourceReferenceIndexByteLimit(index.limits) - index.entryBytes()
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
	if count < 0 || count > sourceSchemaNodeLimit(index.limits)-index.entryCount() {
		return fmt.Errorf("source grammar position limit exceeded")
	}
	remainingBytes := sourceReferenceIndexByteLimit(index.limits) - index.entryBytes()
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
	index := &sourceReferenceIndex{
		positions:                   map[string]sourceReferenceKind{},
		reachableOperationIDs:       map[string]int{},
		reachableOperationPositions: map[string]int{},
		extensions:                  map[string]struct{}{},
		limits:                      limits,
	}
	if err := index.preflightObjectExtensions("", root); err != nil {
		return nil, err
	}
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
	if err := sourceIndexReachableOperations(index, root, form, limits); err != nil {
		return nil, err
	}
	return index, nil
}

func sourceIndexReachableOperations(index *sourceReferenceIndex, root map[string]any, form sourceDocumentForm, limits sourceImportLimits) error {
	rawPaths, declared := root["paths"]
	if !declared {
		return nil
	}
	paths, err := sourceReferenceObject(rawPaths, "paths")
	if err != nil {
		return err
	}
	resolver := sourceReferenceResolver{root: root, limits: limits, form: form, referenceIndex: index}
	for _, path := range sortedSourceMapKeys(paths) {
		if strings.HasPrefix(path, "x-") {
			continue
		}
		pointer := sourceReferenceIndexPointer(sourceJSONPointer(sourceJSONPointer("", "paths"), path))
		if err := resolver.indexReachablePathItemOperations(paths[path], pointer, nil, 0); err != nil {
			return fmt.Errorf("path item %q: %w", path, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) indexReachablePathItemOperations(value any, pointer string, stack map[string]bool, depth int) error {
	pathItem, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("path item must be an object")
	}
	target, _, next, hasReference, err := r.referenceTargetWithCount(pathItem, sourceReferencePathItem, stack, depth, false)
	if err != nil {
		return err
	}
	if hasReference {
		reference, err := sourceReferenceStackAddition(stack, next)
		if err != nil {
			return err
		}
		return r.indexReachablePathItemOperations(target, reference, next, depth+1)
	}
	for _, method := range sourceHTTPMethods {
		rawOperation, declared := pathItem[method]
		if !declared {
			continue
		}
		operation, err := sourceReferenceObject(rawOperation, "operation")
		if err != nil {
			return fmt.Errorf("%s operation: %w", method, err)
		}
		operationPointer := sourceReferenceIndexPointer(sourceJSONPointer(pointer, method))
		if r.referenceIndex.positions[operationPointer] != sourceReferenceOperation {
			return fmt.Errorf("operation %q does not occupy a grammar-defined operation position", operationPointer)
		}
		if r.referenceIndex.reachableOperationPositions[operationPointer] == math.MaxInt {
			return fmt.Errorf("reachable operation occurrence limit exceeded")
		}
		r.referenceIndex.reachableOperationPositions[operationPointer]++
		if rawID, declared := operation["operationId"]; declared {
			if operationID, ok := rawID.(string); ok && operationID != "" {
				if r.referenceIndex.reachableOperationIDs[operationID] == math.MaxInt {
					return fmt.Errorf("reachable operation occurrence limit exceeded")
				}
				r.referenceIndex.reachableOperationIDs[operationID]++
			}
		}
	}
	return nil
}

func sourceIndexOpenAPIComponents(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	components, err := sourceReferenceObject(value, "components")
	if err != nil {
		return err
	}
	if err := index.preflightObjectExtensions(pointer, components); err != nil {
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
	if err := index.preflightObjectExtensions(pointer, pathItem); err != nil {
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
	if err := index.add(pointer, sourceReferenceOperation); err != nil {
		return err
	}
	if err := index.preflightObjectExtensions(pointer, operation); err != nil {
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
	if err := index.preflightObjectExtensions(pointer, parameter); err != nil {
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
	if err := index.preflightObjectExtensions(pointer, requestBody); err != nil {
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
	if err := index.preflightObjectExtensions(pointer, response); err != nil {
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
	if err := index.preflightObjectExtensions(pointer, header); err != nil {
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
		if err := index.preflightObjectExtensions(childPointer, header); err != nil {
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
		if err := index.preflightObjectExtensions(mediaPointer, media); err != nil {
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
			if !sourceRequestFormMediaType(mediaType) {
				return fmt.Errorf("unsupported encoding on request content media %q", mediaType)
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
		if err := index.preflightObjectExtensions(childPointer, encoding); err != nil {
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
	link, err := sourceReferenceObject(value, "link")
	if err != nil {
		return err
	}
	return index.preflightObjectExtensions(pointer, link)
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
	example, err := sourceReferenceObject(value, "example")
	if err != nil {
		return err
	}
	return index.preflightObjectExtensions(pointer, example)
}

func sourceIndexCallback(index *sourceReferenceIndex, value any, pointer string, form sourceDocumentForm) error {
	if err := index.add(pointer, sourceReferenceCallback); err != nil {
		return err
	}
	callback, err := sourceReferenceObject(value, "callback")
	if err != nil {
		return err
	}
	if err := index.preflightObjectExtensions(pointer, callback); err != nil {
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
	scheme, err := sourceReferenceObject(value, "security scheme")
	if err != nil {
		return err
	}
	return index.preflightObjectExtensions(pointer, scheme)
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
	if err := index.preflightObjectExtensions(pointer, object); err != nil {
		return err
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

func sourcePrepareSourceDocument(doc map[string]any, form sourceDocumentForm, limits sourceImportLimits, resolver *sourceReferenceResolver, countBudget *sourceImportCountBudget) error {
	if resolver == nil {
		return fmt.Errorf("source importer has no reference resolver")
	}
	if countBudget == nil {
		return fmt.Errorf("source importer has no count budget")
	}
	index, err := sourceBuildReferenceIndex(doc, form, limits)
	if err != nil {
		return fmt.Errorf("index source grammar positions: %w", err)
	}
	preflight := sourceReferenceResolver{
		root:               doc,
		limits:             limits,
		form:               form,
		referenceIndex:     index,
		schemaCycles:       map[string]struct{}{},
		responseExpansion:  sourceResponseExpansionBudget{limit: sourceResolvedDescriptorLimit(limits)},
		mediaExpansion:     sourceRetainedExpansionBudget{limit: sourceResolvedDescriptorLimit(limits), label: "request media"},
		inboundExpansion:   sourceRetainedExpansionBudget{limit: sourceResolvedDescriptorLimit(limits), label: "inbound event"},
		referenceExpansion: sourceRetainedExpansionBudget{limit: sourceResolvedDescriptorLimit(limits), label: "reference target"},
	}
	if err := preflight.reserveDiscoveredCounts(countBudget); err != nil {
		return fmt.Errorf("reserve source discovery counts: %w", err)
	}
	if err := preflight.preflightDocument(); err != nil {
		return fmt.Errorf("preflight source grammar: %w", err)
	}
	resolver.root = doc
	resolver.limits = limits
	resolver.form = form
	resolver.referenceIndex = index
	resolver.schemaCycles = preflight.schemaCycles
	resolver.schemaReferenceSiblingGaps = preflight.schemaReferenceSiblingGaps
	resolver.references = 0
	resolver.expansion = sourceSchemaExpansionBudget{}
	resolver.responseExpansion = sourceResponseExpansionBudget{limit: sourceResolvedDescriptorLimit(limits)}
	resolver.responseScope = nil
	resolver.mediaExpansion = sourceRetainedExpansionBudget{limit: sourceResolvedDescriptorLimit(limits), label: "request media"}
	resolver.inboundExpansion = sourceRetainedExpansionBudget{limit: sourceResolvedDescriptorLimit(limits), label: "inbound event"}
	resolver.referenceExpansion = sourceRetainedExpansionBudget{limit: sourceResolvedDescriptorLimit(limits), label: "reference target"}
	return nil
}

func (r *sourceReferenceResolver) reserveDiscoveredCounts(countBudget *sourceImportCountBudget) error {
	if r.form.isOpenAPI31() {
		if rawWebhooks, declared := r.root["webhooks"]; declared {
			webhooks, err := sourceReferenceObject(rawWebhooks, "webhooks")
			if err != nil {
				return err
			}
			if err := countBudget.reserveInboundEvents(sourceNonExtensionEntryCount(webhooks)); err != nil {
				return err
			}
		}
	}
	rawPaths, declared := r.root["paths"]
	if !declared {
		return nil
	}
	paths, err := sourceReferenceObject(rawPaths, "paths")
	if err != nil {
		return err
	}
	for _, path := range sortedSourceMapKeys(paths) {
		if strings.HasPrefix(path, "x-") {
			continue
		}
		operations, inboundEvents, err := r.pathItemDiscoveryCounts(paths[path], nil, 0)
		if err != nil {
			return fmt.Errorf("path item %q: %w", path, err)
		}
		if err := countBudget.reserveOperations(operations); err != nil {
			return err
		}
		if err := countBudget.reserveInboundEvents(inboundEvents); err != nil {
			return err
		}
	}
	return nil
}

func (r *sourceReferenceResolver) pathItemDiscoveryCounts(value any, stack map[string]bool, depth int) (int, int, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, 0, fmt.Errorf("path item must be an object")
	}
	target, _, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferencePathItem, stack, depth, false)
	if err != nil {
		return 0, 0, err
	}
	if hasReference {
		return r.pathItemDiscoveryCounts(target, next, depth+1)
	}
	inboundEvents := 0
	if r.form.isOpenAPI() {
		if callbacks, declared := object["callbacks"]; declared {
			count, err := sourceCallbackCount(callbacks)
			if err != nil {
				return 0, 0, err
			}
			inboundEvents += count
		}
	}
	operations := 0
	for _, method := range sourceHTTPMethods {
		rawOperation, declared := object[method]
		if !declared {
			continue
		}
		operation, err := sourceReferenceObject(rawOperation, "operation")
		if err != nil {
			return 0, 0, fmt.Errorf("%s operation: %w", method, err)
		}
		operations++
		if r.form.isOpenAPI() {
			if callbacks, declared := operation["callbacks"]; declared {
				count, err := sourceCallbackCount(callbacks)
				if err != nil {
					return 0, 0, fmt.Errorf("%s operation callbacks: %w", method, err)
				}
				inboundEvents += count
			}
		}
	}
	return operations, inboundEvents, nil
}

func sourceCallbackCount(value any) (int, error) {
	callbacks, err := sourceReferenceObject(value, "callbacks")
	if err != nil {
		return 0, err
	}
	return sourceNonExtensionEntryCount(callbacks), nil
}

func (r *sourceReferenceResolver) preflightDocument() error {
	if rawPaths, declared := r.root["paths"]; declared {
		if err := r.preflightPathItems(rawPaths); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if rawWebhooks, declared := r.root["webhooks"]; declared && r.form.isOpenAPI31() {
			if err := r.preflightInboundPathItems(rawWebhooks); err != nil {
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

func (r *sourceReferenceResolver) preflightInboundPathItems(value any) error {
	items, err := sourceReferenceObject(value, "path items")
	if err != nil {
		return err
	}
	for _, name := range sortedSourceMapKeys(items) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		pathItem, err := r.resolveInboundPathItem(items[name], nil, 0)
		if err != nil {
			return fmt.Errorf("path item %q: %w", name, err)
		}
		if err := r.preflightResolvedPathItem(pathItem); err != nil {
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
	return r.preflightResolvedParameter(parameter)
}

func (r *sourceReferenceResolver) preflightResolvedParameter(parameter map[string]any) error {
	schema, content, err := sourceParameterRepresentation(parameter, r.form)
	if err != nil {
		return err
	}
	if content != nil {
		return validateBoundedParameterContent("source preflight", content, r.form, r.limits)
	}
	schema, err = r.resolveSchema(schema, nil, 0)
	if err != nil {
		return err
	}
	if len(r.schemaCycleReferences(schema)) > 0 {
		return nil
	}
	location, _ := parameter["in"].(string)
	return validateBoundedOperationParameterSchema(schema, r.form, r.limits, location)
}

func (r *sourceReferenceResolver) preflightRequestBody(value any) error {
	body, err := r.resolveRequestBody(value, nil, 0)
	if err != nil {
		return err
	}
	return r.preflightResolvedRequestBody(body)
}

func (r *sourceReferenceResolver) preflightResolvedRequestBody(body map[string]any) error {
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
		if len(r.schemaCycleReferences(schema)) > 0 {
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
	return r.preflightResolvedCallback(callback)
}

func (r *sourceReferenceResolver) preflightResolvedPathItem(pathItem map[string]any) error {
	if parameters, declared := pathItem["parameters"]; declared {
		if err := r.preflightResolvedParameters(parameters); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if callbacks, declared := pathItem["callbacks"]; declared {
			if err := r.preflightResolvedCallbacks(callbacks); err != nil {
				return err
			}
		}
	}
	for _, method := range sourceHTTPMethods {
		rawOperation, declared := pathItem[method]
		if !declared {
			continue
		}
		operation, err := sourceReferenceObject(rawOperation, "operation")
		if err != nil {
			return fmt.Errorf("%s operation: %w", method, err)
		}
		if err := r.preflightResolvedOperation(operation); err != nil {
			return fmt.Errorf("%s operation: %w", method, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightResolvedOperation(operation map[string]any) error {
	if parameters, declared := operation["parameters"]; declared {
		if err := r.preflightResolvedParameters(parameters); err != nil {
			return err
		}
	}
	if r.form.isOpenAPI() {
		if requestBody, declared := operation["requestBody"]; declared {
			body, err := sourceReferenceObject(requestBody, "request body")
			if err != nil {
				return err
			}
			if err := r.preflightResolvedRequestBody(body); err != nil {
				return err
			}
		}
		if callbacks, declared := operation["callbacks"]; declared {
			if err := r.preflightResolvedCallbacks(callbacks); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightResolvedParameters(value any) error {
	parameters, ok := value.([]any)
	if !ok {
		return fmt.Errorf("parameters must be an array")
	}
	for index, rawParameter := range parameters {
		parameter, err := sourceReferenceObject(rawParameter, "parameter")
		if err != nil {
			return fmt.Errorf("parameter %d: %w", index, err)
		}
		if err := r.preflightResolvedParameter(parameter); err != nil {
			return fmt.Errorf("parameter %d: %w", index, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightResolvedCallbacks(value any) error {
	callbacks, err := sourceReferenceObject(value, "callbacks")
	if err != nil {
		return err
	}
	for _, name := range sortedSourceMapKeys(callbacks) {
		if strings.HasPrefix(name, "x-") {
			continue
		}
		callback, err := sourceReferenceObject(callbacks[name], "callback")
		if err != nil {
			return fmt.Errorf("callback %q: %w", name, err)
		}
		if err := r.preflightResolvedCallback(callback); err != nil {
			return fmt.Errorf("callback %q: %w", name, err)
		}
	}
	return nil
}

func (r *sourceReferenceResolver) preflightResolvedCallback(callback map[string]any) error {
	for _, expression := range sortedSourceMapKeys(callback) {
		if strings.HasPrefix(expression, "x-") {
			continue
		}
		pathItem, err := sourceReferenceObject(callback[expression], "path item")
		if err != nil {
			return fmt.Errorf("callback expression %q: %w", expression, err)
		}
		if err := r.preflightResolvedPathItem(pathItem); err != nil {
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
	if err := r.reserveInboundPathItem(value, stack, depth); err != nil {
		return nil, err
	}
	return r.resolveInboundPathItemUnchecked(value, stack, depth)
}

func (r *sourceReferenceResolver) resolveInboundPathItemUnchecked(value any, stack map[string]bool, depth int) (map[string]any, error) {
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
					resolved, err := r.resolveCallbackUnchecked(callbackMap[name], nil, 0)
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

func (r *sourceReferenceResolver) reserveInboundPathItem(value any, stack map[string]bool, depth int) error {
	bytes, err := r.inboundPathItemExpansionBytes(value, stack, depth)
	if err != nil {
		return err
	}
	return r.inboundExpansion.reserve(bytes)
}

func (r *sourceReferenceResolver) reserveInboundCallback(value any, stack map[string]bool, depth int) error {
	bytes, err := r.inboundCallbackExpansionBytes(value, stack, depth)
	if err != nil {
		return err
	}
	return r.inboundExpansion.reserve(bytes)
}

func (r *sourceReferenceResolver) inboundPathItemExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("path item must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferencePathItem, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.inboundPathItemExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	return sourceMapExpandedBytes(object, func(key string, child any) (int64, error) {
		switch {
		case key == "parameters":
			return r.inboundParametersExpansionBytes(child, stack, depth)
		case r.form.isOpenAPI() && key == "callbacks":
			return r.inboundCallbacksExpansionBytes(child, stack, depth)
		case containsSourceString(sourceHTTPMethods, key):
			return r.inboundOperationExpansionBytes(child, stack, depth)
		default:
			return sourceResponseStructuralBytes(child, r.limits)
		}
	})
}

func (r *sourceReferenceResolver) inboundOperationExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	operation, err := sourceReferenceObject(value, "operation")
	if err != nil {
		return 0, err
	}
	return sourceMapExpandedBytes(operation, func(key string, child any) (int64, error) {
		switch {
		case key == "parameters":
			return r.inboundParametersExpansionBytes(child, stack, depth)
		case r.form.isOpenAPI() && key == "requestBody":
			return r.requestBodyExpansionBytes(child, stack, depth)
		case key == "responses":
			return r.inboundResponsesExpansionBytes(child, stack, depth)
		case r.form.isOpenAPI() && key == "callbacks":
			return r.inboundCallbacksExpansionBytes(child, stack, depth)
		default:
			return sourceResponseStructuralBytes(child, r.limits)
		}
	})
}

func (r *sourceReferenceResolver) inboundParametersExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	parameters, ok := value.([]any)
	if !ok {
		return 0, fmt.Errorf("parameters must be an array")
	}
	return sourceArrayExpandedBytes(parameters, func(parameter any) (int64, error) {
		return r.parameterExpansionBytes(parameter, stack, depth)
	})
}

func (r *sourceReferenceResolver) inboundResponsesExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	responses, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("responses must be an object")
	}
	return sourceMapExpandedBytes(responses, func(status string, response any) (int64, error) {
		if strings.HasPrefix(status, "x-") {
			return sourceResponseStructuralBytes(response, r.limits)
		}
		return r.responseExpansionBytes(response, stack, depth)
	})
}

func (r *sourceReferenceResolver) inboundCallbacksExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	callbacks, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("callbacks must be an object")
	}
	return sourceMapExpandedBytes(callbacks, func(name string, callback any) (int64, error) {
		if strings.HasPrefix(name, "x-") {
			return sourceResponseStructuralBytes(callback, r.limits)
		}
		return r.inboundCallbackExpansionBytes(callback, stack, depth)
	})
}

func (r *sourceReferenceResolver) inboundCallbackExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("callback must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceCallback, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.inboundCallbackExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	return sourceMapExpandedBytes(object, func(expression string, pathItem any) (int64, error) {
		if strings.HasPrefix(expression, "x-") {
			return sourceResponseStructuralBytes(pathItem, r.limits)
		}
		return r.inboundPathItemExpansionBytes(pathItem, stack, depth)
	})
}

func (r *sourceReferenceResolver) resolveParameter(value any, stack map[string]bool, depth int) (map[string]any, error) {
	if err := r.reserveParameterEntry(value, stack, depth); err != nil {
		return nil, err
	}
	return r.resolveParameterUnchecked(value, stack, depth)
}

func (r *sourceReferenceResolver) resolveParameterUnchecked(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("parameter must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceParameter, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveParameterUnchecked(target, next, depth+1)
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
	if err := r.reserveRequestBodyEntry(value, stack, depth); err != nil {
		return nil, err
	}
	return r.resolveRequestBodyUnchecked(value, stack, depth)
}

func (r *sourceReferenceResolver) resolveRequestBodyUnchecked(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("request body must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceRequestBody, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveRequestBodyUnchecked(target, next, depth+1)
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

func (r *sourceReferenceResolver) reserveParameterEntry(value any, stack map[string]bool, depth int) error {
	bytes, err := r.parameterExpansionBytes(value, stack, depth)
	if err != nil {
		return err
	}
	return r.mediaExpansion.reserve(bytes)
}

func (r *sourceReferenceResolver) parameterExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("parameter must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceParameter, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.parameterExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	return sourceMapExpandedBytes(object, func(key string, child any) (int64, error) {
		switch {
		case key == "schema":
			return r.responseSchemaExpansionBytes(child, stack, depth)
		case r.form.isOpenAPI() && key == "content":
			return r.responseContentExpansionBytes(child, stack, depth, false)
		case r.form.isOpenAPI() && key == "examples":
			return r.responseExamplesExpansionBytes(child, stack, depth)
		case r.form.isSwagger2() && key == "items":
			return r.responseSchemaExpansionBytes(child, stack, depth)
		default:
			return sourceResponseStructuralBytes(child, r.limits)
		}
	})
}

func (r *sourceReferenceResolver) reserveRequestBodyEntry(value any, stack map[string]bool, depth int) error {
	bytes, err := r.requestBodyExpansionBytes(value, stack, depth)
	if err != nil {
		return err
	}
	return r.mediaExpansion.reserve(bytes)
}

func (r *sourceReferenceResolver) requestBodyExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("request body must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceRequestBody, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.requestBodyExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	return sourceMapExpandedBytes(object, func(key string, child any) (int64, error) {
		if key == "content" {
			return r.responseContentExpansionBytes(child, stack, depth, true)
		}
		return sourceResponseStructuralBytes(child, r.limits)
	})
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
	if hasReference {
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
	return r.responseObjectExpansionBytes(object, stack, depth)
}

func (r *sourceReferenceResolver) responseObjectExpansionBytes(object map[string]any, stack map[string]bool, depth int) (int64, error) {
	return sourceMapExpandedBytes(object, func(key string, value any) (int64, error) {
		switch {
		case r.form.isSwagger2() && key == "schema":
			return r.responseSchemaExpansionBytes(value, stack, depth)
		case r.form.isSwagger2() && key == "headers":
			return r.responseSwaggerHeadersExpansionBytes(value, stack, depth)
		case r.form.isOpenAPI() && key == "content":
			return r.responseContentExpansionBytes(value, stack, depth, false)
		case r.form.isOpenAPI() && key == "headers":
			return r.responseHeadersExpansionBytes(value, stack, depth)
		case r.form.isOpenAPI() && key == "links":
			return r.responseLinksExpansionBytes(value, stack, depth)
		default:
			return sourceResponseStructuralBytes(value, r.limits)
		}
	})
}

func (r *sourceReferenceResolver) responseContentExpansionBytes(value any, stack map[string]bool, depth int, resolveEncoding bool) (int64, error) {
	content, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("content must be an object")
	}
	return sourceMapExpandedBytes(content, func(mediaType string, rawMedia any) (int64, error) {
		media, ok := rawMedia.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("content media %q must be an object", mediaType)
		}
		if _, hasEncoding := media["encoding"]; hasEncoding && resolveEncoding && !sourceRequestFormMediaType(mediaType) {
			return 0, fmt.Errorf("unsupported encoding on request content media %q", mediaType)
		}
		return sourceMapExpandedBytes(media, func(key string, child any) (int64, error) {
			switch key {
			case "schema":
				return r.responseSchemaExpansionBytes(child, stack, depth)
			case "examples":
				return r.responseExamplesExpansionBytes(child, stack, depth)
			case "encoding":
				if !resolveEncoding {
					return 0, fmt.Errorf("unsupported encoding on non-request content media %q", mediaType)
				}
				return r.responseEncodingExpansionBytes(child, stack, depth)
			default:
				return sourceResponseStructuralBytes(child, r.limits)
			}
		})
	})
}

func (r *sourceReferenceResolver) responseSwaggerHeadersExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	headers, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("swagger response headers must be an object")
	}
	return sourceMapExpandedBytes(headers, func(name string, rawHeader any) (int64, error) {
		header, ok := rawHeader.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("swagger response header %q must be an object", name)
		}
		return sourceMapExpandedBytes(header, func(key string, child any) (int64, error) {
			if key == "items" {
				return r.responseSchemaExpansionBytes(child, stack, depth)
			}
			return sourceResponseStructuralBytes(child, r.limits)
		})
	})
}

func (r *sourceReferenceResolver) responseHeadersExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	headers, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("response headers must be an object")
	}
	return sourceMapExpandedBytes(headers, func(name string, header any) (int64, error) {
		bytes, err := r.responseHeaderExpansionBytes(header, stack, depth)
		if err != nil {
			return 0, fmt.Errorf("header %q: %w", name, err)
		}
		return bytes, nil
	})
}

func (r *sourceReferenceResolver) responseHeaderExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("header must be an object")
	}
	if !r.form.isOpenAPI() {
		return sourceResponseStructuralBytes(object, r.limits)
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceHeader, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.responseHeaderExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	return sourceMapExpandedBytes(object, func(key string, child any) (int64, error) {
		switch key {
		case "schema":
			return r.responseSchemaExpansionBytes(child, stack, depth)
		case "content":
			return r.responseContentExpansionBytes(child, stack, depth, false)
		case "examples":
			return r.responseExamplesExpansionBytes(child, stack, depth)
		default:
			return sourceResponseStructuralBytes(child, r.limits)
		}
	})
}

func (r *sourceReferenceResolver) responseExamplesExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	examples, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("examples must be an object")
	}
	return sourceMapExpandedBytes(examples, func(name string, example any) (int64, error) {
		bytes, err := r.responseExampleExpansionBytes(example, stack, depth)
		if err != nil {
			return 0, fmt.Errorf("example %q: %w", name, err)
		}
		return bytes, nil
	})
}

func (r *sourceReferenceResolver) responseExampleExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("example must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceExample, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if !hasReference {
		return sourceResponseStructuralBytes(object, r.limits)
	}
	targetBytes, err := r.responseExampleExpansionBytes(target, next, depth+1)
	if err != nil {
		return 0, err
	}
	referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
	if err != nil {
		return 0, err
	}
	return sourceStructuralAdd(targetBytes, referenceBytes)
}

func (r *sourceReferenceResolver) responseLinksExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	links, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("response links must be an object")
	}
	return sourceMapExpandedBytes(links, func(name string, link any) (int64, error) {
		bytes, err := r.responseLinkExpansionBytes(link, stack, depth)
		if err != nil {
			return 0, fmt.Errorf("link %q: %w", name, err)
		}
		return bytes, nil
	})
}

func (r *sourceReferenceResolver) responseLinkExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("link must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceLink, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.responseLinkExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	if err := r.validateLinkTarget(object); err != nil {
		return 0, err
	}
	return sourceResponseStructuralBytes(object, r.limits)
}

func sourceReferenceRequiresExpansionReservation(kind sourceReferenceKind) bool {
	switch kind {
	case sourceReferencePathItem, sourceReferenceHeader, sourceReferenceCallback, sourceReferenceLink, sourceReferenceExample, sourceReferenceSecurity:
		return true
	default:
		return false
	}
}

func (r *sourceReferenceResolver) reserveReferenceExpansion(kind sourceReferenceKind, value any, stack map[string]bool, depth int) error {
	bytes, err := r.referenceExpansionBytes(kind, value, stack, depth)
	if err != nil {
		return err
	}
	return r.referenceExpansion.reserve(bytes)
}

func (r *sourceReferenceResolver) referenceExpansionBytes(kind sourceReferenceKind, value any, stack map[string]bool, depth int) (int64, error) {
	switch kind {
	case sourceReferencePathItem:
		return r.inboundPathItemExpansionBytes(value, stack, depth)
	case sourceReferenceCallback:
		return r.inboundCallbackExpansionBytes(value, stack, depth)
	case sourceReferenceHeader:
		return r.responseHeaderExpansionBytes(value, stack, depth)
	case sourceReferenceLink:
		return r.responseLinkExpansionBytes(value, stack, depth)
	case sourceReferenceExample:
		return r.responseExampleExpansionBytes(value, stack, depth)
	case sourceReferenceSecurity:
		return r.securitySchemeExpansionBytes(value, stack, depth)
	default:
		return 0, fmt.Errorf("unsupported reference expansion kind %q", kind)
	}
}

func (r *sourceReferenceResolver) securitySchemeExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("security scheme must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceSecurity, stack, depth, false)
	if err != nil {
		return 0, err
	}
	if !hasReference {
		return sourceResponseStructuralBytes(object, r.limits)
	}
	targetBytes, err := r.securitySchemeExpansionBytes(target, next, depth+1)
	if err != nil {
		return 0, err
	}
	referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
	if err != nil {
		return 0, err
	}
	return sourceStructuralAdd(targetBytes, referenceBytes)
}

func (r *sourceReferenceResolver) responseEncodingExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	encodings, ok := value.(map[string]any)
	if !ok {
		return 0, fmt.Errorf("encoding must be an object")
	}
	return sourceMapExpandedBytes(encodings, func(property string, rawEncoding any) (int64, error) {
		encoding, ok := rawEncoding.(map[string]any)
		if !ok {
			return 0, fmt.Errorf("encoding %q must be an object", property)
		}
		return sourceMapExpandedBytes(encoding, func(key string, child any) (int64, error) {
			if key == "headers" {
				return r.responseHeadersExpansionBytes(child, stack, depth)
			}
			return sourceResponseStructuralBytes(child, r.limits)
		})
	})
}

func (r *sourceReferenceResolver) responseSchemaExpansionBytes(value any, stack map[string]bool, depth int) (int64, error) {
	if depth > r.limits.MaxReferenceDepth {
		return 0, fmt.Errorf("schema depth limit exceeded")
	}
	if _, err := sourceSchemaStructuralBytes(value, 0, r.limits); err != nil {
		return 0, err
	}
	if booleanSchema, ok := value.(bool); ok {
		if !r.form.isOpenAPI31() {
			return 0, fmt.Errorf("unsupported %s boolean schema", r.form.Family)
		}
		return sourceResponseStructuralBytes(booleanSchema, r.limits)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return sourceResponseStructuralBytes(value, r.limits)
	}
	if err := sourceRejectDynamicSchemaKeywords(object); err != nil {
		return 0, err
	}
	target, reference, next, hasReference, err := r.referenceTargetWithCount(object, sourceReferenceSchema, stack, depth, false)
	if err != nil {
		var cycle *sourceSchemaReferenceCycleError
		if errors.As(err, &cycle) {
			return sourceResponseStructuralBytes(value, r.limits)
		}
		return 0, err
	}
	if hasReference {
		targetBytes, err := r.responseSchemaExpansionBytes(target, next, depth+1)
		if err != nil {
			return 0, err
		}
		referenceBytes, err := sourceResponseStructuralBytes(reference, r.limits)
		if err != nil {
			return 0, err
		}
		return sourceStructuralAdd(targetBytes, referenceBytes)
	}
	if err := sourceValidateSchemaForm(object, r.form); err != nil {
		return 0, err
	}
	grammar := sourceSchemaGrammarFor(r.form)
	for _, key := range sourceSchemaBearingKeywords {
		if _, exists := object[key]; exists && !grammar.handles(key) {
			return 0, fmt.Errorf("unsupported %s schema keyword %s", r.form.Family, key)
		}
	}
	return r.responseSchemaObjectExpansionBytes(object, grammar, stack, depth)
}

func (r *sourceReferenceResolver) responseSchemaObjectExpansionBytes(object map[string]any, grammar sourceSchemaGrammar, stack map[string]bool, depth int) (int64, error) {
	mapChildren := map[string]bool{}
	for _, key := range grammar.mapChildren {
		mapChildren[key] = true
	}
	schemaChildren := map[string]bool{}
	for _, key := range grammar.schemaChildren {
		schemaChildren[key] = true
	}
	arrayChildren := map[string]bool{}
	for _, key := range grammar.arrayChildren {
		arrayChildren[key] = true
	}
	return sourceMapExpandedBytes(object, func(key string, value any) (int64, error) {
		switch {
		case mapChildren[key]:
			children, ok := value.(map[string]any)
			if !ok {
				return 0, fmt.Errorf("schema %s must be an object", key)
			}
			return sourceMapExpandedBytes(children, func(_ string, child any) (int64, error) {
				return r.responseSchemaExpansionBytes(child, stack, depth+1)
			})
		case grammar.dependencyChildren && key == "dependencies":
			children, ok := value.(map[string]any)
			if !ok {
				return 0, fmt.Errorf("schema dependencies must be an object")
			}
			return sourceMapExpandedBytes(children, func(name string, child any) (int64, error) {
				switch child.(type) {
				case map[string]any, bool:
					return r.responseSchemaExpansionBytes(child, stack, depth+1)
				case []any:
					return sourceResponseStructuralBytes(child, r.limits)
				default:
					return 0, fmt.Errorf("schema dependencies[%q] must be a schema or string array", name)
				}
			})
		case schemaChildren[key]:
			if key == "additionalProperties" && !r.form.isOpenAPI31() {
				if booleanValue, isBoolean := value.(bool); isBoolean {
					return sourceResponseStructuralBytes(booleanValue, r.limits)
				}
			}
			return r.responseSchemaExpansionBytes(value, stack, depth+1)
		case arrayChildren[key]:
			items, ok := value.([]any)
			if !ok {
				return 0, fmt.Errorf("schema %s must be an array", key)
			}
			return sourceArrayExpandedBytes(items, func(item any) (int64, error) {
				return r.responseSchemaExpansionBytes(item, stack, depth+1)
			})
		default:
			return sourceResponseStructuralBytes(value, r.limits)
		}
	})
}

func sourceMapExpandedBytes(object map[string]any, valueBytes func(string, any) (int64, error)) (int64, error) {
	used := int64(2)
	for _, key := range sortedSourceMapKeys(object) {
		bytes, err := valueBytes(key, object[key])
		if err != nil {
			return 0, err
		}
		used, err = sourceStructuralAdd(used, sourceStructuralStringBytes(key))
		if err != nil {
			return 0, err
		}
		used, err = sourceStructuralAdd(used, bytes)
		if err != nil {
			return 0, err
		}
		used, err = sourceStructuralAdd(used, 2)
		if err != nil {
			return 0, err
		}
	}
	return used, nil
}

func sourceArrayExpandedBytes(items []any, valueBytes func(any) (int64, error)) (int64, error) {
	used := int64(2)
	for _, item := range items {
		bytes, err := valueBytes(item)
		if err != nil {
			return 0, err
		}
		used, err = sourceStructuralAdd(used, bytes)
		if err != nil {
			return 0, err
		}
		used, err = sourceStructuralAdd(used, 1)
		if err != nil {
			return 0, err
		}
	}
	return used, nil
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
		return nil, fmt.Errorf("swagger response headers must be an object")
	}
	out := make(map[string]any, len(headers))
	for _, name := range sortedSourceMapKeys(headers) {
		header, ok := headers[name].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("swagger response header %q must be an object", name)
		}
		resolvedHeader := sourceCloneMap(header)
		if items, exists := header["items"]; exists {
			resolved, err := r.resolveSchema(items, stack, depth)
			if err != nil {
				return nil, fmt.Errorf("swagger response header %q items: %w", name, err)
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
		if _, exists := media["encoding"]; exists && resolveEncoding && !sourceRequestFormMediaType(mediaType) {
			return nil, fmt.Errorf("unsupported encoding on request content media %q", mediaType)
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
		object = sourceMergeReferenceObject(resolved, reference)
	}
	if err := r.validateLinkTarget(object); err != nil {
		return nil, err
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
	if depth == 0 && len(stack) == 0 {
		// Bound each retained root schema independently. Aggregate descriptor,
		// response, media, and reference budgets already account for the encoded
		// result; carrying this expansion counter across unrelated roots made a
		// large but finite source fail according to traversal order.
		r.expansion = sourceSchemaExpansionBudget{}
	}
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
	if r.knownSchemaCycleReference(object) {
		for key := range object {
			if !r.referenceSiblingAllowed(sourceReferenceSchema, key) {
				return nil, fmt.Errorf("ambiguous schema reference with sibling field %q", key)
			}
		}
		if r.form.isOpenAPI() && !r.form.isOpenAPI31() {
			if gap := sourceOpenAPI30SchemaReferenceSiblingGap(nil, object); gap != nil {
				r.schemaReferenceSiblingGaps = append(r.schemaReferenceSiblingGaps, *gap)
			}
		}
		return sourceCloneLiteral(object), nil
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceSchema, stack, depth)
	if err != nil {
		var cycle *sourceSchemaReferenceCycleError
		if errors.As(err, &cycle) {
			if r.form.isOpenAPI() && !r.form.isOpenAPI31() {
				if gap := sourceOpenAPI30SchemaReferenceSiblingGap(nil, object); gap != nil {
					r.schemaReferenceSiblingGaps = append(r.schemaReferenceSiblingGaps, *gap)
				}
			}
			return sourceCloneLiteral(object), nil
		}
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
		if r.form.isOpenAPI() && !r.form.isOpenAPI31() {
			if gap := sourceOpenAPI30SchemaReferenceSiblingGap(targetObject, reference); gap != nil {
				r.schemaReferenceSiblingGaps = append(r.schemaReferenceSiblingGaps, *gap)
			}
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
	if err := r.reserveInboundCallback(value, stack, depth); err != nil {
		return nil, err
	}
	return r.resolveCallbackUnchecked(value, stack, depth)
}

func (r *sourceReferenceResolver) resolveCallbackUnchecked(value any, stack map[string]bool, depth int) (map[string]any, error) {
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("callback must be an object")
	}
	target, reference, next, hasReference, err := r.referenceTarget(object, sourceReferenceCallback, stack, depth)
	if err != nil {
		return nil, err
	}
	if hasReference {
		resolved, err := r.resolveCallbackUnchecked(target, next, depth+1)
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
		pathItem, err := r.resolveInboundPathItemUnchecked(object[expression], nil, 0)
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
	for key := range object {
		if !r.referenceSiblingAllowed(kind, key) {
			return nil, nil, nil, false, fmt.Errorf("ambiguous %s reference with sibling field %q", kind, key)
		}
	}
	rawReference, ok := rawRef.(string)
	if !ok {
		return nil, nil, nil, false, fmt.Errorf("external reference %q is unsupported", rawRef)
	}
	ref, err := sourceNormalizeLocalReference(rawReference)
	if err != nil {
		return nil, nil, nil, false, err
	}
	if countReference {
		r.references++
		if r.references > r.limits.MaxReferences {
			return nil, nil, nil, false, fmt.Errorf("reference count limit exceeded")
		}
	}
	if stack != nil && stack[ref] {
		if kind == sourceReferenceSchema {
			if r.schemaCycles == nil {
				r.schemaCycles = map[string]struct{}{}
			}
			r.schemaCycles[ref] = struct{}{}
			return nil, nil, nil, false, &sourceSchemaReferenceCycleError{Reference: ref}
		}
		return nil, nil, nil, false, fmt.Errorf("reference cycle at %q", ref)
	}
	if depth >= r.limits.MaxReferenceDepth {
		return nil, nil, nil, false, fmt.Errorf("reference depth limit exceeded")
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
	if countReference && sourceReferenceRequiresExpansionReservation(kind) {
		if err := r.reserveReferenceExpansion(kind, target, next, depth+1); err != nil {
			return nil, nil, nil, false, err
		}
	}
	return target, object, next, true, nil
}

func (r *sourceReferenceResolver) referenceSiblingAllowed(kind sourceReferenceKind, key string) bool {
	if key == "$ref" || strings.HasPrefix(key, "x-") {
		return true
	}
	// Provider OpenAPI 3.0 descriptions commonly annotate a referenced object.
	// Description, summary, and a schema's readOnly directionality do not change
	// the reference target or its resolved wire shape. A schema type assertion
	// that differs from the target is retained as source-bound, merge-blocking
	// evidence by resolveSchema. All other structural or semantic overrides stay
	// rejected.
	if r.form.isOpenAPI() && (key == "description" || key == "summary" || (kind == sourceReferenceSchema && (key == "readOnly" || key == "type"))) {
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

func sourceOpenAPI30SchemaReferenceSiblingGap(target, reference map[string]any) *sourceContractGap {
	typeSibling, declared := reference["type"]
	if !declared {
		return nil
	}
	targetType, targetDeclaresType := target["type"]
	if targetDeclaresType && reflect.DeepEqual(typeSibling, targetType) {
		return nil
	}
	referencePointer, _ := reference["$ref"].(string)
	if normalized, err := sourceNormalizeLocalReference(referencePointer); err == nil {
		referencePointer = normalized
	}
	return &sourceContractGap{
		Foundation: sourceOpenAPI30ReferenceSiblingFoundation,
		Location:   fmt.Sprintf("schema reference %s", referencePointer),
		Reason:     `OpenAPI 3.0 schema reference retains sibling field "type" as source-bound evidence because its target does not declare the same type`,
	}
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
		if err := r.validateLinkTarget(object); err != nil {
			return fmt.Errorf("link reference does not resolve to a link object: %w", err)
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

func (r *sourceReferenceResolver) validateLinkTarget(object map[string]any) error {
	rawOperationID, hasOperationID := object["operationId"]
	rawOperationRef, hasOperationRef := object["operationRef"]
	if hasOperationID == hasOperationRef {
		return fmt.Errorf("link requires exactly one of operationId or operationRef")
	}
	if r.referenceIndex == nil {
		return fmt.Errorf("reference grammar index is unavailable")
	}
	if hasOperationID {
		operationID, ok := rawOperationID.(string)
		if !ok || operationID == "" {
			return fmt.Errorf("link operationId must be a non-empty string")
		}
		count := r.referenceIndex.reachableOperationIDs[operationID]
		switch count {
		case 0:
			return fmt.Errorf("link operationId %q does not identify an in-artifact operation", operationID)
		case 1:
			return nil
		default:
			return fmt.Errorf("link operationId %q is ambiguous", operationID)
		}
	}
	operationRef, ok := rawOperationRef.(string)
	if !ok || operationRef == "" {
		return fmt.Errorf("link operationRef must be a non-empty string")
	}
	pointer, err := sourceLocalOperationReferencePointer(operationRef)
	if err != nil {
		return err
	}
	target, err := sourcePointer(r.root, pointer)
	if err != nil {
		return err
	}
	if actual, exists := r.referenceIndex.positions[pointer]; !exists || actual != sourceReferenceOperation {
		return fmt.Errorf("link operationRef %q does not resolve to an operation", operationRef)
	}
	if _, ok := target.(map[string]any); !ok {
		return fmt.Errorf("link operationRef %q does not resolve to an operation", operationRef)
	}
	count := r.referenceIndex.reachableOperationPositions[pointer]
	switch count {
	case 0:
		return fmt.Errorf("link operationRef %q does not identify a reachable in-artifact operation", operationRef)
	case 1:
		return nil
	default:
		return fmt.Errorf("link operationRef %q is ambiguous across reachable operations", operationRef)
	}
}

func sourceLocalOperationReferencePointer(raw string) (string, error) {
	pointer, err := sourceNormalizeLocalReference(raw)
	if err != nil {
		return "", err
	}
	if pointer == "#" {
		return "", fmt.Errorf("link operationRef %q must be a local artifact JSON Pointer", raw)
	}
	return pointer, nil
}

func sourceNormalizeLocalReference(raw string) (string, error) {
	if !strings.HasPrefix(raw, "#") {
		return "", fmt.Errorf("external reference %q is unsupported", raw)
	}
	fragment, err := url.PathUnescape(raw[1:])
	if err != nil {
		return "", fmt.Errorf("local reference %q has an invalid percent escape: %w", raw, err)
	}
	if fragment != "" && !strings.HasPrefix(fragment, "/") {
		return "", fmt.Errorf("local reference %q must be a local artifact JSON Pointer", raw)
	}
	if err := sourceValidateJSONPointer(fragment); err != nil {
		return "", fmt.Errorf("local reference %q has an invalid JSON Pointer: %w", raw, err)
	}
	return "#" + fragment, nil
}

func sourceReferenceStackAddition(stack, next map[string]bool) (string, error) {
	var reference string
	for candidate := range next {
		if stack != nil && stack[candidate] {
			continue
		}
		if reference != "" {
			return "", fmt.Errorf("reference traversal has multiple canonical additions")
		}
		reference = candidate
	}
	if reference == "" {
		return "", fmt.Errorf("reference traversal has no canonical addition")
	}
	return reference, nil
}

func sourceValidateJSONPointer(pointer string) error {
	for _, segment := range strings.Split(strings.TrimPrefix(pointer, "/"), "/") {
		for index := 0; index < len(segment); index++ {
			if segment[index] != '~' {
				continue
			}
			if index+1 >= len(segment) || (segment[index+1] != '0' && segment[index+1] != '1') {
				return fmt.Errorf("invalid escape")
			}
			index++
		}
	}
	return nil
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
	out := make(map[string]any, len(target))
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

func importSourceDocumentResult(documentContext sourceImportDocumentContext, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, limits sourceImportLimits, budget *sourceImportBudget) (sourceImportResult, error) {
	if budget == nil {
		return sourceImportResult{}, fmt.Errorf("source importer has no descriptor budget")
	}
	countBudget, err := budget.countBudget(limits)
	if err != nil {
		return sourceImportResult{}, err
	}
	if err := sourcePrepareSourceDocument(doc, form, limits, resolver, countBudget); err != nil {
		return sourceImportResult{}, err
	}
	result := sourceImportResult{Operations: []sourceOperationDescriptor{}}
	rootServers := sourceServerLayer{}
	if form.isOpenAPI() {
		rootServers, err = sourceServerLayerFrom(doc)
		if err != nil {
			return sourceImportResult{}, fmt.Errorf("root servers: %w", err)
		}
	}
	webhooks, err := sourceWebhookEvents(documentContext, doc, form, resolver)
	if err != nil {
		return sourceImportResult{}, err
	}
	for _, event := range webhooks {
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
		result.Gaps = append(result.Gaps, resolver.unreferencedSchemaCycleGaps(result.Operations)...)
		result.Gaps = append(result.Gaps, resolver.unreferencedSchemaReferenceSiblingGaps(result.Operations)...)
		result.Gaps = sourceSortedGaps(result.Gaps)
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
				events, err := sourceCallbackEvents(documentContext, form, "", fmt.Sprintf("paths[%q].callbacks", path), rawCallbacks, resolver)
				if err != nil {
					return sourceImportResult{}, err
				}
				for _, event := range events {
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
			descriptor, err := importSourceOperation(documentContext, doc, form, resolver, path, method, pathParameters, rootServers, pathServers, operation, limits, budget.remaining())
			if err != nil {
				return sourceImportResult{}, err
			}
			if err := budget.add(descriptor, "operation"); err != nil {
				return sourceImportResult{}, err
			}
			result.Operations = append(result.Operations, descriptor)
			if form.isOpenAPI() {
				if rawCallbacks, hasCallbacks := operation["callbacks"]; hasCallbacks {
					events, err := sourceCallbackEvents(documentContext, form, descriptor.SourceID, fmt.Sprintf("paths[%q].%s.callbacks", path, method), rawCallbacks, resolver)
					if err != nil {
						return sourceImportResult{}, err
					}
					for _, event := range events {
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
	result.Gaps = append(result.Gaps, resolver.unreferencedSchemaCycleGaps(result.Operations)...)
	result.Gaps = append(result.Gaps, resolver.unreferencedSchemaReferenceSiblingGaps(result.Operations)...)
	result.Gaps = sourceSortedGaps(result.Gaps)
	return result, nil
}

var sourceHTTPMethods = []string{"delete", "get", "head", "options", "patch", "post", "put", "trace"}

func validateSourceImportPath(path string) error {
	if path == "" || path != strings.TrimSpace(path) || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.Contains(path, "://") || strings.ContainsAny(path, "\r\n?#") {
		return fmt.Errorf("not connector-relative")
	}
	return nil
}

func importSourceOperation(documentContext sourceImportDocumentContext, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver, path, method string, pathParameters []sourceParameterValue, rootServers, pathServers sourceServerLayer, operation map[string]any, limits sourceImportLimits, remainingDescriptorBytes int64) (sourceOperationDescriptor, error) {
	lock := documentContext.Lock
	location := fmt.Sprintf("paths[%q].%s", path, method)
	schemaReferenceSiblingGapsStart := len(resolver.schemaReferenceSiblingGaps)
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
	if documentContext.Document != nil {
		locked, exists := documentContext.lockedRESTOperation(method, path)
		if !exists {
			return sourceOperationDescriptor{}, fmt.Errorf("%s is not present in source document %q inventory", location, documentContext.Document.ID)
		}
		if locked.OperationID != providerID || locked.SourceLocation != location {
			return sourceOperationDescriptor{}, fmt.Errorf("%s disagrees with source document %q inventory", location, documentContext.Document.ID)
		}
		sourceID = locked.ID
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
	// A root server is the connector's ordinary base declaration. A single
	// fixed path/operation origin is representable by the shared declaration-
	// owned base_url override; only variable or ambiguous route sets remain a
	// typed gap.
	if (pathServers.Declared && !sourceServerLayerHasFixedOrigin(pathServers)) ||
		(operationServers.Declared && !sourceServerLayerHasFixedOrigin(operationServers)) {
		servers.Gaps = []sourceContractGap{sourceContractGapFor("cli-operation-route-override-foundation-r1", location+".servers", "provider-declared server routing is variable or ambiguous")}
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
	runtimeGaps := append(sourceRequestGaps(request, form, limits, method), servers.Gaps...)
	runtimeGaps = append(runtimeGaps, resolver.requestSchemaCycleGaps(request)...)
	runtimeGaps = append(runtimeGaps, resolver.responseSchemaCycleGaps(responses)...)
	runtimeGaps = append(runtimeGaps, resolver.schemaReferenceSiblingGaps[schemaReferenceSiblingGapsStart:]...)
	runtimeGaps = sourceSortedGaps(runtimeGaps)
	runtime := sourceRuntimeReachability{MergeBlocked: len(runtimeGaps) > 0, Gaps: runtimeGaps}
	return sourceOperationDescriptor{
		Connector:           lock.Connector,
		Protocol:            "rest",
		SourceID:            sourceID,
		ProviderOperationID: providerID,
		Source:              sourceImportProvenance(documentContext, form, location),
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
		return nil, nil, fmt.Errorf("swagger parameter content is unsupported")
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
			if len(resolver.schemaCycleReferences(parameter.Schema)) == 0 {
				if err := validateBoundedOperationParameterSchema(parameter.Schema, form, limits, parameter.In); err != nil {
					return sourceRequestDescriptor{}, fmt.Errorf("parameter %q: %w", parameter.Name, err)
				}
			}
		} else if err := validateBoundedParameterContent(parameter.Name, parameter.Content, form, limits); err != nil {
			return sourceRequestDescriptor{}, err
		}
		descriptor := sourceParameterDescriptor{
			Name:              parameter.Name,
			Required:          parameter.Required,
			Schema:            parameter.Schema,
			Content:           parameter.Content,
			Wire:              parameter.Wire,
			ExecutionEnvelope: sourceParameterExecutionEnvelopeFor(parameter, limits),
		}
		if err := validateSourceParameterExecutionEnvelope(descriptor, parameter.In, limits); err != nil {
			return sourceRequestDescriptor{}, fmt.Errorf("parameter %q: %w", parameter.Name, err)
		}
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
			return sourceRequestDescriptor{}, fmt.Errorf("swagger request body is ambiguous")
		}
		body := bodyParameters[0]
		if len(resolver.schemaCycleReferences(body.Schema)) == 0 {
			if err := validateBoundedRequestSchema(body.Schema, form, limits, 0); err != nil {
				return sourceRequestDescriptor{}, fmt.Errorf("request body: %w", err)
			}
		}
		mediaType, err := sourceSwaggerRequestMediaType(operation, doc)
		if err != nil {
			return sourceRequestDescriptor{}, err
		}
		request.Body = &sourceRequestBodyDescriptor{
			Required:          body.Required,
			Schema:            body.Schema,
			ExecutionEnvelope: sourceRequestBodyExecutionEnvelope(mediaType, limits, "request.body"),
		}
		if err := validateSourceRequestBodyExecutionEnvelope(request.Body.ExecutionEnvelope, mediaType, limits); err != nil {
			return sourceRequestDescriptor{}, err
		}
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
	if !ok || len(content) == 0 {
		return sourceRequestDescriptor{}, fmt.Errorf("request body requires at least one declared media type")
	}
	for _, mediaType := range sortedSourceMapKeys(content) {
		media, ok := content[mediaType].(map[string]any)
		if !ok {
			return sourceRequestDescriptor{}, fmt.Errorf("request media declaration %q must be an object", mediaType)
		}
		encoding, hasEncoding := media["encoding"]
		formMedia := sourceRequestFormMediaType(mediaType)
		if hasEncoding && !formMedia {
			return sourceRequestDescriptor{}, fmt.Errorf("unsupported encoding on request content media %q", mediaType)
		}
		if !sourceJSONMediaType(mediaType) && !formMedia && mediaType != "text/plain" && mediaType != "application/octet-stream" {
			if !limits.AllowSourceContractGaps {
				return sourceRequestDescriptor{}, fmt.Errorf("unsupported request encoding %q", mediaType)
			}
		}
		schema, ok := media["schema"]
		if !ok {
			if !limits.AllowSourceContractGaps {
				return sourceRequestDescriptor{}, fmt.Errorf("request media %q is missing schema", mediaType)
			}
			schema = true
		}
		if len(resolver.schemaCycleReferences(schema)) == 0 {
			if err := validateBoundedRequestSchema(schema, form, limits, 0); err != nil {
				return sourceRequestDescriptor{}, fmt.Errorf("request media %q: %w", mediaType, err)
			}
		}
		descriptor := sourceRequestMediaDescriptor{
			MediaType:         mediaType,
			Required:          sourceBool(body["required"]),
			Schema:            schema,
			Encoding:          encoding,
			ExecutionEnvelope: sourceRequestBodyExecutionEnvelope(mediaType, limits, "request.media["+strconv.Quote(mediaType)+"]"),
		}
		if err := validateSourceRequestBodyExecutionEnvelope(descriptor.ExecutionEnvelope, mediaType, limits); err != nil {
			return sourceRequestDescriptor{}, fmt.Errorf("request media %q: %w", mediaType, err)
		}
		request.Media = append(request.Media, descriptor)
	}
	if len(request.Media) == 1 {
		media := request.Media[0]
		request.Body = &sourceRequestBodyDescriptor{Required: media.Required, Schema: media.Schema, Encoding: media.Encoding, ExecutionEnvelope: media.ExecutionEnvelope}
		request.MediaType = media.MediaType
		request.Media = nil
	}
	return request, nil
}

func sourceParameterExecutionEnvelopeFor(parameter sourceParameterValue, limits sourceImportLimits) *sourceExecutionEnvelope {
	if !limits.UseExecutionEnvelopes || parameter.Content != nil || !sourceScalarWireSchema(parameter.Schema) {
		return nil
	}
	if parameter.In == "header" {
		effective := sourceBoundedHeaderMaxBytes(parameter.Schema)
		if effective <= 0 {
			return nil
		}
		return &sourceExecutionEnvelope{
			PolicyVersion:  engine.OperationParameterExecutionPolicyVersion,
			Origin:         "provider_and_pm_policy",
			SourceLocation: fmt.Sprintf("request.%s[%q]", parameter.In, parameter.Name),
			Limits: []sourceExecutionLimit{{
				Kind:        "wire_value",
				Unit:        "bytes",
				HardCeiling: engine.MaxOperationHeaderBytes,
				Effective:   effective,
			}},
		}
	}
	if parameter.In != "path" && parameter.In != "query" {
		return nil
	}
	effective, origin := sourcePathQueryExecutionBound(parameter.Schema)
	return &sourceExecutionEnvelope{
		PolicyVersion:  engine.OperationParameterExecutionPolicyVersion,
		Origin:         origin,
		SourceLocation: fmt.Sprintf("request.%s[%q]", parameter.In, parameter.Name),
		Limits: []sourceExecutionLimit{{
			Kind:        "wire_value",
			Unit:        "encoded_bytes",
			Default:     engine.DefaultOperationParameterMaxBytes,
			HardCeiling: engine.MaxOperationParameterMaxBytes,
			Effective:   effective,
		}},
	}
}

func sourcePathQueryExecutionBound(schema any) (int, string) {
	effective := engine.DefaultOperationParameterMaxBytes
	origin := "pm_policy"
	if providerBytes := sourceProjectionFlagMaxBytes(schema, "string"); providerBytes > 0 {
		origin = "provider_and_pm_policy"
		if providerBytes < int64(effective) {
			effective = int(providerBytes)
		}
	}
	return effective, origin
}

func sourceBoundedHeaderMaxBytes(schema any) int {
	object, ok := schema.(map[string]any)
	if !ok {
		return 0
	}
	if object["type"] == "boolean" {
		return len("false")
	}
	if object["type"] != "string" {
		return 0
	}
	maxBytes := 0
	if maxRunes := sourceProjectionSchemaMaxBytes(schema); maxRunes > 0 && maxRunes <= int64(engine.MaxOperationHeaderBytes/utf8.UTFMax) {
		maxBytes = int(maxRunes) * utf8.UTFMax
	}
	for _, raw := range sourceAnySlice(object["enum"]) {
		value, ok := raw.(string)
		if !ok {
			return 0
		}
		if len(value) > maxBytes {
			maxBytes = len(value)
		}
	}
	if maxBytes <= 0 || maxBytes > engine.MaxOperationHeaderBytes {
		return 0
	}
	return maxBytes
}

func sourceRequestBodyExecutionEnvelope(mediaType string, limits sourceImportLimits, sourceLocation string) *sourceExecutionEnvelope {
	if !limits.UseExecutionEnvelopes || !sourceJSONMediaType(sourceNormalizedMediaType(mediaType)) {
		return nil
	}
	return &sourceExecutionEnvelope{
		PolicyVersion:  engine.OperationParameterExecutionPolicyVersion,
		Origin:         "pm_policy",
		SourceLocation: sourceLocation,
		Limits: []sourceExecutionLimit{
			{Kind: "body", Unit: "encoded_bytes", Default: engine.DefaultOperationDirectWriteMaxBytes, HardCeiling: engine.MaxOperationDirectWriteBytes, Effective: engine.DefaultOperationDirectWriteMaxBytes},
			{Kind: "array", Unit: "items", Default: sourceProjectionDefaultArrayItems, HardCeiling: engine.MaxStructuredRESTBodyItems, Effective: sourceProjectionDefaultArrayItems},
			{Kind: "object", Unit: "properties", Default: sourceProjectionDefaultObjectProperties, HardCeiling: engine.MaxStructuredRESTBodyFields, Effective: sourceProjectionDefaultObjectProperties},
			{Kind: "structure", Unit: "depth", Default: engine.MaxStructuredRESTBodyDepth, HardCeiling: engine.MaxStructuredRESTBodyDepth, Effective: engine.MaxStructuredRESTBodyDepth},
			{Kind: "structure", Unit: "nodes", Default: engine.MaxStructuredRESTBodyNodes, HardCeiling: engine.MaxStructuredRESTBodyNodes, Effective: engine.MaxStructuredRESTBodyNodes},
		},
	}
}

func validateSourceRequestBodyExecutionEnvelope(envelope *sourceExecutionEnvelope, mediaType string, limits sourceImportLimits) error {
	if !limits.UseExecutionEnvelopes || !sourceJSONMediaType(sourceNormalizedMediaType(mediaType)) {
		return nil
	}
	if envelope == nil {
		return fmt.Errorf("JSON request body requires a valid PM execution envelope")
	}
	mediaLocation := "request.media[" + strconv.Quote(mediaType) + "]"
	if envelope.SourceLocation != "request.body" && envelope.SourceLocation != mediaLocation {
		return fmt.Errorf("JSON request body requires a valid PM execution envelope")
	}
	expected := sourceRequestBodyExecutionEnvelope(mediaType, limits, envelope.SourceLocation)
	if !reflect.DeepEqual(envelope, expected) {
		return fmt.Errorf("JSON request body requires a valid PM execution envelope")
	}
	return nil
}

func validateSourceParameterExecutionEnvelope(parameter sourceParameterDescriptor, location string, limits sourceImportLimits) error {
	if !limits.UseExecutionEnvelopes || !sourceScalarWireSchema(parameter.Schema) {
		return nil
	}
	envelope := parameter.ExecutionEnvelope
	if location == "header" {
		effective := sourceBoundedHeaderMaxBytes(parameter.Schema)
		if effective == 0 {
			return nil
		}
		if envelope == nil || envelope.PolicyVersion != engine.OperationParameterExecutionPolicyVersion || envelope.Origin != "provider_and_pm_policy" || envelope.SourceLocation != fmt.Sprintf("request.%s[%q]", location, parameter.Name) || len(envelope.Limits) != 1 {
			return fmt.Errorf("executable source header has an invalid PM execution envelope")
		}
		limit := envelope.Limits[0]
		if limit.Kind != "wire_value" || limit.Unit != "bytes" || limit.Default != 0 || limit.HardCeiling != engine.MaxOperationHeaderBytes || limit.Effective != effective {
			return fmt.Errorf("executable source header has an invalid byte execution limit")
		}
		return nil
	}
	if location != "path" && location != "query" {
		return nil
	}
	if envelope == nil {
		return fmt.Errorf("executable source input requires a PM execution envelope")
	}
	effective, origin := sourcePathQueryExecutionBound(parameter.Schema)
	if envelope.PolicyVersion != engine.OperationParameterExecutionPolicyVersion || envelope.Origin != origin || envelope.SourceLocation != fmt.Sprintf("request.%s[%q]", location, parameter.Name) || len(envelope.Limits) != 1 {
		return fmt.Errorf("executable source input has an invalid PM execution envelope")
	}
	limit := envelope.Limits[0]
	if limit.Kind != "wire_value" || limit.Unit != "encoded_bytes" || limit.Default != engine.DefaultOperationParameterMaxBytes || limit.HardCeiling != engine.MaxOperationParameterMaxBytes || limit.Effective != effective {
		return fmt.Errorf("executable source input has an invalid encoded-byte execution limit")
	}
	return nil
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
		return "", fmt.Errorf("swagger request body has no declared consumes media type")
	}
	mediaTypes, err := sourceStringArray(rawConsumes, "Swagger consumes")
	if err != nil {
		return "", err
	}
	if len(mediaTypes) != 1 {
		return "", fmt.Errorf("swagger request body requires exactly one unambiguous media type")
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

func sourceRequestGaps(request sourceRequestDescriptor, form sourceDocumentForm, limits sourceImportLimits, method string) []sourceContractGap {
	var gaps []sourceContractGap
	if len(request.Media) > 1 {
		gaps = append(gaps, sourceContractGapFor("cli-request-media-selection-foundation-r1", "request body media", "provider declares multiple request media types; generated command must select one exact media contract"))
	}
	if request.Body != nil && sourceRequestFormMediaType(request.MediaType) {
		gaps = append(gaps, sourceContractGapFor("cli-request-encoding-foundation-r1", "request body encoding", "provider form request media requires runtime encoding support"))
	}
	for _, group := range []struct {
		location   string
		parameters []sourceParameterDescriptor
	}{
		{location: "path", parameters: request.Path},
		{location: "query", parameters: request.Query},
		{location: "header", parameters: request.Header},
	} {
		for _, parameter := range group.parameters {
			gaps = append(gaps, parameter.Wire.Gaps...)
			if parameter.Schema != nil {
				if err := sourceProjectionOperationParameterGap(parameter.Schema, form, limits, group.location, method); err != nil {
					gaps = append(gaps, sourceContractGapFor("cli-request-schema-foundation-r1", "parameter "+parameter.Name, err.Error()))
				}
			}
		}
	}
	if request.Body != nil {
		if err := sourceProjectionSchemaGap(request.Body.Schema, form, limits); err != nil {
			gaps = append(gaps, sourceContractGapFor("cli-request-schema-foundation-r1", "request body", err.Error()))
		}
	}
	for _, media := range request.Media {
		if err := sourceProjectionSchemaGap(media.Schema, form, limits); err != nil {
			gaps = append(gaps, sourceContractGapFor("cli-request-schema-foundation-r1", "request media "+media.MediaType, err.Error()))
		}
	}
	return sourceSortedGaps(gaps)
}

func validateBoundedOperationParameterSchema(schema any, form sourceDocumentForm, limits sourceImportLimits, location string) error {
	if (location == "path" || location == "query") && sourceStringScalarWireUnion(schema) {
		return nil
	}
	return validateBoundedRequestSchema(schema, form, limits, 0)
}

func sourceProjectionOperationParameterGap(schema any, form sourceDocumentForm, limits sourceImportLimits, location, method string) error {
	if !sourceProjectionMutationMethod(method) && (location == "path" || location == "query") && sourceStringScalarWireUnion(schema) {
		return nil
	}
	if (location == "path" || location == "query") && !sourceScalarWireSchema(schema) {
		return fmt.Errorf("%s parameter requires non-scalar serialization support", location)
	}
	if location == "header" && sourceScalarWireSchema(schema) && !sourceStringScalarWireUnion(schema) && sourceBoundedHeaderMaxBytes(schema) == 0 {
		return fmt.Errorf("unbounded request header requires a compatibility-censused PM byte envelope")
	}
	return sourceProjectionSchemaGap(schema, form, limits)
}

func sourceScalarWireSchema(schema any) bool {
	if sourceStringScalarWireUnion(schema) {
		return true
	}
	object, ok := schema.(map[string]any)
	if !ok {
		return false
	}
	typeName, _ := object["type"].(string)
	switch typeName {
	case "string", "integer", "number", "boolean":
		return true
	default:
		return false
	}
}

// sourceStringScalarWireUnion admits only a source union with an
// unconstrained string arm and scalar alternatives. Path and query values are
// textual on the wire, so that string arm covers every possible encoded value;
// retaining the exact source schema in the descriptor avoids pretending the
// provider only accepts one arm. Object, array, constrained, header, and body
// unions remain gaps because they require a distinct serialization contract.
func sourceStringScalarWireUnion(schema any) bool {
	root, ok := schema.(map[string]any)
	if !ok {
		return false
	}
	arms, ok := root["oneOf"].([]any)
	if !ok || len(arms) == 0 {
		arms, ok = root["anyOf"].([]any)
	}
	if !ok || len(arms) < 2 {
		return false
	}
	hasOpenString := false
	for _, rawArm := range arms {
		arm, ok := rawArm.(map[string]any)
		if !ok || len(arm) != 1 {
			return false
		}
		typeName, ok := arm["type"].(string)
		if !ok {
			return false
		}
		switch typeName {
		case "string":
			hasOpenString = true
		case "integer", "number", "boolean":
		default:
			return false
		}
	}
	return hasOpenString
}

func sourceProjectionSchemaGap(schema any, form sourceDocumentForm, limits sourceImportLimits) error {
	strict := limits
	strict.AllowSourceContractGaps = false
	strict.UseExecutionEnvelopes = true
	if err := validateBoundedRequestSchema(schema, form, strict, 0); err != nil {
		return err
	}
	// The policy-aware semantic pass still validates the entire schema after a
	// missing optional maximum. Projection then classifies representation
	// blockers such as unions and dynamic objects without error-string matching.
	_, projectionErr := sourceProjectionSchema(schema)
	return projectionErr
}

func sourceServerLayerHasFixedOrigin(layer sourceServerLayer) bool {
	items, ok := layer.Servers.([]any)
	if !ok || len(items) != 1 {
		return false
	}
	server, ok := items[0].(map[string]any)
	if !ok {
		return false
	}
	raw, ok := server["url"].(string)
	if !ok || strings.Contains(raw, "{") {
		return false
	}
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Host != "" && (parsed.Scheme == "https" || parsed.Scheme == "http") && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && (parsed.Path == "" || parsed.Path == "/")
}

func sourceContractGapFor(foundation, location, reason string) sourceContractGap {
	return sourceContractGap{Foundation: foundation, Location: location, Reason: reason}
}

func (r *sourceReferenceResolver) requestSchemaCycleGaps(request sourceRequestDescriptor) []sourceContractGap {
	var gaps []sourceContractGap
	for _, group := range []struct {
		location   string
		parameters []sourceParameterDescriptor
	}{
		{location: "path", parameters: request.Path},
		{location: "query", parameters: request.Query},
		{location: "header", parameters: request.Header},
	} {
		for _, parameter := range group.parameters {
			gaps = append(gaps, r.schemaCycleGaps(parameter.Schema, fmt.Sprintf("%s parameter %q schema", group.location, parameter.Name))...)
			gaps = append(gaps, r.schemaCycleGaps(parameter.Content, fmt.Sprintf("%s parameter %q content", group.location, parameter.Name))...)
		}
	}
	if request.Body != nil {
		gaps = append(gaps, r.schemaCycleGaps(request.Body.Schema, "request body schema")...)
	}
	for _, media := range request.Media {
		gaps = append(gaps, r.schemaCycleGaps(media.Schema, fmt.Sprintf("request media %q schema", media.MediaType))...)
	}
	return sourceSortedGaps(gaps)
}

func (r *sourceReferenceResolver) responseSchemaCycleGaps(responses []sourceResponseDescriptor) []sourceContractGap {
	var gaps []sourceContractGap
	for _, response := range responses {
		gaps = append(gaps, r.schemaCycleGaps(response.Declaration, fmt.Sprintf("response %s", response.Status))...)
	}
	return sourceSortedGaps(gaps)
}

func (r *sourceReferenceResolver) unreferencedSchemaCycleGaps(operations []sourceOperationDescriptor) []sourceContractGap {
	if len(r.schemaCycles) == 0 {
		return nil
	}
	used := map[string]bool{}
	for _, operation := range operations {
		for _, reference := range r.schemaCycleReferences(operation.Request) {
			used[reference] = true
		}
		for _, response := range operation.Responses {
			for _, reference := range r.schemaCycleReferences(response.Declaration) {
				used[reference] = true
			}
		}
	}
	var references []string
	for reference := range r.schemaCycles {
		if !used[reference] {
			references = append(references, reference)
		}
	}
	sort.Strings(references)
	gaps := make([]sourceContractGap, 0, len(references))
	for _, reference := range references {
		gaps = append(gaps, sourceSchemaCycleGap(reference, "source schema "+reference))
	}
	return gaps
}

func (r *sourceReferenceResolver) unreferencedSchemaReferenceSiblingGaps(operations []sourceOperationDescriptor) []sourceContractGap {
	if len(r.schemaReferenceSiblingGaps) == 0 {
		return nil
	}
	used := map[string]bool{}
	for _, operation := range operations {
		for _, gap := range operation.Runtime.Gaps {
			if gap.Foundation == sourceOpenAPI30ReferenceSiblingFoundation {
				used[sourceContractGapIdentity(gap)] = true
			}
		}
	}
	seen := map[string]bool{}
	gaps := make([]sourceContractGap, 0, len(r.schemaReferenceSiblingGaps))
	for _, gap := range r.schemaReferenceSiblingGaps {
		identity := sourceContractGapIdentity(gap)
		if used[identity] || seen[identity] {
			continue
		}
		seen[identity] = true
		gaps = append(gaps, gap)
	}
	return sourceSortedGaps(gaps)
}

func sourceContractGapIdentity(gap sourceContractGap) string {
	return gap.Foundation + "\x00" + gap.Location + "\x00" + gap.Reason
}

func (r *sourceReferenceResolver) schemaCycleGaps(value any, location string) []sourceContractGap {
	references := r.schemaCycleReferences(value)
	gaps := make([]sourceContractGap, 0, len(references))
	for _, reference := range references {
		gaps = append(gaps, sourceSchemaCycleGap(reference, location))
	}
	return gaps
}

func (r *sourceReferenceResolver) schemaCycleReferences(value any) []string {
	if len(r.schemaCycles) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var visit func(any)
	visit = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			if rawReference, exists := typed["$ref"]; exists {
				if raw, ok := rawReference.(string); ok {
					if reference, err := sourceNormalizeLocalReference(raw); err == nil {
						if _, isCycle := r.schemaCycles[reference]; isCycle {
							seen[reference] = true
						}
					}
				}
			}
			for _, key := range sortedSourceMapKeys(typed) {
				visit(typed[key])
			}
		case []any:
			for _, item := range typed {
				visit(item)
			}
		}
	}
	visit(value)
	references := make([]string, 0, len(seen))
	for reference := range seen {
		references = append(references, reference)
	}
	sort.Strings(references)
	return references
}

func (r *sourceReferenceResolver) knownSchemaCycleReference(schema map[string]any) bool {
	rawReference, exists := schema["$ref"]
	raw, ok := rawReference.(string)
	if !exists || !ok {
		return false
	}
	reference, err := sourceNormalizeLocalReference(raw)
	if err != nil {
		return false
	}
	_, isCycle := r.schemaCycles[reference]
	return isCycle
}

func sourceSchemaCycleGap(reference, location string) sourceContractGap {
	return sourceContractGapFor(sourceRecursiveSchemaFoundation, location, fmt.Sprintf("provider schema reference cycle at %s is retained without expansion", reference))
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

type sourceRequestSchemaDisposition string

const (
	sourceRequestRepresented                sourceRequestSchemaDisposition = "represented"
	sourceRequestRepresentedWithPolicyBound sourceRequestSchemaDisposition = "represented_with_policy_bound"
	sourceRequestSourceGap                  sourceRequestSchemaDisposition = "source_gap"
	sourceRequestFiniteBudgetGap            sourceRequestSchemaDisposition = "finite_budget_gap"
	sourceRequestMalformedSourceGap         sourceRequestSchemaDisposition = "malformed_source_gap"
)

type sourceRequestSchemaError struct {
	Disposition sourceRequestSchemaDisposition
	Reason      string
}

func (e *sourceRequestSchemaError) Error() string { return e.Reason }

func sourceRequestPolicyBoundError(reason string) error {
	return &sourceRequestSchemaError{Disposition: sourceRequestRepresentedWithPolicyBound, Reason: reason}
}

func sourceRequestSchemaDispositionOf(err error) sourceRequestSchemaDisposition {
	if err == nil {
		return sourceRequestRepresented
	}
	var typed *sourceRequestSchemaError
	if errors.As(err, &typed) {
		return typed.Disposition
	}
	return sourceRequestMalformedSourceGap
}

func validateBoundedRequestSchema(schema any, form sourceDocumentForm, limits sourceImportLimits, depth int) error {
	strict := limits
	strict.AllowSourceContractGaps = false
	err := validateBoundedRequestSchemaWithinEnum(schema, form, strict, depth, false)
	if err != nil && limits.AllowSourceContractGaps {
		return nil
	}
	return err
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
		if err := sourceValidateLengthBounds(object, bounded, limits.UseExecutionEnvelopes); err != nil {
			return err
		}
	case "integer", "number":
		if err := sourceValidateNumericBounds(object, form, !bounded, limits.UseExecutionEnvelopes); err != nil {
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
		if err := sourceValidateArrayBounds(object, bounded, closedTupleLimit, limits.UseExecutionEnvelopes); err != nil {
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

func sourceValidateLengthBounds(object map[string]any, finiteEnum, policyBounded bool) error {
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
	if !finiteEnum && !hasMaximum && !policyBounded {
		return sourceRequestPolicyBoundError("unbounded request schema string has no maxLength")
	}
	return nil
}

func sourceValidateArrayBounds(object map[string]any, finiteEnum bool, closedTupleLimit *int64, policyBounded bool) error {
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
	if !finiteEnum && !hasMaximum && !policyBounded {
		return sourceRequestPolicyBoundError("unbounded request schema array has no maxItems")
	}
	return nil
}

type sourceNumericBound struct {
	value     *big.Rat
	inclusive bool
	present   bool
}

func sourceValidateNumericBounds(object map[string]any, form sourceDocumentForm, requireBoth, policyBounded bool) error {
	lower, err := sourceNumericBoundFor(object, form, "minimum", "exclusiveMinimum", true)
	if err != nil {
		return err
	}
	upper, err := sourceNumericBoundFor(object, form, "maximum", "exclusiveMaximum", false)
	if err != nil {
		return err
	}
	if requireBoth && !lower.present && !policyBounded {
		return sourceRequestPolicyBoundError("unbounded request schema number has no minimum")
	}
	if requireBoth && !upper.present && !policyBounded {
		return sourceRequestPolicyBoundError("unbounded request schema number has no maximum")
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
		if strings.ContainsAny(value[index+1:], "eE") {
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

func sourceImportProvenance(documentContext sourceImportDocumentContext, form sourceDocumentForm, location string) sourceImportSource {
	source := sourceImportSource{
		URL:      documentContext.Artifact.SourceURL,
		SHA256:   strings.ToLower(documentContext.Artifact.SHA256),
		Bytes:    documentContext.Artifact.Bytes,
		Location: location,
		Form:     form.Family,
		Version:  form.Version,
	}
	if documentContext.Document != nil {
		source.DocumentID = documentContext.Document.ID
		source.PublishedURL = documentContext.Document.PublishedSource.SourceURL
		source.PublishedCaptureURL = documentContext.Document.PublishedSource.CaptureURL
		source.PublishedSHA256 = strings.ToLower(documentContext.Document.PublishedSource.SHA256)
		source.PublishedBytes = documentContext.Document.PublishedSource.Bytes
		source.PublishedAdapter = documentContext.Document.PublishedSource.Adapter
	}
	return source
}

func sourceWebhookEvents(documentContext sourceImportDocumentContext, doc map[string]any, form sourceDocumentForm, resolver *sourceReferenceResolver) ([]sourceInboundEventDescriptor, error) {
	lock := documentContext.Lock
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
	events := []sourceInboundEventDescriptor{}
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
			Source:      sourceImportProvenance(documentContext, form, location),
			Declaration: declaration,
			Runtime:     sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{gap}},
		})
	}
	sortSourceInboundEventDescriptors(events)
	return events, nil
}

func sourceCallbackEvents(documentContext sourceImportDocumentContext, form sourceDocumentForm, parentSourceID, location string, raw any, resolver *sourceReferenceResolver) ([]sourceInboundEventDescriptor, error) {
	lock := documentContext.Lock
	callbacks, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("callbacks must be an object")
	}
	events := []sourceInboundEventDescriptor{}
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
			Source:         sourceImportProvenance(documentContext, form, eventLocation),
			Declaration:    declaration,
			Runtime:        sourceRuntimeReachability{MergeBlocked: true, Gaps: []sourceContractGap{gap}},
		})
	}
	sortSourceInboundEventDescriptors(events)
	return events, nil
}

func sourceNonExtensionEntryCount(values map[string]any) int {
	count := 0
	for key := range values {
		if !strings.HasPrefix(key, "x-") {
			count++
		}
	}
	return count
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
	mediaType = sourceNormalizedMediaType(mediaType)
	return mediaType == "application/json" || strings.HasSuffix(mediaType, "+json")
}

func sourceRequestFormMediaType(mediaType string) bool {
	mediaType = sourceNormalizedMediaType(mediaType)
	return mediaType == "application/x-www-form-urlencoded" || (strings.HasPrefix(mediaType, "multipart/") && len(mediaType) > len("multipart/"))
}

func sourceNormalizedMediaType(mediaType string) string {
	base, _, _ := strings.Cut(mediaType, ";")
	return strings.ToLower(strings.TrimSpace(base))
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
		DescriptorSchemaVersion: result.DescriptorSchemaVersion,
		Operations:              append([]sourceOperationDescriptor(nil), result.Operations...),
		InboundEvents:           append([]sourceInboundEventDescriptor(nil), result.InboundEvents...),
		Extensions:              append([]sourceExtensionDescriptor(nil), result.Extensions...),
		GraphQLSchemas:          append([]sourceGraphQLSchemaDescriptor(nil), result.GraphQLSchemas...),
		Gaps:                    append([]sourceContractGap(nil), result.Gaps...),
	}
	sortSourceOperationDescriptors(copyResult.Operations)
	sortSourceInboundEventDescriptors(copyResult.InboundEvents)
	sortSourceExtensions(copyResult.Extensions)
	gaps := append([]sourceContractGap(nil), copyResult.Gaps...)
	for _, operation := range copyResult.Operations {
		gaps = append(gaps, operation.Runtime.Gaps...)
	}
	for _, event := range copyResult.InboundEvents {
		gaps = append(gaps, event.Runtime.Gaps...)
	}
	gaps = sourceSortedGaps(gaps)
	schemaVersion := 2
	if copyResult.DescriptorSchemaVersion != 0 {
		schemaVersion = copyResult.DescriptorSchemaVersion
	}
	document := sourceImportDescriptorDocument{
		SchemaVersion:  schemaVersion,
		Operations:     copyResult.Operations,
		InboundEvents:  copyResult.InboundEvents,
		Extensions:     copyResult.Extensions,
		GraphQLSchemas: copyResult.GraphQLSchemas,
		MergeBlocked:   len(gaps) > 0,
		Gaps:           gaps,
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
	CacheDir  string
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
	if opts.CacheDir != "" {
		cacheDir, cacheErr := sourceImportArtifactCacheDir(opts.CacheDir)
		if cacheErr != nil {
			logln(stderr, "connectorgen source-import: resolve explicit artifact cache:", cacheErr)
			return 1
		}
		fetcher = newSourceImportArtifactCacheFetcher(fetcher, cacheDir, defaultSourceImportLimits())
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
	surface, err := sourceProjectionExecutionSurface(filepath.Join(opts.DefsDir, opts.Connector), opts.Connector)
	if err != nil {
		logln(stderr, "connectorgen source-import: load declaration-owned execution surface:", err)
		return 1
	}
	sourceProjectionNormalizeNonBlockingReadGaps(&result)
	sourceProjectionRestoreSourceBoundDirectReadPathFlags(&surface, result)
	sourceProjectionAnnotateUnreachableReadGaps(surface, &result)
	raw, err := marshalSourceImportResult(result)
	if err != nil {
		logln(stderr, "connectorgen source-import: encode descriptors:", err)
		return 1
	}
	existing, readErr := os.ReadFile(opts.Output)
	if opts.Check {
		projection, projectionErr := projectSourceDescriptorToBundle(filepath.Join(opts.DefsDir, opts.Connector), result, true)
		if projectionErr != nil {
			logln(stderr, "connectorgen source-import: check source-derived bundle projection:", projectionErr)
			return 1
		}
		if readErr != nil || !bytes.Equal(existing, raw) || projection.Changed() {
			logln(stderr, fmt.Sprintf("connectorgen source-import: descriptor or derived bundle projection has drifted (writes=%d cli=%d); rerun without --check after source-lock verification", projection.Writes, projection.CLI))
			return 1
		}
		logln(stdout, fmt.Sprintf("connectorgen source-import: %s, %d operation(s), %d inbound event(s) verified", opts.Connector, len(result.Operations), len(result.InboundEvents)))
		return 0
	}
	if err := os.WriteFile(opts.Output, raw, 0o644); err != nil {
		logln(stderr, "connectorgen source-import: write descriptors:", err)
		return 1
	}
	projection, err := projectSourceDescriptorToBundle(filepath.Join(opts.DefsDir, opts.Connector), result, false)
	if err != nil {
		logln(stderr, "connectorgen source-import: project source-derived bundle:", err)
		return 1
	}
	logln(stdout, fmt.Sprintf("connectorgen source-import: %s, %d operation(s), %d inbound event(s) imported; source projection updated writes=%d cli=%d", opts.Connector, len(result.Operations), len(result.InboundEvents), projection.Writes, projection.CLI))
	return 0
}

func parseSourceImportOptions(args []string) (sourceImportOptions, error) {
	opts := sourceImportOptions{DefsDir: filepath.Join("internal", "connectors", "defs")}
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--check":
			opts.Check = true
		case "--defs", "--out", "--cache-dir":
			if i+1 >= len(args) {
				return sourceImportOptions{}, fmt.Errorf("%s requires a value", arg)
			}
			i++
			switch arg {
			case "--defs":
				opts.DefsDir = args[i]
			case "--out":
				opts.Output = args[i]
			case "--cache-dir":
				opts.CacheDir = args[i]
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
		opts.Output = filepath.Join(opts.DefsDir, opts.Connector, "sources", opts.Connector+"-operation-descriptor.json")
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
	client *http.Client
}

func newHTTPSourceImportFetcher(limits sourceImportLimits) sourceImportFetcher {
	cacheDir, cacheErr := sourceImportArtifactCacheDir("")
	return newSourceImportArtifactCacheFetcher(
		httpSourceImportFetcher{limits: limits, lookup: batchArtifactLookupIPAddr(net.DefaultResolver.LookupIPAddr)},
		cacheDir,
		limits,
		cacheErr,
	)
}

type sourceImportArtifactCacheFetcher struct {
	source   sourceImportFetcher
	cacheDir string
	limits   sourceImportLimits
	cacheErr error
}

func newSourceImportArtifactCacheFetcher(source sourceImportFetcher, cacheDir string, limits sourceImportLimits, cacheErr ...error) sourceImportArtifactCacheFetcher {
	fetcher := sourceImportArtifactCacheFetcher{source: source, cacheDir: cacheDir, limits: limits}
	if len(cacheErr) > 0 {
		fetcher.cacheErr = cacheErr[0]
	}
	return fetcher
}

func (fetcher sourceImportArtifactCacheFetcher) Fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	if fetcher.source == nil {
		return nil, fmt.Errorf("source importer has no fetcher")
	}
	return fetcher.source.Fetch(ctx, sourceURL)
}

// FetchArtifact returns only bytes whose size and digest match the immutable
// lock. A missing, stale, or corrupt cache is never returned; it is replaced
// only by a fresh response that passes the same verification.
func (fetcher sourceImportArtifactCacheFetcher) FetchArtifact(ctx context.Context, artifact sourceImportArtifact) ([]byte, error) {
	if err := validateSourceImportLimits(fetcher.limits); err != nil {
		return nil, err
	}
	if fetcher.source == nil {
		return nil, fmt.Errorf("source importer has no fetcher")
	}
	if fetcher.cacheErr != nil {
		return nil, fmt.Errorf("resolve source-import artifact cache: %w", fetcher.cacheErr)
	}
	cachePath, err := sourceImportArtifactCachePath(fetcher.cacheDir, artifact)
	if err != nil {
		return nil, err
	}
	if cached, present, cacheErr := readSourceImportArtifactCache(cachePath, fetcher.limits.MaxArtifactBytes); cacheErr == nil && present {
		if err := validateSourceImportArtifactBytes(cached, artifact); err == nil {
			return cached, nil
		}
		if err := removeSourceImportArtifactCache(cachePath); err != nil {
			return nil, err
		}
	} else if cacheErr != nil {
		if err := removeSourceImportArtifactCache(cachePath); err != nil {
			return nil, err
		}
	}

	raw, err := fetchSourceImportArtifact(ctx, fetcher.source, artifact)
	if err != nil {
		return nil, err
	}
	if int64(len(raw)) > fetcher.limits.MaxArtifactBytes {
		return nil, fmt.Errorf("artifact byte limit exceeded")
	}
	if err := validateSourceImportArtifactBytes(raw, artifact); err != nil {
		return nil, err
	}
	if err := writeSourceImportArtifactCache(cachePath, raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func sourceImportArtifactCacheDir(explicitRoot string) (string, error) {
	if explicitRoot != "" {
		absoluteRoot, err := filepath.Abs(explicitRoot)
		if err != nil {
			return "", fmt.Errorf("resolve explicit source-import artifact cache directory: %w", err)
		}
		info, err := os.Lstat(absoluteRoot)
		if err != nil {
			return "", fmt.Errorf("explicit source-import artifact cache directory must exist: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return "", fmt.Errorf("explicit source-import artifact cache directory must be an existing non-symlink directory")
		}
		return absoluteRoot, nil
	}
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	if root == "" {
		return "", fmt.Errorf("user cache directory is empty")
	}
	return filepath.Join(root, "polymetrics", "source-import"), nil
}

func sourceImportArtifactCachePath(cacheDir string, artifact sourceImportArtifact) (string, error) {
	if cacheDir == "" {
		return "", fmt.Errorf("source-import artifact cache directory is empty")
	}
	if err := validateSourceImportArtifact(artifact); err != nil {
		return "", err
	}
	return filepath.Join(cacheDir, strings.ToLower(artifact.SHA256)+".artifact"), nil
}

func readSourceImportArtifactCache(path string, maxBytes int64) ([]byte, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, true, fmt.Errorf("open source-import artifact cache: %w", err)
	}
	raw, readErr := io.ReadAll(io.LimitReader(file, maxBytes+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, true, fmt.Errorf("read source-import artifact cache: %w", readErr)
	}
	if closeErr != nil {
		return nil, true, fmt.Errorf("close source-import artifact cache: %w", closeErr)
	}
	if int64(len(raw)) > maxBytes {
		return nil, true, fmt.Errorf("source-import artifact cache exceeds byte limit")
	}
	return raw, true, nil
}

func removeSourceImportArtifactCache(path string) error {
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove invalid source-import artifact cache: %w", err)
	}
	return nil
}

func writeSourceImportArtifactCache(path string, raw []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create source-import artifact cache directory: %w", err)
	}
	temporary, err := os.CreateTemp(dir, ".source-import-artifact-")
	if err != nil {
		return fmt.Errorf("create source-import artifact cache file: %w", err)
	}
	temporaryPath := temporary.Name()
	if err := temporary.Chmod(0o600); err != nil {
		return sourceImportArtifactCacheTemporaryFailure("set source-import artifact cache permissions", err, temporary, temporaryPath)
	}
	if _, err := temporary.Write(raw); err != nil {
		return sourceImportArtifactCacheTemporaryFailure("write source-import artifact cache", err, temporary, temporaryPath)
	}
	if err := temporary.Sync(); err != nil {
		return sourceImportArtifactCacheTemporaryFailure("sync source-import artifact cache", err, temporary, temporaryPath)
	}
	if err := temporary.Close(); err != nil {
		return sourceImportArtifactCacheTemporaryFailure("close source-import artifact cache", err, temporary, temporaryPath)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		if cleanupErr := removeSourceImportArtifactCache(temporaryPath); cleanupErr != nil {
			return fmt.Errorf("publish source-import artifact cache: %w", errors.Join(err, cleanupErr))
		}
		return fmt.Errorf("publish source-import artifact cache: %w", err)
	}
	return nil
}

func sourceImportArtifactCacheTemporaryFailure(operation string, cause error, temporary *os.File, temporaryPath string) error {
	var cleanupErrs []error
	if err := temporary.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
		cleanupErrs = append(cleanupErrs, fmt.Errorf("close source-import artifact cache: %w", err))
	}
	if err := removeSourceImportArtifactCache(temporaryPath); err != nil {
		cleanupErrs = append(cleanupErrs, err)
	}
	if len(cleanupErrs) > 0 {
		return fmt.Errorf("%s: %w", operation, errors.Join(append([]error{cause}, cleanupErrs...)...))
	}
	return fmt.Errorf("%s: %w", operation, cause)
}

func validateSourceImportArtifactBytes(raw []byte, artifact sourceImportArtifact) error {
	digest := sha256.Sum256(raw)
	if int64(len(raw)) != artifact.Bytes || !strings.EqualFold(hex.EncodeToString(digest[:]), artifact.SHA256) {
		return fmt.Errorf("source-lock refresh required: fetched artifact does not match locked bytes and SHA-256")
	}
	return nil
}

func (fetcher httpSourceImportFetcher) Fetch(ctx context.Context, sourceURL string) ([]byte, error) {
	return fetcher.fetch(ctx, sourceURL, batchArtifactURLPolicy{})
}

func (fetcher httpSourceImportFetcher) FetchArtifact(ctx context.Context, artifact sourceImportArtifact) ([]byte, error) {
	if err := validateSourceImportArtifact(artifact); err != nil {
		return nil, err
	}
	return fetcher.fetch(ctx, artifact.SourceURL, batchArtifactURLPolicy{allowIdentityQuery: artifact.IdentityQuery})
}

func (fetcher httpSourceImportFetcher) fetch(ctx context.Context, sourceURL string, policy batchArtifactURLPolicy) ([]byte, error) {
	if err := validateSourceImportLimits(fetcher.limits); err != nil {
		return nil, err
	}
	if fetcher.lookup == nil {
		return nil, fmt.Errorf("source importer has no public address resolver")
	}
	parsed, err := parseBatchArtifactURLWithPolicy(sourceURL, policy)
	if err != nil {
		return nil, fmt.Errorf("validate locked source artifact URL: %w", err)
	}
	if err := validateBatchArtifactRequestURLWithPolicy(ctx, parsed, fetcher.lookup, policy); err != nil {
		return nil, fmt.Errorf("validate locked source artifact destination: %w", err)
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	client := fetcher.client
	if client == nil {
		client = newSourceImportHTTPClient(fetcher.lookup)
	}
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
	client.Timeout = defaultSourceImportFetchTimeout
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return fmt.Errorf("source-lock artifact redirects are not permitted")
	}
	return client
}
