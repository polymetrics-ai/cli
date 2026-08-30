# Jira source-to-seven-lane matrix — Track A

## Task Delivery Header

- Issue: Refs #4402 — Jira — source-to-seven-lane matrix.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` only after independent review and parent integration.
- Delivery: A committed, pushed scoped branch ready for review; this task does not integrate or open a PR.
- Working branch: `feat/4402-jira-track-a-matrix-r1`.
- Task: Add only Jira source-lock-bound seven-lane mapping evidence, a connector-local validator, and issue-scoped planning evidence.
- Verification: Run the focused connector-local Go validator, JSON parsing, non-mutating source/import/projection checks where available, `git diff --check`, and a changed-path audit.

## Scope and boundaries

- Owned paths: `.planning/phases/issue-4402-jira-source-seven-lane-matrix-r1/**`, `internal/connectors/defs/jira/sources/jira-source-lane-matrix.json`, and `internal/connectors/defs/jira/source_lane_matrix_test.go`.
- Provider-fact authority is only `internal/connectors/defs/jira/sources/jira-operation-source-lock.json`. The existing crosswalk and `streams.json` are only connector artifact backlinks and never decide an executable lane.
- No shared engine, `connectorgen` logic, generated projection, runtime, Foundation Atlas, CLI surface, manual, skill, or other connector file is in scope.
- The matrix preserves mapping truth, not command/runtime/certification proof. `mapped_unproven` never means executable.

## Source facts and selected mapping policy

The immutable denominator is 617 `rest.operations` in the schema-v2 Jira lock (`sha256` `e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf`, 2,456,011 bytes, captured `2026-08-19T11:28:09Z`). The lock contains 276 `GET`, 134 `POST`, 118 `PUT`, and 89 `DELETE` rows. Each matrix row repeats its exact source ID, method, path, source location, scope variables, request/success media, source paging facts, and event/cursor facts.

The current Jira crosswalk is an exact 617-to-617 method/path reconciliation: 617 source operations, 617 API-surface endpoints, zero source-only and zero surface-only identities. It is recorded as a boundary backlink, not a second source authority.

| Lane | Source-only applicability rule | Expected cells |
| --- | --- | --- |
| direct_read | Every locked `GET` row; a verb candidate only | 276 `mapped_unproven`, 341 `not_applicable` |
| direct_write | Every locked non-`GET` row (`POST`, `PUT`, `DELETE`) as a mutation-verb candidate | 341 `mapped_unproven`, 276 `not_applicable` |
| binary_download | The three avatar image GET rows whose successful source media includes `image/png` and `image/svg+xml` | 3 `mapped_unproven`, 614 `not_applicable` |
| binary_upload | `addAttachment` with `multipart/form-data`, plus the three avatar loads whose wildcard request body and source response text cite an image | 4 `mapped_unproven`, 613 `not_applicable` |
| ETL through DuckDB | Only a locked `GET` operation with a query parameter exactly named `maxResults` | 95 `mapped_unproven`, 522 `not_applicable` |
| reverse ETL from DuckDB | The same 341 non-`GET` mutation-verb candidates, evaluated independently from direct write | 341 `mapped_unproven`, 276 `not_applicable` |
| sync_transport | `jira.rest.registerDynamicWebhooks`, whose retained request example has `url` and `webhooks[].events`, through the required DuckDB warehouse stage | 1 `missing_foundation`, 616 `not_applicable` |

The pre-existing `streams.json` names only three exact paths in the 95-row ETL cohort: `jira.rest.searchForIssuesUsingJql → issues`, `jira.rest.searchProjects → projects`, and `jira.rest.getAllUsers → users`. They stay `mapped_unproven`; the other 92 cells require a future source-preserving stream projection. The matrix does not convert a legacy stream backlink into runtime proof.

## Foundation Atlas discovery

- `source.retention-import.v1`, `runtime.direct-execution.v1`, `warehouse.stage-etl.v1`, and `warehouse.reverse-etl.v1` are available reuse references for future proof work. Track A does not select or prove an executable adoption.
- `transport.sync-contract.v1` is available for a registered source/destination/mode/conformance intersection through the warehouse. Its owner symbols are `internal/connectors/sync_transport.go#SyncTransportDescriptor` and `internal/synctransport/orchestrator.go#(*Orchestrator).Run`.
- The Jira webhook-registration source row is therefore a genuine missing foundation: `cli-webhook-event-surface-foundation-r1`, a closed inbound Jira webhook receiver and DuckDB-mediated registered source executor. The matrix names its source ID, Atlas lookup, insufficiency, owner symbols, and proof idea. No foundation implementation is authorized in this task; any such work awaits captain approval.

## Execution plan

1. Record source facts and the table above before production artifacts. The mechanical renderer transcribes the locked rows; the stated source-fact policy, not a generator, makes the lane decisions.
2. Add a failing connector-local validator that requires the matrix and rejects hidden rows, missing cells, missing legacy ETL backlinks, and executable disposition promotion.
3. Add the matrix with no `implemented` cells; keep the exact webhook registration as `missing_foundation`.
4. Run focused red/green/edge validation and check-only source/import/projection checks. Do not run a write-mode generator.
5. Review only the owned paths, commit the scoped green files, push the issue branch, and post a no-checkbox proof comment. Do not integrate or open a PR.

## GSD lifecycle trace and manual fallback

`scripts/gsd doctor` succeeded. The canonical source list and prompts for `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` were inspected. The active agent contract prohibits role spawning, so the documented inline/manual fallback is used without weakening red-green-refactor, verification, or review gates. The connector-lane build procedure and required Go guidance were read before implementation.
