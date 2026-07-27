# Verification checklist - Binary v0.1.0 correction

- [x] Branch reconciled to `origin/ci/release-publishing` without force/stash/reset/deletion.
- [x] `website-release` removed from `.github/workflows/release.yml`.
- [x] No PM binary release doc claims website dispatch/deploy.
- [x] No `fm/*` exception remains in `.github/workflows/conventions.yml`.
- [x] `CONTRIBUTING.md` no longer claims `fm/*` branch-name exemption.
- [ ] Website source/tests/workflow/deploy docs and issue-guard repair traces excluded from branch diff.
- [ ] Release Please one-shot `Release-As: 0.1.0` evidence captured.
- [ ] GoReleaser snapshot + `scripts/verify-release-assets.sh dist` pass.
- [ ] Replacement PR opened with exact title and `Refs #67`.
- [ ] no-mistakes validation reaches green/checks-passed.
