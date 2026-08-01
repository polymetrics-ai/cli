# Plan: Gong Official API Parity Completion (#2997)

Issue: #2997
Subissues: #2998, #2999, #3000, #3001, #3002, #3003, #3004
Branch: `fm/cli-gong-parity-wave01-r1`
Worker role: firstmate crewmate, local critical path.

## GSD / skills

- Isolation preflight: `pwd -P` and `git rev-parse --show-toplevel` both resolved to `/Users/karthiksivadas/.treehouse/cli-83d592/12/cli`.
- GSD preflight: `scripts/gsd doctor` and `scripts/gsd list` passed.
- Required command attempted: `scripts/gsd prompt programming-loop init --phase issue-2997-gong-official-api-parity --dry-run` failed with `scripts/gsd: unknown GSD command: programming-loop` in this adapter. Manual-GSD fallback is active for this phase; do not claim the adapter programming-loop command ran.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`; connector parity references and CLI help/docs/website parity reference read.
- Parent orchestration decision: `local_critical_path` for this cycle. All seven subissues target the same Gong connector bundle and command metadata, so parallel mutating workers would collide in one connector-owned write scope. No worker spawned.

## Objective

Complete connector-local Gong parity against the official unauthenticated Gong OpenAPI source without live credentials or provider calls:

- refresh the official operation inventory;
- ensure every documented operation is represented exactly once in `api_surface.json` and `operations.json`;
- add typed bounded command metadata and write schemas for missing operations;
- keep destructive/DELETE operations in scope through typed `destructive` confirmation and reverse ETL plan -> preview -> explicit approval -> execute;
- update fixtures, docs, and connector-owned tests without editing shared runtime files.

## Fresh source inventory

Official source fetched unauthenticated for inventory only: `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`.

Initial source result observed in this worktree: OpenAPI `3.0.1`, `info.version` `V2`, 59 paths and 69 operations:

- GET: 29
- POST: 28
- PUT: 8
- PATCH: 1
- DELETE: 3

Compared with the pre-edit connector `api_surface.json` (67 rows), the official source included two missing operations:

- `GET /v2/targets` (`listTargetDefinitions`) — bounded direct read with required `workspaceId` query flag.
- `POST /v2/targets/{targetId}/assignments` (`uploadAssignments`) — typed multipart reverse ETL upload with required `targetId`, `workspaceId`, optional `validateOnly`, and required CSV file part `file`.

Review-ready parity checkpoint on 2026-08-01 re-fetched the same official OpenAPI source and matched the connector ledger exactly: 69 official operations, 69 `api_surface.json` rows, 0 missing, 0 extra, 0 excluded/planned/blocked rows. Coverage counts are 12 streams, 30 direct reads, and 27 write actions. The checkpoint fixed connector-local discrepancies found during the re-audit: stale `/v2/` write-action paths under a `/v2` base URL, `{meetingId}`/`{taskId}` write paths that needed engine `{{ record.* }}` interpolation, missing required CRM entity-schema query inputs, missing required calls users-access body fields, and non-empty typed direct-read flags including `targets list --workspaceId`.

Gong's optional `validateOnly` query parameter on `POST /v2/targets/{targetId}/assignments` remains intentionally unexposed because the shared write-action path dialect lacks optional query/default support for write paths. Exposing `{{ record.validateOnly }}` would make the provider-optional parameter mandatory. The canonical upload operation is executable with Gong's default `validateOnly=false`, typed `targetId`/`workspaceId`/CSV inputs, destructive confirmation, and reverse ETL plan -> preview -> explicit approval -> execute.

This does not update GitHub issue count tables. The required captain-policy addendum was appended to #2997 and #2998-#3004 without changing existing counts.

## Scope / allowed files

Primary allowed paths:

- `.planning/phases/issue-2997-gong-official-api-parity/**` (GSD/TDD evidence)
- `.planning/traces/**` for the failed programming-loop adapter trace if needed
- `internal/connectors/defs/gong/**`
- `cmd/connectorgen/gong_api_surface_test.go` (connector-owned operation-ledger test)

No shared runtime files, dependencies, live provider calls, credentials, provider writes, certification, VPS/Thaalam/Herdr lifecycle work, pushes to `main`, or merges.

## Implementation slices

1. **Red inventory/test slice**
   - Update/add connector-owned assertions to expect the fresh official count of 69 operations and the two target endpoints.
   - Run a targeted red test before changing the bundle; expected failure is current count 67 / missing target rows.

2. **Definition slice**
   - Add `target_definitions` schema and fixture for `GET /v2/targets`.
   - Add bounded direct-read command metadata for `targets list`.
   - Add `upload_target_assignments` typed multipart write action, command metadata, fixture, `api_surface.json` row, and complete `operations.json` row.
   - Expand `operations.json` from its partial 16-row command ledger to a complete 69-row official-operation ledger.
   - Update docs, metadata, and count/evidence text to say fixture-only and certified=0.

3. **Validation slice**
   - Run `go run ./cmd/connectorgen validate internal/connectors/defs/gong`.
   - Run targeted connector/conformance and connector-owned tests.
   - Run boundary/docs checks appropriate to connector-local command metadata.

## Safety requirements

- No secret values in fixtures, docs, issue bodies, test output summaries, or command metadata.
- Do not add dependencies.
- Do not run live Gong calls or credentialed checks.
- No generic raw HTTP/query/write shell is exposed; every command/action is fixed-target and typed.
- Destructive/delete/write/file-upload actions remain in scope only with typed `destructive` confirmation where applicable and reverse ETL plan -> preview -> explicit approval -> execute.
- If a shared confirmation foundation is missing, record the exact dependency and keep connector-local metadata/tests truthful instead of claiming execution completion.

## Verification checklist

See `VERIFICATION.md` for commands and outcomes.
