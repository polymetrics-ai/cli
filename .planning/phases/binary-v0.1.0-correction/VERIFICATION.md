# Verification checklist - Binary v0.1.0 correction

- [x] Branch reconciled to `origin/ci/release-publishing` without force/stash/reset/deletion.
- [x] `website-release` removed from `.github/workflows/release.yml`.
- [x] No PM binary release doc claims website dispatch/deploy.
- [x] No `fm/*` exception remains in `.github/workflows/conventions.yml`.
- [x] `CONTRIBUTING.md` no longer claims `fm/*` branch-name exemption.
- [x] Website source/tests/workflow/deploy docs and issue-guard repair traces excluded from branch diff.
- [x] Release Please one-shot `Release-As: 0.1.0` evidence captured.
- [x] GoReleaser snapshot + `scripts/verify-release-assets.sh dist` pass.
- [ ] Replacement PR opened with exact title and `Refs #67`.
- [ ] no-mistakes validation reaches green/checks-passed.

## Evidence

- Branch-name script from `.github/workflows/conventions.yml`: `ci/release-publishing` exits 0; `fm/cli-release-and-connector-issues-r1` exits 1; YAML parse passed.
- Net PR diff after correction contains only `.planning/phases/binary-v0.1.0-correction/**`, `CONTRIBUTING.md`, and `docs/release-and-connectors.md`; revoked website/source/test/workflow/issue-guard files are net-clean.
- Commit `29cd7b1ab` body contains `Release-As: 0.1.0`; transient `release-please@latest release-pr --help` confirms `--release-as` support, and upstream README states commit-body `Release-As: x.x.x` opens the specified release PR. Full dry-run was blocked by GitHub GraphQL credentials, so no token was requested or printed.
- `go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean` passed.
- `scripts/verify-release-assets.sh dist` passed before generated `dist/` cleanup; verified six archives plus `checksums.txt`.
