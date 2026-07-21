# Objective

Implement the pure autonomous delivery policy: parent lifecycle stages, dependency DAG validation,
bounded ready-queue selection, write-scope collision arbitration, retry/correction budgets, and
deterministic reconciliation decisions.

Parent: #471
Parent PR: #472
Dependency: #473 (Wave 2)
Branch: `feat/474-shepherd-dependency-policy`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/autonomy-policy.ts`
- `.pi/extensions/shepherd/dependency-graph.ts`
- `.pi/extensions/shepherd/reconciler.ts`
- matching tests
- this issue's GSD/TDD artifacts

Shared controller/extension wiring is reserved for the integration issue.

## Acceptance criteria

- [ ] The state machine covers intake through complete, including correction and human-wait states.
- [ ] Cycles, unknown dependencies, ambiguous scopes, and unsafe transitions fail closed.
- [ ] All dependency-ready non-colliding lanes are selected up to configured concurrency.
- [ ] Mutating scope collisions serialize; read-only research/review can coexist when safe.
- [ ] Retry budgets distinguish transient verification/review failures from hard human gates.
- [ ] Reconciliation is pure and idempotent for the same persisted/Git/GitHub snapshot.
- [ ] Every no-spawn decision emits exactly one repository contract blocker category.

## TDD and verification

Use `gsd-programming-loop` with table/property-style RED cases before implementation. Required
skills: `architecture-patterns`, `javascript-testing-patterns`, `gsd-workstreams`.

```bash
node --test .pi/extensions/shepherd/autonomy-policy.test.ts \
  .pi/extensions/shepherd/dependency-graph.test.ts \
  .pi/extensions/shepherd/reconciler.test.ts
git diff --check
```

Human gates: policy must represent gates but this pure slice performs no external action.
