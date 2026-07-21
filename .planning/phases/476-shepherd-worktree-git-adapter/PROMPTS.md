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
- Downstream artifact: `.planning/phases/476-shepherd-worktree-git-adapter/SUMMARY.md` (completed)
- Verification result: authoritative narrowed local gates pass; full Go/connectors rerun is owned
  by parent integration and GitHub CI under the updated parent policy.

## Exact-head review correction

- Reviewed head: `906a45c53ae1a19c9d2efe1c3f24a64e36ef4d63`.
- Blockers: v1 Shepherd identity compatibility, immutable persisted handoff bindings, and an
  exclusive same-owner writable lease with release and dead-owner resume.
- Warning disposition plan: serialize full Shepherd test files so real Git subprocess load cannot
  invalidate the SDK runner's intentional wall-clock assertions; do not widen those assertions.
- Authorized gates remain TypeScript/Shepherd/Pi/diff-only; no Go, connector, or `make verify` run.
- Result: correction RED `36860ec5`, GREEN `e3669fc4`, and refactor `d91b41a8`; focused 21/21,
  serialized full Shepherd 158/158, strict TypeScript, offline Pi 0.80.6 RPC, and diff/scope pass.
- Ready stacked PR: https://github.com/polymetrics-ai/cli/pull/484.
