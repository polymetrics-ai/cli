# SPEC — Flow Engine (Phase 0)

## Package layout

```
internal/flow/
    manifest.go        — YAML parse + validate → FlowManifest struct
    dag.go             — DAG build + topological sort + cycle detection
    engine.go          — Executor: runs steps, writes ledger, holds lease
    checkpoint.go      — checkpoint read/write backed by internal/app state store
    errors.go          — sentinel errors (ErrCyclicDependency, ErrManifestInvalid, ErrLeaseHeld, …)
    flow_test.go       — table-driven tests (manifest, dag, engine, checkpoint, lease)

internal/cli/flow_cli.go   — pm flow plan|preview|run|status|list dispatcher
```

No new packages. All new files sit inside existing package boundaries.

## Manifest format (YAML authoring)

```yaml
version: 1
name: likely-customers
description: Sync contacts then score them

steps:
  - id: sync-hubspot
    kind: sync
    connection: hubspot-prod
    streams: [contacts, companies]
    out: [contacts, companies]           # tables written

  - id: score-contacts
    kind: query
    sql: |
      SELECT * FROM contacts WHERE email IS NOT NULL
    out: [scored_contacts]               # table written
    in:  [contacts]                      # declares dependency on sync-hubspot
```

### Normalized JSON (internal struct)

```go
// internal/flow/manifest.go
type FlowManifest struct {
    Version     int          `yaml:"version" json:"version"`
    Name        string       `yaml:"name"    json:"name"`
    Description string       `yaml:"description" json:"description"`
    Steps       []FlowStep   `yaml:"steps"   json:"steps"`
}

type FlowStep struct {
    ID         string            `yaml:"id"         json:"id"`
    Kind       StepKind          `yaml:"kind"       json:"kind"`
    Connection string            `yaml:"connection"  json:"connection,omitempty"`
    Streams    []string          `yaml:"streams"    json:"streams,omitempty"`
    SQL        string            `yaml:"sql"        json:"sql,omitempty"`
    In         []string          `yaml:"in"         json:"in"`
    Out        []string          `yaml:"out"        json:"out"`
}

type StepKind string
const (
    KindSync  StepKind = "sync"
    KindQuery StepKind = "query"
)
```

### Validation rules

| Rule | Error |
|------|-------|
| `version` must be 1 | ErrManifestInvalid |
| `name` non-empty, alphanumeric + `-_` only | ErrManifestInvalid |
| each step `id` unique, non-empty | ErrManifestInvalid |
| `kind` must be one of the known kinds | ErrManifestInvalid |
| `sync` step must have `connection` + at least one stream | ErrManifestInvalid |
| `query` step must have non-empty `sql` | ErrManifestInvalid |
| every table in `in` must be produced by some step's `out` | ErrManifestInvalid |

## DAG build

`internal/flow/dag.go` implements Kahn's algorithm (stdlib only):

1. Build adjacency list: for each step, if step B's `in` contains a table in step A's `out`,
   add edge A→B.
2. Detect cycles: if Kahn's algorithm does not consume all nodes, return `ErrCyclicDependency`
   with the remaining node IDs in the error message.
3. Return `[]string` ordered step IDs (topological order).

## Executor

`internal/flow/engine.go`

```go
type Engine struct {
    Manifest   FlowManifest
    App        AppAdapter         // interface wrapping *app.App
    Ledger     LedgerAdapter      // interface wrapping ledger.JSONLedger or ledger.PostgresLedger
    Checkpoint CheckpointStore
    LockDir    string             // .polymetrics/locks/
}

type RunOptions struct {
    DryRun bool
    Force  bool   // clear checkpoints before run
    JSON   bool
}

func (e *Engine) Run(ctx context.Context, opts RunOptions) (RunResult, error)
```

### AppAdapter interface (avoids import cycle)

```go
type AppAdapter interface {
    ETLRun(ctx context.Context, connectionID string, streams []string) (ETLResult, error)
    QuerySQL(ctx context.Context, sql string, limit int) ([]map[string]any, error)
}
```

### Execution flow

1. Acquire `state.FileLock` at `<LockDir>/flow-<name>.lock`. Return `ErrLeaseHeld` on failure.
2. If `opts.Force`, clear all checkpoints for this flow.
3. Compute topological order.
4. For each step in order:
   a. Check checkpoint — skip if `success`.
   b. Write ledger entry status=`running`.
   c. Dispatch to step handler (`runSync` or `runQuery`).
   d. Write ledger entry status=`success` or `failed`.
   e. Write checkpoint `success` on success; on failure stop and return the error with step ID.
5. Release lock.
6. Return `RunResult`.

### RunResult

```go
type RunResult struct {
    FlowName string       `json:"flow_name"`
    Status   string       `json:"status"`   // "ok" | "failed" | "dry_run"
    Steps    []StepResult `json:"steps"`
}

type StepResult struct {
    ID             string `json:"id"`
    Kind           string `json:"kind"`
    Status         string `json:"status"`   // "ok" | "skipped" | "failed" | "dry_run"
    RecordsRead    int    `json:"records_read"`
    RecordsWritten int    `json:"records_written"`
    DurationNs     int64  `json:"duration_ns"`
    Error          string `json:"error,omitempty"`
}
```

## Checkpoint store

`internal/flow/checkpoint.go` — thin wrapper over `.polymetrics/state/flow-checkpoints.json`
(re-uses `internal/state.JSONStore` pattern).

```go
type CheckpointStore interface {
    Get(flowName, stepID string) (string, error)   // returns status or ""
    Set(flowName, stepID, status string) error
    Clear(flowName string) error
}
```

## CLI dispatcher

`internal/cli/flow_cli.go` — new file, same pattern as existing `runETL`, `runReverse`.

```
pm flow plan   [--file <path>] [--force] [--json]
pm flow preview [--file <path>] [--json]
pm flow run    [--file <path>] [--force] [--json]
pm flow status <name> [--json]
pm flow list   [--json]
```

- `--file` defaults to `.polymetrics/flows/<name>.yaml` when name is given, else prompts.
- Wired into the `switch cmd` block in `internal/cli/cli.go` under `case "flow":`.

## Ledger entries

Uses existing `ledger.RunRecord`. `Mode` = `"flow"`, `Operation` = `"<flow-name>/<step-id>"`.

## Lease strategy

Dependency-free: `state.FileLock` at `.polymetrics/locks/flow-<name>.lock`.
Runtime-backed (opt-in, Phase 0 does not implement): `runtime.Module.Leases` via DragonflyDB.
`internal/runtimecheck` will indicate whether Dragonfly is available; for now, always use FileLock.

## Error types

```go
var (
    ErrManifestInvalid    = errors.New("flow: manifest invalid")
    ErrCyclicDependency   = errors.New("flow: cyclic dependency detected")
    ErrLeaseHeld          = errors.New("flow: another run is already in progress")
    ErrStepFailed         = errors.New("flow: step failed")
    ErrUnknownStepKind    = errors.New("flow: unknown step kind")
)
```

## Files to create

| File | Package | Purpose |
|------|---------|---------|
| `internal/flow/manifest.go` | `flow` | Parse + validate |
| `internal/flow/dag.go` | `flow` | DAG + topo sort |
| `internal/flow/engine.go` | `flow` | Executor |
| `internal/flow/checkpoint.go` | `flow` | Checkpoint store |
| `internal/flow/errors.go` | `flow` | Sentinel errors |
| `internal/flow/flow_test.go` | `flow` | All tests |
| `internal/cli/flow_cli.go` | `cli` | CLI dispatcher |

## Files to modify

| File | Change |
|------|--------|
| `internal/cli/cli.go` | Add `case "flow":` to switch |
| `internal/cli/docs.go` | Register `flow` in manual index (if applicable) |

## Design direction

Not applicable — CLI/backend phase, no visual UI.
