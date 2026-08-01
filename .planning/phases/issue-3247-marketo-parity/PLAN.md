# PLAN — issue #3247 Marketo parity final wave

## GSD command path and fallback

- `scripts/gsd doctor` passed on 2026-08-01 in the disposable worktree.
- `scripts/gsd list` passed and reported 69 commands.
- `scripts/gsd prompt programming-loop init --phase issue-3247-marketo-parity --dry-run` failed with `unknown GSD command: programming-loop` even though repo docs reference it.
- Manual GSD fallback is active using `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; no production edits start before this plan/TDD/verification ledger.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-testing`
- `golang-context`
- `golang-concurrency`
- `golang-cli`
- `golang-documentation`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`

## Scope

Connector-local Marketo parity only:

- `internal/connectors/defs/marketo/**`
- Marketo-owned fixtures/docs/CLI metadata generated from official AdobeDocs Marketo Swagger assets
- Dedicated Marketo tests that assert the generated operation ledger and command surface counts
- GSD phase artifacts for this worker

Out of scope: shared runtime behavior, new dependencies, live provider calls, credentials, infrastructure, no-mistakes pipeline, pushing, PR updates, and merge activity.

## Official sources

- `https://raw.githubusercontent.com/AdobeDocs/marketo-developer.en/main/help/rest-api/assets/swagger-asset.json`
- `https://raw.githubusercontent.com/AdobeDocs/marketo-developer.en/main/help/rest-api/assets/swagger-identity.json`
- `https://raw.githubusercontent.com/AdobeDocs/marketo-developer.en/main/help/rest-api/assets/swagger-mapi.json`
- `https://raw.githubusercontent.com/AdobeDocs/marketo-developer.en/main/help/rest-api/assets/swagger-user.json`

Expected official operation inventory from fresh Swagger fetch: 327 operations = asset 166 + identity 2 + mapi 147 + user 12.

## Implementation slices

Status: all local implementation/validation slices complete. Final issue update, local commit, and external status append remain.

1. **Red validation slice**
   - Add a Marketo dedicated count/contract test that currently fails on the 7-row legacy-parity `api_surface.json`.
   - Capture the failing result in `TDD-LEDGER.md`.

2. **Generated definition slice**
   - Generate Marketo `streams.json`, stream schemas, stream fixtures, `writes.json`, `operations.json`, `cli_surface.json`, `api_surface.json`, `certification.json`, and `docs.md` from the official Swagger assets.
   - Keep `spec.json` credential-safe and root-host-based; no secret literals, no credentials, no provider calls.
   - Implement supported JSON GET operations as either fixture-backed ETL streams or bounded direct reads.
   - Implement supported POST/DELETE mutations as named typed reverse-ETL write actions with closed schemas, path/query/body fields, redaction where path identifiers are sensitive, and destructive confirmation for delete/cancel/remove/purge/discard/unapprove/deactivate operations.
   - Keep true binary/InputStream downloads blocked in the operation ledger with official source evidence because the current executable direct-read path is JSON-only.
   - Keep write-query-selector operations blocked until the shared write contract supports a structured typed query map; no write action embeds record values in URL query strings.
   - Represent identity token issuance as a non-data/disallowed operation row, not an executable command.

3. **Validation/fix slice**
   - Run `gofmt` on dedicated test files.
   - Run focused validation, conformance, CLI golden/dynamic tests, build, boundary guard, `git diff --check`, and `make verify`.
   - Fix connector-local failures without weakening shared gates.

4. **Issue update and commit slice**
   - Use `gh-axi` exactly once for the parent/subissue final update after local gates.
   - Commit the clean tested branch.
   - Append final `done:` status with commit hash, counts, and blockers.

## Orchestration decision

Cycle `plan`: `local_critical_path` — this firstmate worker owns a single connector-local write scope in one disposable worktree. Mutating subworkers would require additional isolated worktrees and would exceed the final-wave single-worker setup. Read-only sidecars are not necessary for the critical path because the official Swagger assets are machine-parseable and count-verifiable locally.

## Safety notes

- No live Marketo calls; only public AdobeDocs/GitHub raw Swagger fetches.
- No secrets requested, printed, stored, or fixture-authored.
- No generic HTTP method/path/body, shell, raw query, file, or passthrough command.
- Reverse ETL writes remain `plan -> preview -> approval -> execute`; destructive actions require typed confirmation.
- Binary/InputStream downloads remain blocked/planned until a shared bounded binary transfer executor exists.
