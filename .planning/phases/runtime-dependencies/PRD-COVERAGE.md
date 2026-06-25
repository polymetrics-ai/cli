# PRD Coverage: runtime-dependencies

Source PRD:

- `POLYMETRICS_GO_CLI_MONOLITH_PRD_ARCHITECTURE.md`
- `docs/architecture/runtime-dependencies.md`

Covered:

- PostgreSQL, DragonflyDB, and Temporal Go dependencies added to `go.mod`.
- Runtime health checks implemented in `internal/runtimecheck`.
- PostgreSQL run ledger implemented in `internal/ledger`.
- DragonflyDB lease coordination implemented in `internal/coordination`.
- Performance comparison implemented in `internal/perf`.
- CLI commands:
  - `poly runtime doctor`
  - `poly perf compare`
  - `poly perf compare --runtime`
  - `poly etl run --runtime`
- Generated CLI docs updated under `docs/cli`.
- Runtime performance documentation added under `docs/performance`.

Not fully covered:

- Full PostgreSQL replacement for JSON state.
- Full Temporal workflow execution for ETL and reverse ETL.
- Dragonfly-backed production runner queues and rate limits.
- Runtime integration tests that start containers automatically.

Reason:

The user approved dependency implementation. Full Temporal workflow migration is a larger execution-model change and should be handled as a dedicated follow-up phase with workflow/activity replay tests.

