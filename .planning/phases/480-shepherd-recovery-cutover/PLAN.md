# Plan: #480 Shepherd recovery, audit, and reversible cutover

## Objective

Close the recovery/audit/operator-readiness gap after integrated #479 without reopening the frozen
17-row implementation campaign. Add fault-injected restart coverage, bounded redacted audit evidence,
and reversible legacy-shell cutover preparation. Do not activate deprecation before #481 passes.

Parent issue: #471. Parent PR: #472. Parent branch:
`feat/471-pi-agent-session-shepherd`. Canonical Shepherd branch:
`feat/480-shepherd-recovery-cutover` (the production workspace adapter supports `feat/<issue>-<slug>`).

## GSD and skills

- `scripts/gsd doctor`: pass, 69 commands.
- `programming-loop`: existing manual-GSD fallback; absence was checked once and is not retried.
- Planning path: `scripts/gsd prompt plan-phase 471 --skip-research` plus the universal runtime loop.
- Required sources: `gsd-core`, required-skills routing, Pi/runtime reference, parent-orchestrator
  contracts, issue #480, and #479/#490 completion artifacts.
- Task skills: `architecture-patterns`, `javascript-testing-patterns`, and repository security/docs
  routing. No Go implementation skill applies.

## Scope

Allowed:

- `.pi/extensions/shepherd/recovery.ts`
- `.pi/extensions/shepherd/recovery.test.ts`
- `.pi/extensions/shepherd/audit-log.ts`
- `.pi/extensions/shepherd/audit-log.test.ts`
- bounded recovery/audit fixtures under `.pi/extensions/shepherd/fixtures/issue-480/**`
- `.pi/README.md`
- Shepherd-specific `.agents/agentic-delivery/**` cutover documentation
- `.planning/phases/480-shepherd-recovery-cutover/**`

Existing production recovery/effect/state ports are consumed rather than duplicated. No new
package, credential, deployment, default-branch mutation, generic write tool, or quality-gate
reduction is allowed. Legacy scripts and historical worktrees/branches are retained.

## TDD slice

1. RED: add executable fault-injection tests for restart ordering, ambiguous external effects,
   bounded/redacted audit records, cancellation, stale heads, and reversible cutover state.
2. GREEN: add the smallest adapter/audit implementation over existing production ports.
3. REFACTOR: remove duplication while focused tests stay green.
4. Verify focused tests, the complete sequential Shepherd suite, strict pinned-Pi typecheck,
   Pi-family/provenance/RPC gates, diff/scope, and hosted CI.
5. Run one bounded Codex 5.6-sol xhigh review round. Apply at most one concrete blocker correction
   pass, then verify only dispositions.
6. Integrate only into the non-default parent branch. Keep issue #480 open until #472 lands on
   `main`.

## Acceptance

- Recovery reconciles persisted intent with state, leases, sessions, worktrees, refs, GitHub effects,
  checks/reviews, and human decisions before scheduling.
- Every ambiguous effect either reconciles idempotently or fails closed; duplicate publication and
  decision consumption are prevented.
- Audit records are closed-schema, bounded, redacted, causally linked, and contain no prompt,
  reasoning, credential, or unrestricted output.
- Fault tests cover process kill, power-loss windows, network failure, stale/force-moved refs,
  conflicts, review movement, rate limiting, cancellation, and abort/join.
- Cutover is prepared and reversible but not activated. #481 alone may activate post-canary
  deprecation.
