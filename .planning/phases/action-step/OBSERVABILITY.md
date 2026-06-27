# OBSERVABILITY — Action Step (Phase 1)

## Metrics emitted (in RunResult / StepResult)

| Field | Type | Description |
|-------|------|-------------|
| `records_read` | int | Records fetched from source table |
| `records_written` | int | Records successfully sent to destination |
| `records_failed` | int | Records quarantined in DLQ after MaxRetries |
| `duration_ns` | int64 | Step wall-clock time in nanoseconds |
| `dlq_path` | string | Path to DLQ file if any failures |
| `schema_drift` | bool | True if drift detected (step paused) |

## Ledger entries

All action step events written to `internal/ledger` (JSONLedger by default):

| Operation | Status | When |
|-----------|--------|------|
| `<flow>/<step>` | `running` | Step start |
| `<flow>/<step>` | `receipt` | Per-batch success |
| `<flow>/<step>` | `failed` | Step error |
| `<flow>/<step>` | `dlq` | Record quarantined |
| `<flow>/<step>` | `schema_drift` | Drift detected, step paused |

## Structured log format (stderr, human mode)

```
[action] step=send-emails records=42 sent=40 failed=2 dlq=.polymetrics/dlq/... duration=1.2s
[action] step=send-emails SCHEMA DRIFT: field "email" type changed string→integer — step paused
```

## Error alert on schema drift

When `ErrSchemaDrift` is returned, the engine writes a ledger entry with `Status="schema_drift"`
and prints a human-readable alert to stderr (or --json equivalent).
