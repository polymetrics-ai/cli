# Polymetrics Connector Architecture v2 — Design Document

Status: approved (2026-07-02). Program PRD: `docs/plans/universal-programming-loop-prd.md`.
Validated against `connectors.go`, `defs/*`, `engine/*`, `connsdk/*`, `hooks/*`, `native/*`,
`conformance/*`, `internal/app/sync_modes.go`, and `internal/app/types.go`.

## 0. Executive summary

1. **Connectors become data, not code.** All 556 connector packages under
   `internal/connectors/<name>/` are replaced by JSON definition bundles under
   `internal/connectors/defs/<name>/` (split files, Ruby-style), embedded once via a single
   `go:embed` in `internal/connectors/defs`. A new, well-tested **declarative engine**
   (`internal/connectors/engine`) interprets the bundles at registry build time. A connector like
   aircall goes from 723 lines of Go to **zero Go**.
2. **Escape hatches are additive, never replacements.** Custom behavior attaches as *hooks*
   (`internal/connectors/hooks/<name>/`) or, for non-HTTP systems (databases, queues, files), as
   *full custom connectors* (`internal/connectors/native/<name>/`) that still carry a JSON defs
   bundle for identity/schemas/catalog.
3. **The catalog is the registry.** `catalog_data.json` (646 Airbyte-derived entries), `slug.go`,
   `CatalogAliasConnector`, `NativeCatalogConnector`, `RegisterNativeLive`, and the live/non-live
   distinction are all deleted. The catalog is computed from the defs tree — one entry per
   connector, bare names only, real capability data.
4. **Capability-first-run is enforced by `api_surface.json`** — a per-connector coverage manifest
   mapping every documented API endpoint to a stream, a write action, or an approved exclusion.
   Conformance v2 fails a connector whose API has write endpoints with no write actions declared
   and no exclusion waiver.
5. **Conformance v2 executes the real engine against recorded fixture pages** (embedded
   `fixtures/`), replacing the synthetic `mode=fixture` records that today bypass all real
   request/pagination/cursor logic.

## A. Per-connector directory layout

### Decision: split files, not one manifest.json

Follow the Ruby pattern (`connection_specification.json` / `metadata.json` / `schemas/*.json`):
agent-readability (a 60-line schema file, not a 4,000-line manifest), diff hygiene (one stream =
one file; parallel authoring doesn't conflict), concern separation matching the runtime
(`spec.json` at connection-setup time, `streams.json` at read time, `writes.json` at reverse-ETL
time, `api_surface.json` only by conformance). One deviation from Ruby: request/pagination/cursor
config is **not** code — it is `streams.json`, interpreted by the engine.

### Layout

```
internal/connectors/defs/
  defs.go                     // package defs; //go:embed all */** ; exposes FS
  github/
    metadata.json             // identity, capabilities, rate limits, risk
    spec.json                 // connection specification (JSON Schema draft-07)
    streams.json              // declarative read config: base HTTP + streams
    writes.json               // declarative write actions
    api_surface.json          // API coverage manifest (conformance input)
    schemas/
      issues.json             // per-stream record schema (draft-07 + x- extensions)
      pull_requests.json
    fixtures/
      streams/issues/page_1.json, page_2.json    // recorded API pages
      writes/create_issue.json                   // sample record + expected request
    docs.md                   // human/agent guide
```

`defs/defs.go` is the only Go file:

```go
// Package defs embeds every connector definition bundle.
package defs

import "embed"

//go:embed */metadata.json */spec.json */streams.json */writes.json */api_surface.json */schemas/* */fixtures/** */docs.md
var FS embed.FS
```

(`writes.json` and `fixtures/` are optional per connector; the loader tolerates absence.
Directory name = connector name = the one true identifier: `github`, not `source-github`.)

### `metadata.json` (github example)

```json
{
  "name": "github",
  "display_name": "GitHub",
  "description": "Reads GitHub repository, issue, PR, release, and Actions data; writes issues, PRs, labels, milestones, releases, files, and workflow controls.",
  "integration_type": "api",
  "docs_url": "https://docs.github.com/en/rest",
  "release_stage": "ga",
  "capabilities": { "check": true, "read": true, "write": true, "query": false, "cdc": false, "dynamic_schema": false },
  "batch": { "read_page_size": 100, "write_batch_size": 1 },
  "rate_limit": { "strategy": "retry_after_header", "requests_per_hour": 5000 },
  "risk": {
    "read": "read-only REST calls against the configured repository",
    "write": "creates and mutates issues, PRs, labels, milestones, releases, files, workflow runs",
    "approval": "reverse ETL writes require plan preview and approval token"
  }
}
```

Note: `catalog` is no longer a capability — every connector has a catalog by construction (the
schemas dir). Sync modes are **not** listed here; they are derived per stream (§B.6).

### `spec.json` — pure JSON Schema draft-07

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "GitHub Connection Specification",
  "type": "object",
  "required": ["repository"],
  "properties": {
    "repository": { "type": "string", "description": "owner/repo", "pattern": "^[^/]+/[^/]+$" },
    "base_url":   { "type": "string", "format": "uri", "default": "https://api.github.com" },
    "auth_type":  { "type": "string", "enum": ["auto", "public", "token", "github_app"], "default": "auto" },
    "app_id":          { "type": "string" },
    "installation_id": { "type": "string" },
    "token":       { "type": "string", "x-secret": true },
    "private_key": { "type": "string", "x-secret": true },
    "start_date":  { "type": "string", "format": "date-time" }
  }
}
```

`x-secret: true` is the single source of truth for the config/secret split (replaces
`ConfigField`/`SecretField` Go structs and the catalog's `secret_fields`). The loader partitions
properties into config vs secrets from this flag.

### `schemas/issues.json` — per-stream record schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "issues",
  "type": "object",
  "x-primary-key": ["id"],
  "x-cursor-field": "updated_at",
  "properties": {
    "id":         { "type": "integer" },
    "number":     { "type": "integer" },
    "repository": { "type": "string" },
    "title":      { "type": "string" },
    "state":      { "type": ["string", "null"] },
    "user_login": { "type": ["string", "null"] },
    "labels":     { "type": "array", "items": { "type": "string" } },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "closed_at":  { "type": ["string", "null"], "format": "date-time" }
  }
}
```

`x-primary-key` / `x-cursor-field` are the Ruby extensions. The schema doubles as the **record
projection**: by default the engine emits only declared properties (today's hand-written
`mapRecord` functions become this projection), unless the stream sets
`"projection": "passthrough"`.

### `streams.json` (github, abridged)

```json
{
  "base": {
    "url": "{{ config.base_url }}",
    "user_agent": "polymetrics-go-cli",
    "headers": { "Accept": "application/vnd.github+json", "X-GitHub-Api-Version": "2026-03-10" },
    "auth": [
      { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ config.auth_type in ['auto','token'] }}" },
      { "mode": "custom", "hook": "github_app", "when": "{{ config.auth_type == 'github_app' }}" },
      { "mode": "none",   "when": "{{ config.auth_type == 'public' }}" }
    ],
    "pagination": { "type": "link_header", "size_param": "per_page" },
    "check": { "method": "GET", "path": "/repos/{{ config.repository }}" },
    "error_map": [
      { "status": 401, "hint": "token is missing or expired; re-run pm credentials add github" },
      { "status": 403, "match_body": "rate limit", "class": "rate_limited" },
      { "status": 404, "hint": "repository not found or token lacks access" }
    ]
  },
  "streams": [
    {
      "name": "issues",
      "path": "/repos/{{ config.repository }}/issues",
      "records": { "path": ".", "filter": { "field_absent": "pull_request" } },
      "computed_fields": { "repository": "{{ config.repository }}", "user_login": "{{ record.user.login }}" },
      "incremental": { "cursor_field": "updated_at", "request_param": "since", "param_format": "rfc3339", "start_config_key": "start_date" },
      "query": { "state": "all", "sort": "updated", "direction": "asc" },
      "schema": "schemas/issues.json"
    },
    {
      "name": "workflow_runs",
      "path": "/repos/{{ config.repository }}/actions/runs",
      "records": { "path": "workflow_runs" },
      "incremental": { "cursor_field": "updated_at", "request_param": "created", "param_format": "github_date_range" },
      "schema": "schemas/workflow_runs.json"
    },
    {
      "name": "repository",
      "path": "/repos/{{ config.repository }}",
      "records": { "path": ".", "single_object": true },
      "pagination": { "type": "none" },
      "schema": "schemas/repository.json"
    }
  ]
}
```

Pagination types (all four already exist in `connsdk/paginate.go`): `link_header`, `page_number`,
`offset_limit`, `cursor` (body token), plus `next_url` (aircall's `meta.next_page_link`, a cursor
variant where the token is a full URL) and `none`.

### `writes.json` (github, abridged)

```json
{
  "actions": [
    {
      "name": "create_issue",
      "kind": "create",
      "method": "POST",
      "path": "/repos/{{ config.repository }}/issues",
      "record_schema": {
        "type": "object",
        "required": ["title"],
        "properties": {
          "title": { "type": "string" }, "body": { "type": "string" },
          "labels": { "type": "array", "items": { "type": "string" } },
          "assignees": { "type": "array", "items": { "type": "string" } },
          "milestone": { "type": "integer" }
        }
      },
      "risk": "creates a visible issue in the repository"
    },
    {
      "name": "update_issue",
      "kind": "update",
      "method": "PATCH",
      "path": "/repos/{{ config.repository }}/issues/{{ record.issue_number }}",
      "path_fields": ["issue_number"],
      "record_schema": {
        "type": "object", "required": ["issue_number"], "minProperties": 2,
        "properties": {
          "issue_number": { "type": "integer" }, "title": { "type": "string" },
          "body": { "type": "string" }, "state": { "type": "string", "enum": ["open", "closed"] }
        }
      },
      "risk": "mutates an existing issue"
    },
    {
      "name": "delete_label",
      "kind": "delete",
      "method": "DELETE",
      "path": "/repos/{{ config.repository }}/labels/{{ record.name }}",
      "path_fields": ["name"],
      "record_schema": { "type": "object", "required": ["name"], "properties": { "name": { "type": "string" } } },
      "delete": { "idempotent": true, "missing_ok_status": [404] },
      "risk": "permanently deletes a label",
      "confirm": "destructive"
    },
    {
      "name": "merge_pull_request",
      "kind": "custom",
      "method": "PUT",
      "path": "/repos/{{ config.repository }}/pulls/{{ record.pull_number }}/merge",
      "path_fields": ["pull_number"],
      "record_schema": {
        "type": "object", "required": ["pull_number"],
        "properties": {
          "pull_number": { "type": "integer" },
          "merge_method": { "type": "string", "enum": ["merge", "squash", "rebase"] },
          "commit_title": { "type": "string" }
        }
      },
      "risk": "merges a pull request into its base branch",
      "confirm": "destructive"
    }
  ]
}
```

Write semantics baked into the format:

- **Body construction rule (default)**: every record field not consumed by the path
  (`path_fields`) becomes the JSON body. `body_type: "form"` switches to
  `x-www-form-urlencoded` (stripe). `body: {template}` overrides for reshaped payloads.
  `kind: delete` sends no body unless `body_fields` listed (github `delete_file` needs
  `message`/`sha` — declared via `"body_fields": ["message", "sha", "branch"]`).
- **`record_schema` is draft-07**, validated by the same compiled-schema machinery as `spec.json`.
  This replaces github's hand-written validate/payload functions (~700 lines) for the ~90% of
  actions that are plain payload mapping. Multi-request compound actions (github's
  `create_pull_request` + reviewer follow-up) use a **write hook** (§B.7).
- **`confirm: "destructive"`** feeds the existing plan → preview → approve flow with per-action
  risk tiering.

## B. The declarative runtime engine

### B.1 Package layout

```
internal/connectors/
  connectors.go          // core interfaces, Registry (heavily slimmed)
  connsdk/               // UNCHANGED low-level toolkit: Requester, Authenticator, Paginator, extract, state
  engine/                // NEW: interprets defs bundles
    bundle.go            // Bundle types + loader + validation
    interpolate.go       // {{ ... }} resolver
    connector.go         // engine.Connector implementing connectors.Connector (+Writer, StatefulReader)
    read.go              // stream execution: request build, paginate, extract, project, cursor
    write.go             // action execution: validate, dry-run, execute, delete semantics
    auth.go              // AuthSpec -> connsdk.Authenticator selection
    hooks.go             // hook interfaces + hook registry
    schema.go            // draft-07 compiler (minimal internal impl; no new deps) + x- extensions
    errors.go            // error_map application, typed engine errors
  defs/                  // NEW: embedded JSON bundles (556 dirs)
  hooks/                 // NEW: per-connector Go hooks, only where needed (~15 dirs)
    github/hooks.go
  native/                // full custom connectors (non-HTTP): postgres, warehouse, file, sample, outbox, ...
  conformance/           // conformance v2
cmd/
  connectorgen/          // replaces registrygen + pm-cataloggen: validate | gen | new
```

The engine builds **on top of** connsdk (verified as the right substrate: `Requester.Do/DoForm/
DoJSON` with retry/Retry-After, four paginators, `RecordsAt`/`StringAt`, `Cursor`/`WithCursor`/
`MaxCursor`). connsdk keeps zero JSON-awareness; the engine is the only interpreter.

### B.2 Core Go types

```go
// engine/bundle.go
type Bundle struct {
    Name     string
    Metadata Metadata          // parsed metadata.json
    Spec     *Schema           // compiled spec.json; SecretKeys() from x-secret
    HTTP     HTTPBase          // streams.json "base"
    Streams  []StreamSpec      // streams.json "streams"
    Writes   []WriteAction     // writes.json (nil if absent)
    Schemas  map[string]*StreamSchema // stream name -> compiled schema + PK/cursor
    Surface  *APISurface       // api_surface.json (conformance only)
    Docs     string            // docs.md
    Fixtures fs.FS             // fixtures/ subtree
}

type HTTPBase struct {
    URL        string            `json:"url"`
    UserAgent  string            `json:"user_agent"`
    Headers    map[string]string `json:"headers"`
    Auth       []AuthSpec        `json:"auth"`
    Pagination *PaginationSpec   `json:"pagination"`
    Check      *RequestSpec      `json:"check"`
    ErrorMap   []ErrorRule       `json:"error_map"`
    RateLimit  *RateLimitSpec    `json:"rate_limit,omitempty"`
}

type AuthSpec struct {
    Mode   string `json:"mode"`   // none|bearer|basic|api_key_header|api_key_query|oauth2_client_credentials|custom
    Token  string `json:"token,omitempty"`
    Username, Password string     // basic
    Header, Prefix, Param, Value string // api_key_*
    TokenURL, ClientID, ClientSecret, Scopes string // oauth2_client_credentials
    Hook   string `json:"hook,omitempty"` // custom: hook name resolved via hooks registry
    When   string `json:"when,omitempty"` // condition over config values
}

type StreamSpec struct {
    Name           string            `json:"name"`
    Method         string            `json:"method,omitempty"` // default GET
    Path           string            `json:"path"`
    Query          map[string]string `json:"query,omitempty"`
    Body           map[string]any    `json:"body,omitempty"`   // POST-body streams (GraphQL/search)
    Records        RecordsSpec       `json:"records"`
    Pagination     *PaginationSpec   `json:"pagination,omitempty"` // overrides base
    Incremental    *IncrementalSpec  `json:"incremental,omitempty"`
    ComputedFields map[string]string `json:"computed_fields,omitempty"`
    Projection     string            `json:"projection,omitempty"` // "schema" (default) | "passthrough"
    SchemaRef      string            `json:"schema"`
}

type RecordsSpec struct {
    Path         string      `json:"path"`          // dotted path; "." = body root
    SingleObject bool        `json:"single_object,omitempty"`
    Filter       *FilterSpec `json:"filter,omitempty"` // field_absent / field_equals
}

type IncrementalSpec struct {
    CursorField    string `json:"cursor_field"`
    RequestParam   string `json:"request_param,omitempty"`   // server-side lower bound ("since")
    ParamFormat    string `json:"param_format,omitempty"`    // rfc3339|unix_seconds|date|github_date_range
    StartConfigKey string `json:"start_config_key,omitempty"`
    ClientFiltered bool   `json:"client_filtered,omitempty"` // API has no filter; engine drops old records
}

type WriteAction struct {
    Name         string          `json:"name"`
    Kind         string          `json:"kind"` // create|update|upsert|delete|custom
    Method, Path string
    PathFields   []string        `json:"path_fields,omitempty"`
    BodyType     string          `json:"body_type,omitempty"` // json (default) | form | none
    BodyFields   []string        `json:"body_fields,omitempty"`
    RecordSchema json.RawMessage `json:"record_schema"`
    Delete       *DeleteSpec     `json:"delete,omitempty"` // idempotent, missing_ok_status
    Risk         string          `json:"risk"`
    Confirm      string          `json:"confirm,omitempty"` // "" | "destructive"
    Hook         string          `json:"hook,omitempty"`    // custom executor
}
```

### B.3 Interpolation — deliberately tiny

`{{ config.key }}`, `{{ secrets.key }}`, `{{ record.dotted.path }}`, `{{ cursor }}` plus a fixed
filter set: `urlencode` (default for path segments), `unix_seconds`, `base64`. `when` conditions
support only `==`, `in [...]`, and truthiness. No loops, no arithmetic, no user-defined functions —
anything beyond this is a hook. The interpolator is ~150 lines with table-driven tests, and every
`{{ }}` in every bundle is resolved and type-checked at `connectorgen validate` time against
`spec.json`.

### B.4 Read path (engine/read.go)

```go
func (c *Connector) Read(ctx context.Context, req connectors.ReadRequest, emit func(connectors.Record) error) error
```

1. Resolve `StreamSpec` by `req.Stream`; build `connsdk.Requester` (base URL, headers, auth via
   `selectAuth(cfg)`).
2. Build initial query: static `query` + incremental lower bound from `req.State["cursor"]`
   (fallback `start_config_key`) formatted per `param_format`.
3. Drive the paginator loop; per page: `RecordsAt(body, records.path)` → filter → project through
   stream schema (+ `computed_fields`, which can reach into nested raw JSON, e.g.
   `user_login: {{ record.user.login }}`) → track `MaxCursor` → `emit`.
4. On completion the app layer persists the advanced cursor exactly as today (`internal/app`
   streaming ETL unchanged; `StatefulReader.InitialState` implemented generically by the engine).

Rate limiting: `RateLimitSpec{requests_per_minute}` adds a token-bucket wait inside the requester
loop; Retry-After handling already exists in connsdk.

### B.5 Write path (engine/write.go)

`ValidateWrite` = compile-once `record_schema` validation per record (structural errors carry
record index, matching current behavior). `DryRunWrite` = validation + fully-resolved request
preview (`WritePreview.Warnings` includes resolved method/path with secrets redacted). `Write` =
per-record execution; `kind: delete` honors `missing_ok_status` (a 404 on an idempotent delete
counts as written, not failed). Batch semantics stay one-request-per-record (matches github/stripe
today); `metadata.json.batch.write_batch_size` reserved for future bulk endpoints.

### B.6 Sync modes — derived, never declared

Per stream: `full_refresh_append` and `full_refresh_overwrite` always; `*_deduped` variants iff
`x-primary-key` present; `incremental_append[_deduped]` iff `incremental` block present.
`internal/app/sync_modes.go` unchanged; `ValidateStreamSyncConfig` additionally consults the
stream's derived mode set. This kills the drift between `Manifest.SyncModes`,
`catalog_data.json.supported_sync_modes`, and reality.

### B.7 Escape hatches — when a connector still needs Go

Three tiers, strictly ordered; conformance rejects Go where JSON suffices (a hook must name which
hook-point it needs).

**Tier 1 — declarative only (target ≥ 90% of 556):** everything expressible above.

**Tier 2 — bundle + hooks (`internal/connectors/hooks/<name>/hooks.go`, target ~8%):** the bundle
still defines all streams/writes/schemas; Go attaches at named points:

```go
// engine/hooks.go
type Hooks interface{ ConnectorName() string }

type AuthHook interface {        // e.g. github_app JWT->installation-token exchange, AWS SigV4
    Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec engine.AuthSpec) (connsdk.Authenticator, error)
}
type RecordHook interface {      // per-record post-processing beyond projection
    MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error)
}
type StreamHook interface {      // whole-stream override (async report jobs, CSV downloads, sub-resource fan-out)
    ReadStream(ctx context.Context, stream engine.StreamSpec, req connectors.ReadRequest, rt *engine.Runtime, emit func(connectors.Record) error) (handled bool, err error)
}
type WriteHook interface {       // compound/multi-request actions (github create_pull_request + reviewers)
    ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (handled bool, err error)
}
type CheckHook interface {
    Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *engine.Runtime) (handled bool, err error)
}

func Register(name string, factory func() Hooks) // called from hooks/<name>/init(); wired by generated hooks/hookset
```

Legitimate Tier-2 triggers: signature auth (SigV4, HMAC), token-exchange auth (GitHub App),
multipart/XML bodies, async report polling, response decompression/CSV, sub-resource fan-out reads
(issue → comments per issue), compound writes.

**Tier 3 — full custom connector (`internal/connectors/native/<name>/`, ~10 connectors):**
non-REST protocols — postgres/mysql/snowflake/bigquery (SQL + CDC), amazon-sqs,
file/warehouse/outbox/sample built-ins. These implement `connectors.Connector` directly with the
Ruby component split as the mandated file layout: `connector.go` (entry + registration),
`connection.go`, `reader.go`, `cataloger.go`, `writer.go`, `cdc.go`. **They still ship a defs
bundle** (metadata.json, spec.json, schemas/) so identity, catalog, and docs stay uniform; they
embed `engine.Base` which serves `Definition()`/`Catalog()` from the bundle.

## C. Interface, registry, and catalog changes

### C.1 Connector interface evolution (`internal/connectors/connectors.go`)

```go
type Connector interface {
    Name() string
    Definition() Definition                    // replaces Metadata() + ManifestProvider
    Check(ctx context.Context, cfg RuntimeConfig) error
    Catalog(ctx context.Context, cfg RuntimeConfig) (Catalog, error)
    Read(ctx context.Context, req ReadRequest, emit func(Record) error) error
}

// Write moves OUT of the core interface: ~530 read-only connectors no longer stub it.
type Writer interface {
    ValidateWrite(ctx context.Context, req WriteRequest, records []Record) error
    DryRunWrite(ctx context.Context, req WriteRequest, records []Record) (WritePreview, error)
    Write(ctx context.Context, req WriteRequest, records []Record) (WriteResult, error)
}

// Unchanged optional interfaces: Querier, CDCReader, StatefulReader, LiveConformanceProvider.
// Deleted: ManifestProvider, SchemaMapper (unused).
```

`Definition` is the single replacement for the current triple `Metadata` + `Manifest` +
`ConnectorDefinition`:

```go
type Definition struct {
    Name            string            `json:"name"`
    DisplayName     string            `json:"display_name"`
    Description     string            `json:"description"`
    IntegrationType string            `json:"integration_type"`
    DocsURL         string            `json:"docs_url"`
    ReleaseStage    string            `json:"release_stage"`
    Capabilities    Capabilities      `json:"capabilities"`
    Spec            json.RawMessage   `json:"spec"`
    Streams         []StreamSummary   `json:"streams"`
    WriteActions    []WriteActionInfo `json:"write_actions,omitempty"`
    Risk            RiskSpec          `json:"risk"`
    Icon            *ConnectorIcon    `json:"icon,omitempty"`
}
```

### C.2 Registry

```go
func NewRegistry() *Registry {
    r := &Registry{connectors: map[string]Connector{}}
    bundles, err := engine.LoadAll(defs.FS)          // parse+validate once, cached
    if err != nil { panic(err) }                     // build-time guaranteed by connectorgen validate + tests
    for _, b := range bundles {
        r.Register(engine.New(b, hooks.For(b.Name)))
    }
    nativeset.RegisterInto(r)                        // Tier-3 natives override/extend explicitly
    return r
}
```

Deleted from `connectors.go`: `NewLiveRegistry`, `isLiveFactory`, `liveFactoryNamesCache`,
`RegisterNativeLive`, `nativeLiveNames`, the enabled-catalog-alias loop, `CatalogAliasConnector`.
**Every loaded connector is live** — the enablement ladder collapses because conformance v2 gates
merge, not runtime exposure. Tier-3 natives are wired explicitly by `native/nativeset`; the legacy
process-global `RegisterFactory` path is gone.

### C.3 Generators

- `cmd/connectorgen` owns bundle validation and generated hook/native wiring:
  - `validate` — loads every bundle: draft-07-compiles all schemas, resolves every `{{ }}` against
    spec properties, checks `schema` refs exist, PK fields exist in schema, cursor fields exist,
    write `path_fields ⊆ record_schema.required∪properties`, api_surface coverage (§E). Run in CI
    and as a plain Go test (`engine.LoadAll` in a test gives the same guarantee).
  - `gen` — regenerates the two tiny wiring files: `hooks/hookset/hookset_gen.go` (blank imports
    for `hooks/*`, ~15 lines) and `native/nativeset/nativeset_gen.go` (~10 lines).
  - `new <name>` — scaffolds a defs bundle from templates.
- `cmd/pm-cataloggen` (Airbyte importer) — **deleted**.
- `cmd/iconregistrygen` — unchanged; keys become bare names.

### C.4 Catalog: generated from connectors, and the 646 vs 556 divergence

**Decision: there is no catalog file.** `catalog_data.json` is deleted; the catalog is a view over
loaded definitions: `func (r *Registry) CatalogEntries() []Definition`. An entry exists iff a defs
bundle (or native package) exists. The ~90 Airbyte entries with no implementation and the
source-/destination- duplicate pairs vanish; a one-time `docs/connectors/UNPORTED.md` snapshot
preserves the dropped list. Static consumers get `connectorgen gen --catalog-json
build/catalog.json` — a build artifact, never a source of truth. `ConnectorDefinition`,
`ImplementationStatus`, `RuntimeKind`, `RuntimeCapabilities`, `PMConnectorName`,
`NativeCatalogConnector`, `native_port.go`: all deleted. Only `release_stage` remains as a quality
signal, sourced from metadata.json.

## D. Naming clean-break plan

Everything keys on the bare name (`github`, `stripe`, `postgres`). No aliases, no compatibility
parsing.

| Item | Action |
|---|---|
| `internal/connectors/slug.go` + `slug_test.go` | delete |
| `catalog_data.json`, `ConnectorDefinitionBySlug`, `ConnectorCatalog()` | delete; replace with `Registry.CatalogEntries()` |
| `CatalogAliasConnector`, `NativeCatalogConnector`, `native_catalog_connector.go`, `native_port.go` | delete |
| `RegisterNativeLive`, `NewLiveRegistry`, live-name cache | delete |
| legacy generated registry package/command | delete; replaced by hookset/nativeset + `cmd/connectorgen` |
| `cmd/pm-cataloggen` | delete |
| `Manifest`/`ManifestProvider`/`ConfigField`/`SecretField`/`AuthModeSpec`/`PaginationSpec`/`WriteActionSpec` in `manifest.go` | delete; `Definition` replaces |
| CLI: any `source-`/`destination-` acceptance, `--type source|destination` filters | bare names only; direction filters become `--capability read|write|cdc|query` |
| `internal/app/types.go` connector name values | bare names; no legacy parsing; clear migration error for `source-*` |
| Saved sync state | untouched — `streamStateKey` is `connection:stream`, connector-name-free (verified) |
| Docs: porting guide, status policy, guide.go legacy sections | rewrite for bundle authoring |

Direction is gone as a concept: a connector's "direction" is just `capabilities.read` / presence of
write actions.

## E. Capability-on-first-run policy and conformance v2

### E.1 `api_surface.json`

```json
{
  "api": "GitHub REST API v3",
  "docs": "https://docs.github.com/en/rest",
  "reviewed_at": "2026-07-01",
  "scope": "repository-scoped endpoints; org- and enterprise-admin endpoints out of scope",
  "endpoints": [
    { "method": "GET",    "path": "/repos/{owner}/{repo}/issues",           "covered_by": { "stream": "issues" } },
    { "method": "POST",   "path": "/repos/{owner}/{repo}/issues",           "covered_by": { "write": "create_issue" } },
    { "method": "PATCH",  "path": "/repos/{owner}/{repo}/issues/{number}",  "covered_by": { "write": "update_issue" } },
    { "method": "GET",    "path": "/repos/{owner}/{repo}/traffic/views",
      "excluded": { "category": "requires_elevated_scope", "reason": "push-access-only traffic API; niche analytics" } },
    { "method": "DELETE", "path": "/repos/{owner}/{repo}",
      "excluded": { "category": "destructive_admin", "reason": "repository deletion is never a reverse-ETL action" } }
  ]
}
```

Rules (enforced by `connectorgen validate` + conformance):

1. Every endpoint entry has exactly one of `covered_by` or `excluded`.
2. `covered_by.stream`/`covered_by.write` must resolve to a declared stream/action — and vice
   versa: every stream and write action must appear in the surface.
3. `excluded.category` from the closed vocabulary: `destructive_admin`,
   `requires_elevated_scope`, `binary_payload`, `deprecated`, `non_data_endpoint`,
   `duplicate_of`, `out_of_scope` (with `scope` prose justifying it).
4. **Fail-first-run**: `capabilities.write == false` is only legal when the surface contains zero
   non-excluded POST/PUT/PATCH/DELETE endpoints. Same rule for GET endpoints vs streams.
5. Freshness: `reviewed_at` older than 12 months → warning (not failure).

### E.2 Conformance v2 (`internal/connectors/conformance/`)

Static (per bundle): `spec_schema_valid`, `stream_schemas_valid`, `pk_fields_exist`,
`cursor_fields_exist`, `interpolations_resolve`, `write_schemas_valid`, `surface_complete`,
`docs_present`, `secret_redaction`, `fixtures_present`.

Dynamic (per bundle, fixture-backed): an `httptest.Server` replays
`fixtures/streams/<stream>/page_N.json` (recorded real API pages, keyed by expected request
path+query); the **real engine** runs against it:
- `check_fixture`, `read_fixture_nonempty` per stream with fixtures (first stream mandatory),
- `pagination_terminates` (multi-page fixture consumed exactly once, no infinite loop),
- `records_match_schema` (every emitted record validates against the stream schema),
- `cursor_advances` (incremental: cursor after read == max cursor in fixtures; re-read with that
  cursor sends the right `request_param`),
- `write_validate` + `write_request_shape` per action: `fixtures/writes/<action>.json` =
  `{"record": {...}, "expect": {"method","path","body"}}`; engine dry-run must produce the
  expected request; invalid-record cases must fail validation,
- `delete_semantics` for `kind:delete` actions (404 handling per `missing_ok_status`).

Live (opt-in): `LiveConformanceProvider` supplies real credentials for a nightly job.

**Merge gate**: a bundle merges only when all static + dynamic tests pass. Existence in `defs/`
means enabled.

## F. Code quality conventions

1. **File layout** — declarative connectors: exactly the defs bundle, zero Go. Hook packages:
   single `hooks.go` (+ `hooks_test.go`), hard cap ~300 lines; past that → Tier 3. Native
   connectors: mandated component split, each file < ~400 lines.
2. **Max-custom-code guidance**: `connectorgen validate` reports Go LOC per connector; flags any
   hook package > 300 lines or > 2 hook interfaces.
3. **Naming**: connector name = dir name = registry key = `^[a-z0-9][a-z0-9-]*$`. Streams and
   write actions `snake_case`; actions verb-first (`create_issue`, `merge_pull_request`).
4. **Error handling**: engine errors wrap `connsdk.HTTPError` and attach
   `{connector, stream|action, page|record_index}` via a typed `engine.Error`; `error_map` hints
   surface to the CLI verbatim. Secrets never in errors (`safety.RedactErrorText` stays).
5. **Tests**: table-driven throughout. Engine unit tests own all generic behavior; per-connector
   "tests" are the fixtures themselves. Hook packages require table-driven Go tests against
   `httptest`.
6. **Docs**: `docs.md` per connector with fixed headings (Overview, Auth setup, Streams notes,
   Write actions & risks, Known limits).

## G. Migration diff summary

**Added**: `engine/` (~2,500 lines incl. minimal draft-07 validator), `defs/` (556 bundles),
`hooks/` + hookset (~15 pkgs), `native/` + nativeset, `conformance/`, `cmd/connectorgen`,
`api_surface.json` + `fixtures/` per connector.

**Changed**: `connectors.go` (Definition(), Writer extraction, bundle registry), `native/` moves,
`internal/cli` (bare names, capability filters), `internal/app` (derived sync-mode validation),
porting guide rewrite.

**Removed**: 556 connector Go packages (~250k+ lines), `catalog_data.json`, legacy catalog types,
legacy name parsing/catalog wrappers, v1 native conformance, legacy generated registry wiring,
legacy catalog importer, live-registry split, and the `ImplementationStatus` ladder.

**Sequencing** (each phase leaves the tree green): (1) engine + loader + conformance v2 alongside
legacy; (2) reference ports github/stripe/aircall with engine-vs-legacy parity tests; (3) bulk
migration, conformance warn mode; (4) registry flip + legacy deletion + naming clean break in one
PR; (5) capability completion per connector, surface rules warn→fail per connector as authored.
