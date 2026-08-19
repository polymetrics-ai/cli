## Objective

Update PM orchestrator, worker, issue/PR template, connector implementation, and no-mistakes guidance so connector lanes stop and split when shared foundation work is required.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3583-pm-no-mistakes-connector-lane`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`
- Dependency: none
- Integration status: #3588 was captain-authorized and provisionally integrated into the parent branch as `86b91fc40f46b8653538531fc40c183913676f05` after no-mistakes run `01KZ0SEAKBB9TG7N3SMG97XKJS` passed at `0c321595d7ae4852550a5012a895c3e11f7e8298`; parent PR review/final readiness remains pending.

## Scope

Allowed write scope:

- `.agents/agentic-delivery/**` contracts, workflows, references, and templates
- `.agents/connector-migration/**` instructions if connector implementation routing is defined there
- `.pi/prompts/**` and `.pi/agents/**` only for connector-lane routing text
- `.github/ISSUE_TEMPLATE/**`, `.github/PULL_REQUEST_TEMPLATE*`, or equivalent templates
- repo no-mistakes guidance/config files if present

Do not change core validator code, GitHub workflows/rulesets, or remediation code.

## Required reading / skills

Read `AGENTS.md`, parent-orchestrator contract, issue-agent contract, worker handoff template, GSD universal loop, Pi active orchestration loop, automated review routing, Claude review loop, GSD adapter reference, required-skills routing, CLI parity reference, #3579, #3580, and #3581. Load `gsd-core`, `no-mistakes`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-documentation`, and `golang-lint`.

## GSD/TDD plan

Use `scripts/gsd doctor` and `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`; record fallback if `programming-loop` alias remains unavailable. Update phase PLAN/TDD/VERIFICATION before production edits.

Red evidence first:

- grep/schema/docs validation fails until connector-lane guidance says shared foundation work must become a separate foundation issue/PR;
- worker handoff/PR template validation fails until target connector scope and guard evidence are recorded;
- no-mistakes guidance fails until it stops auto-absorbing generic shared runtime/tooling into connector PRs.

## Acceptance criteria

- Connector implementation instructions require exactly one target connector and forbid unrelated connector/shared runtime/tooling changes in connector lanes.
- PM parent orchestrator routes shared foundation needs into separate sub-issues/PRs before connector implementation proceeds.
- Worker prompts/handoffs include target connector scope, guard command evidence, changed-path compliance, and explicit foundation-PR path for legitimate shared work.
- no-mistakes guidance says connector PRs must not absorb generic shared runtime/tooling or unrelated connector changes; it should ask/stop for foundation split instead.
- Issue and PR templates expose connector implementation scope and guard evidence.

## Verification

```bash
rg -n "connector implementation|foundation|target connector|ownership guard|no-mistakes" .agents .pi .github docs
scripts/gsd doctor
```

Run focused tests if the repository has template/schema checks for these files.

## PR / review requirements

Sub-PR targets the parent branch and uses:

```markdown
Refs #3583
Refs #3579
```

Include docs-only TDD exemption only where no executable check exists; otherwise include red/green evidence, no-mistakes validation, CI, automated review route, and worker handoff. Completed #3588 validation evidence is recorded in `workers/issue-3583/` and parent `VERIFICATION.md`.

## Safety

No secrets, no new dependencies, no quality-gate weakening, no branch protection edits, no push to `main`.
