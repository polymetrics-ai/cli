# Plan: Issue #132 HubSpot CLI Feature Parity Parent Roadmap

Date: 2026-07-10
Runtime: Pi API harness in `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-132-hubspot-cli-parity`
Branch: `feat/132-hubspot-cli-parity`
Parent issue: https://github.com/polymetrics-ai/cli/issues/132

## GSD command path

- `scripts/gsd doctor` — passed, 69 commands registered.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed, JSON parsed with 69 commands.
- `scripts/gsd prompt programming-loop init --phase issue-132-hubspot-cli-parity --dry-run` — **blocked**: pinned registry reports `unknown GSD command: programming-loop`.
- Fallback command prompts used for the live loop:
  - `scripts/gsd prompt plan-phase issue-132-hubspot-cli-parity --skip-research`
  - `scripts/gsd prompt execute-phase issue-132-hubspot-cli-parity --dry-run`
- Fallback policy: adapter health is good, but the pinned official command registry lacks `programming-loop`; use the manual universal programming loop with explicit TDD evidence and record this limitation in `RUN-STATE.json` and PR evidence.

## Required context loaded

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/schemas/orchestration-state.schema.yaml`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-spf13-cobra`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-graphql`
- `golang-lint`

## Parent objective

Deliver HubSpot connector CLI parity against the official HubSpot public OpenAPI collection, preserving full-surface safety:

- Official source: `https://github.com/HubSpot/HubSpot-public-api-spec-collection`
- Baseline: 401 OpenAPI files, 4,396 raw operations, 3,060 unique method/path operations.
- Expected unique method counts: GET 1,038; POST 1,314; PUT 169; PATCH 232; DELETE 307.
- Every operation must be stream, direct read, write, binary, local workflow, duplicate/deprecated/disallowed/auth-internal/product-scope exclusion, or an issue-linked typed blocker with exact evidence.
- No generic raw HTTP write, arbitrary GraphQL mutation body, generic shell write, or generic SQL write surface.
- Reverse ETL remains plan → preview → approval → execute, with typed confirmation for destructive/admin/sensitive actions.

## Parent branch and PR tasks

1. Commit this parent orchestration scaffold as the deliberate parent seed.
2. Push `feat/132-hubspot-cli-parity` to origin.
3. Open a draft parent PR to `main` with `Refs #132`.
4. Keep parent PR draft until all required sub-issues are integrated and full verification passes.
5. Do not merge parent PR to `main`; final merge is human-gated.

## Sub-issue queue

| Issue | Lane | Dependencies | Expected write scope | Status |
|---:|---|---|---|---|
| #134 | CLI surface metadata | Parent PR exists | `.planning/phases/issue-134-*`, `internal/connectors/defs/hubspot/cli_surface.json`, metadata/schema support if needed | selected first |
| #137 | Operation ledger | #134 inventory conventions | `internal/connectors/defs/hubspot/api_surface.json`, generated inventory/tests | queued |
| #140 | Sensitive/admin policy | #137 write classification | write risk metadata, redaction/confirmation tests, policy docs | queued |
| #136 | Stream runner | #134, #137 stream classifications | HubSpot streams/schemas/fixtures, runner tests | queued |
| #138 | Direct reads | #134, #137 direct-read classifications | direct-read policies/executor tests, CLI metadata | queued |
| #139 | Advanced/body/binary engine | #137, #138/#140 blockers | fixed POST body/query schemas, binary policies/tests | queued |
| #135 | Help renderer/docs | #134 plus implemented surfaces | help/manual/website/docs artifacts | queued |

## Orchestration decision

Current Pi API tool surface exposes `read`, `bash`, `edit`, and `write`, but no `subagent` tool and no isolated worker worktrees have been created yet. Mutating workers cannot be safely spawned from this coordinator checkout.

Decision for cycle 1: `local_critical_path` for #134 after parent PR seed, with `not_spawned_runtime_capability_missing` recorded for parallel sub-issue workers. Sub-issues that touch `internal/connectors/defs/hubspot/**` also have a write-scope collision and should run sequentially until the bundle is split into disjoint lanes.

## TDD strategy

- Parent scaffold: planning artifacts only; no production behavior.
- #134 first red tests:
  1. CLI surface validation supports the `binary` command intent only through typed operation metadata, not a raw write or local filesystem escape hatch.
  2. HubSpot CLI surface metadata exists, validates, has no `raw_api` or `direct_write` commands, and maps representative provider-like command families to safe app intents.
  3. HubSpot official inventory metrics match the 3,060-operation baseline when the initial API surface ledger is introduced.
- Keep failing tests before production edits; update `TDD-LEDGER.md` with red/green evidence.

## Verification checklist

Targeted during slices:

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run 'CLISurface|HubSpot'
go test ./internal/connectors/engine -run 'CLISurface|HubSpot'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Full parent handoff gates:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

CLI/help/docs parity gates apply to CLI-visible changes:

```bash
pm help <topic>
pm <namespace>
pm <command> --help
rg -n "hubspot|HubSpot" docs/cli website docs/connectors
```

## Human gates

- Parent PR merge to `main`.
- Live HubSpot credentials or live write tests.
- Auth scope changes.
- New dependencies.
- Destructive/admin external actions.
- Reverse ETL execution beyond plan/preview/approval evidence.
- Binary transfer runtime enablement without explicit safe destination policy.
- Quality gate reductions.
