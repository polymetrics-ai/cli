// Package engine interprets declarative connector definition bundles (defs/)
// on top of the connsdk toolkit.
package engine

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/database"
)

// namePattern is the shared connector/stream/action naming rule (design §A,
// design §F.3): dir name == metadata.name == registry key.
var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var graphQLNamePattern = regexp.MustCompile(`^[_A-Za-z][_0-9A-Za-z]*$`)
var httpHeaderNamePattern = regexp.MustCompile("^[!#$%&'*+.^_`|~0-9A-Za-z-]+$")

// Bundle is a fully loaded and structurally validated connector definition.
type Bundle struct {
	Name                 string
	Metadata             Metadata
	Changefeed           *connectors.ChangefeedDescriptor       // changefeed.json; nil when a connector has not been surveyed
	PollingWatermark     *connectors.PollingWatermarkDescriptor // polling_watermark.json; nil until a native database declaration exists
	SyncTransport        *connectors.SyncTransportDescriptor    // sync_transport.json; nil when no closed source/destination role is declared
	Database             *database.Definition                   // database.json; nil for non-database or unmigrated bundles
	Spec                 *Schema                                // compiled spec.json; SecretKeys() from x-secret
	RawSpec              json.RawMessage                        // verbatim spec.json bytes (F5, REVIEW.md: Definition.Spec must serve this, not a lossy reconstruction); nil for a bundle that never loaded a real spec.json
	HTTP                 HTTPBase                               // streams.json "base"; zero value when no streams.json
	Streams              []StreamSpec                           // streams.json "streams"
	Writes               []WriteAction                          // writes.json "actions"; nil when writes.json absent
	Operations           []OperationSpec                        // operations.json "operations"; nil when operations.json absent
	RawOperations        json.RawMessage                        // verbatim operations.json bytes for validation/audit scanning
	Schemas              map[string]*StreamSchema               // stream name -> compiled schema + PK/cursor
	directWriteEndpoints []directWriteEndpoint                  // runtime projection from shipped rest_write declarations
	CLISurface           *CLISurface                            // cli_surface.json
	RawCLISurface        json.RawMessage                        // verbatim cli_surface.json bytes; nil when absent
	RateLimits           *connsdk.RateLimits                    // rate_limits.json; nil when absent
}

// Metadata is the parsed metadata.json.
type Metadata struct {
	Name            string        `json:"name"`
	DisplayName     string        `json:"display_name"`
	Description     string        `json:"description"`
	IntegrationType string        `json:"integration_type"`
	DocsURL         string        `json:"docs_url,omitempty"`
	ReleaseStage    string        `json:"release_stage"`
	Capabilities    Capabilities  `json:"capabilities"`
	Batch           BatchSpec     `json:"batch,omitempty"`
	RateLimit       RateLimitSpec `json:"rate_limit,omitempty"`
	Risk            RiskSpec      `json:"risk,omitempty"`
}

// Capabilities mirrors metadata.json.capabilities.
type Capabilities struct {
	Check         bool `json:"check"`
	Read          bool `json:"read"`
	Write         bool `json:"write"`
	Query         bool `json:"query"`
	CDC           bool `json:"cdc"`
	DynamicSchema bool `json:"dynamic_schema"`

	// WriteDeclared keeps an omitted JSON member distinct from an explicit
	// false. An automatic source-cited mutation artifact is an opt-out that
	// must never be inferred from Go's false zero value.
	WriteDeclared bool `json:"-"`
}

// UnmarshalJSON preserves whether metadata.json.capabilities explicitly
// declared write. The other capability members retain their existing optional
// semantics; only write carries a source-executable-coverage opt-out.
func (c *Capabilities) UnmarshalJSON(raw []byte) error {
	type capabilitiesDocument struct {
		Check         bool  `json:"check"`
		Read          bool  `json:"read"`
		Write         *bool `json:"write"`
		Query         bool  `json:"query"`
		CDC           bool  `json:"cdc"`
		DynamicSchema bool  `json:"dynamic_schema"`
	}
	var document capabilitiesDocument
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&document); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid JSON: multiple top-level values")
		}
		return err
	}
	*c = Capabilities{
		Check:         document.Check,
		Read:          document.Read,
		Query:         document.Query,
		CDC:           document.CDC,
		DynamicSchema: document.DynamicSchema,
		WriteDeclared: document.Write != nil,
	}
	if document.Write != nil {
		c.Write = *document.Write
	}
	return nil
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
	URL          string               `json:"url"`
	UserAgent    string               `json:"user_agent,omitempty"`
	Headers      map[string]string    `json:"headers,omitempty"`
	Auth         []AuthSpec           `json:"auth,omitempty"`
	Pagination   *PaginationSpec      `json:"pagination,omitempty"`
	Check        *RequestSpec         `json:"check,omitempty"`
	ErrorMap     []ErrorRule          `json:"error_map,omitempty"`
	RateLimit    *RateLimitSpec       `json:"rate_limit,omitempty"`
	Routes       []OperationRouteSpec `json:"routes,omitempty"`
	TenantOrigin *TenantOriginSpec    `json:"tenant_origin,omitempty"`
}

// OperationRouteSpec is one named, declaration-owned provider route. BaseURL
// is either a fixed HTTP origin or the established connector base template;
// Version is a source-traced path prefix which must match the selected
// operation's declared path. The route remains closed because neither field is
// accepted from a command or record.
type OperationRouteSpec struct {
	Name    string `json:"name"`
	BaseURL string `json:"base_url"`
	Version string `json:"version,omitempty"`
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
	Method          string                `json:"method"`
	Path            string                `json:"path"`
	Query           map[string]QueryParam `json:"query,omitempty"`
	Body            map[string]any        `json:"body,omitempty"`
	BodyType        string                `json:"body_type,omitempty"`
	SuccessStatuses []string              `json:"success_statuses,omitempty"`
	MaxBytes        int                   `json:"max_bytes,omitempty"`
}

// AuthSpec describes one candidate authenticator, selected by "when" (first
// match wins).
type AuthSpec struct {
	Mode  string `json:"mode"` // none|bearer|basic|api_key_header|api_key_query|oauth2_client_credentials|oauth2_refresh_token|aws_sigv4|custom
	Token string `json:"token,omitempty"`

	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`

	Header string `json:"header,omitempty"`
	Prefix string `json:"prefix,omitempty"`
	Param  string `json:"param,omitempty"`
	Value  string `json:"value,omitempty"`

	AccessKeyID     string `json:"access_key_id,omitempty"`
	SecretAccessKey string `json:"secret_access_key,omitempty"`
	SessionToken    string `json:"session_token,omitempty"`
	AWSService      string `json:"aws_service,omitempty"`
	AWSRegion       string `json:"aws_region,omitempty"`

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
	// ClientAuthentication controls whether a refresh-token client's declared
	// credentials use the standard form-body or HTTP Basic authentication.
	ClientAuthentication string `json:"client_authentication,omitempty"`

	// RefreshToken is the templated initial refresh token for the
	// oauth2_refresh_token mode (normally "{{ secrets.refresh_token }}"). It is
	// resolved through Interpolate like every other AuthSpec field, so an
	// unresolved config/secrets key is a hard error rather than a silently
	// unauthenticated request.
	//
	// The mode reuses token_url/client_id/client_secret/scopes/extra_params
	// verbatim from oauth2_client_credentials rather than introducing a
	// parallel vocabulary for the same four fields; only the grant differs.
	RefreshToken string `json:"refresh_token,omitempty"`

	// RefreshTokenStoreKey names the secret key under which a PROVIDER-ROTATED
	// refresh token is persisted back to the caller's encrypted local
	// credential store — normally the same key RefreshToken reads from, e.g.
	// "refresh_token".
	//
	// It is declared rather than inferred on purpose. Interpolate resolves a
	// template to its VALUE, so the engine cannot recover a key name from
	// RefreshToken, and pattern-matching "{{ secrets.X }}" back to X would
	// silently overwrite a caller's secret whenever the guess happened to be
	// right and silently do nothing whenever it was not.
	//
	// Omitted means "this provider does not rotate": nothing is ever written,
	// and a rotated value (if one arrives anyway) is held in memory for the
	// process lifetime only. A connector whose provider DOES rotate must
	// declare this key, or it will present an invalidated grant on its next
	// run.
	RefreshTokenStoreKey string `json:"refresh_token_store_key,omitempty"`

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
	// BodyOffsetField moves an offset paginator value into the declared JSON
	// request body instead of sending it as a query parameter.
	BodyOffsetField string `json:"body_offset_field,omitempty"`

	CursorParam     string `json:"cursor_param,omitempty"`
	TokenPath       string `json:"token_path,omitempty"`        // cursor: token from body
	LastRecordField string `json:"last_record_field,omitempty"` // cursor: token from last record (stripe)
	StopPath        string `json:"stop_path,omitempty"`         // cursor: falsy body value stops (stripe)

	NextURLPath string `json:"next_url_path,omitempty"` // next_url type
	// BodyCursorField moves a cursor paginator token into the declared JSON
	// request body instead of sending it as a query parameter.
	BodyCursorField string `json:"body_cursor_field,omitempty"`
	// BodyPageField moves a page-number paginator value into the declared JSON
	// request body instead of sending it as a query parameter.
	BodyPageField string `json:"body_page_field,omitempty"`

	// start_index type — 1-based index pagination that carries its own total,
	// the shape SCIM 2.0 list responses use (RFC 7644 §3.4.2.4):
	//
	//	request:  ?startIndex=1&count=100
	//	response: {"totalResults":N,"itemsPerPage":M,"startIndex":S,"Resources":[…]}
	//
	// Named for the mechanism rather than for SCIM: any API that pages by a
	// 1-based index and reports a total is served by the same walk. Every field
	// below defaults to SCIM's own name, so a SCIM stream declares nothing
	// beyond {"type":"start_index","page_size":N}.
	StartIndexParam string `json:"start_index_param,omitempty"` // default "startIndex"
	CountParam      string `json:"count_param,omitempty"`       // default "count"
	TotalPath       string `json:"total_path,omitempty"`        // default "totalResults"
	StartIndexPath  string `json:"start_index_path,omitempty"`  // default "startIndex"
	// StartIndexBase is a pointer for the same reason StartPage is: an explicit
	// 0 (a 0-based server that still reports a total) must stay distinguishable
	// from an absent key, which unmarshals to the identical Go zero value. nil
	// means "not declared", defaulting to 1 — SCIM's first index.
	StartIndexBase *int `json:"start_index_base,omitempty"`

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
	Name                     string                         `json:"name"`
	Route                    string                         `json:"route,omitempty"`
	Method                   string                         `json:"method,omitempty"` // default GET
	Path                     string                         `json:"path"`
	Query                    map[string]QueryParam          `json:"query,omitempty"`
	Body                     map[string]any                 `json:"body,omitempty"` // POST-body streams
	BodyType                 string                         `json:"body_type,omitempty"`
	GraphQL                  *GraphQLRequestSpec            `json:"graphql,omitempty"`
	Records                  RecordsSpec                    `json:"records"`
	Pagination               *PaginationSpec                `json:"pagination,omitempty"` // overrides base
	Incremental              *IncrementalSpec               `json:"incremental,omitempty"`
	ComputedFields           map[string]string              `json:"computed_fields,omitempty"`
	ResponseFields           map[string]string              `json:"response_fields,omitempty"`
	ResponseHeaderProjection []ResponseHeaderProjectionSpec `json:"response_header_projection,omitempty"`
	ArrayZipProjection       *ArrayZipProjectionSpec        `json:"array_zip_projection,omitempty"`
	Headers                  map[string]string              `json:"headers,omitempty"`
	Projection               string                         `json:"projection,omitempty"` // "schema" (default) | "passthrough"
	SchemaRef                string                         `json:"schema"`
	FanOut                   *FanOutSpec                    `json:"fan_out,omitempty"`
	CartesianConfigFanOut    *CartesianConfigFanOutSpec     `json:"cartesian_config_fan_out,omitempty"`
	DateWindowFanOut         *DateWindowFanOutSpec          `json:"date_window_fan_out,omitempty"`
}

// DateWindowFanOutSpec schedules closed UTC date windows from source-declared
// configuration keys. It is not a transform language or response paginator.
type DateWindowFanOutSpec struct {
	StartDateConfigKey string `json:"start_date_config_key"`
	EndDateConfigKey   string `json:"end_date_config_key"`
	BatchSizeConfigKey string `json:"batch_size_config_key"`
	DateFromQueryParam string `json:"date_from_query_param"`
	DateToQueryParam   string `json:"date_to_query_param"`
	MaxBatchDays       int    `json:"max_batch_days"`
	MaxWindows         int    `json:"max_windows"`
}

// ResponseHeaderProjection binds response-level header names to positional
// values on every extracted row. Each declared mapping is closed: a response
// header must appear exactly once in AllowedHeaders, and its corresponding row
// value must be present before any record can emit.
type ResponseHeaderProjectionSpec struct {
	HeadersPath    string   `json:"headers_path"`
	ValuesPath     string   `json:"values_path"`
	HeaderName     string   `json:"header_name,omitempty"`
	ValueField     string   `json:"value_field,omitempty"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// ArrayZipProjectionSpec maps one extracted response object into records by
// copying source-declared scalar fields and zipping source-declared arrays at
// equal indexes. It is a closed projection, not a general transform language.
type ArrayZipProjectionSpec struct {
	StaticFields []ArrayZipFieldSpec `json:"static_fields,omitempty"`
	ArrayFields  []ArrayZipFieldSpec `json:"array_fields"`
}

// ArrayZipFieldSpec names one emitted field and its source-declared path.
type ArrayZipFieldSpec struct {
	Field string `json:"field"`
	Path  string `json:"path"`
}

// CartesianConfigFanOutSpec expands only the two closed PageSpeed-style config
// axes into declared query parameters. It is not a general transform language.
type CartesianConfigFanOutSpec struct {
	URLConfigKey       string   `json:"url_config_key"`
	StrategyConfigKey  string   `json:"strategy_config_key"`
	URLQueryParam      string   `json:"url_query_param"`
	StrategyQueryParam string   `json:"strategy_query_param"`
	Categories         []string `json:"categories"`
	CategoryQueryParam string   `json:"category_query_param"`
	MaxCombinations    int      `json:"max_combinations"`
	MaxRequests        int      `json:"max_requests"`
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

// QueryParam is a declared query entry for streams, checks, or write actions
// (gap-loop cycle-1 item 3,
// REVIEW-B.md cross-cutting adjudication 2: the recurring
// "optional/config-driven query param not expressible" gap — vitally
// `status`, bitly `size`, calendly `count`/page_size, gmail's two filters,
// searxng wave0 F6 — met the >=3 recurrence threshold). Declared on a stream,
// base check, or write action either as a PLAIN STRING (today's exact dialect: `Template`
// is that string, `OmitWhenAbsent` false, `Default` empty — a template
// referencing an absent config/secrets key is ALWAYS a hard error, zero
// migration risk for every existing bundle) or as an OBJECT
// `{"template": "...", "omit_when_absent": true, "default": "..."}` — an
// explicit opt-in dialect, never a blanket absent-key-falsy change to query
// templating (which would silently convert a mistyped/missing REQUIRED key
// from a fail-loud error into a silently-unfiltered request, the F4
// fail-open class the engine deliberately rejects elsewhere).
//
// WriteAction.Query preserves that dialect and adds one narrow, write-only
// case: an object-form entry with OmitWhenAbsent may omit its own unresolved
// record.* reference. It does not widen caller query input or relax malformed,
// wrong-source, or required-record failures; resolveWriteQueryParams owns the
// distinction while retaining the established config/secrets/incremental rules.
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
	// WrapField replaces each extracted record with an object that carries the
	// raw provider record under this field before projection. It preserves APIs
	// whose stable contract is one opaque provider object per record.
	WrapField string `json:"wrap_field,omitempty"`
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
	ParamFormat    string `json:"param_format,omitempty"`    // rfc3339|rfc3339_utc|unix_seconds|date
	OperatorPrefix string `json:"operator_prefix,omitempty"` // optional comparison prefix such as >=
	StartConfigKey string `json:"start_config_key,omitempty"`
	ClientFiltered bool   `json:"client_filtered,omitempty"`
}

// WriteAction is one entry in writes.json's "actions" array.
type WriteAction struct {
	Name   string `json:"name"`
	Kind   string `json:"kind"` // create|update|upsert|delete|custom
	Method string `json:"method"`
	Path   string `json:"path"`
	// BaseURL pins an action to a declaration-owned alternate provider origin.
	// It is an absolute origin (no path/query/userinfo), never caller input. This
	// covers APIs such as GitHub release uploads whose mutation endpoint is
	// intentionally hosted away from the connector's ordinary REST origin.
	BaseURL string `json:"base_url,omitempty"`
	// AllowedBaseURLOrigins declares the only ordinary connector API origins
	// whose credential may be used with this alternate action origin. It closes
	// a public upload host from receiving a private/Enterprise API credential.
	AllowedBaseURLOrigins []string `json:"allowed_base_url_origins,omitempty"`
	// Route selects one HTTPBase.Routes declaration. It is the preferred
	// declaration-owned routing path for reverse ETL and binary uploads;
	// BaseURL remains only for existing legacy write declarations.
	Route      string   `json:"route,omitempty"`
	PathFields []string `json:"path_fields,omitempty"`
	// Query is the OPTIONAL write-action query-parameter map. It uses the
	// same QueryParam string-or-object dialect as stream and check queries;
	// see QueryParam's doc comment. Write preparation resolves it through
	// resolveWriteQueryParams, preserving that dialect and its one
	// source-locked record.* omission rule without admitting a caller-provided
	// free-form query channel.
	//
	// Absent/empty means the request carries no query string at all, which
	// is exactly the behavior every write action had before this field
	// existed (executeWriteRecord passed a nil url.Values in all six
	// body_type branches). The field is strictly additive and opt-in: no
	// existing bundle changes behavior by its introduction.
	Query map[string]QueryParam `json:"query,omitempty"`
	// RedactFields identifies record fields redacted in generic source-table plan
	// samples and returned write errors. DryRunWrite preview warnings preserve
	// their resolved values.
	RedactFields []string `json:"redact_fields,omitempty"`
	BodyType     string   `json:"body_type,omitempty"` // json (default) | form | none | graphql | json_array | multipart | base64_upload | binary_upload | declared_batch
	// SuccessStatuses optionally narrows generic 2xx success to the exact
	// provider receipt statuses the action declares.
	SuccessStatuses []int `json:"success_statuses,omitempty"`
	// BodyRequired forces an empty JSON object onto the wire when body construction
	// resolves no fields. It is valid only for the default json body type.
	BodyRequired bool                `json:"body_required,omitempty"`
	BodyFields   []string            `json:"body_fields,omitempty"`
	BodyField    string              `json:"body_field,omitempty"`
	BodySchema   json.RawMessage     `json:"body_schema,omitempty"`
	GraphQL      *GraphQLRequestSpec `json:"graphql,omitempty"`
	Multipart    *MultipartSpec      `json:"multipart,omitempty"`
	Base64Upload *Base64UploadSpec   `json:"base64_upload,omitempty"`
	BinaryUpload *BinaryUploadSpec   `json:"binary_upload,omitempty"`
	// DeclaredBatch turns one record into a provider batch whose subrequests
	// may select only the named write actions in AllowedActions. Callers never
	// provide a method, path, headers, or raw body: those are resolved from the
	// existing typed action declarations and sealed into the ordinary preview.
	DeclaredBatch *DeclaredBatchSpec `json:"declared_batch,omitempty"`
	RecordSchema  json.RawMessage    `json:"record_schema"`
	// IdempotencyKeyHeader names a provider-documented request header. Execution
	// generates one fresh key per record and reuses it only across that record's retries.
	IdempotencyKeyHeader string `json:"idempotency_key_header,omitempty"`
	// DynamicFields optionally declares ONE record field as a typed
	// dynamic-key region. Absent means today's exact behavior.
	DynamicFields *DynamicFieldsSpec `json:"dynamic_fields,omitempty"`
	Delete        *DeleteSpec        `json:"delete,omitempty"`
	Risk          string             `json:"risk"`
	// Batchable gates the action out of SourceTable-driven bulk reverse ETL
	// when explicitly false. It is a pointer because bool's zero value is
	// false, and false is the restrictive setting: a plain bool would silently
	// mark every hand-constructed WriteAction as non-batchable. nil means the
	// bundle did not declare it, which is the permissive default every shipped
	// action relies on. Read it through IsBatchable, never directly.
	Batchable        *bool                              `json:"batchable,omitempty"`
	Confirm          string                             `json:"confirm,omitempty"` // legacy: "" | "destructive"
	Confirmation     *ConfirmationSpec                  `json:"confirmation,omitempty"`
	TransportBinding *connectors.TransportActionBinding `json:"transport_binding,omitempty"`
	Hook             string                             `json:"hook,omitempty"`
	// HookFields is a closed declaration-owned list of record fields consumed
	// only by a compound hook's declared follow-up route. They cannot overlap
	// the primary action's path or body fields, which keeps a hook supplement
	// from becoming a generic raw request/body channel.
	HookFields []string `json:"hook_fields,omitempty"`
}

// ConfirmationSpec is the closed, declarative confirmation policy shared by
// write actions and operation executors. Bundle schemas reject unknown kinds
// and fields before this type is decoded.
type ConfirmationSpec struct {
	Kind connectors.ConfirmationKind `json:"kind"`
}

// DynamicFieldsSpec declares ONE record field as a typed dynamic-key region,
// for providers that accept tenant-defined custom fields with no fixed,
// enumerable official set.
//
// It is deliberately NOT a raw-body escape hatch, and every field below exists
// to hold that line. Everything about the region is bundle metadata; only the
// keys and the SCALAR values inside it come from the caller:
//
//   - Values must be scalars drawn from ValueTypes, so no caller input can ever
//     become request STRUCTURE. This is the load-bearing invariant: an object
//     or array value is a hard error, never a coercion.
//   - Keys must match KeyPattern, which is declared in the bundle and is never
//     caller input.
//   - Keys may not collide with path_fields, body_fields, body_field, or any
//     key the body already carries, so a dynamic key can never shadow a
//     structural one.
//   - The region is merged into the JSON body only, AFTER path interpolation.
//     It reaches no other part of the request — not the URL, not the method,
//     not headers.
//   - MaxKeys and MaxValueBytes bound growth.
//
// record_schema stays CLOSED. The bundle declares the container field in it as
// an object; this spec validates the interior separately. That is what lets
// tenant-defined keys become expressible without opening additionalProperties.
type DynamicFieldsSpec struct {
	// Field is the record field holding the dynamic-key map.
	Field string `json:"field"`
	// KeyPattern is the regexp every caller-supplied key must match. It is
	// anchored at both ends when compiled, so a partial match cannot pass.
	KeyPattern string `json:"key_pattern"`
	// MaxKeys bounds how many dynamic keys one record may carry. 0 means the
	// built-in default.
	MaxKeys int `json:"max_keys,omitempty"`
	// ValueTypes is the allow-list of permitted JSON scalar types:
	// string, number, boolean, null. Empty means all four.
	ValueTypes []string `json:"value_types,omitempty"`
	// MaxValueBytes bounds the encoded length of a single dynamic value. 0
	// means the built-in default.
	MaxValueBytes int `json:"max_value_bytes,omitempty"`
	// Target selects where the region lands: "inline" (default) merges it at
	// the body root; "nested" keeps it under Field. Providers differ; both are
	// declarative and neither is caller-controlled.
	Target string `json:"target,omitempty"`
}

// IsBatchable reports whether the action may be executed from a bulk reverse
// ETL plan. Only an explicit "batchable": false in the bundle says no.
//
// Batchability is independent of Confirm: Confirm asks how severe one call is,
// IsBatchable asks whether the action may be fanned out over many records under
// a single approval. An action may declare either, both, or neither.
func (a WriteAction) IsBatchable() bool {
	return a.Batchable == nil || *a.Batchable
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
type MultipartSpec struct {
	MaxBytes int64               `json:"max_bytes,omitempty"`
	Parts    []MultipartPartSpec `json:"parts,omitempty"`
}

// DeclaredBatchSpec describes a closed provider batch envelope over existing
// named write actions. The field names are provider facts owned by the bundle;
// the executor is fixed engine code, not a JSON callback or generic HTTP hook.
// AllowedActions is an explicit allow-list so adding a new connector action
// never silently makes it batch-addressable.
type DeclaredBatchSpec struct {
	MaxActions            int      `json:"max_actions"`
	AllowedActions        []string `json:"allowed_actions"`
	AllowedMethods        []string `json:"allowed_methods"`
	ProviderEnvelopeField string   `json:"provider_envelope_field"`
	ProviderActionsField  string   `json:"provider_actions_field"`
	ProviderMethodField   string   `json:"provider_method_field"`
	ProviderPathField     string   `json:"provider_path_field"`
	ProviderDataField     string   `json:"provider_data_field"`
	InnerBodyField        string   `json:"inner_body_field"`
	ResponseEnvelopeField string   `json:"response_envelope_field"`
	ResponseStatusField   string   `json:"response_status_field"`
}

// Base64UploadSpec describes body_type "base64_upload": a JSON body carrying a
// base64-encoded payload in one declared property. It exists because APIs that
// want an inline encoded upload (Airtable's uploadAttachment among them) had no
// typed route at all — the only alternative would have been letting a caller
// hand the engine a raw body, which is banned outright.
//
// The spec stays deliberately small. The body is built by the ordinary rules
// (body_fields if declared, otherwise every record field minus path_fields) and
// exactly two things then happen: SourceField is REMOVED and ContentField is set
// to the validated base64 string. Everything else an API needs alongside the
// payload — a filename, a content type — is an ordinary record field already
// governed by the action's closed record_schema, so nothing here duplicates a
// constraint the schema dialect can already express.
type Base64UploadSpec struct {
	// Source selects where the payload comes from: "path" (default) reads a
	// local file, "base64" takes an already-encoded string. Both converge on the
	// same validated, canonically re-encoded string.
	Source string `json:"source,omitempty"`

	// SourceField is the record field holding the file path (Source "path") or
	// the encoded payload (Source "base64"). It never reaches the wire — in
	// "path" mode it holds a local filesystem path, and transmitting that would
	// leak the operator's directory layout to the provider.
	SourceField string `json:"source_field"`

	// ContentField is the JSON body property that receives the base64 payload.
	ContentField string `json:"content_field"`

	// MaxDecodedBytes bounds the payload's decoded size. Required and positive:
	// an unbounded inline upload is a memory-exhaustion vector, and the engine
	// additionally clamps this to maxBase64UploadDecodedBytes.
	MaxDecodedBytes int64 `json:"max_decoded_bytes"`

	// MaxEncodedBytes bounds the transmitted (encoded) size. Optional; defaults
	// to the base64 length of MaxDecodedBytes. Declared explicitly because real
	// APIs document the encoded limit — Airtable's attachment cap is 5 MB of
	// base64, not 5 MB of file.
	MaxEncodedBytes int64 `json:"max_encoded_bytes,omitempty"`

	// AllowedMediaTypes is required when this action is exposed as a public
	// binary_upload command. The action-level field stays optional so existing
	// internal-only base64 actions retain their declared compatibility.
	AllowedMediaTypes []string `json:"allowed_media_types,omitempty"`
}

// BinaryUploadSpec describes a declaration-owned application/octet-stream
// body read from a root-confined local file. The path itself is never sent.
// Preview binds its normalized identity and approved SHA-256; execution
// reopens, bounds, and hashes the file before sending the exact bytes once.
type BinaryUploadSpec struct {
	SourceField       string   `json:"source_field"`
	MaxBytes          int64    `json:"max_bytes"`
	AllowedMediaTypes []string `json:"allowed_media_types,omitempty"`
}

type MultipartPartSpec struct {
	Name        string                             `json:"name"`
	Type        string                             `json:"type"`
	Field       string                             `json:"field"`
	ContentType string                             `json:"content_type,omitempty"`
	MediaPolicy connectors.BinaryUploadMediaPolicy `json:"media_policy,omitempty"`
	// AllowedMediaTypes bounds what the part's bytes may sniff as. ContentType
	// is what the bundle asserts to the provider; this is what makes that
	// assertion checkable. Absent means unconstrained; present and empty is a
	// load error, so "bounded" and "unbounded" can never be confused.
	AllowedMediaTypes []string `json:"allowed_media_types,omitempty"`
	Required          bool     `json:"required,omitempty"`
	MaxBytes          int64    `json:"max_bytes,omitempty"`
}

type DeleteSpec struct {
	Idempotent      bool  `json:"idempotent,omitempty"`
	MissingOkStatus []int `json:"missing_ok_status,omitempty"`
}

// directWriteEndpoint is derived exclusively from executable operation JSON.
// It is not a provider-evidence or admission record.
type directWriteEndpoint struct {
	Method            string
	Path              string
	RESTWrite         bool
	GraphQLOperations []string
}

// OperationSpec is one reviewed, typed operation definition. Executors are
// opt-in per kind; unsupported kinds stay metadata-only and unknown kinds are
// rejected by the meta-schema.
type OperationSpec struct {
	ID            string            `json:"id"`
	Kind          string            `json:"kind"`
	Summary       string            `json:"summary"`
	Description   string            `json:"description,omitempty"`
	Risk          string            `json:"risk"`
	Approval      string            `json:"approval"`
	OutputPolicy  string            `json:"output_policy"`
	AuthScopes    []string          `json:"auth_scopes,omitempty"`
	MutationClass string            `json:"mutation_class,omitempty"`
	Destructive   bool              `json:"destructive,omitempty"`
	Confirmation  *ConfirmationSpec `json:"confirmation,omitempty"`
	// Batchable gates this operation out of a bulk plan when explicitly false.
	// It is a pointer because false is restrictive while the omitted default is
	// permissive; see WriteAction.Batchable for the matching write-action
	// contract. Read it through IsBatchable, never directly.
	Batchable       *bool                `json:"batchable,omitempty"`
	SecretSensitive bool                 `json:"secret_sensitive,omitempty"`
	SensitivePolicy *SensitivePolicySpec `json:"sensitive_policy,omitempty"`
	AuditEvent      string               `json:"audit_event,omitempty"`
	// Route selects the connector-owned HTTPBase.Routes declaration used by
	// direct read/write and binary operations. Empty selects the bundle's
	// ordinary declared base URL.
	Route     string                  `json:"route,omitempty"`
	REST      *RESTOperationSpec      `json:"rest,omitempty"`
	GraphQL   *GraphQLOperationSpec   `json:"graphql,omitempty"`
	XML       *XMLOperationSpec       `json:"xml,omitempty"`
	Binary    *BinaryOperationSpec    `json:"binary,omitempty"`
	File      *FileOperationSpec      `json:"file,omitempty"`
	LocalGit  *LocalGitOperationSpec  `json:"local_git,omitempty"`
	LocalFile *LocalFileOperationSpec `json:"local_file,omitempty"`
	Browser   *BrowserOperationSpec   `json:"browser,omitempty"`
	Composite *CompositeOperationSpec `json:"composite,omitempty"`
}

// IsBatchable reports whether the operation may be placed in a bulk plan.
// Only an explicit "batchable": false says no; one direct-write invocation is
// nevertheless always prepared as exactly one request.
func (o OperationSpec) IsBatchable() bool {
	return o.Batchable == nil || *o.Batchable
}

// OperationParameter is one parameter a REST operation accepts, as its
// provider specification declares it.
type OperationParameter struct {
	Name string `json:"name"`
	In   string `json:"in"`
	// CLIName is an optional declaration-owned spelling for a fixed path
	// placeholder. It exists when the runtime's safe placeholder differs from
	// the provider's public resource name; it never changes the wire mapping.
	CLIName    string                  `json:"cli_name,omitempty"`
	Type       string                  `json:"type,omitempty"`
	Required   bool                    `json:"required,omitempty"`
	Repeatable bool                    `json:"repeatable,omitempty"`
	Values     []string                `json:"values,omitempty"`
	Summary    string                  `json:"summary,omitempty"`
	Minimum    *connectors.ExactNumber `json:"minimum,omitempty"`
	Maximum    *connectors.ExactNumber `json:"maximum,omitempty"`
	// Schema and MaxBytes are required for a caller-provided header. Headers
	// are strings on the wire, so their schema is deliberately a bounded
	// string schema rather than a second generic request-body dialect.
	Schema   json.RawMessage `json:"schema,omitempty"`
	MaxBytes int             `json:"max_bytes,omitempty"`
}

type RESTOperationSpec struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	ContentType string `json:"content_type,omitempty"`
	MaxBytes    int    `json:"max_bytes,omitempty"`
	// Pagination is a direct-read operation's own pager. It deliberately
	// shadows neither HTTP.Pagination nor streams.json: an API commonly mixes
	// page/page_size and startIndex/count endpoints, and using the connector
	// global value for both sends a request the provider did not document.
	Pagination *PaginationSpec `json:"pagination,omitempty"`
	// PaginationParameters is the operation-declared, non-CLI subset of this
	// endpoint's provider parameters. It proves that every query mechanic in
	// Pagination is documented by this operation while keeping raw cursors and
	// page selectors out of the command flag surface.
	PaginationParameters []OperationParameter `json:"pagination_parameters,omitempty"`
	// Parameters is the operation's accepted parameter set, rendered from the
	// connector's schema-4 source lock. Command flags reference this bounded
	// set and cannot add an endpoint parameter that the operation lacks.
	//
	// It deliberately carries only what a flag needs (name, location, type,
	// requiredness, enum values, summary) and nothing about paging: page and
	// page-size parameters are excluded here, because paging comes from
	// the connector's declared pagination spec instead.
	Parameters []OperationParameter `json:"parameters,omitempty"`
	Query      map[string]string    `json:"query,omitempty"`
	Body       map[string]any       `json:"body,omitempty"`
	BodySchema json.RawMessage      `json:"body_schema,omitempty"`
	// Multipart is an opt-in, operation-level multipart/form-data contract.
	// It deliberately reuses writes.json's bounded field/file part model; an
	// absent block preserves metadata-only multipart rows until their connector
	// adoption lane supplies an executable declared contract.
	Multipart *MultipartSpec `json:"multipart,omitempty"`
	// RequiredQuery declares query-parameter cardinality for endpoints that
	// must never be called unfiltered — a listing that returns an entire
	// enterprise when no filter is supplied, for example. Every group must be
	// satisfied by at least one of its named parameters carrying a non-blank
	// value on the outgoing request; a value hardcoded in Query counts, since
	// the constraint is about the wire request rather than about who supplied
	// it. Empty (the default) imposes nothing.
	RequiredQuery []RequiredQueryGroup   `json:"required_query,omitempty"`
	Response      *OperationResponseSpec `json:"response,omitempty"`
	Redirect      *OperationRedirectSpec `json:"redirect,omitempty"`
}

// OperationResponseSpec is a narrow result metadata contract. Body bytes and
// media remain governed by the operation's existing output policy/max_bytes;
// only these named, bounded headers may appear in fixed-operation results.
type OperationResponseSpec struct {
	Headers         []OperationResponseHeaderSpec `json:"headers,omitempty"`
	SuccessStatuses []string                      `json:"success_statuses,omitempty"`
}

type OperationResponseHeaderSpec struct {
	Name     string `json:"name"`
	MaxBytes int    `json:"max_bytes"`
}

type OperationRedirectSpec struct {
	MaxHops         int      `json:"max_hops"`
	AllowSameOrigin bool     `json:"allow_same_origin,omitempty"`
	AllowedHosts    []string `json:"allowed_hosts,omitempty"`
}

// RequiredQueryGroup is one "at least one of these" constraint. Several groups
// compose as AND-of-ORs: "at least one of A or B, and at least one of C" is two
// groups, which is the shape of an endpoint requiring both a subject filter and
// a time window.
type RequiredQueryGroup struct {
	AnyOf []string `json:"any_of"`
}

// SensitivePolicySpec is the reverse-ETL sensitive/admin policy for an operation
// whose inputs or effects need encrypted credential storage or elevated-scope
// confirmation. It declares safe secret delivery (never inline CLI by default),
// the provider-specific transform that replaces a generic body template, the
// preflight check that runs without reading secret values, the approval mode,
// and the closed response-secret store route. Secret operation runtime output
// stays complete; no redacting response path is part of this policy.
type SensitivePolicySpec struct {
	InputMode string `json:"input_mode,omitempty"` // env | file | stdin | env_or_file | env_or_stdin (never "inline")
	// RedactFields is legacy compatibility metadata for non-secret operations.
	// Secret operations retain complete runtime output; credential secrecy is
	// provided by encrypted-at-rest storage rather than redaction.
	RedactFields []string `json:"redact_fields,omitempty"`
	Preflight    string   `json:"preflight,omitempty"`     // scope/availability check without reading secret values
	Transform    string   `json:"transform,omitempty"`     // none | github_secret_encryption | provider-specific
	ApprovalMode string   `json:"approval_mode,omitempty"` // typed_confirmation required for secret writes
	// ResponseSecretField and ResponseSecretStoreKey form the only live
	// response-secret route. The value is extracted from this fixed JSON field
	// and handed directly to RuntimeConfig.SecretStore. Runtime output retains
	// the provider response intact; protection is encryption at rest, not field
	// deletion.
	ResponseSecretField    string `json:"response_secret_field,omitempty"`
	ResponseSecretStoreKey string `json:"response_secret_store_key,omitempty"`
}

type GraphQLOperationSpec struct {
	// Document and OperationName are fixed bundle metadata. They are never
	// caller input, which is what keeps an operation declaration from becoming
	// a raw GraphQL transport escape hatch.
	Document      string `json:"document"`
	OperationName string `json:"operation_name"`
	// Path, MaxBytes, and VariablesSchema are required by the executable
	// direct-operation path. They remain optional here because legacy GraphQL
	// stream bindings are metadata-only operations until a generated command
	// supplies the complete closed contract.
	Path            string          `json:"path,omitempty"`
	MaxBytes        int             `json:"max_bytes,omitempty"`
	VariablesSchema json.RawMessage `json:"variables_schema,omitempty"`
	// VariablesPath is legacy metadata for non-executable GraphQL stream
	// bindings. The fixed-operation executor deliberately ignores it: caller
	// variables are admitted only through VariablesSchema at the root.
	VariablesPath string                          `json:"variables_path,omitempty"`
	Pagination    *GraphQLOperationPaginationSpec `json:"pagination,omitempty"`
}

// GraphQLOperationPaginationSpec is the closed cursor contract for a
// fixed-document GraphQL query. The connection and variable names originate
// in the declaration's fixed document; callers may only submit the next
// cursor returned by the preceding result.
type GraphQLOperationPaginationSpec struct {
	ConnectionPath           string `json:"connection_path"`
	CursorVariable           string `json:"cursor_variable"`
	PageSizeVariable         string `json:"page_size_variable,omitempty"`
	BackwardCursorVariable   string `json:"backward_cursor_variable,omitempty"`
	BackwardPageSizeVariable string `json:"backward_page_size_variable,omitempty"`
	MaxPageSize              int    `json:"max_page_size,omitempty"`
}

type XMLOperationSpec struct {
	EnvelopeTemplate string            `json:"envelope_template"`
	ResponsePath     string            `json:"response_path,omitempty"`
	Headers          map[string]string `json:"headers,omitempty"`
}

type BinaryOperationSpec struct {
	Method   string `json:"method"`
	Path     string `json:"path"`
	MaxBytes int    `json:"max_bytes,omitempty"`
	// Parameters admits only the same declaration-owned typed path/query/header
	// parameter dialect as REST operations. In particular, no arbitrary upload
	// metadata or request header map exists.
	Parameters []OperationParameter   `json:"parameters,omitempty"`
	Response   *OperationResponseSpec `json:"response,omitempty"`
	Redirect   *OperationRedirectSpec `json:"redirect,omitempty"`
	// Accept selects one fixed provider-documented response representation.
	// It is declaration-owned and cannot be overridden by command callers.
	Accept string `json:"accept,omitempty"`
	// AllowOverwrite permits replacing an existing destination file.
	AllowOverwrite bool `json:"allow_overwrite,omitempty"`
	// ExtractArchives is DECLARED by two existing github operations but is
	// refused at execution time: archive extraction is zip-slip and
	// decompression-bomb territory and is a separate capability, never a
	// flag. The field is retained so those bundles keep validating.
	ExtractArchives bool `json:"extract_archives,omitempty"`
	// AllowCrossHost permits redirects to ANY other origin. Credentials are
	// stripped on such a hop regardless. Off by default: download endpoints
	// redirect to CDNs constantly and 71 connectors authenticate with a
	// custom header that Go does NOT strip across domains.
	AllowCrossHost bool `json:"allow_cross_host,omitempty"`
	// AllowedHosts permits redirects to exactly these hosts. Credentials are
	// stripped on such a hop regardless.
	AllowedHosts []string `json:"allowed_hosts,omitempty"`
	// ContentTypes is the closed provider media-type allowlist for the streamed
	// response. The runtime parses and enforces it before opening a destination
	// file; an omitted list preserves legacy operations that have no exact media
	// declaration.
	ContentTypes []string `json:"content_types,omitempty"`
	// Charset, when declared for a text_export, is exact and enforced on the
	// response Content-Type. Empty preserves legacy declarations that specify a
	// bounded media type but no provider charset guarantee.
	Charset string `json:"charset,omitempty"`
	// StallTimeoutSeconds bounds how long the download may make NO progress.
	// It is not a wall-clock deadline, which would turn the byte cap into a
	// bandwidth requirement.
	StallTimeoutSeconds int `json:"stall_timeout_seconds,omitempty"`
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

// CLISurface is the parsed cli_surface.json execution command tree.
type CLISurface struct {
	Tagline     string            `json:"tagline"`
	Usage       string            `json:"usage"`
	Groups      []CLICommandGroup `json:"groups,omitempty"`
	GlobalFlags []CLIFlag         `json:"global_flags,omitempty"`
	Commands    []CLICommand      `json:"commands"`
	HelpTopics  []CLIHelpTopic    `json:"help_topics,omitempty"`
}

// CLICommandGroup is a rendered help grouping.
type CLICommandGroup struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Commands []string `json:"commands"`
}

// CLIFlag describes one command or global flag.
type CLIFlag struct {
	Name            string                  `json:"name"`
	Type            string                  `json:"type"`
	Summary         string                  `json:"summary,omitempty"`
	Values          []string                `json:"values,omitempty"`
	MapsTo          string                  `json:"maps_to,omitempty"`
	Format          string                  `json:"format,omitempty"`
	AllowEmpty      *bool                   `json:"allow_empty,omitempty"`
	Minimum         *connectors.ExactNumber `json:"minimum,omitempty"`
	Maximum         *connectors.ExactNumber `json:"maximum,omitempty"`
	Required        bool                    `json:"required,omitempty"`
	Repeatable      bool                    `json:"repeatable,omitempty"`
	EnvOnly         bool                    `json:"env_only,omitempty"`
	AllowBareString bool                    `json:"allow_bare_string,omitempty"`
	// MaxItems/MinItems bound a string_array flag's item count so a bounded
	// provider-search list can be enforced against the flag the user typed, not
	// only against the assembled body.
	MaxItems int `json:"max_items,omitempty"`
	MinItems int `json:"min_items,omitempty"`
	MaxBytes int `json:"max_bytes,omitempty"`
}

// CLIConstraint describes a provider-neutral validation rule over mapped
// command inputs.
type CLIConstraint struct {
	Kind          string   `json:"kind"`
	Fields        []string `json:"fields,omitempty"`
	Left          string   `json:"left"`
	Right         string   `json:"right"`
	Op            string   `json:"op"`
	ValueType     string   `json:"value_type,omitempty"`
	LeftFallback  string   `json:"left_fallback,omitempty"`
	RightFallback string   `json:"right_fallback,omitempty"`
	Message       string   `json:"message,omitempty"`
}

// CommandFoundation is the named missing capability for a deferred command.
// Target makes the future provider endpoint explicit so the runtime can reject
// a policy or excluded row rather than treating it as an executable binding.
type CommandFoundation struct {
	ID        string                  `json:"id"`
	Reason    string                  `json:"reason"`
	Component string                  `json:"component"`
	Evidence  string                  `json:"evidence"`
	Target    CommandFoundationTarget `json:"target"`
}

// CommandFoundationTarget is the exact admitted source and runtime binding of
// a deferred command. Method/path are duplicated with the command reference;
// the stable identities disambiguate operations sharing one transport.
type CommandFoundationTarget struct {
	SourceID            string                 `json:"source_id,omitempty"`
	ProviderOperationID string                 `json:"operation_id,omitempty"`
	Binding             CommandBindingIdentity `json:"binding,omitempty"`
	DestructiveKind     string                 `json:"destructive_kind,omitempty"`
	Method              string                 `json:"method"`
	Path                string                 `json:"path"`
}

// CommandUnsupportedDisposition is source-backed discovery metadata for an
// operation the CLI cannot represent or execute. It is intentionally separate
// from a missing foundation, which describes an implementation gap.
type CommandUnsupportedDisposition struct {
	Reason string                   `json:"reason"`
	Target CommandUnsupportedTarget `json:"target"`
}

type CommandUnsupportedTarget struct {
	SourceID            string `json:"source_id"`
	ProviderOperationID string `json:"operation_id"`
	Method              string `json:"method"`
	Path                string `json:"path"`
}

// CommandBindingIdentity selects one stream, write action, operation, or
// operation-free command independently of its shared transport endpoint.
type CommandBindingIdentity struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

// CLICommand is one provider-inspired command path.
type CLICommand struct {
	Path         string                         `json:"path"`
	Summary      string                         `json:"summary"`
	Intent       string                         `json:"intent"`
	Availability string                         `json:"availability"`
	Stream       string                         `json:"stream,omitempty"`
	Write        string                         `json:"write,omitempty"`
	Flags        []CLIFlag                      `json:"flags,omitempty"`
	Constraints  []CLIConstraint                `json:"constraints,omitempty"`
	Examples     []string                       `json:"examples,omitempty"`
	APISurface   []CLISurfaceEndpointRef        `json:"api_surface,omitempty"`
	OutputPolicy string                         `json:"output_policy,omitempty"`
	RedactFields []string                       `json:"redact_fields,omitempty"`
	Operation    string                         `json:"operation,omitempty"`
	Risk         string                         `json:"risk,omitempty"`
	Approval     string                         `json:"approval,omitempty"`
	Foundation   *CommandFoundation             `json:"foundation_gap,omitempty"`
	Unsupported  *CommandUnsupportedDisposition `json:"unsupported_disposition,omitempty"`
	Notes        string                         `json:"notes,omitempty"`
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
	metadata, changefeed, pollingWatermark, syncTransport, spec, streams, writes, operations, cliSurface, rateLimits *Schema
	err                                                                                                              error
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
	metaSchemas.changefeed = compileMeta(changefeedSchemaJSON)
	metaSchemas.pollingWatermark = compileMeta(pollingWatermarkSchemaJSON)
	metaSchemas.syncTransport = compileMeta(syncTransportSchemaJSON)
	metaSchemas.spec = compileMeta(specSchemaJSON)
	metaSchemas.streams = compileMeta(streamsSchemaJSON)
	metaSchemas.writes = compileMeta(writesSchemaJSON)
	metaSchemas.operations = compileMeta(operationsSchemaJSON)
	metaSchemas.cliSurface = compileMeta(cliSurfaceSchemaJSON)
	metaSchemas.rateLimits = compileMeta(rateLimitsSchemaJSON)
}

// requiredFiles lists the bundle files that must always exist relative to a
// bundle's directory, excepting streams.json (conditionally required).
var requiredFiles = []string{"metadata.json", "spec.json"}

// LoadAllError is the structured error LoadAll returns whenever one or more
// (but not necessarily all) bundle directories under fsys failed to load.
// Failures preserves discovery order (the same sorted directory-name order
// LoadAll iterates) so callers that want per-bundle granularity can use
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
// the currently-loadable subset can proceed with the returned bundles.
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
		b, err := loadBundle(fsys, name)
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
	return loadBundle(fsys, dirName)
}

func loadBundle(fsys fs.FS, dirName string) (Bundle, error) {
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
	changefeed, err := loadChangefeed(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}
	pollingWatermark, err := loadPollingWatermark(sub, dirName, metadata)
	if err != nil {
		return Bundle{}, err
	}
	syncTransport, err := loadSyncTransport(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}
	databaseDefinition, err := loadDatabaseDefinition(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}
	if err := validateCopyWorkerMaximum(syncTransport, databaseDefinition); err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: sync_transport.json: %w", dirName, err)
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
	if err := validateOperationRoutes(Bundle{Name: dirName, HTTP: httpBase}, streams, writes, operations); err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: declaration routes: %w", dirName, err)
	}
	if err := validateOperationRuntimeHeaderIsolation(httpBase, operations); err != nil {
		return Bundle{}, fmt.Errorf("load bundle %s: operations.json: %w", dirName, err)
	}
	directWriteEndpoints := deriveDirectWriteEndpoints(operations)

	schemas, err := loadStreamSchemas(sub, dirName, streams)
	if err != nil {
		return Bundle{}, err
	}

	cliSurface, rawCLISurface, err := loadCLISurface(sub, dirName)
	if err != nil {
		return Bundle{}, err
	}

	rateLimits, err := loadRateLimits(sub, dirName, spec)
	if err != nil {
		return Bundle{}, err
	}

	return Bundle{
		Name:                 dirName,
		Metadata:             metadata,
		Changefeed:           changefeed,
		PollingWatermark:     pollingWatermark,
		SyncTransport:        syncTransport,
		Database:             databaseDefinition,
		Spec:                 spec,
		RawSpec:              rawSpec,
		HTTP:                 httpBase,
		Streams:              streams,
		Writes:               writes,
		Operations:           operations,
		RawOperations:        rawOperations,
		Schemas:              schemas,
		directWriteEndpoints: directWriteEndpoints,
		CLISurface:           cliSurface,
		RawCLISurface:        rawCLISurface,
		RateLimits:           rateLimits,
	}, nil
}

// deriveDirectWriteEndpoints builds the shipped runtime endpoint check solely
// from rest_write and fixed graphql_mutation declarations. It verifies
// internal declaration consistency, not provider documented-surface
// provenance; #3773 owns that evidence model. GraphQL operations that share a
// physical POST path retain their individual IDs so a generated source root
// cannot borrow another root's transport binding.
func deriveDirectWriteEndpoints(operations []OperationSpec) []directWriteEndpoint {
	endpoints := make([]directWriteEndpoint, 0)
	graphQLByPath := make(map[string]int)
	for _, op := range operations {
		switch op.Kind {
		case "rest_write":
			if op.REST == nil {
				continue
			}
			endpoints = append(endpoints, directWriteEndpoint{
				Method:    strings.ToUpper(strings.TrimSpace(op.REST.Method)),
				Path:      op.REST.Path,
				RESTWrite: true,
			})
		case "graphql_mutation":
			if op.GraphQL == nil || strings.TrimSpace(op.GraphQL.Path) == "" {
				continue
			}
			key := http.MethodPost + ":" + op.GraphQL.Path
			if index, ok := graphQLByPath[key]; ok {
				endpoints[index].GraphQLOperations = append(endpoints[index].GraphQLOperations, op.ID)
				continue
			}
			graphQLByPath[key] = len(endpoints)
			endpoints = append(endpoints, directWriteEndpoint{
				Method:            http.MethodPost,
				Path:              op.GraphQL.Path,
				GraphQLOperations: []string{op.ID},
			})
		}
	}
	return endpoints
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

func loadChangefeed(sub fs.FS, dirName string) (*connectors.ChangefeedDescriptor, error) {
	if !fileExists(sub, "changefeed.json") {
		return nil, nil
	}
	raw, err := readFile(sub, "changefeed.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.changefeed.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: changefeed.json: %w", dirName, err)
	}
	var changefeed connectors.ChangefeedDescriptor
	if err := strictDecode(raw, &changefeed); err != nil {
		return nil, fmt.Errorf("load bundle %s: changefeed.json: %w", dirName, err)
	}
	if err := changefeed.Validate(); err != nil {
		return nil, fmt.Errorf("load bundle %s: changefeed.json: %w", dirName, err)
	}
	return changefeed.Clone(), nil
}

// loadPollingWatermark reads the optional native database polling declaration.
// It is a separate file from changefeed.json so a bounded watermark scan does
// not become a CDC claim or an API-surface command.
func loadPollingWatermark(sub fs.FS, dirName string, metadata Metadata) (*connectors.PollingWatermarkDescriptor, error) {
	if !fileExists(sub, "polling_watermark.json") {
		return nil, nil
	}
	if metadata.IntegrationType != "database" {
		return nil, fmt.Errorf("load bundle %s: polling_watermark.json requires metadata integration_type %q", dirName, "database")
	}
	raw, err := readFile(sub, "polling_watermark.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.pollingWatermark.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: polling_watermark.json: %w", dirName, err)
	}
	var declaration connectors.PollingWatermarkDescriptor
	if err := strictDecode(raw, &declaration); err != nil {
		return nil, fmt.Errorf("load bundle %s: polling_watermark.json: %w", dirName, err)
	}
	if err := declaration.Validate(); err != nil {
		return nil, fmt.Errorf("load bundle %s: polling_watermark.json: %w", dirName, err)
	}
	return declaration.Clone(), nil
}

// syncTransportDocument is a versioned on-disk declaration. Its strict JSON
// shape keeps an authored role from becoming executable through an unknown
// field, while the descriptor's own validation preserves the shared closed
// transport vocabulary.
type syncTransportDocument struct {
	SchemaVersion int                                        `json:"schema_version"`
	Source        *connectors.SourceTransportDescriptor      `json:"source_transport,omitempty"`
	Destination   *connectors.DestinationTransportDescriptor `json:"destination_transport,omitempty"`
}

func loadSyncTransport(sub fs.FS, dirName string) (*connectors.SyncTransportDescriptor, error) {
	if !fileExists(sub, "sync_transport.json") {
		return nil, nil
	}
	raw, err := readFile(sub, "sync_transport.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.syncTransport.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: sync_transport.json: %w", dirName, err)
	}
	var document syncTransportDocument
	if err := strictDecode(raw, &document); err != nil {
		return nil, fmt.Errorf("load bundle %s: sync_transport.json: %w", dirName, err)
	}
	if document.SchemaVersion != 1 {
		return nil, fmt.Errorf("load bundle %s: sync_transport.json: unsupported schema version %d", dirName, document.SchemaVersion)
	}
	descriptor := connectors.SyncTransportDescriptor{Source: document.Source, Destination: document.Destination}
	if err := descriptor.Validate(); err != nil {
		return nil, fmt.Errorf("load bundle %s: sync_transport.json: %w", dirName, err)
	}
	return descriptor.Clone(), nil
}

// loadDatabaseDefinition keeps database.json optional for the existing broad
// connector fleet, while making every present declaration pass through the
// shared closed loader during normal engine and connectorgen bundle loading.
// The definition is policy only; this does not register a database driver or
// promote any connector capability.
func loadDatabaseDefinition(sub fs.FS, dirName string) (*database.Definition, error) {
	if !fileExists(sub, "database.json") {
		return nil, nil
	}
	definition, err := database.Load(context.Background(), sub)
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: database.json: %w", dirName, err)
	}
	return &definition, nil
}

// validateCopyWorkerMaximum makes the optional transport COPY declaration a
// projection of the destination's own typed database pool policy. A connector
// may omit it, but it cannot advertise more COPY connections than its
// database.json declares.
func validateCopyWorkerMaximum(transport *connectors.SyncTransportDescriptor, definition *database.Definition) error {
	if transport == nil || transport.Destination == nil || transport.Destination.CopyWorkerMaximum == 0 {
		return nil
	}
	if definition == nil {
		return fmt.Errorf("copy_worker_maximum requires a database resource declaration")
	}
	poolMaximum := definition.Resources().Pool.Maximum
	if transport.Destination.CopyWorkerMaximum > poolMaximum {
		return fmt.Errorf("copy_worker_maximum exceeds declared database pool maximum %d", poolMaximum)
	}
	return nil
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
	if err := validateResponseHeaderProjections(doc.Streams); err != nil {
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: streams.json: %w", dirName, err)
	}
	if err := validateArrayZipProjections(doc.Streams); err != nil {
		return HTTPBase{}, nil, fmt.Errorf("load bundle %s: streams.json: %w", dirName, err)
	}
	if err := validateStaticStreamHeaders(doc.Streams); err != nil {
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
	if err := validateWriteActionNames(doc.Actions); err != nil {
		return nil, fmt.Errorf("load bundle %s: writes.json: %w", dirName, err)
	}
	if err := validateWriteBodies(doc.Actions); err != nil {
		return nil, fmt.Errorf("load bundle %s: writes.json: %w", dirName, err)
	}
	return doc.Actions, nil
}

func validateWriteActionNames(actions []WriteAction) error {
	seen := make(map[string]struct{}, len(actions))
	for index, action := range actions {
		if _, duplicate := seen[action.Name]; duplicate {
			return fmt.Errorf("action %d duplicates write action name %q", index, action.Name)
		}
		seen[action.Name] = struct{}{}
	}
	return nil
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

const maxStaticStreamHeaders = 8

func validateStaticStreamHeaders(streams []StreamSpec) error {
	for streamIndex, stream := range streams {
		if len(stream.Headers) > maxStaticStreamHeaders {
			return fmt.Errorf("stream %d (%q) has %d headers, maximum is %d", streamIndex, stream.Name, len(stream.Headers), maxStaticStreamHeaders)
		}
		for name, value := range stream.Headers {
			if name != "Accept" {
				return fmt.Errorf("stream %d (%q) header %q is not permitted; only fixed Accept headers are supported", streamIndex, stream.Name, name)
			}
			if strings.Contains(value, "{{") || strings.Contains(value, "}}") {
				return fmt.Errorf("stream %d (%q) Accept header must be static", streamIndex, stream.Name)
			}
			mediaType, parameters, err := mime.ParseMediaType(value)
			if err != nil || len(parameters) != 0 || !strings.HasPrefix(strings.ToLower(mediaType), "application/vnd.") || !strings.HasSuffix(strings.ToLower(mediaType), "+json") {
				return fmt.Errorf("stream %d (%q) Accept header %q must be one fixed vendor JSON media type", streamIndex, stream.Name, value)
			}
		}
	}
	return nil
}

func validateResponseHeaderProjections(streams []StreamSpec) error {
	for streamIndex, stream := range streams {
		seenHeaders := make(map[string]struct{})
		for projectionIndex, projection := range stream.ResponseHeaderProjection {
			if strings.TrimSpace(projection.HeadersPath) == "" || strings.TrimSpace(projection.ValuesPath) == "" {
				return fmt.Errorf("stream %d (%q) response_header_projection %d requires headers_path and values_path", streamIndex, stream.Name, projectionIndex)
			}
			if strings.TrimSpace(projection.HeaderName) == "" && projection.HeaderName != "" {
				return fmt.Errorf("stream %d (%q) response_header_projection %d header_name is blank", streamIndex, stream.Name, projectionIndex)
			}
			if strings.TrimSpace(projection.ValueField) == "" && projection.ValueField != "" {
				return fmt.Errorf("stream %d (%q) response_header_projection %d value_field is blank", streamIndex, stream.Name, projectionIndex)
			}
			if len(projection.AllowedHeaders) == 0 {
				return fmt.Errorf("stream %d (%q) response_header_projection %d requires allowed_headers", streamIndex, stream.Name, projectionIndex)
			}
			for _, header := range projection.AllowedHeaders {
				header = strings.TrimSpace(header)
				if header == "" {
					return fmt.Errorf("stream %d (%q) response_header_projection %d has a blank allowed header", streamIndex, stream.Name, projectionIndex)
				}
				if _, duplicate := seenHeaders[header]; duplicate {
					return fmt.Errorf("stream %d (%q) response_header_projection duplicates allowed header %q", streamIndex, stream.Name, header)
				}
				seenHeaders[header] = struct{}{}
			}
		}
	}
	return nil
}

func validateArrayZipProjections(streams []StreamSpec) error {
	for streamIndex, stream := range streams {
		projection := stream.ArrayZipProjection
		if projection == nil {
			continue
		}
		if len(projection.ArrayFields) == 0 {
			return fmt.Errorf("stream %d (%q) array_zip_projection requires array_fields", streamIndex, stream.Name)
		}
		seen := make(map[string]struct{}, len(projection.StaticFields)+len(projection.ArrayFields))
		for _, fields := range [][]ArrayZipFieldSpec{projection.StaticFields, projection.ArrayFields} {
			for _, field := range fields {
				if strings.TrimSpace(field.Field) == "" || strings.TrimSpace(field.Field) != field.Field {
					return fmt.Errorf("stream %d (%q) array_zip_projection has an invalid field name", streamIndex, stream.Name)
				}
				if strings.TrimSpace(field.Path) == "" || strings.TrimSpace(field.Path) != field.Path || strings.Contains(field.Path, "..") {
					return fmt.Errorf("stream %d (%q) array_zip_projection field %q has an invalid path", streamIndex, stream.Name, field.Field)
				}
				if _, duplicate := seen[field.Field]; duplicate {
					return fmt.Errorf("stream %d (%q) array_zip_projection duplicates field %q", streamIndex, stream.Name, field.Field)
				}
				seen[field.Field] = struct{}{}
			}
		}
	}
	return nil
}

// dynamicFieldsValueTypes is the closed set of JSON SCALAR types a dynamic
// value may take. "object" and "array" are deliberately absent and must never
// be added: admitting them would let caller input become request structure,
// which is exactly the escape hatch this primitive exists to avoid.
var dynamicFieldsValueTypes = map[string]bool{
	"string": true, "number": true, "boolean": true, "null": true,
}

// validateDynamicFields enforces the declaration-time half of the dynamic-key
// contract. The execution-time half lives in write.go's applyDynamicFields.
func validateDynamicFields(i int, action WriteAction) error {
	spec := action.DynamicFields
	if spec == nil {
		return nil
	}
	if bodyType := bodyTypeOf(action); bodyType != "json" {
		return fmt.Errorf("action %d (%q) dynamic_fields requires body_type json, got %q", i, action.Name, bodyType)
	}
	field := strings.TrimSpace(spec.Field)
	if field == "" {
		return fmt.Errorf("action %d (%q) dynamic_fields requires field", i, action.Name)
	}
	if strings.TrimSpace(spec.KeyPattern) == "" {
		return fmt.Errorf("action %d (%q) dynamic_fields requires key_pattern", i, action.Name)
	}
	if _, err := compileDynamicKeyPattern(spec.KeyPattern); err != nil {
		return fmt.Errorf("action %d (%q) dynamic_fields key_pattern: %w", i, action.Name, err)
	}
	for _, vt := range spec.ValueTypes {
		if !dynamicFieldsValueTypes[vt] {
			return fmt.Errorf("action %d (%q) dynamic_fields value_types contains unsupported type %q (scalars only)", i, action.Name, vt)
		}
	}
	switch strings.TrimSpace(spec.Target) {
	case "", "inline", "nested":
	default:
		return fmt.Errorf("action %d (%q) dynamic_fields target must be inline or nested, got %q", i, action.Name, spec.Target)
	}
	if spec.MaxKeys < 0 || spec.MaxValueBytes < 0 {
		return fmt.Errorf("action %d (%q) dynamic_fields bounds must be non-negative", i, action.Name)
	}
	// The container field is consumed by the region itself, so it must not
	// also be claimed as a path or body field.
	for _, pf := range action.PathFields {
		if pf == field {
			return fmt.Errorf("action %d (%q) dynamic_fields field %q also declared in path_fields", i, action.Name, field)
		}
	}
	for _, bf := range action.BodyFields {
		if bf == field {
			return fmt.Errorf("action %d (%q) dynamic_fields field %q also declared in body_fields", i, action.Name, field)
		}
	}
	if action.BodyField == field {
		return fmt.Errorf("action %d (%q) dynamic_fields field %q also declared as body_field", i, action.Name, field)
	}
	return nil
}

// compileDynamicKeyPattern anchors the declared pattern at both ends so a
// partial match can never admit a key the bundle did not intend.
func compileDynamicKeyPattern(pattern string) (*regexp.Regexp, error) {
	p := pattern
	if !strings.HasPrefix(p, "^") {
		p = "^" + p
	}
	if !strings.HasSuffix(p, "$") {
		p += "$"
	}
	return regexp.Compile(p)
}

func validateWriteBodies(actions []WriteAction) error {
	actionsByName := make(map[string]WriteAction, len(actions))
	for _, action := range actions {
		actionsByName[action.Name] = action
	}
	for i, action := range actions {
		if err := validateDynamicFields(i, action); err != nil {
			return err
		}
		if err := validateWriteHookFields(i, action); err != nil {
			return err
		}
		bodyType := bodyTypeOf(action)
		if err := validateWriteActionBaseURL(i, action); err != nil {
			return err
		}
		if err := validateWriteActionSuccessStatuses(i, action); err != nil {
			return err
		}
		if action.BodyRequired && bodyType != "json" {
			return fmt.Errorf("action %d (%q) body_required requires body_type json, got %q", i, action.Name, bodyType)
		}
		if header := strings.TrimSpace(action.IdempotencyKeyHeader); header != "" && !httpHeaderNamePattern.MatchString(header) {
			return fmt.Errorf("action %d (%q) idempotency_key_header %q is not a valid HTTP header name", i, action.Name, action.IdempotencyKeyHeader)
		}
		if action.GraphQL != nil && bodyType != "graphql" {
			return fmt.Errorf("action %d (%q) declares graphql but body_type is %q", i, action.Name, bodyType)
		}
		if action.Multipart != nil && bodyType != "multipart" {
			return fmt.Errorf("action %d (%q) declares multipart but body_type is %q", i, action.Name, bodyType)
		}
		if action.Base64Upload != nil && bodyType != "base64_upload" {
			return fmt.Errorf("action %d (%q) declares base64_upload but body_type is %q", i, action.Name, bodyType)
		}
		if action.BinaryUpload != nil && bodyType != "binary_upload" {
			return fmt.Errorf("action %d (%q) declares binary_upload but body_type is %q", i, action.Name, bodyType)
		}
		if action.DeclaredBatch != nil && bodyType != "declared_batch" {
			return fmt.Errorf("action %d (%q) declares declared_batch but body_type is %q", i, action.Name, bodyType)
		}
		switch bodyType {
		case "graphql":
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
		case "json_array":
			if strings.TrimSpace(action.BodyField) == "" {
				return fmt.Errorf("action %d (%q) body_type json_array requires body_field", i, action.Name)
			}
			if len(action.BodySchema) == 0 {
				return fmt.Errorf("action %d (%q) body_type json_array requires body_schema", i, action.Name)
			}
		case "base64_upload":
			if err := validateBase64UploadSpec(i, action); err != nil {
				return err
			}
		case "binary_upload":
			if err := validateBinaryUploadSpec(i, action); err != nil {
				return err
			}
		case "declared_batch":
			if err := validateDeclaredBatchSpec(i, action, actionsByName); err != nil {
				return err
			}
		case "multipart":
			if action.Multipart == nil || len(action.Multipart.Parts) == 0 {
				return fmt.Errorf("action %d (%q) body_type multipart requires multipart.parts", i, action.Name)
			}
			for j, part := range action.Multipart.Parts {
				if strings.TrimSpace(part.Name) == "" || strings.TrimSpace(part.Field) == "" {
					return fmt.Errorf("action %d (%q) multipart part %d requires name and field", i, action.Name, j)
				}
				switch part.Type {
				case "field", "file":
				default:
					return fmt.Errorf("action %d (%q) multipart part %d has unsupported type %q", i, action.Name, j, part.Type)
				}
				if err := validateMultipartMediaTypes(part); err != nil {
					return fmt.Errorf("action %d (%q) multipart part %d: %w", i, action.Name, j, err)
				}
			}
		}
	}
	return nil
}

func validateDeclaredBatchSpec(index int, action WriteAction, actionsByName map[string]WriteAction) error {
	spec := action.DeclaredBatch
	if spec == nil {
		return fmt.Errorf("action %d (%q) body_type declared_batch requires declared_batch", index, action.Name)
	}
	if !strings.EqualFold(strings.TrimSpace(action.Method), http.MethodPost) {
		return fmt.Errorf("action %d (%q) declared_batch method must be POST", index, action.Name)
	}
	if spec.MaxActions < 1 || spec.MaxActions > 64 {
		return fmt.Errorf("action %d (%q) declared_batch max_actions must be between 1 and 64", index, action.Name)
	}
	fields := map[string]string{
		"provider_envelope_field": spec.ProviderEnvelopeField,
		"provider_actions_field":  spec.ProviderActionsField,
		"provider_method_field":   spec.ProviderMethodField,
		"provider_path_field":     spec.ProviderPathField,
		"provider_data_field":     spec.ProviderDataField,
		"inner_body_field":        spec.InnerBodyField,
		"response_envelope_field": spec.ResponseEnvelopeField,
		"response_status_field":   spec.ResponseStatusField,
	}
	for name, field := range fields {
		if !isPreparedWriteBindingField(field) {
			return fmt.Errorf("action %d (%q) declared_batch %s must be a simple field name", index, action.Name, name)
		}
	}
	if spec.ProviderMethodField == spec.ProviderPathField || spec.ProviderMethodField == spec.ProviderDataField || spec.ProviderPathField == spec.ProviderDataField {
		return fmt.Errorf("action %d (%q) declared_batch provider method, path, and data fields must be distinct", index, action.Name)
	}
	allowedMethods := make(map[string]struct{}, len(spec.AllowedMethods))
	for _, raw := range spec.AllowedMethods {
		method := strings.ToUpper(strings.TrimSpace(raw))
		switch method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		default:
			return fmt.Errorf("action %d (%q) declared_batch allowed method %q is unsupported", index, action.Name, raw)
		}
		if _, duplicate := allowedMethods[method]; duplicate {
			return fmt.Errorf("action %d (%q) declared_batch repeats allowed method %q", index, action.Name, method)
		}
		allowedMethods[method] = struct{}{}
	}
	if len(allowedMethods) == 0 {
		return fmt.Errorf("action %d (%q) declared_batch requires allowed_methods", index, action.Name)
	}
	seenActions := make(map[string]struct{}, len(spec.AllowedActions))
	for _, name := range spec.AllowedActions {
		if name == action.Name {
			return fmt.Errorf("action %d (%q) declared_batch cannot select itself", index, action.Name)
		}
		if _, duplicate := seenActions[name]; duplicate {
			return fmt.Errorf("action %d (%q) declared_batch repeats allowed action %q", index, action.Name, name)
		}
		seenActions[name] = struct{}{}
		inner, ok := actionsByName[name]
		if !ok {
			return fmt.Errorf("action %d (%q) declared_batch references unknown write action %q", index, action.Name, name)
		}
		method := strings.ToUpper(strings.TrimSpace(inner.Method))
		if _, ok := allowedMethods[method]; !ok {
			return fmt.Errorf("action %d (%q) declared_batch action %q method %s is outside allowed_methods", index, action.Name, name, method)
		}
		switch bodyTypeOf(inner) {
		case "json", "none":
		default:
			return fmt.Errorf("action %d (%q) declared_batch action %q has unsupported body_type %q", index, action.Name, name, bodyTypeOf(inner))
		}
		if strings.TrimSpace(inner.Hook) != "" || strings.TrimSpace(inner.BaseURL) != "" || strings.TrimSpace(inner.Route) != "" || strings.TrimSpace(inner.IdempotencyKeyHeader) != "" {
			return fmt.Errorf("action %d (%q) declared_batch action %q requires unsupported alternate execution semantics", index, action.Name, name)
		}
		if confirmationKindForWriteAction(inner) == string(connectors.ConfirmationKindDestructive) && confirmationKindForWriteAction(action) != string(connectors.ConfirmationKindDestructive) {
			return fmt.Errorf("action %d (%q) declared_batch selecting destructive action %q requires destructive confirmation", index, action.Name, name)
		}
	}
	if len(seenActions) == 0 {
		return fmt.Errorf("action %d (%q) declared_batch requires allowed_actions", index, action.Name)
	}
	return nil
}

func validateWriteHookFields(i int, action WriteAction) error {
	if len(action.HookFields) == 0 {
		return nil
	}
	if strings.TrimSpace(action.Hook) == "" {
		return fmt.Errorf("action %d (%q) hook_fields requires hook", i, action.Name)
	}
	properties, _, err := recordSchemaTopLevelProperties(action.RecordSchema)
	if err != nil {
		return fmt.Errorf("action %d (%q) hook_fields record_schema: %w", i, action.Name, err)
	}
	seen := make(map[string]struct{}, len(action.HookFields))
	for _, field := range action.HookFields {
		if strings.TrimSpace(field) == "" || strings.TrimSpace(field) != field {
			return fmt.Errorf("action %d (%q) hook_fields contains an invalid field", i, action.Name)
		}
		if _, duplicate := seen[field]; duplicate {
			return fmt.Errorf("action %d (%q) hook_fields duplicates %q", i, action.Name, field)
		}
		seen[field] = struct{}{}
		if _, declared := properties[field]; !declared {
			return fmt.Errorf("action %d (%q) hook_fields field %q is absent from record_schema", i, action.Name, field)
		}
		if containsWriteField(action.PathFields, field) || containsWriteField(action.BodyFields, field) || action.BodyField == field {
			return fmt.Errorf("action %d (%q) hook_fields field %q overlaps the primary request contract", i, action.Name, field)
		}
	}
	return nil
}

func containsWriteField(fields []string, wanted string) bool {
	for _, field := range fields {
		if field == wanted {
			return true
		}
	}
	return false
}

const maxBinaryUploadBytes = int64(64 << 20)

func validateWriteActionBaseURL(i int, action WriteAction) error {
	if strings.TrimSpace(action.BaseURL) == "" && len(action.AllowedBaseURLOrigins) > 0 {
		return fmt.Errorf("action %d (%q) allowed_base_url_origins requires base_url", i, action.Name)
	}
	if strings.TrimSpace(action.BaseURL) != "" {
		if action.BaseURL != strings.TrimSpace(action.BaseURL) || strings.Contains(action.BaseURL, "{{") {
			return fmt.Errorf("action %d (%q) base_url must be one fixed absolute origin", i, action.Name)
		}
		if _, err := fixedHTTPOrigin(action.BaseURL); err != nil {
			return fmt.Errorf("action %d (%q) base_url must be one fixed absolute HTTP origin", i, action.Name)
		}
	}
	seen := make(map[string]struct{}, len(action.AllowedBaseURLOrigins))
	for _, raw := range action.AllowedBaseURLOrigins {
		origin, err := fixedHTTPOrigin(raw)
		if err != nil || raw != strings.TrimSpace(raw) || strings.Contains(raw, "{{") {
			return fmt.Errorf("action %d (%q) allowed_base_url_origins entries must be fixed absolute HTTP origins", i, action.Name)
		}
		if _, duplicate := seen[origin]; duplicate {
			return fmt.Errorf("action %d (%q) allowed_base_url_origins must not repeat %q", i, action.Name, raw)
		}
		seen[origin] = struct{}{}
	}
	return nil
}

func fixedHTTPOrigin(raw string) (string, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return "", fmt.Errorf("not a fixed HTTP origin")
	}
	return strings.ToLower(parsed.Scheme) + "://" + strings.ToLower(parsed.Host), nil
}

func validateWriteActionSuccessStatuses(i int, action WriteAction) error {
	seen := make(map[int]struct{}, len(action.SuccessStatuses))
	for _, status := range action.SuccessStatuses {
		if status < http.StatusOK || status >= http.StatusMultipleChoices {
			return fmt.Errorf("action %d (%q) success_statuses entry %d must be a 2xx status", i, action.Name, status)
		}
		if _, duplicate := seen[status]; duplicate {
			return fmt.Errorf("action %d (%q) success_statuses repeats %d", i, action.Name, status)
		}
		seen[status] = struct{}{}
	}
	return nil
}

func validateBinaryUploadSpec(i int, action WriteAction) error {
	spec := action.BinaryUpload
	if spec == nil {
		return fmt.Errorf("action %d (%q) body_type binary_upload requires binary_upload", i, action.Name)
	}
	if strings.TrimSpace(spec.SourceField) == "" {
		return fmt.Errorf("action %d (%q) binary_upload requires source_field", i, action.Name)
	}
	if spec.MaxBytes <= 0 || spec.MaxBytes > maxBinaryUploadBytes {
		return fmt.Errorf("action %d (%q) binary_upload max_bytes must be between 1 and %d", i, action.Name, maxBinaryUploadBytes)
	}
	if err := validateOptionalUploadMediaTypes(spec.AllowedMediaTypes); err != nil {
		return fmt.Errorf("action %d (%q) binary_upload %w", i, action.Name, err)
	}
	return nil
}

// validateBase64UploadSpec fails a bundle whose base64_upload declaration could
// not be executed safely. Each rule closes a way the declaration could look
// valid but behave wrongly at runtime, which for an upload means transmitting
// something the author did not intend.
func validateBase64UploadSpec(i int, action WriteAction) error {
	spec := action.Base64Upload
	if spec == nil {
		return fmt.Errorf("action %d (%q) body_type base64_upload requires base64_upload", i, action.Name)
	}
	switch spec.Source {
	case "", "path", "base64":
	default:
		return fmt.Errorf("action %d (%q) base64_upload source must be path or base64, got %q", i, action.Name, spec.Source)
	}
	if strings.TrimSpace(spec.SourceField) == "" {
		return fmt.Errorf("action %d (%q) base64_upload requires source_field", i, action.Name)
	}
	if strings.TrimSpace(spec.ContentField) == "" {
		return fmt.Errorf("action %d (%q) base64_upload requires content_field", i, action.Name)
	}
	// In path mode the source field holds a local filesystem path. Naming the
	// same field as the content target would mean the removal and the
	// assignment collide, and whether the path leaks would depend on ordering.
	// In base64 mode the two may legitimately coincide: the field is simply
	// replaced by its canonically re-encoded self.
	if spec.Source != "base64" && spec.SourceField == spec.ContentField {
		return fmt.Errorf("action %d (%q) base64_upload source_field and content_field must differ in path mode", i, action.Name)
	}
	if spec.MaxDecodedBytes <= 0 {
		return fmt.Errorf("action %d (%q) base64_upload requires positive max_decoded_bytes", i, action.Name)
	}
	if spec.MaxDecodedBytes > maxBase64UploadDecodedBytes {
		return fmt.Errorf("action %d (%q) base64_upload max_decoded_bytes %d exceeds the engine ceiling %d", i, action.Name, spec.MaxDecodedBytes, maxBase64UploadDecodedBytes)
	}
	// An encoded bound below the encoded length of the decoded bound can never
	// be satisfied by a payload at the decoded bound, so the pair is
	// contradictory — an authoring mistake, caught here rather than as a
	// runtime rejection of every large payload.
	if spec.MaxEncodedBytes > 0 {
		needed := int64(base64.StdEncoding.EncodedLen(int(spec.MaxDecodedBytes)))
		if spec.MaxEncodedBytes < needed {
			return fmt.Errorf("action %d (%q) base64_upload max_encoded_bytes %d cannot hold max_decoded_bytes %d (needs %d)", i, action.Name, spec.MaxEncodedBytes, spec.MaxDecodedBytes, needed)
		}
	}
	if err := validateOptionalUploadMediaTypes(spec.AllowedMediaTypes); err != nil {
		return fmt.Errorf("action %d (%q) base64_upload %w", i, action.Name, err)
	}
	return nil
}

// validateOptionalUploadMediaTypes keeps the action declaration honest when
// it elects a media policy. Public binary-upload promotion separately requires
// a non-empty list; leaving it absent remains valid for internal-only writes.
func validateOptionalUploadMediaTypes(mediaTypes []string) error {
	if mediaTypes == nil {
		return nil
	}
	if len(mediaTypes) == 0 {
		return fmt.Errorf("allowed_media_types must not be empty; omit it to leave the action unconstrained")
	}
	for _, raw := range mediaTypes {
		if _, _, err := mime.ParseMediaType(raw); err != nil {
			return fmt.Errorf("allowed_media_types entry %q is not a valid media type: %w", raw, err)
		}
	}
	return nil
}

// validateProviderSearchSemantics enforces the provider_search contract at load
// time, so an unbounded or open-bodied declaration cannot ship.
//
// provider_search is a read that carries a fixed POST body containing bounded
// lists. It is a distinct kind rather than a convention over rest_read because
// these rules have to be enforceable, and because #2985's recorded decision is
// that provider search is a separate typed capability — `pm query` stays
// warehouse-focused. Nothing here lets a caller choose the method, the path, or
// a body key the schema does not declare.
func validateProviderSearchSemantics(i int, op OperationSpec) error {
	method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
	if method != "POST" {
		return fmt.Errorf("operation %d (%q) provider_search method must be POST, got %s", i, op.ID, method)
	}
	if isAbsoluteHTTPURL(op.REST.Path) {
		return fmt.Errorf("operation %d (%q) provider_search path must be connector-relative, got an absolute URL", i, op.ID)
	}
	if !strings.EqualFold(strings.TrimSpace(op.REST.ContentType), "application/json") {
		return fmt.Errorf("operation %d (%q) provider_search requires application/json content_type", i, op.ID)
	}
	if op.REST.MaxBytes <= 0 {
		return fmt.Errorf("operation %d (%q) provider_search must declare positive max_bytes", i, op.ID)
	}
	if strings.TrimSpace(op.MutationClass) != "" && op.MutationClass != "none" {
		return fmt.Errorf("operation %d (%q) provider_search is a read and must not declare mutating mutation_class %q", i, op.ID, op.MutationClass)
	}
	if len(op.REST.BodySchema) == 0 {
		return fmt.Errorf("operation %d (%q) provider_search must declare body_schema", i, op.ID)
	}
	var body map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &body); err != nil {
		return fmt.Errorf("operation %d (%q) provider_search body_schema is not an object: %w", i, op.ID, err)
	}
	if closed, ok := body["additionalProperties"].(bool); !ok || closed {
		return fmt.Errorf("operation %d (%q) provider_search body_schema must declare additionalProperties: false so no undeclared body key can be supplied", i, op.ID)
	}
	if _, err := CompileSchema(op.REST.BodySchema); err != nil {
		return fmt.Errorf("operation %d (%q) provider_search body_schema: %w", i, op.ID, err)
	}
	return requireBoundedArrays(i, op.ID, body, "body_schema")
}

// requireBoundedArrays walks a provider_search body schema and refuses any array
// property that does not declare maxItems. This is what makes the kind bounded
// by construction rather than by convention: an unbounded list cannot be
// declared, so it cannot reach a provider.
func requireBoundedArrays(i int, id string, node map[string]any, path string) error {
	if isArrayType(node) {
		if _, ok := node["maxItems"]; !ok {
			return fmt.Errorf("operation %d (%q) provider_search %s declares an array without maxItems; every list must be bounded", i, id, path)
		}
	}
	if isObjectType(node) {
		if closed, ok := node["additionalProperties"].(bool); !ok || closed {
			return fmt.Errorf("operation %d (%q) provider_search %s is an object and must declare additionalProperties: false so no undeclared body key can be supplied inside it", i, id, path)
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		if err := requireBoundedArrays(i, id, items, path+"/items"); err != nil {
			return err
		}
	}
	props, ok := node["properties"].(map[string]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(props))
	for name := range props {
		names = append(names, name)
	}
	sort.Strings(names) // deterministic error for a bundle with several unbounded lists
	for _, name := range names {
		child, ok := props[name].(map[string]any)
		if !ok {
			continue
		}
		if err := requireBoundedArrays(i, id, child, path+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

// isObjectType reports whether a schema node is an object. It recognises the
// single string form ("object") as well as the multi-form type list that
// compileTypes accepts (e.g. ["object","null"]), and a properties-bearing node
// with no string type at all. Every object node in a provider_search body must
// be closed, regardless of how the dialect lets it be declared.
func isObjectType(node map[string]any) bool {
	switch typeOf := node["type"].(type) {
	case string:
		return typeOf == "object"
	case []any:
		for _, t := range typeOf {
			if s, ok := t.(string); ok && s == "object" {
				return true
			}
		}
		return false
	}
	_, hasProps := node["properties"]
	return hasProps
}

// isArrayType reports whether a schema node is an array. It recognises the
// single string form ("array") as well as the multi-form type list that
// compileTypes accepts (e.g. ["array","null"]), and an items-bearing node with
// no string type at all. The bound must hold regardless of how the dialect lets
// a list be declared, so ["array","null"] cannot smuggle an unbounded list in.
func isArrayType(node map[string]any) bool {
	switch typeOf := node["type"].(type) {
	case string:
		if typeOf == "array" {
			return true
		}
		return false
	case []any:
		for _, t := range typeOf {
			if s, ok := t.(string); ok && s == "array" {
				return true
			}
		}
		return false
	}
	if _, ok := node["items"]; ok {
		return true
	}
	return false
}

// validateMultipartMediaTypes enforces the media-type declaration on one part.
// A present-but-empty list is rejected rather than read as "allow anything":
// absent means unconstrained and present means bounded, and a bundle must not be
// able to look bounded while permitting everything.
func validateMultipartMediaTypes(part MultipartPartSpec) error {
	switch part.MediaPolicy {
	case "":
		// Existing declaration semantics continue below.
	case connectors.BinaryUploadMediaPolicyProviderUnrestricted:
		if part.Type != "file" {
			return fmt.Errorf("media_policy %q is only meaningful on a file part, got type %q", part.MediaPolicy, part.Type)
		}
		if part.AllowedMediaTypes != nil {
			return fmt.Errorf("media_policy %q must not declare allowed_media_types", part.MediaPolicy)
		}
		if strings.TrimSpace(part.ContentType) != "" {
			return fmt.Errorf("media_policy %q must not declare content_type", part.MediaPolicy)
		}
		return nil
	default:
		return fmt.Errorf("unsupported media_policy %q", part.MediaPolicy)
	}
	if part.AllowedMediaTypes == nil {
		return nil
	}
	if len(part.AllowedMediaTypes) == 0 {
		return fmt.Errorf("allowed_media_types must not be empty; omit it to leave the part unconstrained")
	}
	if part.Type != "file" {
		return fmt.Errorf("allowed_media_types is only meaningful on a file part, got type %q", part.Type)
	}
	parsed := make([]string, 0, len(part.AllowedMediaTypes))
	for _, raw := range part.AllowedMediaTypes {
		value, _, err := mime.ParseMediaType(raw)
		if err != nil {
			return fmt.Errorf("allowed_media_types entry %q is not a valid media type: %w", raw, err)
		}
		parsed = append(parsed, value)
	}
	declared := strings.TrimSpace(part.ContentType)
	if declared == "" {
		return nil
	}
	want, _, err := mime.ParseMediaType(declared)
	if err != nil {
		return fmt.Errorf("content_type %q is not a valid media type: %w", declared, err)
	}
	for _, allowed := range parsed {
		if strings.EqualFold(allowed, want) {
			return nil
		}
	}
	return fmt.Errorf("content_type %q is not among its own allowed_media_types %s", declared, strings.Join(part.AllowedMediaTypes, ", "))
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

// validateRequiredQuery rejects a required_query group that could never be
// satisfied. The meta-schema already enforces a non-empty any_of; this catches
// the blank-name case it cannot express. Both failures are load errors rather
// than silent no-ops: a constraint that never fires is worse than an absent one,
// because the bundle author believes the endpoint is protected.
func validateRequiredQuery(i int, op OperationSpec) error {
	if op.REST == nil {
		return nil
	}
	for j, group := range op.REST.RequiredQuery {
		if len(group.AnyOf) == 0 {
			return fmt.Errorf("operation %d (%q) required_query group %d must name at least one parameter", i, op.ID, j)
		}
		for _, name := range group.AnyOf {
			if strings.TrimSpace(name) == "" {
				return fmt.Errorf("operation %d (%q) required_query group %d has a blank parameter name", i, op.ID, j)
			}
		}
	}
	return nil
}

// validateOperationMultipartSemantics closes the operation-level multipart
// contract before a connector can expose it as an executable rest_write. The
// ordinary writes.json multipart path remains its own established contract;
// this is deliberately opt-in so a historical content_type annotation cannot
// become a live upload merely by loading a new engine version.
func validateOperationMultipartSemantics(i int, op OperationSpec) error {
	if op.REST == nil || op.REST.Multipart == nil {
		return nil
	}
	if op.Kind != "rest_write" {
		return fmt.Errorf("operation %d (%q) rest.multipart is only valid for rest_write operations, got %q", i, op.ID, op.Kind)
	}
	if op.REST.ContentType != "multipart/form-data" {
		return fmt.Errorf("operation %d (%q) rest.multipart requires literal content_type multipart/form-data", i, op.ID)
	}
	if strings.TrimSpace(op.REST.Path) == "" || isAbsoluteHTTPURL(op.REST.Path) || strings.HasPrefix(strings.TrimSpace(op.REST.Path), "//") {
		return fmt.Errorf("operation %d (%q) rest.multipart endpoint must be connector-relative", i, op.ID)
	}
	if op.REST.MaxBytes <= 0 {
		return fmt.Errorf("operation %d (%q) rest.multipart requires positive rest.max_bytes response capture limit", i, op.ID)
	}
	if len(op.REST.BodySchema) == 0 {
		return fmt.Errorf("operation %d (%q) rest.multipart requires body_schema", i, op.ID)
	}

	var bodySchema map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &bodySchema); err != nil {
		return fmt.Errorf("operation %d (%q) rest.multipart body_schema is not an object: %w", i, op.ID, err)
	}
	if !isObjectType(bodySchema) {
		return fmt.Errorf("operation %d (%q) rest.multipart body_schema must be an object", i, op.ID)
	}
	if _, err := CompileSchema(op.REST.BodySchema); err != nil {
		return fmt.Errorf("operation %d (%q) rest.multipart body_schema: %w", i, op.ID, err)
	}
	if err := requireClosedMultipartBodySchema(i, op.ID, bodySchema, "body_schema"); err != nil {
		return err
	}

	multipart := op.REST.Multipart
	if multipart.MaxBytes <= 0 {
		return fmt.Errorf("operation %d (%q) rest.multipart requires a positive aggregate max_bytes", i, op.ID)
	}
	if len(multipart.Parts) == 0 {
		return fmt.Errorf("operation %d (%q) rest.multipart requires non-empty parts", i, op.ID)
	}
	partNames := make(map[string]struct{}, len(multipart.Parts))
	for partIndex, part := range multipart.Parts {
		name := strings.TrimSpace(part.Name)
		if name == "" || strings.TrimSpace(part.Field) == "" {
			return fmt.Errorf("operation %d (%q) rest.multipart part %d requires name and field", i, op.ID, partIndex)
		}
		if _, duplicate := partNames[name]; duplicate {
			return fmt.Errorf("operation %d (%q) rest.multipart part %d duplicates name %q", i, op.ID, partIndex, name)
		}
		partNames[name] = struct{}{}

		fieldSchema, required, err := multipartBodySchemaField(bodySchema, part.Field)
		if err != nil {
			return fmt.Errorf("operation %d (%q) rest.multipart part %d must reference a declared body field %q: %w", i, op.ID, partIndex, part.Field, err)
		}
		switch part.Type {
		case "field":
			// The compiled closed body schema remains the source of truth for
			// every scalar/object type a field part may carry.
		case "file":
			if !part.Required || !required || !multipartSchemaString(fieldSchema) {
				return fmt.Errorf("operation %d (%q) rest.multipart file part %q must reference a required string body field", i, op.ID, part.Name)
			}
			if part.MaxBytes <= 0 {
				return fmt.Errorf("operation %d (%q) rest.multipart file part %q requires a positive max_bytes", i, op.ID, part.Name)
			}
			if strings.TrimSpace(part.ContentType) == "" && len(part.AllowedMediaTypes) == 0 && part.MediaPolicy == "" {
				return fmt.Errorf("operation %d (%q) rest.multipart file part %q requires declared media policy", i, op.ID, part.Name)
			}
		default:
			return fmt.Errorf("operation %d (%q) rest.multipart part %d has unsupported type %q", i, op.ID, partIndex, part.Type)
		}
		if declared := strings.TrimSpace(part.ContentType); declared != "" {
			if _, _, err := mime.ParseMediaType(declared); err != nil {
				return fmt.Errorf("operation %d (%q) rest.multipart part %d content_type %q is not a valid media type: %w", i, op.ID, partIndex, part.ContentType, err)
			}
		}
		if err := validateMultipartMediaTypes(part); err != nil {
			return fmt.Errorf("operation %d (%q) rest.multipart part %d: %w", i, op.ID, partIndex, err)
		}
	}
	return nil
}

// requireClosedMultipartBodySchema recursively closes every object reachable
// from an operation multipart body. Command flags can materialize dotted body
// paths, so closing only the root would still leave an undeclared nested body
// key available to a caller.
func requireClosedMultipartBodySchema(i int, id string, node map[string]any, path string) error {
	if isObjectType(node) {
		if closed, ok := node["additionalProperties"].(bool); !ok || closed {
			return fmt.Errorf("operation %d (%q) rest.multipart %s is an object and must declare additionalProperties: false", i, id, path)
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		if err := requireClosedMultipartBodySchema(i, id, items, path+"/items"); err != nil {
			return err
		}
	}
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return nil
	}
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		child, ok := properties[name].(map[string]any)
		if !ok {
			continue
		}
		if err := requireClosedMultipartBodySchema(i, id, child, path+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

// multipartBodySchemaField resolves a dotted part field through the declared
// body schema and reports whether every object edge is required. File parts
// need that stronger guarantee so a missing path cannot silently become an
// optional inline payload at execution time.
func multipartBodySchemaField(root map[string]any, field string) (map[string]any, bool, error) {
	current := root
	required := true
	for _, segment := range strings.Split(field, ".") {
		if strings.TrimSpace(segment) == "" {
			return nil, false, fmt.Errorf("field path contains an empty segment")
		}
		properties, ok := current["properties"].(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("field path has no declared property %q", segment)
		}
		raw, ok := properties[segment]
		if !ok {
			return nil, false, fmt.Errorf("field path has no declared property %q", segment)
		}
		next, ok := raw.(map[string]any)
		if !ok {
			return nil, false, fmt.Errorf("field path property %q is not a schema object", segment)
		}
		if !multipartSchemaRequired(current, segment) {
			required = false
		}
		current = next
	}
	return current, required, nil
}

func multipartSchemaRequired(node map[string]any, name string) bool {
	required, ok := node["required"].([]any)
	if !ok {
		return false
	}
	for _, raw := range required {
		if candidate, ok := raw.(string); ok && candidate == name {
			return true
		}
	}
	return false
}

func multipartSchemaString(node map[string]any) bool {
	typeName, ok := node["type"].(string)
	return ok && typeName == "string"
}

func validateOperationSemantics(i int, op OperationSpec) error {
	if err := validateRequiredQuery(i, op); err != nil {
		return err
	}
	if err := validateOperationHeaderParameters(op); err != nil {
		return fmt.Errorf("operation %d (%q): %w", i, op.ID, err)
	}
	if err := validateOperationResponseContract(op); err != nil {
		return fmt.Errorf("operation %d (%q): %w", i, op.ID, err)
	}
	if err := validateOperationMultipartSemantics(i, op); err != nil {
		return err
	}
	switch op.Kind {
	case "rest_read":
		if err := validateRESTOperationPagination(op); err != nil {
			return fmt.Errorf("operation %d (%q) rest_read: %w", i, op.ID, err)
		}
		method := strings.ToUpper(strings.TrimSpace(op.REST.Method))
		if method != "GET" && method != "POST" {
			return fmt.Errorf("operation %d (%q) rest_read method must be GET or POST, got %s", i, op.ID, method)
		}
		if method == "POST" && len(op.REST.BodySchema) == 0 {
			return fmt.Errorf("operation %d (%q) rest_read POST must declare body_schema", i, op.ID)
		}
		// text/plain is a new, closed operation contract. Validate it at load
		// time so a bundle cannot declare raw input without its root-string
		// schema. Keep existing non-text POST metadata on its established
		// loader path; the operation direct-read preflight performs its stricter
		// executable-contract validation only when a command names it.
		if method == "POST" && operationDirectReadContentType(op) == "text/plain" {
			if err := validateOperationDirectReadTextPlainContract(op); err != nil {
				return fmt.Errorf("operation %d (%q) rest_read POST: %w", i, op.ID, err)
			}
		}
		if strings.TrimSpace(op.MutationClass) != "" && op.MutationClass != "none" {
			return fmt.Errorf("operation %d (%q) rest_read must not declare mutating mutation_class %q", i, op.ID, op.MutationClass)
		}
	case "rest_status":
		if op.REST == nil {
			return fmt.Errorf("operation %d (%q) rest_status requires rest metadata", i, op.ID)
		}
		if method := strings.ToUpper(strings.TrimSpace(op.REST.Method)); method != http.MethodHead {
			return fmt.Errorf("operation %d (%q) rest_status method must be HEAD, got %s", i, op.ID, method)
		}
		if op.OutputPolicy != "status" {
			return fmt.Errorf("operation %d (%q) rest_status output_policy must be status", i, op.ID)
		}
		if op.REST.MaxBytes <= 0 || op.REST.MaxBytes > 1024 {
			return fmt.Errorf("operation %d (%q) rest_status max_bytes must be between 1 and 1024", i, op.ID)
		}
		if len(op.REST.Body) != 0 || len(op.REST.BodySchema) != 0 || strings.TrimSpace(op.REST.ContentType) != "" {
			return fmt.Errorf("operation %d (%q) rest_status must not declare a request body", i, op.ID)
		}
	case "provider_search":
		if err := validateProviderSearchSemantics(i, op); err != nil {
			return err
		}
	case "graphql_query":
		if err := validateGraphQLOperationDeclaration(op, "query"); err != nil {
			return fmt.Errorf("operation %d (%q) graphql_query: %w", i, op.ID, err)
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
	case "graphql_mutation":
		if err := validateGraphQLOperationDeclaration(op, "mutation"); err != nil {
			return fmt.Errorf("operation %d (%q) graphql_mutation: %w", i, op.ID, err)
		}
		if strings.TrimSpace(op.MutationClass) == "" || op.MutationClass == "none" {
			return fmt.Errorf("operation %d (%q) %s must declare mutation_class", i, op.ID, op.Kind)
		}
		if strings.TrimSpace(op.Approval) == "" || op.Approval == "none" {
			return fmt.Errorf("operation %d (%q) %s must declare approval requirements", i, op.ID, op.Kind)
		}
	case "xml_import":
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
		if accept := strings.TrimSpace(op.Binary.Accept); accept != "" {
			if _, _, err := mime.ParseMediaType(accept); err != nil {
				return fmt.Errorf("operation %d (%q) binary_download accept %q is not a valid media type: %w", i, op.ID, op.Binary.Accept, err)
			}
		}
		if err := validateOperationBinaryContentTypes(op); err != nil {
			return fmt.Errorf("operation %d (%q) binary_download: %w", i, op.ID, err)
		}
	case "text_export":
		if method := strings.ToUpper(strings.TrimSpace(op.Binary.Method)); method != http.MethodGet {
			return fmt.Errorf("operation %d (%q) text_export method must be GET, got %s", i, op.ID, method)
		}
		if op.Binary.MaxBytes <= 0 {
			return fmt.Errorf("operation %d (%q) text_export must declare positive max_bytes", i, op.ID)
		}
		if !strings.EqualFold(strings.TrimSpace(op.Binary.Accept), "text/csv") {
			return fmt.Errorf("operation %d (%q) text_export accept must be text/csv", i, op.ID)
		}
		if op.OutputPolicy != "file_manifest" {
			return fmt.Errorf("operation %d (%q) text_export output_policy must be file_manifest", i, op.ID)
		}
		if err := validateOperationBinaryContentTypes(op); err != nil {
			return fmt.Errorf("operation %d (%q) text_export: %w", i, op.ID, err)
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

// validateRESTOperationPagination binds a per-operation pager to the query
// parameters imported from that endpoint's source contract. This is kept out
// of Parameters deliberately: those values must never be exposed as raw CLI
// flags alongside --page/--page-cursor.
func validateRESTOperationPagination(op OperationSpec) error {
	if op.REST == nil || op.REST.Pagination == nil {
		return nil
	}
	if op.Kind != "rest_read" {
		return fmt.Errorf("pagination is only supported by rest_read operations")
	}
	spec := *op.REST.Pagination
	if _, err := newPaginator(spec, spec.PageSize, ""); err != nil {
		return fmt.Errorf("pagination is invalid: %w", err)
	}
	expected := restPaginationQueryParameters(spec)
	if len(expected) == 0 {
		return nil
	}
	expectedSet := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		expectedSet[name] = struct{}{}
	}
	source := make(map[string]struct{}, len(op.REST.PaginationParameters))
	for _, parameter := range op.REST.PaginationParameters {
		name := strings.TrimSpace(parameter.Name)
		if parameter.In != "query" || name == "" {
			return fmt.Errorf("pagination source parameter must be a named query parameter")
		}
		source[name] = struct{}{}
	}
	if len(source) == 0 {
		return fmt.Errorf("pagination declares query mechanics but has no source pagination_parameters")
	}
	for _, name := range expected {
		if _, ok := source[name]; !ok {
			return fmt.Errorf("pagination parameter %q disagrees with source pagination_parameters", name)
		}
	}
	for name := range source {
		if _, ok := expectedSet[name]; !ok {
			return fmt.Errorf("source pagination parameter %q is not used by the declared pagination", name)
		}
	}
	return nil
}

func restPaginationQueryParameters(spec PaginationSpec) []string {
	var names []string
	appendName := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		for _, existing := range names {
			if existing == name {
				return
			}
		}
		names = append(names, name)
	}
	switch strings.TrimSpace(spec.Type) {
	case "page_number":
		appendName(spec.PageParam)
		appendName(spec.SizeParam)
	case "offset_limit":
		appendName(spec.OffsetParam)
		appendName(spec.LimitParam)
		appendName(spec.SizeParam)
	case "cursor":
		appendName(spec.CursorParam)
		appendName(spec.SizeParam)
	case "start_index":
		appendName(valueOrDefault(spec.StartIndexParam, defaultStartIndexParam))
		appendName(valueOrDefault(spec.CountParam, defaultStartIndexCount))
		appendName(spec.SizeParam)
	case "next_url":
		appendName(spec.SizeParam)
		appendName(spec.LimitParam)
		appendName(spec.OffsetParam)
	case "link_header":
		appendName(spec.SizeParam)
	case "none", "":
		// These have no operation query mechanics to verify.
	}
	return names
}

// validateSensitivePolicy enforces the sensitive/admin reverse-ETL policy model
// (#41). An operation that is secret_sensitive or has mutation_class "secret"
// must declare a sensitive_policy with: a non-inline input_mode and
// approval_mode "typed_confirmation". The transform,
// when set, must be a known value. Live secret writes remain blocked in this
// issue; live execution is admitted only through the closed secret-store
// response contract below.
func validateSensitivePolicy(i int, op OperationSpec) error {
	isSecret := op.SecretSensitive || strings.EqualFold(op.MutationClass, "secret")
	if !isSecret {
		// Non-secret operations may still declare a policy (e.g. admin actions
		// that redact fields) but are not forced to.
		if op.SensitivePolicy == nil {
			return nil
		}
	} else if op.SensitivePolicy == nil {
		return fmt.Errorf("operation %d (%q) is secret_sensitive but declares no sensitive_policy (input_mode, approval_mode)", i, op.ID)
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
	switch strings.ToLower(strings.TrimSpace(p.Transform)) {
	case "", "none", "github_secret_encryption":
		// allowed
	default:
		return fmt.Errorf("operation %d (%q) sensitive_policy transform %q is not a known value", i, op.ID, p.Transform)
	}
	if isSecret && !strings.EqualFold(p.ApprovalMode, "typed_confirmation") {
		return fmt.Errorf("operation %d (%q) sensitive_policy approval_mode must be typed_confirmation for secret writes", i, op.ID)
	}
	if (p.ResponseSecretField == "") != (p.ResponseSecretStoreKey == "") {
		return fmt.Errorf("operation %d (%q) sensitive_policy response_secret_field and response_secret_store_key must be declared together", i, op.ID)
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
			Raw:         append(json.RawMessage(nil), raw...),
		}
	}
	return out, nil
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
	for index, command := range surface.Commands {
		if command.Foundation != nil {
			if err := ValidateCommandEndpoint(command.Foundation.Target.Method, command.Foundation.Target.Path); err != nil {
				return nil, nil, fmt.Errorf("load bundle %s: cli_surface.json: command %d foundation target: %w", dirName, index, err)
			}
		}
		if command.Unsupported != nil {
			if err := ValidateCommandEndpoint(command.Unsupported.Target.Method, command.Unsupported.Target.Path); err != nil {
				return nil, nil, fmt.Errorf("load bundle %s: cli_surface.json: command %d unsupported target: %w", dirName, index, err)
			}
		}
	}
	return &surface, json.RawMessage(raw), nil
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
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("invalid JSON: multiple top-level values")
		}
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
