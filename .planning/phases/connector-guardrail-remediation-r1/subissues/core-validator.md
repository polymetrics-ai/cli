## Objective

Add the target-scope contract and core changed-path ownership validator for connector implementation PRs.

## Parent

- Parent issue: #3579
- Parent PR: https://github.com/polymetrics-ai/cli/pull/3580
- Parent branch: `fix/3579-connector-path-ownership-guardrails`
- Sub-issue branch: `fix/3581-target-scope-core-validator`
- Sub-PR base: `fix/3579-connector-path-ownership-guardrails`

## Background

The existing `connectorgen boundary` guard scans shared production Go for connector-specific policy. It does not know a target connector and cannot reject unrelated connector definitions, unrelated generated docs, or generic shared runtime/tooling changes in connector implementation PRs.

## Scope

Allowed write scope for this sub-issue:

- `cmd/connectorgen/**` for the validator CLI surface
- `internal/connectors/boundary/**` for reusable ownership validation
- focused tests under those packages
- minimal validator usage docs under `docs/migration/**` if needed

Do not edit GitHub workflow/ruleset wiring, PM/no-mistakes guidance, or remediation ledgers in this slice.

## Required reading

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `docs/migration/connector-boundary-guard.md`
- Parent issue #3579 and parent PR #3580

## Required skills

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-documentation`

## GSD/TDD plan

Use repo-local GSD through Pi or shell:

```bash
scripts/gsd doctor
scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run
```

`programming-loop` script alias is currently unavailable; record the same fallback used by the parent PLAN if still unavailable. Create/update this phase's `PLAN.md`, `TDD-LEDGER.md`, and `VERIFICATION.md` before production edits.

Red tests first:

- exactly one connector slug must be declared in machine-readable scope;
- target connector auto-detection from changed paths works when no label/scope marker is present;
- shared runtime/tooling paths fail;
- unrelated connector definitions fail;
- unrelated generated connector docs/website paths fail;
- connector-owned defs/fixtures/docs and narrow shared indexes/goldens pass.

## Acceptance criteria

- Machine-readable connector implementation scope exists and contains exactly one connector slug.
- Validator can infer connector implementation scope from changed paths so label/tag omission cannot skip the check.
- Validator uses closed allowed path classes.
- Validator rejects shared runtime/tooling, unrelated connectors, unrelated generated docs, and unrelated generated website churn.
- Validator allows target defs, target fixtures/tests, target generated docs/website/manual outputs, and narrow shared indexes/goldens only under tested rules.
- Connector-lane exception/config edits cannot silently weaken the gate.
- CLI usage supports machine-readable output.

## Verification

```bash
gofmt -w cmd internal
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --help
go build ./cmd/pm
```

Run broader gates if the implementation touches shared packages beyond the listed scope.

## PR / review requirements

Open a sub-PR against `fix/3579-connector-path-ownership-guardrails` with:

```markdown
Refs #3581
Refs #3579
```

Record GSD/TDD/skill evidence, red/green verification, no-mistakes evidence, and automated review coverage. Use the worker handoff template.

## Safety

No secrets, no new dependencies, no credentialed connector checks, no reverse ETL execution, no generic raw write tools, no pushes to `main`.
