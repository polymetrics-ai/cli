# Objective

Integrate the policy, in-process worker runtime, workspace/Git, GitHub orchestration, decisions, and
evidence ports into a complete dependency-aware autonomous controller and `/pm-shepherd` UX.

Parent: #471
Parent PR: #472
Dependencies: #474, #475, #476, #477, and #478 (Wave 4)
Branch: `feat/479-shepherd-autonomous-controller`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/{controller,extension,index,domain}.ts`
- `.pi/extensions/shepherd/autonomous-controller.ts`
- `.pi/extensions/shepherd/{arguments,state-store}.ts`
- narrowly typed autonomous state, intake, verification, scoped-workspace, concrete GitHub
  transport, Codex-review, and effect-journal modules under `.pi/extensions/shepherd/`
- `.pi/extensions/shepherd/{workspace-adapter,git-adapter}.ts` only for typed coordinator-parent
  refresh plus stale-child rebase/reclaim capabilities; generic shell execution remains forbidden
- matching tests
- `.pi/README.md` command/operator section
- CLI help/manual parity artifacts applicable to `/pm-shepherd`
- this issue's GSD/TDD artifacts

This is the deliberate shared integration slice. The expanded scope above is required because the
preflight at parent head `2a89142e` proved that the read-only v1 parser/state DTO and immutable
initial-base workspace API cannot represent the autonomous lifecycle. See
`../traces/479-preflight-interface-audit.md`.

## Acceptance criteria

- [ ] `start` drives intake through scheduling without requiring `--read-only`, a target PR, or an
      external shell driver; `canary` retains the explicit read-only contract.
- [ ] Every ready non-colliding child is dispatched up to configured concurrency; dependencies and
      collisions explain deterministic waiting.
- [ ] Workers plan, implement red-green-refactor, verify, open/update PRs, review, correct, and
      integrate through authoritative adapters with bounded retry budgets.
- [ ] Read-only research/review sessions remain internal roles. Every top-level child issue is a
      scoped mutating issue-to-PR lane, matching #478's integration-roster contract.
- [ ] `status`, `stop`, and `resume` expose durable stage/lane/dependency/review/gate truth and join
      all child work correctly across stop/shutdown races.
- [ ] A validated v2 autonomous state persists plan identity, child stages, exact Git/GitHub facts,
      retry budgets, prepared/applied effects, review dispositions, and human waits without storing
      prompts, hidden reasoning, raw model output, credentials, or unrestricted logs.
- [ ] One coordinator lease fences a parent run while per-child leases and runtime reservations
      allow disjoint issue workspaces to execute concurrently; shared parent integration serializes.
- [ ] After any child advances the parent branch, stale siblings refresh/rebase or reclaim through
      typed Git capabilities and must repeat exact-head verification and Codex review.
- [ ] Human-gate comments cause a durable wait; prepare/publish/observe/consume/apply checkpoints
      survive crashes and a valid consumed reply resumes its exact bound effect once.
- [ ] Parent finalization revalidates exact head and requests fresh human merge approval; no agent
      score or prose can bypass it.
- [ ] Bare/help/invalid command behavior and documentation are contextual and deterministic.

## TDD and verification

Use the complete fake-port RED matrix in `../traces/479-preflight-interface-audit.md` before wiring.
It includes disjoint mutator concurrency, parent refresh/rebase, fault injection at every external
effect boundary, cancellation/join ordering, stale-generation fencing, and exact-head re-review.
Required skills: `gsd-programming-loop`, `gsd-workstreams`, `javascript-testing-patterns`,
`architecture-patterns`, CLI help/docs parity.

```bash
node --test .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
```

Human gates: this slice must wait for broker decisions; tests must prove it cannot self-approve.
