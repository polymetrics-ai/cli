# Issue 156 — Zendesk Support connector parity plan

## Context

- Parent: https://github.com/polymetrics-ai/cli/issues/156; subissues #157-#163.
- Branch: `fm/cli-zendesk-support-parity-wave01-r1` in isolated treehouse worktree.
- Write scope: connector-local Zendesk Support bundle metadata and fixtures under `internal/connectors/defs/zendesk-support/**`, plus required GSD artifacts, generated docs/website surfaces needed to expose connector-owned operation metadata, and lint-only shared Go housekeeping from the documentation/lint pass.
- Official sources inventoried: Zendesk Support introduction and Support API OAS 2.0.0 at `https://developer.zendesk.com/zendesk/oas.yaml`.
- Current local snapshot before edits: 33 streams, 27 write actions, no `operations.json`, no `cli_surface.json`, 76 api-surface rows. Official OAS has 625 operations.

## GSD command path

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt plan-phase issue-156-zendesk-support-parity --skip-research` generated the planning prompt used here.
- Required `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-support-parity --dry-run` was attempted first, but this repo-local adapter currently reports `unknown GSD command: programming-loop`. Manual GSD programming-loop fallback is recorded here and in the ledger; do not claim the unavailable command ran.
- `scripts/gsd prompt quick "issue-156 Zendesk Support parity connector-local implementation"` generated the execution prompt used as the local critical-path fallback.
- `scripts/gsd prompt quick "issue-156 Zendesk Support parity checkpoint re-audit"` generated the checkpoint prompt for the 2026-08-01 source-backed re-audit.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-documentation`
- `golang-cli`
- CLI help/docs/website parity reference

## Parent orchestration decision

`local_critical_path`: firstmate launched this worker in an isolated worktree with one connector-wide write scope. The seven subissues all collide on `internal/connectors/defs/zendesk-support/**`, so no mutating subagents are spawned from this checkout. Read-only research may be done inline or with tools; implementation remains local.

## Implementation slices

1. **Issue metadata addendum**
   - Append an idempotent captain-policy addendum to #156-#163 with `gh-axi`.
   - Preserve existing issue bodies and count tables.
   - State destructive/delete operations are in scope with typed `destructive` confirmation; no fabricated implementation counts.

2. **Official operation inventory and ledger**
   - Fetch and parse Zendesk Support OAS 2.0.0 without live credentials.
   - Generate a complete `api_surface.json` ledger with every official operation exactly once.
   - Map existing streams and writes to official rows where equivalent.
   - Add blocked `operation` rows for not-yet-executable direct, binary, CDC/changefeed, admin, sensitive, and destructive operations; avoid legacy blanket `excluded` rows in operation-ledger mode.
   - Add supplemental rows only where existing connector streams/writes are not present in the Support OAS but must remain covered for validation.

3. **Typed operation and command metadata**
   - Add `operations.json` with fixed, connector-scoped typed REST/binary operation contracts for blocked ledger rows.
   - Add `cli_surface.json` with provider-style command metadata that references streams, writes, and blocked operations without exposing raw method/path/body escape hatches.
   - For destructive/delete operations, set explicit risk/approval text and typed-confirmation notes.

4. **Docs and conformance evidence**
   - Update `docs.md` overview/known limits with official ledger counts and blocked-by-default safety.
   - Preserve existing stream/write fixtures; add fixtures only if new executable streams or write actions are introduced.
   - Run connector-local validation and conformance.

5. **Parity checkpoint re-audit**
   - Re-fetch the official Zendesk Support OAS with a browser-style user agent, parse OpenAPI 3.0.3 / `info.version` 2.0.0, and compare method/path keys against `api_surface.json`, `operations.json`, `cli_surface.json`, streams, and writes.
   - Treat missing, stale, duplicate, unclassified, unknown covered refs, missing command refs, missing blocked-operation metadata, or destructive/delete safety gaps as findings to fix before review-ready.
   - Record truthful post-audit counts without claiming live certification or additional implemented operations.

## Safety gates

- No live credentials, provider calls, external writes, certification, VPS/Thaalam, dependencies, behavior-changing shared runtime edits, or merges.
- No secrets in fixtures, docs, metadata, command examples, or GitHub issue edits.
- Reverse ETL remains plan → preview → explicit approval → execute; destructive operations require typed `destructive` confirmation before execution.
- Shared executor gaps are recorded as blocked operation metadata rather than claimed as implemented. File-upload rows are limited to the currently supported connector-local `direction`/`path`/`max_bytes` metadata, with source-backed method/query/content-type/body contracts recorded in operation descriptions pending a shared file-operation schema/validator.

## Expected checkpoint commits

1. Plan/addendum/ledger generation checkpoint after local validation.
2. Follow-up fix checkpoint only if validation/review finds connector-local issues.
