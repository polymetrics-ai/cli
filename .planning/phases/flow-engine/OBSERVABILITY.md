# OBSERVABILITY — Flow Engine (Phase 0)

## Logging

All log output uses `fmt.Fprintf(stderr, ...)` for human mode. In `--json` mode, structured
JSON is written to stdout. No third-party logging library is introduced.

### Log events

| Event | Output |
|-------|--------|
| Flow started | stderr: `flow <name>: starting (<n> steps)` |
| Step started | stderr: `  step <id> [<kind>]: running` |
| Step skipped | stderr: `  step <id> [<kind>]: skipped (checkpointed)` |
| Step succeeded | stderr: `  step <id> [<kind>]: ok (<records> records, <ms>ms)` |
| Step failed | stderr: `  step <id> [<kind>]: FAILED: <error>` |
| Flow succeeded | stderr: `flow <name>: ok` |
| Flow failed | stderr: `flow <name>: FAILED at step <id>` |
| Lock acquired | stderr (debug only, not default) |
| Stale lock removed | stderr: `flow <name>: removed stale lock (pid <pid> not running)` |

## Metrics (in RunResult JSON envelope)

Every `--json` response includes per-step and per-flow:

| Metric | Field |
|--------|-------|
| Records read | `steps[].records_read` |
| Records written | `steps[].records_written` |
| Step wall time | `steps[].duration_ns` |
| Step status | `steps[].status` |
| Flow status | `status` |

## Ledger (audit trail)

Every step and flow run writes to `.polymetrics/logs/ledger.jsonl`. The ledger is the
durable record for:
- run start time (`created_at`)
- outcome (`status`)
- records processed

This provides an offline audit log queryable with `jq` or `pm query`.

## What is not measured in Phase 0

- Per-connector latency breakdown (added per connector in existing `app.ETLRun`)
- Schema-drift events (Phase 1)
- DLQ depth (Phase 1)
- Token counts (Phase 5)

## Future phases

Phase 3 (scheduling) will add cron-triggered run metrics. Phase 5 (agent mode) will add
token/byte efficiency benchmarks.
