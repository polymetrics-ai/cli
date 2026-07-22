# TDD ledger

## Evidence policy

The phase required behavior-level RED → GREEN → REFACTOR evidence for all 17 preflight rows. The
history does not prove that process completely. Commit `9848410e` added
`production-controller.test.ts` while `production-controller.ts` did not exist, so it established a
missing-module/compile RED only. The program's own stable-head convergence contract says that a
missing-module or compile-only failure is not behavior-level RED. No saved failing output proves
that all 17 assertions first executed against a compiling scaffold and failed for their intended
reason.

Later focused and full-suite GREEN runs are valid regression evidence for the final implementation,
but they cannot retroactively manufacture the missing behavior-level RED checkpoint. This process
gap remains open in the phase record even if every final functional gate passes.

## Slice ledger

| Slice | RED evidence | GREEN/regression evidence | Refactor/correction evidence | Process status |
|---|---|---|---|---|
| Plan and production contract | No separate behavior RED; `2e488a43` added contract and tests together. | Contract tests validate exact schema-2 input, bounded values, mutating-only top-level work, and hostile-shape rejection. | Subsequent validator and state corrections reuse the frozen DTOs. | GREEN regression; RED not evidenced. |
| Intake, scheduler, and bounded verification | No separate behavior RED; commits `7bfe115b`, `78034a8d`, and `a2cf90be` co-located tests and production. | Intake cancellation/non-symlink tests, deterministic dependency/collision/capacity tests, and fixed executable/argv/cwd/output/timeout tests. | Missing-intake error corrected in `bec16415`. | GREEN regression; RED not evidenced. |
| Durable state, effects, and recovery barrier | No separate behavior RED; commits `35787518`, `616a3429`, and `6fea793d` co-located tests and production. | CAS/binding/budget/generation tests plus prepared → observed → applied journal and ordered recovery tests. | Retry, lock, ownership, checkpoint, and exact merge evidence corrections in `236260dc` through `cf955711`. | GREEN regression; RED not evidenced. |
| Controller trajectory | `9848410e` is a test-only missing-module/compile RED, explicitly insufficient as behavior RED. | `81d3062d` supplied the controller; controller tests exercise dependency-ready parallel work, correction, stale-parent refresh, budget waits, stop/join, sibling abort, binding rejection, durable effect acknowledgement, and observed merge. | `d84e4740` added race coverage; later checkpoint and merge-evidence corrections refined state transitions. | Functional GREEN evidence; behavior RED gap remains. |
| Workspace/Git lifecycle | No separate behavior RED for the initial slice; production and tests arrived together. | Lifecycle tests cover stable ownership, isolated claims, verification-before-commit, refresh/rebase/reverify, per-run join/release, and commit/push reconciliation. | Lease, orphan-lock, resume ownership, refreshed-base, and planned-parent binding fixes include `86238846`, `16994d1c`, `2fba70be`, `d2f71aa0`, and `b31424b1`. | GREEN regression; RED not evidenced. |
| GitHub, review, and human decisions | No separate behavior RED; `3628cc90` and later parent-lifecycle commits co-located tests and production. | Deterministic transport, exact-head review/dispositions, fail-closed GitHub evidence, exact human commands, draft→ready, and observed human merge tests. | Exact parent lifecycle hardened in `d138fb99`; final-head fixes are recorded by the final verification head. | GREEN regression; RED not evidenced. |
| Production child pipeline and runtime composition | No separate behavior RED; `7fecbe86` added pipeline and tests together. Runtime composition began as an uncommitted test scaffold and must be judged from final history. | Child tests cover all stage checkpoints, correction dispositions, refresh, intervention, timeout reconciliation, hostile evidence, abort, and close. Runtime factory/composition tests are recorded in final verification. | Planned-parent binding fixed in `b31424b1`; later corrections are listed in the final Git head. | Final-head GREEN pending in this ledger until verification is frozen. |
| Documentation | Docs-only; behavior RED is not applicable. | Schema-2 README sample is parsed and validated by the production plan validator; documentation grep, secret scan, JSON parse, and diff checks are recorded in `VERIFICATION.md`. | Stale schema-1/local-MVP claims removed. | Pending final docs gate. |

## TDD disposition

- Functional result: determined by the final 17-row matrix and machine gates in `VERIFICATION.md`.
- Historical process result: **OPEN — behavior-level RED evidence is incomplete**.
- No future test-only commit should be described as the original RED for already implemented
  behavior. Any additional correction must add a genuinely failing behavior test first, capture its
  failure, then implement the smallest GREEN change and rerun the stable-head gates.
