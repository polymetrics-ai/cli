# Plan: Gong CLI Parity Parent Orchestration

Parent issue: #133
Parent branch: `feat/133-gong-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/232 (draft, base `main`)
Default branch: `main`

## GSD command path

- `scripts/gsd doctor` — pass.
- `scripts/gsd verify-pi` — pass.
- `scripts/gsd list --json` — pass; output truncated by harness but command exited 0.
- `scripts/gsd prompt plan-phase 133 --skip-research --tdd` — rendered plan prompt to `/tmp/gsd-plan-phase-133.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-cli-parity --dry-run` — adapter returned `unknown GSD command: programming-loop`.
- Fallback: use repo-local Pi `/pm-orchestrate` + `/pm-gsd-loop` prompt bodies and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`; record manual-GSD fallback until `scripts/gsd` exposes `programming-loop`.

## Required skills loaded

- GSD/runtime: `gsd-core`, `caveman`.
- Go/CLI/connector: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `golang-lint`.
- Repo references: `required-skills-routing.md`, `gsd-pi-adapter.md`, `cli-help-docs-website-parity.md`, connector migration handoff/conventions/design docs.

## Mission

Deliver Gong connector CLI parity against the official public Gong OpenAPI 3.0.1 spec at `https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=` and parent issue #133.

Official spec fetch evidence on 2026-07-09:

| Metric | Count |
|---|---:|
| Paths | 57 |
| Operations | 67 |
| GET | 28 |
| POST | 27 |
| PUT | 8 |
| PATCH | 1 |
| DELETE | 3 |

Current bundle baseline:

- `internal/connectors/defs/gong/` exists.
- Streams: `users`, `calls`, `scorecards`.
- `api_surface.json`: 10 entries, 3 `covered_by.stream`, 7 legacy `excluded` rows.
- The exact public spec shows `/v2/calls/extensive` and `/v2/calls/transcript` are `POST` read-query operations, not `GET` out-of-scope exclusions.

## Sub-issue queue and dependencies

| Issue | Lane | Write scope | Dependencies | Initial state |
|---:|---|---|---|---|
| #144 | Operation ledger | `internal/connectors/defs/gong/api_surface.json`, docs, connector-specific validation tests | none; unlocks other lanes | local critical path |
| #141 | CLI surface metadata | `internal/connectors/defs/gong/cli_surface.json`, metadata validation/tests | #144 for exact operation IDs/classification | queued |
| #142 | Help renderer/docs | help/docs/website/generated artifacts | #141 metadata | queued |
| #143 | Stream runner | Gong stream definitions/schemas/fixtures/runner tests | #144, may parallel with #145 after ledger | queued |
| #145 | Direct reads | direct-read operation metadata/executor tests | #144; may need #146 body policy for POST query shapes | queued |
| #146 | Advanced body/binary | bounded body/binary policy and fixtures | #144 | queued |
| #147 | Sensitive/admin policy | `writes.json`, sensitive policy/redaction/confirmation tests | #144 and #146 for payload policy | queued |

## Orchestration decision

Decision: `not_spawned_runtime_capability_missing`.

Reason: this Pi API session exposes only `read`, `bash`, `edit`, and `write`; no `subagent` tool is available, so mutating workers cannot be spawned into isolated worktrees from this session. Continue with #144 as `local_critical_path` and record future worker prompts/handoffs in the state ledger.

## Parent PR plan

1. Create parent planning artifacts and #144 red test.
2. Land green #144 slice locally.
3. Commit only scoped files; do not add untracked `PI_CONNECTOR_PROMPT.md` unless explicitly requested.
4. Push `feat/133-gong-cli-parity` after local targeted validation.
5. Keep draft parent PR #232 open until all required sub-issues are integrated and final verification passes.
6. Do not merge parent PR to `main`; final merge is human-gated.

## Full-surface safety constraints

- No secrets requested, printed, stored, summarized, or embedded in fixtures.
- No credentialed Gong checks unless explicitly requested.
- No generic HTTP write, raw GraphQL mutation, generic shell write, or generic SQL write surface.
- Reverse ETL remains plan → preview → approval → execute.
- Sensitive/admin/destructive endpoints are typed reverse-ETL candidates, not permanent exclusions.
- Binary/input payload endpoints require bounded local input/output policy and no path traversal or broad paths.

## Verification plan

Targeted during #144:

```bash
go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
go test ./cmd/connectorgen -count=1
go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1
```

Broader before parent handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```
