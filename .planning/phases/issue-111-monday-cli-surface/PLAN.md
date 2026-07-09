# Plan — issue #111 Monday CLI surface metadata

## Objective

Produce and validate `internal/connectors/defs/monday/cli_surface.json` that maps Monday provider-like commands into safe app intents without adding raw API/write escape hatches.

## GSD mode

- `scripts/gsd prompt plan-phase issue-111-monday-cli-surface --skip-research` generated the planning prompt.
- `scripts/gsd prompt programming-loop ...` is unavailable in this repo-local registry, so manual GSD programming-loop fallback is active.
- Red-first TDD remains required before production metadata edits.

## Required skills

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`, `golang-documentation`, `golang-spf13-cobra`, `golang-lint`.

## Slice

1. Add tests asserting Monday loads an embedded command surface, validates cleanly, and exposes implemented ETL commands only for the existing safe streams (`boards`, `items`, `users`, `teams`, `tags`).
2. Add `cli_surface.json` with provider-like groups and commands:
   - implemented ETL: `board list`, `item list`, `user list`, `team list`, `tag list`.
   - planned/blocked direct reads and mutations: metadata only; no executable raw GraphQL/API/write commands.
3. Ensure examples do not contain secret-looking values and reverse ETL entries (if any partial/planned) include risk/approval text.
4. Run targeted validation/tests.

## Safety constraints

- No credentials, token values, or live Monday API calls.
- No generic `api`, raw GraphQL, raw HTTP, shell, or SQL write command.
- No implemented reverse ETL command in this lane.
- All mutation/admin/sensitive surfaces stay blocked/planned and are refined in #114/#117.

## Verification

```bash
go test ./cmd/connectorgen -run 'TestMondayCLISurface' -count=1
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayCLISurface' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs/monday --json
```
