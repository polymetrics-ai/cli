# Plan: Chatwoot CLI Parity Parent Orchestration

Parent issue: #148
Parent branch: `feat/148-chatwoot-cli-parity`
Parent PR: pending creation
Default branch: `main`
Connector: `chatwoot`
Definition scope: `internal/connectors/defs/chatwoot/`

## GSD command path

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass; output large/truncated in terminal, command succeeded.
- `scripts/gsd prompt programming-loop init --phase issue-148-chatwoot-cli-parity --dry-run`: unavailable (`scripts/gsd: unknown GSD command: programming-loop`). Trace: `traces/gsd-programming-loop-unavailable.txt`.
- Fallback prompt traces generated:
  - `scripts/gsd prompt quick --validate "Parent orchestration for issue #148 Chatwoot CLI parity and local critical path planning for issue #149"`
  - `scripts/gsd prompt plan-phase 1 --skip-research`

Because the programming-loop command is absent from the current adapter registry, this phase uses the manual GSD universal loop from `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` plus `.pi/prompts/pm-gsd-loop.md`. TDD evidence remains mandatory before production edits.

## Required skills loaded

- `gsd-core`
- `caveman` for compact orchestration status/handoffs
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-lint`

## Parent objective

Bring Chatwoot to full connector CLI parity against the official Swagger source:
`https://raw.githubusercontent.com/chatwoot/chatwoot/develop/swagger/swagger.json`.

Baseline from issue #148:

- 144 official operations across 89 paths.
- Method split: POST 41, GET 62, PATCH 21, DELETE 18, PUT 2.
- Current bundle baseline: 71 mapped API entries, 7 streams, 6 write actions.

## Work queue and dependencies

| Issue | Lane | Initial write scope | Dependencies | Execution decision |
| --- | --- | --- | --- | --- |
| #149 | CLI surface metadata | `internal/connectors/defs/chatwoot/{metadata.json,api_surface.json,cli_surface.json,docs.md}`, `.planning/phases/issue-149-*` | parent plan/PR seed | `local_critical_path` |
| #152 | Operation ledger | `internal/connectors/defs/chatwoot/api_surface.json`, ledger artifacts | #149 official inventory | `not_spawned_dependency_blocked` |
| #151 | Stream runner | `internal/connectors/defs/chatwoot/{streams.json,schemas,fixtures}` | #149/#152 inventory | `not_spawned_dependency_blocked` |
| #153 | Direct reads | `internal/connectors/defs/chatwoot/cli_surface.json`, direct-read runner metadata | #149/#152, possible #154 output policy | `not_spawned_dependency_blocked` |
| #154 | Advanced query/binary engine | provider-specific query/body/binary metadata | #149/#152 | `not_spawned_dependency_blocked` |
| #155 | Sensitive/admin policy | write risk tiers, redaction, destructive confirms | #149/#152 | `not_spawned_dependency_blocked` |
| #150 | Help renderer/docs | runtime help/docs/website parity | #149 command metadata | `not_spawned_dependency_blocked` |

## Orchestration decision

Current Pi API tool surface in this session does not expose the `subagent` tool, so mutating worker dispatch is unavailable. The coordinator will take the local critical path for parent setup and issue #149. Record `not_spawned_runtime_capability_missing` for worker fanout until a Pi session with `subagent` is available.

## Slice boundaries

### Slice 0 — parent seed and plan checkpoint

- Create parent/issue planning artifacts.
- Commit the plan checkpoint as the deliberate parent seed diff so a draft parent PR can be opened.
- Push `feat/148-chatwoot-cli-parity`; open parent PR to `main` as draft with `Refs #148`.

### Slice 1 — issue #149 local critical path

- Capture red validation showing official Swagger has 144 operations while current `api_surface.json` has 71 entries.
- Refresh Chatwoot `api_surface.json` from the official source without enabling raw generic writes.
- Add Chatwoot `cli_surface.json` mapping provider-shaped commands to implemented streams/writes and planned/blocked safe intents.
- Update metadata/docs to avoid overclaiming unsupported direct-read/binary/sensitive/admin execution.
- Run targeted validation and tests.
- Commit and push a green slice; open sub-PR targeting the parent branch when branch isolation is practical.

## TDD approach

- Red validation for #149: a script compares official Swagger operation count and current surface count; expected fail before metadata refresh.
- Green validation: official count equals `api_surface.json` endpoint count and method split matches issue #148 baseline.
- Connector validation: `go run ./cmd/connectorgen validate internal/connectors/defs/chatwoot`.
- CLI-surface validation: `go test ./cmd/connectorgen -run CLISurface -count=1` after adding `cli_surface.json`.

## Safety gates

- No secrets, credential values, live connector checks, or credentialed API calls.
- No new dependencies.
- No raw generic HTTP write, raw SQL write, raw GraphQL mutation, or shell tool exposure.
- Reverse ETL remains plan → preview → approval → execute.
- Destructive/admin/sensitive operations stay blocked by default or require typed reverse-ETL action policy in later slices.
- Parent PR merge to `main` remains human-gated.
