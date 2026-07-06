# GitHub Stream Runner

Issue: #36
Branch: `feat/36-github-stream-runner`
Parent issue: #44

## Goal

Execute GitHub CLI-surface commands that are already mapped to existing ETL streams through a
generic connector command runner.

## Scope

- Allow only command-surface rows with `intent=etl`, `availability=implemented`, and `stream`.
- Map declared CLI flags to stream query overrides when `maps_to` starts with `query.`.
- Preserve bounded reads with the existing `LimitEmitter` path.
- Return structured policy errors for reverse ETL, direct writes, raw API, local workflows, partial,
  planned, unknown commands, unknown flags, and invalid enum values.
- Add a small app resolver so provider commands can use saved credentials without duplicating secret
  loading in the CLI.

## Out Of Scope

- Reverse ETL dispatch.
- Direct read operation ledger execution.
- GraphQL execution.
- Raw API execution.
- Generic HTTP, SQL, or shell tooling.

## Red/Green Plan

1. Add a red engine test proving `ReadRequest.Query` overrides static stream query values.
2. Add red runner tests for implemented stream commands and blocked commands.
3. Add red CLI integration tests for `pm github issue list` and blocked `issue create`.
4. Implement the minimal stream-only runner.
5. Wire provider command fallback in `internal/cli`.
6. Refactor only after targeted tests are green.

## Verification

- `go test ./internal/connectors/engine -run QueryOverride`
- `go test ./internal/connectors/commandrunner`
- `go test ./internal/cli -run GitHubCommandSurface`
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli`
- `go vet ./...`
- `go build ./cmd/pm`
- `make verify`

## Human Gates

- Any credential scope changes.
- Any write execution path.
- Any raw API, GraphQL, generic HTTP, generic SQL, or shell execution path.
- Parent PR merge into `main`.
