# Plan: Zendesk Direct Read

Parent issue: #156
Sub-issue: #161
Stack base: `feat/160-zendesk-operation-ledger` until prior stacked PRs are integrated.
Sub-issue branch: `feat/161-zendesk-direct-read`

## GSD Command Path

- GSD health already verified in this session with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Planning prompt: `scripts/gsd prompt plan-phase issue-156-zendesk-complete-implementation --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run`; adapter returned `unknown GSD command: programming-loop`, so manual GSD/TDD fallback remains active.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Implement bounded, typed Zendesk direct-read commands for safe JSON GET operations without exposing a raw generic HTTP escape hatch.

## Scope

- Extend the generic direct-read engine/runner to support a bounded `json` output policy.
- Generate implemented `cli_surface.json` commands for Zendesk `direct_read` operation-ledger rows.
- Add path-variable flags and simple query-parameter flags from the official OAS where they are safe identifiers.
- Mark direct-read API surface rows with `covered_by.direct_read` command paths; leave binary reads, streams, and writes blocked for their lanes.

## Non-goals

- Do not implement binary downloads (#162).
- Do not implement streams (#159).
- Do not implement reverse-ETL writes (#163).
- Do not run credentialed Zendesk checks or execute live connector reads.

## TDD / Red-Green Plan

1. Add a red test proving the command runner accepts a generic `json` direct-read output policy and rejects unsupported policies.
2. Add a Zendesk bundle test proving implemented direct-read command count matches 282 OAS direct-read rows and API surface coverage references known commands.
3. Implement the generic `json` output policy and generated Zendesk direct-read command metadata.
4. Validate with focused engine/commandrunner/connectorgen tests and connector docs validation.

## Safety Rules

- Direct reads are GET-only, connector-relative, bounded to `MaxDirectReadBytes`, and JSON-decoded.
- No command accepts arbitrary URLs or methods.
- No secrets in command metadata or docs.
- Query/path flags are allow-listed per operation metadata; unknown flags fail.

## CLI Help / Docs / Website Parity

This issue changes runtime connector command metadata for `pm zendesk ...`. #158 owns full help renderer/docs parity, but this slice must keep connector manuals/generated docs consistent and verify non-credentialed inspection.

## Verification

Targeted:

```bash
go test ./internal/connectors/commandrunner -run 'DirectRead' -count=1
go test ./internal/connectors/engine -run 'Zendesk|DirectRead' -count=1
go test ./cmd/connectorgen -run 'CLISurface|Surface' -count=1
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
