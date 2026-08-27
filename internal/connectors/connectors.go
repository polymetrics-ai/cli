package connectors

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf16"
	"unicode/utf8"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/synccontract"
	"polymetrics.ai/internal/warehouse"
)

var ErrUnsupportedOperation = errors.New("unsupported connector operation")
var ErrOpaqueCursorOrderUnavailable = errors.New("opaque cursor order is unavailable")

// AuthenticationAdmission is the production request boundary installed by
// app.Open. Implementations admit against durable cohort health before running
// operation and persist only explicitly verified authentication failures.
type AuthenticationAdmission interface {
	Execute(context.Context, func(context.Context) error) error
}

// AuthenticationFailureClassifier is implemented by a connector that can
// prove an error is an authentication failure from its declared provider
// contract or native protocol code. It must never classify by error text.
type AuthenticationFailureClassifier interface {
	AuthenticationFailureVerified(error) bool
}

// RateParkingAdmission rejects a provider scope while durable parked work is
// waiting to resume. The scope is already an opaque project-local projection.
type RateParkingAdmission interface {
	Admit(RateLimitScopeKey) error
}

// VerifiedAuthenticationError marks a failure classified by a connector's
// declared error map or native protocol code. Generic 401s and text matching
// must never construct this type.
type VerifiedAuthenticationError struct {
	Err error
}

func (e *VerifiedAuthenticationError) Error() string {
	if e == nil || e.Err == nil {
		return "verified authentication failure"
	}
	return e.Err.Error()
}

func (e *VerifiedAuthenticationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func MarkVerifiedAuthenticationFailure(err error) error {
	if err == nil {
		return nil
	}
	var verified *VerifiedAuthenticationError
	if errors.As(err, &verified) {
		return err
	}
	return &VerifiedAuthenticationError{Err: err}
}

func IsVerifiedAuthenticationFailure(err error) bool {
	var verified *VerifiedAuthenticationError
	return errors.As(err, &verified)
}

// MarkConnectorAuthenticationFailure applies only a connector-owned typed
// classifier. It lets closed transport executors report through the same
// production admission runtime as ordinary Connector.Read/Write calls.
func MarkConnectorAuthenticationFailure(connector Connector, err error) error {
	if err == nil || IsVerifiedAuthenticationFailure(err) {
		return err
	}
	classifier, ok := connector.(AuthenticationFailureClassifier)
	if ok && classifier.AuthenticationFailureVerified(err) {
		return MarkVerifiedAuthenticationFailure(err)
	}
	return err
}

// GroupRateLimitScopeKeys returns one opaque admission key for every policy
// governing a physical request. One-policy routes preserve their existing key;
// multi-policy routes use a one-way digest so parking blocks the exact group
// without exposing any member scope.
func GroupRateLimitScopeKeys(scopes ...RateLimitScopeKey) (RateLimitScopeKey, error) {
	if len(scopes) == 0 {
		return "", errors.New("rate-limit scope group is unavailable")
	}
	values := make([]string, len(scopes))
	for index, scope := range scopes {
		if scope == "" {
			return "", errors.New("rate-limit scope group contains an empty scope")
		}
		values[index] = string(scope)
	}
	sort.Strings(values)
	values = compactStrings(values)
	if len(values) == 1 {
		return RateLimitScopeKey(values[0]), nil
	}
	digest := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return RateLimitScopeKey(fmt.Sprintf("rl-group-v1-%x", digest[:])), nil
}

func compactStrings(values []string) []string {
	if len(values) < 2 {
		return values
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

// RateLimitParkingScopeResolver converts a terminal typed rate-limit result
// into the same opaque scope group used at the physical send admission path.
type RateLimitParkingScopeResolver interface {
	RateLimitParkingScope(context.Context, RuntimeConfig, string, error) (RateLimitScopeKey, error)
}

var defaultRegistryBuilder = struct {
	mu sync.RWMutex
	fn func() *Registry
}{}

// RegisterDefaultRegistryBuilder installs the process default registry builder.
// Wave 6 uses this to let the bundle-backed registry live in a package that can
// import engine/defs without creating a connectors<->engine cycle.
func RegisterDefaultRegistryBuilder(fn func() *Registry) {
	defaultRegistryBuilder.mu.Lock()
	defer defaultRegistryBuilder.mu.Unlock()
	defaultRegistryBuilder.fn = fn
}

func registeredDefaultRegistryBuilder() func() *Registry {
	defaultRegistryBuilder.mu.RLock()
	defer defaultRegistryBuilder.mu.RUnlock()
	return defaultRegistryBuilder.fn
}

type Record map[string]any

type Capabilities struct {
	Check   bool `json:"check"`
	Catalog bool `json:"catalog"`
	Read    bool `json:"read"`
	Write   bool `json:"write"`
	Query   bool `json:"query"`
	CDC     bool `json:"cdc,omitempty"`
}

type Metadata struct {
	Name            string         `json:"name"`
	DisplayName     string         `json:"display_name"`
	IntegrationType string         `json:"integration_type"`
	Description     string         `json:"description"`
	Capabilities    Capabilities   `json:"capabilities"`
	Icon            *ConnectorIcon `json:"icon,omitempty"`
}

type Field struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Stream struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Fields       []Field  `json:"fields"`
	PrimaryKey   []string `json:"primary_key"`
	CursorFields []string `json:"cursor_fields"`
	// Schema is the verbatim draft-07 record schema backing this catalog
	// stream. Static bundles and provider discovery both pass through
	// StreamFromSchema so consumers never need a source-specific schema path.
	Schema json.RawMessage `json:"schema,omitempty"`
}

type Catalog struct {
	Connector string           `json:"connector"`
	Streams   []Stream         `json:"streams"`
	Discovery *DiscoveryStatus `json:"discovery,omitempty"`
}

// DiscoveryStatus makes a provider-derived catalog's freshness and completeness
// explicit. It intentionally carries only safe object names and lifecycle
// facts: provider response bodies and credentials never belong in a catalog.
type DiscoveryStatus struct {
	Complete     bool               `json:"complete"`
	Cached       bool               `json:"cached,omitempty"`
	Stale        bool               `json:"stale,omitempty"`
	UsedFallback bool               `json:"used_fallback,omitempty"`
	RefreshedAt  time.Time          `json:"refreshed_at,omitempty"`
	ExpiresAt    time.Time          `json:"expires_at,omitempty"`
	Failures     []DiscoveryFailure `json:"failures,omitempty"`
}

// DiscoveryFailure records the scope of an incomplete discovery without
// retaining an upstream error, which can contain provider-supplied data.
type DiscoveryFailure struct {
	Object   string `json:"object,omitempty"`
	Stage    string `json:"stage"`
	Attempts int    `json:"attempts,omitempty"`
}

// StreamFromSchema constructs the public catalog projection of one draft-07
// record schema. It is the shared boundary for static bundle schemas and
// dynamically discovered schemas: fields, primary key, cursor, and raw schema
// must therefore agree regardless of where the schema came from.
func StreamFromSchema(name, description string, raw json.RawMessage) (Stream, error) {
	if strings.TrimSpace(name) == "" {
		return Stream{}, errors.New("catalog stream name is required")
	}
	if len(raw) == 0 {
		return Stream{}, fmt.Errorf("catalog stream %q schema is required", name)
	}

	var doc struct {
		Type        string                     `json:"type"`
		Properties  map[string]json.RawMessage `json:"properties"`
		PrimaryKey  []string                   `json:"x-primary-key"`
		CursorField string                     `json:"x-cursor-field"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return Stream{}, fmt.Errorf("catalog stream %q schema: %w", name, err)
	}
	if doc.Type != "object" {
		return Stream{}, fmt.Errorf("catalog stream %q schema type must be object", name)
	}

	stream := Stream{
		Name:        name,
		Description: description,
		PrimaryKey:  append([]string(nil), doc.PrimaryKey...),
		Schema:      append(json.RawMessage(nil), raw...),
	}
	if doc.CursorField != "" {
		stream.CursorFields = []string{doc.CursorField}
	}

	fieldNames := make([]string, 0, len(doc.Properties))
	for fieldName := range doc.Properties {
		fieldNames = append(fieldNames, fieldName)
	}
	sort.Strings(fieldNames)
	for _, fieldName := range fieldNames {
		fieldType, err := catalogFieldType(doc.Properties[fieldName])
		if err != nil {
			return Stream{}, fmt.Errorf("catalog stream %q field %q: %w", name, fieldName, err)
		}
		stream.Fields = append(stream.Fields, Field{Name: fieldName, Type: fieldType})
	}

	if err := validateCatalogSchemaContract(stream); err != nil {
		return Stream{}, err
	}
	return stream, nil
}

// catalogFieldType preserves the conventional Field.Type projection while the
// raw schema remains authoritative. A nullable schema reports its concrete
// non-null type; an omitted type stays empty rather than being guessed.
func catalogFieldType(raw json.RawMessage) (string, error) {
	var property struct {
		Type json.RawMessage `json:"type"`
	}
	if err := json.Unmarshal(raw, &property); err != nil {
		return "", fmt.Errorf("invalid property schema: %w", err)
	}
	if len(property.Type) == 0 || string(property.Type) == "null" {
		return "", nil
	}

	var single string
	if err := json.Unmarshal(property.Type, &single); err == nil {
		return single, nil
	}

	var union []string
	if err := json.Unmarshal(property.Type, &union); err != nil {
		return "", fmt.Errorf("type must be a string or array of strings")
	}
	for _, candidate := range union {
		if candidate != "null" {
			return candidate, nil
		}
	}
	return "null", nil
}

func validateCatalogSchemaContract(stream Stream) error {
	known := make(map[string]struct{}, len(stream.Fields))
	for _, field := range stream.Fields {
		known[field.Name] = struct{}{}
	}
	for _, field := range stream.PrimaryKey {
		if _, ok := known[field]; !ok {
			return fmt.Errorf("catalog stream %q primary key field %q is absent from schema", stream.Name, field)
		}
	}
	for _, field := range stream.CursorFields {
		if _, ok := known[field]; !ok {
			return fmt.Errorf("catalog stream %q cursor field %q is absent from schema", stream.Name, field)
		}
	}
	return nil
}

// SecretStore writes a single rotated secret back to the caller's credential
// store. It exists for provider-rotated credentials — an OAuth2 refresh token
// that the provider replaces on every exchange and invalidates the old value
// of, so dropping the new one silently breaks the connector on its next run.
//
// It is deliberately narrow. There is no Get (the current values already
// arrive in RuntimeConfig.Secrets), no Delete, and no enumeration: nothing here
// gives a connector a way to read secrets it was not given, or to reach a
// credential other than its own.
//
// Implementations must persist encrypted and locally — the CLI's is backed by
// internal/vault (AES-256-GCM under .polymetrics/vault) — must be safe for
// concurrent use, and must never place a secret value in an error string.
//
// A nil SecretStore is valid and means "this caller has no credential store".
// Rotation is then held in memory for the process lifetime; it is never
// downgraded to a plaintext write.
type SecretStore interface {
	PutSecret(ctx context.Context, key, value string) error
}

type RuntimeConfig struct {
	ProjectDir string            `json:"-"`
	Config     map[string]string `json:"config"`
	Secrets    map[string]string `json:"-"`
	// CoordinationIdentity carries opaque auth/rate inputs only. It has no
	// credential, binding, provider, profile, or approval-revision preimage.
	CoordinationIdentity  CoordinationIdentity `json:"-"`
	CredentialRevision    string               `json:"-"`
	ConfigurationDigest   string               `json:"-"`
	WriteApprovalScope    string               `json:"-"`
	ApprovedPayloadSHA256 map[string]string    `json:"-"`
	// ForceCatalogRefresh bypasses an in-process discovery cache. App-level
	// `pm catalog refresh` sets it deliberately; it is never persisted or
	// inferred from credentials.
	ForceCatalogRefresh bool `json:"-"`
	// ResolvedCatalog is an already durable, account-scoped catalog supplied
	// by the application to a connector invocation. It prevents a reader from
	// independently rediscovering an account on every command. It contains
	// provider schema metadata only, never connection configuration or secrets.
	ResolvedCatalog *Catalog `json:"-"`
	// SecretStore, when set, persists a provider-rotated secret back to the
	// caller's encrypted credential store. Optional; see SecretStore.
	SecretStore SecretStore `json:"-"`
	// AuthenticationAdmission and RateParkingAdmission are installed only by
	// the production composition root. They contain no credential material.
	AuthenticationAdmission AuthenticationAdmission `json:"-"`
	RateParkingAdmission    RateParkingAdmission    `json:"-"`
	// BudgetCoordinator owns the opaque lifecycle reservation for a declared
	// physical request. It receives only declaration-owned policy data and
	// opaque scopes; it never receives request, response, or credential values.
	BudgetCoordinator connsdk.BudgetCoordinator `json:"-"`
}

// PayloadApprovalKey identifies a file field within an approved write batch.
func PayloadApprovalKey(recordIndex int, field string) string {
	return fmt.Sprintf("%d:%s", recordIndex, field)
}

type OpaqueCursorState struct {
	Token   []byte
	Present bool
}

// ReadContinuation is engine-owned pagination state for a bounded source
// scan. Its bytes are opaque outside the connector engine: callers may retain
// and return an exact value through a durable checkpoint, but cannot select a
// provider cursor, URL, query, or pagination strategy with it.
type ReadContinuation struct {
	Kind  string
	Token []byte
}

// Clone preserves the exact opaque continuation bytes across request and
// result boundaries.
func (c *ReadContinuation) Clone() *ReadContinuation {
	if c == nil {
		return nil
	}
	clone := *c
	clone.Token = append([]byte(nil), c.Token...)
	return &clone
}

// ReadBudgetStoppedError reports that an engine scan reached its declared
// page budget while a further provider page was known to exist. It is not EOF
// and cannot be treated as a completed collection.
type ReadBudgetStoppedError struct {
	Continuation ReadContinuation
}

func (e *ReadBudgetStoppedError) Error() string {
	if e == nil {
		return "source pagination stopped at its page budget"
	}
	return "source pagination stopped at its page budget before exhaustion"
}

type SourceOrderedCursorReader interface {
	CursorStateFromRecord(Record, string) (OpaqueCursorState, error)
	ValidateCursorField(RuntimeConfig, string) error
	CompareCursorStates(OpaqueCursorState, OpaqueCursorState) (int, error)
}

type ReadRequest struct {
	Stream      string
	Config      RuntimeConfig
	State       map[string]string
	CursorState OpaqueCursorState
	Query       map[string]string
	Limit       int
	// MaxPages is an optional caller-side request cap. Positive values only
	// tighten a declared stream limit; zero leaves it unchanged and a negative
	// value is rejected.
	MaxPages int
	// Continuation is accepted only by an engine-owned bounded-source resume.
	// It is deliberately not mapped from command input or connector config.
	Continuation *ReadContinuation
	// PageDeadline bounds one retryable provider page request. It is used by
	// closed transports and never represents a deadline for an entire stream.
	PageDeadline time.Duration
	// ObservePageFetch receives only an elapsed duration after a provider page
	// request returns. It carries no route, headers, payload, or credentials.
	ObservePageFetch func(time.Duration)
}

type DirectReadRequest struct {
	Method       string
	Path         string
	Config       RuntimeConfig
	PathParams   map[string]string
	Query        map[string]string
	MaxBytes     int
	OutputPolicy string
	RedactFields []string
	// Page selects an addressable page (page_number/offset_limit strategies);
	// PageCursor selects one by opaque token. They are mutually exclusive, and
	// both are answered by the previous read's DirectReadPage.
	Page       int
	PageCursor string
}

type OperationDirectReadRequest struct {
	Operation  string
	Config     RuntimeConfig
	PathParams map[string]string
	Query      map[string]string
	// CommandBindings seals the exact caller-controlled fields declared by the
	// generated command descriptor. The engine revalidates this set against the
	// loaded bundle before using it, so legacy descriptors can remain executable
	// while undeclared direct callers stay closed.
	CommandBindings *OperationDirectReadBindings
	// Headers contains only values for exact, provider-declared non-auth
	// header parameters. The operation engine validates the declaration and
	// values before constructing a runtime or issuing I/O.
	Headers      map[string]string
	HeaderValues map[string][]string
	Body         map[string]any
	// RawBody is available only to an operation that explicitly declares a
	// text/plain POST with a root string body schema. It is a pointer so an
	// absent body is distinct from an intentionally empty one; the operation
	// schema decides whether either is valid.
	RawBody      *string
	MaxBytes     int
	OutputPolicy string
	RedactFields []string
	// Page and PageCursor mirror DirectReadRequest's navigation inputs.
	Page       int
	PageCursor string
}

// OperationDirectReadBindings is the closed command-to-operation projection
// for one direct read. It contains field names only; values remain in the
// request maps and are independently type/size validated before dispatch.
type OperationDirectReadBindings struct {
	Path    []string
	Query   []string
	Body    []string
	RawBody bool
}

type DirectReadResult struct {
	Connector string `json:"connector"`
	Operation string `json:"operation,omitempty"`
	Method    string `json:"method"`
	Path      string `json:"path"`
	Status    int    `json:"status"`
	Body      any    `json:"body"`
	// Receipt is the complete bounded provider response. Body remains the
	// declaration's convenience projection; Receipt retains the raw transport
	// representation and every ordinary response header for audit/retry work.
	Receipt *ProviderResponseReceipt `json:"receipt,omitempty"`
	// Headers contains only response headers explicitly admitted by the exact
	// operation response contract. Declared provider values are preserved
	// verbatim unless a declaration-owned output-secret field classifies them;
	// arbitrary provider metadata is never an output channel.
	Headers map[string]OperationResponseHeader `json:"headers,omitempty"`
	// GraphQL is present only for a declared fixed-document GraphQL operation.
	// It deliberately exposes a small, redacted response summary rather than
	// a provider error envelope, so a provider cannot turn errors/extensions
	// into a secret-bearing output side channel.
	GraphQL *GraphQLResponseMetadata `json:"graphql,omitempty"`
	// OutputSecretFields is declaration-owned response sensitivity metadata.
	// It never leaves the internal connector/runner boundary; the public
	// projection replaces only the matching scalar values while retaining keys.
	OutputSecretFields []string `json:"-"`
	// Page answers "is this all of it?" — the question a direct-read caller
	// previously had no way to ask. A single Status+Body cannot represent
	// page 2..N, so before this field a truncated collection and a complete
	// one were indistinguishable at status 200.
	Page DirectReadPage `json:"page"`
}

// ProviderResponseReceipt is shared by direct reads and binary downloads.
// Binary success never inlines file bytes: Body is confined-file metadata and
// BodyRaw remains empty while BodyBytes reports the streamed byte count.
type ProviderResponseReceipt struct {
	ResponseReceived bool                               `json:"response_received"`
	Status           int                                `json:"status"`
	Headers          map[string]OperationResponseHeader `json:"headers,omitempty"`
	BodyPresent      bool                               `json:"body_present"`
	BodyBytes        int64                              `json:"body_bytes"`
	BodyRaw          string                             `json:"body_raw,omitempty"`
	BodyRawEncoding  string                             `json:"body_raw_encoding,omitempty"`
	Body             any                                `json:"body,omitempty"`
}

// OperationResponseHeader is one declared provider response header. Values are
// complete when ordinary. Redacted and Masked preserve compatibility for
// declaration-owned output-secret fields only.
type OperationResponseHeader struct {
	Values   []string `json:"values,omitempty"`
	Redacted bool     `json:"redacted,omitempty"`
	// Masked is the persisted reverse-ETL projection spelling for the same
	// declared secret boundary. Keep both markers so direct operation results
	// and full provider write receipts retain their established JSON contracts.
	Masked bool `json:"masked,omitempty"`
}

// GraphQLResponseMetadata is the bounded protocol metadata retained for a
// fixed-document GraphQL operation. Body remains the declared operation's
// data value; errors and rate-limit fields never alter the document or its
// selection set.
type GraphQLResponseMetadata struct {
	PartialData bool                 `json:"partial_data,omitempty"`
	Errors      []GraphQLResultError `json:"errors,omitempty"`
	RateLimit   *GraphQLRateLimit    `json:"rate_limit,omitempty"`
}

// GraphQLResultError reports one provider GraphQL error.
type GraphQLResultError struct {
	Message string `json:"message"`
}

// GraphQLRateLimit is the fixed, scalar rate-limit selection a declared
// GraphQL operation may return. It is observation metadata; requester-side
// admission and accounting remain the engine's declared rate-limit policy.
type GraphQLRateLimit struct {
	Limit     int    `json:"limit,omitempty"`
	Cost      int    `json:"cost,omitempty"`
	Remaining int    `json:"remaining,omitempty"`
	ResetAt   string `json:"reset_at,omitempty"`
}

// Reasons a direct read's page is not confirmed complete. They are declared
// beside DirectReadPage so the engine that sets them and the CLI that renders
// them cannot drift.
//
// The four not-paged reasons are deliberately distinct. Telling a caller that
// a connector "declares no pagination strategy" when it declares a working one
// — and most of them do — is the same class of untruth as an unsignalled
// truncation, just pointed the other way.
const (
	DirectReadPageReasonMorePages = "more_pages"
	// DirectReadPageReasonNoPagination: the bundle declares no pagination spec
	// at all, so nothing could prove where this page sits.
	DirectReadPageReasonNoPagination = "pagination_not_declared"
	// DirectReadPageReasonDeclaredNone: the bundle explicitly declares
	// pagination type "none". One request is the whole request it knows how to
	// make, but the declaration is not proof that the provider agrees.
	DirectReadPageReasonDeclaredNone = "pagination_declared_none"
	// DirectReadPageReasonNotAddressable: a real strategy is declared, but it
	// cannot page THIS request — a POST read carries its selection in a body
	// no query-param strategy can express.
	DirectReadPageReasonNotAddressable = "pagination_not_addressable_for_request"
	// DirectReadPageReasonInvalidSpec: a strategy is declared but its spec is
	// unusable, so paging degraded to a single page for this connector alone.
	DirectReadPageReasonInvalidSpec = "pagination_spec_invalid"
	DirectReadPageReasonAmbiguous   = "ambiguous_collection_shape"
	// DirectReadPageReasonSizeNotRequested: the declared strategy stops on a
	// SHORT page, but the size it compared against is not the size the request
	// carried, so the comparison proves nothing. Usually the spec names no
	// size/limit parameter at all and the provider applied its own default;
	// it also covers a caller-supplied size the walk could not adopt as its
	// threshold. Either way, asserting completeness would claim something
	// nothing measured.
	DirectReadPageReasonSizeNotRequested = "page_size_not_requested"
)

// DirectReadPage is the page context of one direct read.
//
// A direct read is page-wise EXPLORATION, not bulk extraction: it returns one
// page and tells the caller how to reach the next. (Bulk extraction is the ETL
// path, which stores what it reads; a direct read does not.) Complete is the
// load-bearing field — false means records exist that this result does not
// contain, and Reason says why the page stopped there.
type DirectReadPage struct {
	// Strategy is the pagination type the bundle declares, or "none" when it
	// declares one that cannot page (or declares nothing at all).
	Strategy string `json:"strategy"`
	// Records counts the rows on THIS page; 0 for a response that is not a
	// collection. When Reason is ambiguous_collection_shape it counts every
	// top-level array element instead, because which array pages is unknown
	// but how many rows arrived is not.
	Records int `json:"records"`
	// Size is the page size actually put on the wire. It is omitted when the
	// declared spec names no size parameter, since the provider then applied
	// its own default and no size was requested at all.
	Size int `json:"size,omitempty"`
	// Number is the addressable page number. Only page_number and offset_limit
	// have one; cursor, next_url and link_header address pages by opaque token
	// and leave this zero.
	Number int `json:"number,omitempty"`
	// HasMore reports that the provider has at least one further page.
	HasMore bool `json:"has_more"`
	// NextNumber is the page to request next, set only for addressable
	// strategies.
	NextNumber int `json:"next_number,omitempty"`
	// NextCursor is the opaque token to request next, set for the strategies
	// that have no addressable page number.
	NextCursor string `json:"next_cursor,omitempty"`
	// Complete is true only when this page is provably the whole collection.
	Complete bool `json:"complete"`
	// Reason names why Complete is false — one of the DirectReadPageReason
	// constants above.
	Reason string `json:"reason,omitempty"`
}

type DirectReader interface {
	DirectRead(context.Context, DirectReadRequest) (DirectReadResult, error)
}

type OperationDirectReader interface {
	OperationDirectRead(context.Context, OperationDirectReadRequest) (DirectReadResult, error)
}

// OperationDirectReadPreflighter exposes the no-network metadata admission
// check for one declared direct-read operation. Command preflight passes its
// exact operation-to-command binding through this interface so a command cannot
// claim availability "implemented" unless the runtime accepts its operation
// kind, endpoint, cap, and output policy.
type OperationDirectReadPreflighter interface {
	PreflightOperationDirectRead(operation, method, path string, maxBytes int, outputPolicy string) error
}

// OperationDirectReadBindingPreflighter proves the exact command-owned
// path/query/body mappings before commandrunner accepts an operation-backed
// direct read. Endpoint preflight alone cannot prevent an otherwise valid
// operation from receiving undeclared caller fields.
type OperationDirectReadBindingPreflighter interface {
	PreflightOperationDirectReadBindings(operation string, pathFields, queryFields, bodyFields []string, rawBody bool) error
}

// SourceBoundReadPreflighter verifies that a source-projected direct read
// still names the exact locked source operation carried by its selected engine
// operation. It has no URL, header, method, or request-body escape hatch.
type SourceBoundReadPreflighter interface {
	PreflightSourceBoundRead(operation, sourceOperation, method, path string) error
}

// SourceBoundStreamReadPreflighter performs the matching no-network proof for
// an existing ETL stream. It keeps a collection command on the stream executor
// only when its declaration-owned stream and source-bound operation agree.
type SourceBoundStreamReadPreflighter interface {
	PreflightSourceBoundStreamRead(stream, sourceOperation, method, path string) error
}

// SourceBoundOriginPreflighter checks the one declared source origin using
// public configuration only. Command dispatch invokes it before App credential
// resolution, so a caller cannot cause source-bound credential/auth state to
// materialize merely by selecting another provider origin.
type SourceBoundOriginPreflighter interface {
	PreflightSourceBoundOperationOrigin(operation string, cfg RuntimeConfig) error
	PreflightSourceBoundStreamOrigin(stream string, cfg RuntimeConfig) error
}

// OperationStructuredJSONVariablePreflighter exposes the deliberately narrow
// admission check for a structured CLI value in a fixed GraphQL operation.
// The operation declaration remains the only authority for which one
// top-level variable may receive an object/array; this interface must never be
// generalized into a request-body or raw-GraphQL parser.
type OperationStructuredJSONVariablePreflighter interface {
	PreflightOperationStructuredJSONVariable(operation, variable string) error
}

// OperationDirectWriteRequest is one declared, typed rest_write invocation.
//
// A caller must obtain PreviewDigest from PreviewOperationDirectWrite before
// execution. Destructive operations additionally require approval evidence
// issued for that exact preview; the engine consumes it at its shared write
// gate immediately before dispatch.
type OperationDirectWriteRequest struct {
	Operation  string
	Config     RuntimeConfig
	PathParams map[string]string
	Query      map[string]string
	// Headers contains only exact declaration-owned non-auth header values.
	// They are included in the preview digest that authorizes execution.
	Headers      map[string]string
	HeaderValues map[string][]string
	Body         map[string]any
	OutputPolicy string
	// RedactFields remains part of the request contract for compatibility, but
	// rest_write does not strip runtime content from it.
	RedactFields  []string
	Approval      *WriteApprovalEvidence
	PreviewDigest string
}

// OperationDirectWriteResult is the typed result of one declared REST or
// fixed-document GraphQL mutation. Provider-returned output is retained
// verbatim; output_policy shapes parsing only. System-generated diagnostics,
// plans, and logs remain separate secret-safe surfaces.
type OperationDirectWriteResult struct {
	Connector          string                             `json:"connector"`
	Operation          string                             `json:"operation"`
	Method             string                             `json:"method"`
	Path               string                             `json:"path"`
	ResponseReceived   bool                               `json:"response_received"`
	Status             int                                `json:"status"`
	Headers            map[string]OperationResponseHeader `json:"headers,omitempty"`
	BodyPresent        bool                               `json:"body_present"`
	BodyBytes          int                                `json:"body_bytes"`
	BodyRaw            string                             `json:"body_raw,omitempty"`
	BodyRawEncoding    string                             `json:"body_raw_encoding,omitempty"`
	Body               any                                `json:"body"`
	GraphQL            *GraphQLResponseMetadata           `json:"graphql,omitempty"`
	OutputSecretFields []string                           `json:"-"`
	// RequestSensitiveValues is populated only by the declaration-owned
	// executor and is deliberately not serializable or caller-controlled. It
	// protects system-generated diagnostics; provider output is classified only
	// by OutputSecretFields.
	RequestSensitiveValues []string `json:"-"`
}

// OperationDirectWriteMetadata is the no-network operation metadata needed by
// the connector-command plan lifecycle. It is intentionally a closed summary
// rather than a raw operation definition, so callers cannot turn it into a
// generic HTTP-write escape hatch.
type OperationDirectWriteMetadata struct {
	Operation             string
	MutationClass         string
	Risk                  string
	Approval              string
	ConfirmationChallenge string
	OutputPolicy          string
	Batchable             bool
	StructuredBody        bool
	// PayloadFileFields is nil for non-multipart operations. For a declared
	// multipart operation it is the closed set of body paths whose local-file
	// identities must be captured before preview, even when their names do not
	// follow a file_path convention.
	PayloadFileFields   []string
	PayloadFileMaxBytes map[string]int64
	// RedactFields is the operation's declared sensitive_policy.redact_fields.
	// It is the ONLY redaction source for an operation-backed reverse plan:
	// operation IDs and write-action names are separate namespaces that
	// collide by name in at least one bundle, so resolving one against the
	// other would withhold an unrelated set.
	RedactFields []string
}

// OperationDirectWriter is implemented by connectors that can preview and
// execute a declared REST or fixed-document GraphQL mutation through the
// shared write gate.
type OperationDirectWriter interface {
	PreviewOperationDirectWrite(context.Context, OperationDirectWriteRequest) (WritePreview, error)
	OperationDirectWrite(context.Context, OperationDirectWriteRequest) (OperationDirectWriteResult, error)
}

// OperationDirectWritePreflighter exposes the no-network admission check for
// a command's exact direct-write binding. A command cannot be executable merely
// because an operation has compatible metadata: its method, provider-relative
// path, output policy, and query-field bindings must match the fixed
// declaration that the runtime will dispatch.
type OperationDirectWritePreflighter interface {
	PreflightOperationDirectWrite(operation, method, path, outputPolicy string, queryFields ...string) error
}

type OperationDirectWriteBindingPreflighter interface {
	PreflightOperationDirectWriteBindings(operation string, pathFields, bodyFields []string) error
}

type OperationDirectWriteBodyMaterializer interface {
	MaterializeOperationDirectWriteBody(operation string, mappings map[string]any) (map[string]any, error)
}

type OperationDirectWriteBodyValueResolver interface {
	ResolveOperationDirectWriteBodyValue(operation string, body map[string]any, path string) (any, bool, error)
}

// OperationDirectWriteBodyPlanTransformer is the closed plan/reconstitution
// companion for a declaration-owned structured body. It has no method, URL,
// header, or raw-body authority: every accepted path is resolved against the
// operation's bound body schema.
type OperationDirectWriteBodyPlanTransformer interface {
	WithholdOperationDirectWriteBodyFields(operation string, body map[string]any, fields []string) (map[string]any, []string, error)
	RedactOperationDirectWriteBodyFields(operation string, body map[string]any, fields []string) (map[string]any, error)
	MergeOperationDirectWriteBodyFragments(operation string, base, overlay map[string]any) (map[string]any, error)
	OperationDirectWriteBodyPathContains(operation, parent, child string) (bool, error)
}

// OperationStructuredJSONBodyPreflighter proves that one named top-level
// operation body field is a closed, bounded object or array in the operation
// declaration. It deliberately accepts neither a raw body nor a dotted path:
// command runners use it before parsing a json flag so source declarations,
// not callers, own the resulting request structure.
type OperationStructuredJSONBodyPreflighter interface {
	PreflightOperationStructuredJSONBodyField(operation, field string) error
}

// OperationDirectWriteMetadataProvider exposes the plan-safe metadata for a
// declared REST or fixed-document GraphQL mutation without preparing
// credentials or making a network request.
type OperationDirectWriteMetadataProvider interface {
	OperationDirectWriteMetadata(operation string) (OperationDirectWriteMetadata, error)
}

// OperationBinaryDownloadRequest is one bounded binary/file download driven by
// a declared binary_download or text_export operation.
//
// DestRoot is required and is the directory the download is confined beneath;
// there is no implicit destination, because a CLI that guesses where to write
// a file is a CLI that eventually writes it somewhere it should not.
type OperationBinaryDownloadRequest struct {
	Operation  string
	Config     RuntimeConfig
	PathParams map[string]string
	Query      map[string]string
	// Headers contains only exact declaration-owned non-auth header values.
	Headers      map[string]string
	HeaderValues map[string][]string
	// MaxBytes may only lower the operation's declared cap, never raise it.
	MaxBytes int64
	DestRoot string
	// FileName optionally names the file within DestRoot. It must be a local,
	// single-segment name; traversal is refused.
	FileName     string
	RedactFields []string
}

// OperationBinaryDownloadResult describes what landed on disk. Bytes are never
// inlined: a 25 MiB attachment would become a 34 MiB JSON line.
type OperationBinaryDownloadResult struct {
	Connector string                             `json:"connector"`
	Operation string                             `json:"operation"`
	Method    string                             `json:"method"`
	Path      string                             `json:"path"`
	Record    Record                             `json:"record"`
	Status    int                                `json:"status"`
	Headers   map[string]OperationResponseHeader `json:"headers,omitempty"`
	Receipt   *ProviderResponseReceipt           `json:"receipt,omitempty"`
}

// OperationBinaryDownloader is implemented by connectors that can execute a
// declared binary_download or text_export operation.
type OperationBinaryDownloader interface {
	OperationBinaryDownload(context.Context, OperationBinaryDownloadRequest) (OperationBinaryDownloadResult, error)
}

type OperationBinaryDownloadPreflighter interface {
	PreflightOperationBinaryDownload(operation, method, path string) error
}

// OperationStatusCheckRequest selects one declared response-less HEAD
// operation. It has no body, output policy, or pagination channel, so it
// cannot become a JSON direct-read escape hatch.
type OperationStatusCheckRequest struct {
	Operation  string
	Config     RuntimeConfig
	PathParams map[string]string
	Query      map[string]string
	// Headers contains only exact declaration-owned non-auth header values.
	Headers      map[string]string
	HeaderValues map[string][]string
}

// OperationStatusCheckResult intentionally exposes only bounded response
// metadata. A final non-2xx response remains metadata rather than an HTTP
// error, and HEAD response bodies are never decoded or surfaced.
type OperationStatusCheckResult struct {
	Connector string                             `json:"connector"`
	Operation string                             `json:"operation"`
	Method    string                             `json:"method"`
	Path      string                             `json:"path"`
	Status    int                                `json:"status"`
	BodyBytes int                                `json:"body_bytes"`
	Headers   map[string]OperationResponseHeader `json:"headers,omitempty"`
	// Receipt is the complete bounded HEAD response. Headers remains the
	// declaration-admitted convenience projection; Receipt preserves all
	// provider metadata so a terminal error does not erase audit evidence.
	Receipt *ProviderResponseReceipt `json:"receipt,omitempty"`
}

type OperationStatusChecker interface {
	OperationStatusCheck(context.Context, OperationStatusCheckRequest) (OperationStatusCheckResult, error)
}

type OperationStatusCheckPreflighter interface {
	PreflightOperationStatusCheck(operation, method, path, outputPolicy string) error
}

var ErrReadLimitReached = errors.New("connector read limit reached")

func LimitEmitter(limit int, emit func(Record) error) func(Record) error {
	if limit <= 0 {
		return emit
	}
	count := 0
	return func(record Record) error {
		if count >= limit {
			return ErrReadLimitReached
		}
		if err := emit(record); err != nil {
			return err
		}
		count++
		if count >= limit {
			return ErrReadLimitReached
		}
		return nil
	}
}

func IgnoreReadLimit(err error) error {
	if errors.Is(err, ErrReadLimitReached) {
		return nil
	}
	return err
}

func RejectLegacyConnectorName(name string) error {
	if !IsLegacyConnectorName(name) {
		return nil
	}
	return fmt.Errorf("connector %q uses a legacy source-/destination- prefix; use bare connector name %q", name, legacyBareConnectorName(name))
}

func IsLegacyConnectorName(name string) bool {
	normalized := strings.TrimSpace(strings.ToLower(name))
	return strings.HasPrefix(normalized, "source-") || strings.HasPrefix(normalized, "destination-")
}

func legacyBareConnectorName(name string) string {
	normalized := strings.TrimSpace(strings.ToLower(name))
	normalized = strings.TrimPrefix(normalized, "source-")
	normalized = strings.TrimPrefix(normalized, "destination-")
	return normalized
}

type WriteRequest struct {
	Stream    string
	Table     string
	Action    string
	Overwrite bool
	// DeliveryOccurrence is an internal, durable-workset identity supplied by
	// checkpointed destinations. It is never a provider parameter or a
	// caller-selected idempotency key; the engine hashes it with the sealed
	// preview/action/index so retries stay stable without aliasing worksets.
	DeliveryOccurrence string
	Config             RuntimeConfig
	PrimaryKey         []string
	Approval           *WriteApprovalEvidence
}

// ConfirmationKind is the closed runtime vocabulary for an explicit write
// confirmation. It is deliberately not a caller-defined prompt.
type ConfirmationKind string

const ConfirmationKindDestructive ConfirmationKind = "destructive"

// WriteConfirmation is the typed, closed confirmation attached to an
// explicitly approved write request.
type WriteConfirmation struct {
	Kind ConfirmationKind `json:"kind"`
}

// ParseWriteConfirmation maps CLI and persisted values into the closed
// confirmation vocabulary.
func ParseWriteConfirmation(raw string) (WriteConfirmation, error) {
	switch ConfirmationKind(strings.TrimSpace(raw)) {
	case "":
		return WriteConfirmation{}, nil
	case ConfirmationKindDestructive:
		return WriteConfirmation{Kind: ConfirmationKindDestructive}, nil
	default:
		return WriteConfirmation{}, fmt.Errorf("unsupported confirmation kind %q", strings.TrimSpace(raw))
	}
}

type WriteResult struct {
	RecordsWritten    int                     `json:"records_written"`
	RecordsFailed     int                     `json:"records_failed"`
	RecordsUnchanged  int                     `json:"records_unchanged,omitempty"`
	ProviderResponses []WriteProviderResponse `json:"provider_responses,omitempty"`
}

// WriteProviderResponse is the verbatim provider result captured for one named
// typed write action. System-generated diagnostics are carried separately.
type WriteProviderResponse struct {
	RecordIndex     int                            `json:"record_index"`
	Status          int                            `json:"status"`
	Headers         map[string]WriteProviderHeader `json:"headers"`
	BodyPresent     bool                           `json:"body_present"`
	BodyBytes       int                            `json:"body_bytes"`
	BodyRaw         string                         `json:"body_raw,omitempty"`
	BodyRawEncoding string                         `json:"body_raw_encoding,omitempty"`
	Body            any                            `json:"body"`
	BodyEncoding    string                         `json:"body_encoding,omitempty"`
}

// WriteProviderHeader and OperationResponseHeader share the same complete
// provider-header representation. The alias prevents the declarative reverse
// result path from narrowing a declaration-owned operation result.
type WriteProviderHeader = OperationResponseHeader

// SanitizeWriteResultForOutput returns a public clone of the complete internal
// receipt, preserving provider-returned data verbatim.
func SanitizeWriteResultForOutput(result WriteResult, _ map[string]string) WriteResult {
	out := result
	out.ProviderResponses = make([]WriteProviderResponse, len(result.ProviderResponses))
	for i, response := range result.ProviderResponses {
		out.ProviderResponses[i] = cloneWriteProviderResponse(response)
	}
	return out
}

// SanitizeOperationDirectWriteResultForOutput applies the same public-boundary
// clone and also masks declaration-classified response secret locations.
func SanitizeOperationDirectWriteResultForOutput(result OperationDirectWriteResult, _ map[string]string) OperationDirectWriteResult {
	declaredValues := providerOutputSecretValues(result.Body, result.OutputSecretFields)
	body := cloneProviderOutputValue(result.Body)
	for _, path := range result.OutputSecretFields {
		maskProviderOutputPath(body, path)
	}
	out := result
	out.Headers = cloneProviderHeaders(result.Headers)
	out.BodyRaw = sanitizeProviderResponseRawAtPaths(result.BodyRaw, result.BodyRawEncoding, nil, result.OutputSecretFields)
	out.Body = cloneProviderOutputValue(body)
	out.OutputSecretFields = append([]string(nil), result.OutputSecretFields...)
	out.RequestSensitiveValues = append([]string(nil), result.RequestSensitiveValues...)
	if result.GraphQL != nil {
		graphql := *result.GraphQL
		graphql.Errors = append([]GraphQLResultError(nil), result.GraphQL.Errors...)
		graphqlSecrets := appendDeclaredSecretVariants(nil, declaredValues)
		for i := range graphql.Errors {
			graphql.Errors[i].Message = redactWriteResultString(graphql.Errors[i].Message, graphqlSecrets)
		}
		if result.GraphQL.RateLimit != nil {
			rateLimit := *result.GraphQL.RateLimit
			graphql.RateLimit = &rateLimit
		}
		out.GraphQL = &graphql
	}
	return out
}

// SanitizeOperationDirectWriteResultForPlaintextState returns the narrow
// persistence-only projection for a failed direct write. Provider output is
// intentionally verbatim at the caller-facing boundary; this helper exists
// because ReverseRun is also written to the unencrypted state.json file.
//
// Only concrete values supplied through the credential secret store are
// withheld here. It is not a general response-redaction layer and must not be
// used for CLI output or provider-result return values.
func SanitizeOperationDirectWriteResultForPlaintextState(result OperationDirectWriteResult, secrets map[string]string) OperationDirectWriteResult {
	out := SanitizeOperationDirectWriteResultForOutput(result, nil)
	masked := configuredWriteResultSecrets(secrets)
	if len(masked) == 0 {
		return out
	}
	out.Headers = redactOperationResponseHeadersForPlaintextState(out.Headers, masked)
	out.BodyRaw = redactProviderDiagnosticText(out.BodyRaw, masked)
	out.Body = redactProviderOutputValueForPlaintextState(out.Body, masked)
	if rawBody, ok := out.Body.(string); ok && result.BodyRaw != "" && rawBody == result.BodyRaw {
		out.Body = out.BodyRaw
	}
	if out.GraphQL != nil {
		for index := range out.GraphQL.Errors {
			out.GraphQL.Errors[index].Message = redactProviderDiagnosticText(out.GraphQL.Errors[index].Message, masked)
		}
	}
	return out
}

func redactOperationResponseHeadersForPlaintextState(headers map[string]OperationResponseHeader, secrets []string) map[string]OperationResponseHeader {
	out := cloneProviderHeaders(headers)
	for name, header := range out {
		for index := range header.Values {
			header.Values[index] = redactProviderDiagnosticText(header.Values[index], secrets)
		}
		out[name] = header
	}
	return out
}

func redactProviderOutputValueForPlaintextState(value any, secrets []string) any {
	switch typed := value.(type) {
	case string:
		return redactProviderDiagnosticText(typed, secrets)
	case json.Number:
		redacted := redactProviderDiagnosticText(typed.String(), secrets)
		if redacted == typed.String() {
			return typed
		}
		return redacted
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = redactProviderOutputValueForPlaintextState(item, secrets)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for index, item := range typed {
			out[index] = redactProviderOutputValueForPlaintextState(item, secrets)
		}
		return out
	case []string:
		out := make([]string, len(typed))
		for index, item := range typed {
			out[index] = redactProviderDiagnosticText(item, secrets)
		}
		return out
	case []byte:
		return []byte(redactProviderDiagnosticText(string(typed), secrets))
	default:
		return value
	}
}

func cloneWriteProviderResponse(response WriteProviderResponse) WriteProviderResponse {
	out := response
	out.Headers = cloneProviderHeaders(response.Headers)
	out.Body = cloneProviderOutputValue(response.Body)
	return out
}

func cloneProviderHeaders(headers map[string]OperationResponseHeader) map[string]OperationResponseHeader {
	if headers == nil {
		return nil
	}
	out := make(map[string]OperationResponseHeader, len(headers))
	for name, header := range headers {
		clone := header
		clone.Values = append([]string(nil), header.Values...)
		out[name] = clone
	}
	return out
}

// SanitizeGraphQLResponseMetadataForOutput clones provider-returned GraphQL
// metadata. A response diagnostic may legitimately equal credential bytes, so
// only declaration-owned output-secret fields are eligible for masking at a
// public boundary.
func SanitizeGraphQLResponseMetadataForOutput(metadata *GraphQLResponseMetadata, _ map[string]string) *GraphQLResponseMetadata {
	return cloneGraphQLResponseMetadata(metadata)
}

func cloneGraphQLResponseMetadata(metadata *GraphQLResponseMetadata) *GraphQLResponseMetadata {
	if metadata == nil {
		return nil
	}
	out := *metadata
	out.Errors = append([]GraphQLResultError(nil), metadata.Errors...)
	if metadata.RateLimit != nil {
		rateLimit := *metadata.RateLimit
		out.RateLimit = &rateLimit
	}
	return &out
}

func sanitizeWriteProviderResponse(response WriteProviderResponse, secrets []string) WriteProviderResponse {
	_ = secrets
	out := response
	out.Headers = cloneProviderHeaders(response.Headers)
	out.BodyRaw = response.BodyRaw
	out.Body = cloneProviderOutputValue(response.Body)
	return out
}

func sanitizeProviderHeaders(headers map[string]OperationResponseHeader, secrets []string) map[string]OperationResponseHeader {
	_ = secrets
	return cloneProviderHeaders(headers)
}

func cloneProviderOutputValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = cloneProviderOutputValue(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = cloneProviderOutputValue(item)
		}
		return out
	case []string:
		return append([]string(nil), typed...)
	case []byte:
		return append([]byte(nil), typed...)
	default:
		return value
	}
}

func sanitizeProviderOutputValue(value any, secrets []string) any {
	_ = secrets
	return cloneProviderOutputValue(value)
}

func sanitizeProviderResponseValue(value any, secrets []string) any {
	_ = secrets
	return cloneProviderOutputValue(value)
}

func sanitizeProviderOutputScalar(value string, secrets []string) string {
	_ = secrets
	return value
}

// SanitizeProviderOutputForOutput deep-clones an arbitrary decoded provider
// value while preserving provider-returned fields, keys, and values.
func SanitizeProviderOutputForOutput(value any, _ map[string]string) any {
	return cloneProviderOutputValue(value)
}

func SanitizeProviderResponseHeadersForOutput(headers map[string]OperationResponseHeader, _ map[string]string) map[string]OperationResponseHeader {
	return cloneProviderHeaders(headers)
}

// SanitizeDirectReadPageForOutput clones page navigation while retaining the
// provider-issued cursor verbatim.
func SanitizeDirectReadPageForOutput(page DirectReadPage, _ map[string]string) DirectReadPage {
	return page
}

// SanitizeProviderResponseReceiptForOutput clones a complete response receipt
// while retaining provider-returned fields, keys, and values verbatim.
func SanitizeProviderResponseReceiptForOutput(receipt ProviderResponseReceipt, _ map[string]string) ProviderResponseReceipt {
	return SanitizeProviderResponseReceiptWithDeclaredSecretsForOutput(receipt, nil, nil)
}

// SanitizeProviderResponseReceiptWithDeclaredSecretsForOutput is the sole
// public projection of an immutable provider receipt. Declared field paths
// identify values (never keys) to mask.
func SanitizeProviderResponseReceiptWithDeclaredSecretsForOutput(receipt ProviderResponseReceipt, _ map[string]string, declaredFields []string) ProviderResponseReceipt {
	out := receipt
	out.Headers = cloneProviderHeaders(receipt.Headers)
	body := cloneProviderOutputValue(receipt.Body)
	maskProviderOutputPaths(body, declaredFields)
	out.BodyRaw = sanitizeProviderResponseRawAtPaths(receipt.BodyRaw, receipt.BodyRawEncoding, nil, declaredFields)
	out.Body = cloneProviderOutputValue(body)
	if rawBody, ok := receipt.Body.(string); ok && receipt.BodyRawEncoding == "base64" && rawBody == receipt.BodyRaw {
		out.Body = out.BodyRaw
	}
	return out
}

// SanitizeDirectReadResultForOutput projects an engine-owned result at the
// command boundary. The engine result remains immutable provider evidence;
// this clone is the only representation suitable for stdout or persistence.
func SanitizeDirectReadResultForOutput(result DirectReadResult, _ map[string]string) DirectReadResult {
	var declaredValues []string
	if result.Receipt != nil {
		declaredValues = providerOutputSecretValues(result.Receipt.Body, result.OutputSecretFields)
	}
	out := result
	out.Headers = cloneProviderHeaders(result.Headers)
	out.Body = cloneProviderOutputValue(result.Body)
	out.OutputSecretFields = append([]string(nil), result.OutputSecretFields...)
	if result.Receipt != nil {
		receipt := SanitizeProviderResponseReceiptWithDeclaredSecretsForOutput(*result.Receipt, nil, result.OutputSecretFields)
		out.Receipt = &receipt
	}
	if result.GraphQL != nil {
		graphql := *result.GraphQL
		graphql.Errors = append([]GraphQLResultError(nil), result.GraphQL.Errors...)
		graphqlSecrets := appendDeclaredSecretVariants(nil, declaredValues)
		for i := range graphql.Errors {
			graphql.Errors[i].Message = redactWriteResultString(graphql.Errors[i].Message, graphqlSecrets)
		}
		if result.GraphQL.RateLimit != nil {
			rateLimit := *result.GraphQL.RateLimit
			graphql.RateLimit = &rateLimit
		}
		out.GraphQL = &graphql
	}
	return out
}

// SanitizeOperationStatusCheckResultForOutput mirrors direct-read receipt
// safety without adding a response-body surface to a HEAD operation.
func SanitizeOperationStatusCheckResultForOutput(result OperationStatusCheckResult, _ map[string]string) OperationStatusCheckResult {
	out := result
	out.Headers = cloneProviderHeaders(result.Headers)
	if result.Receipt != nil {
		receipt := SanitizeProviderResponseReceiptForOutput(*result.Receipt, nil)
		out.Receipt = &receipt
	}
	return out
}

// SanitizeOperationBinaryDownloadResultForOutput keeps binary transfer bytes
// out of the record while projecting its bounded provider receipt safely.
func SanitizeOperationBinaryDownloadResultForOutput(result OperationBinaryDownloadResult, _ map[string]string) OperationBinaryDownloadResult {
	out := result
	out.Headers = cloneProviderHeaders(result.Headers)
	if result.Receipt != nil {
		receipt := SanitizeProviderResponseReceiptForOutput(*result.Receipt, nil)
		out.Receipt = &receipt
	}
	return out
}

func sanitizeProviderResponseRawAtPaths(value, encoding string, secrets, declaredFields []string) string {
	if len(declaredFields) == 0 {
		return sanitizeProviderResponseRaw(value, encoding, secrets)
	}
	raw := []byte(value)
	base64Encoded := encoding == "base64"
	if base64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(value)
		if err != nil {
			return sanitizeProviderResponseRaw(value, encoding, secrets)
		}
		raw = decoded
	}
	if publicJSON, ok := sanitizeProviderJSON(raw, secrets, declaredFields); ok {
		raw = publicJSON
	} else {
		return sanitizeProviderResponseRaw(value, encoding, secrets)
	}
	if base64Encoded {
		return base64.StdEncoding.EncodeToString(raw)
	}
	return string(raw)
}

func sanitizeProviderJSON(raw []byte, secrets, declaredFields []string) ([]byte, bool) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return nil, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, false
	}
	original := cloneProviderOutputValue(decoded)
	maskProviderOutputPaths(decoded, declaredFields)
	public := sanitizeProviderOutputValue(decoded, secrets)
	// Provider bytes are evidence, not a formatting suggestion. A parse is
	// necessary to find a concrete scalar to withhold, but when no declared or
	// configured value matched it must not change whitespace, key order,
	// numeric spelling, escapes, or non-ASCII byte representation.
	if reflect.DeepEqual(original, public) {
		return append([]byte(nil), raw...), true
	}
	encoded, err := json.Marshal(public)
	if err != nil {
		return nil, false
	}
	return encoded, true
}

func sanitizeProviderResponseBody(raw, rawEncoding string, body any, secrets []string) (string, any) {
	_ = rawEncoding
	_ = secrets
	sanitizedRaw := raw
	sanitizedBody := cloneProviderOutputValue(body)
	if rawBody, ok := body.(string); ok && rawBody == raw {
		sanitizedBody = sanitizedRaw
	}
	return sanitizedRaw, sanitizedBody
}

func sanitizeProviderResponseRaw(value, encoding string, secrets []string) string {
	_ = encoding
	_ = secrets
	return value
}

type providerResponseJSONStringSpan struct {
	start int
	end   int
	key   bool
}

type providerResponseJSONContainerState uint8

const (
	providerResponseJSONObjectKeyOrEnd providerResponseJSONContainerState = iota
	providerResponseJSONObjectColon
	providerResponseJSONObjectValue
	providerResponseJSONObjectCommaOrEnd
	providerResponseJSONArrayValueOrEnd
	providerResponseJSONArrayCommaOrEnd
	providerResponseJSONContainerInvalid
)

type providerResponseJSONContainer struct {
	kind  byte
	state providerResponseJSONContainerState
}

type providerResponseJSONScanner struct {
	containers []providerResponseJSONContainer
}

func redactProviderResponseText(value string, secrets []string) string {
	if len(secrets) == 0 {
		return value
	}
	stringSpans, isJSON := providerResponseJSONStrings(value)
	if !isJSON || len(stringSpans) == 0 {
		return redactProviderResponseTextSegment(value, secrets)
	}
	var out strings.Builder
	out.Grow(len(value))
	offset := 0
	for _, span := range stringSpans {
		out.WriteString(redactProviderResponseTextSegment(value[offset:span.start], secrets))
		if span.key {
			out.WriteString(value[span.start:span.end])
		} else {
			out.WriteString(redactProviderResponseJSONString(value[span.start:span.end], secrets))
		}
		offset = span.end
	}
	out.WriteString(redactProviderResponseTextSegment(value[offset:], secrets))
	return out.String()
}

func redactProviderResponseTextSegment(value string, secrets []string) string {
	for _, secret := range secrets {
		value = redactConcreteSecretTextWithBoundary(value, secret, providerResponseTokenBoundary)
	}
	return value
}

func providerResponseJSONStrings(value string) ([]providerResponseJSONStringSpan, bool) {
	spans := make([]providerResponseJSONStringSpan, 0)
	scanner := providerResponseJSONScanner{}
	for offset := 0; offset < len(value); {
		if isProviderResponseJSONWhitespace(value[offset]) {
			offset++
			continue
		}
		if value[offset] != '"' {
			scanner.consume(value, &offset)
			continue
		}
		start := offset
		offset++
		terminated := false
		for offset < len(value) {
			switch value[offset] {
			case '\\':
				offset += 2
			case '"':
				offset++
				terminated = true
			default:
				offset++
			}
			if terminated {
				break
			}
		}
		if !terminated {
			break
		}
		spans = append(spans, providerResponseJSONStringSpan{start: start, end: offset, key: scanner.consumeString(value, offset)})
	}
	return spans, len(spans) != 0
}

func (scanner *providerResponseJSONScanner) consume(value string, offset *int) {
	switch value[*offset] {
	case '{', '[':
		scanner.open(value[*offset])
		*offset = *offset + 1
	case '}', ']':
		scanner.close(value[*offset])
		*offset = *offset + 1
	case ':':
		scanner.colon()
		*offset = *offset + 1
	case ',':
		scanner.comma()
		*offset = *offset + 1
	default:
		start := *offset
		for *offset < len(value) && !isProviderResponseJSONWhitespace(value[*offset]) && !isProviderResponseJSONStructuralByte(value[*offset]) {
			*offset = *offset + 1
		}
		scanner.literal(value[start:*offset])
	}
}

func (scanner *providerResponseJSONScanner) consumeString(value string, end int) bool {
	container := scanner.top()
	if container == nil {
		return false
	}
	next := end
	for next < len(value) && isProviderResponseJSONWhitespace(value[next]) {
		next++
	}
	if container.kind == '{' && container.state == providerResponseJSONObjectKeyOrEnd && next < len(value) && value[next] == ':' {
		container.state = providerResponseJSONObjectColon
		return true
	}
	switch container.state {
	case providerResponseJSONObjectValue:
		container.state = providerResponseJSONObjectCommaOrEnd
	case providerResponseJSONArrayValueOrEnd:
		container.state = providerResponseJSONArrayCommaOrEnd
	default:
		container.state = providerResponseJSONContainerInvalid
	}
	return false
}

func (scanner *providerResponseJSONScanner) open(kind byte) {
	if container := scanner.top(); container != nil {
		switch container.state {
		case providerResponseJSONObjectValue:
			container.state = providerResponseJSONObjectCommaOrEnd
		case providerResponseJSONArrayValueOrEnd:
			container.state = providerResponseJSONArrayCommaOrEnd
		default:
			container.state = providerResponseJSONContainerInvalid
			return
		}
	}
	state := providerResponseJSONObjectKeyOrEnd
	if kind == '[' {
		state = providerResponseJSONArrayValueOrEnd
	}
	scanner.containers = append(scanner.containers, providerResponseJSONContainer{kind: kind, state: state})
}

func (scanner *providerResponseJSONScanner) close(kind byte) {
	container := scanner.top()
	if container == nil {
		return
	}
	if (kind != '}' || container.kind != '{') && (kind != ']' || container.kind != '[') {
		container.state = providerResponseJSONContainerInvalid
		return
	}
	complete := providerResponseJSONContainerComplete(*container)
	scanner.containers = scanner.containers[:len(scanner.containers)-1]
	if !complete {
		if parent := scanner.top(); parent != nil {
			parent.state = providerResponseJSONContainerInvalid
		}
	}
}

func (scanner *providerResponseJSONScanner) colon() {
	container := scanner.top()
	if container == nil {
		return
	}
	if container.kind == '{' && container.state == providerResponseJSONObjectColon {
		container.state = providerResponseJSONObjectValue
		return
	}
	container.state = providerResponseJSONContainerInvalid
}

func (scanner *providerResponseJSONScanner) comma() {
	container := scanner.top()
	if container == nil {
		return
	}
	switch container.kind {
	case '{':
		if container.state == providerResponseJSONObjectCommaOrEnd {
			container.state = providerResponseJSONObjectKeyOrEnd
			return
		}
		container.state = providerResponseJSONContainerInvalid
	case '[':
		if container.state == providerResponseJSONArrayCommaOrEnd {
			container.state = providerResponseJSONArrayValueOrEnd
			return
		}
		container.state = providerResponseJSONContainerInvalid
	}
}

func (scanner *providerResponseJSONScanner) literal(value string) {
	container := scanner.top()
	if container == nil {
		return
	}
	if !json.Valid([]byte(value)) {
		container.state = providerResponseJSONContainerInvalid
		return
	}
	switch container.state {
	case providerResponseJSONObjectValue:
		container.state = providerResponseJSONObjectCommaOrEnd
	case providerResponseJSONArrayValueOrEnd:
		container.state = providerResponseJSONArrayCommaOrEnd
	default:
		container.state = providerResponseJSONContainerInvalid
	}
}

func (scanner *providerResponseJSONScanner) top() *providerResponseJSONContainer {
	if len(scanner.containers) == 0 {
		return nil
	}
	return &scanner.containers[len(scanner.containers)-1]
}

func providerResponseJSONContainerComplete(container providerResponseJSONContainer) bool {
	switch container.kind {
	case '{':
		return container.state == providerResponseJSONObjectKeyOrEnd || container.state == providerResponseJSONObjectCommaOrEnd
	case '[':
		return container.state == providerResponseJSONArrayValueOrEnd || container.state == providerResponseJSONArrayCommaOrEnd
	default:
		return false
	}
}

func isProviderResponseJSONStructuralByte(value byte) bool {
	switch value {
	case '{', '}', '[', ']', ':', ',', '"':
		return true
	default:
		return false
	}
}

func isProviderResponseJSONWhitespace(value byte) bool {
	return value == ' ' || value == '\t' || value == '\n' || value == '\r'
}

type providerResponseDecodedByteSpan struct {
	start int
	end   int
}

type providerResponseTextSpan struct {
	start int
	end   int
}

func redactProviderResponseJSONString(value string, secrets []string) string {
	decoded, decodedSpans, ok := decodeProviderResponseJSONString(value)
	if !ok {
		return redactProviderResponseTextSegment(value, secrets)
	}
	matches := providerResponseSecretTextSpans(decoded, secrets)
	if len(matches) == 0 {
		return value
	}
	var out strings.Builder
	out.Grow(len(value))
	offset := 0
	for _, match := range matches {
		start := decodedSpans[match.start].start
		end := decodedSpans[match.end-1].end
		out.WriteString(value[offset:start])
		out.WriteString("[masked]")
		offset = end
	}
	out.WriteString(value[offset:])
	return out.String()
}

func decodeProviderResponseJSONString(value string) (string, []providerResponseDecodedByteSpan, bool) {
	if len(value) < 2 || value[0] != '"' || value[len(value)-1] != '"' {
		return "", nil, false
	}
	decoded := make([]byte, 0, len(value)-2)
	spans := make([]providerResponseDecodedByteSpan, 0, len(value)-2)
	appendBytes := func(start, end int, bytes []byte) {
		decoded = append(decoded, bytes...)
		for range bytes {
			spans = append(spans, providerResponseDecodedByteSpan{start: start, end: end})
		}
	}
	for offset := 1; offset < len(value)-1; {
		if value[offset] != '\\' {
			appendBytes(offset, offset+1, []byte{value[offset]})
			offset++
			continue
		}
		if offset+1 >= len(value)-1 {
			return "", nil, false
		}
		start := offset
		switch value[offset+1] {
		case '"', '\\', '/':
			appendBytes(start, offset+2, []byte{value[offset+1]})
			offset += 2
		case 'b':
			appendBytes(start, offset+2, []byte{'\b'})
			offset += 2
		case 'f':
			appendBytes(start, offset+2, []byte{'\f'})
			offset += 2
		case 'n':
			appendBytes(start, offset+2, []byte{'\n'})
			offset += 2
		case 'r':
			appendBytes(start, offset+2, []byte{'\r'})
			offset += 2
		case 't':
			appendBytes(start, offset+2, []byte{'\t'})
			offset += 2
		case 'u':
			if offset+6 > len(value)-1 {
				return "", nil, false
			}
			r, ok := providerResponseJSONHexRune(value[offset+2 : offset+6])
			if !ok {
				return "", nil, false
			}
			offset += 6
			end := offset
			if r >= 0xD800 && r <= 0xDBFF && offset+6 <= len(value)-1 && value[offset] == '\\' && value[offset+1] == 'u' {
				low, validLow := providerResponseJSONHexRune(value[offset+2 : offset+6])
				if validLow && low >= 0xDC00 && low <= 0xDFFF {
					r = utf16.DecodeRune(r, low)
					offset += 6
					end = offset
				} else {
					r = utf8.RuneError
				}
			} else if r >= 0xD800 && r <= 0xDFFF {
				r = utf8.RuneError
			}
			appendBytes(start, end, []byte(string(r)))
		default:
			return "", nil, false
		}
	}
	return string(decoded), spans, true
}

func providerResponseJSONHexRune(value string) (rune, bool) {
	if len(value) != 4 {
		return 0, false
	}
	var result rune
	for _, digit := range value {
		result <<= 4
		switch {
		case digit >= '0' && digit <= '9':
			result += digit - '0'
		case digit >= 'a' && digit <= 'f':
			result += digit - 'a' + 10
		case digit >= 'A' && digit <= 'F':
			result += digit - 'A' + 10
		default:
			return 0, false
		}
	}
	return result, true
}

func providerResponseSecretTextSpans(value string, secrets []string) []providerResponseTextSpan {
	spans := make([]providerResponseTextSpan, 0)
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		for offset := 0; offset < len(value); {
			index := strings.Index(value[offset:], secret)
			if index < 0 {
				break
			}
			start := offset + index
			end := start + len(secret)
			if providerOutputTokenBoundary(value, start, end) && !providerResponseTextSpanOverlaps(spans, start, end) {
				spans = append(spans, providerResponseTextSpan{start: start, end: end})
			}
			offset = end
		}
	}
	sort.Slice(spans, func(i, j int) bool {
		if spans[i].start == spans[j].start {
			return spans[i].end < spans[j].end
		}
		return spans[i].start < spans[j].start
	})
	return spans
}

func providerResponseTextSpanOverlaps(spans []providerResponseTextSpan, start, end int) bool {
	for _, span := range spans {
		if start < span.end && span.start < end {
			return true
		}
	}
	return false
}

func providerOutputValueAtPath(value any, path string) (any, bool) {
	current := value
	for _, segment := range strings.Split(strings.TrimPrefix(path, "body."), ".") {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[segment]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func maskProviderOutputPath(value any, path string) {
	segments := strings.Split(strings.TrimPrefix(path, "body."), ".")
	maskProviderOutputSegments(value, segments)
}

func maskProviderOutputPaths(value any, paths []string) {
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}
		maskProviderOutputPath(value, path)
	}
}

func maskProviderOutputSegments(value any, segments []string) {
	if len(segments) == 0 {
		return
	}
	switch current := value.(type) {
	case map[string]any:
		next, present := current[segments[0]]
		if !present {
			return
		}
		if len(segments) == 1 {
			current[segments[0]] = "[masked]"
			return
		}
		maskProviderOutputSegments(next, segments[1:])
	case []any:
		for _, item := range current {
			maskProviderOutputSegments(item, segments)
		}
	}
}

func providerOutputSecretValues(value any, paths []string) []string {
	values := make([]string, 0, len(paths))
	for _, path := range paths {
		segments := strings.Split(strings.TrimPrefix(strings.TrimSpace(path), "body."), ".")
		if len(segments) == 0 || segments[0] == "" {
			continue
		}
		collectProviderOutputSecretValues(value, segments, &values)
	}
	return values
}

func collectProviderOutputSecretValues(value any, segments []string, values *[]string) {
	if len(segments) == 0 {
		if scalar, ok := providerOutputSecretScalar(value); ok {
			*values = append(*values, scalar)
		}
		return
	}
	switch current := value.(type) {
	case map[string]any:
		next, present := current[segments[0]]
		if present {
			collectProviderOutputSecretValues(next, segments[1:], values)
		}
	case []any:
		for _, item := range current {
			collectProviderOutputSecretValues(item, segments, values)
		}
	}
}

func appendDeclaredSecretVariants(secrets, declared []string) []string {
	out := append([]string(nil), secrets...)
	seen := make(map[string]struct{}, len(out))
	for _, value := range out {
		seen[value] = struct{}{}
	}
	for _, value := range declared {
		for _, variant := range configuredSecretVariants(value) {
			if _, present := seen[variant]; present {
				continue
			}
			seen[variant] = struct{}{}
			out = append(out, variant)
		}
	}
	return out
}

func providerOutputSecretScalar(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		return typed, typed != ""
	case json.Number:
		return typed.String(), typed.String() != ""
	default:
		return "", false
	}
}

func SanitizeWriteErrorForOutput(err error, secrets map[string]string) string {
	if err == nil {
		return ""
	}
	return redactProviderDiagnosticText(err.Error(), configuredWriteResultSecrets(secrets))
}

// Diagnostics are not provider scalar values: an upstream error may quote a
// credential inside surrounding prose. Redact concrete known material here,
// while provider receipts continue to use exact scalar matching above.
func redactProviderDiagnosticText(value string, secrets []string) string {
	for _, secret := range secrets {
		if secret != "" {
			value = strings.ReplaceAll(value, secret, "[masked]")
		}
	}
	return value
}

func configuredWriteResultSecrets(secrets map[string]string) []string {
	values := make([]string, 0, len(secrets)*7)
	seen := make(map[string]struct{}, len(secrets)*7)
	for _, value := range secrets {
		for _, variant := range configuredSecretVariants(value) {
			if _, found := seen[variant]; found {
				continue
			}
			seen[variant] = struct{}{}
			values = append(values, variant)
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if len(values[i]) == len(values[j]) {
			return values[i] < values[j]
		}
		return len(values[i]) > len(values[j])
	})
	return values
}

func configuredSecretVariants(value string) []string {
	if value == "" {
		return nil
	}
	bytes := []byte(value)
	variants := []string{
		value,
		url.QueryEscape(value),
		url.PathEscape(value),
		base64.StdEncoding.EncodeToString(bytes),
		base64.RawStdEncoding.EncodeToString(bytes),
		base64.URLEncoding.EncodeToString(bytes),
		base64.RawURLEncoding.EncodeToString(bytes),
	}
	if escaped, err := json.Marshal(value); err == nil && len(escaped) >= 2 {
		variants = append(variants, string(escaped[1:len(escaped)-1]))
	}
	return variants
}

func redactWriteResultString(value string, secrets []string) string {
	for _, secret := range secrets {
		value = redactConcreteSecretText(value, secret)
	}
	return value
}

func redactConcreteSecretText(value, secret string) string {
	return redactConcreteSecretTextWithBoundary(value, secret, providerOutputTokenBoundary)
}

func redactConcreteSecretTextWithBoundary(value, secret string, boundary func(string, int, int) bool) string {
	if secret == "" {
		return value
	}
	var out strings.Builder
	for offset := 0; ; {
		index := strings.Index(value[offset:], secret)
		if index < 0 {
			out.WriteString(value[offset:])
			return out.String()
		}
		start := offset + index
		end := start + len(secret)
		out.WriteString(value[offset:start])
		if boundary(value, start, end) {
			out.WriteString("[masked]")
		} else {
			out.WriteString(secret)
		}
		offset = end
	}
}

func providerOutputTokenBoundary(value string, start, end int) bool {
	return (start == 0 || !isProviderOutputTokenByte(value[start-1])) &&
		(end == len(value) || !isProviderOutputTokenByte(value[end]))
}

func providerResponseTokenBoundary(value string, start, end int) bool {
	return providerResponseTokenBoundaryBefore(value, start) && providerResponseTokenBoundaryAfter(value, end)
}

func providerResponseTokenBoundaryBefore(value string, start int) bool {
	if start == 0 {
		return true
	}
	if decoded, ok := providerResponseEscapedRuneBefore(value, start); ok {
		return !isProviderResponseTokenRune(decoded)
	}
	return !isProviderOutputTokenByte(value[start-1])
}

func providerResponseTokenBoundaryAfter(value string, end int) bool {
	if end == len(value) {
		return true
	}
	if decoded, ok := providerResponseEscapedRuneAfter(value, end); ok {
		return !isProviderResponseTokenRune(decoded)
	}
	return !isProviderOutputTokenByte(value[end])
}

func providerResponseEscapedRuneBefore(value string, end int) (rune, bool) {
	start := end - 6
	if start < 0 || value[start] != '\\' || value[start+1] != 'u' || providerResponseEscapedBackslash(value, start) {
		return 0, false
	}
	return providerResponseJSONHexRune(value[start+2 : end])
}

func providerResponseEscapedRuneAfter(value string, start int) (rune, bool) {
	end := start + 6
	if end > len(value) || value[start] != '\\' || value[start+1] != 'u' || providerResponseEscapedBackslash(value, start) {
		return 0, false
	}
	return providerResponseJSONHexRune(value[start+2 : end])
}

func providerResponseEscapedBackslash(value string, slash int) bool {
	count := 0
	for offset := slash; offset >= 0 && value[offset] == '\\'; offset-- {
		count++
	}
	return count%2 == 0
}

func isProviderResponseTokenRune(value rune) bool {
	return value < utf8.RuneSelf && isProviderOutputTokenByte(byte(value))
}

func isProviderOutputTokenByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' ||
		value == '_' || value == '-' || value == '.' || value == '~' || value == '%'
}

type QueryRequest struct {
	SQL    string
	Stream string
	Limit  int
	Config RuntimeConfig
}

type QueryResult struct {
	Rows []Record `json:"rows"`
}

type WritePreview struct {
	RecordsStaged  int                 `json:"records_staged"`
	Action         string              `json:"action"`
	Warnings       []string            `json:"warnings,omitempty"`
	Digest         string              `json:"digest,omitempty"`
	ApprovalTarget WriteApprovalTarget `json:"approval_target,omitempty"`
}

type CDCReadRequest struct {
	Stream     string
	Config     RuntimeConfig
	State      map[string]string
	Checkpoint *synccontract.CheckpointEnvelope
	// TransactionReceiver is the production whole-transaction boundary. A
	// changefeed may call the legacy per-event callback for compatibility only
	// when this receiver is absent; application dispatch supplies it so a
	// source cannot mistake callback return for a durable downstream receipt.
	TransactionReceiver CDCTransactionReceiver
	// CheckpointCommitter receives the next source state only after its page's
	// emitted events have been durably accepted by the caller.
	CheckpointCommitter ChangefeedCheckpointCommitter
	// DurableCheckpointCommitter persists the product-wide opaque checkpoint
	// envelope. Native sources that need source identity and protocol positions
	// use this port instead of serializing a parallel scalar state map.
	DurableCheckpointCommitter DurableChangefeedCheckpointCommitter
}

type CDCEvent struct {
	Operation string `json:"operation"`
	Record    Record `json:"record"`
	State     Record `json:"state,omitempty"`
}

// CDCTransaction is one source-committed transaction whose events can be
// consumed once with bounded memory. Its identity is opaque-safe and contains
// no provider transaction value.
type CDCTransaction struct {
	id            string
	records       int64
	contentDigest string
	stream        func(context.Context, func(CDCEvent) error) error
}

// NewCDCTransaction constructs a committed transaction around a one-shot
// event stream. Native changefeeds use it only after their own commit boundary.
func NewCDCTransaction(id string, records int64, stream func(context.Context, func(CDCEvent) error) error) (CDCTransaction, error) {
	return newCDCTransaction(id, records, "", stream)
}

// NewCDCTransactionWithContentDigest constructs a committed transaction whose
// source-stage content digest is required by an artifact-bound durable receipt.
func NewCDCTransactionWithContentDigest(id string, records int64, contentDigest string, stream func(context.Context, func(CDCEvent) error) error) (CDCTransaction, error) {
	if !validCDCArtifactDigest(contentDigest) {
		return CDCTransaction{}, errors.New("CDC transaction content digest is invalid")
	}
	return newCDCTransaction(id, records, contentDigest, stream)
}

func newCDCTransaction(id string, records int64, contentDigest string, stream func(context.Context, func(CDCEvent) error) error) (CDCTransaction, error) {
	if strings.TrimSpace(id) == "" || len(id) > 1024 {
		return CDCTransaction{}, errors.New("CDC transaction identity is invalid")
	}
	if records < 0 {
		return CDCTransaction{}, errors.New("CDC transaction record count cannot be negative")
	}
	if stream == nil {
		return CDCTransaction{}, errors.New("CDC transaction event stream is required")
	}
	return CDCTransaction{id: id, records: records, contentDigest: contentDigest, stream: stream}, nil
}

// ID returns the opaque-safe committed transaction identity.
func (t CDCTransaction) ID() string { return t.id }

// Records returns the source-declared event count.
func (t CDCTransaction) Records() int64 { return t.records }

// ContentDigest is the exact source-stage content identity when the native
// changefeed has one. An empty value cannot bind a warehouse recovery receipt.
func (t CDCTransaction) ContentDigest() string { return t.contentDigest }

// StreamEvents visits the transaction in source order exactly once.
func (t CDCTransaction) StreamEvents(ctx context.Context, emit func(CDCEvent) error) error {
	if ctx == nil {
		return errors.New("CDC transaction stream context is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if emit == nil || t.stream == nil {
		return errors.New("CDC transaction event callback is required")
	}
	return t.stream(ctx, emit)
}

// CDCTransactionReceipt is durable downstream evidence for one complete CDC
// transaction. Its acknowledgement is constructible only through the sync
// contract's durable acknowledgement constructor.
type CDCTransactionReceipt struct {
	id               string
	acknowledgement  synccontract.DownstreamAcknowledgement
	artifactManifest string
	durable          bool
}

const cdcArtifactManifestVersion = 1

// CDCArtifactManifest is the private, versioned evidence binding a durable
// CDC receipt to the exact connection-owned warehouse generation and files.
// It is never included in public run or provider output.
type CDCArtifactManifest struct {
	Version          int    `json:"version"`
	ConnectionID     string `json:"connection_id"`
	Stream           string `json:"stream"`
	GenerationID     int64  `json:"generation_id"`
	TransactionKey   string `json:"transaction_key"`
	Records          int64  `json:"records"`
	ContentSHA256    string `json:"content_sha256"`
	RawWALSHA256     string `json:"raw_wal_sha256"`
	FinalTableSHA256 string `json:"final_table_sha256"`
}

// NewCDCArtifactManifest makes the exact source/staged transaction identity
// available only to the durable recovery-receipt path.
func NewCDCArtifactManifest(connectionID, stream string, generationID int64, transactionKey string, records int64, contentSHA256, rawWALSHA256, finalTableSHA256 string) (CDCArtifactManifest, error) {
	manifest := CDCArtifactManifest{
		Version:          cdcArtifactManifestVersion,
		ConnectionID:     connectionID,
		Stream:           stream,
		GenerationID:     generationID,
		TransactionKey:   transactionKey,
		Records:          records,
		ContentSHA256:    contentSHA256,
		RawWALSHA256:     rawWALSHA256,
		FinalTableSHA256: finalTableSHA256,
	}
	if err := manifest.Validate(); err != nil {
		return CDCArtifactManifest{}, err
	}
	return manifest, nil
}

// Validate refuses incomplete, unbounded, or non-digest artifact evidence
// before it can bind a recovered source checkpoint.
func (m CDCArtifactManifest) Validate() error {
	if m.Version != cdcArtifactManifestVersion || strings.TrimSpace(m.ConnectionID) == "" || len(m.ConnectionID) > 1024 ||
		strings.TrimSpace(m.Stream) == "" || len(m.Stream) > 1024 || m.GenerationID <= 0 ||
		strings.TrimSpace(m.TransactionKey) == "" || len(m.TransactionKey) > 1024 || m.Records < 0 ||
		!validCDCArtifactDigest(m.ContentSHA256) || !validCDCArtifactDigest(m.RawWALSHA256) || !validCDCArtifactDigest(m.FinalTableSHA256) {
		return errors.New("CDC artifact manifest is invalid")
	}
	return nil
}

// NewCDCTransactionReceipt constructs a receipt after the named sink has made
// the complete transaction durable.
func NewCDCTransactionReceipt(id, sink string, durableAt time.Time) (CDCTransactionReceipt, error) {
	return newCDCTransactionReceipt(id, sink, durableAt, "")
}

// NewCDCTransactionReceiptWithArtifactManifest creates a private
// artifact-bound receipt after all manifest artifacts are durable.
func NewCDCTransactionReceiptWithArtifactManifest(id, sink string, durableAt time.Time, manifest CDCArtifactManifest) (CDCTransactionReceipt, error) {
	if err := manifest.Validate(); err != nil {
		return CDCTransactionReceipt{}, err
	}
	payload, err := json.Marshal(manifest)
	if err != nil {
		return CDCTransactionReceipt{}, fmt.Errorf("encode CDC artifact manifest: %w", err)
	}
	return newCDCTransactionReceipt(id, sink, durableAt, string(payload))
}

// NewCDCTransactionReceiptWithArtifactManifestJSON restores the bounded
// private manifest persisted by a native transaction stage; it accepts no
// public provider body or caller-selected path.
func NewCDCTransactionReceiptWithArtifactManifestJSON(id, sink string, durableAt time.Time, payload string) (CDCTransactionReceipt, error) {
	manifest, canonical, err := parseCDCArtifactManifest(payload)
	if err != nil {
		return CDCTransactionReceipt{}, err
	}
	if _, err := NewCDCTransactionReceiptWithArtifactManifest(id, sink, durableAt, manifest); err != nil {
		return CDCTransactionReceipt{}, err
	}
	return newCDCTransactionReceipt(id, sink, durableAt, canonical)
}

func newCDCTransactionReceipt(id, sink string, durableAt time.Time, artifactManifest string) (CDCTransactionReceipt, error) {
	if strings.TrimSpace(id) == "" || len(id) > 1024 {
		return CDCTransactionReceipt{}, errors.New("CDC transaction receipt identity is invalid")
	}
	acknowledgement, err := synccontract.NewDurableDownstreamAcknowledgement(sink, durableAt)
	if err != nil {
		return CDCTransactionReceipt{}, err
	}
	return CDCTransactionReceipt{id: id, acknowledgement: acknowledgement, artifactManifest: artifactManifest, durable: true}, nil
}

// ID returns the receiver-owned durable receipt identity.
func (r CDCTransactionReceipt) ID() string { return r.id }

// HasArtifactManifest reports whether this private durable receipt carries
// exact warehouse artifact evidence.
func (r CDCTransactionReceipt) HasArtifactManifest() bool {
	return r.durable && r.artifactManifest != ""
}

// Acknowledgement returns the checkpoint admission produced by this receipt.
func (r CDCTransactionReceipt) Acknowledgement() (synccontract.DownstreamAcknowledgement, error) {
	if !r.durable || strings.TrimSpace(r.id) == "" {
		return synccontract.DownstreamAcknowledgement{}, errors.New("durable CDC transaction receipt is unavailable")
	}
	return r.acknowledgement, nil
}

// ArtifactManifest returns private exact artifact evidence. It is deliberately
// unavailable for ordinary unbound CDC receipts and never becomes result data.
func (r CDCTransactionReceipt) ArtifactManifest() (CDCArtifactManifest, error) {
	if !r.durable || strings.TrimSpace(r.id) == "" {
		return CDCArtifactManifest{}, errors.New("durable CDC transaction receipt is unavailable")
	}
	manifest, _, err := parseCDCArtifactManifest(r.artifactManifest)
	return manifest, err
}

// ArtifactManifestJSON returns a validated canonical private payload for the
// transaction-stage receipt store. It is not a public output projection.
func (r CDCTransactionReceipt) ArtifactManifestJSON() (string, error) {
	if !r.durable || strings.TrimSpace(r.id) == "" {
		return "", errors.New("durable CDC transaction receipt is unavailable")
	}
	_, canonical, err := parseCDCArtifactManifest(r.artifactManifest)
	return canonical, err
}

func parseCDCArtifactManifest(payload string) (CDCArtifactManifest, string, error) {
	if len(payload) == 0 || len(payload) > 8<<10 {
		return CDCArtifactManifest{}, "", errors.New("CDC artifact manifest is unavailable")
	}
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.DisallowUnknownFields()
	var manifest CDCArtifactManifest
	if err := decoder.Decode(&manifest); err != nil {
		return CDCArtifactManifest{}, "", fmt.Errorf("decode CDC artifact manifest: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return CDCArtifactManifest{}, "", errors.New("CDC artifact manifest has trailing values")
		}
		return CDCArtifactManifest{}, "", fmt.Errorf("decode CDC artifact manifest trailing value: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return CDCArtifactManifest{}, "", err
	}
	canonical, err := json.Marshal(manifest)
	if err != nil {
		return CDCArtifactManifest{}, "", fmt.Errorf("encode CDC artifact manifest: %w", err)
	}
	return manifest, string(canonical), nil
}

func validCDCArtifactDigest(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	for _, character := range value {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

// CDCTransactionReceiver makes one complete source transaction durable and
// returns its receipt. It is deliberately separate from CDCReader so a
// per-event callback cannot be promoted into a durability claim.
type CDCTransactionReceiver interface {
	ReceiveCDCTransaction(context.Context, CDCTransaction) (CDCTransactionReceipt, error)
}

// CDCTransactionReceiptRestorer re-attaches a crash-recovered source-stage
// receipt to the downstream receiver that originally produced it. A source
// calls it before retrying the receipt's checkpoint; events are not re-emitted.
type CDCTransactionReceiptRestorer interface {
	RestoreCDCTransactionReceipt(context.Context, string, CDCTransactionReceipt) error
}

type WriteValidator interface {
	ValidateWrite(ctx context.Context, req WriteRequest, records []Record) error
}

type DeclarativeTypedDestination interface {
	Connector
	DefinitionProvider
	WriteValidator
	PreflightWriteRecordFieldMapping(actionName string, fields []string) error
	DeclarativeTypedDestinationActionDigest(actionName string) (string, error)
	// DeclarativeTypedDestinationIdempotencyHeader returns the provider-owned
	// header from the compiled action. A transport delivery claim alone is not
	// evidence that an action can safely replay an ambiguous write.
	DeclarativeTypedDestinationIdempotencyHeader(actionName string) (string, error)
}

// DeclarativeTypedDestinationReadBackRequest is a declaration-owned bounded
// provider-state read. Operation remains opaque to shared orchestration.
type DeclarativeTypedDestinationReadBackRequest struct {
	Operation              string
	Runtime                RuntimeConfig
	MaxRecords             int
	Receipt                json.RawMessage
	ReceiptLocator         DestinationReceiptLocator
	ActionDefinitionSHA256 string
}

// DeclarativeTypedDestinationReadBackReceipt is private, bounded evidence
// extracted from provider write responses. It is only decoded by the exact
// connector-owned read-back method and never becomes printable result output.
type DeclarativeTypedDestinationReadBackReceipt struct {
	Version                int      `json:"version"`
	ActionDefinitionSHA256 string   `json:"action_definition_sha256"`
	Locators               []string `json:"locators"`
}

// NewDeclarativeTypedDestinationReadBackReceipt creates the private, bounded
// bridge from a successful typed write to its declared read-back. Callers pass
// only scalars already extracted under the declaration's response locator.
func NewDeclarativeTypedDestinationReadBackReceipt(actionDefinitionSHA256 string, locator DestinationReceiptLocator, locators []string, maxRecords int) (json.RawMessage, error) {
	if err := locator.Validate(); err != nil {
		return nil, fmt.Errorf("declarative destination receipt locator: %w", err)
	}
	if len(strings.TrimSpace(actionDefinitionSHA256)) != sha256.Size*2 {
		return nil, fmt.Errorf("declarative destination receipt requires an action definition digest")
	}
	if maxRecords < 1 || len(locators) == 0 || len(locators) > maxRecords {
		return nil, fmt.Errorf("declarative destination receipt locator count is outside its read-back bound")
	}
	copyLocators := make([]string, len(locators))
	for index, value := range locators {
		if value == "" || len(value) > locator.MaxValueBytes {
			return nil, fmt.Errorf("declarative destination receipt locator value is outside its byte bound")
		}
		copyLocators[index] = value
	}
	receipt, err := json.Marshal(DeclarativeTypedDestinationReadBackReceipt{
		Version:                1,
		ActionDefinitionSHA256: actionDefinitionSHA256,
		Locators:               copyLocators,
	})
	if err != nil {
		return nil, fmt.Errorf("encode declarative destination read-back receipt: %w", err)
	}
	if len(receipt) > synccontract.MaxPrivateReceiptBytes {
		return nil, fmt.Errorf("declarative destination receipt exceeds its byte bound")
	}
	return receipt, nil
}

// ParseDeclarativeTypedDestinationReadBackReceipt validates the exact private
// receipt before it can drive a declared query parameter. It never accepts a
// fallback output body, route, or user-provided selector.
func ParseDeclarativeTypedDestinationReadBackReceipt(raw json.RawMessage, actionDefinitionSHA256 string, locator DestinationReceiptLocator, maxRecords int) ([]string, error) {
	if len(raw) == 0 || len(raw) > synccontract.MaxPrivateReceiptBytes {
		return nil, fmt.Errorf("declarative destination read-back requires a bounded private receipt")
	}
	if err := locator.Validate(); err != nil {
		return nil, fmt.Errorf("declarative destination receipt locator: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var receipt DeclarativeTypedDestinationReadBackReceipt
	if err := decoder.Decode(&receipt); err != nil {
		return nil, fmt.Errorf("decode declarative destination read-back receipt: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("declarative destination read-back receipt contains multiple JSON values")
		}
		return nil, fmt.Errorf("decode declarative destination read-back receipt: %w", err)
	}
	if receipt.Version != 1 || !strings.EqualFold(receipt.ActionDefinitionSHA256, actionDefinitionSHA256) || maxRecords < 1 || len(receipt.Locators) == 0 || len(receipt.Locators) > maxRecords {
		return nil, fmt.Errorf("declarative destination read-back receipt does not match the declared action")
	}
	locators := make([]string, len(receipt.Locators))
	for index, value := range receipt.Locators {
		if value == "" || len(value) > locator.MaxValueBytes {
			return nil, fmt.Errorf("declarative destination read-back receipt locator is outside its byte bound")
		}
		locators[index] = value
	}
	return locators, nil
}

// DeclarativeTypedDestinationReadBack is required by the generic typed
// destination adapter. Connectors without a real provider read cannot claim
// provider-state read-back from local receipt checks.
type DeclarativeTypedDestinationReadBack interface {
	ReadBackDeclarativeDestination(context.Context, DeclarativeTypedDestinationReadBackRequest) ([]Record, error)
}

type DryRunWriter interface {
	DryRunWrite(ctx context.Context, req WriteRequest, records []Record) (WritePreview, error)
}

type Querier interface {
	Query(ctx context.Context, req QueryRequest) (QueryResult, error)
}

type CDCReader interface {
	ReadCDC(ctx context.Context, req CDCReadRequest, emit func(CDCEvent) error) error
}

// ChangefeedCheckpointCommitter persists a source position only after the
// caller's CDC event callback has durably accepted that position's records.
// The polling-watermark executor treats the state as opaque key/value data so
// the durable, versioned #3810 sync envelope can replace this narrow adapter
// without changing transport or delivery logic.
type ChangefeedCheckpointCommitter interface {
	CommitChangefeedCheckpoint(ctx context.Context, state map[string]string) error
}

// DurableChangefeedCheckpointCommitter receives a fully structured
// synccontract envelope after the caller has durably accepted the emitted
// source transaction. Implementations commit through
// synccontract.CommitAfterDownstreamAcknowledgement; the connector never
// treats a received record or an unacknowledged candidate as resumable state.
type DurableChangefeedCheckpointCommitter interface {
	CommitDurableChangefeedCheckpoint(context.Context, synccontract.CheckpointEnvelope) error
}

// ChangefeedStatus is the closed lifecycle vocabulary for a declared
// changefeed. A status is not itself an executable capability: only an
// implemented descriptor with a matching ChangefeedExecutor is discoverable.
type ChangefeedStatus string

const (
	ChangefeedStatusImplemented ChangefeedStatus = "implemented"
	ChangefeedStatusPlanned     ChangefeedStatus = "planned"
	ChangefeedStatusUnsupported ChangefeedStatus = "unsupported"
	ChangefeedStatusUnknown     ChangefeedStatus = "unknown"
)

// ChangefeedMechanism is the closed taxonomy of consumable source mechanisms.
// A snapshot pagination cursor is not a changefeed unless its descriptor
// establishes one of these contracts.
type ChangefeedMechanism string

const (
	ChangefeedMechanismLogicalReplication ChangefeedMechanism = "logical_replication"
	// ChangefeedMechanismBinlogReplication identifies MySQL/MariaDB's binary
	// log replication protocol. It is distinct from PostgreSQL logical
	// replication: each mechanism has different server prerequisites,
	// checkpoint positions, and decoders.
	ChangefeedMechanismBinlogReplication ChangefeedMechanism = "binlog_replication"
	ChangefeedMechanismIncrementalCursor ChangefeedMechanism = "incremental_cursor"
	ChangefeedMechanismWebhook           ChangefeedMechanism = "webhook"
	ChangefeedMechanismEventStream       ChangefeedMechanism = "event_stream"
	ChangefeedMechanismPollingWatermark  ChangefeedMechanism = "polling_watermark"
)

// ChangefeedSource records the provider artifact that supports a declared
// changefeed contract. The retrieval date is a date-only ISO-8601 value.
type ChangefeedSource struct {
	ArtifactURL     string `json:"artifact_url"`
	ArtifactVersion string `json:"artifact_version"`
	RetrievedAt     string `json:"retrieved_at"`
}

// ChangefeedExecutorRef names the fixed executor selected by an implemented
// descriptor. It is never supplied by a CLI caller.
type ChangefeedExecutorRef struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// ChangefeedCheckpoint declares the durable position and recovery contract.
// State advances only after the descriptor's CommitAfter condition is met.
type ChangefeedCheckpoint struct {
	Kind        string   `json:"kind"`
	Keys        []string `json:"keys"`
	CommitAfter string   `json:"commit_after"`
	OnInvalid   string   `json:"on_invalid"`
}

// ChangefeedDelivery records the source ordering, duplicate, and delete
// guarantees. Values are provider-specific declarations rather than generic
// promises made by the CLI.
type ChangefeedDelivery struct {
	Ordering   string   `json:"ordering"`
	Duplicates string   `json:"duplicates"`
	Deletes    string   `json:"deletes"`
	DedupeKey  []string `json:"dedupe_key,omitempty"`
}

// PollingWatermarkField identifies a provider record value using a dotted
// object path. The value is never supplied by a CLI caller.
type PollingWatermarkField struct {
	Path string `json:"path"`
}

// PollingWatermarkValue identifies the provider ordering value and its
// lossless representation. Timestamp, monotonic sequence, and opaque cursor
// values intentionally have different validation and safety-lag semantics.
type PollingWatermarkValue struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}

// PollingWatermarkDeletionEndpoint declares a provider deletion feed that a
// polling source adapter must fetch in addition to ordinary records. It is
// deliberately descriptive: the shared executor owns event/checkpoint
// semantics while a closed connector transport owns the provider request.
type PollingWatermarkDeletionEndpoint struct {
	Path        string `json:"path"`
	RecordsPath string `json:"records_path"`
}

// PollingWatermarkSpec is the complete declaration consumed by the shared
// polling-watermark executor. Inclusive boundaries deliberately replay the
// page edge, so delivery is at-least-once rather than falsely exactly-once.
// A timestamp safety lag of zero is an explicit opt-out and can lose late
// arrivals whose provider clock is behind the committed watermark.
type PollingWatermarkSpec struct {
	Watermark        PollingWatermarkValue             `json:"watermark"`
	TieBreaker       PollingWatermarkField             `json:"tie_breaker"`
	Boundary         string                            `json:"boundary"`
	SafetyLagSeconds int                               `json:"safety_lag_seconds"`
	PageSize         int                               `json:"page_size"`
	MaxPages         int                               `json:"max_pages"`
	RequestBudget    int                               `json:"request_budget"`
	SoftDelete       *PollingWatermarkField            `json:"soft_delete,omitempty"`
	DeletionEndpoint *PollingWatermarkDeletionEndpoint `json:"deletion_endpoint,omitempty"`
}

const maxPollingWatermarkSafetyLagSeconds = int64(1<<63-1) / int64(time.Second)

// ChangefeedDescriptor is the evidence-backed, declarative contract for a
// connector changefeed. It is distinct from CDCReader because a reader method
// can be an unsupported migration stub.
type ChangefeedDescriptor struct {
	Status           ChangefeedStatus       `json:"status"`
	Mechanism        ChangefeedMechanism    `json:"mechanism"`
	Source           ChangefeedSource       `json:"source"`
	Reason           string                 `json:"reason,omitempty"`
	Executor         *ChangefeedExecutorRef `json:"executor,omitempty"`
	Checkpoint       *ChangefeedCheckpoint  `json:"checkpoint,omitempty"`
	Delivery         *ChangefeedDelivery    `json:"delivery,omitempty"`
	Streams          []string               `json:"streams,omitempty"`
	PollingWatermark *PollingWatermarkSpec  `json:"polling_watermark,omitempty"`
}

// ChangefeedExecutorDescriptor is the runtime half of a declared changefeed.
// It reports the executor lifecycle and mechanism; an implemented declaration
// also supplies the executor and checkpoint contract needed for promotion.
// Provider evidence remains bundle-owned.
type ChangefeedExecutorDescriptor struct {
	Status     ChangefeedStatus      `json:"status"`
	Mechanism  ChangefeedMechanism   `json:"mechanism"`
	Executor   ChangefeedExecutorRef `json:"executor"`
	Checkpoint ChangefeedCheckpoint  `json:"checkpoint"`
}

// ChangefeedDescriptorProvider reports the lifecycle and mechanism associated
// with a CDC reader. It is deliberately separate from CDCReader so method-set
// coincidence cannot advertise capability; only a matching implemented
// declaration can do that.
type ChangefeedDescriptorProvider interface {
	ChangefeedExecutorDescriptor() ChangefeedExecutorDescriptor
}

// ChangefeedExecutor is the explicit runtime admission contract. A connector
// must both execute the legacy migration method and report a matching modern
// descriptor before an implemented bundle declaration becomes public CDC.
type ChangefeedExecutor interface {
	CDCReader
	ChangefeedDescriptorProvider
}

// Validate verifies the descriptor's closed vocabulary and the minimum
// evidence needed to make an implemented or unsupported claim. It is used at
// bundle-load time and never treats an incomplete declaration as executable.
func (d ChangefeedDescriptor) Validate() error {
	if !d.Status.valid() {
		return fmt.Errorf("unsupported changefeed status %q", d.Status)
	}
	if !d.Mechanism.valid() {
		return fmt.Errorf("unsupported changefeed mechanism %q", d.Mechanism)
	}
	if strings.TrimSpace(d.Source.ArtifactURL) == "" || strings.TrimSpace(d.Source.ArtifactVersion) == "" || strings.TrimSpace(d.Source.RetrievedAt) == "" {
		return errors.New("changefeed source requires artifact_url, artifact_version, and retrieved_at")
	}
	artifactURL, err := url.ParseRequestURI(d.Source.ArtifactURL)
	if err != nil || artifactURL.Host == "" || (artifactURL.Scheme != "http" && artifactURL.Scheme != "https") {
		return errors.New("changefeed source artifact_url must be an absolute http or https URL")
	}
	if _, err := time.Parse("2006-01-02", d.Source.RetrievedAt); err != nil {
		return fmt.Errorf("changefeed source retrieved_at must be an ISO-8601 date: %w", err)
	}

	switch d.Status {
	case ChangefeedStatusImplemented:
		if d.Executor == nil || strings.TrimSpace(d.Executor.Kind) == "" || strings.TrimSpace(d.Executor.ID) == "" {
			return errors.New("implemented changefeed requires a named executor")
		}
		if err := validateChangefeedCheckpoint(d.Checkpoint); err != nil {
			return err
		}
		if err := validateChangefeedDelivery(d.Delivery); err != nil {
			return err
		}
		if err := validateChangefeedKeys("streams", d.Streams); err != nil {
			return err
		}
		if d.Mechanism == ChangefeedMechanismPollingWatermark {
			if err := validatePollingWatermark(d); err != nil {
				return err
			}
		} else if d.PollingWatermark != nil {
			return errors.New("polling_watermark declaration requires polling_watermark mechanism")
		}
	case ChangefeedStatusUnsupported:
		if strings.TrimSpace(d.Reason) == "" {
			return errors.New("unsupported changefeed requires a reason")
		}
		if d.Executor != nil || d.Checkpoint != nil || d.Delivery != nil || d.PollingWatermark != nil {
			return errors.New("unsupported changefeed cannot declare an executor, checkpoint, delivery, or polling watermark")
		}
	}
	return nil
}

// IsImplemented reports whether the declaration is complete enough to be
// considered for execution. A true result still needs a matching executor.
func (d ChangefeedDescriptor) IsImplemented() bool {
	return d.Status == ChangefeedStatusImplemented && d.Validate() == nil
}

// MatchesExecutor reports whether executor is the runtime counterpart of this
// implemented descriptor. Matching requires the status, mechanism, named
// executor, and checkpoint contract to agree exactly.
func (d ChangefeedDescriptor) MatchesExecutor(executor ChangefeedExecutorDescriptor) bool {
	if !d.IsImplemented() || executor.Status != ChangefeedStatusImplemented || d.Executor == nil || d.Checkpoint == nil {
		return false
	}
	return d.Mechanism == executor.Mechanism &&
		d.Executor.Kind == executor.Executor.Kind &&
		d.Executor.ID == executor.Executor.ID &&
		d.Checkpoint.Kind == executor.Checkpoint.Kind &&
		sameStrings(d.Checkpoint.Keys, executor.Checkpoint.Keys) &&
		d.Checkpoint.CommitAfter == executor.Checkpoint.CommitAfter &&
		d.Checkpoint.OnInvalid == executor.Checkpoint.OnInvalid
}

// Clone returns a defensive copy suitable for a public Definition projection.
func (d ChangefeedDescriptor) Clone() *ChangefeedDescriptor {
	clone := d
	clone.Streams = append([]string(nil), d.Streams...)
	if d.Executor != nil {
		executor := *d.Executor
		clone.Executor = &executor
	}
	if d.Checkpoint != nil {
		checkpoint := *d.Checkpoint
		checkpoint.Keys = append([]string(nil), d.Checkpoint.Keys...)
		clone.Checkpoint = &checkpoint
	}
	if d.Delivery != nil {
		delivery := *d.Delivery
		delivery.DedupeKey = append([]string(nil), d.Delivery.DedupeKey...)
		clone.Delivery = &delivery
	}
	if d.PollingWatermark != nil {
		polling := *d.PollingWatermark
		if d.PollingWatermark.SoftDelete != nil {
			softDelete := *d.PollingWatermark.SoftDelete
			polling.SoftDelete = &softDelete
		}
		if d.PollingWatermark.DeletionEndpoint != nil {
			deletionEndpoint := *d.PollingWatermark.DeletionEndpoint
			polling.DeletionEndpoint = &deletionEndpoint
		}
		clone.PollingWatermark = &polling
	}
	return &clone
}

// HasImplementedChangefeed is the single capability projection rule. Legacy
// CDCReader presence alone is insufficient; callers must use this helper when
// deciding whether to expose CDC publicly.
func HasImplementedChangefeed(c Connector, descriptor *ChangefeedDescriptor) bool {
	if descriptor == nil || !descriptor.IsImplemented() {
		return false
	}
	executor, ok := c.(ChangefeedExecutor)
	if !ok {
		return false
	}
	return descriptor.MatchesExecutor(executor.ChangefeedExecutorDescriptor())
}

func (s ChangefeedStatus) valid() bool {
	switch s {
	case ChangefeedStatusImplemented, ChangefeedStatusPlanned, ChangefeedStatusUnsupported, ChangefeedStatusUnknown:
		return true
	default:
		return false
	}
}

func (m ChangefeedMechanism) valid() bool {
	switch m {
	case ChangefeedMechanismLogicalReplication, ChangefeedMechanismBinlogReplication, ChangefeedMechanismIncrementalCursor, ChangefeedMechanismWebhook, ChangefeedMechanismEventStream, ChangefeedMechanismPollingWatermark:
		return true
	default:
		return false
	}
}

func validateChangefeedCheckpoint(checkpoint *ChangefeedCheckpoint) error {
	if checkpoint == nil || strings.TrimSpace(checkpoint.Kind) == "" || strings.TrimSpace(checkpoint.CommitAfter) == "" || strings.TrimSpace(checkpoint.OnInvalid) == "" || len(checkpoint.Keys) == 0 {
		return errors.New("implemented changefeed requires checkpoint kind, keys, commit_after, and on_invalid")
	}
	seen := make(map[string]struct{}, len(checkpoint.Keys))
	for _, key := range checkpoint.Keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return errors.New("changefeed checkpoint keys cannot be empty")
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("changefeed checkpoint key %q is duplicated", key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func validateChangefeedDelivery(delivery *ChangefeedDelivery) error {
	if delivery == nil || strings.TrimSpace(delivery.Ordering) == "" || strings.TrimSpace(delivery.Duplicates) == "" || strings.TrimSpace(delivery.Deletes) == "" {
		return errors.New("implemented changefeed requires ordering, duplicates, and deletes guarantees")
	}
	if err := validateChangefeedKeys("delivery dedupe_key", delivery.DedupeKey); err != nil {
		return err
	}
	return nil
}

func validatePollingWatermark(d ChangefeedDescriptor) error {
	polling := d.PollingWatermark
	if polling == nil {
		return errors.New("implemented polling_watermark changefeed requires polling_watermark declaration")
	}
	if d.Executor == nil || d.Executor.Kind != "engine" || d.Executor.ID != "polling_watermark" {
		return errors.New("implemented polling_watermark changefeed requires executor engine/polling_watermark")
	}
	if err := validatePollingWatermarkPath("watermark path", polling.Watermark.Path); err != nil {
		return err
	}
	if err := validatePollingWatermarkPath("tie_breaker path", polling.TieBreaker.Path); err != nil {
		return err
	}
	switch polling.Watermark.Kind {
	case "timestamp", "monotonic_sequence", "opaque_cursor":
	default:
		return fmt.Errorf("unsupported polling watermark kind %q", polling.Watermark.Kind)
	}
	if polling.Boundary != "inclusive" {
		return errors.New("polling watermark boundary must be inclusive to prevent tie loss")
	}
	if polling.SafetyLagSeconds < 0 {
		return errors.New("polling watermark safety_lag_seconds cannot be negative")
	}
	if int64(polling.SafetyLagSeconds) > maxPollingWatermarkSafetyLagSeconds {
		return fmt.Errorf("polling watermark safety_lag_seconds exceeds the maximum duration-safe value of %d", maxPollingWatermarkSafetyLagSeconds)
	}
	if polling.Watermark.Kind != "timestamp" && polling.SafetyLagSeconds != 0 {
		return errors.New("polling watermark safety_lag_seconds is only valid for timestamp watermarks")
	}
	if polling.PageSize <= 0 || polling.MaxPages <= 0 || polling.RequestBudget <= 0 {
		return errors.New("polling watermark requires positive page_size, max_pages, and request_budget")
	}
	if polling.DeletionEndpoint != nil && polling.RequestBudget < 2 {
		return errors.New("polling watermark deletion_endpoint requires request_budget of at least 2")
	}
	if d.Checkpoint == nil || !sameStrings(d.Checkpoint.Keys, []string{polling.Watermark.Path, polling.TieBreaker.Path}) {
		return errors.New("polling watermark checkpoint keys must be watermark path then tie_breaker path")
	}
	if d.Delivery == nil || d.Delivery.Duplicates != "at_least_once" {
		return errors.New("polling watermark delivery must declare duplicates at_least_once")
	}
	if polling.SoftDelete != nil && polling.DeletionEndpoint != nil {
		return errors.New("polling watermark may declare either soft_delete or deletion_endpoint, not both")
	}
	if polling.SoftDelete != nil {
		if err := validatePollingWatermarkPath("soft_delete path", polling.SoftDelete.Path); err != nil {
			return err
		}
	}
	if polling.DeletionEndpoint != nil {
		if err := validatePollingWatermarkEndpointPath(polling.DeletionEndpoint.Path); err != nil {
			return err
		}
		if err := validatePollingWatermarkPath("deletion_endpoint records_path", polling.DeletionEndpoint.RecordsPath); err != nil {
			return err
		}
	}
	if polling.SoftDelete == nil && polling.DeletionEndpoint == nil && d.Delivery.Deletes != "not_available" {
		return errors.New("polling watermark hard deletes are not observable; delivery deletes must be not_available")
	}
	if (polling.SoftDelete != nil || polling.DeletionEndpoint != nil) && d.Delivery.Deletes != "tombstone" {
		return errors.New("polling watermark observable deletes must declare tombstone delivery")
	}
	return nil
}

func validatePollingWatermarkEndpointPath(path string) error {
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "//") || strings.Contains(path, "..") || strings.ContainsAny(path, "\r\n?#") {
		return fmt.Errorf("polling watermark deletion_endpoint path %q is not a safe connector-relative path", path)
	}
	return nil
}

func validatePollingWatermarkPath(name, path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("polling watermark %s is required", name)
	}
	for _, segment := range strings.Split(path, ".") {
		if segment == "" {
			return fmt.Errorf("polling watermark %s contains an empty path segment", name)
		}
		for index, runeValue := range segment {
			if (runeValue >= 'a' && runeValue <= 'z') || (runeValue >= 'A' && runeValue <= 'Z') || runeValue == '_' || runeValue == '-' || (index > 0 && runeValue >= '0' && runeValue <= '9') {
				continue
			}
			return fmt.Errorf("polling watermark %s contains unsafe path segment %q", name, segment)
		}
	}
	return nil
}

func validateChangefeedKeys(name string, keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("implemented changefeed requires %s", name)
	}
	seen := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			return fmt.Errorf("changefeed %s cannot contain empty values", name)
		}
		if _, ok := seen[key]; ok {
			return fmt.Errorf("changefeed %s value %q is duplicated", name, key)
		}
		seen[key] = struct{}{}
	}
	return nil
}

func sameStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

type StatefulReader interface {
	InitialState(ctx context.Context, stream string, cfg RuntimeConfig) (map[string]string, error)
}

type SchemaMapper interface {
	MapSchema(ctx context.Context, stream Stream) (Stream, error)
}

type LiveConformanceProvider interface {
	LiveConformanceConfig(ctx context.Context) (RuntimeConfig, bool, error)
}

type Connector interface {
	Name() string
	Metadata() Metadata
	Check(ctx context.Context, cfg RuntimeConfig) error
	Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error)
	Read(ctx context.Context, req ReadRequest, emit func(Record) error) error
	Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error)
}

type LocalWarehouseMaterializer interface {
	MaterializesLocalWarehouse() bool
}

type Registry struct {
	connectors            map[string]Connector
	iconCoverageValidated bool
}

func NewEmptyRegistry() *Registry {
	return &Registry{connectors: make(map[string]Connector)}
}

func NewRegistry() *Registry {
	if builder := registeredDefaultRegistryBuilder(); builder != nil {
		registry := builder()
		registry.MustValidateIconCoverage()
		return registry
	}
	r := NewEmptyRegistry()
	r.RegisterBuiltins()
	r.MustValidateIconCoverage()
	return r
}

// RegisterBuiltins adds the primitive local connectors that are implemented in
// this package rather than in defs/. They are not legacy per-connector packages.
func (r *Registry) RegisterBuiltins() {
	r.Register(Sample{})
	r.Register(File{})
	r.Register(Warehouse{})
	r.Register(Outbox{})
}

func (r *Registry) Register(c Connector) {
	r.connectors[c.Name()] = c
	r.iconCoverageValidated = false
}

func (r *Registry) Get(name string) (Connector, bool) {
	c, ok := r.connectors[name]
	return c, ok
}

func (r *Registry) List() []Metadata {
	out := make([]Metadata, 0, len(r.connectors))
	for _, connector := range r.connectors {
		out = append(out, MetadataOf(connector))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (r *Registry) CatalogEntries() []Definition {
	list := r.List()
	out := make([]Definition, 0, len(list))
	for _, item := range list {
		connector, ok := r.Get(item.Name)
		if !ok {
			continue
		}
		def, ok := DefinitionOf(connector)
		if !ok {
			manifest := ManifestOf(connector)
			def = Definition{
				Name:            manifest.Metadata.Name,
				DisplayName:     manifest.Metadata.DisplayName,
				Description:     manifest.Metadata.Description,
				IntegrationType: manifest.Metadata.IntegrationType,
				Capabilities:    manifest.Metadata.Capabilities,
				Streams:         streamSummariesFromManifest(manifest),
				WriteActions:    writeActionInfosFromManifest(manifest),
				Risk:            manifest.Risk,
			}
		}
		// Catalog is a public capability projection. The fallback definition
		// has no changefeed descriptor, so it remains false even if a legacy
		// connector happens to expose CDCReader or a hand-authored metadata bit.
		def.Capabilities.CDC = HasImplementedChangefeed(connector, def.Changefeed)
		def.Icon = MetadataOf(connector).Icon
		out = append(out, def)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

type Sample struct{}

func (Sample) Name() string { return "sample" }

func (Sample) Metadata() Metadata {
	return Metadata{
		Name:            "sample",
		DisplayName:     "Sample",
		IntegrationType: "api",
		Description:     "Built-in deterministic source connector for local development and tests.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (Sample) Check(ctx context.Context, cfg RuntimeConfig) error {
	return ctx.Err()
}

func (s Sample) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := ctx.Err(); err != nil {
		return Catalog{}, err
	}
	return Catalog{Connector: s.Name(), Streams: []Stream{
		{
			Name:        "customers",
			Description: "Sample customer records.",
			PrimaryKey:  []string{"id"},
			CursorFields: []string{
				"updated_at",
			},
			Fields: []Field{
				{Name: "id", Type: "string"},
				{Name: "name", Type: "string"},
				{Name: "email", Type: "string"},
				{Name: "plan", Type: "string"},
				{Name: "updated_at", Type: "timestamp"},
			},
		},
		{
			Name:         "events",
			Description:  "Sample event records.",
			PrimaryKey:   []string{"id"},
			CursorFields: []string{"occurred_at"},
			Fields: []Field{
				{Name: "id", Type: "string"},
				{Name: "customer_id", Type: "string"},
				{Name: "event", Type: "string"},
				{Name: "occurred_at", Type: "timestamp"},
			},
		},
	}}, nil
}

func (Sample) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	var records []Record
	switch req.Stream {
	case "customers", "":
		records = []Record{
			{"id": "cus_001", "name": "Ada Lovelace", "email": "ada@example.com", "plan": "enterprise", "updated_at": "2026-06-20T10:00:00Z"},
			{"id": "cus_002", "name": "Grace Hopper", "email": "grace@example.com", "plan": "team", "updated_at": "2026-06-21T12:30:00Z"},
			{"id": "cus_003", "name": "Katherine Johnson", "email": "katherine@example.com", "plan": "starter", "updated_at": "2026-06-22T09:15:00Z"},
		}
	case "events":
		records = []Record{
			{"id": "evt_001", "customer_id": "cus_001", "event": "signed_in", "occurred_at": "2026-06-22T10:00:00Z"},
			{"id": "evt_002", "customer_id": "cus_002", "event": "upgraded", "occurred_at": "2026-06-22T11:00:00Z"},
		}
	default:
		return fmt.Errorf("sample stream %q not found", req.Stream)
	}
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := emit(copyRecord(record)); err != nil {
			return err
		}
	}
	return nil
}

func (Sample) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	return WriteResult{}, ErrUnsupportedOperation
}

type File struct{}

func (File) Name() string { return "file" }

func (File) Metadata() Metadata {
	return Metadata{
		Name:            "file",
		DisplayName:     "File",
		IntegrationType: "file",
		Description:     "Reads local JSONL or CSV files as source streams.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Read: true},
	}
}

func (File) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := cfg.Config["path"]
	if path == "" {
		return errors.New("file connector requires config path")
	}
	_, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat file source %s: %w", path, err)
	}
	return nil
}

func (f File) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := f.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	stream := cfg.Config["stream"]
	if stream == "" {
		stream = strings.TrimSuffix(filepath.Base(cfg.Config["path"]), filepath.Ext(cfg.Config["path"]))
	}
	return Catalog{Connector: f.Name(), Streams: []Stream{{
		Name:        stream,
		Description: "Local file stream.",
	}}}, nil
}

func (File) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	path := req.Config.Config["path"]
	if path == "" {
		return errors.New("file connector requires config path")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open file source %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	switch strings.ToLower(filepath.Ext(path)) {
	case ".csv":
		return readCSV(ctx, file, emit)
	default:
		return readJSONL(ctx, file, emit)
	}
}

func (File) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	return WriteResult{}, ErrUnsupportedOperation
}

type Warehouse struct{}

// LocalWarehouseDestinationTransportReference is the concrete local Parquet
// materializer. Its declaration lives with the primitive, rather than being
// inferred from the legacy Connector.Write capability.
var LocalWarehouseDestinationTransportReference = TransportExecutorReference{
	Family: TransportExecutorFamilyNativeDatabase,
	ID:     "local_parquet_warehouse",
}

// LocalWarehouseDestinationTransportConformance is admitted only by the
// production composition's matching factory; a descriptor alone cannot admit
// an unregistered materializer.
var LocalWarehouseDestinationTransportConformance = ConformanceEvidenceReference{
	Suite: "local_parquet_warehouse",
	RunID: "connection_owned_v1",
}

const localWarehouseDestinationTransportAction = "materialize_local_parquet"

func (Warehouse) Name() string { return "warehouse" }

func (Warehouse) MaterializesLocalWarehouse() bool { return true }

// SyncTransportDescriptor declares the local warehouse's real, closed
// destination role. It intentionally does not infer that role from Read or
// Write: the reference-bound production adapter owns the durable application
// and read-back contract.
func (Warehouse) SyncTransportDescriptor() *SyncTransportDescriptor {
	return &SyncTransportDescriptor{Destination: &DestinationTransportDescriptor{
		Executor: LocalWarehouseDestinationTransportReference,
		EligibleActions: []string{
			localWarehouseDestinationTransportAction,
		},
		Modes: []synccontract.Mode{
			synccontract.ModeFullOverwrite,
			synccontract.ModeFullAppend,
			synccontract.ModeIncrementalUpsert,
			synccontract.ModeIncrementalDedupe,
			synccontract.ModeIncrementalDedupeHistory,
		},
		Delivery: DeliveryGuarantees{
			Idempotency: DeliveryIdempotencyKeyed,
			Ordering:    DeliveryOrderingSource,
			Deletes:     DeliveryDeletesTombstone,
		},
		Conformance:     LocalWarehouseDestinationTransportConformance,
		Acknowledgement: TransportAcknowledgementDurableWarehouse,
		ApplyStrategies: []DestinationApplyStrategy{
			{Mode: synccontract.ModeFullOverwrite, Strategy: ApplyStrategyReplace, Action: localWarehouseDestinationTransportAction},
			{Mode: synccontract.ModeFullAppend, Strategy: ApplyStrategyAppend, Action: localWarehouseDestinationTransportAction},
			{Mode: synccontract.ModeIncrementalUpsert, Strategy: ApplyStrategyMerge, Action: localWarehouseDestinationTransportAction},
			{Mode: synccontract.ModeIncrementalDedupe, Strategy: ApplyStrategyDedupe, Action: localWarehouseDestinationTransportAction},
			{Mode: synccontract.ModeIncrementalDedupeHistory, Strategy: ApplyStrategyDedupeHistory, Action: localWarehouseDestinationTransportAction},
		},
	}}
}

func (Warehouse) Metadata() Metadata {
	return Metadata{
		Name:            "warehouse",
		DisplayName:     "Local Warehouse",
		IntegrationType: "database",
		Description:     "Local Parquet warehouse destination queried by the embedded DuckDB engine.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Read: true, Write: true, Query: true},
	}
}

func (Warehouse) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.MkdirAll(warehousePath(cfg), 0o700)
}

// Catalog lists the tables materialized under the per-connection layout. One
// table name can belong to several connections, so it is reported once, and a
// read that does not say which connection it means is refused rather than
// answered from whichever file happened to be found first.
func (w Warehouse) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := w.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	tables, faults, err := warehouse.Tables(warehousePath(cfg))
	if err != nil {
		return Catalog{}, err
	}
	// A damaged ownership record hides only its own connection's tables. The
	// catalog still lists every healthy connection's, because one connection's
	// problem stays that connection's problem. It is reported only when there
	// is nothing left to list, so an incomplete catalog is never mistaken for
	// an empty warehouse.
	if len(tables) == 0 && len(faults) > 0 {
		return Catalog{}, warehouse.FaultsError("", faults)
	}
	owners := make(map[string][]string, len(tables))
	for _, table := range tables {
		// A root-level table's empty connection is kept as an entry of its
		// own: it has no owning connection, and naming one would attribute
		// rows to a connection that never produced them.
		owners[table.Name] = append(owners[table.Name], table.Connection)
	}
	streams := make([]Stream, 0, len(owners))
	for name, connections := range owners {
		named := make([]string, 0, len(connections))
		attributed := 0
		for _, connection := range connections {
			// A name held by a connection and by an unattributed root-level
			// file at once is held by both. Listing only the connection would
			// read as sole ownership and contradict the ambiguity error the
			// very next read of that name raises, which names both.
			if connection == "" {
				named = append(named, warehouse.UnattributedConnection)
				continue
			}
			attributed++
			named = append(named, connection)
		}
		description := "Warehouse table " + name
		if attributed > 0 {
			description += " (connection " + strings.Join(named, ", ") + ")"
		}
		streams = append(streams, Stream{Name: name, Description: description})
	}
	sort.Slice(streams, func(i, j int) bool { return streams[i].Name < streams[j].Name })
	return Catalog{Connector: w.Name(), Streams: streams}, nil
}

// Read resolves a table inside the per-connection layout. Config["connection"]
// scopes the read when more than one connection materializes the same name.
func (Warehouse) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	table, err := warehouse.FindTable(warehousePath(req.Config), req.Stream, req.Config.Config["connection"])
	if err != nil {
		return err
	}
	// The table is Parquet, and the warehouse package holds the single Parquet
	// implementation in the process. Reading it with a second one is how two
	// readers of the same file come to disagree.
	return warehouse.ReadTable(ctx, table.Path, func(row warehouse.Row) error {
		return emit(Record(row))
	})
}

// warehouseWriteTable is the single owner of which file a warehouse write
// resolves to and which names it accepts. A direct connector write carries no
// connection identity, so its rows stay at the root rather than entering a
// connection's directory, where they would be indistinguishable from that
// connection's own synced data. The table name is validated rather than folded
// into a safe filename, so a write and the read that follows it always name
// the same file. ValidateWrite exists only to predict Write, so it asks this
// rather than restating it: a predictor that carries its own copy of the rule
// stops predicting the moment either copy moves.
func warehouseWriteTable(req WriteRequest) (string, error) {
	table := req.Table
	if table == "" {
		table = req.Stream
	}
	return warehouse.PathComponent("table", table)
}

// ValidateWrite refuses a write this connector could never perform before the
// caller commits to it. A reverse plan is stored and approved ahead of the
// write it performs, and there is no way to edit a stored plan, so a refusal
// that only arrives at write time would leave an approved plan that can never
// run. It asks the same two questions Write does, in the same order, rather
// than carrying its own copy of either.
func (Warehouse) ValidateWrite(_ context.Context, req WriteRequest, _ []Record) error {
	if _, err := warehouseWriteTable(req); err != nil {
		return err
	}
	return warehouse.CheckLegacyTableFormat(warehousePath(req.Config))
}

func (Warehouse) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return WriteResult{}, err
	}
	dir := warehousePath(req.Config)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{}, fmt.Errorf("create warehouse directory: %w", err)
	}
	component, err := warehouseWriteTable(req)
	if err != nil {
		return WriteResult{}, err
	}
	// pm does not write into a warehouse it will not read. Reads of a
	// pre-Parquet warehouse are refused, so a write that reported success into
	// one would tell the caller the rows landed somewhere nothing can resolve.
	if err := warehouse.CheckLegacyTableFormat(dir); err != nil {
		return WriteResult{}, err
	}
	path := filepath.Join(dir, component+warehouse.TableFileExt)
	// A Parquet file cannot be appended to once closed, so an appending write
	// reads the rows already there and rewrites the table with both sets. The
	// table is derived and rewritten wholesale everywhere else in pm for the
	// same reason; this keeps the direct write surface consistent with it.
	rows := make([]warehouse.Row, 0, len(records))
	if !req.Overwrite {
		if _, err := os.Stat(path); err == nil {
			if err := warehouse.ReadTable(ctx, path, func(row warehouse.Row) error {
				rows = append(rows, row)
				return nil
			}); err != nil {
				return WriteResult{}, err
			}
		} else if !os.IsNotExist(err) {
			return WriteResult{}, fmt.Errorf("open warehouse table %s: %w", component, err)
		}
	}
	for _, record := range records {
		if err := ctx.Err(); err != nil {
			return WriteResult{}, err
		}
		rows = append(rows, warehouse.Row(record))
	}
	if err := warehouse.WriteTable(ctx, path, rows); err != nil {
		return WriteResult{}, err
	}
	return WriteResult{RecordsWritten: len(records)}, nil
}

type Outbox struct{}

func (Outbox) Name() string { return "outbox" }

func (Outbox) Metadata() Metadata {
	return Metadata{
		Name:            "outbox",
		DisplayName:     "Local Outbox",
		IntegrationType: "api",
		Description:     "Local JSONL destination that records reverse ETL writes and receipts.",
		Capabilities:    Capabilities{Check: true, Catalog: true, Write: true},
	}
}

func (Outbox) Check(ctx context.Context, cfg RuntimeConfig) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return os.MkdirAll(outboxPath(cfg), 0o700)
}

func (o Outbox) Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error) {
	if err := o.Check(ctx, cfg); err != nil {
		return Catalog{}, err
	}
	return Catalog{Connector: o.Name(), Streams: []Stream{{Name: "records", Description: "Reverse ETL outbox records."}}}, nil
}

func (Outbox) Read(ctx context.Context, req ReadRequest, emit func(Record) error) error {
	return ErrUnsupportedOperation
}

func (Outbox) Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error) {
	if err := ctx.Err(); err != nil {
		return WriteResult{}, err
	}
	dir := outboxPath(req.Config)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return WriteResult{}, fmt.Errorf("create outbox directory: %w", err)
	}
	name := req.Table
	if name == "" {
		name = "records"
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	enriched := make([]Record, 0, len(records))
	for _, record := range records {
		r := copyRecord(record)
		r["_outbox_action"] = req.Action
		r["_outbox_written_at"] = now
		enriched = append(enriched, r)
	}
	file, err := os.OpenFile(tablePath(dir, name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return WriteResult{}, fmt.Errorf("open outbox %s: %w", name, err)
	}
	defer func() {
		_ = file.Close()
	}()
	n, err := writeJSONL(ctx, file, enriched)
	if err != nil {
		return WriteResult{}, err
	}
	return WriteResult{RecordsWritten: n}, nil
}

func readJSONL(ctx context.Context, r io.Reader, emit func(Record) error) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return err
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var record Record
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return fmt.Errorf("decode jsonl record: %w", err)
		}
		if err := emit(record); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan jsonl: %w", err)
	}
	return nil
}

func readCSV(ctx context.Context, r io.Reader, emit func(Record) error) error {
	reader := csv.NewReader(r)
	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read csv header: %w", err)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read csv row: %w", err)
		}
		record := make(Record, len(header))
		for i, name := range header {
			if i < len(row) {
				record[name] = row[i]
			}
		}
		if err := emit(record); err != nil {
			return err
		}
	}
}

func writeJSONL(ctx context.Context, w io.Writer, records []Record) (int, error) {
	enc := json.NewEncoder(w)
	for i, record := range records {
		if err := ctx.Err(); err != nil {
			return i, err
		}
		if err := enc.Encode(record); err != nil {
			return i, fmt.Errorf("encode jsonl record: %w", err)
		}
	}
	return len(records), nil
}

func warehousePath(cfg RuntimeConfig) string {
	if cfg.Config["path"] != "" {
		return cfg.Config["path"]
	}
	return filepath.Join(cfg.ProjectDir, "warehouse")
}

func outboxPath(cfg RuntimeConfig) string {
	if cfg.Config["path"] != "" {
		return cfg.Config["path"]
	}
	return filepath.Join(cfg.ProjectDir, "outbox")
}

func tablePath(dir, table string) string {
	return filepath.Join(dir, safeName(table)+".jsonl")
}

func safeName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	var b strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		case r == '.' || r == '/' || r == ' ':
			b.WriteRune('_')
		}
	}
	if b.Len() == 0 {
		return "records"
	}
	return b.String()
}

func copyRecord(in Record) Record {
	out := make(Record, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
