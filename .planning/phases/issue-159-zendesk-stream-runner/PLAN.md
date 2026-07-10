# Plan: Zendesk Stream Runner

Parent issue: #156
Sub-issue: #159
Stack base: `feat/161-zendesk-direct-read` while prior Zendesk stacked lanes await integration.
Sub-issue branch: `feat/159-zendesk-stream-runner`

## GSD Command Path

- GSD health already verified in this session with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Planning prompt: `scripts/gsd prompt plan-phase issue-156-zendesk-complete-implementation --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run`; adapter returned `unknown GSD command: programming-loop`, so manual GSD/TDD fallback remains active.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Enable safe ETL stream-backed Zendesk commands for top-level collection GET operations that do not require arbitrary per-command path parameters.

## Scope

- Generate `streams.json` entries and schemas for top-level Zendesk collection responses inferred from official OAS response array properties.
- Use bounded cursor-style pagination where Zendesk responses expose `next_page` links, with cross-host following disabled.
- Convert the corresponding `api_surface.json` rows from direct-read command coverage to stream coverage.
- Convert matching CLI commands from direct-read commands to stream-backed ETL commands.

## Non-goals

- Do not implement path-parameter child collection streams that require user-specific IDs in connection config.
- Do not implement binary downloads (#162) or writes (#163).
- Do not run live Zendesk reads.

## TDD / Red-Green Plan

1. Add a red test requiring at least 70 safe top-level Zendesk stream-backed commands and validating stream/API surface coverage.
2. Generate stream definitions and minimal permissive schemas from the official OAS response array keys.
3. Keep direct-read commands for point reads and parameterized child reads not eligible for top-level ETL streams.
4. Validate engine load, connectorgen static checks, docs validation, and no credentialed execution.

## Safety Rules

- Streams are GET-only and connector-relative.
- Streams use existing batch-size/limit bounds from the ETL runner.
- Pagination follows provider `next_page` only with cross-host disabled.
- No secrets, no credentials, no external connector calls.

## CLI Help / Docs / Website Parity

Generated connector manuals are refreshed for the new stream-backed commands. #158 owns broader renderer and website parity.

## Verification

Targeted:

```bash
go test ./internal/connectors/engine -run 'ZendeskStream|ZendeskDirectRead|ZendeskOperationLedger' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface|Schema' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
./pm docs validate --connectors-dir docs/connectors
```

Before handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
