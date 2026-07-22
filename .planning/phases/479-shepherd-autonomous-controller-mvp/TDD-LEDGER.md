# TDD ledger

## RED

Behavior contract and compiling port/state scaffold committed before controller behavior. Command:

```bash
node --test --test-concurrency=1 .pi/extensions/shepherd/autonomous-controller.test.ts
```

Result: 1 executed, 0 passed, exactly 1 intended failure at
`AutonomousShepherdController.start`: `autonomous Shepherd MVP is not implemented`. The trajectory
covers parallel independent children, dependency waiting, execute/verify/review/integrate ordering,
durable state, stop/join/resume, the final human gate, and absence of a merge-main capability.

## GREEN

The production controller now schedules two non-colliding children, releases capacity as each
finishes, honors dependencies, advances execute/verify/review/integrate ports, persists schema-v2
state, joins stop, resumes only unfinished work, and creates a durable parent-merge wait with no
merge capability. Autonomous parser/extension routing is separate from the retained v1 canary.

Focused result: 25/25 pass across controller, local adapters, arguments, and extension. Strict
TypeScript passes all Shepherd production modules against the pinned Pi 0.80.6 types. Offline Pi
RPC loads the extension and reports `pm-shepherd`.

## REFACTOR

The concurrent effect assertion was narrowed from global serial order to per-child stage order;
parallel siblings may interleave by design. No production behavior was expanded and no speculative
hardening cycle was started.

## Single blocker-only review

One `gpt-5.6-sol/xhigh` exact-range review was run against base
`c958d7e608eac4afedceea953408a8fd4cfcb566`. It found only trajectory-blocking defects. The same
bounded correction pass:

- replaced the class workspace with the exact plain capability required by the hardened runtime;
- allocated one max-concurrency-one runtime per child lane so parallel mutators do not share a
  mutation lease, while retaining the same runtime across that child's ordered roles;
- keyed the concurrency test by workspace identity and bounded its wait; and
- routed status through the matching active canary before persisted autonomous fallback.

No second review or hardening cycle was started. Final focused result is 25/25; the proportional
Shepherd-safe suite is 1,339 passed, 0 failed, and 1 intentional skip.
