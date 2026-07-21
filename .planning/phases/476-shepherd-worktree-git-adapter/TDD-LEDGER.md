# Issue 476 TDD Ledger

## Policy

Production code does not begin until deterministic tests fail for the intended missing behavior.
Tests use bounded temporary local Git repositories and no credentials or network services.

| Slice | RED command/evidence | GREEN command/evidence | Refactor/broad evidence | Status |
|---|---|---|---|---|
| Typed Git adapter | `node --test .pi/extensions/shepherd/workspace-adapter.test.ts .pi/extensions/shepherd/git-adapter.test.ts` → exit 1, `ERR_MODULE_NOT_FOUND` for absent `git-adapter.ts` | Same command → 16/16 pass | Strict no-emit TypeScript passed; raw Git errors reduced to bounded exit evidence | green/refactored |
| Isolated workspace ownership | Same command → exit 1 before collection because the required adapter modules do not exist | Genuine temporary bare-remote/worktree cases pass | Full Shepherd suite 153/153 pass | green/refactored |
| Crash/retry and collision safety | Deterministic genuine-repository cases were present but could not collect until the missing adapters existed | Exact retry, owner collision, concurrent owners, alias branch, path collision, stale base, dirty preservation, and unrelated-head cases pass | Canonical claim metadata stores owner hashes rather than caller text; no release/cleanup API | green/refactored |
| Correction: state identity parity | Focused run: adapter identity differed from `target-evidence.ts` for the same checkout | v1 repository/worktree identity parity passes for coordinator and linked worktree | Strict production TypeScript passes | green |
| Correction: immutable handoff claim | Focused run: all three mutable handoff variants were accepted and persisted claim lacked `allowedScopes` | atomic claim/binding records reject workspace and persisted-field tampering | Strict production TypeScript passes | green |
| Correction: exclusive writable lease | Focused run: both same-owner contenders fulfilled; returned workspace had no lease capability | same-owner race yields one lease; release/retry and dead-owner resume pass | Existing append-only `FileStateStore` fencing reused | green |

Correction RED command: `node --test .pi/extensions/shepherd/workspace-adapter.test.ts
.pi/extensions/shepherd/git-adapter.test.ts` → 21 tests, 16 passed, 5 failed. The five failures map
directly to the reviewed contracts; no production file was changed before this run.

Correction GREEN command: the same focused command → 21 tests passed, 0 failed. Strict no-emit
TypeScript over both production adapters and their imports also passed against the cached Pi 0.80.6
Node type surface.

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
- Execution decision, TDD gate cycle: `local_critical_path` — RED evidence captured before either
  production adapter file exists.
- Execution decision, execute cycle: `local_critical_path` — minimal typed Git and workspace
  adapters implemented inside the assigned files.
- Execution decision, refactor cycle: `local_critical_path` — validation, canonical scope handling,
  root pinning, and secret-safe Git failure reduction completed; focused and full suites pass.
- Execution decision, verify cycle: `local_critical_path` — authoritative narrowed local gates pass;
  parent integration owns the full Go/connectors rerun under the updated policy.
- Execution decision, summary cycle: `local_critical_path` — durable evidence and stacked PR handoff
  finalized without automated-review or merge authority.
