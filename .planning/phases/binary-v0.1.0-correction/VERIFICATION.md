# Verification checklist - Binary v0.1.0 correction

- [x] Replacement branch `ci/pm-v0.1.0-release` created from `origin/main`.
- [x] `ci/release-publishing` left intact; no rewrite, force-push, or deletion performed.
- [x] No `website-release` job exists in `.github/workflows/release.yml`.
- [x] No PM binary release doc claims website publication.
- [x] No `fm/*` exception remains in `.github/workflows/conventions.yml`.
- [x] `CONTRIBUTING.md` no longer claims `fm/*` branch-name exemption.
- [x] Website source/tests/workflow docs and issue-guard repair traces excluded from branch diff.
- [x] Release Please one-shot `Release-As: 0.1.0` evidence captured.
- [x] Previous GoReleaser snapshot + `scripts/verify-release-assets.sh dist` evidence preserved.
- [x] Replacement PR #539 opened with exact title and `Refs #67`.
- [x] Superseded pointers added to PR #528 and PR #538.
- [x] PR #538 closed unmerged after replacement confirmation.

## Evidence

- Branch-name script from `.github/workflows/conventions.yml`: `ci/pm-v0.1.0-release` exits 0; `fm/cli-release-and-connector-issues-r1` exits 1; YAML parse passed.
- Net replacement branch diff contains only `.planning/phases/binary-v0.1.0-correction/**`, `CONTRIBUTING.md`, and `docs/release-and-connectors.md`; revoked website/source/test/workflow/issue-guard files are net-clean.
- Commit `7d019a11` body contains `Release-As: 0.1.0`; transient `release-please@latest release-pr --help` confirms `--release-as` support, and upstream README states commit-body `Release-As: x.x.x` opens the specified release PR. Full dry-run was blocked by GitHub GraphQL credentials, so no token was requested or printed.
- `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean` passed.
- `scripts/verify-release-assets.sh dist` passed before generated `dist/` cleanup; verified six archives plus `checksums.txt`.
- Focused clean-branch verification passed: `git log --oneline origin/main..HEAD`, `git merge-base HEAD origin/main`, `git diff --name-status origin/main..HEAD`, `git diff --exit-code` for release/website/conventions/issue-guard paths, `git show --no-patch --format=%B 7d019a11`, and changed-doc `rg` checks for forbidden website workflow/deploy/dispatch markers.
- PR #539: https://github.com/polymetrics-ai/cli/pull/539.
