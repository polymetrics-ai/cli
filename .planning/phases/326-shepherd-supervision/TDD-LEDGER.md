# TDD Ledger: Shepherd Supervision

## Workflow preflight — 2026-07-12

- Issue #326 was narrowed before edits to the existing Shepherd launcher lifecycle.
- `github-issue-first-delivery` and `gsd-programming-loop` are active. The repository GSD adapter
  still omits `programming-loop`; the installed helper plus universal/manual TDD loop is the
  recorded fallback.
- Parent branch `fix/323-auto-loop-hardening` and draft PR #324 exist; the branch starts from the
  verified alignment commit `aef7fe86`.
- The task is shell/control-plane work; no Go or UI implementation skill is required for this slice.
- Production edits: none at this checkpoint.

## RED matrix

| Scenario | Current unsafe behavior | Required result |
|---|---|---|
| 32 concurrent starts | multiple providers may launch | one winner; 31 `CONTROLLER_HELD` |
| hard deadline + grandchild | child liveness can extend turn | TERM/KILL entire exact group; no validator |
| leader exits, child lives | root exit can look complete | durable orphan HALT; no checkpoint/validator |
| validator HALT | process-only exit | atomic HALT before quiescent exit |
| resume after HALT | prompt is read and loop restarts | reject before prompt/provider; bytes unchanged |
| signal during turn | descendants may survive | drain exact group; `recovery_required` |
| fence moves | old controller keeps mutating | drain and fail `CONTROL_FENCE_MISMATCH` |
| pause/resume at cap | turn counter resets | persisted monotonic ordinal and cap |

RED command/output will be appended after the tests exist and before production changes.

## RED checkpoint — 2026-07-12

- Added `scripts/tests/pi-shepherd-supervision.sh`. It copies the exact launcher body into a
  temporary fixture and mechanically removes only the Phase 0 guard; production has no enable flag
  or environment bypass. Fake Pi processes, synthetic local state, and test-owned PIDs are used.
- `bash -n scripts/tests/pi-shepherd-supervision.sh` -> pass.
- `bash scripts/tests/pi-shepherd-supervision.sh` -> expected exit 1 with 11 failures:
  - all 32 concurrent controllers launched an orchestrator; no contender received the required
    controller-held result;
  - hard deadline exited 3, started a validator, and wrote no halted/quiescent control state;
  - leader exit with a live descendant was accepted, validator started, and the child survived;
  - validator HALT produced no durable control latch;
  - signal termination left its descendant alive and wrote no recovery state;
  - the turn cap/ordinal was not persisted for resume.
- Production files changed before RED: none.
- Expected-failure integrity: the harness cleaned every recorded test child on exit; no external
  provider, GitHub mutation, credential, or non-test process was touched.

## RED harness audit

- An independent read-only audit confirmed each baseline failure maps to a missing launcher
  behavior, then identified potential false-green/test-safety gaps before implementation.
- The harness now removes exactly one sentinel-delimited Phase 0 block, uses a clean environment,
  verifies nonce-bound PID identity before cleanup, preserves same-user bystanders outside the
  role group, measures hard-deadline elapsed time, requires exact contention exits/codes, renames
  rather than chmods the resume prompt, and adds the previously missing fence-movement case.
- Audited RED: `bash scripts/tests/pi-shepherd-supervision.sh` exits 1 with 13 expected assertions,
  including 32/32 controller launches, natural deadline completion, live orphan, missing durable
  HALT/recovery, absent active fence, and reset turn cap. Bystanders survive, and only nonce-bound
  test processes are eligible for cleanup.
