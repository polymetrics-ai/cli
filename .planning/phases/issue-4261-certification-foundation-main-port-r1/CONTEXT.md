# Issue 4261 — Certification foundation port to main

## Fixed decisions

- Port only the certification-foundation content from `ac2944115` onto the
  current `origin/main` base (`d842c815a6ab5a73cbcf154569f7426d2720e46a`).
- Target the pull request at `main`; the integration branch is not a merge
  candidate.
- Regenerate generated GitHub certification artifacts from this branch. The
  regenerated output is authoritative if it differs from the source commit.
- Exclude release metadata, unrelated planning deletions, and
  `internal/app/sync_modes.go`, which is absent from the source foundation
  commit.
- The explicitly authorized GitHub proof reads only the designated Keychain
  credentials at point of use. No credential value is written, logged, passed
  in argv, or included in evidence.

## Delivery constraints

- The desired code/content source is commit `ac2944115`; the desired merge
  base is `origin/main`, not the source branch.
- The finished diff must leave `CHANGELOG.md` and
  `.release-please-manifest.json` unchanged relative to `origin/main`.
- Local `make verify` is required before each push and is installed as the
  worktree's pre-push guard.
