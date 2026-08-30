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
| direct_read | 160 locked `GET` rows plus the exact provider summaries `List commits with include/exclude` and `List commits for revision using include/exclude` on two `POST` rows | 162 `mapped_unproven`, 135 `not_applicable` |
| direct_write | Every source mutation candidate: 49 `DELETE`, 44 `PUT`, and 42 `POST` rows after those two exact documented list-read exceptions | 135 `mapped_unproven`, 162 `not_applicable` |
| binary_download | 13 exact retained source rows whose locked summary/response text cites raw diff/patch/log/file/download/GPG material | 13 `mapped_unproven`, 284 `not_applicable` |
| binary_upload | Two exact retained `POST` source rows that explicitly say upload a download artifact or upload a file | 2 `mapped_unproven`, 295 `not_applicable` |
| etl | Only a locked successful response that references `#/components/schemas/paginated_*`; GET alone is insufficient | 70 `mapped_unproven`, 227 `not_applicable` |
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
