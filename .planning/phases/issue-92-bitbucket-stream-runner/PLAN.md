# Plan: Stream-backed Bitbucket connector command execution for safe read commands.

Sub-issue: #92
Parent issue: #79
Parent branch: `feat/79-bitbucket-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/128 (draft)
Execution mode: `local_critical_path` in this Pi API session because no subagent tool is exposed.

## GSD command path

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed; 69 commands.
- `scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd` — generated and followed.
- `scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run` — unavailable (`scripts/gsd: unknown GSD command: programming-loop`).
- Manual fallback remains active: `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`.

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-context`, `golang-concurrency`, `golang-graphql`, `golang-lint`, `golang-spf13-cobra`

## Objective

Stream-backed Bitbucket connector command execution for safe read commands.

## Implementation approach

- Keep Bitbucket credentials optional and never request or log secret values.
- Use the official Bitbucket Swagger at `https://api.bitbucket.org/swagger.json` for operation paths/counts.
- Implement only bounded, explicit app intents. Do not add raw generic HTTP write, shell, SQL, local-git, or arbitrary mutation escape hatches.
- Reverse ETL-capable commands must remain plan → preview → approval → execute and carry risk/approval metadata.
- Destructive/admin/sensitive operations are blocked by default unless represented as explicit, confirmed write actions with policy metadata.

## Verification checklist

- Targeted red/green tests for this lane.
- `jq . internal/connectors/defs/bitbucket/*.json internal/connectors/defs/bitbucket/schemas/*.json`
- `go test ./cmd/connectorgen -run Bitbucket -count=1`
- `go test ./internal/cli ./internal/connectors/commandrunner ./internal/connectors/engine ./cmd/connectorgen -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/pm`
- `make verify`

## CLI help/docs/website parity

Applicable. Verify `pm help bitbucket`, `pm bitbucket`, `pm bitbucket --help`, `docs/cli/**`, `docs/connectors/**`, and website generated data where behavior or surfaced connector metadata changes.

## Outcome

Completed: stream-backed Bitbucket commands execute through connector commandrunner for reviewed read intents; stream schemas and fixture replay coverage added. Full local gates passed; see `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json` for evidence.
