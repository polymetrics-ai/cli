# Review — Time-boxed rebase-safe post-commit push

## Manual GSD code-review fallback

The generated `code-review` prompt was resolved, but the task requires a
single worker and the available runtime cannot provide an isolated review
agent. A focused manual review was performed after the executable harness and
shellcheck passed.

## Scope reviewed

- `.githooks/post-commit`
- `scripts/tests/post-commit-autopush.sh`
- `docs/GUIDE.md`
- the companion GSD evidence

## Follow-up findings

- R1 was legitimate: Git can remove a manual merge, cherry-pick, or revert
  marker before `post-commit`. A tracked `prepare-commit-msg` companion now
  records a Git-resolved operation snapshot in the worktree Git directory, and
  `post-commit` consumes it only when its parent matches the just-created
  commit.
- R2 was legitimate: a local remote `HEAD` symref can be absent or stale. The
  detached child resolves the remote's live symbolic `HEAD`, logs an unavailable
  or matching default branch, and returns without push.
- R3 was legitimate: the old read/check/write sequence was not a lease. A
  common-Git-dir sibling directory now atomically protects the rate decision and
  timestamp replacement.

The focused executable harness covers each follow-up alongside the original
non-force rejection and detached-push behavior.
