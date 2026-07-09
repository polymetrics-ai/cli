# Plan: Front CLI Parity Parent Orchestration

Parent issue: #188
Parent branch: `feat/188-front-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/224 (draft)
Connector: `front`
Definition folder: `internal/connectors/defs/front/`

## GSD command path

- Preflight: `scripts/gsd doctor` — passed.
- Pi verification: `scripts/gsd verify-pi` — passed.
- Command inventory: `scripts/gsd list --json` — completed.
- Planning prompt used: `scripts/gsd prompt plan-phase 188 --skip-research --tdd`.
- Programming-loop prompt attempted: `scripts/gsd prompt programming-loop init --phase issue-188-front-cli-parity --dry-run` — unavailable (`unknown GSD command: programming-loop`).
- Pi-local prompt fallback source: `.pi/prompts/pm-gsd-loop.md` read and followed manually.

## Manual GSD fallback

The repo-local `scripts/gsd` adapter is healthy, but this checkout's command registry does not expose
`programming-loop` or `pm-gsd-loop` as shell prompt commands. This phase therefore uses the manual
GSD universal loop: plan, record TDD ledger, capture red validation before behavior changes, execute
small green slices, update verification evidence, and stop at human gates.

## Required skills loaded

- `gsd-core`
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
- `golang-spf13-cobra`
- `golang-spf13-viper`
- `golang-lint`

Required references loaded:

- `AGENTS.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/pi-active-orchestration-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/coderabbit-review-loop.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

## Baseline

- Official source: https://dev.frontapp.com/llms.txt (public docs index fetched successfully).
- Parent issue baseline: 342 official operations.
- Method split: GET 216, POST 69, PATCH 31, DELETE 26.
- Candidate taxonomy: 178 ETL stream candidates, 28 binary/file candidates, 126 reverse-ETL write candidates, 10 direct-read candidates.
- Current bundle baseline: `api_surface.json` maps 10 entries, `streams.json` declares 6 streams, no `writes.json` exists, and `metadata.json` advertises `write: false`.

## Full-surface safety target

Every official Front operation must resolve to exactly one of:

1. `covered_by.stream` for syncable structured collection reads.
2. `covered_by.direct_read` / `covered_by.direct_reads` for bounded safe read commands.
3. `covered_by.write` for named reverse-ETL actions.
4. `covered_by.binary_read` or bounded direct-read/binary metadata policy.
5. An explicit block only when duplicate, deprecated, disallowed, auth-internal, or outside product scope.

Sensitive/admin/destructive endpoints are not blanket exclusions. They must become typed reverse-ETL
actions with schemas, path fields, risk/approval text, redaction, and `confirm: destructive` where
appropriate, unless a precise product-scope or safety reason blocks them.

Forbidden: generic shell, generic HTTP write, generic SQL write, raw generic GraphQL mutation, unsafe
binary downloads, secret exposure, credentialed connector checks without request, and reverse ETL
execution outside plan → preview → approval → execute.

## Work queue and dependencies

| Issue | Lane | Dependencies | Primary write scope |
|---:|---|---|---|
| #189 | CLI surface metadata | parent PR exists | `internal/connectors/defs/front/`, `.planning/phases/issue-189-front-cli-surface-metadata/` |
| #192 | Operation ledger | #189 official inventory | `internal/connectors/defs/front/api_surface.json`, `.planning/phases/issue-192-front-operation-ledger/` |
| #194 | Advanced query/binary engine | #192 classifications | connector engine/direct-read/binary metadata files, `internal/connectors/defs/front/` |
| #191 | Stream runner | #189 + #192, and #194 when query/body support is needed | `internal/connectors/defs/front/streams.json`, schemas, fixtures |
| #193 | Direct reads | #192 + #194 where needed | direct-read metadata/runner, `internal/connectors/defs/front/` |
| #195 | Sensitive/admin policy | #192 + write classification | write policy schemas/metadata, `internal/connectors/defs/front/` |
| #190 | Help renderer/docs | #189 metadata first; update after each implemented surface | CLI help/docs/website/generated help artifacts |

## Slice boundaries

### Slice 0 — parent seed and PR

- Create parent orchestration artifacts.
- Commit and push the parent seed branch.
- Open a draft parent PR from `feat/188-front-cli-parity` to `main` using `Refs #188`.
- No production code or connector definitions change in this slice.

### Slice 1 — #189 CLI surface metadata

- Create a stacked sub-issue branch from the parent branch after the parent PR exists.
- Capture red validation that current Front surface metadata is incomplete against the official 342-operation baseline.
- Refresh `api_surface.json` and CLI surface metadata from official Front sources.
- Map provider operations into safe app intents without generic raw write escape hatches.
- Commit and push the plan checkpoint before production edits, then red-test/green slices.

### Later slices

- #192 locks the complete operation ledger and resolves duplicate/deprecated/disallowed/auth-internal/product-scope classifications.
- #191/#193/#194 implement stream/direct-read/binary support behind bounded policies.
- #195 adds typed write/destructive/admin risk policy and approval gates.
- #190 keeps CLI help/manual/website/generated artifacts aligned as surfaces become real.

## TDD strategy

- Planning-only slice: red test not applicable; validate JSON/YAML/Markdown and state honesty.
- #189 first red artifact: a focused validation/test that proves the current Front surface is incomplete (10 entries vs 342 official operations) and/or missing CLI surface metadata.
- Behavior-changing slices: add or update failing Go tests before connector engine/CLI changes.
- JSON-only connector slices: use `go run ./cmd/connectorgen validate internal/connectors/defs/front` plus targeted tests for new schema/metadata rules.

## Verification checklist

Initial parent seed:

- [ ] `jq empty .planning/phases/issue-188-front-cli-parity/*.json`
- [ ] `git diff --check`
- [x] confirm parent PR exists after push (`https://github.com/polymetrics-ai/cli/pull/224`)

Issue implementation gates before handoff:

- [ ] `gofmt -w cmd internal`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] CLI help/docs/website parity checks for changed surfaces: `pm help <topic>`, `pm <namespace>`, `pm <command> --help`, and docs/website grep or generator checks as applicable.

## Automated review route

- Parent PR targeting `main`: keep draft while sub-issues land; when non-draft, prefer CodeRabbit automatic review and do not post redundant manual commands.
- Stacked sub-PRs targeting `feat/188-front-cli-parity`: record CodeRabbit behavior. If skipped due to non-default base, parent PR fallback coverage is required for integrated ranges.
- Copilot backup only when CodeRabbit is skipped/rate-limited/disabled/unavailable and coverage blocks progress.

## Spawn decision

No mutating worker spawned in this API session because the `subagent` tool is not available in the
active tool set. Recorded as `not_spawned_runtime_capability_missing`. The local critical-path
action is to seed the parent PR; issue #189 will then run locally or in an isolated Pi worker if an
interactive Pi session with `subagent` is available.

## Human gates

- Parent PR merge to `main`.
- New dependencies.
- Auth scope changes or `gh auth refresh`.
- Secrets/credential access or credentialed connector checks.
- Destructive external actions.
- Reverse ETL execution.
- Production deploys.
- Quality-gate reductions.
- Generic shell, generic HTTP write, generic SQL write, or unrestricted raw API tooling.
