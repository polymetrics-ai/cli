# Bitbucket source-to-seven-lane matrix — Track A

## Task Delivery Header

- Issue: Refs #4381 — Bitbucket — source-to-seven-lane matrix.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: A committed and pushed scoped branch, ready for independent review and not integrated into the parent branch.
- Working branch: `feat/4381-bitbucket-track-a-matrix-r1`.
- Task: Add only Bitbucket source-lock-bound seven-lane mapping evidence, its connector-local validator, and issue-scoped planning evidence. Preserve every retained source row; record crosswalk-only identities as a source-boundary reconciliation, not rows.
- Verification: Run the connector-local Go validation test, JSON parsing, focused source/import/projection checks that do not rewrite definitions, `git diff --check`, and a changed-path audit.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| All 297 immutable lock operations remain visible with seven cells | live | The connector-local validator loads the real lock and fails if a matrix ID is missing, duplicated, foreign, or lacks a lane cell. |
| The 34 crosswalk-only method/path identities are visible but not promoted to source rows | live | The validator recomputes the crosswalk-minus-lock method/path set and rejects a missing, extra, or row-promoted boundary identity. |
| Lane classifications remain source-backed and not execution claims | live | The validator cross-checks method, exact POST read exceptions, response `paginated_*` schema refs, binary source signals, and webhook delivery sources against the retained lock; no Track A cell is `implemented`. |
| The real webhook runtime gap stays explicit | live | The validator requires each four source-backed webhook subscription cell to name `cli-webhook-event-surface-foundation-r1` and rejects an executable promotion. |

## Scope and boundaries

- Owned paths: `.planning/phases/issue-4381-bitbucket-source-seven-lane-matrix-r1/**`, `internal/connectors/defs/bitbucket/sources/bitbucket-source-lane-matrix.json`, and `internal/connectors/defs/bitbucket/source_lane_matrix_test.go`.
- Provider-fact authority is only `internal/connectors/defs/bitbucket/sources/bitbucket-operation-source-lock.json`; the crosswalk is used only for its explicitly non-source boundary reconciliation.
- No `connectorgen` output, runtime, shared foundation, definition projection, manual/skill, CLI surface, or other connector files are in scope.
- The matrix records mapping truth, not command/runtime/certification proof. `mapped_unproven` never means executable.

## Source facts and selected mapping policy

The immutable denominator is 297 `rest.operations` in the schema-v2 Bitbucket lock (`sha256` `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3`, 1,359,673 bytes, captured `2026-08-19T11:28:09Z`). Each row repeats source ID, protocol, method, path, provider operation ID, exact source location, scope variables, media, pagination evidence, and event/cursor disposition.

The crosswalk has 331 method/path identities. Its 34 identities absent from the lock are recorded in `source_boundary_reconciliation` using method-plus-path identity; they are neither hidden nor matrix rows. The lock has no method/path identity absent from the crosswalk.

| Lane | Source-only applicability rule | Expected cells |
| --- | --- | --- |
| direct_read | The locked provider operation summary begins with a source-semantic bounded-read action (`get`, `list`, `search`, `compare`, `retrieve`, or `check`). HTTP method and operation ID are retained facts but never classify the lane. | 162 `mapped_unproven`, 135 `not_applicable` |
| direct_write | The locked provider operation summary begins with a source-semantic mutation action (`delete`, `update`, `create`, `add`, `remove`, `unapprove`, `approve`, `watch`, `set`, `upload`, `stop`, `run`, `resolve`, `request`, `reopen`, `merge`, `fork`, `decline`, or `bulk`). | 135 `mapped_unproven`, 162 `not_applicable` |
| binary_download | 13 exact retained source rows whose locked summary/response text cites raw diff/patch/log/file/download/GPG material | 13 `mapped_unproven`, 284 `not_applicable` |
| binary_upload | Two exact retained `POST` source rows that explicitly say upload a download artifact or upload a file | 2 `mapped_unproven`, 295 `not_applicable` |
| etl | Only a locked successful response that resolves to a source-contract schema with both string `next` and array `values`; request page/cursor controls are retained as source evidence. A read method or list summary alone is insufficient. | 73 `mapped_unproven`, 224 `not_applicable` |
| reverse_etl | The same 135 source mutation candidates, evaluated independently from direct write | 135 `mapped_unproven`, 162 `not_applicable` |
| sync_transport | Four source-backed repository/workspace webhook create/update rows whose locked text describes delivery of selected events | 4 `missing_foundation`, 293 `not_applicable` |

## Foundation Atlas discovery

The current Atlas was consulted before mapping:

- `source.retention-import.v1`, `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, and `warehouse.reverse-etl.v1` are available reuse contracts. Track A does not select or prove any executable adoption.
- `transport.sync-contract.v1` is available for registered source/destination transports through a durable warehouse, but it explicitly requires an exact registered source, destination, mode, and conformance intersection. It does not supply a Bitbucket inbound webhook receiver.
- The four webhook subscription cells therefore name the actual gap `cli-webhook-event-surface-foundation-r1`: a closed inbound receiver with Bitbucket HMAC verification, source-cited event scope, durable warehouse acknowledgement/checkpoint, and replay/conformance proof. No Atlas entry owns that contract today. No foundation implementation is authorized here.

## Execution plan

1. Record the source facts and the exact decisions above before production artifacts; use a read-only lock enumerator only to transcribe retained facts into rows. The renderer is mechanical; the source-fact policy above is the decision authority.
2. Add a failing connector-local validator that requires the matrix file and rejects hidden rows, missing cells, and crosswalk boundary drift.
3. Add the matrix, keeping every runtime-related candidate `mapped_unproven` or the four webhook cells `missing_foundation`.
4. Run focused red/green/edge validation and the source/import/projection checks in check-only mode. Do not run a write-mode generator.
5. Review changed paths and diffs, commit the green scoped files, push the issue branch, and post a no-checkbox evidence comment. Do not integrate or open a PR.

## GSD lifecycle trace and manual fallback

`scripts/gsd doctor` succeeded. Every canonical command was resolved through `scripts/gsd sources`: `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`. Generated prompts for all five were inspected and executed inline. This isolated task has no compatible role-spawn runtime and the active agent contract prohibits spawning roles, so the documented inline/manual fallback is used without weakening the red-green-refactor, verification, or review gates.

Required skills loaded for the focused Go validator: `golang-how-to`, `golang-error-handling`, `golang-security`, `golang-structs-interfaces`, `golang-design-patterns`, `golang-safety`, and `golang-testing`; connector procedure and Foundation Atlas guidance were also applied. CLI help/manual/website parity is not applicable because Track A adds no command, flag, runtime surface, generated manual, or website artifact.

---

## Semantic-source repair addendum — 2026-08-31

### Task Delivery Header

- Issue: Refs #4381 — Bitbucket — source-to-seven-lane matrix.
- Base branch: `feat/4381-bitbucket-track-a-matrix-r1` at `d1de8a9dc45ed9f4feab3e92e6b6aa8fd0b2231b`.
- Merges into: `feat/4381-bitbucket-track-a-matrix-r1 → fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: A committed, locally verified, pushed semantic-repair branch with a completion/re-review comment on #4381. It remains unintegrated and no merge is authorized.
- Working branch: `fix/4381-bitbucket-semantic-lane-r1`.
- Task: Repair only the Bitbucket Track A matrix and its connector-local validator so direct-read/direct-write/reverse-ETL classification comes from the locked provider operation semantics rather than an HTTP-method rule, and ETL classification comes from source request/response continuation facts rather than a response-schema-name prefix. Preserve every source row, every documented mutation/delete, the existing binary evidence, source-boundary reconciliation, and the existing webhook sync gap.
- Verification: Add red/green/edge tests for a semantic non-GET read, mutation exclusion from direct read, schema-name-independent pagination, missing continuation, and source-row/count/backlink coverage; run focused and package tests, race and vet checks, JSON/agent-contract/diff checks, and record any pre-existing broader-gate result without altering it.

### Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A non-GET documented list read is directly readable without an HTTP-method allow-list | live | The real locked POST `List commits…` row is classified from its source summary as a read and the focused validator accepts the corresponding cell. |
| A documented mutation/delete stays independently mapped for direct write and reverse ETL | live | The real locked create/delete rows are classified from their source summaries as mutations, rejected as direct reads, and require both write cells. |
| Search-result ETL is classified from request/response continuation facts, not schema spelling | live | Each retained `search_result_page` operation has source-documented page/pagelen query facts and a resolved response schema with `next` and `values`; deleting that continuation evidence makes the focused validator fail. |
| A non-continuable response is not promoted to ETL | live | A synthetic source contract with `values` but no `next` is rejected by the pagination helper. |
| Matrix backlinks and deterministic counts cover the revised denominator | live | The focused validator recomputes all source-lane eligibility from the real 297-row lock and rejects an altered search ETL cell or backlink/count drift. |
| No runtime/certification/foundation behavior changes | live | `git diff --name-only` is restricted to this phase evidence, Bitbucket matrix, and its package-local test; no provider I/O or credential is used. |

### Decision and Atlas disposition

- CodeGraph discovery: no `.codegraph/` directory exists in the frozen target worktree, so the connector-local validator and retained source files were inspected directly.
- The provider source lock already retains both operation nodes and the `source_contract.components.schemas` needed for source-semantic classification. This repair is a connector-local mapping correction, not an importer or runtime-foundation change.
- Atlas consulted: `source.retention-import.v1` and `source.projection-admission.v1` are **reuse** context only; `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, and `warehouse.reverse-etl.v1` are not selected or claimed executable by Track A.
- `transport.sync-contract.v1` remains **actual gap** context for the four existing Bitbucket webhook subscription cells. Pagination never implies sync transport and this repair does not change those cells.
- No new Foundation Atlas entry, source lock, crosswalk, runtime engine, transport, certification, or generated artifact may change in this slice.
- The matrix-level `mapping_policy` is an executable contract: its direct-read, direct-write/reverse-ETL, and ETL text must match these semantic/structural selectors, and connector-local tests reject stale method or schema-name wording.

### Red–Green–Refactor execution

1. **Red:** Add focused assertions against the real lock for POST list-read semantics and the three search-result continuation contracts; assert a synthetic `values`-only response cannot qualify. Expect the current prefix-based ETL helper to miss the three search operations.
2. **Green:** Resolve successful response `$ref`s structurally within the retained `source_contract` and classify `next` plus array `values` as continuation evidence. Derive semantic read/mutation classification from the locked operation summary rather than HTTP method/operation ID. Update only the affected matrix evidence and counts.
3. **Refactor:** Keep the helper bounded, deterministic, and source-only; reject unknown summary actions rather than silently calling them reads or writes. Rerun focused, package, race, vet, JSON, agent-contract, and diff checks.

### GSD and skill trace

`scripts/gsd doctor`, all five `scripts/gsd sources` lookups, and `go run ./cmd/agentcontractgen check` passed. Generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` were inspected. This isolated worker has no compatible Pi GSD execution runtime and may not spawn a role, so the documented inline/manual fallback is used and recorded here.

Required skills loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, the shared `go-engineering` guidance, and `connector-lane-build-order`. CLI help/manual/website parity remains not applicable because no command, flag, generated help, manual, or website artifact changes.
