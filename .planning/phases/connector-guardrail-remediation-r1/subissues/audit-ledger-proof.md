## Objective

Record all eight first-merge dispositions and produce end-to-end enforcement proof for the target-aware connector ownership guard.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3587-first-eight-audit-ledger-proof`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`
- Dependencies: #3581 and #3582 for final guard proof; remediation ledger rows from #3584, #3585, and #3586.

## Scope

Allowed write scope:

- historical audit/disposition ledger under `docs/migration/**` or `.planning/phases/connector-guardrail-remediation-r1/**`
- guard proof fixtures/tests under the ownership validator package
- parent verification transcript artifacts

Do not edit workflow/ruleset wiring, PM/no-mistakes guidance, or remediation code beyond proof fixtures.

## Required reading / skills

Read `AGENTS.md`, issue-agent contract, GSD universal loop, GSD adapter reference, required-skills routing, first-eight audit report, #3579, #3580, and all sub-issues. Load `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, and `golang-lint`.

## GSD/TDD plan

Use `scripts/gsd doctor` and `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`; record fallback if `programming-loop` alias remains unavailable. Update phase PLAN/TDD/VERIFICATION before production edits.

Red evidence first:

- ledger schema/completeness test fails until all eight PRs have rows;
- end-to-end fixture fails until positive and negative ownership cases pass;
- label-bypass fixture fails until detection ignores label omission.

## Acceptance criteria

- Ledger records all eight first-merge dispositions:
  - HubSpot #3529
  - Stripe #3530
  - Bitbucket #3531
  - Zendesk Support #3532
  - Google Ads #3535
  - Freshchat #3536
  - Xero #3537 compliant baseline
  - Asana #3538 compliant baseline
- Each row includes PR URL, merge SHA, connector, path classes, verdict, forward disposition, owner sub-issue, and evidence.
- Positive fixtures prove target defs/fixtures/docs and narrow shared indexes/goldens are allowed.
- Negative fixtures prove shared runtime/tooling, unrelated connector changes, unrelated generated churn, and label omission fail.
- Explicit foundation PR path is documented and tested.
- Parent verification transcript lists local gates, CI, automated review coverage, required remote guard read-back, and no-mistakes readiness.

## Verification

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --base origin/main --scope-file <positive-fixture>
go run ./cmd/connectorgen ownership . --base origin/main --scope-file <negative-fixture>
```

Full parent verification is run by the orchestrator after all sub-issues integrate.

## PR / review requirements

Sub-PR targets the parent branch and uses:

```markdown
Refs #3587
Refs #3579
```

Include GSD/TDD/skill evidence, no-mistakes validation, CI, automated review route, and worker handoff.

## Safety

No secrets, no credentialed connector checks, no new dependencies, no push to `main`.
