# API-CONTRACT — wave0-engine-harness (Go interfaces added/changed)

Rule of the phase: **additive only**. The existing `connectors.Connector` interface
(`internal/connectors/connectors.go:256`), `Manifest`/`ManifestProvider`
(`internal/connectors/manifest.go:52,66`), registry constructors, and all optional interfaces
(`WriteValidator`, `DryRunWriter`, `Querier`, `CDCReader`, `StatefulReader`,
`LiveConformanceProvider`) are UNCHANGED. The design-doc §C.1 interface evolution (Definition
replacing Metadata, Writer extraction) is a wave6 change.

## 1. package `connectors` — additions (new file `internal/connectors/definition.go`)

```go
// Definition is the unified connector descriptor introduced by architecture v2.
// In wave0 it coexists with Metadata/Manifest; in wave6 it replaces them.
type Definition struct {
    Name            string            `json:"name"`
    DisplayName     string            `json:"display_name"`
    Description     string            `json:"description"`
    IntegrationType string            `json:"integration_type"`
    DocsURL         string            `json:"docs_url"`
    ReleaseStage    string            `json:"release_stage"`
    Capabilities    Capabilities      `json:"capabilities"` // existing struct, connectors.go:108
    Spec            json.RawMessage   `json:"spec"`         // verbatim spec.json
    Streams         []StreamSummary   `json:"streams"`
    WriteActions    []WriteActionInfo `json:"write_actions,omitempty"`
    Risk            RiskSpec          `json:"risk"`         // existing struct, manifest.go:26
    Icon            *ConnectorIcon    `json:"icon,omitempty"`
}

type StreamSummary struct {
    Name        string   `json:"name"`
    Description string   `json:"description,omitempty"`
    PrimaryKey  []string `json:"primary_key,omitempty"`
    CursorField string   `json:"cursor_field,omitempty"`
    SyncModes   []string `json:"sync_modes"` // DERIVED (design §B.6), never authored
}

type WriteActionInfo struct {
    Name    string `json:"name"`
    Kind    string `json:"kind"`
    Method  string `json:"method"`
    Path    string `json:"path"`
    Risk    string `json:"risk"`
    Confirm string `json:"confirm,omitempty"`
}

// DefinitionProvider is implemented by engine-backed and Tier-3 connectors in
// wave0; the method joins the core Connector interface in wave6.
type DefinitionProvider interface {
    Definition() Definition
}
```

## 2. package `engine` (new: `internal/connectors/engine`)

```go
// bundle.go
func LoadAll(fsys fs.FS) ([]Bundle, error)
func Load(fsys fs.FS, name string) (Bundle, error)

type Bundle struct {
    Name     string
    Metadata Metadata
    Spec     *Schema                  // compiled spec.json; SecretKeys() from x-secret
    HTTP     HTTPBase                 // streams.json "base" (zero value when no streams.json)
    Streams  []StreamSpec
    Writes   []WriteAction            // nil when writes.json absent
    Schemas  map[string]*StreamSchema // stream name -> compiled schema + PK/cursor
    Surface  *APISurface
    Docs     string
    Fixtures fs.FS                    // fixtures/ subtree; nil when absent
}
// Metadata, HTTPBase, AuthSpec, StreamSpec, RecordsSpec, FilterSpec, IncrementalSpec,
// WriteAction, DeleteSpec, ErrorRule, RateLimitSpec: field-for-field per design §B.2;
// PaginationSpec extended with LastRecordField/StopPath/NextURLPath/MaxPages (DATA-MODEL §2).

// schema.go
type Schema struct{ /* opaque compiled form */ }
func CompileSchema(raw json.RawMessage) (*Schema, error)
func (s *Schema) Validate(v any) error         // instance validation
func (s *Schema) SecretKeys() []string         // x-secret properties
func (s *Schema) Properties() []string
type StreamSchema struct {
    *Schema
    PrimaryKey  []string // x-primary-key
    CursorField string   // x-cursor-field
}

// interpolate.go
type Vars struct {
    Config  map[string]string
    Secrets map[string]string
    Record  map[string]any // nil outside record contexts
    Cursor  string
}
func Interpolate(template string, vars Vars) (string, error)           // urlencode default off
func InterpolatePath(template string, vars Vars) (string, error)       // urlencode default on
func EvalWhen(cond string, vars Vars) (bool, error)
func ResolveCheck(template string, specKeys map[string]bool) error     // connectorgen-time

// auth.go
func selectAuth(cfg connectors.RuntimeConfig, specs []AuthSpec, h Hooks) (connsdk.Authenticator, error)

// paginate.go
func newPaginator(spec PaginationSpec, pageSize int) (connsdk.Paginator, error)

// errors.go
type Error struct {
    Connector, Stream, Action string
    Page, RecordIndex         int    // -1 when not applicable
    Class, Hint               string // from error_map
    Err                       error  // wraps *connsdk.HTTPError etc.
}
func (e *Error) Error() string       // redacted via safety.RedactErrorText
func (e *Error) Unwrap() error
func applyErrorMap(rules []ErrorRule, err error) (class, hint string)

// hooks.go — interfaces exactly per design §B.7
type Hooks interface{ ConnectorName() string }
type AuthHook interface {
    Authenticator(ctx context.Context, cfg connectors.RuntimeConfig, spec AuthSpec) (connsdk.Authenticator, error)
}
type RecordHook interface {
    MapRecord(stream string, raw, projected connsdk.Record) (connsdk.Record, bool, error)
}
type StreamHook interface {
    ReadStream(ctx context.Context, stream StreamSpec, req connectors.ReadRequest, rt *Runtime, emit func(connectors.Record) error) (handled bool, err error)
}
type WriteHook interface {
    ExecuteWrite(ctx context.Context, action WriteAction, rec connectors.Record, rt *Runtime) (handled bool, err error)
}
type CheckHook interface {
    Check(ctx context.Context, cfg connectors.RuntimeConfig, rt *Runtime) (handled bool, err error)
}
func RegisterHooks(name string, factory func() Hooks)  // called from hooks/<name>/init()
func HooksFor(name string) Hooks                       // nil when none registered

// connector.go
func New(b Bundle, h Hooks) *Connector
type Connector struct{ /* unexported */ }
// Satisfies (compile-asserted in tests):
//   connectors.Connector            (Name, Metadata, Check, Catalog, Read, Write)
//   connectors.WriteValidator, connectors.DryRunWriter
//   connectors.StatefulReader
//   connectors.ManifestProvider     (manifest synthesized from bundle)
//   connectors.DefinitionProvider
type Runtime struct { // passed to Stream/Write/Check hooks
    Requester *connsdk.Requester
    Bundle    *Bundle
    Config    connectors.RuntimeConfig
}

// Base for Tier-3 natives (native/postgres embeds it):
type Base struct{ /* bundle-backed */ }
func NewBase(b Bundle) Base
func (b Base) Name() string
func (b Base) Metadata() connectors.Metadata
func (b Base) Definition() connectors.Definition
func DerivedSyncModes(s StreamSpec, sch *StreamSchema) []string // §B.6 truth table
```

## 3. package `conformance` (new: `internal/connectors/conformance`)

```go
type Check struct {
    Name   string `json:"name"`   // e.g. "pagination_terminates"
    Passed bool   `json:"passed"`
    Error  string `json:"error,omitempty"`
}
type Report struct {
    Connector string  `json:"connector"`
    Passed    bool    `json:"passed"`
    Static    []Check `json:"static"`
    Dynamic   []Check `json:"dynamic"`
}
func Run(ctx context.Context, b engine.Bundle, h engine.Hooks) Report
// Test entrypoint: TestConformance iterates engine.LoadAll(defs.FS) with t.Run(name, ...)
```

## 4. package `certify` (new: `internal/connectors/certify`)

```go
type Options struct {
    Connector string
    Stream    string            // default: first cursor stream, else first
    Limit     int               // default 50
    Modes     []string          // default all 5
    Config    map[string]string // connector config for credentials add
    SecretEnv map[string]string // field -> ENV name
    KeepWork  bool
}
type Runner struct{ /* unexported */ }
func NewRunner(o Options) *Runner
func (r *Runner) Run(ctx context.Context) (Report, error) // wave0: source stages 0–11 only

type Report struct { /* fields per certification design §A; see DATA-MODEL §6 */ }
func (rep *Report) Save(dir string) error
func LoadReport(path string) (Report, error)

// cliharness.go
type CLIResult struct {
    ExitCode int
    Stdout, Stderr string
    Envelope map[string]any // parsed --json output; nil for text runs
    Kind     string
}
func (h *Harness) Run(args ...string) CLIResult              // injects --root, captures, scans
func (h *Harness) MustKind(res CLIResult, kind string, exit int) error
func ScanForSecrets(text string, secrets []string) []string  // exact/base64/urlencoded
```

## 5. Tools

```go
// cmd/connectorgen: subcommands
//   validate [dir] [--json]   exit 0/1
//   gen                       writes hooks/hookset/hookset_gen.go, native/nativeset/nativeset_gen.go
//   new <name>                scaffolds defs/<name>/
// cmd/inventorygen: writes docs/migration/inventory.json (DATA-MODEL §4)
// cmd/registrygen: UNCHANGED except skip map += {defs, engine, hooks, native, conformance, certify}
//   (cmd/registrygen/main.go:30)
```

## 6. Explicit non-changes (compat assertions worth testing)

- `connectors.RegisterFactory` / `RegisterNativeLive` semantics untouched
  (`connectors.go:60,49`); legacy goldens still registered.
- `connsdk` public surface untouched (engine adds paginators in its own package).
- `internal/app`, `internal/cli` untouched; CLI JSON envelopes unchanged.
- `internal/connectors/native/postgres` performs NO registration in wave0 (guard test, T-17).
