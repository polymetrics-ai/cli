# Verification checklist — Time-boxed rebase-safe post-commit push

## Required behavioral checks

- [x] Real two-commit rebase, manual merge/cherry-pick/revert, clean squash, clean no-commit cherry-pick, and clean revert completions invoke the hook and record no push.
- [x] `main`, a detached HEAD, and a stale locally tracked remote default record no push.
- [x] Every effective push endpoint is checked for its own live default before a remote with multiple `pushurl` values can send.
- [x] A branch-specific push remote, then `remote.pushDefault`, then the tracking remote receives the feature commit according to Git's push precedence.
- [x] A feature branch pushes once, skips a second commit inside 600 seconds, and catches up after expiry.
- [x] Concurrent forced linked worktrees compare-and-swap the same expired per-branch rate ref and schedule one receiver call.
- [x] A real non-fast-forward push is rejected without a force option, is logged once, and leaves the commit successful.
- [x] A deliberately delayed push does not delay the commit.
- [x] A delayed child cannot send while an operation is active or a later operation-generated branch tip.
- [x] `PM_NO_AUTOPUSH=1` skips the hook.
- [x] `docs/GUIDE.md` documents manual enablement and opt-out; no repository config, installer, CI, or server hook changes exist.

## Local command checklist

- [x] `shellcheck -s sh .githooks/pm-autopush-operation-state .githooks/prepare-commit-msg .githooks/post-commit scripts/tests/post-commit-autopush.sh`
- [x] `sh scripts/tests/post-commit-autopush.sh`
- [x] `make docs-check`
- [x] `git diff --check`
- [x] Focused manual code/security review of the hook, test harness, and documentation.
