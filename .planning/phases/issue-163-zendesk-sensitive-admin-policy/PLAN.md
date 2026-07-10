# Plan: Zendesk Sensitive/Admin Policy

Parent issue: #156
Sub-issue: #163
Stack base: `feat/159-zendesk-stream-runner` while prior Zendesk stacked lanes await integration.
Sub-issue branch: `feat/163-zendesk-sensitive-admin-policy`

## GSD Command Path

- GSD health already verified in this session with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Planning prompt: `scripts/gsd prompt plan-phase issue-156-zendesk-complete-implementation --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run`; adapter returned `unknown GSD command: programming-loop`, so manual GSD/TDD fallback remains active.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Implement Zendesk mutating operations as typed reverse-ETL write actions behind existing plan → preview → approval → execute gates, with destructive confirmation where required.

## Scope

- Generate endpoint-specific `writes.json` actions for non-deprecated Zendesk POST/PUT/PATCH/DELETE operations.
- Generate reverse-ETL command metadata and API surface coverage for those actions.
- Require path variables in record schemas, infer request-body fields from official OAS schemas where possible, and mark destructive DELETE operations with `confirm: destructive`.
- Keep deprecated mutating operations blocked in operation-ledger rows.

## Non-goals

- Do not execute any Zendesk write.
- Do not request or store credentials.
- Do not add raw generic HTTP write or arbitrary body escape hatch.
- Do not add new dependencies.

## TDD / Red-Green Plan

1. Add a red test requiring 295 Zendesk write actions/commands and destructive confirmation coverage for 85 DELETE operations.
2. Generate write actions and reverse-ETL command metadata from the official OAS.
3. Validate static schemas, API surface coverage, generated docs, and no credentialed execution.

## Safety Rules

- All write execution remains under existing connector-command reverse-ETL plan/preview/approval/run flow.
- DELETE actions require `confirm: destructive`.
- Record schemas include path fields as required and mark sensitive-looking fields for preview redaction through existing CLI redaction.
- No direct-write/raw API/generic HTTP write command is introduced.

## CLI Help / Docs / Website Parity

Generated connector manuals are refreshed for the new reverse-ETL commands. #158 owns broader renderer and website parity.

## Verification

Targeted:

```bash
go test ./internal/connectors/engine -run 'ZendeskWrite|ZendeskStream|ZendeskDirectRead|ZendeskOperationLedger' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface|Write' -count=1
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
