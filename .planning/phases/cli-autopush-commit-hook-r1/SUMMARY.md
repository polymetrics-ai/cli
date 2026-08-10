# Summary — Time-boxed rebase-safe post-commit push

## Delivered

- Added a POSIX `post-commit` hook that only schedules a normal feature-branch
  push when no Git operation is in progress, HEAD is attached and non-default,
  the live remote default has been checked, and the per-branch 600-second window
  has expired.
- Kept rate and status state below the shared common Git directory, resolved by
  `git rev-parse --git-path`, with an atomic lease so linked worktrees do not fork
  or race rate-limit state.
- Added a worktree-local `prepare-commit-msg` companion that snapshots a
  Git-resolved operation before commit cleanup, allowing `post-commit` to refuse
  manually completed merges, cherry-picks, and reverts too.
- Detached the push from the commit terminal, recorded failed/rejected pushes as
  one local line, and supplied no force or force-with-lease path.
- Added an executable local-Git harness and documented explicit enablement plus
  `PM_NO_AUTOPUSH=1` opt-out. No hook configuration, installer, CI workflow, or
  server hook changed.

## Verification

- `shellcheck -s sh .githooks/post-commit scripts/tests/post-commit-autopush.sh`
- `sh scripts/tests/post-commit-autopush.sh`
- `make docs-check`
- `git diff --check`

The harness covers live and stale remote defaults, detached HEAD, opt-out, rate
window and catch-up, concurrent linked-worktree leasing, a receive-delayed
asynchronous push, a real two-commit rebase, manual merge/cherry-pick/revert
completion, and a real non-fast-forward rejected push.
