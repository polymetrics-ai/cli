# Summary — Time-boxed rebase-safe post-commit push

## Delivered

- Added a POSIX `post-commit` hook that only schedules a normal feature-branch
  push when no Git operation is in progress, HEAD is attached and non-default,
  every effective push endpoint's live default has been checked, and the
  per-branch 600-second window has expired; it honors Git's branch-specific,
  global, then tracking push destination precedence.
- Kept status state below the shared common Git directory and recorded the rate
  timestamp behind a per-branch Git ref, so `git update-ref` makes linked
  worktree rate decisions atomic without a stale lease.
- Added a worktree-local `prepare-commit-msg` companion that records a
  commit-scoped Git-resolved operation marker before cleanup, using a shared
  operation helper so `post-commit` also refuses amended squash and clean
  no-commit completions.
- Detached the push from the commit terminal, recorded failed/rejected pushes as
  one local line, supplied no force or force-with-lease path, and bound each
  detached child to the HEAD that scheduled it after it rechecks Git operation
  state.
- Added an executable local-Git harness and documented explicit enablement plus
  `PM_NO_AUTOPUSH=1` opt-out. No hook configuration, installer, CI workflow, or
  server hook changed.

## Verification

- `shellcheck -s sh .githooks/pm-autopush-operation-state .githooks/prepare-commit-msg .githooks/post-commit scripts/tests/post-commit-autopush.sh`
- `sh scripts/tests/post-commit-autopush.sh`
- `make docs-check`
- `git diff --check`

The harness covers live, stale, and pushurl-specific remote defaults, detached
HEAD, opt-out, rate window and catch-up, concurrent linked-worktree rate-ref
compare-and-swap, a receive-delayed asynchronous push, delayed children during
and after a manual merge, a real two-commit rebase, manual, clean, and amended
squash/cherry-pick/revert completion, configured push-target precedence, and a
real non-fast-forward rejected push.
