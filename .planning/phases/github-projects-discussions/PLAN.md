# PLAN: GitHub Projects And Discussions GraphQL Reads

## Tasks

1. Add red tests for engine GraphQL query-variable interpolation and optional cursor omission.
2. Add red tests for GitHub bundle loading and CLI surface mappings for project/discussion reads.
3. Extend the engine narrowly:
   - add `query.*` template namespace
   - pass `ReadRequest.Query` into GraphQL variable resolution
   - support `omit_when_empty` on GraphQL variable template objects
4. Add GitHub GraphQL read streams, schemas, fixtures, and operation ledger rows:
   - `projects`
   - `project_items`
   - `discussions`
   - `discussion`
5. Update GitHub `cli_surface.json` so read commands are implemented ETL streams.
6. Update GitHub docs and generated website bundle/catalog data.
7. Run validation:
   - focused red/green tests
   - `go run ./cmd/connectorgen validate internal/connectors/defs/github --json`
   - `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen`
   - `go test ./internal/connectors/conformance -run 'TestConformance/github'`
   - `go vet ./...`
   - `go build ./cmd/pm`

## Spawn Decision

- `spawned`: read-only workflow adapter audit and GitHub mapping audit sidecars.
- `not_spawned_write_scope_collision`: no mutating subagent for shared engine/GitHub bundle files in
  coordinator checkout. Implementation stays local for this slice.

## Human Gates

- No new dependency.
- No auth-scope refresh.
- No mutation execution.
- Parent PR merge to `main` remains human-gated.
