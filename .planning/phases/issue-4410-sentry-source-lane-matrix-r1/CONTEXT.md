# Sentry source-to-seven-lane matrix — #4410

## Task Delivery Header

- Issue: Refs #4410 — Sentry source-to-seven-lane matrix.
- Base branch: `origin/main` at `813f457a925f7ee3fe3bea101a43e445992c8552`.
- Merges into: `main`.
- Delivery: Scoped branch committed, pushed, and evidenced in #4410; no pull request or merge is opened by this task.
- Working branch: `fix/4410-sentry-track-a-r1`.
- Task: Restore only absent connector-local retained Sentry source artifacts after byte/digest verification against Batch R1 parent `dc481bac`, then create a source-row × seven-lane mapping matrix with a local reconciliation test. Preserve every source ID and make no runtime, executor, certification, credential, or shared-control change.
- Verification: Exact parent-byte verification, source-row/cell/count reconciliation, red/green/edge Go tests, JSON validation, formatting, scoped source/map checks, and an issue completion proof with every source ID.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every retained source ID stays represented exactly once | live | The local test decodes the lock and matrix, rejects hidden or duplicate IDs, and reconciles their counts. |
| Every row has seven cited lane cells and source facts | live | The local test recomputes source facts and expected mapping dispositions from the retained lock. |
| Pageable reads and writes cannot be replaced by direct cells | live | In-memory adversarial cases reject a missing ETL/sync cell for a pageable row and a missing reverse-ETL cell for a write. |
| Retained source artifacts preserve Batch R1 bytes | live | SHA-256 checks compare each copied artifact to the exact Batch R1 parent bytes. |
| No runtime capability is claimed or changed | live | Matrix state validation permits source-only states for this Track A slice and changed-path review excludes shared/runtime paths. |

## Scope decision

`origin/main` has no Sentry `sources/` directory. Batch R1 parent `dc481bac` contains retained source lock material. This task will copy only the missing retained source material after verifying exact bytes and provider digest identity; it will not copy Batch R1 declarations, writes, sync transport, certification sweep, or shared mapping controls.

Copied source-local files are exactly the parent blobs, verified after copying:

| File | SHA-256 | Bytes |
| --- | --- | ---: |
| `sentry-operation-source-lock.json` | `383633e6c8403b78c44d9841cd271ec49dbbcecc5778f6c74db1ec162ef1a059` | 4,134,778 |
| `sentry-operation-descriptor.json` | `0f138ba59ce9b29c6cc331c5122c974f2540ab93337e77f56fd34e767e39657c` | 3,305,062 |
| `sentry-operation-crosswalk.json` | `1651cb64ab3dfc08912f63f45d6f9fb52d59fb23c34c9473a2b16554080dd69c` | 210,981 |
| `sentry-declaration-disposition.json` | `87ce883dddbb7846f32cd90174e31517af7ffe1aec3e191dba4d854f2b9de231` | 611,974 |

The copied lock pins the provider OpenAPI document at `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435`, captured `2026-08-19T11:28:09Z`, with 223 REST operations.

## Foundation Atlas decision

The Batch R1 Atlas was consulted before any `missing_foundation` label. `source.retention-import.v1`, `source.projection-admission.v1`, `warehouse.stage-etl.v1`, `warehouse.reverse-etl.v1`, and `transport.sync-contract.v1` are catalogued as available, but none is invoked or claimed by this mapping-only delivery. Every cell therefore remains either `mapped_unproven` or source-evidenced `not_applicable`; no runtime foundation gap is named.

`docs/connector-terminology.md`, cited by the issue, is absent from both current main and the Batch R1 parent. The connector-lane build-order terminology was used instead; this is a documentation-reference gap, not a source, mapping, or runtime gap.

## GSD and skill record

Ran `scripts/gsd doctor` and resolved the required `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts. The current task is explicitly single-worker and the adapter cannot provide compatible isolated planning roles; manual inline GSD is therefore recorded here and in the plan/ledger rather than spawning agents.

Loaded skills: `connector-lane-build-order`, `go-engineering` (fundamentals and agentic ETL), and the repository required-skills routing. Required Go skills selected for this source-backed local validator are recorded in `PLAN.md` before implementation.
