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
- matching tests
- `.pi/README.md` command/operator section
- CLI help/manual parity artifacts applicable to `/pm-shepherd`
- this issue's GSD/TDD artifacts

## Acceptance criteria

- [ ] `start` drives intake through scheduling without requiring an external shell driver.
- [ ] Every ready non-colliding child is dispatched up to configured concurrency; dependencies and
      collisions explain deterministic waiting.
- [ ] Workers plan, implement red-green-refactor, verify, open/update PRs, review, correct, and
      integrate through authoritative adapters with bounded retry budgets.
- [ ] `status`, `stop`, and `resume` expose durable stage/lane/dependency/review/gate truth and join
      all child work correctly across stop/shutdown races.
- [ ] Human-gate comments cause a durable wait; a valid consumed reply resumes exactly once.
- [ ] Parent finalization revalidates exact head and requests fresh human merge approval; no agent
      score or prose can bypass it.
- [ ] Bare/help/invalid command behavior and documentation are contextual and deterministic.

## TDD and verification

Use end-to-end fake-port RED scenarios before wiring. Required skills: `gsd-programming-loop`,
`gsd-workstreams`, `javascript-testing-patterns`, `architecture-patterns`, CLI help/docs parity.

```bash
node --test .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
```

Human gates: this slice must wait for broker decisions; tests must prove it cannot self-approve.
