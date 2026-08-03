# PLAN — issue 3585 shared engine/runner remediation

## Task contract

- Sub-issue: #3585 — Stripe, Freshchat, and Google Ads shared engine/runner/connectorgen remediation.
- Parent: #3579, parent draft PR #3580.
- Branch: `fix/3585-shared-engine-runner-remediation`.
- Base/PR target: `fix/3579-connector-path-ownership-guardrails`.
- Worker mode: fallback mutating GSD/TDD implementation worker in Pi; do not spawn subagents.

## Required reading completed

- `AGENTS.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.agents/agentic-delivery/contracts/worker-handoff-template.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- GitHub issue #3585, parent issue #3579, parent PR #3580, dependency issue #3581 via `gh-axi`
- Captain authority record and first-eight audit report sections for PR #3530, #3535, and #3536

## Skills loaded / recorded

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

## GSD trace

- `scripts/gsd doctor` passed in this worktree.
- `scripts/gsd prompt execute-phase connector-guardrail-remediation-r1 --dry-run` saved to `traces/gsd-execute-phase-dry-run.txt`.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1 --dry-run` failed with `unknown GSD command: programming-loop`; manual GSD universal loop fallback is active and recorded in `traces/gsd-programming-loop-init-dry-run.txt`.

## Scope boundaries

Allowed writes for this worker:

- `.planning/phases/connector-guardrail-remediation-r1/workers/issue-3585/**`
- Focused tests/artifacts for PR #3530, #3535, and #3536 dispositions
- Production edits only if a failing test proves a forward correction is required, and only under:
  - `internal/connectors/engine/**`
  - `internal/connectors/commandrunner/**`
  - `internal/connectors/command_surface.go`
  - `cmd/connectorgen/**` (avoid ownership-validator-adjacent edits until #3581 integrates)
  - `internal/cli/cli.go` only if Google Ads command-surface wiring is demonstrably broken

Out of scope / do not edit:

- HubSpot/Bitbucket remediation, Zendesk/generated remediation, unrelated connector docs/website, workflows/rulesets, PM/no-mistakes guidance, parent issue/PR bodies, parent state ledger, active connector campaign branches, and `main`.

## Implementation slices

### Slice 1 — worker-owned red ledger test

Create a Go test in this worker directory that fails until a durable disposition ledger exists and includes required rows for every audited shared/runtime path from PR #3530, #3535, and #3536.

Expected RED: `go test ./.planning/phases/connector-guardrail-remediation-r1/workers/issue-3585` fails because `DISPOSITION-LEDGER.md` is missing required rows.

### Slice 2 — inspect existing shared foundations

Inspect current tests/code for the audited behaviors:

- Stripe #3530: `internal/connectors/engine/write.go` and `write_test.go` DELETE/no-body behavior.
- Freshchat #3536 and Google Ads #3535: `internal/connectors/commandrunner/runner.go` / `runner_test.go` command runner validation/mapping behavior.
- Google Ads #3535: `cmd/connectorgen/main.go`, `cmd/connectorgen/validate.go`, `cmd/connectorgen/main_test.go`, `internal/connectors/command_surface.go`, `internal/cli/cli.go`.

If coverage is already general, pass/fail evidence supports preserving the foundation with ledger ownership. If a failing focused test proves a bug, implement the smallest forward correction under the allowed shared path family.

### Slice 3 — disposition ledger/proof

Write `DISPOSITION-LEDGER.md` with one explicit disposition per audited path, citing:

- PR URL and connector
- Merge SHA
- Path classification
- Ownership decision
- Whether the behavior is preserved, corrected, or deferred to another sub-issue
- Verification evidence

Include Google Ads connector-owned hook/native paths as allowed connector-owned surface, and record the unrelated Gong definition as out of this worker's remediation scope / handed to #3586.

### Slice 4 — verification, commit, push, PR

Run targeted verification, then broader feasible gates. Commit only this worker's files and any justified focused code/test changes. Push with Alfred SSH command only. Open a sub-PR to the parent branch using `gh-axi`, with `Refs #3585` and `Refs #3579`.

## Production-change decision rule

Default to ledger/proof-only. Do not edit production Go unless a fresh failing test demonstrates a current bug or missing guard-adjacent behavior. Preserve valid shared foundations only when existing or newly added tests prove generic behavior and the ledger names the foundation ownership.

## Safety

No secrets, no credentialed connector checks, no new dependencies, no reverse ETL execution, no generic raw write tools, no force-push, no `main` push/merge.
