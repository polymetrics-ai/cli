# Objective

Turn the existing read-only Shepherd implementation into a release-quality autonomous control-plane
foundation: durable state and lease ownership, lifecycle linearization, exact target evidence,
framework-independent ports, and fail-closed persisted-state validation.

Parent: #471
Parent PR: #472
Dependency: none (Wave 1 critical path)
Branch: `feat/473-shepherd-control-plane-foundation`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/{domain,state-store,target-evidence,sdk-runner,controller,extension,index}.ts`
- matching tests beside those files
- this issue's GSD/TDD artifacts

No autonomous scheduling, GitHub mutation, or worker-write functionality belongs in this slice.

## Acceptance criteria

- [ ] First-wins stop/shutdown cancellation is linearizable and joins every owned child/save.
- [ ] Lease epoch rollover cannot return an orphan/non-authoritative lease under cleanup races.
- [ ] A state store pins root path plus device/inode for its lifetime; root replacement fails closed.
- [ ] Root-security failures are never swallowed as best-effort cleanup errors.
- [ ] Persisted assessed states require nonempty lanes, finite aggregates, coherent outcomes, and
      aggregate hard-gate coverage.
- [ ] Epoch bounds, malformed journals, stale evidence, exact head/PR URL, and restart reconciliation
      are adversarially tested.
- [ ] macOS documents and enforces a private trusted same-user state root; it does not claim a
      hostile same-UID boundary without native descriptor-relative operations.
- [ ] Existing read-only canary behavior remains green as foundation regression coverage.

## TDD and verification

Use `gsd-programming-loop`: add deterministic RED race/invariant tests, implement the smallest
GREEN fix, then refactor. Required skills: `javascript-testing-patterns`, `architecture-patterns`,
and repository security routing.

```bash
node --test .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
```

Human gates: none for local implementation. Do not merge the sub-PR unless parent-orchestrator
verification and review coverage are recorded. Use `Refs`, not closing keywords.
