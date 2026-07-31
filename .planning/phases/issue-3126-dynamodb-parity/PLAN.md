# GSD Plan — issue 3126 DynamoDB parity wave 04

## Context

- Branch: `fm/cli-dynamodb-parity-wave04-r1` in disposable worktree `/Users/karthiksivadas/.treehouse/cli-83d592/41/cli`.
- Primary issue: #3126. Capability children: #3127-#3133, fetched with `gh-axi issue view <n> --full -R polymetrics-ai/cli` before edits.
- GSD command evidence: `scripts/gsd doctor` passed. `scripts/gsd prompt programming-loop ...` is unavailable in the current repo-local registry (`unknown GSD command: programming-loop`), so this phase uses the repo-local `/gsd-plan-phase` prompt path plus manual programming-loop/TDD artifacts. Prompt generated with `scripts/gsd prompt plan-phase issue-3126-dynamodb-parity --skip-research --skip-verify`.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-spf13-cobra`, `context-mode`.

## Objective

Move `dynamodb` from legacy Scan-only native parity to a truthful complete official API ledger and connector-local implementation for every official DynamoDB/DynamoDB Streams operation supported by the existing contract:

- ETL streams for bounded read/list/describe/read-transaction operations.
- Bounded direct/provider reads for point/keyed/query operations.
- CDC/changefeed support for DynamoDB Streams operation surfaces.
- Typed reverse-ETL write actions with schema validation, dry-run previews, plan → preview → approval → execute semantics, destructive confirmation notes, and idempotency documentation.
- Binary/import/export operations are not executed unless the current shared runtime has a binary operation executor; if absent, ledger them as blocked with exact shared-runtime dependency.
- Raw PartiQL/query, unrestricted scans, arbitrary expressions/bodies, generic HTTP/shell/file/passthrough escapes remain disallowed.

## Source/audit plan

1. Re-audit official Smithy models from AWS SDK Go v2 (`dynamodb.json`, `dynamodb-streams.json`) and official operation list docs.
2. Reconcile 57 DynamoDB + 4 DynamoDB Streams operations against parent lane counts: 23 ETL/read, 26 reverse ETL write, 3 direct query/search, 2 binary, 4 CDC, 3 excluded = 61 total.
3. Update `internal/connectors/defs/dynamodb/api_surface.json` with one row per official operation and exact executable/blocked/excluded disposition.

## Implementation slices

1. **Red tests / characterization**: add native DynamoDB tests for all operation classifications, stream targets, write schemas/previews/execution, operation direct reads, CDC changefeed helpers, and fixture inventory. Tests use `httptest.Server` only; no AWS/provider calls.
2. **Connector defs**: expand `metadata.json`, `spec.json`, `streams.json`, schemas, `writes.json`, `operations.json`, `cli_surface.json`, fixtures, docs/manual/skill to the complete ledger.
3. **Native execution**: extend `internal/connectors/native/dynamodb` only. Add generic signed JSON-RPC helper and closed operation registry. Implement bounded streams, direct reads, CDC operations, ValidateWrite, DryRunWrite, Write. Keep SigV4 and endpoint safety; no shared runtime edits.
4. **Docs/generated surfaces**: update DynamoDB connector docs and generated connector docs/skills/website data if the repo provides connector-local generators; avoid unrelated churn.
5. **Issue addendum**: append idempotent captain-policy addendum to all eight issues with truthful counts after local verification.

## Safety constraints

- No live provider calls, credentials, secrets, writes, certification, VPS, Thaalam, Herdr lifecycle, or daemon restarts.
- No new dependencies.
- No shared engine/CLI behavior edits unless an existing local test proves connector-local code cannot meet the contract; stop before shared runtime behavior changes.
- Reverse ETL execution remains gated by existing plan/preview/approval/execute flows; local tests call connector `Write` only against `httptest.Server` with synthetic payloads.

## Expected final counts

Target post-change ledger, pending implementation verification:

| total | implemented | fixture-tested | blocked/planned | excluded/not-applicable | certified |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 61 | 56 | 56 | 2 | 3 | 0 |

Blocked/planned: `ExportTableToPointInTime`, `ImportTable` (binary/import-export shared-runtime executor absent). Excluded/not-applicable: `BatchExecuteStatement`, `ExecuteStatement`, `ExecuteTransaction` (raw PartiQL statement surfaces disallowed by task and Connector Guard).

## Verification checklist

See `VERIFICATION.md`. Required gates from brief: focused connectorgen validation/conformance for `dynamodb`, focused CLI tests, `go build ./cmd/pm`, `make connector-boundary`, `make verify`, `git diff --check`.
