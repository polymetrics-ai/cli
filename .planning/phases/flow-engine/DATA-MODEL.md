# DATA-MODEL — Flow Engine (Phase 0)

## In-memory structs

Defined in `internal/flow/manifest.go` and `internal/flow/engine.go`.

### FlowManifest

| Field | Type | Description |
|-------|------|-------------|
| Version | int | Must be 1 |
| Name | string | Alphanumeric + `-_`, unique per project |
| Description | string | Optional |
| Steps | []FlowStep | Ordered list; engine derives execution order from DAG |

### FlowStep

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Unique within manifest |
| Kind | StepKind | "sync" or "query" |
| Connection | string | Connection ID (sync only) |
| Streams | []string | Stream names (sync only) |
| SQL | string | SQL query (query only) |
| In | []string | Tables consumed; declares DAG edges |
| Out | []string | Tables produced; declares DAG edges |

### RunResult

| Field | Type | Description |
|-------|------|-------------|
| FlowName | string | Manifest name |
| Status | string | "ok" / "failed" / "dry_run" |
| Steps | []StepResult | Per-step results in execution order |

### StepResult

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Step ID |
| Kind | string | Step kind |
| Status | string | "ok" / "skipped" / "failed" / "dry_run" |
| RecordsRead | int | Records read by step |
| RecordsWritten | int | Records written by step |
| DurationNs | int64 | Wall time in nanoseconds |
| Error | string | Error message if failed (omitempty) |

## Persisted state

### Ledger entries

File: `.polymetrics/logs/ledger.jsonl`

Reuses existing `ledger.RunRecord`. For flow runs:

| Field | Value |
|-------|-------|
| Mode | "flow" |
| Operation | "<flow-name>" (flow entry) or "<flow-name>/<step-id>" (step entry) |
| Status | "running" / "success" / "failed" |
| RecordsRead | per step |
| RecordsWritten | per step |

### Checkpoint store

File: `.polymetrics/state/flow-checkpoints.json`

```json
{
  "daily-sync": {
    "sync-hubspot": "success",
    "score-contacts": "success"
  }
}
```

Top-level key: flow name. Second-level key: step ID. Value: "success" (only success is
persisted; "failed" and "running" are transient).

### Lock files

File: `.polymetrics/locks/flow-<name>.lock`

Content: PID of the running process as a decimal integer followed by newline.

## Schema migrations

None. All stores are append-only JSONL (ledger) or overwrite-on-write JSON (checkpoint).
No schema migration gate is triggered.

## Warehouse tables

Flow steps produce and consume tables in the existing local JSONL warehouse under
`.polymetrics/warehouse/`. No new warehouse schema is introduced; table names come from
step `out` declarations and are written by the existing `app.ETLRun` and `app.QuerySQL`
implementations.
