## Intent

Refs #4183. This stacked direct PR targets `integration/4015-mvp-flat-r1` and implements the additive PostgreSQL-to-PostgreSQL transformed `full_overwrite` vertical slice.

## What changed

- Characterized and fixed `cli-full-overwrite-per-page-truncate-dataloss-r1`: a two-page production-binary test proved the historical per-page truncate loss, and both legacy and Arrow routes now use one run-scoped shadow/publish lifecycle.
- Added closed, normalized, hash-bound `TransformPlanV1` plus `pm connections create --transform-file`. It supports typed projection/rename, timezone-preserving timestamp, date, checked integer multiply/cast, uppercase, modulo, and `not_equal`; arbitrary SQL remains forbidden.
- Added connector-neutral Arrow segment/controller/byte-credit contracts in `internal/synctransport`, versioned immutable Arrow-Parquet manifests in `internal/warehouse`, and transform/receipt contracts in `internal/connectors/database`.
- Isolated PostgreSQL range extraction and reusable pgx binary `CopyFromSource`/shadow publication in `internal/connectors/native/postgres`. The fast path does not build `map[string]any` or a Go struct per output row, and has no INSERT fallback.
- Added durable source/transform/Parquet close+fsync/COPY/index-build/publish+receipt/checkpoint/wall counters, decimal MB/s, MiB/s, and the 3 GiB segment admission safety check.

### Connector seam

A future source (S3 Parquet, MySQL, MongoDB, etc.) implements `synctransport.ArrowRangeExtractor`; a future destination (ClickHouse, MongoDB, etc.) implements `synctransport.ArrowBulkDestination` / `ArrowFullOverwriteRun`. Both inherit the Arrow/DuckDB transform, immutable segment/manifest, byte-credit control, deterministic receipt sequencing, per-unit deadline handling, and checkpoint-after-readback controller. PostgreSQL alone provides the current range primitive and pgx binary COPY adapter. No PostgreSQL/pgx type crosses the shared ports.

MySQL was not added: the optional variant needs a separately verified native extractor primitive and live proof; no substrate change was made to force it into this slice.

## Red / Green / Refactor evidence

- **Happy:** live binary tests retain all pages and read exact projected/typed/filtered target rows, one receipt, and durable counters.
- **Bad:** typed missing-segment, invalid source identity, transform hash/file, manifest, credit, and COPY-value refusals assert no extractor/applier/file side effect before refusal.
- **Edge:** zero source, nulls, timestamp units, cancelled credits, duplicate/corrupt manifests, repeated transformer batches, two-page overwrite, and the 3 GiB boundary are named tests. The complete ledger is `TDD-LEDGER.md` in the phase evidence.
- Manual GSD fallback was recorded because this worker had no compatible isolated Pi role runtime. Resolved commands: `scripts/gsd prompt discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.

## 5 GB measurement (honest result)

Input bytes are pre-transform projected source Arrow buffers, excluding Parquet, pgwire, target storage, and checkpoint bytes. The opt-in live proof moved 5,368,947,776 logical bytes for the realistic mapping in 48.032609708 s: **111.78 decimal MB/s / 106.60 MiB/s**, below the required 200 MB/s / 25 s gate. Identity control was 105.68 MB/s / 100.79 MiB/s in 50.80 s. The bottleneck is source read (16.699 s) plus binary COPY (27.638 s) on this one-host Docker diagnostic; report and peak disk data are committed in `fastpath-5gb-proof.json`.

The test intentionally fails only when the explicit 5 GB performance opt-in is set; without it, it visibly skips. It never claims the gate passed.

## Verification

- `go test -count=1 -timeout 20m ./internal/connectors/database ./internal/connectors/native/postgres ./internal/synctransport ./internal/warehouse`
- Targeted `internal/app` transform/approval/measurement and `internal/cli` transform-file tests; tagged 5 GB skip without opt-in.
- Live Docker Unix-socket production-binary correctness test for two-page overwrite and transformed Arrow COPY.
- Scoped `go vet`, `go build ./cmd/pm`, `pm help connections`, bare `pm connections`, command help, generated docs validation, and smoke flow.
- Agent-contract, connector validation/surface/boundary, connector canon, pinned-build, Homebrew notification, and release-target checks.

## CLI parity / skills / safety

Generated `docs/cli/connections.md` and website ETL/CLI reference document `--transform-file`; runtime help and bare namespace behavior were verified. Required Go/CLI/testing/database/concurrency/performance/observability skills are recorded in `PLAN.md`. The route uses existing plan/preview/approval, keeps approval tokens off argv and persistent artifacts, and preserves the legacy local warehouse table and JSONL WAL.

## Review

Opening this stacked PR triggers the repository's automatic Claude review. No Copilot fallback was requested; review disposition is pending the automatic run. No human merge is requested.
