# Objective

Implement typed, auditable Git and isolated-worktree operations for autonomous issue workers, with
canonical ownership, safe branch/base rules, collision prevention, and exact-head verification.

Parent: #471
Parent PR: #472
Dependency: #473 (Wave 2)
Branch: `feat/476-shepherd-worktree-git-adapter`
PR base: `feat/471-pi-agent-session-shepherd`

## Allowed write scope

- `.pi/extensions/shepherd/workspace-adapter.ts`
- `.pi/extensions/shepherd/git-adapter.ts`
- matching tests and bounded fixtures
- this issue's GSD/TDD artifacts

Do not dispatch agents or mutate GitHub in this slice.

## Acceptance criteria

- [ ] One mutating issue maps to one canonical branch and one isolated worktree outside the
      coordinator checkout; duplicate or aliased ownership fails closed.
- [ ] Branch names, parent base, paths, remotes, and repository identity are validated as untrusted.
- [ ] Typed operations cover only required status/fetch/branch/commit/push/diff/worktree actions.
- [ ] No force push, direct push to `main`, destructive reset, arbitrary refspec, or unbounded path
      is available to an autonomous worker.
- [ ] Dirty/untracked/unique state is preserved and reported; cleanup requires a separate policy.
- [ ] Every handoff records exact base/head SHAs, changed scope, verification state, and PR base.
- [ ] Crash/retry is idempotent and cannot create two active mutators for one branch/worktree.

## TDD and verification

Use temporary local repositories and deterministic RED safety cases. Required skills:
`javascript-testing-patterns`, `architecture-patterns`, `gitops-workflow`.

```bash
node --test .pi/extensions/shepherd/workspace-adapter.test.ts \
  .pi/extensions/shepherd/git-adapter.test.ts
git diff --check
```

Human gates: destructive cleanup and any direct-default-branch action remain unavailable.
