# Plan: #481 Shepherd CLI Architecture v2 canary

## Objective

After #480 integrates, prove the production Shepherd against the exact authoritative #397 / draft
PR #438 topology without taking over or merging that parent program. Combine deterministic sandbox
trajectories with a supported read-only live `/pm-shepherd canary` reconciliation, then activate the
prepared legacy-shell deprecation only after the canary and final parent gates pass.

Parent issue: #471. Parent PR: #472. Dependency: #480. Canonical Shepherd branch:
`feat/481-shepherd-cli-architecture-canary`; the prior issue text's `test/481-...` branch is
superseded because the production workspace adapter owns the canonical `feat/<issue>-<slug>` form.

## GSD and skills

- `scripts/gsd doctor`: pass, 69 commands.
- Existing manual-GSD programming-loop fallback applies; no repeated unavailable-command retry.
- Required sources: `gsd-core`, parent contracts, Pi/runtime reference, issue #481, #397/#438
  authoritative GitHub state, and final #480 handoff.
- Task skills: `gsd-verify-work`, `gsd-code-review`, `e2e-testing-patterns`, parent orchestration,
  and JavaScript/TypeScript testing/security routing. No Go implementation skill applies.

## Scope

Allowed:

- `.pi/extensions/shepherd/cli-architecture-canary.ts`
- `.pi/extensions/shepherd/cli-architecture-canary.test.ts`
- bounded canary fixtures under `.pi/extensions/shepherd/fixtures/issue-481/**`
- `.pi/README.md` and the exact Shepherd-specific deprecation marker prepared by #480
- `.planning/phases/481-shepherd-cli-architecture-canary/**`
- parent-owned final #471 traces only after integration

The canary may inspect #397/#438 through typed/read-only authority. It must not merge #438, mutate
its parent branch, duplicate its workers, consume its human gates, or request secrets. Current
reconciliation records #438 at `21d195aff0c7bd60b3bf54f14b1ce165cec9e03f`, draft/open and
GitHub `DIRTY`/conflicting; that state is evidence to report, never a reason to rewrite it.

## TDD slice

1. RED: add deterministic tests for exact #397/#438 reconciliation, two independent eligible lanes,
   dependency/collision serialization, restart, one synthetic exact human gate, no secret
   persistence, and no #438 merge/mutation authority.
2. GREEN: implement the bounded canary harness and post-pass deprecation receipt.
3. REFACTOR while focused tests stay green.
4. Run the supported read-only live `/pm-shepherd canary --issue 397 --pr 438 --read-only
   --backend sdk-inproc --experimental` only after deterministic gates; record exact head and result.
5. Verify focused/full Shepherd, strict Pi 0.80.10 typecheck, family/provenance/RPC, diff/scope, and
   hosted CI.
6. Run one bounded Codex 5.6-sol xhigh review round and at most one concrete correction pass.
7. Integrate into the non-default #471 parent; leave #481 open until #472 lands on `main`.

## Acceptance

- Exact live #397/#438 state is reconciled before any action and no consumer mutation occurs.
- Deterministic canary proves two independent isolated lanes plus dependency/collision serialization.
- Exercised lanes carry GSD/TDD, verification, stacked-PR, review/correction, and integration
  evidence without fabricating GitHub success.
- One synthetic or designated human request survives restart and consumes one exact allowlisted
  answer.
- #438 remains draft/unmerged and no secret appears in prompts, state, logs, comments, or tests.
- The prepared rollback path becomes deprecated only after a bound passing canary receipt.
