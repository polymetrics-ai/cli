## Objective

Forward-remediate and record dispositions for HubSpot #3529 and Bitbucket #3531 shared CLI/app/manifest/runtime path findings without rewriting `main` history.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3584-hubspot-bitbucket-forward-remediation`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`
- Dependency: #3581

## Scope

Allowed write scope:

- remediation ledger/proof entries for PR #3529 and #3531
- targeted tests for preserving or correcting foundations introduced by those PRs
- only audited HubSpot/Bitbucket shared path families when a forward code correction is required:
  - `internal/cli/**` HubSpot shared CLI changes
  - `internal/app/**`, `internal/connectors/engine/**`, `internal/connectors/manifest.go`, `internal/connectors/bundleregistry/**` Bitbucket shared changes
  - related docs only where needed to record disposition

Do not edit Stripe/Freshchat/Google Ads engine/runner/connectorgen remediation scope, generated unrelated-connector remediation scope, or core validator/wiring scope unless a test exposes a direct integration bug.

## Required reading / skills

Read `AGENTS.md`, issue-agent contract, GSD universal loop, GSD adapter reference, required-skills routing, first-eight audit report sections for #3529/#3531, #3579, #3580, and #3581. Load `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-lint`.

## GSD/TDD plan

Use `scripts/gsd doctor` and `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`; record fallback if `programming-loop` alias remains unavailable. Update phase PLAN/TDD/VERIFICATION before production edits.

Red evidence first:

- ledger completeness test or fixture fails until all #3529/#3531 audited paths have disposition;
- focused regression fails before preserving/correcting any shared CLI/app/manifest/runtime behavior;
- guard fixture shows these path classes would fail in a connector lane.

## Acceptance criteria

- Every shared CLI/app/manifest/runtime path listed for #3529 and #3531 has an explicit disposition: preserved foundation with evidence, corrected forward, removed forward, or deferred to named foundation issue.
- Valid foundations are preserved only with tests or narrow rationale.
- Harmful/unrelated churn is removed or corrected by forward commits only.
- No history rewrite, no force-push, no blanket revert.
- Ledger cites PR URLs, merge SHAs, path classifications, disposition, and verification evidence.

## Verification

```bash
go test ./internal/cli ./internal/app ./internal/connectors/engine ./internal/connectors/bundleregistry
go test ./internal/connectors/boundary ./cmd/connectorgen
```

Broaden to `go test ./...` if code changes affect shared behavior.

## PR / review requirements

Sub-PR targets the parent branch and uses:

```markdown
Refs #3584
Refs #3579
```

Include GSD/TDD/skill evidence, no-mistakes validation, CI, automated review route, and worker handoff.

## Safety

No secrets, no credentialed connector checks, no new dependencies, no reverse ETL execution, no push to `main`.
