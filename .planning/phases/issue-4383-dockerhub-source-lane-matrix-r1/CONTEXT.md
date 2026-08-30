# Docker Hub source-to-seven-lane matrix — #4383

## Task Delivery Header

- Issue: Refs #4383 — Docker Hub source-to-seven-lane matrix.
- Base branch: `origin/main` at `813f457a925f7ee3fe3bea101a43e445992c8552`.
- Merges into: `main`.
- Delivery: Scoped branch committed, pushed, and evidenced in #4383; no pull request or merge is opened by this task.
- Working branch: `fix/4383-dockerhub-track-a-r1`.
- Task: Restore only absent, byte-verified Docker Hub source sidecars from Batch R1 parent `dc481bac`, then add a connector-local source-operation × seven-lane mapping matrix and local reconciliation test. Preserve all source IDs and make no runtime, executor, generator, certification, credential, or shared-control change.
- Verification: Parent-byte verification, source-row/cell/count reconciliation, red/green/edge Go tests, JSON validation, formatting, scoped source/map checks, changed-path review, and a no-checkbox issue completion proof.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every retained source ID is represented exactly once | live | The local test decodes the lock and matrix, rejects hidden and duplicate IDs, and reconciles the exact count. |
| Every source row has seven cited lane cells and source facts | live | The local test recomputes source identity, media, scope, pagination, callback/cursor absence, and expected source-only lane dispositions. |
| Pageable or extractable reads and writes retain independent ETL/sync or reverse-ETL cells | live | In-memory adversarial cases reject missing ETL/sync cells for source paging/extractability and missing reverse-ETL cells for source mutations. |
| Retained source artifacts preserve Batch R1 bytes | live | SHA-256 checks compare each copied sidecar to the exact Batch R1 parent blob. |
| No runtime capability is claimed or changed | live | Matrix states remain source-only and changed-path review excludes shared/runtime paths. |

## Scope decision

`origin/main` has no `internal/connectors/defs/dockerhub/sources/` directory. Batch R1 parent `dc481bac` contains four Docker Hub source sidecars. This slice copies only those exact sidecars after byte verification, and adds only a connector-local matrix/test and planning evidence. It deliberately does not copy Batch R1 `writes.json`, `sync_transport.json`, `certification-sweep.json`, or any shared mapping control.

| File | Parent SHA-256 | Bytes |
| --- | --- | ---: |
| `dockerhub-operation-source-lock.json` | `0a9224a085305dd51037e2f3d723d53cef9659625c0146587a97207747011bc3` | 225,254 |
| `dockerhub-operation-crosswalk.json` | `5bdece35e1a8d7931f97fd3e05d7a515eb2e2f6a4b3a972c9b0154dd11642199` | 100,020 |
| `dockerhub-declaration-disposition.json` | `ca1254eb95a1d0d15fdd9defa6327184d258fe2fc599e132c76ed89dc6db7032` | 160,761 |
| `dockerhub-reverse-etl-action-audit.json` | `2dd132e6f1eddc43db2ab0b09a6d63ad61f4876624f6b27c4fc85f5634634169` | 12,235 |

The lock pins `https://docs.docker.com/reference/api/hub/latest.yaml` at SHA-256 `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756`, captured `2026-08-19T11:28:09Z`, with 54 REST operations: 24 GET, 27 mutations (including 6 DELETE), and 3 HEAD. The source facts include 9 GET rows with paging-shaped parameters and 9 SCIM-media rows.

The local contract treats only a source response whose dereferenced top-level
schema is an array as an extractable collection fact. Docker Hub has two of
those rows (member listing and member export); one overlaps pagination, so
there are 10 independent source-backed ETL/sync candidates. Nested arrays in
arbitrary response objects are retained as source schema detail, not promoted
to executable collection semantics.

## Foundation Atlas and media decision

The Batch R1 Foundation Atlas was consulted before any `missing_foundation` label. `source.retention-import.v1`, `source.projection-admission.v1`, `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, `warehouse.reverse-etl.v1`, and `transport.sync-contract.v1` are catalogued as available, but none is invoked or claimed by this mapping-only delivery. Every matrix cell must therefore be either `mapped_unproven` or source-evidenced `not_applicable`; no runtime foundation gap is named.

The known SCIM `application/scim+json` question is **not** an executor limitation on current main: `internal/connectors/engine/structured_rest_body.go` and `direct_write.go` explicitly accept it as the closed JSON family, with `TestOperationDirectWriteContentTypeAllowsClosedJSONFamily`; `cmd/connectorgen` admits it in `supportedDirectWriteContentType`. The matrix will retain exact SCIM media facts and must not relabel them as binary. No candidate runtime gap is recorded.

`docs/connector-terminology.md`, cited by the build-order skill and issue, is absent from current main and Batch R1 parent. The connector-lane build-order contract is used as the terminology fallback; this is a documentation-reference gap only.

## GSD and skill record

Ran `scripts/gsd doctor` and resolved the required `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts. The task is explicitly single-worker and this runner has no compatible isolated Pi role; the manual inline GSD fallback is recorded here and in the plan/ledger instead of spawning roles.

Loaded skills: `connector-lane-build-order`; `go-engineering` (fundamentals and agentic ETL); repository routing `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.
