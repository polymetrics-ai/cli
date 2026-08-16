## Intent

Refs #4183. This stacked direct PR targets `integration/4015-mvp-flat-r1` and implements the additive PostgreSQL-to-PostgreSQL transformed `full_overwrite` vertical slice.

## Benchmark re-qualification — gate still not met

Input bytes are pre-transform projected source Arrow buffers; they exclude Parquet, pgwire, target storage, and checkpoint bytes.

**Historical, resource-contended result:** 5,368,947,776 logical input bytes in 48.032609708 s: **111.78 decimal MB/s (106.60 MiB/s)**. This ran inside a **2-CPU / 2-GiB Colima VM** while macOS Spotlight mass re-indexing was contending for CPU (`mds_stores` at 303% CPU).

**Quiet re-qualification:** 5,368,947,776 logical input bytes in 48.000221500 s: **111.85 decimal MB/s (106.67 MiB/s)**. This ran inside an **8-CPU / 16-GiB Colima VM**; `mds_stores` was 20.2% CPU at launch, below the ~50% quiet-machine threshold. **This remains BELOW the 200 MB/s / 25 s commitment.**

The named bottleneck remains the **sequential source-read plus binary-COPY critical path** (16.869 s source read and 27.727 s binary COPY in the quiet run). More cores will not yield a linear 3× gain because this path does not parallelize with cores; extra cores can help PostgreSQL WAL, index/background work, and the DuckDB vectorised transform. Parallel partitioned extraction and the bounded pipeline were deliberately out of scope for this slice and remain the path to the gate. The near-identical qualified pair retains the original number while showing that the current ceiling is the sequential path, not a core-scalable claim.

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

### CI repair round 1

- Regenerated the three `connections` golden transcript entries only after verifying that the diff contains the intended `--transform-file` synopsis token and the closed-plan `TRANSFORMS` help section; the transcript test passes.
- Regenerated `website/lib/docs.generated.ts` with `pnpm --dir website run gen:docs`; a repeat run is byte-stable and website type-checking passes.
- Dependency Review's transitive `github.com/apache/thrift@v0.21.0` finding is resolved by upgrading the direct `github.com/apache/arrow-go/v18` dependency to `v18.7.0`, which selects Thrift `v0.24.0`. `go mod verify`, targeted Arrow/PostgreSQL suites, and `go build ./cmd/pm` pass.
- The accepted independent review corrected a test-harness regression: ordinary `afterApply` now runs after acknowledgement construction. The inherently per-page source-failure and page-one/page-two rows retain page-mode coverage; full-overwrite has honest inverse/final-CAS coverage at pre-publish abort, post-receipt read-back stale-writer, and post-final-checkpoint terminal boundaries. Production publish-then-checkpoint is unchanged. The two post-final-checkpoint cancellation/failure cases are explicitly inapplicable for full-overwrite until their product observation point is decided; no test is skipped or disabled.
- The load-sensitive stale-writer test now synchronizes at receipt read-back, not page apply, and passes 20 consecutive runs at the existing 20-second timeout. `make test` passes in full.
- New test classes: **happy** — post-ack ordinary callback and distinct full-overwrite page/publish/read-back lifecycle; **bad** — pre-publication and wrong-receipt refusals before read-back, plus typed stale/missing final-checkpoint conflicts; **edge** — repeated acknowledgement/abort, source failure and cancellation before publish, two-page receipt-to-final-CAS racing, and unrelated final-checkpoint rebase. These are App construction-path tests; the existing tagged binary tests remain the production PostgreSQL proof.
- CodeQL repair: a valid `TransformPlanV1` reaches the Postgres compiler only through the closed parser and the regular `--transform-file` admission capped at 64 KiB; the native representation is unexported. Rather than rely on that bound, the map capacity hint now performs no arithmetic. The disk probe removes only the impossible unsigned `Bavail < 0` branch and retains its block-size refusal and multiplication-overflow clamp. No benchmark value, production lifecycle, or externally observable behavior changed; existing Postgres-plan **happy/bad** and disk-capacity **happy/bad/edge** tests cover the preserved results. Targeted tests, `go vet ./...`, and full `make test` pass. `security/snyk` is pre-existing on the base branch and intentionally outside this PR.
- Lint repair: Arrow v18.7 deprecates `arrow.Record` and `NewRecord`; the migration uses `arrow.RecordBatch`/`NewRecordBatch` at all eight scoped sites (including zero-row/null edge tests) with no behavior change or suppression. `make lint` passes with **0 issues** and full `make test` passes; this is now recorded before push as well as in the status file.

## CLI parity / skills / safety

Generated `docs/cli/connections.md` and website ETL/CLI reference document `--transform-file`; runtime help and bare namespace behavior were verified. Required Go/CLI/testing/database/concurrency/performance/observability skills are recorded in `PLAN.md`. The route uses existing plan/preview/approval, keeps approval tokens off argv and persistent artifacts, and preserves the legacy local warehouse table and JSONL WAL.

## Review

Opening this stacked PR triggers the repository's automatic Claude review. No Copilot fallback was requested; review disposition is pending the automatic run. No human merge is requested.
