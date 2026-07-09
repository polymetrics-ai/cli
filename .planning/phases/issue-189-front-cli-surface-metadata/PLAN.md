# Plan: Front CLI Surface Metadata

Parent issue: #188
Sub-issue: #189
Parent PR: https://github.com/polymetrics-ai/cli/pull/224
Parent branch: `feat/188-front-cli-parity`
Sub-issue branch: `feat/189-front-cli-surface-metadata`
Connector: `front`

## GSD command path

- Preflight inherited from parent: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json`.
- Planning prompt used: `scripts/gsd prompt plan-phase 189 --skip-research --tdd`.
- Programming-loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-189-front-cli-surface-metadata --dry-run` — unavailable (`unknown GSD command: programming-loop`).
- Manual GSD fallback: use `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` directly.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-spf13-cobra`
- `golang-spf13-viper`
- `golang-lint`

## Objective

Produce and validate Front CLI/API surface metadata from official Front docs, mapping provider
operations into safe Polymetrics app intents without generic raw HTTP/write escape hatches.

## Scope

Allowed files for this slice:

- `internal/connectors/defs/front/api_surface.json`
- `internal/connectors/defs/front/cli_surface.json`
- `internal/connectors/defs/front/metadata.json` when capability/risk metadata must reflect the surface inventory.
- `internal/connectors/defs/front/docs.md` only for metadata/scope notes needed by validation.
- `.planning/phases/issue-189-front-cli-surface-metadata/**`

Out of scope for this slice unless needed by validation only:

- Implementing new stream schemas/fixtures.
- Implementing write actions.
- Implementing direct-read/binary executors.
- Runtime help renderer changes.
- Website/generated docs changes.

## Current baseline

- Official baseline from parent issue: 342 operations: GET 216, POST 69, PATCH 31, DELETE 26.
- Current `api_surface.json`: 10 endpoint entries.
- Current implemented streams: `contacts`, `conversations`, `inboxes`, `tags`, `teammates`, `channels`.
- Current writes: none.
- Current `cli_surface.json`: absent.

## Implementation slices

### Slice 1 — red validation and source capture

1. Capture a failing focused check proving the current Front metadata is incomplete against the official baseline.
2. Store non-secret source notes under this phase directory; do not commit raw huge public docs unless needed.
3. Confirm `go run ./cmd/connectorgen validate internal/connectors/defs` remains the static gate.

### Slice 2 — safe CLI surface metadata

1. Add `internal/connectors/defs/front/cli_surface.json` with provider-inspired Front command groups.
2. Mark implemented commands only where they resolve to existing streams.
3. Mark unimplemented write/direct-read/binary/admin commands as `planned`, `unsupported_api`, `unsafe_or_disallowed`, `unsupported_local`, or other existing validator-compatible non-implemented availability values.
4. Do not mark raw API, direct write, or generic mutation commands as implemented.
5. Ensure examples contain no secret-looking strings and favor `--json`.

### Slice 3 — API surface inventory bridge

1. Refresh `api_surface.json` enough for #189 metadata parity without overclaiming implementation.
2. Preserve current implemented stream coverage.
3. Add official operation rows that are needed by `cli_surface.json` references and/or parent ledger bootstrapping.
4. If the full 342-operation classification is too large for #189, record remaining rows as #192 operation-ledger scope in `SUMMARY.md` with exact blocker category rather than silently omitting.

## TDD strategy

- Red: scripted or Go-based validation fails while Front has 10 `api_surface` entries and no `cli_surface.json`.
- Green: `cli_surface.json` validates and implemented commands resolve only to existing streams.
- Refactor: JSON formatting and docs wording only unless Go validation gaps appear.

## Verification checklist

Focused:

- [ ] red metadata completeness check captured.
- [ ] `jq empty internal/connectors/defs/front/api_surface.json internal/connectors/defs/front/cli_surface.json`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go test ./cmd/connectorgen -run CLISurface`
- [ ] `go test ./internal/connectors/engine -run CLISurface`

Broader before sub-PR handoff:

- [ ] `gofmt -w cmd internal` (only if Go files change)
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## CLI help/docs/website parity

This slice adds connector metadata that will feed CLI help/docs later. Runtime help/docs/website
rendering is #190. This PR must explicitly state the #190 dependency and not claim runtime command
availability beyond existing stream-backed metadata.

## Safety gates

- No secrets.
- No credentialed Front checks.
- No reverse ETL execution.
- No generic raw HTTP writes.
- No generic SQL/shell write tooling.
- No new dependencies.
- No destructive/admin action execution.
