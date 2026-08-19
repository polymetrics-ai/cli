## Objective

Forward-remediate and record dispositions for Stripe #3530, Freshchat #3536, and Google Ads #3535 shared engine, commandrunner, command-surface, and connectorgen findings without rewriting `main` history.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3585-shared-engine-runner-remediation`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`
- Dependency: #3581 for any `cmd/connectorgen` ownership-validator-adjacent edits; ledger/test reconnaissance may begin after parent PR exists.

## Scope

Allowed write scope:

- remediation ledger/proof entries for PR #3530, #3535, and #3536
- focused tests for preserving or correcting audited shared behavior
- only audited shared path families when a forward code correction is required:
  - `internal/connectors/engine/**`
  - `internal/connectors/commandrunner/**`
  - `internal/connectors/command_surface.go`
  - `cmd/connectorgen/**` excluding ownership-validator files changed by #3581 until #3581 integrates
  - `internal/cli/cli.go` only for Google Ads command-surface wiring, if needed

Do not edit HubSpot/Bitbucket CLI/app remediation, unrelated generated connector docs, GitHub workflow/ruleset wiring, or PM/no-mistakes guidance.

## Required reading / skills

Read `AGENTS.md`, issue-agent contract, GSD universal loop, GSD adapter reference, required-skills routing, first-eight audit report sections for #3530/#3535/#3536, #3579, #3580, and #3581. Load `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `golang-lint`.

## GSD/TDD plan

Use `scripts/gsd doctor` and `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run`; record fallback if `programming-loop` alias remains unavailable. Update phase PLAN/TDD/VERIFICATION before production edits.

Red evidence first:

- ledger completeness test/fixture fails until #3530/#3535/#3536 audited shared paths have disposition;
- focused regression fails before preserving/correcting engine write DELETE/no-body, commandrunner, or connectorgen behavior;
- guard fixture shows shared runtime/tooling and unrelated connector definitions would fail in a connector lane.

## Acceptance criteria

- Every shared engine/runner/connectorgen/command-surface path listed for #3530, #3535, and #3536 has an explicit disposition.
- Required general foundations are preserved only with explicit ownership and tests.
- Unrelated or harmful behavior is removed or corrected by forward commits only.
- Google Ads hook/native edits are recognized as connector-owned native/hook surface, while unrelated `gong` definition and shared runtime changes are not.
- No history rewrite, no force-push, no blanket revert.
- Ledger cites PR URLs, merge SHAs, path classifications, disposition, and verification evidence.

## Verification

```bash
go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen
go test ./internal/connectors/boundary ./cmd/connectorgen
```

Broaden to `go test ./...` if code changes affect shared behavior.

## PR / review requirements

Sub-PR targets the parent branch and uses:

```markdown
Refs #3585
Refs #3579
```

Include GSD/TDD/skill evidence, no-mistakes validation, CI, automated review route, and worker handoff.

## Safety

No secrets, no credentialed connector checks, no new dependencies, no reverse ETL execution, no push to `main`.
