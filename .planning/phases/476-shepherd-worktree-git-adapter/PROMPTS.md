# Issue 476 Prompt Snapshot

## Kickoff

- Objective: implement typed, auditable Git and isolated-worktree operations for Shepherd issue
  workers, with canonical ownership, safe branch/base rules, collision prevention, exact-head
  verification, and non-destructive crash/retry behavior.
- Immutable base: `e659d6f1b666f58748e2d8c86599ceb4bbc62ff8`
- Branch: `feat/476-shepherd-worktree-git-adapter`
- Parent base: `feat/471-pi-agent-session-shepherd`
- Scope: the two adapter files, matching tests/fixtures, and this issue phase directory only.
- Human gates: destructive cleanup and any direct-default-branch action remain unavailable.
- GSD command result: `programming-loop` absent from repo adapter; manual lifecycle required.
- Downstream artifact: `.planning/phases/476-shepherd-worktree-git-adapter/SUMMARY.md` (in progress)
- Verification result: pending

