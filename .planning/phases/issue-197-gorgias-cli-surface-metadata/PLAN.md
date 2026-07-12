# Plan: Gorgias CLI Surface Metadata

Parent issue: #196  
Sub-issue: #197  
Parent PR: https://github.com/polymetrics-ai/cli/pull/229  
Parent branch: `feat/196-gorgias-cli-parity`  
Sub-issue branch: `feat/197-gorgias-cli-surface-metadata`  
Connector: `gorgias`

## GSD command path

- Parent preflight: `scripts/gsd doctor`, `scripts/gsd verify-pi`, `scripts/gsd list --json`.
- Planning prompt: `scripts/gsd prompt plan-phase 197 --skip-research`.
- Execution prompt: `scripts/gsd prompt execute-phase issue-197-gorgias-cli-surface-metadata --dry-run`.
- Programming-loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-197-gorgias-cli-surface-metadata --dry-run` — unavailable (`unknown GSD command: programming-loop`). Manual GSD universal runtime fallback is active.

## Required skills loaded

- `gsd-core`
- `caveman`
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
- `golang-lint`
- `golang-spf13-cobra` if CLI command-tree code changes become necessary

## Objective

Create and validate Gorgias CLI/API surface metadata that maps provider-style commands into safe Polymetrics intents without claiming runtime support that later lanes own.

## Allowed files for this slice

- `internal/connectors/defs/gorgias/cli_surface.json`
- `internal/connectors/defs/gorgias/api_surface.json` only for metadata/scope/source refresh needed by this lane
- `internal/connectors/defs/gorgias/metadata.json` only for capability/risk wording aligned with metadata inventory
- `internal/connectors/defs/gorgias/docs.md` only for metadata/scope notes needed by validation
- `.planning/phases/issue-197-gorgias-cli-surface-metadata/**`
- `.planning/phases/issue-196-gorgias-cli-parity/ORCHESTRATION-STATE.json` only for parent state update if this branch opens a sub-PR

## Out of scope

- Implementing new streams, schemas, fixtures, or pagination.
- Implementing write actions.
- Implementing direct-read or binary executors.
- Runtime help renderer changes, website docs, or generated CLI docs beyond validation notes.
- Full 114-operation classification if it becomes too large for this slice; #200 owns the complete operation ledger.

## Current baseline

- Official parent baseline: 114 operations: GET 59, PATCH 22, DELETE 16, POST 17.
- Candidate taxonomy: 47 ETL stream candidates, 55 reverse-ETL write candidates, 7 direct-read candidates, 5 binary/file candidates.
- Current `api_surface.json`: 11 endpoint entries.
- Current streams: `tickets`, `customers`, `messages`, `satisfaction_surveys`.
- Current writes: none.
- Current `cli_surface.json`: absent.

## Implementation slices

### Slice 1 — red validation and source capture

1. Capture red metadata-completeness check against the official 114-operation baseline.
2. Capture non-secret official source notes from https://developers.gorgias.com/llms.txt and any linked ReadMe/OpenAPI source reachable without credentials.
3. Do not commit raw large public specs unless needed; keep concise source capture in this phase.

### Slice 2 — safe CLI surface metadata

1. Add `cli_surface.json` with provider-inspired Gorgias command groups.
2. Mark implemented commands only for current stream-backed commands.
3. Mark write, direct-read, binary, and admin commands as planned/unsupported/unsafe as appropriate; do not expose generic raw API/direct write tooling.
4. Include `--json` examples and no secret-looking literals.
5. Reference `api_surface` only for implemented stream commands that are covered by existing `api_surface.json` rows.

### Slice 3 — API/metadata scope refresh

1. Update `metadata.json` and `docs.md` wording so they do not overclaim full CLI parity.
2. Preserve `capabilities.write=false` until #203/#write-action lanes add real write actions.
3. If full API classification is deferred to #200, record the exact deferral and red evidence in `SUMMARY.md` and parent state.

## TDD strategy

- Red: metadata completeness check fails because current Gorgias has 11 `api_surface` rows and no `cli_surface.json` versus the 114-operation baseline.
- Green: `cli_surface.json` parses and validates; implemented commands resolve only to existing streams/API rows; whole definition validation stays green.
- Refactor: JSON formatting and doc wording only unless Go validator gaps appear.

## Verification checklist

Focused:

- [x] Red metadata completeness check captured.
- [ ] `jq empty internal/connectors/defs/gorgias/api_surface.json internal/connectors/defs/gorgias/cli_surface.json`
- [ ] `go test ./cmd/connectorgen -run CLISurface`
- [ ] `go test ./internal/connectors/engine -run CLISurface`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `git diff --check`

Broader before sub-PR handoff:

- [ ] `gofmt -w cmd internal` (only if Go files change)
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`

## CLI help/docs/website parity

This slice adds connector metadata consumed by later help/docs lanes. Runtime help rendering, `docs/cli/**`, website docs, and generated artifacts are #198 scope. This PR must not claim Gorgias runtime command availability beyond the current stream-backed metadata.

## Safety gates

- No secrets requested, printed, stored, summarized, or invented.
- No credentialed Gorgias checks.
- No reverse ETL execution.
- No destructive/admin external actions.
- No new dependencies.
- No generic shell, generic HTTP write, generic SQL write, direct_write, raw_api, or raw GraphQL mutation tooling.
