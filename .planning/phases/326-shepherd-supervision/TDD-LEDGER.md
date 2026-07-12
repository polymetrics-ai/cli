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

