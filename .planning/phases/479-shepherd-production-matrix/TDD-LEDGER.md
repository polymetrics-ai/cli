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
| Production child pipeline and runtime composition | No separate behavior RED for the original slice; `7fecbe86` added pipeline and tests together. | Final combined production suite passes 808/808 and reaches the genuine Pi host composition. | Planned-parent binding fixed in `b31424b1`; exact effect recovery, parent lifecycle, and integration-CAS corrections end at code head `91692415`. | Functional GREEN; original behavior RED gap remains. |
| Documentation | Docs-only; behavior RED is not applicable. | Schema-2 README sample is parsed and validated; command RPC, documentation, JSON, and diff gates are recorded in `VERIFICATION.md`. | Stale schema-1/local-MVP claims removed; exact CAS/recovery behavior documented. | PASS. |

## Final correction RED → GREEN evidence

| Review blocker | Behavior RED | GREEN correction | Final evidence |
|---|---|---|---|
| Generation-2 child intervention recovery | `0fe22e9e`: fresh-process generation-2 request failed because recovery reconstructed generation 1. | `e2dedad7`: recovery binds the child decision to current execution generation while resource identities retain resource generation. | Runtime 10/10; recovery 83/83; controller 34/34. |
| Start/resume default-branch authority | `06e50e21`: resume accepted remote-default drift and a `trunk` → `main` parent target. | `a8104613`: start/resume re-inspect exact repository/remote authority and reject default aliases. | Host/extension 25/25; real offline Pi RPC passed. |
| Atomic parent-base integration | `5ef7ba15`: GitHub's head-only merge path accepted a parent race and did not provide an exact base CAS. | `37dbc42c`: deterministic exact-base/head merge plus `--force-with-lease` parent ref update; GitHub becomes receipt-only. | Exact CAS 3/3; orchestrator 515/515; pipeline 17/17. |
| Crash after Git CAS before receipt | `32a0d50e`: fresh-process retry stopped at `stale_parent` after one physical CAS and no receipt. | `91692415`: retry proves/reuses the exact deterministic merge, publishes one receipt, and still rejects unrelated parent movement. | Pipeline/orchestrator/evidence 666/666; combined final suite 808/808. |

## TDD disposition

- Functional result: **PASS — all 17 rows are implemented and verified** at production code head
  `91692415`; the later test/docs freeze does not change production code.
- Historical process result: **OPEN — behavior-level RED evidence is incomplete**.
- No future test-only commit should be described as the original RED for already implemented
  behavior. Any additional correction must add a genuinely failing behavior test first, capture its
  failure, then implement the smallest GREEN change and rerun the stable-head gates.
