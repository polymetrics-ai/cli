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
- [ ] Replacement PR opened with exact title and `Refs #67`.
- [ ] Superseded pointers added to PR #528 and PR #538.
- [ ] PR #538 closed unmerged after replacement confirmation.

## Evidence

- Branch-name script from `.github/workflows/conventions.yml`: `ci/pm-v0.1.0-release` exits 0; `fm/cli-release-and-connector-issues-r1` exits 1; YAML parse passed.
- Net replacement branch diff contains only `.planning/phases/binary-v0.1.0-correction/**`, `CONTRIBUTING.md`, and `docs/release-and-connectors.md`; revoked website/source/test/workflow/issue-guard files are net-clean.
- Clean candidate commit SHA pending first replacement commit; transient `release-please@latest release-pr --help` confirms `--release-as` support, and upstream README states commit-body `Release-As: x.x.x` opens the specified release PR. Full dry-run was blocked by GitHub GraphQL credentials, so no token was requested or printed.
- `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean` passed.
- `scripts/verify-release-assets.sh dist` passed before generated `dist/` cleanup; verified six archives plus `checksums.txt`.
