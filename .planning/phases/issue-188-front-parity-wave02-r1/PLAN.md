# PLAN — issue #188 Front documented connector parity

## Scope

Worker branch: `fm/cli-front-parity-wave02-r1`.

Allowed write scope from firstmate brief:

- `internal/connectors/defs/front/**`
- Front-owned fixtures/CLI metadata/docs under that bundle
- GSD/planning artifacts for this worker
- GitHub issue bodies for #188 and #189-#195 only, via idempotent captain-policy addendum

Explicit non-goals / hard gates:

- No live Front calls, credentials, provider writes, certification, no-mistakes, push, PR, merge, or shared daemon changes.
- No shared runtime/foundation edits, no other connector edits, no new dependencies.
- No generic HTTP write/read/query escape hatch. Every operation must be fixed-target and typed or blocked/planned with evidence.
- Reverse ETL remains plan → preview → explicit approval → execute; destructive/admin writes require typed destructive confirmation.

## Required reading and skills loaded

- `AGENTS.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/workflows/parent-issue-orchestration-loop.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `docs/plans/universal-programming-loop-prd.md`
- `docs/prompts/universal-programming-loop-prompts.md`

Required skills loaded: `gsd-core`, `context-mode`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-safety`, `golang-lint`.

## GSD command path

- `scripts/gsd doctor` passed.
- `scripts/gsd prompt programming-loop init --phase issue-188-front-parity --dry-run` was unavailable in this adapter (`unknown GSD command: programming-loop`).
- Pi programming-loop prompt source `.pi/prompts/pm-gsd-loop.md` and universal loop were read and followed manually.
- Deterministic GSD trace generated with `scripts/gsd prompt quick --full ...` at `.planning/traces/issue-188-front-parity-wave02-r1-gsd-quick.prompt.md`.

## Parent/subissue graph inventory

Parent issue #188 owns subissues #189 (ledger), #190 (ETL/read + changefeed), #191 (direct/provider-search + binary), #192 (reverse ETL writes), #193 (CLI/config/help), #194 (fixtures/docs/conformance/Connector Guard), and #195 (certification/release evidence). Firstmate assigned this worker the combined connector-local implementation across all child lanes.

## Official-source inventory plan

1. Fetch/index `https://dev.frontapp.com/llms.txt` and reference markdown pages.
2. Parse embedded OpenAPI snippets per reference page.
3. Reconcile parent issue counts:
   - 112 ETL read operations
   - 118 reverse ETL write operations
   - 5 direct/provider-search/query operations
   - 5 binary/download operations
   - 4 CDC/changefeed surfaces
   - 11 excluded/not-applicable application-channel/voice/plugin surfaces
4. Preserve the duplicate `PATCH /channels/{channel_id}` docs entry as a duplicate ledger row while keeping one executable write action.

## Implementation slices

1. **Red validation:** run current Front validate/conformance and capture current gap evidence.
2. **Ledger generation:** replace `api_surface.json` with operation-ledger v1 rows covering all documented operations exactly once by operation source.
3. **Read surfaces:** expand `streams.json`, `spec.json`, schemas, and docs for all ETL/read and GET CDC surfaces. Keep existing six read fixtures and add/update first-stream fixture evidence.
4. **Write surfaces:** add `writes.json` for fixed-target Front write actions. Every action gets a bounded record schema, path fields, risk text, redaction, destructive confirmation, and DELETE idempotency notes.
5. **Direct/binary metadata:** add `operations.json` and `cli_surface.json` for implemented typed direct reads, planned binary downloads, and excluded/not-applicable fixed operations without generic passthrough.
6. **Fixtures/conformance:** generate sanitized write fixtures for executable writes and minimal certification/docs evidence. Run connector-local gates.
7. **GitHub addendum:** append idempotent captain-policy addendum to #188 and #189-#195 with `gh-axi`, preserving existing counts/body text.
8. **Commit:** commit completed connector bundle and planning artifacts; stop before no-mistakes/push/PR.

## Orchestration decision

`local_critical_path`: firstmate assigned this worker one combined parent/subissue scope in an isolated disposable worktree. Mutating sub-workers were not spawned because the allowed write scope is a single connector bundle plus shared issue bodies; splitting would create write-scope collisions. Read-only scouting was performed inline with context-mode and official source parsing.
