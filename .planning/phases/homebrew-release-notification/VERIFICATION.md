# Verification checklist — Homebrew release notification

## Focused checks

- [ ] `scripts/tests/homebrew-release-notify.sh`
- [ ] `make release-workflow-check`

## Static coverage requirements

- [ ] Authorized dry-run notification uses only `.github/workflows/pm-formula-update.yml` with `dry_run=true`.
- [ ] Missing App credentials fail without PAT, `GITHUB_TOKEN`, `GH_TOKEN`, or ambient fallback.
- [ ] Wrong event ordering / failed upstream verification fails before notification.
- [ ] Duplicate notifications are deterministic and dry-run only.
- [ ] Malformed tags, wrong repositories, and wrong workflows fail closed.
- [ ] Website workflow cannot trigger the Homebrew notification path.
- [ ] No live mutation path (`dry_run=false`) is enabled.

## Broader local gates if time permits before handoff

- [ ] `git diff --check`
- [ ] `make verify` or bounded subset if full gate is deferred to firstmate/no-mistakes.

## Deferred / forbidden in this task

- [ ] Do not dispatch the live Homebrew workflow during implementation or validation.
- [ ] Do not publish a PM release.
- [ ] Do not create Homebrew formula branches or PRs.
- [ ] Do not access secret values or change repository/App settings.
