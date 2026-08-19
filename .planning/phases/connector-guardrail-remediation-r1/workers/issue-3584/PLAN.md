# PLAN — issue #3584 HubSpot / Bitbucket forward remediation

## Isolation and branch

- Worker directory: `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3584-hubspot-bitbucket`
- Branch: `fix/3584-hubspot-bitbucket-forward-remediation`
- Base branch: `fix/3579-connector-path-ownership-guardrails`
- Parent issue / PR: #3579 / #3580
- Sub-issue: #3584
- Spawn decision: `spawned`

Isolation checked before edits:

```bash
pwd -P
git rev-parse --show-toplevel
git status --short --branch
```

Observed: cwd and git top-level both point at this worker directory; branch is `fix/3584-hubspot-bitbucket-forward-remediation`; worktree was clean.

## Required reading loaded

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`
- `.agents/agentic-delivery/workflows/claude-review-loop.md`
- `.agents/agentic-delivery/workflows/stacked-parent-subissue-workflow.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` (loaded for CLI-surface audit awareness; no CLI behavior/help change planned)
- `.planning/config.json`, `.planning/PROJECT.md`, `.planning/ROADMAP.md`, `.planning/STATE.md`
- `docs/plans/universal-programming-loop-prd.md`, `docs/prompts/universal-programming-loop-prompts.md`
- `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, `docs/architecture/connector-architecture-v2-design.md`
- Current phase `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`
- `.planning/phases/connector-guardrail-remediation-r1/subissues/hubspot-bitbucket-remediation.md`
- First-eight audit report sections for PR #3529 and #3531 from `/Users/karthiksivadas/karthik-agent-workspace/data/cli-first-eight-connector-guardrail-audit-r1/report.md`
- `gh-axi issue view 3579 --repo polymetrics-ai/cli --full`
- `gh-axi issue view 3584 --repo polymetrics-ai/cli --full`
- `gh-axi issue view 3581 --repo polymetrics-ai/cli --full`
- `gh-axi pr view 3580 --repo polymetrics-ai/cli --comments --reviews --full`

## GSD mode

Commands run:

```bash
scripts/gsd doctor
scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run
scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run
```

- `scripts/gsd doctor`: passed.
- `programming-loop` alias: unavailable (`scripts/gsd: unknown GSD command: programming-loop`, exit 1).
- Fallback: manual GSD universal runtime loop with available `execute-phase` prompt trace; TDD, verification, review, and handoff requirements remain active.

## Required skills loaded

- `gsd-core`
- `no-mistakes`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`
- `golang-lint`

Task-specific rule hooks:

- `gsd-core` implementation workflow steps 1-9: issue-first, GSD prompt, plan/TDD/verification before edits, red/green/refactor evidence, commit/PR evidence.
- `golang-testing` rules 1, 3, 5: named table/subtests, independent tests, observable behavior over internals.
- `golang-error-handling` rules 1, 2, 7: check returned errors, wrap with context, avoid log-and-return duplication.
- `golang-security` security thinking model 1-3 and common-mistakes secret rule: identify trust boundaries, attacker control, blast radius; no secrets in fixtures/logs.
- `golang-safety` rules 2, 4, 6, 7: safe assertions, initialized maps, defensive copies, loop/resource cleanup awareness.
- `golang-design-patterns` rules 5, 9, 20, 21: error-first flow, timeouts for external calls, avoid new deps, design for testability.
- `golang-structs-interfaces` interface rules: small consumer-owned interfaces; accept interfaces/return structs; avoid premature abstraction.
- `golang-documentation` writing principles: concise, intent over paraphrase, no invented context, preserve obligation language.
- `golang-lint` suppression rules 1-4 and workflow rules 1-3: no blanket suppressions, justify suppressions, lint/format after significant changes.

## Scope and non-scope

Allowed writes in this worker:

- `.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584/**`
- focused tests proving issue #3584 ledger/proof completeness
- `internal/cli/**`, `internal/app/**`, `internal/connectors/manifest.go`, `internal/connectors/bundleregistry/**` only if a focused regression requires a forward code correction

Non-scope / stop-or-defer:

- Parent issue body, parent PR body, parent branch direct edits, shared parent state outside this worker directory
- `cmd/connectorgen/**`, ownership validator files, CI workflows, PM/no-mistakes guidance, generated unrelated docs
- #3585 shared engine/runner/connectorgen remediation; Bitbucket engine-path code corrections are ledgered/deferred, not edited here
- Secrets, credentialed connector checks, reverse ETL execution, dependencies, pushes to `main`, parent PR merge

## Slice plan

### Slice A — worker audit ledger TDD

1. Add worker-owned Go ledger-completeness test under `.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584/`.
2. Red: run `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3584`; expect failure until the remediation ledger exists and all #3529/#3531 shared paths are dispositioned.
3. Green: add `remediation-ledger.json` with PR URLs, merge SHAs, path classifications, connector-lane verdicts, dispositions, and evidence for every audited shared path.

### Slice B — shared path disposition review

1. Inspect merge commits `41a00398a88db809b4e799a59fea381ace5cc06e` (#3529) and `bfe785464d04fd73dba0c4a70f36e23dd84da3d0` (#3531) for shared paths.
2. Preserve valid foundations only when evidence exists (existing focused tests or gate commands).
3. If a harmful CLI/app/manifest/bundleregistry issue is found, add a focused failing regression before the forward fix.
4. If a Bitbucket engine-path correction is needed, record dependency to #3585 instead of editing engine code.

### Slice C — verification / PR

1. Run worker ledger test.
2. Run issue verification:
   - `go test ./internal/cli ./internal/app ./internal/connectors/bundleregistry`
   - `go test ./internal/connectors/boundary ./cmd/connectorgen`
3. Run broader local gates when feasible:
   - `gofmt -w cmd internal`
   - `go vet ./...`
   - `go build ./cmd/pm`
   - `make verify` if time/infra allow
4. Commit and push branch.
5. Open sub-PR to `fix/3579-connector-path-ownership-guardrails` with `Refs #3584` and `Refs #3579`.
6. Run scoped no-mistakes if feasible; stop on daemon error or `ask-user` finding.

## Planned artifacts

- `remediation-ledger.json`: canonical per-path disposition file for #3529/#3531.
- `ledger_test.go`: worker-owned completeness/schema/verdict test.
- `TDD-LEDGER.md`: red/green evidence and command transcript summary.
- `VERIFICATION.md`: local gates, no-mistakes, PR/review status.
