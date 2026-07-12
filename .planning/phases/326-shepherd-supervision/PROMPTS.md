# Prompt Snapshot: Issue #326

## Worker contract

- Objective: fence and supervise the existing `scripts/pi-shepherd-loop.sh` lifecycle.
- Output: one stacked, test-first shell change with focused process/concurrency evidence and a
  condensed handoff.
- Tool guidance: local temporary repositories and fake executables only; use `apply_patch` for repo
  edits; no model/provider invocation.
- Boundaries: production launcher + supervision test + Make target + issue memory only; Phase 0
  remains closed; no alternate controller implementation or destructive cleanup.

## Review focus

- No pre-guard effects or production bypass.
- Lock is worktree-wide and held for controller lifetime.
- Authority uses the exact persisted fence, never PID alone.
- Hard deadlines cannot be extended by descendants.
- Teardown targets only the registered process group and proves quiescence.
- HALT/recovery is durable before exit and cannot be resumed by flags or environment.

