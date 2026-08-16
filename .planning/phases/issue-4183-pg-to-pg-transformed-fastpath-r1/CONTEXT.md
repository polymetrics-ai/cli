# Context — PostgreSQL transformed fast path R1

## Task Delivery Header

- Issue: Refs #4183 — feat(postgres): add transformed full-overwrite binary-COPY fast path; child of #3972 and #4097.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → docs/4015-connector-release-certification → main`.
- Delivery: A direct PR from `fm/cli-pg-to-pg-transformed-fastpath-r1` targeting `integration/4015-mvp-flat-r1`, with the API-reported base verified after opening.
- Working branch: `fm/cli-pg-to-pg-transformed-fastpath-r1`.
- Task: Implement the approved, additive PostgreSQL-to-PostgreSQL transformed `full_overwrite` vertical slice. The substrate is connector-neutral: source extraction and destination bulk application are ports; Arrow batches, closed transforms, durable segments/manifests, credit control, receipts, checkpoints, and per-unit timing contain no PostgreSQL, pgx, or SQL types. PostgreSQL supplies the range extractor and binary-COPY shadow-publish adapter only.
- Verification: Red/green unit and production-composition tests; live two-plus-page PostgreSQL container test; tagged binary correctness/performance harness; targeted Go tests, vet, binary build, individual verify gates, `verify-work`, code review, and API PR-base readback.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Multi-page full overwrite retains every source page | live | A two-plus-page PostgreSQL test reads the published target and asserts every expected row; a last-page-only target fails. |
| Closed transform rejects invalid/drifted plans before I/O | live | Production construction returns the typed plan/hash refusal and the injected extractor records zero calls. |
| Neutral segment pipeline carries typed transformed rows without row maps | fake | Unit tests use a typed in-memory Arrow batch port because live tests must not inspect internal allocation shape; the binary live test proves the port is reachable. |
| COPY/shadow publish writes the transformed target exactly once | live | A PostgreSQL container observes transformed counts/aggregates and a single matching receipt after replay from a post-receipt checkpoint fault. |
| Progress survives success and failure before cleanup | live | Durable run state contains phase counters and elapsed intervals after an injected stage/apply failure and after a successful run. |
| 2–3 GB gate is measured honestly | fake | The qualified separate-volume host is unavailable in this worktree; the tagged binary harness and its preflight/JSON assertions ship but remain unscored here. |

## Assertion Rule

Every live test asserts an externally observable target row set, receipt, checkpoint, or durable run record. No acceptance claim rests on an exit code or `err == nil` alone.

## Binding decisions

- Read the performance report's `Next PR: one honest vertical slice`, `In scope`, `Explicitly out of scope`, and `Exact 2–3 GB proof plan` as the design contract. No redesign is authorized.
- PR #4182 is present in the base. It established durable progress and per-unit deadline behavior; this slice must preserve no overall run deadline and must emit payload-free phase telemetry before cleanup on success and failure.
- `full_overwrite` is the only production fast apply. Other modes, CDC, object stores, user-owned tables, unlogged staging, custom pgwire codecs, and legacy warehouse changes remain excluded.
- The transformed production path is not an identity relay: it must support the specified closed projection/rename, declared types, scalar expressions, and filtering. Cross-segment production dedupe remains excluded; its kernel proof may not be presented as the production mode.
- The PostgreSQL adapter must use binary `CopyFrom`; no bulk-path INSERT/UPDATE/DELETE is permitted. Logged shadow publication, index/constraint build, deterministic receipt insertion, and checkpoint reconciliation preserve exactly-once ordering.
- A future connector inherits the neutral `Extract`/Arrow pipeline, transform, segment, manifest, credit, receipt, and checkpoint contracts. It implements a source extractor and/or destination bulk applier without a PostgreSQL type leaking into those shared layers.

