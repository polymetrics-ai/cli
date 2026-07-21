# Issue 476 Summary

Status: `in_progress`

Implemented a typed Git process adapter and isolated-worktree policy adapter for Shepherd. The Git
port binds repository/worktree/origin identity, exposes bounded status/fetch/branch/commit/push/
diff/worktree operations, and verifies exact local and remote heads. The workspace policy derives
the only valid issue branch and path, persists hashed ownership claims, rejects aliases/collisions,
and reconciles an exact retry without removing dirty or unique state.

Focused tests pass 16/16 and the full Shepherd suite passes 153/153. Strict production TypeScript
also passes against the installed Pi 0.80.6 environment. Repository-wide gates and PR creation are
still pending.

No dependency, controller, domain, runner, extension, GitHub integration, destructive cleanup,
force/reset, default-branch push, arbitrary refspec, or unrestricted path capability was added.
