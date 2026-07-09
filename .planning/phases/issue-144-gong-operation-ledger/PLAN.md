# Plan: Gong Operation Ledger (#144)

Issue: #144
Parent issue: #133
Parent branch: `feat/133-gong-cli-parity`
Suggested sub-branch: `feat/144-gong-operation-ledger`
Active mode: local critical path inside parent branch because this Pi API session has no `subagent` tool.

## GSD / skills

- GSD preflight: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json` passed.
- GSD plan prompt: `scripts/gsd prompt plan-phase 133 --skip-research --tdd` rendered.
- Programming-loop adapter gap: `scripts/gsd prompt programming-loop init --phase issue-133-gong-cli-parity --dry-run` failed (`unknown GSD command: programming-loop`). Manual-GSD fallback recorded in parent plan.
- Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-lint`.

## Objective

Replace Gong's stale 10-row `api_surface.json` with an exact operation ledger for the public Gong OpenAPI 3.0.1 spec:

- Source: `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=`
- Fetched unauthenticated on 2026-07-09.
- 57 paths, 67 operations: GET 28, POST 27, PUT 8, PATCH 1, DELETE 3.

## Scope

Allowed files:

- `.planning/phases/issue-144-gong-operation-ledger/**`
- `.planning/phases/issue-133-gong-cli-parity-parent/**`
- `cmd/connectorgen/gong_api_surface_test.go`
- `internal/connectors/defs/gong/api_surface.json`
- `internal/connectors/defs/gong/docs.md`

Out of scope for this slice:

- `writes.json` execution or reverse ETL execution.
- Direct-read executor implementation.
- Binary upload/download executor implementation.
- Raw generic HTTP/GraphQL/shell/SQL write surfaces.
- Live Gong credentials or authenticated app calls.

## Classification target

Current stale exclusions must be removed or corrected:

- `/v2/calls/extensive` is `POST`, read-query candidate (#145/#146), not `GET` out-of-scope.
- `/v2/calls/transcript` is `POST`, read-query candidate (#145/#146), not `GET` out-of-scope.
- `/v2/stats/interaction` is `POST`, read-query/stat query candidate (#145/#146), not `GET` out-of-scope.
- `/v2/stats/activity/trackers` is not present in the exact public spec; do not carry it as an endpoint.
- `/v2/workspaces` and `/v2/library/folders` remain official `GET` read candidates.
- `/v2/settings/webhooks` is not present in the exact public spec; do not carry it as an endpoint.

Every official operation must have exactly one of:

- existing `covered_by.stream` for `users`, `calls`, `scorecards`; or
- metadata-only `operation` with `blocked_by_default: true`, non-empty reason, source URL/notes where required, and an issue-linked path to #145/#146/#147.

No legacy `excluded` rows remain in ledger mode.

## Red / green / refactor plan

1. Add `cmd/connectorgen/gong_api_surface_test.go` with expected spec counts, no legacy exclusions, corrected POST read-query paths, and stale-path absence.
2. Run red test; it must fail against current 10-row `api_surface.json`.
3. Replace `internal/connectors/defs/gong/api_surface.json` with operation-ledger version 1 covering all 67 operations.
4. Update `internal/connectors/defs/gong/docs.md` known limits to state exact spec coverage and metadata-only blocked rows without overclaiming execution.
5. Run targeted green tests and validator.
6. Run broader `go run ./cmd/connectorgen validate internal/connectors/defs` before handoff.

## Expected test assertions

- `operation_ledger_version == 1`.
- `len(endpoints) == 67`.
- total methods: GET 28, POST 27, PUT 8, PATCH 1, DELETE 3.
- covered stream endpoints: 3 (`GET /v2/users`, `GET /v2/calls`, `GET /v2/settings/scorecards`).
- operation rows: 64.
- legacy `excluded` rows: 0.
- every operation row is blocked by default with non-empty reason.
- stale paths `/v2/stats/activity/trackers` and `/v2/settings/webhooks` absent.
- `/v2/calls/extensive`, `/v2/calls/transcript`, `/v2/stats/interaction` present as `POST` operation rows.

## Human gates

- No live Gong credentials.
- No write execution.
- No binary payload transfer.
- No new dependencies.
- No parent PR merge to `main`.
