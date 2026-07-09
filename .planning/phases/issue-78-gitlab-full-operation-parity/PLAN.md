# Plan: GitLab Full Operation Parity Follow-up (#78)

Parent issue: #78
Parent PR: #127
Branch: `feat/78-gitlab-cli-parity`

## GSD Evidence

- Lane prompt: `scripts/gsd prompt execute-phase issue-78-gitlab-full-operation-parity --tdd`
- Programming-loop prompt: `scripts/gsd prompt programming-loop init --phase issue-78-gitlab-full-ops-parity --dry-run` returned `unknown GSD command: programming-loop`; manual universal GSD loop continues.

## Required Skills Loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- CLI help/docs/website parity reference

## User Request

Check how GitHub implemented broad operation parity and apply the same approach to GitLab: convert the full inventoried GitLab operation ledger into implemented, typed command-surface coverage where safe under existing gates.

## GitHub Model Observed

- `scripts/gen-github-parity.py` converts non-deprecated, non-duplicate operation-ledger rows into:
  - `operations.json` entries;
  - `writes.json` typed write actions for mutation endpoints;
  - `cli_surface.json` implemented commands;
  - `api_surface.json` `covered_by` rows.
- Operation-backed direct/binary commands remain feature-gated at runtime by `commandrunner` (`operation ... executor is not implemented`) unless a stream/direct-read/write action path exists.
- Write-backed commands use reverse ETL write actions and still require plan → preview → approval → execute.

## Safety Boundaries

- No raw generic GitLab API command.
- No credentialed GitLab checks.
- No secrets or new dependencies.
- Mutating GitLab endpoints may be represented only as typed write actions/commands behind reverse ETL plan → preview → approval → execute.
- Binary and non-GET direct reads may be represented as operation-backed commands but remain feature-gated unless a bounded executor exists.
- Do not enable arbitrary GraphQL mutation or generic body-template escape hatch.

## TDD / Execution Plan

1. Add a red test that GitLab full parity mirrors the GitHub coverage model: no non-deprecated/non-disallowed official OpenAPI operation remains blocked only as `api_surface.operation`; every such row is `covered_by`, and counts show broad command/write/operation coverage. ✅
2. Implement a GitLab parity generator modeled after `scripts/gen-github-parity.py`. ✅
3. Regenerate GitLab `api_surface.json`, `writes.json`, and `cli_surface.json` while retaining the existing full `operations.json` ledger. ✅
4. Update GitLab docs/website generated data to explain broad typed operation coverage and runtime gates. ✅
5. Validate with focused tests, connectorgen validation, docs generation/website generation, and final Go gates. In progress.

## Implemented Scope

- GitLab now mirrors the GitHub broad-parity scaffold: all non-deprecated official REST operations are covered by a stream, bounded direct read, operation-backed read/binary/HEAD command, or typed reverse-ETL write action.
- Covered endpoint rows: 1,142 of 1,145 rows (1,144 official + `/users` compatibility; 3 deprecated remain blocked).
- Implemented commands: 1,142 (`etl` 4, `direct_read` 501, `reverse_etl` 637).
- Runtime-safe executable slice remains bounded: 4 ETL streams and 4 APISurface direct reads run today; operation-backed reads stay feature-gated; writes require reverse ETL plan → preview → approval → execute.

## Verification Plan

```bash
go test ./cmd/connectorgen -run 'GitLab.*Operation|GitLab.*Full|GitHub' -count=1
go test ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli -run 'GitLab|Operation|DirectRead' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
cd website && pnpm run gen:website-data
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
