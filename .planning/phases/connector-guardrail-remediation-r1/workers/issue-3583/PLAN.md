# PLAN — issue 3583 PM/no-mistakes connector lane

## Issue contract

- Sub-issue: #3583 — Connector guardrail: PM orchestrator and no-mistakes integration
- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580 (draft)
- Branch: `fix/3583-pm-no-mistakes-connector-lane`
- Base branch: `fix/3579-connector-path-ownership-guardrails`
- Spawn decision for this worker: `spawned`
- Write scope: `.agents/agentic-delivery/**`, `.agents/connector-migration/**` instructions if needed, `.pi/prompts/**`, `.pi/agents/**`, `.github/ISSUE_TEMPLATE/**`, `.github/pull_request_template.md`, repo no-mistakes guidance/config if present, and this worker artifact directory only.

## GSD mode

- `scripts/gsd doctor`: passed before planning in this worker checkout.
- `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`: generated `/tmp/gsd-execute-phase-3583.prompt.md` (87 lines) and loaded.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run`: unavailable in this adapter (`scripts/gsd: unknown GSD command: programming-loop`); using manual GSD universal runtime loop plus available execute-phase prompt. TDD, verification, review, and no-mistakes requirements are not weakened.

## Required skills loaded

- `gsd-core` (`.pi/skills/gsd-core/SKILL.md`)
- `no-mistakes` (`/Users/karthiksivadas/.agents/skills/no-mistakes/SKILL.md`)
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-security`
- `golang-safety`
- `golang-error-handling`
- `golang-documentation`
- `golang-lint`

Applied skill rules for this docs/guidance slice:

- gsd-core workflow rules 1–9: issue-first, GSD command, plan/TDD/verification before edits, red/green evidence, commit checkpoints.
- no-mistakes active validation-step boundary and ask-user escalation: do not drive nested pipeline control; stop on `ask-user` findings.
- golang-how-to Skill loading table: docs/CLI/review work loads testing, security, documentation, lint, and error-handling skills together.
- golang-documentation Writing Principles and Step 7: preserve obligations; docs/guidance must explain constraints and CLI/application behavior accurately.
- golang-testing Best Practices Summary #5: validation checks should constrain observable behavior, not implementation detail.
- golang-security Security Thinking Model #1–#3 and Common Mistakes: keep secrets/generic shell/raw write tools out of guidance.
- golang-safety Best Practices Summary #10 and Resource Safety: prefer explicit zero-risk guard wording that prevents accidental workflow misuse.
- golang-error-handling Best Practices #2, #7, #14: stop/propagate blockers with context; do not hide human-gate decisions.
- golang-lint Suppressing Lint Warnings rules #1–#4: do not weaken lint/quality gates or add blanket bypass guidance.

## Minimal green slices

### Slice A — PM/orchestrator and worker guidance

- Add connector implementation lane guardrails to parent orchestration contracts/workflows and Pi worker/orchestrator prompts.
- Require exactly one target connector in connector implementation worker prompts.
- Require orchestrators to stop and split shared runtime/tooling, connector schema/foundation, or unrelated connector changes into a separate foundation issue/PR before connector implementation proceeds.

### Slice B — connector migration instruction set

- Update connector rollout prompt, ownership rules, checklist, and validation gates so connector PRs never absorb generic shared runtime/tooling or unrelated connector changes.
- Add ownership guard command/evidence requirements and target connector scope to handoffs.

### Slice C — templates/no-mistakes evidence

- Update worker handoff and PR/issue templates to expose target connector scope, changed-path compliance, ownership guard evidence, and foundation PR path.
- Add no-mistakes guidance in repo templates/contracts so connector PR validation stops/asks for foundation split instead of auto-absorbing generic shared changes.

## Implementation status

- Slice A: implemented in PM parent orchestration contracts/workflows and Pi orchestrator/worker/planner prompts.
- Slice B: implemented in connector migration rollout prompt, ownership rules, checklist, validation gates, README, and passb-expander metadata.
- Slice C: implemented in worker handoff, issue prompt, issue template, PR template, issue-agent contract, and task-skill matrix.

## TDD / validation strategy

Use grep-based docs validation because this slice updates orchestration/template guidance and no executable schema exists for these markdown/YAML prompt contracts. Red evidence must fail before production guidance edits, then pass after edits.

Planned red checks:

```bash
rg -n "exactly one target connector" .agents .pi .github
rg -n "ownership guard evidence" .agents .pi .github
rg -n "foundation issue/PR|foundation PR" .agents .pi .github
rg -n "no-mistakes.*foundation split|foundation split.*no-mistakes" .agents .pi .github
```

## Verification plan

Required by issue:

```bash
rg -n "connector implementation|foundation|target connector|ownership guard|no-mistakes" .agents .pi .github docs
scripts/gsd doctor
```

Additional scoped verification:

```bash
git diff --check
```

Go gates (`gofmt -w cmd internal`, `go vet ./...`, `go build ./cmd/pm`, focused Go tests, `make verify`) are not required for docs/template-only changes that do not edit `cmd` or `internal`; if time permits, run `go test ./...` or `make verify` only after confirming no unrelated long-running blocker.

## Commit / push checkpoints

1. Planning artifact checkpoint after this PLAN/TDD/VERIFICATION seed is current.
2. Red evidence checkpoint if useful after failed grep validation is captured.
3. Green implementation checkpoint after docs/templates pass scoped verification.
4. Push branch `fix/3583-pm-no-mistakes-connector-lane` to origin.
5. Open sub-PR to `fix/3579-connector-path-ownership-guardrails` with `Refs #3583` and `Refs #3579`.

## Safety

- No secrets or credentialed connector checks.
- No new dependencies.
- No GitHub workflow/ruleset edits; #3582 owns workflow wiring.
- No parent issue/PR body or parent branch direct edits.
- No push to `main`; no parent PR merge.
