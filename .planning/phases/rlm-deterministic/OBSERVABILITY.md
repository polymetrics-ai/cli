# OBSERVABILITY — Phase 2: RLM Deterministic Backend

## Structured log lines (stderr)

All RLM log output goes to stderr. Format matches existing `internal/cli` style: key=value pairs.

### Run start
```
[rlm] start mode=deterministic in=contacts out=lead_scores spec=likely-customers
```

### Run complete
```
[rlm] done mode=deterministic in=contacts out=lead_scores records_read=42 records_scored=42 records_failed=0 duration=1.2ms
```

### Per-record parse error (only emitted when records_failed > 0)
```
[rlm] parse_error line=17 error="invalid character '}' looking for beginning of value"
```

### Dry run indicator
```
[rlm] dry_run=true out_table_not_written
```

---

## Machine-readable envelope (stdout, --json flag)

```json
{
  "mode": "deterministic",
  "in_table": "contacts",
  "out_table": "lead_scores",
  "records_read": 42,
  "records_scored": 42,
  "records_failed": 0,
  "duration_ns": 1234567,
  "dry_run": false
}
```

This is the `RunResult` struct encoded as JSON. Stable contract (see API-CONTRACT.md).

---

## Ledger entry

One `ledger.RunRecord` appended per run (success or failure):

| Field | Value |
|---|---|
| `id` | UUID generated per run |
| `mode` | `"rlm"` |
| `operation` | backend mode: `"deterministic"`, `"fixture"`, `"model"` |
| `status` | `"success"` or `"error"` |
| `records_read` | from RunResult |
| `records_written` | from RunResult.RecordsScored |
| `duration_ns` | nanoseconds |
| `created_at` | UTC time of run start |

The ledger write is best-effort: a ledger write failure is logged to stderr but does not fail the run.

---

## Metrics (this phase — counters only, no external sink)

The following counters are available in `RunResult` and surfaced in the JSON envelope:

| Counter | Description |
|---|---|
| `records_read` | Total lines read from InTable |
| `records_scored` | Lines successfully scored and included in OutTable |
| `records_failed` | Lines skipped due to parse or scoring error |
| `duration_ns` | Wall-clock duration of the Run call in nanoseconds |

No metrics push to an external sink in this phase. Flow-level aggregation (Phase 0) may wrap these.

---

## What is NOT in scope for this phase

- Prometheus/OpenTelemetry metrics export (no new deps allowed).
- Distributed tracing.
- Per-feature score breakdown in output (may be added in a future phase as opt-in `--verbose`).
