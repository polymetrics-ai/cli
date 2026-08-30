# Vercel source-to-seven-lane matrix — Track A

## Task Delivery Header

- Issue: Refs #4421 — Vercel — source-to-seven-lane matrix.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` only after independent review and parent integration.
- Delivery: a committed, pushed scoped branch ready for review; this task does not integrate or open a PR.
- Working branch: `feat/4421-vercel-track-a-matrix-r1`.
- Task: add only Vercel source-lock-bound seven-lane mapping evidence, a connector-local validator, and issue-scoped planning evidence.
- Verification: focused connector-local Go validation, JSON parsing, non-mutating source/import/projection checks where available, `git diff --check`, and a changed-path audit.

## Scope and boundaries

- Owned paths: `.planning/phases/issue-4421-vercel-source-seven-lane-matrix-r1/**`, `internal/connectors/defs/vercel/sources/vercel-source-lane-matrix.json`, and `internal/connectors/defs/vercel/source_lane_matrix_test.go`.
- Provider-fact authority is only `internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json`. The crosswalk, existing definitions, and stream declarations are backlinks only; they never decide executable lane status.
- No shared engine, `connectorgen` logic, generated projection, runtime, Foundation Atlas, CLI surface, manual, skill, or other connector file is in scope.
- The matrix preserves mapping truth, not command/runtime/certification proof. `mapped_unproven` never means executable.

## Source facts and selected mapping policy

The immutable denominator is 400 `rest.operations` in the schema-v2 Vercel lock (`sha256` `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28`, 10,463,249 bytes, captured `2026-08-19T11:28:09Z`). Its methods are 159 `GET`, 120 `POST`, 56 `DELETE`, 43 `PATCH`, 18 `PUT`, and four `HEAD`. Each matrix row repeats exact source identity/citation, scope variables, request and success-response media, selected pagination facts, and event/cursor facts.

The current Vercel crosswalk has 400 exact source identities and 22 API-surface-only method/path identities. The latter remain explicit `not_source_row` boundary records; they do not affect the source denominator.

| Lane | Source-only applicability rule | Expected cells |
| --- | --- | --- |
| direct_read | Every locked `GET` row; a verb candidate only | 159 `mapped_unproven`, 241 `not_applicable` |
| direct_write | Every locked `POST`, `PUT`, `PATCH`, or `DELETE` row; a mutation-verb candidate only | 237 `mapped_unproven`, 163 `not_applicable` |
| binary_download | No retained successful response has selected binary-download media evidence | 400 `not_applicable` |
| binary_upload | `uploadArtifact`, `uploadProjectAvatar`, and `uploadFile`, each with retained `application/octet-stream` request media | 3 `mapped_unproven`, 397 `not_applicable` |
| ETL through DuckDB | The 22 GET rows with a required `pagination` response wrapper and either `limit` + (`since`/`until`/`cursor`/`from`) or `page` + `per_page` retained query controls | 22 `mapped_unproven`, 378 `not_applicable` |
| reverse ETL from DuckDB | The same 237 mutation-verb candidates, independently evaluated from direct write | 237 `mapped_unproven`, 163 `not_applicable` |
| sync transport | `vercel.rest.createWebhook`, whose source request body requires `url` and `events` | 1 `missing_foundation`, 399 `not_applicable` |

The four selected ETL candidates that exactly backlink to existing Vercel streams are `getDomains → domains`, `getProjects → projects`, `getTeams → teams`, and `listAliases → aliases`. They remain `mapped_unproven`; the other 18 selected ETL candidates require a future source-preserving stream projection. Existing `/v6/deployments` and `/v1/edge-config` stream definitions are visibly boundary-only rather than source-lock operations.

## Foundation Atlas discovery

- `source.retention-import.v1`, `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, and `warehouse.reverse-etl.v1` are available reuse references for later proof work. Track A does not select or prove executable adoption.
- `transport.sync-contract.v1` is available for registered source/destination/mode/conformance intersections through DuckDB. Its relevant owners are `internal/connectors/sync_transport.go#SyncTransportDescriptor` and `internal/synctransport/orchestrator.go#(*Orchestrator).Run`.
- The retained Vercel webhook-registration source row is a genuine missing foundation: `cli-webhook-event-surface-foundation-r1`, a closed inbound Vercel webhook receiver and DuckDB-mediated registered source executor. The matrix records exact source evidence, Atlas lookup, insufficiency, owner symbols, and a proof idea. No foundation implementation is authorized; it awaits captain approval.

## Execution plan

1. Record the source facts and lane policy before production artifacts. A mechanical renderer transcribes locked rows; the stated source-fact policy—not a generator—makes mapping decisions.
2. Add a failing connector-local validator that rejects hidden rows, missing cells, crosswalk-boundary drift, stream backlink drift, and executable-disposition promotion.
3. Add the matrix with no `implemented` cells; preserve `createWebhook` as its single `missing_foundation` sync cell.
4. Run focused red/green/edge validation and check-only source/import/projection checks. Do not run a write-mode generator.
5. Review only owned paths, commit the scoped green files, push the issue branch, and post a no-checkbox proof comment. Do not integrate or open a PR.

## GSD lifecycle trace and manual fallback

`scripts/gsd doctor` succeeded. The canonical source list and generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` are being followed inline. The active agent contract prohibits role spawning, so the documented inline/manual fallback is used without weakening red-green-refactor, verification, or review gates. The connector-lane build procedure and required Go guidance were loaded before implementation.
