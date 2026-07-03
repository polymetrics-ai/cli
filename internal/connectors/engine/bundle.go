// Package engine interprets declarative connector definition bundles (defs/)
// on top of the connsdk toolkit.
package engine

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
)

// namePattern is the shared connector/stream/action naming rule (design §A,
// design §F.3): dir name == metadata.name == registry key.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)

// Bundle is a fully loaded and structurally validated connector definition.
type Bundle struct {
	Name     string
	Metadata Metadata
	Spec     *Schema                  // compiled spec.json; SecretKeys() from x-secret
	RawSpec  json.RawMessage          // verbatim spec.json bytes (F5, REVIEW.md: Definition.Spec must serve this, not a lossy reconstruction); nil for a bundle that never loaded a real spec.json
	HTTP     HTTPBase                 // streams.json "base"; zero value when no streams.json
	Streams  []StreamSpec             // streams.json "streams"
	Writes   []WriteAction            // writes.json "actions"; nil when writes.json absent
	Schemas  map[string]*StreamSchema // stream name -> compiled schema + PK/cursor
	Surface  *APISurface              // api_surface.json
	Docs     string                   // docs.md
	Fixtures fs.FS                    // fixtures/ subtree; nil when absent
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
// behavior (e.g. "hook-covered; proven live by
// internal/connectors/paritytest/<name>"), never just assert the skip.
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
	Records        RecordsSpec           `json:"records"`
	Pagination     *PaginationSpec       `json:"pagination,omitempty"` // overrides base
	Incremental    *IncrementalSpec      `json:"incremental,omitempty"`
	ComputedFields map[string]string     `json:"computed_fields,omitempty"`
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
	Name         string          `json:"name"`
	Kind         string          `json:"kind"` // create|update|upsert|delete|custom
	Method       string          `json:"method"`
	Path         string          `json:"path"`
	PathFields   []string        `json:"path_fields,omitempty"`
	BodyType     string          `json:"body_type,omitempty"` // json (default) | form | none
	BodyFields   []string        `json:"body_fields,omitempty"`
	RecordSchema json.RawMessage `json:"record_schema"`
	Delete       *DeleteSpec     `json:"delete,omitempty"`
	Risk         string          `json:"risk"`
	Confirm      string          `json:"confirm,omitempty"` // "" | "destructive"
	Hook         string          `json:"hook,omitempty"`
}

// DeleteSpec describes idempotent-delete semantics for a delete write action.
type DeleteSpec struct {
	Idempotent      bool  `json:"idempotent,omitempty"`
	MissingOkStatus []int `json:"missing_ok_status,omitempty"`
}

// APISurface is the parsed api_surface.json (conformance input only).
type APISurface struct {
	API        string            `json:"api"`
	Docs       string            `json:"docs,omitempty"`
	ReviewedAt string            `json:"reviewed_at,omitempty"`
	Scope      string            `json:"scope,omitempty"`
	Endpoints  []SurfaceEndpoint `json:"endpoints"`
}

// SurfaceEndpoint is one api_surface.json endpoint entry.
type SurfaceEndpoint struct {
	Method    string            `json:"method,omitempty"`
	Path      string            `json:"path,omitempty"`
	CoveredBy *SurfaceCoverage  `json:"covered_by,omitempty"`
	Excluded  *SurfaceExclusion `json:"excluded,omitempty"`
}

// SurfaceCoverage names the stream or write action that covers an endpoint.
type SurfaceCoverage struct {
	Stream string `json:"stream,omitempty"`
	Write  string `json:"write,omitempty"`
}

// SurfaceExclusion names why an endpoint is intentionally out of scope.
type SurfaceExclusion struct {
	Category string `json:"category"`
	Reason   string `json:"reason,omitempty"`
}

// metaSchemas holds the compiled meta-schemas used to validate the bundle
// files themselves, lazily compiled once from the embedded schema/ dir.
var metaSchemas = struct {
	metadata, spec, streams, writes, apiSurface *Schema
	err                                         error
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
}

// requiredFiles lists the bundle files that must always exist relative to a
// bundle's directory, excepting streams.json (conditionally required).
var requiredFiles = []string{"metadata.json", "spec.json", "api_surface.json", "docs.md"}

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

	schemas, err := loadStreamSchemas(sub, dirName, streams)
	if err != nil {
		return Bundle{}, err
	}

	surface, err := loadAPISurface(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	docs, err := readFileString(sub, "docs.md")
	if err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: %w", dirName, err)
	}

	fixtures := loadFixtures(sub)

	return Bundle{
		Name:     dirName,
		Metadata: metadata,
		Spec:     spec,
		RawSpec:  rawSpec,
		HTTP:     httpBase,
		Streams:  streams,
		Writes:   writes,
		Schemas:  schemas,
		Surface:  surface,
		Docs:     docs,
		Fixtures: fixtures,
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
	return doc.Actions, nil
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
