// Package engine interprets declarative connector definition bundles (defs/)
// on top of the connsdk toolkit.
package engine

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// namePattern is the shared connector/stream/action naming rule (design §A,
// design §F.3): dir name == metadata.name == registry key.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var graphQLNamePattern = regexp.MustCompile(`^[_A-Za-z][_0-9A-Za-z]*$`)

// Bundle is a fully loaded and structurally validated connector definition.
type Bundle struct {
	Name          string
	Metadata      Metadata
	Spec          *Schema                  // compiled spec.json; SecretKeys() from x-secret
	RawSpec       json.RawMessage          // verbatim spec.json bytes (F5, REVIEW.md: Definition.Spec must serve this, not a lossy reconstruction); nil for a bundle that never loaded a real spec.json
	HTTP          HTTPBase                 // streams.json "base"; zero value when no streams.json
	Streams       []StreamSpec             // streams.json "streams"
	Writes        []WriteAction            // writes.json "actions"; nil when writes.json absent
	Operations    []OperationSpec          // operations.json "operations"; nil when operations.json absent
	RawOperations json.RawMessage          // verbatim operations.json bytes for validation/audit scanning
	Schemas       map[string]*StreamSchema // stream name -> compiled schema + PK/cursor
	Surface       *APISurface              // api_surface.json
	CLISurface    *CLISurface              // cli_surface.json
	RawCLISurface json.RawMessage          // verbatim cli_surface.json bytes; nil when absent
	Docs          string                   // docs.md
	Fixtures      fs.FS                    // fixtures/ subtree; nil when absent
}

// Metadata is the parsed metadata.json.
type Metadata struct {
	Name            string             `json:"name"`
	DisplayName     string             `json:"display_name"`
	Description     string             `json:"description"`
	IntegrationType string             `json:"integration_type"`
	DocsURL         string             `json:"docs_url,omitempty"`
	ReleaseStage    string             `json:"release_stage"`
	Capabilities    Capabilities       `json:"capabilities"`
	Batch           BatchSpec          `json:"batch,omitempty"`
	RateLimit       RateLimitSpec      `json:"rate_limit,omitempty"`
	Risk            RiskSpec           `json:"risk,omitempty"`
	Conformance     *ConformanceMarker `json:"conformance,omitempty"`
}

// ConformanceMarker is an OPTIONAL, explicit opt-out from
// conformance's dynamic (fixture-replay) checks, declarable at bundle level
// (Metadata.Conformance — e.g. a custom-auth-only connector whose hook
// conformance's synthetic config can never satisfy) or at stream level
// (StreamSpec.Conformance — e.g. a stream whose real reads dispatch entirely
// through a Tier-2 StreamHook that a declarative fixture replay cannot
// exercise). Reason is required whenever SkipDynamic is true
// (connectorgen validate's ruleConformanceSkipReason enforces this); it must
// name the authoritative substitute that actually proves the skipped
// behavior (e.g. "hook-covered; proven live by hook/native tests or archived
// pre-deletion parity evidence"), never just assert the skip.
type ConformanceMarker struct {
	SkipDynamic bool   `json:"skip_dynamic,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

// Capabilities mirrors metadata.json.capabilities.
type Capabilities struct {
	Check         bool `json:"check"`
	Read          bool `json:"read"`
	Write         bool `json:"write"`
	Query         bool `json:"query"`
	CDC           bool `json:"cdc"`
	DynamicSchema bool `json:"dynamic_schema"`
}

// BatchSpec mirrors metadata.json.batch.
type BatchSpec struct {
	ReadPageSize   int `json:"read_page_size,omitempty"`
	WriteBatchSize int `json:"write_batch_size,omitempty"`
}

// RiskSpec mirrors metadata.json.risk.
type RiskSpec struct {
	Read     string `json:"read,omitempty"`
	Write    string `json:"write,omitempty"`
	Approval string `json:"approval,omitempty"`
}

// HTTPBase is streams.json's "base" section: shared HTTP configuration for
// every stream in the bundle.
type HTTPBase struct {
	URL        string            `json:"url"`
	UserAgent  string            `json:"user_agent,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Auth       []AuthSpec        `json:"auth,omitempty"`
	Pagination *PaginationSpec   `json:"pagination,omitempty"`
	Check      *RequestSpec      `json:"check,omitempty"`
	ErrorMap   []ErrorRule       `json:"error_map,omitempty"`
	RateLimit  *RateLimitSpec    `json:"rate_limit,omitempty"`
}

// RequestSpec is a method+path(+query) request descriptor (used by "check").
//
// ENGINE DIALECT ADDITION (checkquery-ledger.md): Query mirrors
// StreamSpec.Query's existing string-or-object QueryParam dialect verbatim
// (per hardening-ledger.md's suggested follow-up shape) rather than a plain
// map[string]string, since 148 bundles under defs/ declare base.check.query
// with the same plain-string-template shape stream.Query already supports,
// and reusing the identical type gives check.query the same
// omit_when_absent/default escape hatches for free with no new dialect to
// learn. Before this field existed, base.check.query was either a load-time
// meta-schema rejection (post-hardening) or a silently-dropped no-op
// (pre-hardening) — see read.go's Check() for how it is now resolved+sent.
type RequestSpec struct {
	Method string                `json:"method"`
	Path   string                `json:"path"`
	Query  map[string]QueryParam `json:"query,omitempty"`
}

// AuthSpec describes one candidate authenticator, selected by "when" (first
// match wins).
type AuthSpec struct {
	Mode  string `json:"mode"` // none|bearer|basic|api_key_header|api_key_query|oauth2_client_credentials|custom
	Token string `json:"token,omitempty"`

	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	Header string `json:"header,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Param  string `json:"param,omitempty"`
	Value  string `json:"value,omitempty"`

	TokenURL     string `json:"token_url,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	Scopes       string `json:"scopes,omitempty"`
	// ExtraParams (S4 engine mini-wave item 4: auth0's audience form param,
	// box's box_subject_type/box_subject_id) are additional templated
	// key->value form params sent on every oauth2_client_credentials token
	// request, alongside grant_type/client_id/client_secret/scope. Each
	// value is resolved via ordinary Interpolate against the same Vars as
	// every other AuthSpec field — a hard error on an unresolved
	// config/secrets key, exactly like ClientID/ClientSecret (never silently
	// dropped). connsdk.OAuth2ClientCredentials already exposes an
	// ExtraParams url.Values field; this is the engine-side dialect that
	// populates it (connsdk itself needed no change).
	ExtraParams map[string]string `json:"extra_params,omitempty"`

	Hook string `json:"hook,omitempty"` // custom: hook name resolved via hooks registry
	When string `json:"when,omitempty"` // condition over config values
}

// PaginationSpec is the wave0 dialect (extends design §A examples to cover
// stripe's cursor+last_record_field shape and the next_url/allow_cross_host
// additions from the coordinator's Wave A corrections).
type PaginationSpec struct {
	Type string `json:"type"` // none|link_header|page_number|offset_limit|cursor|next_url

	SizeParam string `json:"size_param,omitempty"`
	PageParam string `json:"page_param,omitempty"`
	// StartPage is a pointer (S4 engine mini-wave item 1) so an EXPLICIT
	// "start_page": 0 (algolia/auth0/beamer/braze/clickup-api/concord/
	// customerly/dolibarr/harness/hubplanner-shaped genuinely 0-indexed
	// APIs) is distinguishable from an absent/omitted start_page — a plain
	// Go int cannot represent that distinction, since JSON-unmarshaling a
	// missing key produces the exact same zero value as an explicit 0.
	// nil means "not declared", defaulting to page 1 (newPaginator); a
	// non-nil pointer to 0 must be honored as the literal first page
	// number, never coerced.
	StartPage *int `json:"start_page,omitempty"`

	LimitParam  string `json:"limit_param,omitempty"`
	OffsetParam string `json:"offset_param,omitempty"`

	CursorParam     string `json:"cursor_param,omitempty"`
	TokenPath       string `json:"token_path,omitempty"`        // cursor: token from body
	LastRecordField string `json:"last_record_field,omitempty"` // cursor: token from last record (stripe)
	StopPath        string `json:"stop_path,omitempty"`         // cursor: falsy body value stops (stripe)

	NextURLPath string `json:"next_url_path,omitempty"` // next_url type

	PageSize int `json:"page_size,omitempty"`
	MaxPages int `json:"max_pages,omitempty"`

	// AllowCrossHost opts a next_url/Link-header follow out of the same-host
	// SSRF guard (THREAT-MODEL §3). Defaults to false; none of the wave0
	// goldens set it.
	AllowCrossHost bool `json:"allow_cross_host,omitempty"`
}

// ErrorRule is one error_map entry.
type ErrorRule struct {
	Status    int    `json:"status"`
	MatchBody string `json:"match_body,omitempty"`
	Class     string `json:"class,omitempty"`
	Hint      string `json:"hint,omitempty"`
}

// RateLimitSpec adds an inter-request wait inside the requester loop.
type RateLimitSpec struct {
	RequestsPerMinute int `json:"requests_per_minute,omitempty"`
}

// StreamSpec is one entry in streams.json's "streams" array.
type StreamSpec struct {
	Name           string                `json:"name"`
	Method         string                `json:"method,omitempty"` // default GET
	Path           string                `json:"path"`
	Query          map[string]QueryParam `json:"query,omitempty"`
	Body           map[string]any        `json:"body,omitempty"` // POST-body streams
	GraphQL        *GraphQLRequestSpec   `json:"graphql,omitempty"`
	Records        RecordsSpec           `json:"records"`
	Pagination     *PaginationSpec       `json:"pagination,omitempty"` // overrides base
	Incremental    *IncrementalSpec      `json:"incremental,omitempty"`
	ComputedFields map[string]string     `json:"computed_fields,omitempty"`
	ResponseFields map[string]string     `json:"response_fields,omitempty"`
	Projection     string                `json:"projection,omitempty"` // "schema" (default) | "passthrough"
	SchemaRef      string                `json:"schema"`
	Conformance    *ConformanceMarker    `json:"conformance,omitempty"`
	FanOut         *FanOutSpec           `json:"fan_out,omitempty"`
}

// FanOutSpec declares a sub-resource fan-out read (S4 engine mini-wave item
// 2: appfollow/bigmailer/breezy-hr/campayn/eventzilla/everhour/finnworlds/
// k6-cloud/metricool/cisco-meraki/configcat and 15+ quarantined/partial
// connectors whose real read is "list N parent ids, then repeat the WHOLE
// per-stream request/pagination/incremental sequence once per id, stamping
// the parent id onto every child record"). IDsFrom resolves the id list
// exactly ONCE per Read() call (before the first sub-sequence starts); Into
// decides how each resolved id is threaded into every request of its
// sub-sequence; StampField, when set, writes the id onto every emitted
// record of that sub-sequence (post-projection, alongside computed_fields).
type FanOutSpec struct {
	IDsFrom    FanOutIDsFrom `json:"ids_from"`
	Into       FanOutInto    `json:"into"`
	StampField string        `json:"stamp_field,omitempty"`
}

// FanOutIDsFrom is EXACTLY ONE of ConfigKey (a config value holding a
// comma-separated id list, e.g. appfollow's app_collection_ids) or Request (a
// preliminary GET, fully paginated to exhaustion using the stream's OWN base
// pagination spec, whose extracted records yield one id per record at
// IDField) — declaring both, or neither, is a read-time error (newFanOutIDs),
// mirroring cursor pagination's token_path/last_record_field mutual
// exclusivity (bundle.go's own PaginationSpec doc comment).
type FanOutIDsFrom struct {
	ConfigKey string            `json:"config_key,omitempty"`
	Request   *FanOutIDsRequest `json:"request,omitempty"`
}

// FanOutIDsRequest is the preliminary "list every parent id" request: Path is
// interpolated exactly like a stream's own Path (config/secrets templates,
// urlencoded-by-default path segments); RecordsPath is the dotted path
// (RecordsSpec.Path semantics) where the id records live in each page's
// body; IDField names the field on each extracted record holding the id
// value. Paginated with the stream's own effective pagination spec (base or
// stream-level override) — a fan-out id-listing request is not itself
// declared with its own pagination block; it reuses the child stream's.
type FanOutIDsRequest struct {
	Path        string `json:"path"`
	RecordsPath string `json:"records_path"`
	IDField     string `json:"id_field"`
}

// FanOutInto is EXACTLY ONE of QueryParam (the resolved id is added as a
// query parameter on every request of that id's sub-sequence — appfollow's
// apps_id=<id>) or PathVar (the resolved id becomes referenceable in the
// stream's own Path template as "{{ fanout.id }}" — declaring PathVar does
// NOT change what string is substituted for the literal name "fanout.id";
// PathVar exists so a future dialect could support multiple named fan-out
// vars, but today only "{{ fanout.id }}" is ever resolved).
type FanOutInto struct {
	QueryParam string `json:"query_param,omitempty"`
	PathVar    string `json:"path_var,omitempty"`
}

// QueryParam is one stream.Query entry (gap-loop cycle-1 item 3,
// REVIEW-B.md cross-cutting adjudication 2: the recurring
// "optional/config-driven query param not expressible" gap — vitally
// `status`, bitly `size`, calendly `count`/page_size, gmail's two filters,
// searxng wave0 F6 — met the >=3 recurrence threshold). Declared in
// streams.json either as a PLAIN STRING (today's exact dialect: `Template`
// is that string, `OmitWhenAbsent` false, `Default` empty — a template
// referencing an absent config/secrets key is ALWAYS a hard error, zero
// migration risk for every existing bundle) or as an OBJECT
// `{"template": "...", "omit_when_absent": true, "default": "..."}` — an
// explicit opt-in dialect, never a blanket absent-key-falsy change to query
// templating (which would silently convert a mistyped/missing REQUIRED key
// from a fail-loud error into a silently-unfiltered request, the F4
// fail-open class the engine deliberately rejects elsewhere).
//
// OmitWhenAbsent and Default are mutually usable but conceptually distinct:
// OmitWhenAbsent means "leave the param off the request entirely when its
// template resolves to an unresolved/absent key" (vitally's status filter);
// Default means "send this literal instead of hard-erroring, when the
// template's referenced key is absent" (calendly's page_size — closes the
// same gap class as a legacy in-code default). If both are set,
// OmitWhenAbsent takes priority conceptually but read.go's buildInitialQuery
// checks Default first only when OmitWhenAbsent is false — see that
// function's doc comment. Declaring both on the same entry is unusual
// authoring (contradictory intents) but not itself a validate-time error;
// bundle authors should pick one.
type QueryParam struct {
	Template       string `json:"template"`
	OmitWhenAbsent bool   `json:"omit_when_absent,omitempty"`
	Default        string `json:"default,omitempty"`
}

// UnmarshalJSON accepts EITHER a bare JSON string (sets Template, leaves
// OmitWhenAbsent/Default at their zero values — today's exact dialect) OR a
// JSON object matching QueryParam's fields verbatim. Any other JSON shape
// (number, array, bool, null) is a decode error.
func (q *QueryParam) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		q.Template = s
		q.OmitWhenAbsent = false
		q.Default = ""
		return nil
	}
	type alias QueryParam
	var obj alias
	if err := json.Unmarshal(data, &obj); err != nil {
		return fmt.Errorf("query param: expected a string or an object with a \"template\" field: %w", err)
	}
	*q = QueryParam(obj)
	return nil
}

// RecordsSpec describes how to extract records from a page body.
type RecordsSpec struct {
	Path         string      `json:"path"` // dotted path; "." = body root
	SingleObject bool        `json:"single_object,omitempty"`
	Filter       *FilterSpec `json:"filter,omitempty"`

	// KeyedObject (S4 engine mini-wave item 3: appfigures/alpha-vantage/
	// exchange-rates-shaped APIs) treats the JSON OBJECT found at Path as a
	// map of arbitrary-id -> record, exploding EACH VALUE into its own
	// record — e.g. {"111": {...}, "222": {...}} becomes 2 records — instead
	// of the ordinary RecordsAt behavior of treating a bare object as ONE
	// record (the whole object passed through verbatim). Mutually exclusive
	// with SingleObject in practice (SingleObject makes no sense once the
	// object's VALUES are the records, not the object itself), though the
	// loader does not enforce that in wave0 (mirrors FilterSpec's own
	// documented-but-unenforced mutual exclusivity).
	KeyedObject bool `json:"keyed_object,omitempty"`
	// KeyField, when set, stamps the source object's key (map key, e.g.
	// "111") onto that field of the exploded record BEFORE projection — so
	// it participates in schema projection/computed_fields exactly like any
	// other raw field. Ignored when KeyedObject is false.
	KeyField string `json:"key_field,omitempty"`
}

// FilterSpec is one of field_absent / field_equals (mutually exclusive by
// convention; the loader does not enforce this in wave0).
type FilterSpec struct {
	FieldAbsent string         `json:"field_absent,omitempty"`
	FieldEquals map[string]any `json:"field_equals,omitempty"`
}

// IncrementalSpec describes cursor-based incremental reads for a stream.
type IncrementalSpec struct {
	CursorField    string `json:"cursor_field"`
	RequestParam   string `json:"request_param,omitempty"`
	ParamFormat    string `json:"param_format,omitempty"` // rfc3339|unix_seconds|date|github_date_range
	StartConfigKey string `json:"start_config_key,omitempty"`
	ClientFiltered bool   `json:"client_filtered,omitempty"`
}

// WriteAction is one entry in writes.json's "actions" array.
type WriteAction struct {
	Name         string              `json:"name"`
	Kind         string              `json:"kind"` // create|update|upsert|delete|custom
	Method       string              `json:"method"`
	Path         string              `json:"path"`
	PathFields   []string            `json:"path_fields,omitempty"`
	BodyType     string              `json:"body_type,omitempty"` // json (default) | form | none | graphql
	BodyFields   []string            `json:"body_fields,omitempty"`
	GraphQL      *GraphQLRequestSpec `json:"graphql,omitempty"`
	RecordSchema json.RawMessage     `json:"record_schema"`
	Delete       *DeleteSpec         `json:"delete,omitempty"`
	Risk         string              `json:"risk"`
	Confirm      string              `json:"confirm,omitempty"` // "" | "destructive"
	Hook         string              `json:"hook,omitempty"`
}

// GraphQLRequestSpec describes a fixed GraphQL document whose variables are
// filled from declared templates. It is intentionally not a raw query escape
// hatch: Document is bundle metadata, never user input.
type GraphQLRequestSpec struct {
	Document      string         `json:"document"`
	OperationName string         `json:"operation_name,omitempty"`
	Variables     map[string]any `json:"variables,omitempty"`
}

// DeleteSpec describes idempotent-delete semantics for a delete write action.
type DeleteSpec struct {
	Idempotent      bool  `json:"idempotent,omitempty"`
	MissingOkStatus []int `json:"missing_ok_status,omitempty"`
}

// APISurface is the parsed api_surface.json (conformance input only).
type APISurface struct {
	API                    string            `json:"api"`
	Docs                   string            `json:"docs,omitempty"`
	ReviewedAt             string            `json:"reviewed_at,omitempty"`
	OperationLedgerVersion int               `json:"operation_ledger_version,omitempty"`
	Scope                  string            `json:"scope,omitempty"`
	Endpoints              []SurfaceEndpoint `json:"endpoints"`
}

// SurfaceEndpoint is one api_surface.json endpoint entry.
type SurfaceEndpoint struct {
	Method    string            `json:"method,omitempty"`
	Path      string            `json:"path,omitempty"`
	CoveredBy *SurfaceCoverage  `json:"covered_by,omitempty"`
	Excluded  *SurfaceExclusion `json:"excluded,omitempty"`
	Operation *SurfaceOperation `json:"operation,omitempty"`
}

// SurfaceCoverage names the executable connector surface that covers an endpoint.
type SurfaceCoverage struct {
	Stream      string   `json:"stream,omitempty"`
	Write       string   `json:"write,omitempty"`
	DirectRead  string   `json:"direct_read,omitempty"`
	DirectReads []string `json:"direct_reads,omitempty"`
}

// SurfaceExclusion names why an endpoint is intentionally out of scope.
type SurfaceExclusion struct {
	Category string `json:"category"`
	Reason   string `json:"reason,omitempty"`
}

// SurfaceOperation classifies a tracked endpoint that is not yet executable.
// Operation rows are metadata only and must remain blocked by default.
type SurfaceOperation struct {
	Model            string `json:"model"`
	Status           string `json:"status"`
	Risk             string `json:"risk"`
	BlockedByDefault bool   `json:"blocked_by_default"`
	Reason           string `json:"reason"`
	SourceURL        string `json:"source_url,omitempty"`
	Notes            string `json:"notes,omitempty"`
	DuplicateOf      string `json:"duplicate_of,omitempty"`
}

// OperationSpec is one reviewed, typed operation definition. The first phase
// loads and validates these definitions only; executors are added in later
// issue slices and every unknown kind remains rejected by the meta-schema.
type OperationSpec struct {
	ID              string                  `json:"id"`
	Kind            string                  `json:"kind"`
	Summary         string                  `json:"summary"`
	Description     string                  `json:"description,omitempty"`
	SourceURL       string                  `json:"source_url,omitempty"`
	Risk            string                  `json:"risk"`
	Approval        string                  `json:"approval"`
	OutputPolicy    string                  `json:"output_policy"`
	AuthScopes      []string                `json:"auth_scopes,omitempty"`
	MutationClass   string                  `json:"mutation_class,omitempty"`
	Destructive     bool                    `json:"destructive,omitempty"`
	SecretSensitive bool                    `json:"secret_sensitive,omitempty"`
	SensitivePolicy *SensitivePolicySpec    `json:"sensitive_policy,omitempty"`
	AuditEvent      string                  `json:"audit_event,omitempty"`
	REST            *RESTOperationSpec      `json:"rest,omitempty"`
	GraphQL         *GraphQLOperationSpec   `json:"graphql,omitempty"`
	XML             *XMLOperationSpec       `json:"xml,omitempty"`
	Binary          *BinaryOperationSpec    `json:"binary,omitempty"`
	File            *FileOperationSpec      `json:"file,omitempty"`
	LocalGit        *LocalGitOperationSpec  `json:"local_git,omitempty"`
	LocalFile       *LocalFileOperationSpec `json:"local_file,omitempty"`
	Browser         *BrowserOperationSpec   `json:"browser,omitempty"`
	Composite       *CompositeOperationSpec `json:"composite,omitempty"`
}

type RESTOperationSpec struct {
	Method string            `json:"method"`
	Path   string            `json:"path"`
	Query  map[string]string `json:"query,omitempty"`
}

// SensitivePolicySpec is the reverse-ETL sensitive/admin policy for an operation
// whose inputs or effects must never leak (secrets, variables, elevated-scope
// admin actions). It declares how secret values may be supplied (never inline
// CLI by default), which record fields must be redacted everywhere, the
// provider-specific transform that replaces a generic body template (e.g.
// GitHub's fetch-public-key + libsodium-encrypt flow), the preflight check that
// runs without reading secret values, and the approval mode that requires typed
// confirmation. The first sensitive/admin policy issue (#41) implements schema
// and validator support only; live secret writes remain blocked.
type SensitivePolicySpec struct {
	InputMode    string   `json:"input_mode,omitempty"`    // env | file | stdin | env_or_file | env_or_stdin (never "inline")
	RedactFields []string `json:"redact_fields,omitempty"` // record fields redacted in docs/previews/logs/errors
	Preflight    string   `json:"preflight,omitempty"`     // scope/availability check without reading secret values
	Transform    string   `json:"transform,omitempty"`     // none | github_secret_encryption | provider-specific
	ApprovalMode string   `json:"approval_mode,omitempty"` // typed_confirmation required for secret writes
}

type GraphQLOperationSpec struct {
	Document      string         `json:"document"`
	OperationName string         `json:"operation_name"`
	VariablesPath string         `json:"variables_path,omitempty"`
	Pagination    map[string]any `json:"pagination,omitempty"`
}

type XMLOperationSpec struct {
	EnvelopeTemplate string            `json:"envelope_template"`
	ResponsePath     string            `json:"response_path,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
}

type BinaryOperationSpec struct {
	Method          string `json:"method"`
	Path            string `json:"path"`
	MaxBytes        int    `json:"max_bytes,omitempty"`
	AllowOverwrite  bool   `json:"allow_overwrite,omitempty"`
	ExtractArchives bool   `json:"extract_archives,omitempty"`
}

type FileOperationSpec struct {
	Direction string `json:"direction"`
	Path      string `json:"path,omitempty"`
	MaxBytes  int    `json:"max_bytes,omitempty"`
}

type LocalGitOperationSpec struct {
	Action      string   `json:"action"`
	AllowedArgs []string `json:"allowed_args,omitempty"`
}

type LocalFileOperationSpec struct {
	Action   string `json:"action"`
	Path     string `json:"path,omitempty"`
	MaxBytes int    `json:"max_bytes,omitempty"`
}

type BrowserOperationSpec struct {
	Action string `json:"action"`
	URL    string `json:"url,omitempty"`
}

type CompositeOperationSpec struct {
	Steps []string `json:"steps"`
}

// CLISurface is the parsed cli_surface.json. It is docs/help metadata only:
// it maps provider-style command paths to existing streams, write actions,
// API-surface rows, or explicit unsupported/planned classifications.
type CLISurface struct {
	Tagline     string            `json:"tagline"`
	Usage       string            `json:"usage"`
	SourceCLI   *CLISourceCLI     `json:"source_cli,omitempty"`
	Groups      []CLICommandGroup `json:"groups,omitempty"`
	GlobalFlags []CLIFlag         `json:"global_flags,omitempty"`
	Commands    []CLICommand      `json:"commands"`
	HelpTopics  []CLIHelpTopic    `json:"help_topics,omitempty"`
}

// CLISourceCLI names the external provider CLI used as a parity reference.
type CLISourceCLI struct {
	Name      string `json:"name"`
	Docs      string `json:"docs,omitempty"`
	Reference string `json:"reference,omitempty"`
	Source    string `json:"source,omitempty"`
}

// CLICommandGroup is a rendered help grouping.
type CLICommandGroup struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Commands []string `json:"commands"`
}

// CLIFlag describes one command or global flag.
type CLIFlag struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Summary string   `json:"summary,omitempty"`
	Values  []string `json:"values,omitempty"`
	MapsTo  string   `json:"maps_to,omitempty"`
}

// CLICommand is one provider-inspired command path.
type CLICommand struct {
	Path          string                  `json:"path"`
	Summary       string                  `json:"summary"`
	Intent        string                  `json:"intent"`
	Availability  string                  `json:"availability"`
	Stream        string                  `json:"stream,omitempty"`
	Write         string                  `json:"write,omitempty"`
	SourceCLIPath string                  `json:"source_cli_path,omitempty"`
	SourceURL     string                  `json:"source_url,omitempty"`
	Flags         []CLIFlag               `json:"flags,omitempty"`
	Examples      []string                `json:"examples,omitempty"`
	APISurface    []CLISurfaceEndpointRef `json:"api_surface,omitempty"`
	OutputPolicy  string                  `json:"output_policy,omitempty"`
	Operation     string                  `json:"operation,omitempty"`
	Risk          string                  `json:"risk,omitempty"`
	Approval      string                  `json:"approval,omitempty"`
	Notes         string                  `json:"notes,omitempty"`
}

// CLISurfaceEndpointRef points from a command to a tracked api_surface row.
type CLISurfaceEndpointRef struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// CLIHelpTopic is one rendered help topic.
type CLIHelpTopic struct {
	Name    string `json:"name"`
	Summary string `json:"summary"`
}

// metaSchemas holds the compiled meta-schemas used to validate the bundle
// files themselves, lazily compiled once from the embedded schema/ dir.
var metaSchemas = struct {
	metadata, spec, streams, writes, apiSurface, operations, cliSurface *Schema
	err                                                                 error
}{}

func init() {
	compileMeta := func(raw string) *Schema {
		if metaSchemas.err != nil {
			return nil
		}
		sch, err := CompileSchema(json.RawMessage(raw))
		if err != nil {
			metaSchemas.err = err
			return nil
		}
		return sch
	}
	metaSchemas.metadata = compileMeta(metadataSchemaJSON)
	metaSchemas.spec = compileMeta(specSchemaJSON)
	metaSchemas.streams = compileMeta(streamsSchemaJSON)
	metaSchemas.writes = compileMeta(writesSchemaJSON)
	metaSchemas.apiSurface = compileMeta(apiSurfaceSchemaJSON)
	metaSchemas.operations = compileMeta(operationsSchemaJSON)
	metaSchemas.cliSurface = compileMeta(cliSurfaceSchemaJSON)
}

// requiredFiles lists the bundle files that must always exist relative to a
// bundle's directory, excepting streams.json (conditionally required).
//
// api_surface.json is intentionally not required here. It is a
// conformance/authoring artifact, not runtime input for check/read/write/catalog
// execution, and production defs.FS excludes it to keep cmd/pm small. When the
// file is present, loadAPISurface still parses and validates it so disk-backed
// validation keeps the full coverage gate.
var requiredFiles = []string{"metadata.json", "spec.json", "docs.md"}

// LoadAllError is the structured error LoadAll returns whenever one or more
// (but not necessarily all) bundle directories under fsys failed to load.
// Failures preserves discovery order (the same sorted directory-name order
// LoadAll iterates) so callers that want per-bundle granularity (e.g.
// conformance's TestConformance, which reports one failing subtest per
// bundle name rather than a single opaque batch failure) can do so via
// errors.As instead of parsing Error()'s message.
type LoadAllError struct {
	Failures []BundleLoadFailure
}

// BundleLoadFailure names one bundle directory that failed to load and the
// error Load returned for it.
type BundleLoadFailure struct {
	Name string
	Err  error
}

// GetFailures returns e.Failures, tolerating a nil *LoadAllError (the shape
// errors.As leaves its target in when the wrapped error chain contained no
// *LoadAllError at all, e.g. when LoadAll returned a nil error) so callers
// that iterate "whatever failed, if anything" never need their own nil
// check first.
func (e *LoadAllError) GetFailures() []BundleLoadFailure {
	if e == nil {
		return nil
	}
	return e.Failures
}

func (e *LoadAllError) Error() string {
	names := make([]string, 0, len(e.Failures))
	msgs := make([]string, 0, len(e.Failures))
	for _, f := range e.Failures {
		names = append(names, f.Name)
		msgs = append(msgs, f.Err.Error())
	}
	return fmt.Sprintf("load all bundles: %d bundle(s) failed to load (%s): %s",
		len(e.Failures), strings.Join(names, ", "), strings.Join(msgs, "; "))
}

// LoadAll loads and validates every bundle directory found at the root of
// fsys. Non-directory root entries are skipped. An empty tree is not an
// error (returns zero bundles).
//
// ENGINE HARDENING (hardening-ledger.md): a single bundle that fails to load
// no longer hides every OTHER bundle in fsys. LoadAll attempts every bundle
// directory (never aborts partway through) and, if one or more failed,
// returns the bundles that DID load cleanly alongside a non-nil *LoadAllError
// naming every failure. This mirrors cmd/connectorgen validate's own
// long-standing per-bundle isolation (validateBundleDir already turns one
// bundle's engine.Load error into an isolated Finding rather than aborting
// the whole validate run) — with ~400 independently-authored bundles under
// defs/, fail-fast-on-first-error made fleet-wide discovery (the same path
// production bundle-registry construction and every defs.FS-wide test in
// this repo uses) an all-or-nothing proposition, which is exactly the
// failure mode this change closes. Callers that must treat any failure as
// fatal still get a non-nil error to check (via plain err != nil, or
// errors.As(&LoadAllError{}) for the per-bundle detail); callers that want
// the currently-loadable subset (this package's and conformance's
// fleet-wide tests) can proceed with the returned bundles regardless.
func LoadAll(fsys fs.FS) ([]Bundle, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("load all bundles: read root: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	bundles := make([]Bundle, 0, len(names))
	var loadErr LoadAllError
	for _, name := range names {
		b, err := Load(fsys, name)
		if err != nil {
			loadErr.Failures = append(loadErr.Failures, BundleLoadFailure{Name: name, Err: err})
			continue
		}
		bundles = append(bundles, b)
	}
	if len(loadErr.Failures) > 0 {
		return bundles, &loadErr
	}
	return bundles, nil
}

// Load loads and structurally validates a single bundle directory named
// dirName at the root of fsys.
func Load(fsys fs.FS, dirName string) (Bundle, error) {
	if metaSchemas.err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: meta-schemas failed to compile: %w", dirName, metaSchemas.err)
	}

	sub, err := fs.Sub(fsys, dirName)
	if err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: %w", dirName, err)
	}

	for _, f := range requiredFiles {
		if _, err := fs.Stat(sub, f); err != nil {
			return Bundle{}, fmt.Errorf("load bundle %s: missing required file %s", dirName, f)
		}
	}

	metadata, err := loadMetadata(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	spec, rawSpec, err := loadSpec(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	httpBase, streams, err := loadStreams(sub, dirName, metadata)
	if err != nil {
		return Bundle{}, err
	}

	writes, err := loadWrites(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	operations, rawOperations, err := loadOperations(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	schemas, err := loadStreamSchemas(sub, dirName, streams)
	if err != nil {
		return Bundle{}, err
	}

	surface, err := loadAPISurface(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	cliSurface, rawCLISurface, err := loadCLISurface(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	docs, err := readFileString(sub, "docs.md")
	if err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: %w", dirName, err)
	}

	fixtures := loadFixtures(sub)

	return Bundle{
		Name:          dirName,
		Metadata:      metadata,
		Spec:          spec,
		RawSpec:       rawSpec,
		HTTP:          httpBase,
		Streams:       streams,
		Writes:        writes,
		Operations:    operations,
		RawOperations: rawOperations,
		Schemas:       schemas,
		Surface:       surface,
		CLISurface:    cliSurface,
		RawCLISurface: rawCLISurface,
		Docs:          docs,
		Fixtures:      fixtures,
	}, nil
}

func loadMetadata(sub fs.FS, dirName string) (Metadata, error) {
	raw, err := readFile(sub, "metadata.json")
	if err != nil {
		return Metadata{}, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.metadata.Validate(mustDecodeAny(raw)); err != nil {
		return Metadata{}, fmt.Errorf("load bundle %s: metadata.json: %w", dirName, err)
	}

	var m Metadata
	if err := strictDecode(raw, &m); err != nil {
		return Metadata{}, fmt.Errorf("load bundle %s: metadata.json: %w", dirName, err)
	}

	if !namePattern.MatchString(m.Name) {
		return Metadata{}, fmt.Errorf("load bundle %s: metadata.json name %q does not match %s", dirName, m.Name, namePattern.String())
	}
	if m.Name != dirName {
		return Metadata{}, fmt.Errorf("load bundle %s: directory name %q does not match metadata.json name %q", dirName, dirName, m.Name)
	}

	return m, nil
}

// loadSpec returns both the compiled *Schema (used for runtime interpolation
// checks, SecretKeys()/Properties()/RequiredKeys()) and the VERBATIM raw
// spec.json bytes it already read (F5, REVIEW.md: the loader previously
// discarded these after compiling, forcing Definition.Spec to lossily
// reconstruct the config surface from the compiled Schema alone — dropping
// types/enums/defaults/required/descriptions).
func loadSpec(sub fs.FS, dirName string) (*Schema, json.RawMessage, error) {
	raw, err := readFile(sub, "spec.json")
	if err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.spec.Validate(mustDecodeAny(raw)); err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: spec.json: %w", dirName, err)
	}
	sch, err := CompileSchema(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: spec.json: %w", dirName, err)
	}
	return sch, json.RawMessage(raw), nil
}

func loadStreams(sub fs.FS, dirName string, metadata Metadata) (HTTPBase, []StreamSpec, error) {
	exists := fileExists(sub, "streams.json")
	if !exists {
		if metadata.Capabilities.DynamicSchema {
			return HTTPBase{}, nil, nil
		}
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: missing required file streams.json (required unless capabilities.dynamic_schema is true)", dirName)
	}

	raw, err := readFile(sub, "streams.json")
	if err != nil {
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.streams.Validate(mustDecodeAny(raw)); err != nil {
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: streams.json: %w", dirName, err)
	}

	var doc struct {
		Base    HTTPBase     `json:"base"`
		Streams []StreamSpec `json:"streams"`
	}
	if err := strictDecode(raw, &doc); err != nil {
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: streams.json: %w", dirName, err)
	}
	if err := validateStreamGraphQL(doc.Streams); err != nil {
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: streams.json: %w", dirName, err)
	}
	return doc.Base, doc.Streams, nil
}

func loadWrites(sub fs.FS, dirName string) ([]WriteAction, error) {
	if !fileExists(sub, "writes.json") {
		return nil, nil
	}
	raw, err := readFile(sub, "writes.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.writes.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: writes.json: %w", dirName, err)
	}
	var doc struct {
		Actions []WriteAction `json:"actions"`
	}
	if err := strictDecode(raw, &doc); err != nil {
		return nil, fmt.Errorf("load bundle %s: writes.json: %w", dirName, err)
	}
	if err := validateWriteGraphQL(doc.Actions); err != nil {
		return nil, fmt.Errorf("load bundle %s: writes.json: %w", dirName, err)
	}
	return doc.Actions, nil
}

func validateStreamGraphQL(streams []StreamSpec) error {
	for i, stream := range streams {
		if stream.GraphQL == nil {
			continue
		}
		if len(stream.Body) > 0 {
			return fmt.Errorf("stream %d (%q) cannot declare both body and graphql", i, stream.Name)
		}
		if method := strings.ToUpper(methodOrDefault(stream.Method)); method != "POST" {
			return fmt.Errorf("stream %d (%q) graphql stream method must be POST, got %s", i, stream.Name, method)
		}
		if err := validateGraphQLSpec(stream.GraphQL, "query"); err != nil {
			return fmt.Errorf("stream %d (%q): %w", i, stream.Name, err)
		}
	}
	return nil
}

func validateWriteGraphQL(actions []WriteAction) error {
	for i, action := range actions {
		bodyType := bodyTypeOf(action)
		if action.GraphQL != nil && bodyType != "graphql" {
			return fmt.Errorf("action %d (%q) declares graphql but body_type is %q", i, action.Name, bodyType)
		}
		if bodyType != "graphql" {
			continue
		}
		if action.GraphQL == nil {
			return fmt.Errorf("action %d (%q) body_type graphql requires graphql", i, action.Name)
		}
		if len(action.BodyFields) > 0 {
			return fmt.Errorf("action %d (%q) body_type graphql cannot declare body_fields", i, action.Name)
		}
		if method := strings.ToUpper(methodOrDefault(action.Method)); method != "POST" {
			return fmt.Errorf("action %d (%q) graphql action method must be POST, got %s", i, action.Name, method)
		}
		if err := validateGraphQLSpec(action.GraphQL, "mutation"); err != nil {
			return fmt.Errorf("action %d (%q): %w", i, action.Name, err)
		}
	}
	return nil
}

func validateGraphQLSpec(spec *GraphQLRequestSpec, operationKind string) error {
	if spec == nil {
		return fmt.Errorf("graphql is required")
	}
	doc := strings.TrimSpace(spec.Document)
	if doc == "" {
		return fmt.Errorf("graphql.document is required")
	}
	if strings.Contains(doc, "{{") || strings.Contains(doc, "}}") {
		return fmt.Errorf("graphql.document must be fixed bundle metadata, not a template")
	}
	if operationKind != "" && !graphQLDocumentStartsWith(doc, operationKind) {
		return fmt.Errorf("graphql.document must start with %s", operationKind)
	}
	opName := strings.TrimSpace(spec.OperationName)
	if opName == "" {
		return fmt.Errorf("graphql.operation_name is required")
	}
	if !graphQLNamePattern.MatchString(opName) {
		return fmt.Errorf("graphql.operation_name %q is not a valid GraphQL name", opName)
	}
	for name := range spec.Variables {
		if !graphQLNamePattern.MatchString(name) {
			return fmt.Errorf("graphql variable %q is not a valid GraphQL name", name)
		}
	}
	if err := validateGraphQLVariables(spec.Variables); err != nil {
		return err
	}
	return nil
}

func graphQLDocumentStartsWith(doc, kind string) bool {
	if !strings.HasPrefix(doc, kind) {
		return false
	}
	if len(doc) == len(kind) {
		return true
	}
	switch doc[len(kind)] {
	case ' ', '\t', '\n', '\r', '(', '{':
		return true
	default:
		return false
	}
}

func validateGraphQLVariables(vars map[string]any) error {
	for name, value := range vars {
		if err := validateGraphQLVariableValue(name, value); err != nil {
			return err
		}
	}
	return nil
}

func validateGraphQLDefaultForType(def, typ string) error {
	switch typ {
	case "integer":
		if _, err := strconv.ParseInt(def, 10, 64); err != nil {
			return fmt.Errorf("must be a valid integer, got %q", def)
		}
	case "number":
		if _, err := strconv.ParseFloat(def, 64); err != nil {
			return fmt.Errorf("must be a valid number, got %q", def)
		}
	case "boolean":
		if _, err := strconv.ParseBool(def); err != nil {
			return fmt.Errorf("must be a valid boolean, got %q", def)
		}
	}
	return nil
}

func validateGraphQLVariableValue(name string, value any) error {
	obj, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	if _, isTemplate := obj["template"]; isTemplate {
		if _, ok := obj["template"].(string); !ok {
			return fmt.Errorf("graphql variable %q template must be a string", name)
		}
		for key := range obj {
			if key != "template" && key != "type" && key != "omit_when_empty" && key != "default" {
				return fmt.Errorf("graphql variable %q template object has unsupported key %q", name, key)
			}
		}
		if def, ok := obj["default"]; ok {
			defStr, ok := def.(string)
			if !ok {
				return fmt.Errorf("graphql variable %q default must be a string", name)
			}
			if typ, _ := obj["type"].(string); typ != "" && typ != "string" {
				if err := validateGraphQLDefaultForType(defStr, typ); err != nil {
					return fmt.Errorf("graphql variable %q default %v", name, err)
				}
			}
		}
		if omit, ok := obj["omit_when_empty"]; ok {
			if _, ok := omit.(bool); !ok {
				return fmt.Errorf("graphql variable %q omit_when_empty must be a boolean", name)
			}
		}
		if typ, ok := obj["type"].(string); ok {
			switch typ {
			case "", "string", "integer", "number", "boolean":
			default:
				return fmt.Errorf("graphql variable %q has unsupported type %q", name, typ)
			}
		}
		return nil
	}
	for childName, childValue := range obj {
		if err := validateGraphQLVariableValue(childName, childValue); err != nil {
			return err
		}
	}
	return nil
}

func loadOperations(sub fs.FS, dirName string) ([]OperationSpec, json.RawMessage, error) {
	if !fileExists(sub, "operations.json") {
		return nil, nil, nil
	}
	raw, err := readFile(sub, "operations.json")
	if err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.operations.Validate(mustDecodeAny(raw)); err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: operations.json: %w", dirName, err)
	}
	var doc struct {
		Operations []OperationSpec `json:"operations"`
	}
	if err := strictDecode(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: operations.json: %w", dirName, err)
	}
	if err := validateOperations(doc.Operations); err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: operations.json: %w", dirName, err)
	}
	return doc.Operations, raw, nil
}

func validateOperations(ops []OperationSpec) error {
	seen := map[string]bool{}
	for i, op := range ops {
		if seen[op.ID] {
			return fmt.Errorf("operation %d has duplicate operation id %q", i, op.ID)
		}
		seen[op.ID] = true

		block, count := operationExecutionBlock(op)
		if count != 1 {
			return fmt.Errorf("operation %d (%q) must declare exactly one execution block, got %d", i, op.ID, count)
		}
		expected := expectedOperationBlock(op.Kind)
		if expected == "" {
			return fmt.Errorf("operation %d (%q) has unsupported kind %q", i, op.ID, op.Kind)
		}
		if block != expected {
			return fmt.Errorf("operation %d (%q) kind %q must declare %s block, got %s", i, op.ID, op.Kind, expected, block)
		}
		if err := validateOperationSemantics(i, op); err != nil {
			return err
		}
	}
	return nil
}

func operationExecutionBlock(op OperationSpec) (string, int) {
	var block string
	count := 0
	add := func(name string, present bool) {
		if !present {
			return
		}
		block = name
		count++
	}
	add("rest", op.REST != nil)
	add("graphql", op.GraphQL != nil)
	add("xml", op.XML != nil)
	add("binary", op.Binary != nil)
	add("file", op.File != nil)
	add("local_git", op.LocalGit != nil)
	add("local_file", op.LocalFile != nil)
	add("browser", op.Browser != nil)
	add("composite", op.Composite != nil)
	return block, count
}

func expectedOperationBlock(kind string) string {
	switch kind {
	case "rest_read", "rest_write":
		return "rest"
	case "graphql_query", "graphql_mutation":
		return "graphql"
	case "xml_export", "xml_import":
		return "xml"
	case "binary_download":
		return "binary"
	case "file_upload":
		return "file"
	case "local_git":
		return "local_git"
	case "local_file":
		return "local_file"
	case "browser_open":
		return "browser"
	case "stream_etl", "composite":
		return "composite"
	default:
		return ""
	}
}

func validateOperationSemantics(i int, op OperationSpec) error {
	switch op.Kind {
	case "rest_read":
		if method := strings.ToUpper(strings.TrimSpace(op.REST.Method)); method != "GET" {
			return fmt.Errorf("operation %d (%q) rest_read method must be GET, got %s", i, op.ID, method)
		}
	case "rest_write":
		method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
		if method == "GET" || method == "HEAD" || method == "" {
			return fmt.Errorf("operation %d (%q) rest_write method must be mutating, got %s", i, op.ID, method)
		}
		if strings.TrimSpace(op.MutationClass) == "" || op.MutationClass == "none" {
			return fmt.Errorf("operation %d (%q) rest_write must declare mutation_class", i, op.ID)
		}
		if strings.TrimSpace(op.Approval) == "" || op.Approval == "none" {
			return fmt.Errorf("operation %d (%q) rest_write must declare approval requirements", i, op.ID)
		}
	case "graphql_mutation", "xml_import":
		if strings.TrimSpace(op.MutationClass) == "" || op.MutationClass == "none" {
			return fmt.Errorf("operation %d (%q) %s must declare mutation_class", i, op.ID, op.Kind)
		}
		if strings.TrimSpace(op.Approval) == "" || op.Approval == "none" {
			return fmt.Errorf("operation %d (%q) %s must declare approval requirements", i, op.ID, op.Kind)
		}
	case "binary_download":
		if method := strings.ToUpper(strings.TrimSpace(op.Binary.Method)); method != "GET" {
			return fmt.Errorf("operation %d (%q) binary_download method must be GET, got %s", i, op.ID, method)
		}
		if op.Binary.MaxBytes <= 0 {
			return fmt.Errorf("operation %d (%q) binary_download must declare positive max_bytes", i, op.ID)
		}
	case "file_upload":
		if op.File.Direction != "upload" {
			return fmt.Errorf("operation %d (%q) file_upload direction must be upload, got %s", i, op.ID, op.File.Direction)
		}
		if op.File.MaxBytes <= 0 {
			return fmt.Errorf("operation %d (%q) file_upload must declare positive max_bytes", i, op.ID)
		}
		if strings.TrimSpace(op.Approval) == "" || op.Approval == "none" {
			return fmt.Errorf("operation %d (%q) file_upload must declare approval requirements", i, op.ID)
		}
	case "local_file":
		if op.LocalFile.Action == "write" || op.LocalFile.Action == "mkdir" {
			if strings.TrimSpace(op.Approval) == "" || op.Approval == "none" {
				return fmt.Errorf("operation %d (%q) local_file mutation must declare approval requirements", i, op.ID)
			}
		}
		if op.LocalFile.Action == "write" && op.LocalFile.MaxBytes <= 0 {
			return fmt.Errorf("operation %d (%q) local_file write must declare positive max_bytes", i, op.ID)
		}
	}
	if err := validateSensitivePolicy(i, op); err != nil {
		return err
	}
	return nil
}

// validateSensitivePolicy enforces the sensitive/admin reverse-ETL policy model
// (#41). An operation that is secret_sensitive or has mutation_class "secret"
// must declare a sensitive_policy with: a non-inline input_mode, at least one
// redact_fields entry, and approval_mode "typed_confirmation". The transform,
// when set, must be a known value. Live secret writes remain blocked in this
// issue; this is schema + validator support only.
func validateSensitivePolicy(i int, op OperationSpec) error {
	isSecret := op.SecretSensitive || strings.EqualFold(op.MutationClass, "secret")
	if !isSecret {
		// Non-secret operations may still declare a policy (e.g. admin actions
		// that redact fields) but are not forced to.
		if op.SensitivePolicy == nil {
			return nil
		}
	} else if op.SensitivePolicy == nil {
		return fmt.Errorf("operation %d (%q) is secret_sensitive but declares no sensitive_policy (input_mode, redact_fields, approval_mode)", i, op.ID)
	}
	p := op.SensitivePolicy
	switch strings.ToLower(strings.TrimSpace(p.InputMode)) {
	case "", "inline":
		if isSecret {
			return fmt.Errorf("operation %d (%q) sensitive_policy input_mode must not be inline; secret values must come from env/file/stdin", i, op.ID)
		}
	case "env", "file", "stdin", "env_or_file", "env_or_stdin":
		// allowed
	default:
		return fmt.Errorf("operation %d (%q) sensitive_policy input_mode %q is not a known value", i, op.ID, p.InputMode)
	}
	if isSecret && len(p.RedactFields) == 0 {
		return fmt.Errorf("operation %d (%q) sensitive_policy must declare at least one redact_fields entry", i, op.ID)
	}
	switch strings.ToLower(strings.TrimSpace(p.Transform)) {
	case "", "none", "github_secret_encryption":
		// allowed
	default:
		return fmt.Errorf("operation %d (%q) sensitive_policy transform %q is not a known value", i, op.ID, p.Transform)
	}
	if isSecret && !strings.EqualFold(p.ApprovalMode, "typed_confirmation") {
		return fmt.Errorf("operation %d (%q) sensitive_policy approval_mode must be typed_confirmation for secret writes", i, op.ID)
	}
	return nil
}

func loadStreamSchemas(sub fs.FS, dirName string, streams []StreamSpec) (map[string]*StreamSchema, error) {
	if len(streams) == 0 {
		return map[string]*StreamSchema{}, nil
	}
	out := make(map[string]*StreamSchema, len(streams))
	for _, s := range streams {
		raw, err := readFile(sub, s.SchemaRef)
		if err != nil {
			return nil, fmt.Errorf("load bundle %s: stream %s: schema %s: %w", dirName, s.Name, s.SchemaRef, err)
		}
		sch, err := CompileSchema(raw)
		if err != nil {
			return nil, fmt.Errorf("load bundle %s: stream %s: schema %s: %w", dirName, s.Name, s.SchemaRef, err)
		}
		out[s.Name] = &StreamSchema{
			Schema:      sch,
			PrimaryKey:  sch.PrimaryKeys(),
			CursorField: sch.CursorFieldName(),
		}
	}
	return out, nil
}

func loadAPISurface(sub fs.FS, dirName string) (*APISurface, error) {
	if !fileExists(sub, "api_surface.json") {
		return nil, nil
	}
	raw, err := readFile(sub, "api_surface.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.apiSurface.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: api_surface.json: %w", dirName, err)
	}
	var surface APISurface
	if err := strictDecode(raw, &surface); err != nil {
		return nil, fmt.Errorf("load bundle %s: api_surface.json: %w", dirName, err)
	}
	return &surface, nil
}

func loadCLISurface(sub fs.FS, dirName string) (*CLISurface, json.RawMessage, error) {
	if !fileExists(sub, "cli_surface.json") {
		return nil, nil, nil
	}
	raw, err := readFile(sub, "cli_surface.json")
	if err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.cliSurface.Validate(mustDecodeAny(raw)); err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: cli_surface.json: %w", dirName, err)
	}
	var surface CLISurface
	if err := strictDecode(raw, &surface); err != nil {
		return nil, nil, fmt.Errorf("load bundle %s: cli_surface.json: %w", dirName, err)
	}
	return &surface, json.RawMessage(raw), nil
}

func loadFixtures(sub fs.FS) fs.FS {
	if !dirExists(sub, "fixtures") {
		return nil
	}
	fixturesFS, err := fs.Sub(sub, "fixtures")
	if err != nil {
		return nil
	}
	return fixturesFS
}

func readFile(fsys fs.FS, name string) ([]byte, error) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	return data, nil
}

func readFileString(fsys fs.FS, name string) (string, error) {
	data, err := readFile(fsys, name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func fileExists(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && !info.IsDir()
}

func dirExists(fsys fs.FS, name string) bool {
	info, err := fs.Stat(fsys, name)
	return err == nil && info.IsDir()
}

// strictDecode decodes raw into dst via encoding/json with
// DisallowUnknownFields, independent of and in addition to the meta-schema
// Validate() pass every caller already runs first.
//
// ENGINE HARDENING (hardening-ledger.md): the meta-schema pass alone is not
// sufficient defense-in-depth for this specific mistake class — a bundle
// author (or a future edit to the meta-schema files themselves) could
// silently reopen the hole a bare {"type":"object"} sub-schema left open
// (internal/connectors/defs/rentcast's now-repaired "base.query", still
// exactly reproduced today by ~150 bundles' "base.check.query": RequestSpec
// only has Method/Path, so that JSON silently did nothing at runtime while
// passing every gate). DisallowUnknownFields rejects any key not matched by
// a STRUCT field on dst (or on any nested struct/pointer-to-struct it
// decodes into); fields typed as a map (HTTPBase.Headers, StreamSpec.Query/
// Body/ComputedFields, FilterSpec.FieldEquals, RecordSchema, ...) remain
// deliberately open, since those are genuinely caller-defined free-form
// key sets, not a fixed dialect surface.
func strictDecode(raw []byte, dst any) error {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("%w", err)
	}
	return nil
}

// mustDecodeAny decodes raw JSON into a generic any for meta-schema
// validation. Callers only pass already-well-formed-enough bytes (read from
// disk/embed); a decode failure here is folded into the returned error by
// the caller's json.Unmarshal step that follows, so this helper degrades to
// nil on error rather than panicking.
func mustDecodeAny(raw []byte) any {
	var v any
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.UseNumber()
	if err := dec.Decode(&v); err != nil {
		return nil
	}
	return v
}
