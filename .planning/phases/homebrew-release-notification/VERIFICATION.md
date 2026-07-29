# Verification checklist — Homebrew release notification

## Focused checks

- [x] `scripts/tests/homebrew-release-notify.sh` — passed
- [x] `make release-workflow-check` — passed

## Static coverage requirements

- [x] Authorized dry-run notification uses only `.github/workflows/pm-formula-update.yml` with `dry_run=true`.
- [x] Missing App credentials fail without PAT, `GITHUB_TOKEN`, `GH_TOKEN`, or ambient fallback.
- [x] Wrong event ordering / failed upstream verification fails before notification.
- [x] Duplicate notifications are deterministic and dry-run only.
- [x] Malformed tags, wrong repositories, and wrong workflows fail closed.
- [x] Website workflow cannot trigger the Homebrew notification path.
- [x] No live mutation path (`dry_run=false`) is enabled.

## Broader local gates if time permits before handoff

- [x] `git diff --check` — passed
- [x] `shellcheck scripts/notify-homebrew-formula-update.sh scripts/tests/homebrew-release-notify.sh` — passed
- [x] YAML syntax check via Ruby `YAML.load_file` — passed
- [ ] `make verify` or bounded subset if full gate is deferred to firstmate/no-mistakes.

## Deferred / forbidden in this task

- [ ] Do not dispatch the live Homebrew workflow during implementation or validation.
- [ ] Do not publish a PM release.
- [ ] Do not create Homebrew formula branches or PRs.
- [ ] Do not access secret values or change repository/App settings.
