# Plan: Zendesk Help Renderer

Parent issue: #156
Sub-issue: #158
Stack base: `feat/163-zendesk-sensitive-admin-policy` while prior Zendesk lanes await integration.
Sub-issue branch: `feat/158-zendesk-help-renderer`

## GSD Command Path

- GSD health already verified in this session with `scripts/gsd doctor`, `scripts/gsd verify-pi`, and `scripts/gsd list --json`.
- Planning prompt: `scripts/gsd prompt plan-phase issue-156-zendesk-complete-implementation --skip-research`.
- Programming loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-156-zendesk-complete-implementation --dry-run`; adapter returned `unknown GSD command: programming-loop`, so manual GSD/TDD fallback remains active.

## Required Skills Loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-spf13-cobra`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.

## Objective

Expose safe runtime help for generated connector command surfaces so `pm zendesk`, `pm help zendesk`, and `pm zendesk <command> --help` work without credentials.

## Scope

- Add connector-surface help fallback for dynamic connector namespaces.
- Render command-specific help from `cli_surface.json` metadata.
- Keep invalid command paths as usage/policy errors.
- Update tests and generated docs as applicable.

## Non-goals

- Do not add website UI changes in this slice; generated connector manuals remain canonical for now.
- Do not execute live Zendesk reads or writes.
- Do not request credentials.

## TDD / Red-Green Plan

1. Add red CLI tests for `pm help zendesk`, bare `pm zendesk`, and `pm zendesk read list-tickets --help`.
2. Implement connector help fallback and command-detail renderer.
3. Verify docs/manual generation and non-credentialed inspect/help behavior.

## Safety Rules

- Help rendering must not resolve credentials or run connector checks.
- Help text must preserve reverse-ETL plan → preview → approval → execute wording.
- No generic raw write/API command is documented as executable.

## Verification

Targeted:

```bash
go test ./internal/cli -run 'Zendesk.*Help|Connector.*Help' -count=1
./pm help zendesk
./pm zendesk
./pm zendesk read list-tickets --help
./pm connectors inspect zendesk --json
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
