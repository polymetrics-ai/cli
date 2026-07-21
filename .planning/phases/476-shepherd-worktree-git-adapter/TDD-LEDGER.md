# Issue 476 TDD Ledger

## Policy

Production code does not begin until deterministic tests fail for the intended missing behavior.
Tests use bounded temporary local Git repositories and no credentials or network services.

| Slice | RED command/evidence | GREEN command/evidence | Refactor/broad evidence | Status |
|---|---|---|---|---|
| Typed Git adapter | Pending | Pending | Pending | planned |
| Isolated workspace ownership | Pending | Pending | Pending | planned |
| Crash/retry and collision safety | Pending | Pending | Pending | planned |

## Required safety cases

- canonical repository identity is stable across linked worktrees and rejects a mismatched binding
- canonical issue branch, parent base, path, remote, ref, and relative scope validation
- typed command construction cannot express force, reset, arbitrary refspec, or cleanup
- default-branch push is unavailable
- dirty, untracked, conflicted, and unique state remains present and is reported
- base/head SHA evidence is exact and validated as 40 lowercase hexadecimal characters
- exact existing workspace reconciliation is idempotent after crash/retry
- alias or duplicate branch/worktree ownership fails closed
- concurrent create attempts cannot yield two active mutators

## GSD gate

- Mode: `manual_gsd_fallback`
- Adapter failure: `scripts/gsd: unknown GSD command or prompt: programming-loop`
- Execution decision, plan cycle: `local_critical_path` — this worker already owns one isolated
  sub-issue worktree and no further delegation is authorized or needed.

