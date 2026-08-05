## Objective

Wire the target-aware connector ownership validator into GitHub Actions, local fast-feedback hooks, ergonomic connector implementation presentation, and the authoritative required remote merge gate.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3582-connector-ownership-ci-gate`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`
- Dependency: #3581

## Scope

Allowed write scope:

- `.github/workflows/**`
- `.github/**` hook/template files only where needed for required-check or label presentation
- `Makefile` or scripts invoking the same validator
- local hook setup/config already present in the repo
- narrow runbook docs under `docs/migration/**`

Do not change core validator behavior outside this issue except integration-call bugs discovered by tests.

## Required reading / skills

Read `AGENTS.md`, issue-agent contract, GSD universal loop, automated review routing, Claude review loop, GSD Pi adapter reference, required-skills routing, CLI parity reference, #3579, #3580, and #3581. Load `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, and `golang-lint`.

## GSD/TDD plan

Use `scripts/gsd doctor` and `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`; record the `programming-loop` alias fallback if it remains unavailable. Update phase PLAN/TDD/VERIFICATION before production edits.

Red evidence first:

- workflow/local hook fixture proves the guard command is called;
- a connector implementation diff without label still runs/fails when violating ownership;
- local hook uses same validator as CI;
- docs state local hooks are bypassable and remote required CI is authoritative.

## Acceptance criteria

- `connector-implementation` label/tag/scope presentation exists for ergonomics and auditability only.
- GitHub Actions runs the target-aware ownership validator automatically on every relevant PR and `main` push.
- Failing ownership check makes CI non-green.
- Local pre-commit/pre-push fast-feedback hook invokes the same validator.
- Documentation is technically truthful: local hooks are bypassable; required remote CI plus ruleset/branch protection is authoritative.
- GitHub ruleset or branch protection requires the guard check for `main`; read-back evidence is recorded. If permission is denied, record exact blocker and stop claiming completion.

## Verification

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --help
make connector-boundary
gh-axi api /repos/polymetrics-ai/cli/rulesets
gh-axi api /repos/polymetrics-ai/cli/branches/main/protection
```

## PR / review requirements

Sub-PR targets the parent branch and uses:

```markdown
Refs #3582
Refs #3579
```

Include GSD/TDD/skill evidence, scoped no-mistakes validation, CI, automated review route, and worker handoff.

## Safety

No secrets, no `gh auth refresh`, no dependency additions, no quality-gate weakening, no push to `main`. Escalate permission denial for branch protection/rulesets.
