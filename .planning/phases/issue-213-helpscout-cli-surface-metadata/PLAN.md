# Plan: Help Scout CLI Surface Metadata

Parent issue: #212
Sub-issue: #213
Parent branch: `feat/212-helpscout-cli-parity`
Sub-issue branch: `feat/213-helpscout-cli-surface-metadata`
Connector slug: `help-scout`

## GSD / Skills Evidence

- GSD planning prompt: `scripts/gsd prompt plan-phase 213 --skip-research --tdd` captured in `PROMPTS.md`.
- Manual programming-loop fallback inherited from parent: `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`).
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`.
- CLI help/docs/website parity reference loaded. Runtime/help/doc changes are not in this metadata-only slice unless validation requires docs.md updates.

## Scope

1. Refresh Help Scout surface inventory from `https://developer.helpscout.com/mailbox-api/`.
2. Update `internal/connectors/defs/help-scout/api_surface.json` with the official endpoint inventory without blanket sensitive/admin exclusions.
3. Add `internal/connectors/defs/help-scout/cli_surface.json` metadata mapping safe app intents to existing streams and future lanes.
4. Update `metadata.json`/`docs.md` only as needed to avoid overclaiming current coverage.
5. Keep all unimplemented writes/direct reads/binary reads blocked-by-default or planned; do not add raw generic API write tools.

## Source Inventory Baseline

A credential-free crawl of the official docs navigation found:

- 146 endpoint pages under `/mailbox-api/endpoints/.../`.
- 146 endpoint page request snippets.
- 145 unique method/path pairs.
- Method split: GET 79, POST 21, PUT 20, PATCH 6, DELETE 19.
- Duplicate docs path: `GET /v2/conversations/123/threads/456/original-source` appears on both JSON and RFC 822 thread-source pages; it must be classified as a duplicate or represented without double-counting execution coverage.
- Existing bundle covers 4 stream endpoints: conversations, customers, mailboxes, users.

## Implementation Steps

1. Confirm the parent PR exists before production edits on the sub-issue branch.
2. Add a metadata red gate:
   - Add/refresh `cli_surface.json` and `api_surface.json`.
   - Run `go run ./cmd/connectorgen validate internal/connectors/defs` and capture any expected failures.
3. Normalize official sample paths from numeric examples to templated path variables in `api_surface.json` where appropriate.
4. Preserve current stream coverage for `conversations`, `customers`, `mailboxes`, and `users`.
5. For unimplemented direct-read/write/binary/sensitive/admin operations, use operation-ledger rows or planned `cli_surface.json` entries that do not claim executable coverage.
6. Validate JSON and connector definitions.
7. Update TDD ledger, verification, summary, and orchestration state.
8. Commit and push a green sub-issue slice; open sub-PR to the parent branch with `Refs #213` and `Refs #212`.

## TDD / Validation Plan

- Red: connector validation should fail if `cli_surface.json` references unknown streams/endpoints or if updated `api_surface.json` omits current streams.
- Green: validation passes with no `cli_surface_*`, `surface_*`, or secret-literal findings.
- Refactor: reduce duplicated metadata only after validation is green; keep JSON explicit and reviewable.

## Verification Checklist

Focused:

```bash
jq empty internal/connectors/defs/help-scout/*.json internal/connectors/defs/help-scout/schemas/*.json
go test ./cmd/connectorgen -run CLISurface
go test ./internal/connectors/engine -run CLISurface
go run ./cmd/connectorgen validate internal/connectors/defs
```

Issue-level:

```bash
go test ./cmd/connectorgen ./internal/connectors/engine
go test ./internal/connectors/conformance -run 'TestConformance/help-scout'
go build ./cmd/pm
```

Before handoff if shared docs or runtime surfaces change:

```bash
pm help connectors
pm connectors
pm connectors inspect help-scout --help
rg -n "help-scout|Help Scout" docs/cli docs/connectors website
```

## Non-Goals

- No credentialed Help Scout checks.
- No live API calls beyond public documentation fetches.
- No implementation of new stream runners, direct reads, binary downloads, or reverse-ETL writes in this slice.
- No new dependencies.
- No connector slug rename or duplicate `helpscout` bundle.

## Human Gates

Same as parent: secrets, credentialed checks, dependencies, destructive external actions, raw generic write tools, reverse ETL execution, quality gate reduction, and merge to `main`.
