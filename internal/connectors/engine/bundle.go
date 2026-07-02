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

// RequestSpec is a minimal method+path request descriptor (used by "check").
type RequestSpec struct {
	Method string `json:"method"`
	Path   string `json:"path"`
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
	StartPage int    `json:"start_page,omitempty"`

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
	Name           string            `json:"name"`
	Method         string            `json:"method,omitempty"` // default GET
	Path           string            `json:"path"`
	Query          map[string]string `json:"query,omitempty"`
	Body           map[string]any    `json:"body,omitempty"` // POST-body streams
	Records        RecordsSpec       `json:"records"`
	Pagination     *PaginationSpec   `json:"pagination,omitempty"` // overrides base
	Incremental    *IncrementalSpec  `json:"incremental,omitempty"`
	ComputedFields map[string]string `json:"computed_fields,omitempty"`
	Projection     string            `json:"projection,omitempty"` // "schema" (default) | "passthrough"
	SchemaRef      string            `json:"schema"`
}

// RecordsSpec describes how to extract records from a page body.
type RecordsSpec struct {
	Path         string      `json:"path"` // dotted path; "." = body root
	SingleObject bool        `json:"single_object,omitempty"`
	Filter       *FilterSpec `json:"filter,omitempty"`
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

// LoadAll loads and validates every bundle directory found at the root of
// fsys. Non-directory root entries are skipped. An empty tree is not an
// error (returns zero bundles).
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
	for _, name := range names {
		b, err := Load(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("load all bundles: %s: %w", name, err)
		}
		bundles = append(bundles, b)
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

	spec, err := loadSpec(sub, dirName)
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
	if err := json.Unmarshal(raw, &m); err != nil {
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

func loadSpec(sub fs.FS, dirName string) (*Schema, error) {
	raw, err := readFile(sub, "spec.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.spec.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: spec.json: %w", dirName, err)
	}
	sch, err := CompileSchema(raw)
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: spec.json: %w", dirName, err)
	}
	return sch, nil
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
	if err := json.Unmarshal(raw, &doc); err != nil {
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
	if err := json.Unmarshal(raw, &doc); err != nil {
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
	if err := json.Unmarshal(raw, &surface); err != nil {
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
